package node

import (
	"context"
	"sync"

	"github.com/libp2p/go-libp2p-core/event"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tcfw/didem/internal/config"
	"github.com/tcfw/didem/pkg/storage"
)

type Node struct {
	p2p     *p2pHost
	storage storage.Storage
	logger  *logrus.Logger
}

func (n *Node) Storage() storage.Storage {
	return n.storage
}

func (n *Node) P2P() *p2pHost {
	return n.p2p
}

func NewNode(ctx context.Context, opts ...NodeOption) (*Node, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	n := &Node{}

	n.p2p, err = newP2PHost(ctx, cfg)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		if err := opt(n); err != nil {
			return nil, err
		}
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
			return err
		}

		peerinfo, _ := peer.AddrInfoFromP2pAddr(ma)
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := n.p2p.host.Connect(ctx, *peerinfo); err != nil {
				n.logger.Warning(err)
			} else {
				n.logger.Debug("Connection established with bootstrap node:", *peerinfo)
			}
		}()
	}
	wg.Wait()

	return nil
}
