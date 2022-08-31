package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tcfw/didem/pkg/did/w3cdid"
	"github.com/tcfw/didem/pkg/tx"
)

func TestValidDID(t *testing.T) {
	s := NewMemStore()
	v := NewTxValidator(s)

	tx := &tx.Tx{
		Signature: make([]byte, 10),
		Action:    tx.TxActionAdd,
		Data: &tx.DID{
			Document: &w3cdid.Document{
				ID: "did:example:1234",
			},
		},
	}

	err := v.isDIDTxValid(context.Background(), tx)

	assert.NoError(t, err)
}
