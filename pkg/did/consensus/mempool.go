package consensus

import (
	"container/heap"

	"github.com/tcfw/didem/pkg/tx"
)

type MemPool interface {
	AddTx(*tx.Tx) error
	GetTx() *tx.Tx
}

var (
	_ MemPool = (*TxMemPool)(nil)
)

type TxList []*tx.Tx

func (h TxList) Len() int           { return len(h) }
func (h TxList) Less(i, j int) bool { return h[i].Ts < h[j].Ts }

func (h TxList) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *TxList) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(*tx.Tx))
}

func (h *TxList) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type TxMemPool struct {
	plist TxList
}

func NewTxMemPool() *TxMemPool {
	l := &TxMemPool{
		plist: make([]*tx.Tx, 0),
	}

	heap.Init(l.plist)

	return l
}

func (m *TxMemPool) GetTx() *tx.Tx {
	if m.plist.Len() > 0 {
		t := heap.Pop(&m.plist)
		return t.(*tx.Tx)
	}

	return nil
}

func (m *TxMemPool) AddTx(tx *tx.Tx) error {
	heap.Push(&m.plist, tx)
	return nil
}
