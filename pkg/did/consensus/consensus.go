package consensus

import (
	"context"
	"math/big"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	pubsubMsgsChanName = "/did/msgs"
	pubsubTxChanName   = "/did/tx"
)

type Consensus struct {
	id     peer.ID
	logger logrus.Logger

	db         Db
	memPool    MemPool
	blockStore BlockStore

	beacon <-chan int64
	p2p    *p2p

	state           State
	currentProposer peer.ID
	propsalState    State
}

func (c *Consensus) proposer() <-chan peer.ID {
	bCh := make(chan peer.ID)

	go func() {
		for b := range c.beacon {
			nodes, err := c.db.Nodes()
			if err != nil {
				c.logger.WithError(err).Error("getting node list")
				continue
			}

			m := big.NewInt(0)
			m.SetInt64(b)
			mn := m.Mod(m, big.NewInt(int64(len(nodes)))).Int64()

			p := nodes[mn]
			bCh <- p
		}
	}()

	return bCh
}

func (c *Consensus) Start() error {
	go c.watchProposer()

	if err := c.subscribeMsgs(); err != nil {
		return errors.Wrap(err, "watching msgs")
	}

	return c.subscribeTx()
}

func (c *Consensus) subscribeMsgs() error {
	sub, err := c.p2p.Msgs(pubsubMsgsChanName)
	if err != nil {
		return errors.Wrap(err, "subscribing to consensus msgs")
	}

	go func() {
		for msg := range sub {
			go c.OnMsg(msg)
		}
	}()

	return nil
}

func (c *Consensus) subscribeTx() error {
	sub, err := c.p2p.Msgs(pubsubTxChanName)
	if err != nil {
		return errors.Wrap(err, "subscribing to tx msgs")
	}

	go func() {
		for tx := range sub {
			go c.OnMsg(tx)
		}
	}()

	return nil
}

func (c *Consensus) watchProposer() {
	for id := range c.proposer() {
		c.propsalState.Step = propose

		if id == c.id {
			go c.StartRound()
		} else {
			//TODO(tcfw) start propose timeout
		}
	}
}

func (c *Consensus) StartRound() error {
	c.propsalState.AmProposer = true
	c.propsalState.lockedRound = -1
	c.propsalState.lockedValue = cid.Cid{}
	c.propsalState.Step = propose
	c.propsalState.Height = c.state.Height
	c.propsalState.Round = 0
	c.propsalState.Block = cid.Cid{}
	c.propsalState.PreVotes = make(map[peer.ID]*ConsensusMsgVote)
	c.propsalState.PreCommits = make(map[peer.ID]*ConsensusMsgVote)

	n, err := c.db.Nodes()
	if err != nil {
		return errors.Wrap(err, "getting nodes")
	}
	c.propsalState.f = (uint64(len(n))/3)*2 + 1

	//build block
	//upload block
	//broadcast proposal
	//wait for updates
	//append timeout

	return nil
}

func (c *Consensus) OnMsg(msg *Msg) {
	node, err := c.db.Node(msg.From)
	if err != nil {
		c.logger.WithError(err).Error("fetching node")
		return
	}
	if node == nil {
		c.logger.Error("unable to find node")
		return
	}

	signInfo, err := signatureData(msg)
	if err != nil {
		c.logger.WithError(err).Error("making msg signature data")
		return
	}

	hasValidSignature := false
	for _, k := range node.Keys {
		if err := Verify(signInfo, msg.Signature, k); err == nil {
			hasValidSignature = true
			break
		}
	}
	if !hasValidSignature {
		c.logger.Error("no valid signature for msg from peer")
		return
	}

	switch msg.Type {
	case MsgTypeConsensus:
		c.onConsensusMsg(msg.Consensus, msg.From)
	case MsgTypeTx:
		c.onTx(msg.Tx, msg.From)
	case MsgTypeBlock:
	}
}

func (c *Consensus) onConsensusMsg(msg *ConsensusMsg, from peer.ID) {
	switch msg.Type {
	case ConsensusMsgTypeNewRound:
		c.onNewRound(msg.NewRound, from)
	case ConsensusMsgTypeProposal:
		c.onProposal(msg.Proposal, from)
	case ConsensusMsgTypeVote:
		c.onVote(msg.Vote, from)
	case ConsensusMsgTypeBlock:
		c.onBlock(msg.Block, from)
	}
}

func (c *Consensus) onTx(msg *TxMsg, from peer.ID) {
	if err := c.memPool.AddTx(msg.Tx); err != nil {
		c.logger.WithError(err).Error("adding tx to mempool")
	}
}

func (c *Consensus) onVote(msg *ConsensusMsgVote, from peer.ID) {
	if from == c.currentProposer ||
		msg.Height != c.propsalState.Height ||
		msg.BlockID != c.propsalState.Block.String() ||
		msg.Round != c.propsalState.Round {
		return
	}

	switch msg.Type {
	case VoteTypePreVote:
		c.onPreVote(msg, from)
	case VoteTypePreCommit:
		c.OnPreCommit(msg, from)
	}
}

func (c *Consensus) onNewRound(msg *ConsensusMsgNewRound, from peer.ID) {
	if from != c.currentProposer {
		return
	}

	c.propsalState.lockedRound = -1
	c.propsalState.lockedValue = cid.Cid{}
	c.propsalState.AmProposer = false
	c.propsalState.Step = propose
	c.propsalState.Height = msg.Height
	c.propsalState.Round = msg.Round
	c.propsalState.Block = cid.Cid{}
	c.propsalState.PreVotes = make(map[peer.ID]*ConsensusMsgVote)
	c.propsalState.PreCommits = make(map[peer.ID]*ConsensusMsgVote)

	n, err := c.db.Nodes()
	if err != nil {
		c.logger.WithError(err).Error("getting nodes")
		return
	}
	c.propsalState.f = (uint64(len(n))/3)*2 + 1
}

func (c *Consensus) onProposal(msg *ConsensusMsgProposal, from peer.ID) {
	if from != c.currentProposer || c.propsalState.Step != propose {
		return
	}

	cid, err := cid.Parse(msg.BlockID)
	if err != nil {
		c.logger.WithError(err).Error("unable to parse CID")
		return
	}

	block, err := c.blockStore.getBlock(cid)
	if err != nil {
		c.logger.WithError(err).Error("getting block")
		return
	}

	valid := false
	if err := block.IsValid(c.blockStore); err != nil {
		c.logger.WithError(err).Error("invalid block")
	} else {
		valid = true
	}

	if !valid {
		if err := c.sendVote(VoteTypePreVote, ""); err != nil {
			c.logger.WithError(err).Error("sending nil prevote")
			return
		}
	} else {
		if err := c.sendVote(VoteTypePreVote, msg.BlockID); err != nil {
			c.logger.WithError(err).Error("sending prevote")
			return
		}
	}

	c.propsalState.Step = prevote

	//start timeouts
}

func (c *Consensus) onPreVote(msg *ConsensusMsgVote, from peer.ID) {
	if c.propsalState.Step != prevote {
		return
	}

	c.propsalState.PreVotes[from] = msg

	if uint64(len(c.propsalState.PreVotes)) > c.propsalState.f {
		id := c.propsalState.Block.String()
		if err := c.sendVote(VoteTypePreCommit, id); err != nil {
			c.logger.WithError(err).Error("sending precommit")
			return
		}
		c.propsalState.Step = precommit
	}
}

func (c *Consensus) OnPreCommit(msg *ConsensusMsgVote, from peer.ID) {
	if c.propsalState.Step != precommit {
		return
	}

	c.propsalState.PreCommits[from] = msg

	if c.propsalState.AmProposer && uint64(len(c.propsalState.PreCommits)) > c.propsalState.f {
		//send block msg
	}
}

func (c *Consensus) sendNewRound() error {
	msg := &ConsensusMsgNewRound{}

	return c.sendMsg(msg)
}

func (c *Consensus) sendVote(t VoteType, value string) error {
	msg := &ConsensusMsgVote{
		Type:      t,
		Height:    c.propsalState.Height,
		Round:     c.propsalState.Round,
		BlockID:   value,
		Timestamp: time.Now(),
		Validator: c.id.String(),
	}

	return c.sendMsg(msg)
}

func (c *Consensus) sendProposal(cid string) error {
	msg := &ConsensusMsgProposal{
		Height:    c.propsalState.Height,
		Round:     c.propsalState.Round,
		BlockID:   cid,
		Timestamp: time.Now(),
	}

	return c.sendMsg(msg)
}

func (c *Consensus) sendMsg(msg interface{}) error {
	wrapper := &ConsensusMsg{}

	switch msg.(type) {
	case *ConsensusMsgNewRound:
		wrapper.Type = ConsensusMsgTypeNewRound
		wrapper.NewRound = msg.(*ConsensusMsgNewRound)
	case *ConsensusMsgProposal:
		wrapper.Type = ConsensusMsgTypeProposal
		wrapper.Proposal = msg.(*ConsensusMsgProposal)
	case *ConsensusMsgVote:
		wrapper.Type = ConsensusMsgTypeVote
		wrapper.Vote = msg.(*ConsensusMsgVote)
	case *ConsensusMsgBlock:
		wrapper.Type = ConsensusMsgTypeBlock
		wrapper.Block = msg.(*ConsensusMsgBlock)
	}

	hl := &Msg{
		Type:      MsgTypeConsensus,
		From:      c.id,
		Consensus: wrapper,
		Timestamp: time.Now(),
	}

	return c.p2p.PublishContext(context.Background(), pubsubMsgsChanName, hl)
}

func (c *Consensus) onBlock(msg *ConsensusMsgBlock, from peer.ID) {
	if from != c.currentProposer ||
		c.propsalState.Step != precommit ||
		msg.CID != c.propsalState.lockedValue.String() ||
		c.propsalState.lockedRound != int32(msg.Round) {
		return
	}

	c.state.Height = msg.Height
	c.state.ParentBlock = c.state.Block
	c.state.Block, _ = cid.Parse(msg.CID)
}
