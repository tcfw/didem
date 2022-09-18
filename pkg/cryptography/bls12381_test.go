package cryptography

import (
	"testing"

	"github.com/multiformats/go-multibase"
	"github.com/stretchr/testify/assert"
)

func TestVerifyBls12381(t *testing.T) {
	sk := NewBls12381PrivateKey()
	pk := sk.Public().(*Bls12381PublicKey)

	pkb, err := pk.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	pkmb, err := multibase.Encode(multibase.Base58BTC, pkb)
	if err != nil {
		t.Fatal(err)
	}

	msg := []byte("abc")

	sig, err := sk.Sign(nil, msg, nil)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := ValidateBls12381(VerificationMethod{PublicKeyMultibase: pkmb}, sig, msg)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, ok)
}
