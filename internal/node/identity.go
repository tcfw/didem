package node

import (
	"context"
	"crypto/rand"
	"io/ioutil"
	"os"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tcfw/didem/internal/config"
)

func getIdentity(ctx context.Context, cfg *config.Config, l *logrus.Logger) (libp2p.Option, error) {
	id := cfg.P2P().IdentityFile
	_, err := os.Stat(id)
	if errors.Is(err, os.ErrNotExist) {
		if err := generateIdentity(ctx, cfg, l); err != nil {
			return nil, errors.Wrap(err, "creating new identity")
		}
	} else if err != nil {
		return nil, errors.Wrap(err, "checking identity file")
	} else {
		l.Debugf("using existing Ed25519 identity")
	}

	idB, err := ioutil.ReadFile(id)
	if err != nil {
		return nil, errors.Wrap(err, "reading identity file")
	}

	priv, err := crypto.UnmarshalPrivateKey(idB)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling private key")
	}

	return libp2p.Identity(priv), nil
}

func generateIdentity(ctx context.Context, cfg *config.Config, l *logrus.Logger) error {
	l.Debugf("creating a new Ed25519 identity")
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 0, rand.Reader)
	if err != nil {
		return errors.Wrap(err, "generating priv key")
	}

	b, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return errors.Wrap(err, "marshaling new private key")
	}

	return ioutil.WriteFile(cfg.P2P().IdentityFile, b, 0600)
}
