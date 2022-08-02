package storage

import (
	"context"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"github.com/tcfw/didem/pkg/tx"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	ErrNotFound = errors.New("not found")

	_ Store = (*MemStore)(nil)
)

type MemStore struct {
	store    map[cid.Cid][]byte
	txbIndex map[tx.TxID]BlockID
	bState   map[BlockID]BlockState
	mu       sync.RWMutex
}

func NewMemStore() *MemStore {
	return &MemStore{
		store:    make(map[cid.Cid][]byte),
		txbIndex: make(map[tx.TxID]BlockID),
		bState:   make(map[BlockID]BlockState),
	}
}

func (m *MemStore) putObj(obj interface{}) cid.Cid {
	d, _ := msgpack.Marshal(obj)

	h, _ := multihash.Sum(d, multihash.SHA3_256, multihash.DefaultLengths[multihash.SHA3_256])
	cid := cid.NewCidV1(cid.Raw, h)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.store[cid] = d

	return cid
}

func (m *MemStore) getObj(id cid.Cid) []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.store[id]
}

func (m *MemStore) PutTx(ctx context.Context, tx *tx.Tx) (cid.Cid, error) {
	return m.putObj(tx), nil
}

func (m *MemStore) GetTx(ctx context.Context, id tx.TxID) (*tx.Tx, error) {
	d := m.getObj(cid.Cid(id))
	if d == nil {
		return nil, ErrNotFound
	}

	tx := &tx.Tx{}
	if err := msgpack.Unmarshal(d, tx); err != nil {
		return nil, errors.Wrap(err, "unmarshalling ")
	}

	return tx, nil
}

func (m *MemStore) PutBlock(ctx context.Context, b *Block) (cid.Cid, error) {
	return m.putObj(b), nil
}

func (m *MemStore) GetTxBlock(ctx context.Context, id tx.TxID) (*Block, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	blkId, ok := m.txbIndex[id]
	if !ok {
		return nil, ErrNotFound
	}

	return m.GetBlock(ctx, blkId)
}

func (m *MemStore) MarkBlock(ctx context.Context, b BlockID, s BlockState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	c := m.bState[b]
	m.bState[b] = s

	if c == BlockStateValidated && s == BlockStateAccepted {
		m.indexBlockTx(ctx, b)
	}

	return nil
}

func (m *MemStore) indexBlockTx(ctx context.Context, b BlockID) {

}

func (m *MemStore) GetBlock(ctx context.Context, id BlockID) (*Block, error) {
	d := m.getObj(cid.Cid(id))
	if d == nil {
		return nil, ErrNotFound
	}

	b := &Block{}
	if err := msgpack.Unmarshal(d, b); err != nil {
		return nil, errors.Wrap(err, "unmarshalling ")
	}

	return b, nil
}

func (m *MemStore) PutSet(ctx context.Context, txt *TxSet) (cid.Cid, error) {
	return m.putObj(txt), nil
}

func (m *MemStore) GetSet(ctx context.Context, id cid.Cid) (*TxSet, error) {
	d := m.getObj(id)
	if d == nil {
		return nil, ErrNotFound
	}

	txt := &TxSet{}
	if err := msgpack.Unmarshal(d, txt); err != nil {
		return nil, errors.Wrap(err, "unmarshalling ")
	}

	return txt, nil
}
