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

const (
	TxType_PKPublish TxType = iota + 1
)

type TxID cid.Cid

type Tx struct {
	Version uint8       `msgpack:"v"`
	Ts      int64       `msgpack:"t"`
	Type    TxType      `msgpack:"T"`
	Data    interface{} `msgpack:"d,noinline"`
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
		//DIDComm decode
	default:
		return errors.New("unknown data type")
	}

	return nil
}
