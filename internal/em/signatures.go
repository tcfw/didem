package em

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"

	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

func verify(pk crypto.PublicKey, b []byte) error {
	return errors.New("mismatching signature")
}

func sign(pk crypto.PrivateKey, b []byte) ([]byte, error) {
	h := sha3.Sum384(b)

	switch t := pk.(type) {
	case *ecdsa.PrivateKey:
		return pk.(*ecdsa.PrivateKey).Sign(rand.Reader, h[:], nil)
	case *ed25519.PrivateKey:
		return pk.(*ed25519.PrivateKey).Sign(rand.Reader, b, nil)
	case *rsa.PrivateKey:
		return pk.(*rsa.PrivateKey).Sign(rand.Reader, h[:], nil)
	default:
		return nil, errors.Errorf("unknown private key type: %T", t)
	}
}
