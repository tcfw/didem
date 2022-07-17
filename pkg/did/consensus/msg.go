package consensus

import (
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type MsgType uint16

const (
	MsgTypeTx MsgType = iota + 1
	MsgTypeBlock
	MsgTypeConsensus
)

type ConsensusMsgType uint8

const (
	ConsensusMsgTypeNewRound ConsensusMsgType = iota + 1
	ConsensusMsgTypeProposal
	ConsensusMsgTypeVote
	ConsensusMsgTypePreCommit
	ConsensusMsgTypeBlock
)

type Msg struct {
	Type      MsgType       `msgpack:"t"`
	Tx        *TxMsg        `msgpack:"tx,omitempty"`
	Block     *BlockMsg     `msgpack:"bl,omitempty"`
	Consensus *ConsensusMsg `msgpack:"cn,omitempty"`
}

func (m *Msg) Marshal() ([]byte, error) {
	return msgpack.Marshal(m)
}

type TxMsg struct{}
type BlockMsg struct{}

type ConsensusMsg struct {
	Type     ConsensusMsgType      `msgpack:"t"`
	NewRound *ConsensusMsgNewRound `msgpack:"nr,omitempty"`
	Proposal *ConsensusMsgProposal `msgpack:"po,omitempty"`
	Vote     *ConsensusMsgVote     `msgpack:"v,omitempty"`
	Block    *ConsensusMsgBlock    `msgpack:"b,omitempty"`
}

type ConsensusMsgNewRound struct {
	Height                int64 `msgpack:"h"`
	Round                 int32 `msgpack:"r"`
	SecondsSinceStartTime int64 `msgpack:"s"`
	LastCommitRound       int32 `msgpack:"l"`
}

type ConsensusMsgProposal struct {
	Height    uint64    `msgpack:"h,string"`
	Round     uint32    `msgpack:"r"`
	BlockID   string    `msgpack:"id"`
	Timestamp time.Time `msgpack:"ts"`
	Signature []byte    `msgpack:"s"`
}

type VoteType uint8

const (
	VoteTypePreVote VoteType = iota + 1
	VoteTypePreCommit
	VoteTypeProposal
)

type ConsensusMsgVote struct {
	Type             VoteType  `msgpack:"t"`
	Height           uint64    `msgpack:"h,omitempty"`
	Round            uint32    `msgpack:"r,omitempty"`
	BlockID          string    `msgpack:"id"`
	Timestamp        time.Time `msgpack:"ts"`
	ValidatorAddress []byte    `msgpack:"v_a,omitempty"`
	ValidatorIndex   int32     `msgpack:"v_i,omitempty"`
	Signature        []byte    `msgpack:"s,omitempty"`
}

type ConsensusMsgBlock struct {
	Height uint64 `msgpack:"h"`
	Round  uint32 `msgpack:"r"`
	CID    string `msgpack:"c"`
}
