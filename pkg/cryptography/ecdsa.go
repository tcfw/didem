package cryptography

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha512"
	"io"

	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

type Secp256k1PrivateKey struct {
	*ecdsa.PrivateKey
}

func NewEcdsaSecp256k1PrivateKey() (*Secp256k1PrivateKey, error) {
	pk, err := ecdsa.GenerateKey(ethCrypto.S256(), rand.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "generating ecdsa key")
	}

	return &Secp256k1PrivateKey{pk}, nil
}

func (p *Secp256k1PrivateKey) Bytes() ([]byte, error) {
	return ethCrypto.FromECDSA(p.PrivateKey), nil
}

func (p *Secp256k1PrivateKey) Sign(_ io.Reader, digest []byte, _ crypto.SignerOpts) ([]byte, error) {
	dig := digest

	if len(dig) > 128 {
		h := sha512.Sum512(digest)
		dig = h[:]
	}

	return ethCrypto.Sign(dig, p.PrivateKey)
}

func (p *Secp256k1PrivateKey) Public() crypto.PublicKey {
	return Secp256k1PublicKey{p.PublicKey}
}

func NewSecp256k1PublicKey(d []byte) (*Secp256k1PublicKey, error) {
	pub, err := ethCrypto.UnmarshalPubkey(d)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling ecdsa pub key")
	}

	return &Secp256k1PublicKey{*pub}, nil
}

type Secp256k1PublicKey struct {
	ecdsa.PublicKey
}

func (p *Secp256k1PublicKey) Bytes() ([]byte, error) {
	return ethCrypto.FromECDSAPub(&p.PublicKey), nil
}

func (p *Secp256k1PublicKey) Verify(sig, msg []byte) (bool, error) {
	return ethCrypto.VerifySignature(
		ethCrypto.FromECDSAPub(&p.PublicKey),
		msg,
		sig,
	), nil
}

func ValidateEcdsaSecp256k1(vm VerificationMethod, signature []byte, msg []byte) (bool, error) {
	dig := msg

	if len(dig) > 128 {
		h := sha512.Sum512(msg)
		dig = h[:]
	}

	pkbytes, err := decodeMultibase(vm.PublicKeyMultibase)
	if err != nil {
		return false, errors.Wrap(err, "decoding multibase")
	}

	pub, err := NewSecp256k1PublicKey(pkbytes)
	if err != nil {
		return false, errors.Wrap(err, "unmarshalling public key")
	}

	return pub.Verify(signature, dig)
}
