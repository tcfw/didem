package consensus

import (
	"math/big"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
)

type Consensus struct {
	beacon <-chan int64

	db     Db
	logger logrus.Logger
}

func (c *Consensus) Proposer() <-chan peer.ID {
	bCh := make(chan peer.ID)

	go func() {
		for b := range c.beacon {
			nodes, err := c.db.Nodes()
			if err != nil {
				c.logger.WithError(err).Error("getting node list")
				continue
			}

			m := big.NewInt(0)
			m.SetInt64(b)
			mn := m.Mod(m, big.NewInt(int64(len(nodes)))).Int64()

			p := nodes[mn]
			bCh <- p
		}
	}()

	return bCh
}
