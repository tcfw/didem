package node

import (
	"context"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
)

type Node interface {
	P2P() P2P
}

type P2P interface {
	Connect(context.Context, peer.AddrInfo) error
	Open(context.Context, peer.ID, protocol.ID) (network.Stream, error)
	FindProvider(context.Context, cid.Cid) ([]peer.AddrInfo, error)
}
