//go:generate go run github.com/vektra/mockery/v2 --name Db

package consensus

import (
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
)

type Db interface {
	Nodes() ([]peer.ID, error)
	Node(peer.ID) (*Node, error)
}

type Node struct {
	Id   peer.ID
	Did  string
	Keys []crypto.PubKey
}

type dbImpl struct{}

func (db *dbImpl) Nodes() ([]peer.ID, error) {
	return nil, errors.New("not impl")
}

func (db *dbImpl) Node(id peer.ID) (*Node, error) {
	return nil, errors.New("not impl")
}
