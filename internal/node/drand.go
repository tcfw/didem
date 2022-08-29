package node

import (
	"context"
	"encoding/binary"
	"encoding/hex"

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

var chainHash, _ = hex.DecodeString("8990e7a9aaed2ffed73dbd7092123d6f289930540d7651336225dc172e51b2ce")

const (
	nthTick = 10 //10*30s=ever 5 minutes
)

func newDrandClient(ps *pubsub.PubSub) (client.Client, error) {
	logger := log.NewKitLoggerFrom(log.LoggerTo(logging.Entry().WithField("component", "drand").WriterLevel(logrus.DebugLevel)))

	c, err := client.New(
		client.WithLogger(logger),
		pubsubClient.WithPubsub(ps),
		client.WithChainHash(chainHash),
		client.From(http.ForURLs(urls, chainHash)...),
	)
	if err != nil {
		return nil, errors.Wrap(err, "constructing drand client")
	}

	return c, nil
}

func (n *Node) RandomSource() <-chan int64 {
	dstCh := make(chan int64)
	srcCh := make(<-chan client.Result)

	if n.drand != nil {
		logging.Entry().Debug("Using drand as beacon source")
		srcCh = n.drand.Watch(context.Background())
	} else {
		//TODO(tcfw) support random beacon from pubsub directly
		logging.Entry().Fatal("network beacon source not implemented yet")
	}

	go func() {
		logging.Entry().Debug("starting random source")

		for tick := range srcCh {
			if tick.Round()%nthTick != 0 {
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
