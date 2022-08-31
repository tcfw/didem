package storage

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"github.com/tcfw/didem/pkg/did/genesis"
	"github.com/tcfw/didem/pkg/did/w3cdid"
	"github.com/tcfw/didem/pkg/tx"
	"github.com/vmihailenco/msgpack/v5"
)

var (

	_ Store = (*MemStore)(nil)
)

type MemStore struct {
	mu     sync.RWMutex
	metaMu sync.RWMutex

	objects  map[cid.Cid][]byte
	txbIndex map[tx.TxID]BlockID
	bState   map[BlockID]BlockState
	nodes    map[string]cid.Cid
	dids     map[string]cid.Cid
	claims   map[string][]cid.Cid

	lastBlock BlockID

	genesisApplied bool
}

func NewMemStore() *MemStore {
	return &MemStore{
		objects:   make(map[cid.Cid][]byte),
		txbIndex:  make(map[tx.TxID]BlockID),
		bState:    make(map[BlockID]BlockState),
		nodes:     make(map[string]cid.Cid),
		dids:      make(map[string]cid.Cid),
		claims:    make(map[string][]cid.Cid),
		lastBlock: BlockID(cid.Undef),
	}
}

func (m *MemStore) UpdateLastApplied(_ context.Context, id BlockID) error {
	m.lastBlock = id
	return nil
}

func (m *MemStore) GetLastApplied(ctx context.Context) (*Block, error) {
	if m.lastBlock == BlockID(cid.Undef) {
		return nil, nil
	}

	return m.GetBlock(ctx, m.lastBlock)
}

func (m *MemStore) putObj(obj interface{}) cid.Cid {
	d, _ := msgpack.Marshal(obj)

	h, _ := multihash.Sum(d, multihash.SHA3_256, multihash.DefaultLengths[multihash.SHA3_256])
	cid := cid.NewCidV1(cid.Raw, h)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.objects[cid] = d

	return cid
}

func (m *MemStore) getObj(id cid.Cid) []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.objects[id]
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
	if err := tx.Unmarshal(d); err != nil {
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

func (m *MemStore) AllTx(ctx context.Context, b *Block) (map[tx.TxID]*tx.Tx, error) {
	return m.allTxCids(ctx, b)
}

func (m *MemStore) allTxCids(ctx context.Context, b *Block) (map[tx.TxID]*tx.Tx, error) {
	txSeen := map[tx.TxID]*tx.Tx{}
	visited := cid.NewSet()
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

		set, err := m.GetSet(ctx, setCid)
		if err != nil {
			return nil, errors.Wrap(err, "getting root trie")
		}

		if set.Tx != nil {
			t, err := m.GetTx(ctx, tx.TxID(*set.Tx))
			if err != nil {
				return nil, errors.Wrap(err, "getting root tx")
			}

			txSeen[tx.TxID(*set.Tx)] = t

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

	return txSeen, nil
}

func (m *MemStore) MarkBlock(ctx context.Context, b BlockID, s BlockState) error {
	m.mu.Lock()
	c := m.bState[b]
	m.bState[b] = s
	m.mu.Unlock()

	if c == BlockStateValidated && s == BlockStateAccepted {
		if err := m.indexBlockTx(ctx, b); err != nil {
			return err
		}
	}

	return nil
}

func (m *MemStore) indexBlockTx(ctx context.Context, b BlockID) error {
	m.metaMu.Lock()
	defer m.metaMu.Unlock()

	block, err := m.GetBlock(ctx, b)
	if err != nil {
		return err
	}

	txs, err := m.allTxCids(ctx, block)
	if err != nil {
		return errors.Wrap(err, "getting block tx")
	}

	for c, t := range txs {
		m.txbIndex[c] = b

		cid, err := cid.Parse(cid.Cid(c))
		if err != nil {
			return errors.Wrap(err, "parsing tx id")
		}

		switch t.Type {
		case tx.TxType_DID:
			did := t.Data.(*tx.DID).Document.ID
			m.dids[did] = cid
		case tx.TxType_VC:
			// m.claims[]
		case tx.TxType_Node:
			n := t.Data.(*tx.Node)

			switch t.Action {
			case tx.TxActionAdd:
				m.nodes[n.Id] = cid
			case tx.TxActionRevoke:
				delete(m.nodes, n.Id)
			}
		}
	}

	return nil
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

func (m *MemStore) LookupDID(_ context.Context, id string) (*w3cdid.Document, error) {
	m.metaMu.RLock()
	defer m.metaMu.RUnlock()

	cid, ok := m.dids[id]
	if !ok {
		return nil, ErrNotFound
	}

	d := m.getObj(cid)
	if d == nil {
		return nil, ErrNotFound
	}

	doc := &w3cdid.Document{}
	if err := msgpack.Unmarshal(d, doc); err != nil {
		return nil, errors.Wrap(err, "unmarshalling ")
	}

	return doc, nil
}

func (m *MemStore) DIDHistory(_ context.Context, id string) ([]*tx.Tx, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MemStore) Claims(_ context.Context, id string) ([]*tx.Tx, error) {
	m.metaMu.RLock()
	defer m.metaMu.RUnlock()

	cids, ok := m.claims[id]
	if !ok || len(cids) == 0 {
		return nil, ErrNotFound
	}

	txs := make([]*tx.Tx, 0, len(cids))
	for _, cid := range cids {
		d := m.getObj(cid)
		if d == nil {
			return nil, ErrNotFound
		}

		tx := &tx.Tx{}
		if err := msgpack.Unmarshal(d, tx); err != nil {
			return nil, errors.Wrap(err, "unmarshalling ")
		}
	}

	return txs, nil
}

func (m *MemStore) Nodes() ([]string, error) {
	m.metaMu.RLock()
	defer m.metaMu.RUnlock()

	n := make([]string, len(m.nodes))

	var i int
	for k := range m.nodes {
		n[i] = k
		i++
	}

	sort.Strings(n)

	return n, nil
}

func (m *MemStore) Node(ctx context.Context, id string) (*tx.Node, error) {
	m.metaMu.RLock()
	defer m.metaMu.RUnlock()

	cid, ok := m.nodes[id]
	if !ok {
		return nil, ErrNotFound
	}

	d := m.getObj(cid)
	if d == nil {
		return nil, ErrNotFound
	}

	t := &tx.Tx{}
	if err := t.Unmarshal(d); err != nil {
		return nil, errors.Wrap(err, "unmarshalling")
	}

	return t.Data.(*tx.Node), nil
}

func (m *MemStore) HasGenesisApplied() bool {
	return m.genesisApplied
}

func (m *MemStore) ApplyGenesis(*genesis.Info) error {
	m.genesisApplied = true

	return nil
}

func (m *MemStore) Stop() error {
	return nil
}
