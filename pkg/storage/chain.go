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
	Tx       cid.Cid   `msgpack:"t,omitempty"`

	cid cid.Cid
}

type cidList []cid.Cid

func (cl cidList) Len() int           { return len(cl) }
func (cl cidList) Less(i, j int) bool { return bytes.Compare(cl[i].Bytes(), cl[j].Bytes()) == -1 }
func (cl cidList) Swap(i, j int)      { cl[i], cl[j] = cl[j], cl[i] }

func NewTxSet(s Store, txs []cid.Cid) (*TxSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sort.Sort(cidList(txs))

	n := len(txs)

	if n > MaxBlockTxCount {
		return nil, errors.New("too many tx for set")
	}

	nodes := []*TxSet{}

	for _, t := range txs {
		ttx := &TxSet{Tx: t}

		c, err := s.PutSet(ctx, ttx)
		if err != nil {
			return nil, errors.Wrap(err, "storing root txset")
		}

		ttx.cid = c
		nodes = append(nodes, ttx)
	}

	for len(nodes) > 1 {
		if len(nodes)%2 == 1 {
			nodes = append(nodes, nodes[len(nodes)-1])
		}

		parents := []*TxSet{}

		for i := 0; i < len(nodes); i += 2 {
			n := &TxSet{
				Children: []cid.Cid{
					nodes[i].cid,
					nodes[i+1].cid,
				},
			}

			c, err := s.PutSet(ctx, n)
			if err != nil {
				return nil, errors.Wrap(err, "storing root txset")
			}
			n.cid = c

			parents = append(parents, n)
		}

		nodes = parents
	}

	return nodes[0], nil
}

type Validator interface {
	IsBlockValid(context.Context, *Block) error
	IsTxValid(context.Context, *tx.Tx) error
}

type TxValidator struct {
	s Store
}

func NewTxValidator(s Store) *TxValidator {
	return &TxValidator{s}
}

func (v *TxValidator) IsBlockValid(ctx context.Context, b *Block) error {
	txs, err := v.AllTx(ctx, b)
	if err != nil {
		return errors.Wrap(err, "getting block txs")
	}

	for _, tx := range txs {
		if err := v.IsTxValid(ctx, tx); err != nil {
			return errors.Wrap(err, "invalid tx in block")
		}
	}

	return nil
}

func (v *TxValidator) IsTxValid(ctx context.Context, t *tx.Tx) error {
	//TODO
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

		tx, err := v.s.GetTx(ctx, tx.TxID(set.Tx))
		if err != nil {
			return nil, errors.Wrap(err, "getting root tx")
		}

		txSeen[set.Tx.KeyString()] = tx

		//max check
		if len(txSeen) > MaxBlockTxCount {
			return nil, errors.New("block containers too many tx")
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
