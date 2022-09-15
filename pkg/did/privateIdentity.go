package did

import (
	"crypto"
	"crypto/ed25519"
	"io"

	"github.com/multiformats/go-multihash"
	"github.com/tcfw/didem/pkg/cryptography"
)

// PrivateIdentity contains the secret identity whilst providing a means
// of producting the public identity of a key
type PrivateIdentity interface {
	PrivateKey() crypto.PrivateKey
	PublicIdentity() (*PublicIdentity, error)
}

type Bls12381G2Identity struct {
	sk *cryptography.Bls12381PrivateKey
}

func GenerateBls12381G2Identity() (PrivateIdentity, error) {
	b := &Bls12381G2Identity{
		sk: cryptography.NewBls12381PrivateKey(),
	}

	return b, nil
}

func (b *Bls12381G2Identity) PrivateKey() crypto.PrivateKey {
	return b.sk
}

func (b *Bls12381G2Identity) PublicIdentity() (*PublicIdentity, error) {
	pk := b.sk.Public().(*cryptography.Bls12381PublicKey)

	pkraw, err := pk.Bytes()
	if err != nil {
		return nil, err
	}

	mh, err := multihash.Sum(pkraw, multihash.SHA3_384, multihash.DefaultLengths[multihash.SHA3_384])
	if err != nil {
		return nil, err
	}

	id := mh.B58String()

	pub := &PublicIdentity{ID: id, PublicKeys: []PublicKey{{Key: pk}}}
	return pub, nil
}

type Ed25519Identity struct {
	sk ed25519.PrivateKey
}

// NewEd25519Identity creates a new Ed25519 type identity
func NewEd25519Identity(sk []byte) *Ed25519Identity {
	return &Ed25519Identity{sk: ed25519.PrivateKey(sk)}
}

// GenerateEd25519Identity generates a new Ed25519 private identity from a given
// random source
func GenerateEd25519Identity(rand io.Reader) (PrivateIdentity, error) {
	_, sk, err := ed25519.GenerateKey(rand)
	if err != nil {
		return nil, err
	}

	return &Ed25519Identity{sk}, nil
}

// PublicIdentity converts public identity of the Ed25519 to be used in chain information
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
