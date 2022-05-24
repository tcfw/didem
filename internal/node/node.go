package node

import (
	"github.com/sirupsen/logrus"
	"github.com/tcfw/didem/pkg/storage"
)

type Node struct {
	storage storage.Storage
	logger  *logrus.Logger
}

func (n *Node) Storage() storage.Storage {
	return n.storage
}

func NewNode(opts ...NodeOption) (*Node, error) {
	n := &Node{}

	for _, opt := range opts {
		if err := opt(n); err != nil {
			return nil, err
		}
	}

	return n, nil
}

func (n *Node) ListenAndServe() error {
	n.logger.Info("Starting listening")
	select {}
	return nil
}

func (n *Node) Stop() error {
	n.logger.Warn("Shutting down")

	return nil
}
