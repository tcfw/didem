package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tcfw/didem/pkg/did/w3cdid"
	"github.com/tcfw/didem/pkg/tx"
)

func TestMemStore(t *testing.T) {
	m := NewMemStore()

	obj := &tx.Tx{
		Version: 1,
		Ts:      time.Now().Unix(),
		Type:    tx.TxType_DID,
		Action:  tx.TxActionAdd,
		Data:    &tx.DID{Document: &w3cdid.Document{}},
	}

	id, err := m.PutTx(context.Background(), obj)
	if err != nil {
		t.Fatal(err)
	}

	tx, err := m.GetTx(context.Background(), tx.TxID(id))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, obj, tx)
}
