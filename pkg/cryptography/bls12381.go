package cryptography

import (
	"crypto"
	"io"

	"github.com/drand/kyber"
	bls "github.com/drand/kyber-bls12381"
	sig "github.com/drand/kyber/sign/bls"
	"github.com/drand/kyber/util/random"
	"github.com/pkg/errors"
)

var (
	_ crypto.PrivateKey = (*Bls12381PrivateKey)(nil)
	_ crypto.PublicKey  = (*Bls12381PublicKey)(nil)

	pairing = bls.NewBLS12381Suite()
)

func NewBls12381PrivateKey() *Bls12381PrivateKey {
	return &Bls12381PrivateKey{
		pairing.G1().Scalar().Pick(random.New()),
	}
}

func NewBls12391PrivateKeyFromBytes(b []byte) (*Bls12381PrivateKey, error) {
	k := &Bls12381PrivateKey{
		pairing.G1().Scalar().One(),
	}

	if err := k.sk.UnmarshalBinary(b); err != nil {
		return nil, err
	}

	return k, nil
}

type Bls12381PrivateKey struct {
	sk kyber.Scalar
}

func (b *Bls12381PrivateKey) Sign(_ io.Reader, digest []byte, _ crypto.SignerOpts) (signature []byte, err error) {
	scheme := sig.NewSchemeOnG2(pairing)
	return scheme.Sign(b.sk, digest)
}

func (b *Bls12381PrivateKey) Public() crypto.PublicKey {
	pk := pairing.G1().Point().Mul(b.sk, nil)
	return &Bls12381PublicKey{pk}
}

func (b *Bls12381PrivateKey) Equal(obls crypto.PrivateKey) bool {
	return b.sk.Equal(obls.(*Bls12381PrivateKey).sk)
}

func NewBls12381PublicKey(d []byte) (*Bls12381PublicKey, error) {
	pk := &Bls12381PublicKey{pairing.G1().Point()}
	if err := pk.UnmarshalBinary(d); err != nil {
		return nil, err
	}

	return pk, nil

}

type Bls12381PublicKey struct {
	kyber.Point
}

func (b *Bls12381PublicKey) Bytes() ([]byte, error) {
	return b.MarshalBinary()
}

func (b *Bls12381PublicKey) Verify(signature, msg []byte) (bool, error) {
	scheme := sig.NewSchemeOnG2(pairing)
	if err := scheme.Verify(b.Point, msg, signature); err != nil {
		return false, err
	}

	return true, nil
}

func ValidateBls12381(vm VerificationMethod, signature []byte, msg []byte) (bool, error) {
	pkbytes, err := decodeMultibase(vm.PublicKeyMultibase)
	if err != nil {
		return false, errors.Wrap(err, "decoding multibase")
	}

	pk := &Bls12381PublicKey{pairing.G1().Point()}
	if err := pk.UnmarshalBinary(pkbytes); err != nil {
		return false, err
	}

	return pk.Verify(signature, msg)
}

func AggregateBls12381Signatures(sigs ...[]byte) ([]byte, error) {
	scheme := sig.NewSchemeOnG2(pairing)
	return scheme.AggregateSignatures(sigs...)
}

func AggregateBls12381PublicKeys(pk ...*Bls12381PublicKey) kyber.Point {
	scheme := sig.NewSchemeOnG2(pairing)

	points := make([]kyber.Point, len(pk))

	for i, p := range pk {
		points[i] = p.Point
	}

	return scheme.AggregatePublicKeys(points...)
}
