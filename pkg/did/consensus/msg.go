package consensus

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/tcfw/didem/pkg/storage"
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
	ConsensusMsgTypeEvidence
)

type Msg struct {
	Type      MsgType       `msgpack:"t"`
	From      peer.ID       `msgpack:"p"`
	Chain     []byte        `msgpack:"c"`
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
	Tx  *tx.Tx `msgpack:"t"`
	TTL int    `msgpack:"ttl"`
}

type BlockMsg struct {
	Block *storage.Block `msgpack:"b"`
}

type ConsensusMsg struct {
	Type     ConsensusMsgType      `msgpack:"t"`
	NewRound *ConsensusMsgNewRound `msgpack:"n,omitempty"`
	Proposal *ConsensusMsgProposal `msgpack:"p,omitempty"`
	Vote     *ConsensusMsgVote     `msgpack:"v,omitempty"`
	Block    *ConsensusMsgBlock    `msgpack:"b,omitempty"`
	Evidence *ConsensusMsgEvidence `msgpack:"e,omitempty"`
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
	Height    uint64    `msgpack:"h"`
	Round     uint32    `msgpack:"r"`
	BlockID   string    `msgpack:"id"`
	Timestamp time.Time `msgpack:"ts"`
	Validator string    `msgpack:"v"`
}

type ConsensusMsgEvidence struct {
	ConsensusMsgVote
	Accepted bool `msgpack:"a"`
}

type ConsensusMsgBlock struct {
	Height     uint64 `msgpack:"h"`
	Round      uint32 `msgpack:"r"`
	CID        string `msgpack:"c"`
	Validators []byte `msgpack:"v"`
	Signature  []byte `msgpack:"s"`
}
