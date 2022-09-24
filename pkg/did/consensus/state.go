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

	pvOnce resync.Once
	pcOnce resync.Once
	sbOnce resync.Once
}

func (s *State) resetOnces() {
	s.pvOnce = resync.Once{}
	s.pcOnce = resync.Once{}
	s.sbOnce = resync.Once{}
}
