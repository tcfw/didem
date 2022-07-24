package consensus

import (
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"
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
	return msgpack.Marshal(msg)
}
