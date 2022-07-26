package consensus

import (
	"context"
	"math/big"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	pubsubMsgsChanName = "/did/msgs"
	pubsubTxChanName   = "/did/tx"

	timeoutPropose   = 1 * time.Minute
	timeoutPrevote   = 1 * time.Minute
	timeoutPrecommit = 1 * time.Minute
)

type Consensus struct {
	id     peer.ID
	priv   crypto.PrivKey
	logger *logrus.Entry

	db         Db
	memPool    MemPool
	blockStore BlockStore

	beacon <-chan int64
	p2p    *p2p

	state        State
	propsalState State

	timerPropose   *time.Timer
	timerPrevote   *time.Timer
	timerPrecommit *time.Timer
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
	c.timerPropose = time.NewTimer(1 * time.Minute)
	if !c.timerPropose.Stop() {
		<-c.timerPropose.C //clear timer
	}
	c.timerPrevote = time.NewTimer(1 * time.Minute)
	if !c.timerPrevote.Stop() {
		<-c.timerPrevote.C //clear timer
	}
	c.timerPrecommit = time.NewTimer(1 * time.Minute)
	if !c.timerPrecommit.Stop() {
		<-c.timerPrecommit.C //clear timer
	}

	go c.watchProposer()
	go c.watchTimeouts()

	if err := c.subscribeMsgs(); err != nil {
		return errors.Wrap(err, "watching msgs")
	}

	return c.subscribeTx()
}

func (c *Consensus) watchTimeouts() {
	for {
		select {
		case <-c.timerPropose.C:
			c.onTimeoutProposal()
		case <-c.timerPrevote.C:
			c.onTimeoutPrevote()
		case <-c.timerPrecommit.C:
			c.onTimeoutPrecommit()
		}
	}
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
		c.propsalState.Proposer = id

		c.StartRound(false)
	}
}

func (c *Consensus) StartRound(inc bool) error {
	c.propsalState.AmProposer = c.propsalState.Proposer == c.id
	c.propsalState.lockedRound = 0
	c.propsalState.lockedValue = cid.Undef
	c.propsalState.Step = propose
	c.propsalState.Height = c.state.Height + 1
	if inc {
		c.propsalState.Round++
	} else {
		c.propsalState.Round = 1
		c.propsalState.Block = cid.Undef
	}
	c.propsalState.PreVotes = make(map[peer.ID]*ConsensusMsgVote)
	c.propsalState.PreCommits = make(map[peer.ID]*ConsensusMsgVote)

	n, err := c.db.Nodes()
	if err != nil {
		return errors.Wrap(err, "getting nodes")
	}
	c.propsalState.f = (uint64(len(n))/3)*2 + 1

	if !c.propsalState.AmProposer {
		c.timerPropose.Reset(timeoutPropose)
		return nil
	}

	if err := c.sendNewRound(); err != nil {
		return errors.Wrap(err, "sending new round")
	}

	//build block
	//upload block

	c.propsalState.Block = cid.Undef

	//broadcast proposal
	if err := c.sendProposal(); err != nil {
		return errors.Wrap(err, "sending proposal")
	}

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

	signData, err := signatureData(msg)
	if err != nil {
		c.logger.WithError(err).Error("making msg signature data")
		return
	}

	hasValidSignature := false
	for _, k := range node.Keys {
		if err := Verify(signData, msg.Signature, k); err == nil {
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
	if from == c.propsalState.Proposer ||
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
	if from != c.propsalState.Proposer {
		return
	}

	c.propsalState.lockedRound = 0
	c.propsalState.lockedValue = cid.Undef
	c.propsalState.AmProposer = false
	c.propsalState.Step = propose
	c.propsalState.Height = msg.Height
	c.propsalState.Round = msg.Round
	c.propsalState.Block = cid.Undef
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
	if from != c.propsalState.Proposer || c.propsalState.Step != propose {
		return
	}

	bc, err := cid.Parse(msg.BlockID)
	if err != nil {
		c.logger.WithError(err).Error("unable to parse CID")
		return
	}

	c.propsalState.Block = bc

	block, err := c.blockStore.getBlock(bc)
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
		c.propsalState.lockedValue = c.propsalState.Block
		c.propsalState.lockedRound = c.propsalState.Round

		c.timerPrevote.Reset(timeoutPrevote)
	}
}

func (c *Consensus) OnPreCommit(msg *ConsensusMsgVote, from peer.ID) {
	if c.propsalState.Step != precommit {
		return
	}

	c.propsalState.PreCommits[from] = msg

	if uint64(len(c.propsalState.PreCommits)) > c.propsalState.f {
		if c.propsalState.AmProposer {
			//send block msg
		} else {
			//expect block msg
			c.timerPrecommit.Reset(timeoutPrecommit)
		}
	}
}

func (c *Consensus) onTimeoutProposal() {
	if c.propsalState.Height != c.state.Height &&
		c.propsalState.Step != propose {
		return
	}

	c.propsalState.Step = prevote
	c.sendVote(VoteTypePreVote, "")
}

func (c *Consensus) onTimeoutPrevote() {
	if c.propsalState.Height != c.state.Height &&
		c.propsalState.Step != prevote {
		return
	}

	c.propsalState.Step = precommit
	c.sendVote(VoteTypePreCommit, "")
}

func (c *Consensus) onTimeoutPrecommit() {
	c.StartRound(true)
}

func (c *Consensus) sendNewRound() error {
	msg := &ConsensusMsgNewRound{
		Height:          c.propsalState.Height,
		Round:           c.propsalState.Round,
		LastCommitRound: c.state.Round,
		Timestamp:       time.Now(),
	}

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

func (c *Consensus) sendProposal() error {
	msg := &ConsensusMsgProposal{
		Height:    c.propsalState.Height,
		Round:     c.propsalState.Round,
		BlockID:   c.propsalState.Block.String(),
		Timestamp: time.Now(),
	}

	return c.sendMsg(msg)
}

func (c *Consensus) sendMsg(msg interface{}) error {
	wrapper := &ConsensusMsg{}

	switch msg := msg.(type) {
	case *ConsensusMsgNewRound:
		wrapper.Type = ConsensusMsgTypeNewRound
		wrapper.NewRound = msg
	case *ConsensusMsgProposal:
		wrapper.Type = ConsensusMsgTypeProposal
		wrapper.Proposal = msg
	case *ConsensusMsgVote:
		wrapper.Type = ConsensusMsgTypeVote
		wrapper.Vote = msg
	case *ConsensusMsgBlock:
		wrapper.Type = ConsensusMsgTypeBlock
		wrapper.Block = msg
	}

	hl := &Msg{
		Type:      MsgTypeConsensus,
		From:      c.id,
		Consensus: wrapper,
		Timestamp: time.Now(),
	}

	signData, err := signatureData(hl)
	if err != nil {
		return errors.Wrap(err, "unable to create sign data")
	}

	sig, err := c.priv.Sign(signData)
	if err != nil {
		return errors.Wrap(err, "unable to sign")
	}
	hl.Signature = sig

	return c.p2p.PublishContext(context.Background(), pubsubMsgsChanName, hl)
}

func (c *Consensus) onBlock(msg *ConsensusMsgBlock, from peer.ID) {
	if from != c.propsalState.Proposer ||
		c.propsalState.Step != precommit ||
		msg.CID != c.propsalState.lockedValue.String() ||
		c.propsalState.lockedRound != msg.Round {
		return
	}

	c.state.Height = c.propsalState.Height
	c.state.ParentBlock = *c.state.Block
	c.state.Block = c.propsalState.lockedValue
}
