package node

import (
	"context"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	connmgriFace "github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"

	"github.com/tcfw/didem/internal/config"
	"github.com/tcfw/didem/pkg/did"
)

func newP2PHost(ctx context.Context, n *Node, cfg *config.Config) (*p2pHost, error) {
	var err error
	h := &p2pHost{
		n: n,
	}

	id, err := getIdentity(ctx, cfg, n.logger)
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

	opts := []libp2p.Option{
		id,
		listeningAddrs,
		libp2p.DefaultTransports,
		libp2p.DefaultResourceManager,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.ConnectionManager(h.connMgr),
		libp2p.Peerstore(h.peerStore),
		libp2p.NATPortMap(),
		libp2p.EnableNATService(),
		libp2p.ForceReachabilityPublic(),
	}

	if cfg.P2P().Relay {
		n.logger.Warn("p2p relay enabled")
		opts = append(opts, libp2p.EnableRelay(), libp2p.EnableAutoRelay())
	}

	h.host, err = libp2p.NewWithoutDefaults(opts...)
	if err != nil {
		return nil, errors.Wrap(err, "creating libp2p host")
	}

	h.dht, err = dht.New(ctx, h.host)
	if err != nil {
		return nil, errors.Wrap(err, "initing DHT")
	}

	go func() {
		time.Sleep(2 * time.Second)
		if err := h.dht.Bootstrap(context.Background()); err != nil {
			n.logger.WithError(err).Error("bootstrapping DHT")
		}
	}()

	h.discovery = discovery.NewRoutingDiscovery(h.dht)

	h.pubsub, err = newGossipSub(ctx, cfg, h)
	if err != nil {
		return nil, err
	}

	h.advertiseTimer = *time.NewTicker(1 * time.Minute)
	go h.watchAdvertise(ctx)
	go func() {
		time.Sleep(3 * time.Second)
		h.doAdvertise(ctx)
	}()

	return h, nil
}

func newGossipSub(ctx context.Context, cfg *config.Config, h *p2pHost) (*pubsub.PubSub, error) {
	p, err := pubsub.NewGossipSub(ctx, h.host,
		pubsub.WithPeerExchange(true),
		pubsub.WithStrictSignatureVerification(true),
		pubsub.WithDiscovery(h.discovery),
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

type p2pHost struct {
	host host.Host

	n         *Node
	peerStore peerstore.Peerstore
	connMgr   connmgriFace.ConnManager
	pubsub    *pubsub.PubSub
	dht       *dht.IpfsDHT
	discovery *discovery.RoutingDiscovery

	advertiseTimer time.Ticker
}

func (p *p2pHost) Connect(ctx context.Context, info peer.AddrInfo) error {
	return p.host.Connect(ctx, info)
}

func (p *p2pHost) Open(ctx context.Context, peer peer.ID, protocol protocol.ID) (network.Stream, error) {
	return p.host.NewStream(ctx, peer, protocol)
}

func (p *p2pHost) FindProvider(ctx context.Context, c cid.Cid) ([]peer.AddrInfo, error) {
	return p.dht.FindProviders(ctx, c)
}

func (p *p2pHost) watchAdvertise(ctx context.Context) {
	for range p.advertiseTimer.C {
		go p.doAdvertise(ctx)
	}
}

func (p *p2pHost) doAdvertise(ctx context.Context) {
	p.n.logger.Debug("advertising identities started")

	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	ids, err := p.n.advertisableIdentities(ctx)
	if err != nil {
		p.n.logger.WithError(err).Error("getting ids to advertise")
		return
	}

	var wg sync.WaitGroup

	for _, id := range ids {
		wg.Add(1)

		go func(id did.PrivateIdentity) {
			defer wg.Done()

			if err := p.provideIdentity(ctx, id); err != nil {
				p.n.logger.WithError(err).Error("failed to advertise identity")
			}
		}(id)
	}

	wg.Wait()

	p.n.logger.Debug("advertisement finished")
}

func (p *p2pHost) provideIdentity(ctx context.Context, id did.PrivateIdentity) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pid, err := id.PublicIdentity()
	if err != nil {
		return err
	}

	mh, err := multihash.FromB58String(pid.ID)
	if err != nil {
		return err
	}

	p.n.logger.WithField("id", pid.ID).Debug("advertising identity")

	return p.dht.Provide(ctx, cid.NewCidV1(cid.Raw, mh), true)
}

func (p *p2pHost) Peers() []peer.ID {
	return p.host.Network().Peers()
}
