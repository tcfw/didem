package cryptography

import (
	"crypto"
	"crypto/ed25519"

	"github.com/pkg/errors"

	"github.com/multiformats/go-multibase"
)

func decodeMultibase(mb string) ([]byte, error) {
	_, d, err := multibase.Decode(mb)
	return d, err
}

func EncodeMultibase(publicKey crypto.PublicKey) (string, error) {
	var raw []byte

	switch t := publicKey.(type) {
	case ed25519.PublicKey:
		raw = []byte(t)
	case *Bls12381PublicKey:
		b, err := t.Bytes()
		if err != nil {
			return "", err
		}
		raw = b
	default:
		return "", errors.Errorf("unsupported pk type: %T", t)

	}

	return multibase.Encode(multibase.Base58BTC, raw)
}
