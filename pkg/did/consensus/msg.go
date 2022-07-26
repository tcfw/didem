package consensus

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/tcfw/didem/pkg/tx"
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
	ConsensusMsgTypeBlock
)

type Msg struct {
	Type      MsgType       `msgpack:"t"`
	From      peer.ID       `msgpack:"p"`
	Tx        *TxMsg        `msgpack:"tx,omitempty"`
	Block     *BlockMsg     `msgpack:"bl,omitempty"`
	Consensus *ConsensusMsg `msgpack:"cn,omitempty"`
	Timestamp time.Time     `msgpack:"ts"`
	Signature []byte        `msgpack:"s"`
	Key       []byte        `msgpack:"k,omptempty"`
}

func (m *Msg) Marshal() ([]byte, error) {
	return msgpack.Marshal(m)
}

type TxMsg struct {
	Tx *tx.Tx `msgpack:"t"`
}

type BlockMsg struct {
	Block *Block `msgpack:"b"`
}

type ConsensusMsg struct {
	Type     ConsensusMsgType      `msgpack:"t"`
	NewRound *ConsensusMsgNewRound `msgpack:"nr,omitempty"`
	Proposal *ConsensusMsgProposal `msgpack:"po,omitempty"`
	Vote     *ConsensusMsgVote     `msgpack:"v,omitempty"`
	Block    *ConsensusMsgBlock    `msgpack:"b,omitempty"`
}

type ConsensusMsgNewRound struct {
	Height          uint64    `msgpack:"h"`
	Round           uint32    `msgpack:"r"`
	LastCommitRound uint32    `msgpack:"l"`
	Timestamp       time.Time `msgpack:"ts"`
}

type ConsensusMsgProposal struct {
	Height    uint64    `msgpack:"h,string"`
	Round     uint32    `msgpack:"r"`
	BlockID   string    `msgpack:"id"`
	Timestamp time.Time `msgpack:"ts"`
}

type VoteType uint8

const (
	VoteTypePreVote VoteType = iota + 1
	VoteTypePreCommit
	VoteTypeProposal
)

type ConsensusMsgVote struct {
	Type      VoteType  `msgpack:"t"`
	Height    uint64    `msgpack:"h,omitempty"`
	Round     uint32    `msgpack:"r,omitempty"`
	BlockID   string    `msgpack:"id"`
	Timestamp time.Time `msgpack:"ts"`
	Validator string    `msgpack:"v,omitempty"`
	Signature []byte    `msgpack:"s,omitempty"`
}

type ConsensusMsgBlock struct {
	Height    uint64 `msgpack:"h"`
	Round     uint32 `msgpack:"r"`
	CID       string `msgpack:"c"`
	Signature []byte `msgpack:"s,omitempty"`
}
