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

type Validator struct {
	s Storage
}

func (b *Block) IsValid() error {
	return nil
}

func (v *Validator) AllTx(ctx context.Context, b *Block) ([]*tx.Tx, error) {
	txSeen := map[string]*tx.Tx{}
	visited := map[cid.Cid]struct{}{}
	trieQ := []cid.Cid{b.TxRoot}

	for len(trieQ) != 0 {
		//pop
		el := trieQ[0]
		trieQ = trieQ[1:]

		//just incase we encounter a loop (bad proposer?)
		if _, ok := visited[el]; ok {
			continue
		} else {
			visited[el] = struct{}{}
		}

		trie, err := v.s.GetTrie(ctx, el)
		if err != nil {
			return nil, errors.Wrap(err, "getting root trie")
		}

		if trie.Tx != nil {
			tx, err := v.s.GetTx(ctx, *trie.Tx)
			if err != nil {
				return nil, errors.Wrap(err, "getting root tx")
			}

			txSeen[trie.Tx.KeyString()] = tx
			if len(txSeen) > maxBlockTxCount {
				return nil, errors.New("block containers too many tx")
			}
		}

		//push
		if len(trie.Children) > 0 {
			trieQ = append(trieQ, trie.Children...)
		}
	}

	txList := make([]*tx.Tx, 0, len(txSeen))

	for _, tx := range txSeen {
		txList = append(txList, tx)
	}

	return txList, nil
}
