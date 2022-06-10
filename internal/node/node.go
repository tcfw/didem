package node

import (
	"context"
	"sync"

	"github.com/libp2p/go-libp2p-core/event"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tcfw/didem/internal/config"
	"github.com/tcfw/didem/internal/did"
	"github.com/tcfw/didem/internal/em"
	didIface "github.com/tcfw/didem/pkg/did"
	"github.com/tcfw/didem/pkg/storage"

	nodeIface "github.com/tcfw/didem/pkg/node"
)

type Node struct {
	p2p              *p2pHost
	storage          storage.Storage
	identityResolver didIface.Resolver
	idStore          didIface.IdentityStore

	handlers map[protocol.ID]interface{}

	logger *logrus.Logger
}

func (n *Node) Storage() storage.Storage {
	return n.storage
}

func (n *Node) P2P() nodeIface.P2P {
	return n.p2p
}

func (n *Node) ID() didIface.IdentityStore {
	return n.idStore
}

func (n *Node) Resolver() didIface.Resolver {
	return n.identityResolver
}

func (n *Node) Did() *did.Handler {
	return n.handlers[did.ProtocolID].(*did.Handler)
}

func (n *Node) Em() *em.Handler {
	return n.handlers[em.ProtocolID].(*em.Handler)
}

func NewNode(ctx context.Context, opts ...NodeOption) (*Node, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	n := &Node{
		handlers: make(map[protocol.ID]interface{}),
	}

	for _, opt := range opts {
		if err := opt(n); err != nil {
			return nil, err
		}
	}

	n.p2p, err = newP2PHost(ctx, n, cfg)
	if err != nil {
		return nil, err
	}

	go n.watchEvents()

	if err := n.setupStreamHandlers(); err != nil {
		return nil, errors.Wrap(err, "attaching stream handlers")
	}

	if err := n.bootstrap(ctx, cfg); err != nil {
		return nil, errors.Wrap(err, "bootstrapping p2p")
	}

	return n, nil
}

func (n *Node) watchEvents() {
	sub, err := n.p2p.host.EventBus().Subscribe(event.WildcardSubscription)
	if err != nil {
		n.logger.WithError(err).Error("subscribing to p2p events")
		return
	}

	defer sub.Close()
	for e := range sub.Out() {
		switch eventType := e.(type) {
		case event.EvtLocalAddressesUpdated:
			evt := e.(event.EvtLocalAddressesUpdated)
			for _, addr := range evt.Current {
				if addr.Action != event.Maintained {
					actionStr := "added"
					if addr.Action == event.Removed {
						actionStr = "removed"
					}
					n.logger.WithField("addr", addr.Address.String()).WithField("action", actionStr).Info("updated reachability")
				}
			}
		default:
			n.logger.WithField("event", e).Debugf("unknown event %T", eventType)

		}
	}
}

func (n *Node) ListenAndServe() error {
	n.logger.WithField("addrs", n.p2p.host.Addrs()).WithField("id", n.p2p.host.ID().String()).Info("Starting listening")

	select {}
}

func (n *Node) Stop() error {
	n.logger.Warn("Shutting down")

	return nil
}

func (n *Node) bootstrap(ctx context.Context, cfg *config.Config) error {
	n.logger.Debugf("bootstrapping P2P host")

	peers := cfg.P2P().BootstrapPeers
	if len(peers) == 0 {
		n.logger.Debug("no bootstrapping peers")
	}

	var wg sync.WaitGroup

	for _, peerAddr := range peers {
		ma, err := multiaddr.NewMultiaddr(peerAddr)
		if err != nil {
			return errors.Wrap(err, "parsing bootstrap multiaddr")
		}

		peerinfo, _ := peer.AddrInfoFromP2pAddr(ma)
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := n.p2p.host.Connect(ctx, *peerinfo); err != nil {
				n.logger.WithField("peer", peerinfo.String()).WithError(err).Warning("failed to connect to bootstrap peer")
			} else {
				n.logger.Debug("Connection established with bootstrap peer:", *peerinfo)
			}
		}()
	}
	wg.Wait()

	return nil
}
