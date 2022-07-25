//go:generate go run github.com/vektra/mockery/v2 --name Db

package consensus

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"github.com/tcfw/didem/pkg/tx"
)

type Db interface {
	Nodes() ([]peer.ID, error)
	Node(peer.ID) (*tx.Node, error)
}

type dbImpl struct{}

func (db *dbImpl) Nodes() ([]peer.ID, error) {
	return nil, errors.New("not impl")
}

func (db *dbImpl) Node(id peer.ID) (*tx.Node, error) {
	return nil, errors.New("not impl")
}
