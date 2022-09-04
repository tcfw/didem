package consensus

import (
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
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
	Proposer    peer.ID
	Height      uint64
	Round       uint32
	Step        Step
	Block       cid.Cid
	ParentBlock cid.Cid

	f uint64

	voteMu     sync.Mutex
	PreVotes   map[peer.ID]*Msg
	PreCommits map[peer.ID]*Msg

	PreVotesEvidence   map[peer.ID]*ConsensusMsgEvidence
	PreCommitsEvidence map[peer.ID]*ConsensusMsgEvidence

	lockedValue cid.Cid
	lockedRound uint32

	pvOnce sync.Once
	pcOnce sync.Once
	sbOnce sync.Once
}

func (s *State) resetOnces() {
	s.pvOnce = sync.Once{}
	s.pcOnce = sync.Once{}
	s.sbOnce = sync.Once{}
}
