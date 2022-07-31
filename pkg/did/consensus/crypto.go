package consensus

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign/bls"
)

func Verify(msg, signature []byte, from peer.ID, key kyber.Point) error {
	d, err := from.MarshalBinary()
	if err != nil {
		return errors.Wrap(err, "marshaling id")
	}

	d = append(d, msg...)

	err = bls.Verify(bn256.NewSuite(), key, d, signature)
	if err != nil {
		return err
	}

	return nil
}

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
