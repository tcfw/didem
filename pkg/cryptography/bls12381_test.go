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

func TestBls12381Aggregates(t *testing.T) {
	msg := []byte("abc")
	sigs := [][]byte{}

	pks := []*Bls12381PublicKey{}

	for i := 0; i < 10; i++ {
		sk := NewBls12381PrivateKey()
		pk := sk.Public()
		pks = append(pks, pk.(*Bls12381PublicKey))

		sig, err := sk.Sign(nil, msg, nil)
		if err != nil {
			t.Fatal(err)
		}
		sigs = append(sigs, sig)
	}

	aggPubP := AggregateBls12381PublicKeys(pks...)
	aggSig, err := AggregateBls12381Signatures(sigs...)
	if err != nil {
		t.Fatal(err)
	}

	pub := &Bls12381PublicKey{aggPubP}
	ok, err := pub.Verify(aggSig, msg)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, ok)
}
