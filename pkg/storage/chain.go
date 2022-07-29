package storage

import (
	"context"

	"github.com/ipfs/go-cid"
	"github.com/pkg/errors"
	"github.com/tcfw/didem/pkg/tx"
)

const (
	MaxBlockTxCount = 1000
	MaxSetSize      = 100
	Version         = 1
)

type BlockID string

type Block struct {
	Version   uint32   `msgpack:"v"`
	ID        BlockID  `magpack:"i"`
	Parent    BlockID  `msgpack:"p"`
	Height    uint64   `msgpack:"h"`
	CreatedAt int64    `msgpack:"t"`
	Proposer  string   `msgpack:"w"`
	Signers   []byte   `msgpack:"sn"`
	Signature []byte   `msgpack:"s"`
	Nonce     [32]byte `msgpack:"n"`
	Bloom     []byte   `msgpack:"b"`
	TxRoot    cid.Cid  `msgpack:"x"`
}

type TxSet struct {
	Children []cid.Cid `msgpack:"c"`
	Txs      []cid.Cid `msgpack:"t,omitempty"`
}

func NewTxSet(s Store, txs []cid.Cid) (*TxSet, error) {
	n := len(txs)

	if n > MaxBlockTxCount {
		return nil, errors.New("too many tx for set")
	}

	if n < MaxSetSize {
		return &TxSet{Txs: txs}, nil
	}

	//TODO(tcfw)
	return nil, errors.New("not implemented")
}

type Validator interface {
	IsValid(b *Block) error
}

type TxValidator struct {
	s Store
}

func NewTxValidator(s Store) *TxValidator {
	return &TxValidator{s}
}

func (v *TxValidator) IsValid(b *Block) error {
	return nil
}

func (v *TxValidator) AllTx(ctx context.Context, b *Block) ([]*tx.Tx, error) {
	txSeen := map[string]*tx.Tx{}
	visited := cid.Set{}
	queue := []cid.Cid{b.TxRoot}

	for len(queue) != 0 {
		//pop
		setCid := queue[0]
		queue = queue[1:]

		//just incase we encounter a loop (bad proposer?)
		if visited.Has(setCid) {
			continue
		} else {
			visited.Add(setCid)
		}

		set, err := v.s.GetSet(ctx, setCid)
		if err != nil {
			return nil, errors.Wrap(err, "getting root trie")
		}

		//max check
		if len(set.Txs) > MaxSetSize {
			return nil, errors.Wrap(err, "too many tx in set")
		}

		for _, tcid := range set.Txs {
			tx, err := v.s.GetTx(ctx, tcid)
			if err != nil {
				return nil, errors.Wrap(err, "getting root tx")
			}

			txSeen[tcid.KeyString()] = tx

			//max check
			if len(txSeen) > MaxBlockTxCount {
				return nil, errors.New("block containers too many tx")
			}
		}

		//push
		if len(set.Children) > 0 {
			queue = append(queue, set.Children...)
		}
	}

	txList := make([]*tx.Tx, 0, len(txSeen))
	for _, tx := range txSeen {
		txList = append(txList, tx)
	}

	return txList, nil
}
