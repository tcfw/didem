package node

import (
	"context"

	"github.com/libp2p/go-libp2p"
	connmgriFace "github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"

	"github.com/tcfw/didem/internal/config"
)

type p2pHost struct {
	host      host.Host
	peerStore peerstore.Peerstore
	connMgr   connmgriFace.ConnManager
	pubsub    *pubsub.PubSub
}

func newP2PHost(ctx context.Context, cfg *config.Config) (*p2pHost, error) {
	var err error
	h := &p2pHost{}

	id, err := getIdentity(ctx, cfg)
	if err != nil {
		return nil, err
	}

	listeningAddrs, err := buildListeningAddrs(ctx, cfg)
	if err != nil {
		return nil, err
	}

	h.connMgr, err = connmgr.NewConnManager(
		cfg.P2P().Connections.PeersCountLow,
		cfg.P2P().Connections.PeersCountHigh,
	)
	if err != nil {
		return nil, err
	}

	h.peerStore, err = pstoremem.NewPeerstore()
	if err != nil {
		return nil, err
	}

	h.host, err = libp2p.NewWithoutDefaults(
		id,
		listeningAddrs,
		libp2p.DefaultTransports,
		libp2p.DefaultResourceManager,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.ConnectionManager(h.connMgr),
		libp2p.Peerstore(h.peerStore),
		libp2p.EnableNATService(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "creating libp2p host")
	}

	h.pubsub, err = newGossipSub(ctx, cfg, h.host)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func newGossipSub(ctx context.Context, cfg *config.Config, h host.Host) (*pubsub.PubSub, error) {
	p, err := pubsub.NewGossipSub(ctx, h,
		pubsub.WithPeerExchange(true),
		pubsub.WithStrictSignatureVerification(true),
	)
	if err != nil {
		return nil, errors.Wrap(err, "creating gossipsub router")
	}

	return p, nil
}

func buildListeningAddrs(ctx context.Context, cfg *config.Config) (libp2p.Option, error) {
	maAddrs := []multiaddr.Multiaddr{}

	for _, addr := range cfg.P2P().ListenAddrs {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		maAddrs = append(maAddrs, maddr)
	}

	return libp2p.ListenAddrs(maAddrs...), nil
}