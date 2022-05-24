package tx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	tx := &Tx{
		Version: Version1,
		Ts:      time.Now().Unix(),
		Type:    TxType_PKPublish,
		Data: &PKPublish{
			Nonce:     []byte{1, 2, 3, 4, 5},
			PublicKey: []byte{1, 2, 3, 4, 5},
		},
	}

	b, err := tx.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	txRB := &Tx{}

	if err := txRB.Unmarshal(b); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, tx, txRB)
}
