package cryptography

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/multiformats/go-multibase"
	"github.com/stretchr/testify/assert"
)

func TestEd25519GoodSignature(t *testing.T) {
	did := "did:example:1234"

	pk, sk, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	mb, err := multibase.Encode(multibase.Base64, pk)
	if err != nil {
		t.Fatal(err)
	}

	vm := &VerificationMethod{
		ID:                 did,
		Type:               Ed25519VerificationKey2018,
		Controller:         did,
		PublicKeyMultibase: mb,
	}

	msg := []byte("test")

	sig := ed25519.Sign(sk, msg)

	ok, err := ValidateEd25519(*vm, sig, msg)
	assert.NoError(t, err)
	assert.True(t, ok)
}
