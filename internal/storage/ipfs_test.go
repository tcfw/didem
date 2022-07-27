package storage

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tcfw/didem/pkg/tx"
	"golang.org/x/crypto/sha3"
)

func TestIPFSAdd(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ipfs, err := NewIPFSStorage(ctx)
	if err != nil {
		t.Fatal(err)
	}

	tx := &tx.Tx{}

	id, err := ipfs.PutTx(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}

	txb, _ := tx.Marshal()
	txbh := sha3.Sum384(txb)

	expected := hex.EncodeToString(txbh[:])

	idhex := hex.EncodeToString(id.Bytes())

	assert.Equal(t, expected, idhex)

	txrb, err := ipfs.GetTx(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, tx, txrb)
}
