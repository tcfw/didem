package consensus

import (
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/tcfw/didem/internal/utils/resync"
)

type Step uint8

const (
	propose Step = iota
	prevote
	precommit
	block
)

type State struct {
	AmProposer  bool
	Proposer    string
	Height      uint64
	Round       uint32
	Step        Step
	Block       cid.Cid
	ParentBlock cid.Cid

	f  uint64
	mu sync.Mutex

	voteMu     sync.Mutex
	PreVotes   map[peer.ID]*Msg
	PreCommits map[peer.ID]*Msg

	PreVotesEvidence   map[string]*ConsensusMsgEvidence
	PreCommitsEvidence map[string]*ConsensusMsgEvidence

	lockedValue cid.Cid
	lockedRound uint32

	pvOnce resync.Once
	pcOnce resync.Once
	sbOnce resync.Once
}

func (s *State) setStep(st Step) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Step = st
}

func (s *State) resetOnces() {
	s.pvOnce.Reset()
	s.pcOnce.Reset()
	s.sbOnce.Reset()
}
