package did

import (
	"crypto"
	"crypto/ed25519"
	"io"

	"github.com/multiformats/go-multihash"
)

type PrivateIdentity interface {
	PrivateKey() crypto.PrivateKey
	PublicIdentity() (*PublicIdentity, error)
}

type Ed25519Identity struct {
	sk ed25519.PrivateKey
}

func NewEd25519Identity(sk []byte) *Ed25519Identity {
	return &Ed25519Identity{sk: ed25519.PrivateKey(sk)}
}

func GenerateEd25519Identity(rand io.Reader) (PrivateIdentity, error) {
	_, sk, err := ed25519.GenerateKey(rand)
	if err != nil {
		return nil, err
	}

	return &Ed25519Identity{sk}, nil
}

func (e *Ed25519Identity) PublicIdentity() (*PublicIdentity, error) {
	pk := e.sk.Public().(ed25519.PublicKey)

	mh, err := multihash.Sum(pk, multihash.SHA3_384, multihash.DefaultLengths[multihash.SHA3_384])
	if err != nil {
		return nil, err
	}

	id := mh.B58String()

	pub := &PublicIdentity{ID: id, PublicKeys: []PublicKey{{Key: pk}}}

	return pub, nil
}

func (e *Ed25519Identity) PrivateKey() crypto.PrivateKey {
	return e.sk
}
