//go:generate go run github.com/vektra/mockery/v2 --name Db

package consensus

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
)

type Db interface {
	Nodes() ([]peer.ID, error)
}

type dbImpl struct{}

func (db *dbImpl) Nodes() ([]peer.ID, error) {
	return nil, errors.New("not impl")
}
