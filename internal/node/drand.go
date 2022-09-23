package node

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"strings"

	"github.com/drand/drand/chain"
	"github.com/drand/drand/client"
	"github.com/drand/drand/client/http"
	"github.com/drand/drand/log"
	pubsubClient "github.com/drand/drand/lp2p/client"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
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
)

const (
	nthTick = 10 //10*30s=ever 5 minutes
)

func newDrandClient(ps *pubsub.PubSub) (client.Client, error) {
	logger := log.NewKitLoggerFrom(log.LoggerTo(logging.Entry().WithField("component", "drand").WriterLevel(logrus.DebugLevel)))

	cinfo, err := chain.InfoFromJSON(strings.NewReader(drandChainInfo))
	if err != nil {
		return nil, errors.Wrap(err, "reading chain info")
	}

	logging.Entry().Debugf("drand chain info %+v", cinfo)

	c, err := client.New(
		client.WithChainInfo(cinfo),
		client.WithAutoWatch(),
		client.WithLogger(logger),
		pubsubClient.WithPubsub(ps),
		client.From(http.ForURLs(urls, drandHash)...),
	)
	if err != nil {
		return nil, errors.Wrap(err, "constructing drand client")
	}

	logging.Entry().Info("drand client initiated successfully")

	return c, nil
}

func (n *Node) RandomSource() <-chan int64 {
	dstCh := make(chan int64)
	srcCh := make(<-chan client.Result)

	if n.drand == nil {
		logging.Entry().Fatal("network beacon source not implemented yet")
	}
	//TODO(tcfw) support random beacon from pubsub directly

	logging.Entry().Debug("Using drand as beacon source")

	srcCh = n.drand.Watch(context.Background())

	go func() {
		logging.Entry().Debug("starting random source")

		for tick := range srcCh {
			if tick.Round()%nthTick != 0 {
				logging.Entry().WithField("t", nthTick-(tick.Round()%nthTick)).Debug("got beacon, but is not tick")
				continue
			}

			randomness := tick.Randomness()
			var v uint64
			binary.BigEndian.PutUint64(randomness, v)

			select {
			case dstCh <- int64(v):
				logging.Entry().Debug("beacon consumed")
			default:
			}
		}
	}()

	return dstCh

}
