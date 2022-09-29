package node

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"strings"
	"time"

	"github.com/drand/drand/chain"
	"github.com/drand/drand/client"
	"github.com/drand/drand/client/http"
	"github.com/drand/drand/log"
	pubsubClient "github.com/drand/drand/lp2p/client"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	dnsaddr "github.com/multiformats/go-multiaddr-dns"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tcfw/didem/internal/utils/logging"
)

var urls = []string{
	"https://api.drand.sh",
	"https://api2.drand.sh",
	"https://api3.drand.sh",
	"https://drand.cloudflare.com",
}

var (
	drandChainInfo = `{
		"public_key":"868f005eb8e6e4ca0a47c8a77ceaa5309a47978a7c71bc5cce96366b5d7a569937c529eeda66c7293784a9402801af31",
		"period":30,
		"genesis_time":1595431050,
		"hash":"8990e7a9aaed2ffed73dbd7092123d6f289930540d7651336225dc172e51b2ce",
		"groupHash":"176f93498eac9ca337150b46d21dd58673ea4e3581185f869672e59fa4cb390a"
	}`

	drandHash, _ = hex.DecodeString("8990e7a9aaed2ffed73dbd7092123d6f289930540d7651336225dc172e51b2ce")

	drandGSBootstrapPeers = []string{
		"/dnsaddr/api.drand.sh",
		"/dnsaddr/api2.drand.sh",
		"/dnsaddr/api3.drand.sh",
	}
)

const (
	nthTick = 10 //10*30s=ever 5 minutes
)

func newDrandClient(h *p2pHost, raw bool) (client.Client, error) {
	logger := log.NewKitLoggerFrom(log.LoggerTo(logging.Entry().WithField("component", "drand").WriterLevel(logrus.DebugLevel)))

	cinfo, err := chain.InfoFromJSON(strings.NewReader(drandChainInfo))
	if err != nil {
		return nil, errors.Wrap(err, "reading chain info")
	}

	opts := []client.Option{
		client.WithChainInfo(cinfo),
		client.WithAutoWatch(),
		client.WithLogger(logger),
		pubsubClient.WithPubsub(h.pubsub),
	}

	if raw {
		opts = append(opts, client.From(http.ForURLs(urls, drandHash)...))
	} else {
		logging.Entry().Warn("pubsub beacon source is experimental")

		for _, p := range drandGSBootstrapPeers {
			go func(p string) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				mas, err := dnsaddr.Resolve(ctx, multiaddr.StringCast(p))
				if err != nil {
					logging.WithError(err).Error(err)
					return
				}

				for _, ma := range mas {
					pi, err := peer.AddrInfoFromP2pAddr(ma)
					if err != nil {
						logging.WithError(err).Error(err)
						return
					}

					h.connMgr.TagPeer(pi.ID, "drand", 0)

					if err := h.Connect(context.Background(), *pi); err != nil {
						logging.WithError(err).Error(err)
						return
					}
				}
			}(p)
		}
	}

	c, err := client.New(
		opts...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "constructing drand client")
	}

	logging.Entry().Info("drand client initiated successfully")

	return c, nil
}

func (n *Node) RandomSource() <-chan uint64 {
	dstCh := make(chan uint64)
	srcCh := make(chan client.Result)

	if n.drand != nil {
		logging.Entry().Debug("Using drand as beacon source")

		go func() {
			for b := range n.drand.Watch(context.Background()) {
				srcCh <- b
			}
		}()
		// } else {
		// 	logging.Entry().Warn("Pubsub drand beacons are experimental")

		// 	go func() {
		// 		pubsubTopic := fmt.Sprintf("/drand/pubsub/v0.0.0/%s", hex.EncodeToString(drandHash))

		// 		t, err := n.p2p.pubsub.Join(pubsubTopic)
		// 		if err != nil {
		// 			logging.WithError(err).Error("getting drand topic")
		// 			return
		// 		}

		// 		sub, err := t.Subscribe()
		// 		if err != nil {
		// 			logging.WithError(err).Error("subbing to drand")
		// 		}

		// 		for {
		// 			msg, err := sub.Next(context.Background())
		// 			if err != nil {
		// 				logging.WithError(err).Error("pubsub drand msg")
		// 			}

		// 			var rand drand.PublicRandResponse
		// 			err = proto.Unmarshal(msg.Data, &rand)
		// 			if err != nil {
		// 				logging.WithError(err).Warn("gossip client", "unmarshal random error", "err", err)
		// 				continue
		// 			}

		// 			logging.Entry().Warn("got successful pubsub drand beacon")

		// 			srcCh <- &client.RandomData{
		// 				Rnd:               rand.Round,
		// 				Random:            rand.Randomness,
		// 				Sig:               rand.Signature,
		// 				PreviousSignature: rand.PreviousSignature,
		// 			}
		// 		}
		// 	}()
	}

	go func() {
		logging.Entry().Debug("starting random source")

		for tick := range srcCh {
			if tick.Round()%nthTick != 0 {
				logging.Entry().WithField("t", nthTick-(tick.Round()%nthTick)).Debug("got beacon, but is not tick")
				continue
			}

			randomness := tick.Randomness()
			v := binary.BigEndian.Uint64(randomness)

			select {
			case dstCh <- v:
				logging.Entry().Debug("beacon consumed")
			default:
			}
		}
	}()

	return dstCh

}
