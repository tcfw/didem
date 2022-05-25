package node

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tcfw/didem/internal/storage"
	storageIface "github.com/tcfw/didem/pkg/storage"
)

type NodeOption func(*Node) error

func WithStorage(s storageIface.Storage) NodeOption {
	return func(n *Node) error {
		n.storage = s
		return nil
	}
}

func WithLogger(l *logrus.Logger) NodeOption {
	return func(n *Node) error {
		n.logger = l
		return nil
	}
}

func WithDefaultOptions(ctx context.Context) NodeOption {
	return func(n *Node) error {
		n.logger = logrus.StandardLogger()

		ipfs, err := storage.NewIPFSStorage(ctx)
		if err != nil {
			return errors.Wrap(err, "initing storage")
		}
		n.storage = ipfs

		return nil
	}
}
