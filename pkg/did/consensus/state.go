package consensus

import (
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
)

type Step uint8

const (
	propose Step = iota
	prevote
	precommit
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

	PreVotes   map[peer.ID]*Msg
	PreCommits map[peer.ID]*Msg

	PreVotesEvidence   map[peer.ID]*ConsensusMsgEvidence
	PreCommitsEvidence map[peer.ID]*ConsensusMsgEvidence

	lockedValue cid.Cid
	lockedRound uint32
}
