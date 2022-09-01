package storage

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/multiformats/go-multibase"
	"github.com/stretchr/testify/assert"
	"github.com/tcfw/didem/pkg/did/w3cdid"
	"github.com/tcfw/didem/pkg/did/w3cdid/cryptography"
	"github.com/tcfw/didem/pkg/tx"
)

func TestValidDID(t *testing.T) {
	s := NewMemStore()
	v := NewTxValidator(s)

	pk, sk, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	mb, err := multibase.Encode(multibase.Base64, pk)
	if err != nil {
		t.Fatal(err)
	}

	tx := &tx.Tx{
		Version: tx.Version1,
		Ts:      time.Now().Unix(),
		Action:  tx.TxActionAdd,
		From:    "did:example:1234",
		Data: &tx.DID{
			Document: &w3cdid.Document{
				ID: "did:example:1234",
				VerificationMethod: []cryptography.VerificationMethod{{
					ID:                 "did:example:1234",
					Controller:         "did:example:1234",
					Type:               cryptography.Ed25519VerificationKey2018,
					PublicKeyMultibase: mb,
				}},
			},
		},
	}

	msg, err := tx.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	tx.Signature = ed25519.Sign(sk, msg)

	err = v.isDIDTxValid(context.Background(), tx)

	assert.NoError(t, err)
}

func TestInvalidTx(t *testing.T) {
	s := NewMemStore()
	v := NewTxValidator(s)

	ttx := &tx.Tx{
		Data: nil,
	}

	err := v.IsTxValid(context.Background(), ttx)
	assert.ErrorIs(t, err, ErrTxVersionNotSupported)

	ttx.Version = tx.Version1

	err = v.IsTxValid(context.Background(), ttx)
	assert.ErrorIs(t, err, ErrTxMissingTimestamp)

	ttx.Ts = time.Now().Unix()

	err = v.IsTxValid(context.Background(), ttx)
	assert.ErrorIs(t, err, ErrTxMissingFrom)

	ttx.From = "did:example:1234"

	err = v.IsTxValid(context.Background(), ttx)
	assert.ErrorIs(t, err, ErrTxMissingData)

	ttx.Data = &tx.DID{
		Document: &w3cdid.Document{
			ID: "did:example:1234",
		},
	}

	err = v.IsTxValid(context.Background(), ttx)
	assert.ErrorIs(t, err, ErrTxUnsupportedType)

	ttx.Type = tx.TxType_DID

	err = v.IsTxValid(context.Background(), ttx)
	assert.ErrorIs(t, err, ErrOpNotSupported)

	ttx.Action = tx.TxActionAdd

	err = v.isDIDTxValid(context.Background(), ttx)
	assert.ErrorIs(t, err, ErrDIDInvalidSignature)
}
