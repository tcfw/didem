package node

import (
	"context"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/tcfw/didem/internal/config"
	"github.com/tcfw/didem/pkg/did"
	"github.com/tcfw/didem/pkg/storage"
)

type Node interface {
	P2P() P2P
	Resolver() did.Resolver
	ID() did.IdentityStore
	Cfg() *config.Config
	Storage() storage.Store
	RandomSource() <-chan uint64
}

type P2P interface {
	Connect(context.Context, peer.AddrInfo) error
	Open(context.Context, peer.ID, protocol.ID) (network.Stream, error)
	FindProvider(context.Context, cid.Cid) ([]peer.AddrInfo, error)
	Peers() []peer.ID
	Host() host.Host
	PubSub() *pubsub.PubSub
}
