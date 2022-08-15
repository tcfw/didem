package node

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"io/ioutil"
	"os"
	"path"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"github.com/tcfw/didem/internal/config"
	"github.com/tcfw/didem/internal/utils/logging"
	"github.com/tcfw/didem/pkg/did"

	ipfsConfig "github.com/ipfs/kubo/config"
)

func getIdentity(ctx context.Context, cfg *config.Config) (libp2p.Option, error) {
	id := cfg.P2P().IdentityFile
	_, err := os.Stat(id)
	if errors.Is(err, os.ErrNotExist) {
		if err := generateIdentity(ctx, cfg); err != nil {
			return nil, errors.Wrap(err, "creating new identity")
		}
	} else if err != nil {
		return nil, errors.Wrap(err, "checking identity file")
	} else {
		logging.Entry().Debugf("using existing Ed25519 identity")
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

func generateIdentity(ctx context.Context, cfg *config.Config) error {
	logging.Entry().Debugf("creating a new Ed25519 identity")

	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 0, rand.Reader)
	if err != nil {
		return errors.Wrap(err, "generating priv key")
	}

	b, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return errors.Wrap(err, "marshaling new private key")
	}

	if err := os.MkdirAll(path.Dir(cfg.P2P().IdentityFile), 0600); err != nil {
		return errors.Wrap(err, "making identity config path")
	}

	return ioutil.WriteFile(cfg.P2P().IdentityFile, b, 0600)
}

func identityToIPFSConfigIdentity(fp string) (ipfsConfig.Identity, error) {
	ident := ipfsConfig.Identity{}

	idB, err := ioutil.ReadFile(fp)
	if err != nil {
		return ident, errors.Wrap(err, "reading identity file")
	}

	sk, err := crypto.UnmarshalPrivateKey(idB)
	if err != nil {
		return ident, errors.Wrap(err, "unmarshaling private key")
	}

	pk := sk.GetPublic()

	skbytes, err := crypto.MarshalPrivateKey(sk)
	if err != nil {
		return ident, err
	}
	ident.PrivKey = base64.StdEncoding.EncodeToString(skbytes)

	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return ident, err
	}
	ident.PeerID = id.Pretty()

	return ident, nil
}

func (n *Node) advertisableIdentities(ctx context.Context) ([]did.PrivateIdentity, error) {
	l, err := n.idStore.List()
	if err != nil {
		return nil, errors.Wrap(err, "getting id list")
	}

	return l, nil
}
