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
	Height      uint64
	Round       uint32
	Step        Step
	Block       cid.Cid
	ParentBlock cid.Cid

	f uint64

	PreVotes   map[peer.ID]*ConsensusMsgVote
	PreCommits map[peer.ID]*ConsensusMsgVote

	lockedValue cid.Cid
	lockedRound int32
}
