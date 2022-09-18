package consensus

import (
	"github.com/drand/kyber"
	"github.com/tcfw/didem/pkg/cryptography"
)

// signatureData takes in a new message and removes the signature if any
// and marshals the message to be used when creating a signature of the
// message
func signatureData(msg *Msg) ([]byte, error) {
	sig := msg.Signature
	msg.Signature = nil
	d, err := msg.Marshal()
	if err != nil {
		return nil, err
	}

	msg.Signature = sig
	return d, nil
}

func AggregateSignatures(sigs ...[]byte) ([]byte, error) {
	return cryptography.AggregateBls12381Signatures(sigs...)
}

func AggregatePublicKeys(pk ...*cryptography.Bls12381PublicKey) kyber.Point {
	return cryptography.AggregateBls12381PublicKeys(pk...)
}
