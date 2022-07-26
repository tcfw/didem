package storage

import (
	"context"

	"github.com/ipfs/go-cid"
	"github.com/pkg/errors"
	"github.com/tcfw/didem/pkg/tx"
)

const (
	maxBlockTxCount = 1000
)

type BlockID string

type Block struct {
	Version   uint32  `msgpack:"v"`
	ID        BlockID `magpack:"i"`
	Parent    BlockID `msgpack:"p"`
	Height    uint64  `msgpack:"h"`
	CreatedAt uint64  `msgpack:"t"`
	Proposer  string  `msgpack:"w"`
	Signers   uint32  `msgpack:"sn"`
	Signature []byte  `msgpack:"s"`
	Nonce     []byte  `msgpack:"n"`
	TxRoot    cid.Cid `msgpack:"x"`
}

type TxTrie struct {
	Children []cid.Cid `msgpack:"c"`
	Tx       *cid.Cid  `msgpack:"t,omitempty"`
}

type Validator interface {
	IsValid(b *Block) error
}

type TxValidator struct {
	s Store
}

func (v *TxValidator) IsValid(b *Block) error {
	return nil
}

func (v *TxValidator) AllTx(ctx context.Context, b *Block) ([]*tx.Tx, error) {
	txSeen := map[string]*tx.Tx{}
	visited := map[cid.Cid]struct{}{}
	queue := []cid.Cid{b.TxRoot}

	for len(queue) != 0 {
		//pop
		trieCid := queue[0]
		queue = queue[1:]

		//just incase we encounter a loop (bad proposer?)
		if _, ok := visited[trieCid]; ok {
			continue
		} else {
			visited[trieCid] = struct{}{}
		}

		trie, err := v.s.GetTrie(ctx, trieCid)
		if err != nil {
			return nil, errors.Wrap(err, "getting root trie")
		}

		if trie.Tx != nil {
			tx, err := v.s.GetTx(ctx, *trie.Tx)
			if err != nil {
				return nil, errors.Wrap(err, "getting root tx")
			}

			txSeen[trie.Tx.KeyString()] = tx

			//max check
			if len(txSeen) > maxBlockTxCount {
				return nil, errors.New("block containers too many tx")
			}
		}

		//push
		if len(trie.Children) > 0 {
			queue = append(queue, trie.Children...)
		}
	}

	txList := make([]*tx.Tx, 0, len(txSeen))
	for _, tx := range txSeen {
		txList = append(txList, tx)
	}

	return txList, nil
}
