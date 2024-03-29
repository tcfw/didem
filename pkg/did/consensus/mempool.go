package consensus

import (
	"container/heap"
	"sync"

	"github.com/tcfw/didem/pkg/tx"
)

// MemPool provides a means of storing new TXs that are NOT in the current
// block chain. The MemPool is to provide TXs in sequental time as stated
// by the TX, not but the time received
type MemPool interface {
	AddTx(*tx.Tx, int) error
	GetTx() *tx.Tx
	Len() int
}

var (
	_ MemPool = (*TxMemPool)(nil)
)

type TxList []*tx.Tx

type TxMemPool struct {
	plist TxList
	mu    sync.Mutex
}

// NewTxMemPool creates a new empty MemPool
func NewTxMemPool() *TxMemPool {
	l := &TxMemPool{
		plist: make([]*tx.Tx, 0),
	}

	heap.Init(l)

	return l
}

func (m *TxMemPool) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.plist)
}

func (m *TxMemPool) Less(i, j int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.plist[i].Ts < m.plist[j].Ts
}

func (m *TxMemPool) Swap(i, j int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.plist[i], m.plist[j] = m.plist[j], m.plist[i]
}

func (m *TxMemPool) Push(x interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.plist = append(m.plist, x.(*tx.Tx))
}

func (m *TxMemPool) Pop() interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	old := m.plist
	n := len(old)
	x := old[n-1]
	m.plist = old[0 : n-1]
	return x
}

// GetTx gets the next, most recent TX in the pool
func (m *TxMemPool) GetTx() *tx.Tx {
	if m.Len() > 0 {
		t := heap.Pop(m)
		return t.(*tx.Tx)
	}

	return nil
}

// AddTx pushes a new TX into the MemPool and updates the priorities
// of any existing TX in the pool
func (m *TxMemPool) AddTx(tx *tx.Tx, expires int) error {
	heap.Push(m, tx)
	return nil
}
