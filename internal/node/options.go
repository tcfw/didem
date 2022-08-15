package node

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tcfw/didem/internal/storage"
	"github.com/tcfw/didem/pkg/did"
	storageIface "github.com/tcfw/didem/pkg/storage"
)

type NodeOption func(*Node) error

func WithStorage(s storageIface.Store) NodeOption {
	return func(n *Node) error {
		n.storage = s
		return nil
	}
}

func WithDefaultOptions(ctx context.Context) NodeOption {
	return func(n *Node) error {
		idFile := n.cfg.P2P().IdentityFile
		repo := n.cfg.P2P().IpfsRepo

		ident, err := identityToIPFSConfigIdentity(idFile)
		if err != nil {
			return errors.Wrap(err, "converting ident from config")
		}

		ipfs, err := storage.NewIPFSStorage(ctx, ident, repo)
		if err != nil {
			return errors.Wrap(err, "initing storage")
		}
		n.storage = ipfs

		return nil
	}
}

func WithIdentityStore(store did.IdentityStore) NodeOption {
	return func(n *Node) error {
		n.idStore = store
		return nil
	}
}
