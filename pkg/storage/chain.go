package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"sort"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/pkg/errors"
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

func (b BlockID) String() string {
	return cid.Cid(b).String()
}

type Block struct {
	Version   uint32   `msgpack:"v"`
	ID        BlockID  `magpack:"i"`
	Parent    BlockID  `msgpack:"p"`
	Height    uint64   `msgpack:"h"`
	CreatedAt int64    `msgpack:"t"`
	Proposer  string   `msgpack:"w"`
	Signature []byte   `msgpack:"s"`
	Nonce     [32]byte `msgpack:"n"`
	Bloom     []byte   `msgpack:"b"`
	TxRoot    cid.Cid  `msgpack:"x"`
}

func NewBlock(ctx context.Context, s Store, parent BlockID, txs []cid.Cid) (*Block, error) {
	var height uint64
	var err error

	if parent != BlockID(cid.Undef) {
		pb, err := s.GetBlock(ctx, parent)
		if err != nil {
			return nil, errors.Wrap(err, "getting parent")
		}
		height = pb.Height + 1
	}

	b := &Block{
		Version:   Version,
		Parent:    parent,
		Height:    height,
		CreatedAt: time.Now().Unix(),
	}

	if _, err := rand.Read(b.Nonce[:]); err != nil {
		return nil, errors.Wrap(err, "making nonce")
	}

	b.Bloom, err = MakeBloom(txs)
	if err != nil {
		return nil, errors.Wrap(err, "creating bloom")
	}

	rCid, err := NewTxSet(s, txs)
	if err != nil {
		return nil, errors.Wrap(err, "making tx set")
	}

	b.TxRoot = rCid.Cid()

	return b, nil
}

type TxSet struct {
	Children []cid.Cid `msgpack:"c"`
	Tx       *cid.Cid  `msgpack:"t,omitempty"`

	cid cid.Cid
}

func (txs *TxSet) Cid() cid.Cid {
	return txs.cid
}

type cidList []cid.Cid

func (cl cidList) Len() int           { return len(cl) }
func (cl cidList) Less(i, j int) bool { return bytes.Compare(cl[i].Bytes(), cl[j].Bytes()) == -1 }
func (cl cidList) Swap(i, j int)      { cl[i], cl[j] = cl[j], cl[i] }

// NewTxSet constructs a new Merkle tree of TX CIDs and stores the TxSet in the
// provided store as the Merkle tree reduces to its root node
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
		ttx := &TxSet{Tx: &t}

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
