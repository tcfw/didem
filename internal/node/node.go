package node

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tcfw/didem/internal/config"
	"github.com/tcfw/didem/pkg/storage"
)

type Node struct {
	p2p     *p2pHost
	storage storage.Storage
	logger  *logrus.Logger
}

func (n *Node) Storage() storage.Storage {
	return n.storage
}

func (n *Node) P2P() *p2pHost {
	return n.p2p
}

func NewNode(ctx context.Context, opts ...NodeOption) (*Node, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	n := &Node{}

	p2phost, err := newP2PHost(ctx, cfg)
	if err != nil {
		return nil, err
	}
	n.p2p = p2phost

	for _, opt := range opts {
		if err := opt(n); err != nil {
			return nil, err
		}
	}

	if err := n.setupStreamHandlers(); err != nil {
		return nil, errors.Wrap(err, "attaching stream handlers")
	}

	return n, nil
}

func (n *Node) ListenAndServe() error {
	n.logger.WithField("addrs", n.p2p.host.Addrs()).WithField("id", n.p2p.host.ID().String()).Info("Starting listening")

	select {}
}

func (n *Node) Stop() error {
	n.logger.Warn("Shutting down")

	return nil
}
