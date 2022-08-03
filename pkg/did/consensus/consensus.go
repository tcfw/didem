package consensus

import (
	"bytes"
	"context"
	"math/big"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tcfw/didem/internal/utils/logging"
	"github.com/tcfw/didem/pkg/storage"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign/bls"
	"go.dedis.ch/kyber/v3/util/random"
)

const (
	pubsubMsgsChanName = "/did/msgs"
	pubsubTxChanName   = "/did/tx"
)

var (
	timeoutPropose   = 1 * time.Minute
	timeoutPrevote   = 1 * time.Minute
	timeoutPrecommit = 1 * time.Minute
	timeoutBlock     = 1 * time.Minute
)

type Consensus struct {
	chain      []byte
	id         peer.ID
	signingKey kyber.Scalar

	db         Db
	memPool    MemPool
	blockStore storage.Store
	validator  storage.Validator

	beacon <-chan int64
	p2p    *p2p

	state        State
	propsalState State

	timerPropose   *time.Timer
	timerPrevote   *time.Timer
	timerPrecommit *time.Timer
	timerBlock     *time.Timer
	stopTimerLoop  chan struct{}
}

func NewConsensus(h host.Host, p *pubsub.PubSub, opts ...Option) (*Consensus, error) {
	c := &Consensus{
		id:      h.ID(),
		p2p:     newP2P(h.ID(), p),
		memPool: NewTxMemPool(),
	}

	c.setupTimers()

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Consensus) Start() error {
	c.setupTimers()

	go c.watchProposer()
	go c.watchTimeouts()

	if err := c.subscribeMsgs(); err != nil {
		return errors.Wrap(err, "watching msgs")
	}

	return c.subscribeTx()
}

func (c *Consensus) proposer() <-chan peer.ID {
	bCh := make(chan peer.ID)

	go func() {
		for b := range c.beacon {
			nodes, err := c.db.Nodes()
			if err != nil {
				logging.WithError(err).Error("getting node list")
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

func (c *Consensus) setupTimers() {
	c.timerPropose = time.NewTimer(1 * time.Minute)
	stopTimer(c.timerPropose)

	c.timerPrevote = time.NewTimer(1 * time.Minute)
	stopTimer(c.timerPrevote)

	c.timerPrecommit = time.NewTimer(1 * time.Minute)
	stopTimer(c.timerPrecommit)

	c.timerBlock = time.NewTimer(1 * time.Minute)
	stopTimer(c.timerBlock)

	c.stopTimerLoop = make(chan struct{})
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
		case <-c.timerBlock.C:
			c.onTimeoutBlock()
		case <-c.stopTimerLoop:
			return
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
			c.OnMsg(msg)
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
	c.propsalState.PreVotes = make(map[peer.ID]*Msg)
	c.propsalState.PreCommits = make(map[peer.ID]*Msg)
	c.propsalState.PreVotesEvidence = make(map[peer.ID]*ConsensusMsgEvidence)
	c.propsalState.PreCommitsEvidence = make(map[peer.ID]*ConsensusMsgEvidence)

	if inc {
		c.propsalState.Round++
		c.propsalState.Step = prevote
	} else {
		c.propsalState.Round = 1
		c.propsalState.Block = cid.Undef
	}

	logging.Entry().WithFields(logrus.Fields{
		"Height":   c.propsalState.Height,
		"Round":    c.propsalState.Round,
		"Proposer": c.propsalState.AmProposer,
	}).Info("starting a new round")

	if !c.propsalState.AmProposer {
		if c.propsalState.Block == cid.Undef && c.propsalState.Round > 1 {
			//still waiting for a proposal
			c.propsalState.Step = propose
		}
		restartTimer(c.timerPropose, timeoutPropose)
		return nil
	}

	restartTimer(c.timerPrevote, timeoutPrevote)

	if err := c.sendNewRound(); err != nil {
		return errors.Wrap(err, "sending new round")
	}

	n, err := c.db.Nodes()
	if err != nil {
		return errors.Wrap(err, "getting nodes")
	}

	c.propsalState.f = (uint64(len(n))/3)*2 + 1

	//build & upload block
	if c.propsalState.Block == cid.Undef {
		block, err := c.makeBlock()
		if err != nil {
			return errors.Wrap(err, "making new block")
		}

		if block != nil {
			c.propsalState.Block, err = c.blockStore.PutBlock(context.Background(), block)
			if err != nil {
				return errors.Wrap(err, "storing new block")
			}
		}
	}

	if err := c.sendProposal(); err != nil {
		return errors.Wrap(err, "sending proposal")
	}

	return nil
}

func (c *Consensus) OnMsg(msg *Msg) {
	if !bytes.Equal(msg.Chain, c.chain) {
		logging.Error("received msg from different chain")
		return
	}

	logging.Entry().WithField("msg", msg).Info("received msg")

	node, err := c.db.Node(msg.From)
	if err != nil {
		logging.WithError(err).Error("fetching node")
		return
	}
	if node == nil {
		logging.Error("unable to find node")
		return
	}

	sd, err := msg.From.MarshalBinary()
	if err != nil {
		logging.WithError(err).Error("marshalling id")
		return
	}

	signData, err := signatureData(msg)
	if err != nil {
		logging.WithError(err).Error("making msg signature data")
		return
	}

	sd = append(sd, signData...)

	hasValidSignature := false
	for _, k := range node.Keys {
		_, pub := bls.NewKeyPair(bn256.NewSuite(), random.New())
		if err := pub.UnmarshalBinary(k); err != nil {
			logging.WithError(err).Error("unmarshaling key")
			return
		}
		if err := bls.Verify(bn256.NewSuite(), pub, sd, msg.Signature); err == nil {
			hasValidSignature = true
			break
		}
	}
	if !hasValidSignature {
		logging.Error("no valid signature for msg from peer")
		return
	}

	switch msg.Type {
	case MsgTypeConsensus:
		c.onConsensusMsg(msg, msg.From)
	case MsgTypeTx:
		go c.onTx(msg.Tx, msg.From)
	case MsgTypeBlock:
	}
}

func (c *Consensus) onConsensusMsg(msg *Msg, from peer.ID) {
	switch msg.Consensus.Type {
	case ConsensusMsgTypeNewRound:
		c.onNewRound(msg.Consensus.NewRound, from)
	case ConsensusMsgTypeProposal:
		c.onProposal(msg.Consensus.Proposal, from)
	case ConsensusMsgTypeVote:
		c.onVote(msg, from)
	case ConsensusMsgTypeBlock:
		c.onBlock(msg.Consensus.Block, from)
	case ConsensusMsgTypeEvidence:
		c.onEvidence(msg.Consensus.Evidence, from)
	}
}

func (c *Consensus) onTx(msg *TxMsg, from peer.ID) {
	if err := c.memPool.AddTx(msg.Tx, msg.TTL); err != nil {
		logging.WithError(err).Error("adding tx to mempool")
	}
}

func (c *Consensus) onVote(msg *Msg, from peer.ID) {
	vote := msg.Consensus.Vote

	if from == c.propsalState.Proposer {
		logging.Error("ignoring vote from proposer")
		return
	} else if vote.Height != c.propsalState.Height ||
		vote.BlockID != c.propsalState.Block.String() ||
		vote.Round != c.propsalState.Round {
		logging.Error("ignoring vote, invalid state")
		return
	}

	switch vote.Type {
	case VoteTypePreVote:
		c.onPreVote(msg)
	case VoteTypePreCommit:
		c.OnPreCommit(msg)
	}

	if c.propsalState.AmProposer {
		if err := c.sendEvidence(vote); err != nil {
			logging.WithError(err).Error("sending evidence")
		}
	}
}

func (c *Consensus) onNewRound(msg *ConsensusMsgNewRound, from peer.ID) {
	if from != c.propsalState.Proposer {
		logging.Error("ignorining new round not from current proposer")
		return
	}

	logging.Entry().WithFields(logrus.Fields{
		"Height": msg.Height,
		"Round":  msg.Round,
	}).Info("new round starting")

	c.propsalState.lockedRound = 0
	c.propsalState.lockedValue = cid.Undef
	c.propsalState.Step = propose
	c.propsalState.Height = msg.Height
	c.propsalState.Round = msg.Round
	c.propsalState.PreVotes = make(map[peer.ID]*Msg)
	c.propsalState.PreCommits = make(map[peer.ID]*Msg)
	c.propsalState.PreVotesEvidence = make(map[peer.ID]*ConsensusMsgEvidence)
	c.propsalState.PreCommitsEvidence = make(map[peer.ID]*ConsensusMsgEvidence)

	restartTimer(c.timerPropose, timeoutPropose)

	logging.Entry().WithFields(logrus.Fields{
		"Height": msg.Height,
		"Round":  msg.Round,
	}).Info("waiting for proposal to start")
}

func (c *Consensus) onProposal(msg *ConsensusMsgProposal, from peer.ID) {
	if from != c.propsalState.Proposer {
		logging.Error("ignorining proposal not from current proposer")
		return
	} else if c.propsalState.Step != propose {
		logging.Error("ignorining proposal, wrong step")
		return
	}

	logging.Entry().WithFields(logrus.Fields{
		"Height": msg.Height,
		"Round":  msg.Round,
		"Value":  msg.BlockID,
	}).Info("new proposal processing")

	bc, err := c.validate(msg.BlockID)
	if err != nil {
		logging.WithError(err).Error("validating block")
	}

	c.propsalState.Block = bc

	stopTimer(c.timerPropose)

	if bc == cid.Undef {
		if err := c.sendVote(VoteTypePreVote, ""); err != nil {
			logging.WithError(err).Error("sending nil prevote")
			return
		}
	} else {
		if err := c.sendVote(VoteTypePreVote, msg.BlockID); err != nil {
			logging.WithError(err).Error("sending prevote")
			return
		}
	}

	restartTimer(c.timerPrevote, timeoutPrevote)
	c.propsalState.Step = prevote
}

func (c *Consensus) validate(value string) (cid.Cid, error) {
	cv, err := cid.Parse(value)
	if err != nil {
		return cid.Undef, errors.Wrap(err, "unable to parse CID")
	}

	block, err := c.blockStore.GetBlock(context.Background(), storage.BlockID(cv))
	if err != nil {
		return cid.Undef, errors.Wrap(err, "getting block")
	}

	if err := c.validator.IsValid(block); err != nil {
		return cid.Undef, errors.Wrap(err, "invalid block")
	}

	return cv, nil
}

func (c *Consensus) onPreVote(msg *Msg) {
	if c.propsalState.Step != prevote {
		return
	}

	if c.propsalState.AmProposer {
		restartTimer(c.timerPrecommit, timeoutPrecommit)
	}

	c.propsalState.PreVotes[msg.From] = msg

	if uint64(len(c.propsalState.PreVotes)) >= c.propsalState.f {
		stopTimer(c.timerPrevote)

		id := c.propsalState.Block.String()
		if err := c.sendVote(VoteTypePreCommit, id); err != nil {
			logging.WithError(err).Error("sending precommit")
			return
		}

		c.propsalState.Step = precommit
		c.propsalState.lockedValue = c.propsalState.Block
		c.propsalState.lockedRound = c.propsalState.Round

		restartTimer(c.timerPrecommit, timeoutPrecommit)
	}
}

func (c *Consensus) OnPreCommit(msg *Msg) {
	if c.propsalState.Step != precommit {
		return
	}

	if c.propsalState.AmProposer {
		restartTimer(c.timerPrecommit, timeoutPrecommit)
	}

	c.propsalState.PreVotes[msg.From] = msg

	if uint64(len(c.propsalState.PreCommits)) >= c.propsalState.f {
		stopTimer(c.timerPrecommit)

		if c.propsalState.AmProposer {
			msg, err := c.sendBlock()
			if err != nil {
				logging.WithError(err).Error("failed to send blockmsg")
				return
			}

			c.onBlock(msg, c.id)
		} else {
			restartTimer(c.timerBlock, timeoutBlock)
		}
	}
}

func (c *Consensus) onEvidence(msg *ConsensusMsgEvidence, from peer.ID) {
	if from != c.propsalState.Proposer {
		return
	}

	switch msg.Type {
	case VoteTypePreVote:
		c.propsalState.PreVotesEvidence[from] = msg
	case VoteTypePreCommit:
		c.propsalState.PreCommitsEvidence[from] = msg
	}
}

func (c *Consensus) onBlock(msg *ConsensusMsgBlock, from peer.ID) {
	if from != c.propsalState.Proposer ||
		c.propsalState.Step != precommit ||
		msg.CID != c.propsalState.lockedValue.String() ||
		c.propsalState.lockedRound != msg.Round {
		return
	}

	stopTimer(c.timerBlock)

	c.state.Height = c.propsalState.Height
	c.state.ParentBlock = c.state.Block
	c.state.Block = c.propsalState.lockedValue
}

func (c *Consensus) onTimeoutProposal() {
	if c.propsalState.Height != c.state.Height &&
		c.propsalState.Step != propose {
		return
	}

	logging.Entry().WithFields(logrus.Fields{
		"Height": c.propsalState.Height,
		"Round":  c.propsalState.Round,
	}).Info("timing out proposal")

	c.propsalState.Step = prevote
	c.sendVote(VoteTypePreVote, "")

	restartTimer(c.timerPrevote, timeoutPrevote)
}

func (c *Consensus) onTimeoutPrevote() {
	if c.propsalState.Height != c.state.Height &&
		c.propsalState.Step != prevote {
		return
	}

	logging.Entry().WithFields(logrus.Fields{
		"Height": c.propsalState.Height,
		"Round":  c.propsalState.Round,
	}).Info("timing out prevote")

	c.propsalState.Step = precommit
	c.sendVote(VoteTypePreCommit, "")

	restartTimer(c.timerPrecommit, timeoutPrecommit)
}

func (c *Consensus) onTimeoutPrecommit() {
	c.StartRound(true)
}

func (c *Consensus) onTimeoutBlock() {
	if c.propsalState.Height != c.state.Height &&
		c.propsalState.Step != precommit {
		return
	}

	logging.Entry().WithFields(logrus.Fields{
		"Height": c.propsalState.Height,
		"Round":  c.propsalState.Round,
	}).Info("timing out block")

	//TODO(tcfw): what to do waiting for signatures?
}

func (c *Consensus) sendEvidence(m *ConsensusMsgVote) error {
	msg := &ConsensusMsgEvidence{
		ConsensusMsgVote: *m,
		Accepted:         true,
	}

	return c.sendMsg(msg)
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

func (c *Consensus) sendBlock() (*ConsensusMsgBlock, error) {
	msg := &ConsensusMsgBlock{
		Height: c.propsalState.Height,
		Round:  c.propsalState.lockedRound,
		CID:    c.propsalState.lockedValue.String(),
	}

	pks := make([]kyber.Point, 0, len(c.propsalState.PreCommits))
	sigs := make([][]byte, 0, len(c.propsalState.PreCommits))

	for p, vote := range c.propsalState.PreCommits {
		n, err := c.db.Node(p)
		if err != nil {
			return nil, errors.Wrap(err, "finding node")
		}
		for _, k := range n.Keys {
			if k.String() == vote.Consensus.Vote.Validator {
				pks = append(pks, k)
			}
		}

		sigs = append(sigs, vote.Signature)
	}
	k := bls.AggregatePublicKeys(bn256.NewSuite(), pks...)
	v, err := k.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "aggregating public keys")
	}
	msg.Validators = v

	s, err := bls.AggregateSignatures(bn256.NewSuite(), sigs...)
	if err != nil {
		return nil, errors.Wrap(err, "aggregating signatures")
	}
	msg.Signature = s

	return msg, c.sendMsg(msg)
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
	case *ConsensusMsgEvidence:
		wrapper.Type = ConsensusMsgTypeEvidence
		wrapper.Evidence = msg
	case *ConsensusMsgBlock:
		wrapper.Type = ConsensusMsgTypeBlock
		wrapper.Block = msg
	}

	hl := &Msg{
		Type:      MsgTypeConsensus,
		From:      c.id,
		Chain:     c.chain,
		Consensus: wrapper,
		Timestamp: time.Now(),
	}

	d, err := c.id.MarshalBinary()
	if err != nil {
		return errors.Wrap(err, "marshaling id")
	}

	signData, err := signatureData(hl)
	if err != nil {
		return errors.Wrap(err, "unable to create sign data")
	}

	d = append(d, signData...)

	sig, err := bls.Sign(bn256.NewSuite(), c.signingKey, d)
	if err != nil {
		return errors.Wrap(err, "signing vote data")
	}

	hl.Signature = sig

	logging.Entry().WithField("msg", hl).Info("sending msg")

	return c.p2p.PublishContext(context.Background(), pubsubMsgsChanName, hl)
}

func restartTimer(t *time.Timer, d time.Duration) {
	stopTimer(t)
	t.Reset(d)
}

func stopTimer(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}
