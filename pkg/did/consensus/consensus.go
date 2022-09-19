package consensus

import (
	"bytes"
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tcfw/didem/internal/utils/logging"
	"github.com/tcfw/didem/pkg/cryptography"
	"github.com/tcfw/didem/pkg/storage"
)

const (
	pubsubMsgsChanName = "/did/msgs"
	pubsubTxChanName   = "/did/tx"
)

var (
	timeoutPropose       = 1 * time.Minute
	timeoutPrevote       = 1 * time.Minute
	timeoutPrecommit     = 1 * time.Minute
	timeoutBlock         = 1 * time.Minute
	blockVoteGracePeriod = 10 * time.Second
)

type Consensus struct {
	chain      []byte
	id         peer.ID
	signingKey *cryptography.Bls12381PrivateKey

	genesis *storage.GenesisInfo

	memPool   MemPool
	store     storage.Store
	validator storage.Validator

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

// NewConsensus initiates a new consensus engine by checking if the genesis has been
// fully applied to the metadata store, setting up the round timeouts and restores
// the current consensus state from the metadata store
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

	if c.genesis != nil && !c.store.HasGenesisApplied() {
		if err := c.store.ApplyGenesis(c.genesis); err != nil {
			return nil, errors.Wrap(err, "applying genesis")
		}
	}

	b, err := c.store.GetLastApplied(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "fetching last applied block")
	}
	if b != nil {
		logging.Entry().WithFields(logrus.Fields{
			"Height": b.Height,
			"Parent": b.Parent.String(),
			"Block":  b.ID.String(),
		}).Info("loaded last block as")

		c.state.Block = cid.Cid(b.ID)
		c.state.Height = b.Height
		c.state.ParentBlock = cid.Cid(b.Parent)
	}

	return c, nil
}

func (c *Consensus) Validator() storage.Validator {
	return c.validator
}

// Start initiates watching for new proposers and timeouts as well as subscribing
// to new messages from other nodes
func (c *Consensus) Start() error {
	c.setupTimers()

	go c.watchProposer()
	go c.watchTimeouts()

	if err := c.subscribeMsgs(); err != nil {
		return errors.Wrap(err, "watching msgs")
	}

	return c.subscribeTx()
}

func (c *Consensus) State() *State {
	return &c.state
}

func (c *Consensus) ChainID() string {
	return string(c.chain)
}

// proposer watches for new propsers from the given randomness beacon
func (c *Consensus) proposer() <-chan peer.ID {
	bCh := make(chan peer.ID)

	go func() {
		for b := range c.beacon {
			nodes, err := c.store.Nodes()
			if err != nil {
				logging.WithError(err).Error("getting node list")
				continue
			}

			m := big.NewInt(0)
			m.SetInt64(b)
			mn := m.Mod(m, big.NewInt(int64(len(nodes)))).Int64()

			p := nodes[mn]
			bCh <- peer.ID(p)
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

		if err := c.StartRound(false); err != nil {
			logging.WithError(err).Error("starting round")
		}
	}
}

// StartRound begins a new round of consensus from other nodes, assuming a new height
// and clearing out all current votes and evidence. Start round also sets the required
// number of votes required from the round by inspecting the current number of nodes
// still active in the metadata store. If the current proposer is the current node
// the node will attempt to create a new proposal/block and distribute to all other
// nodes
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
	c.propsalState.resetOnces()

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

	n, err := c.store.Nodes()
	if err != nil {
		return errors.Wrap(err, "getting nodes")
	}

	for i, ni := range n {
		if ni == c.id.String() {
			n[i] = n[len(n)-1]
			n = n[:len(n)-1]
		}
	}

	c.propsalState.f = ((uint64(len(n)) / 3) * 2) + 1

	//build & upload block
	if c.propsalState.Block.Equals(cid.Undef) {
		block, err := c.makeBlock()
		if err != nil {
			return errors.Wrap(err, "making new block")
		}

		if block != nil {
			c.propsalState.Block, err = c.store.PutBlock(context.Background(), block)
			if err != nil {
				return errors.Wrap(err, "storing new block")
			}
		}
	}

	if err := c.sendProposal(); err != nil {
		return errors.Wrap(err, "sending proposal")
	}

	c.propsalState.Step = prevote

	return nil
}

// OnMsg takes in a new message from another node, validates it's signature, and
// forwards the message onto the specific message type handlers
func (c *Consensus) OnMsg(msg *Msg) {
	if !bytes.Equal(msg.Chain, c.chain) {
		logging.Error("received msg from different chain")
		return
	}

	logging.Entry().WithField("msg", msg).Info("received msg")

	node, err := c.store.Node(context.Background(), msg.From.String())
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
		logging.WithError(err).Error("marshaling id")
		return
	}

	signData, err := signatureData(msg)
	if err != nil {
		logging.WithError(err).Error("making msg signature data")
		return
	}

	sd = append(sd, signData...)

	pk, err := cryptography.NewBls12381PublicKey(node.Key)
	if err != nil {
		logging.WithError(err).Error("unmarshalling node pk")
	}

	if ok, err := pk.Verify(msg.Signature, sd); !ok || err != nil {
		logging.Error("invalid signature")
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

// onVote handles any new vote from another node, filtering out if the
// current node is not the current propser. If the propser is the current
// node, the propser sends back evidence of acceptance of the nodes vote
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

	c.propsalState.voteMu.Lock()
	defer c.propsalState.voteMu.Unlock()

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

// onNewRound picks up on the propser starting a new round of consensus and populates
// the propsal state based of the proposers future round values
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
	c.propsalState.pvOnce = sync.Once{}
	c.propsalState.pcOnce = sync.Once{}
	c.propsalState.sbOnce = sync.Once{}

	restartTimer(c.timerPropose, timeoutPropose)

	logging.Entry().WithFields(logrus.Fields{
		"Height": msg.Height,
		"Round":  msg.Round,
	}).Info("waiting for proposal to start")
}

// onProposal takes in the block proposal from the current proposer and
// validates the propsed block. If the block is found to be valid, a vote
// is sent
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

	c.propsalState.pvOnce.Do(func() {
		c.propsalState.Block = bc

		stopTimer(c.timerPropose)

		if bc == cid.Undef {
			if err := c.sendVote(VoteTypePreVote, cid.Undef.String()); err != nil {
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
	})
}

func (c *Consensus) validate(value string) (cid.Cid, error) {
	if value == cid.Undef.String() {
		return cid.Undef, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cv, err := cid.Parse(value)
	if err != nil {
		return cid.Undef, errors.Wrap(err, "unable to parse CID")
	}

	blockID := storage.BlockID(cv)

	block, err := c.store.GetBlock(ctx, blockID)
	if err != nil {
		return cid.Undef, errors.Wrap(err, "getting block")
	}

	if err := c.validator.IsBlockValid(ctx, block, true); err != nil {
		return cid.Undef, errors.Wrap(err, "invalid block")
	}

	return cv, nil
}

// onPreVote collects prevotes from other nodes. If the number of votes is above
// the required threshold, the round state is progressed to precommit
func (c *Consensus) onPreVote(msg *Msg) {
	if c.propsalState.Step != prevote {
		return
	}

	if c.propsalState.AmProposer {
		restartTimer(c.timerPrecommit, timeoutPrecommit)
	}

	c.propsalState.PreVotes[msg.From] = msg

	if uint64(len(c.propsalState.PreVotes)) > c.propsalState.f {
		c.propsalState.pcOnce.Do(func() {
			stopTimer(c.timerPrevote)

			c.propsalState.Step = precommit
			c.propsalState.lockedValue = c.propsalState.Block
			c.propsalState.lockedRound = c.propsalState.Round

			id := c.propsalState.Block.String()
			if err := c.sendVote(VoteTypePreCommit, id); err != nil {
				logging.WithError(err).Error("sending precommit")
				return
			}

			restartTimer(c.timerPrecommit, timeoutPrecommit)
		})
	}
}

// OnPreCommit collects precommit votes from other nodes. If the number of precommit
// votes collected is above the required threshold, the consensus state is progressed
// to the block state
func (c *Consensus) OnPreCommit(msg *Msg) {
	if c.propsalState.Step != precommit {
		return
	}

	if c.propsalState.AmProposer {
		restartTimer(c.timerPrecommit, timeoutPrecommit)
	}

	c.propsalState.PreCommits[msg.From] = msg

	if uint64(len(c.propsalState.PreCommits)) > c.propsalState.f {
		c.propsalState.sbOnce.Do(func() {
			stopTimer(c.timerPrecommit)

			c.propsalState.Step = block

			if c.propsalState.AmProposer {
				//Give some time for other nodes to collect the evidence
				time.AfterFunc(blockVoteGracePeriod, func() {
					msg, err := c.sendBlock()
					if err != nil {
						logging.WithError(err).Error("failed to send blockmsg")
						return
					}

					c.onBlock(msg, c.id)
				})
			} else {
				restartTimer(c.timerBlock, timeoutBlock)
			}
		})
	}
}

// onEvidence collects evidence from the proposer node
func (c *Consensus) onEvidence(msg *ConsensusMsgEvidence, from peer.ID) {
	if from != c.propsalState.Proposer {
		return
	}

	c.propsalState.voteMu.Lock()
	defer c.propsalState.voteMu.Unlock()

	switch msg.Type {
	case VoteTypePreVote:
		c.propsalState.PreVotesEvidence[from] = msg
	case VoteTypePreCommit:
		c.propsalState.PreCommitsEvidence[from] = msg
	}
}

// onBlock collects block completion messages sent by the propser and
// updates the final current consensus state
func (c *Consensus) onBlock(msg *ConsensusMsgBlock, from peer.ID) {
	if from != c.propsalState.Proposer ||
		c.propsalState.Step != block ||
		msg.CID != c.propsalState.lockedValue.String() ||
		c.propsalState.lockedRound != msg.Round {
		return
	}

	if !c.propsalState.lockedValue.Equals(cid.Undef) {
		if err := c.store.MarkBlock(context.Background(), storage.BlockID(c.propsalState.lockedValue), storage.BlockStateAccepted); err != nil {
			logging.WithError(err).Error("marking block as accepted")
		}
	}

	stopTimer(c.timerBlock)

	//TODO(tcfw) validate the final signature from the propser

	if err := c.store.UpdateLastApplied(context.Background(), storage.BlockID(c.propsalState.lockedValue)); err != nil {
		logging.WithError(err).Error("marking block as latest")
	}

	c.state.Height = c.propsalState.Height
	c.state.ParentBlock = c.state.Block
	c.state.Block = c.propsalState.lockedValue
}

// onTimeoutProposal sends an empty prevote when the propser node doesn't
// send a valid proposal message in the alloted time
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
	c.propsalState.pvOnce.Do(func() {
		if err := c.sendVote(VoteTypePreVote, ""); err != nil {
			logging.Error(err)
		}

		restartTimer(c.timerPrevote, timeoutPrevote)
	})
}

// onTimeoutPrevote sents an empty precommit when the propser node doesn't
// send a valid prevote evidence message
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

	c.propsalState.pcOnce.Do(func() {
		if err := c.sendVote(VoteTypePreCommit, ""); err != nil {
			logging.Error(err)
		}

		restartTimer(c.timerPrecommit, timeoutPrecommit)
	})
}

// onTimeoutPrecommit assumes a new round is to be started when the
// propser does not send precommit evidence for the current round
func (c *Consensus) onTimeoutPrecommit() {
	if err := c.StartRound(true); err != nil {
		logging.WithError(err).Error("starting new round after timeout")
	}
}

// onTimeoutBlock assumes the propser failed to create a block/signature
// confirmation message for the currnet round and the round has failed
func (c *Consensus) onTimeoutBlock() {
	if c.propsalState.Height != c.state.Height &&
		c.propsalState.Step != precommit {
		return
	}

	logging.Entry().WithFields(logrus.Fields{
		"Height": c.propsalState.Height,
		"Round":  c.propsalState.Round,
	}).Info("timing out block")

	//TODO(tcfw): when waiting for signatures?
}

func (c *Consensus) sendEvidence(m *ConsensusMsgVote) error {
	msg := &ConsensusMsgEvidence{
		ConsensusMsgVote: *m,
		Accepted:         true,
	}

	switch m.Type {
	case VoteTypePreVote:
		c.propsalState.PreVotesEvidence[peer.ID(m.Validator)] = msg
	case VoteTypePreCommit:
		c.propsalState.PreCommitsEvidence[peer.ID(m.Validator)] = msg
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

	pks := make([]*cryptography.Bls12381PublicKey, 0, len(c.propsalState.PreCommits))
	sigs := make([][]byte, 0, len(c.propsalState.PreCommits))

	for p, vote := range c.propsalState.PreCommits {
		n, err := c.store.Node(context.Background(), p.String())
		if err != nil {
			return nil, errors.Wrap(err, "finding node")
		}

		pk, err := cryptography.NewBls12381PublicKey(n.Key)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshalling node key")
		}

		pks = append(pks, pk)

		sigs = append(sigs, vote.Signature)
	}

	k := AggregatePublicKeys(pks...)
	v, err := k.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "aggregating public keys")
	}
	msg.Validators = v

	s, err := AggregateSignatures(sigs...)
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

	sig, err := c.signingKey.Sign(nil, d, nil)
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
