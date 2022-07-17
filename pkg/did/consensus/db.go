package consensus

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
)

type Db struct{}

func (db *Db) Nodes() ([]peer.ID, error) {
	return nil, errors.New("not impl")
}
