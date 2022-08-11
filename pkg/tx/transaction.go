package tx

import (
	"github.com/ipfs/go-cid"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	Version1 uint8 = 1
)

type TxType int8
type TxAction int8

const (
	TxType_DID TxType = iota + 1
	TxType_VC
	TxType_Node
)

const (
	TxActionAdd TxAction = iota
	TxActionUpdate
	TxActionRevoke
)

type TxID cid.Cid

type Tx struct {
	Version   uint8       `msgpack:"v"`
	Ts        int64       `msgpack:"t"`
	From      string      `msgpack:"f"`
	Type      TxType      `msgpack:"e"`
	Action    TxAction    `msgpack:"a"`
	Data      interface{} `msgpack:"d,noinline"`
	Signature []byte      `msgpack:"s"`
}

func (t *Tx) Marshal() ([]byte, error) {
	b, err := msgpack.Marshal(t)
	if err != nil {
		return nil, errors.Wrap(err, "mashaling tx")
	}

	return b, nil
}

func (t *Tx) Unmarshal(b []byte) error {
	t.Data = &msgpack.RawMessage{}

	if err := msgpack.Unmarshal(b, t); err != nil {
		return err
	}

	switch t.Type {
	case TxType_DID:
		//DID decode
		t.Data = &DID{}
	case TxType_VC:
		//VC decode
	case TxType_Node:
		//Node decode
		t.Data = &Node{}
	default:
		return errors.New("unknown data type")
	}

	if err := msgpack.Unmarshal(b, t); err != nil {
		return err
	}

	return nil
}
