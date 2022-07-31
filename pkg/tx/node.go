package tx

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"go.dedis.ch/kyber/v3"
)

type Node struct {
	Id   peer.ID
	Did  string
	Keys []kyber.Point
}
