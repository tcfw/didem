package storage

import (
	"bytes"
	"context"
	"sort"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/pkg/errors"
	"github.com/tcfw/didem/pkg/tx"
)

const (
	MaxBlockTxCount = 1000
	MaxSetSize      = 100
	Version         = 1
)

type BlockState int

const (
	BlockStateUnvalidated = iota
	BlockStateValidated
	BlockStateAccepted
)

type BlockID cid.Cid

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

type cidList []cid.Cid

func (cl cidList) Len() int           { return len(cl) }
func (cl cidList) Less(i, j int) bool { return bytes.Compare(cl[i].Bytes(), cl[j].Bytes()) == -1 }
func (cl cidList) Swap(i, j int)      { cl[i], cl[j] = cl[j], cl[i] }

func NewTxSet(s Store, txs []cid.Cid) (cid.Cid, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sort.Sort(cidList(txs))

	n := len(txs)

	if n > MaxBlockTxCount {
		return cid.Undef, errors.New("too many tx for set")
	}

	var root *TxSet

	//TODO(tcfw)

	c, err := s.PutSet(ctx, root)
	if err != nil {
		return cid.Undef, errors.Wrap(err, "storing root txset")
	}

	return c, nil
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
			tx, err := v.s.GetTx(ctx, tx.TxID(tcid))
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
