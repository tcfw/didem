package consensus

import (
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pkg/errors"
)

func Verify(msg, signature []byte, key crypto.PubKey) error {
	ok, err := key.Verify(msg, signature)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("invalid signature")
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
