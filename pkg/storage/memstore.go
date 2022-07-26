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
	store map[cid.Cid][]byte
	mu    sync.RWMutex
}

func NewMemStore() *MemStore {
	return &MemStore{
		store: make(map[cid.Cid][]byte),
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

func (m *MemStore) GetTx(ctx context.Context, id cid.Cid) (*tx.Tx, error) {
	d := m.getObj(id)
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

func (m *MemStore) GetBlock(ctx context.Context, id cid.Cid) (*Block, error) {
	d := m.getObj(id)
	if d == nil {
		return nil, ErrNotFound
	}

	b := &Block{}
	if err := msgpack.Unmarshal(d, b); err != nil {
		return nil, errors.Wrap(err, "unmarshalling ")
	}

	return b, nil
}

func (m *MemStore) PutTrie(ctx context.Context, txt *TxTrie) (cid.Cid, error) {
	return m.putObj(txt), nil
}

func (m *MemStore) GetTrie(ctx context.Context, id cid.Cid) (*TxTrie, error) {
	d := m.getObj(id)
	if d == nil {
		return nil, ErrNotFound
	}

	txt := &TxTrie{}
	if err := msgpack.Unmarshal(d, txt); err != nil {
		return nil, errors.Wrap(err, "unmarshalling ")
	}

	return txt, nil
}
