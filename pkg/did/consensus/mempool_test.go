package consensus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tcfw/didem/pkg/tx"
)

func TestMemPoolAddPriority(t *testing.T) {
	m := NewTxMemPool()

	if err := m.AddTx(&tx.Tx{Ts: 2}, 0); err != nil {
		t.Fatal(err)
	}

	if err := m.AddTx(&tx.Tx{Ts: 1}, 0); err != nil {
		t.Fatal(err)
	}

	if err := m.AddTx(&tx.Tx{Ts: 3}, 0); err != nil {
		t.Fatal(err)
	}

	assert.Len(t, m.plist, 3)

	t1 := m.GetTx()
	t2 := m.GetTx()
	t3 := m.GetTx()

	assert.Equal(t, int64(1), t1.Ts)
	assert.Equal(t, int64(2), t2.Ts)
	assert.Equal(t, int64(3), t3.Ts)
}
