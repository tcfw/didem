package tx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign/bls"
	"go.dedis.ch/kyber/v3/util/random"
)

func TestTxEncodeDecode(t *testing.T) {
	_, pub := bls.NewKeyPair(bn256.NewSuite(), random.New())
	pubB, _ := pub.MarshalBinary()

	tx := &Tx{
		Version: Version1,
		Ts:      time.Now().Unix(),
		Type:    TxType_Node,
		Data: &Node{
			Id:  "1111",
			Did: "did:example:abcdefghijklmnopqrstuvwxyz0123456789",
			Key: pubB,
		},
	}

	b, err := tx.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	txrb := &Tx{}
	if err := txrb.Unmarshal(b); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, tx, txrb)
}
