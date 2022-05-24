package tx

import (
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	Version1 uint8 = 1
)

type ID []byte

type TxType int8

const (
	TxType_PKPublish TxType = iota + 1
)

type Tx struct {
	Version  uint8       `msgpack:"v"`
	Children [2]ID       `msgpack:"c"`
	Ts       int64       `msgpack:"t"`
	Type     TxType      `msgpack:"T"`
	Data     interface{} `msgpack:"d,noinline"`
	Proof    Proof       `msgpack:"p,inline"`
}

type Proof struct {
	Signatures []TxSignature `msgpack:"s"`
}

type TxSignature struct {
	PubkTxId  ID     `msgpack:"p"`
	Signature []byte `msgpack:"s"`
}

type PKPublish struct {
	Nonce     []byte `msgpack:"n"`
	KeyType   uint16 `msgpack:"t"`
	PublicKey []byte `msgpack:"pk"`
}

func (t *Tx) Marshal() ([]byte, error) {
	b, err := msgpack.Marshal(t)
	if err != nil {
		return nil, errors.Wrap(err, "mashaling tx")
	}

	return b, nil
}

func (t *Tx) Unmarshal(b []byte) error {
	if err := msgpack.Unmarshal(b, t); err != nil {
		return err
	}

	switch t.Type {
	case TxType_PKPublish:
		raw, err := msgpack.Marshal(t.Data)
		if err != nil {
			return err
		}

		pkp := &PKPublish{}
		if err := msgpack.Unmarshal(raw, pkp); err != nil {
			return err
		}

		t.Data = pkp
	default:
		return errors.New("unknown data type")
	}

	return nil
}
