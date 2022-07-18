package comm

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"

	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

func sign(pk crypto.PrivateKey, b []byte) ([]byte, error) {
	h := sha3.Sum512(b)

	switch t := pk.(type) {
	case *ecdsa.PrivateKey:
		return ecdsa.SignASN1(rand.Reader, pk.(*ecdsa.PrivateKey), h[:])
	case ed25519.PrivateKey:
		s := ed25519.Sign(pk.(ed25519.PrivateKey), b)
		return s, nil
	case *rsa.PrivateKey:
		return rsa.SignPKCS1v15(rand.Reader, pk.(*rsa.PrivateKey), crypto.SHA3_512, h[:])
	default:
		return nil, errors.Errorf("unknown private key type: %T", t)
	}
}

func verify(pk crypto.PublicKey, b []byte, sig []byte) error {
	h := sha3.Sum512(b)

	var valid bool

	switch t := pk.(type) {
	case *rsa.PublicKey:
		rsa.VerifyPKCS1v15(pk.(*rsa.PublicKey), crypto.SHA3_512, h[:], sig)
	case ed25519.PublicKey:
		valid = ed25519.Verify(pk.(ed25519.PublicKey), b, sig)
	case *ecdsa.PublicKey:
		valid = ecdsa.VerifyASN1(pk.(*ecdsa.PublicKey), h[:], sig)
	default:
		return errors.Errorf("unknown public key type: %T", t)
	}

	if !valid {
		return errors.New("invalid signature")
	}

	return nil

}
