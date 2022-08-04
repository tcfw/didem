package consensus

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	bhost "github.com/libp2p/go-libp2p/p2p/host/blank"
	swarmt "github.com/libp2p/go-libp2p/p2p/net/swarm/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tcfw/didem/pkg/did/consensus/mocks"
	"github.com/tcfw/didem/pkg/storage"
	"github.com/tcfw/didem/pkg/tx"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign/bls"
	"go.dedis.ch/kyber/v3/util/random"
)

func TestCheckBn256PointMarhsaling(t *testing.T) {
	_, pk := bls.NewKeyPair(bn256.NewSuite(), random.New())

	b, err := pk.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	p := bn256.NewSuite().G2().Point()
	p.UnmarshalBinary(b)

	assert.True(t, pk.Equal(p))
}

func newConsensusPubSubNet(t *testing.T, ctx context.Context, n int) ([]host.Host, []*Consensus, MemPool) {
	hosts := getNetHosts(t, ctx, n)

	psubs := getGossipsubs(ctx, hosts)

	peers := []peer.ID{}
	instances := []*Consensus{}

	sks := make([]kyber.Scalar, 0, n)
	pks := make([]kyber.Point, 0, n)

	for i := 0; i < n; i++ {
		sk, pk := bls.NewKeyPair(bn256.NewSuite(), random.New())
		sks = append(sks, sk)
		pks = append(pks, pk)
	}

	db := mocks.NewDb(t)
	db.On("Nodes").Maybe().Return(peers, nil)
	nodeRet := func(id peer.ID) *tx.Node {
		for i, h := range hosts {
			b, _ := pks[i].MarshalBinary()
			if h.ID().String() == id.String() {
				return &tx.Node{
					Id:  h.ID().String(),
					Did: fmt.Sprintf("did:didem:%s", h.ID().String()),
					Key: b,
				}
			}
		}
		return nil
	}

	db.On("Node", mock.Anything).Maybe().Return(nodeRet, nil)

	blockStore := storage.NewMemStore()
	memPool := NewTxMemPool()

	chain := make([]byte, 10)
	rand.Read(chain)

	for i, h := range hosts {
		peers = append(peers, h.ID())
		c := &Consensus{
			id:         h.ID(),
			chain:      chain,
			signingKey: sks[i],
			memPool:    memPool,
			blockStore: blockStore,
			validator:  storage.NewTxValidator(blockStore),
			db:         db,
			p2p: &p2p{
				self:   h.ID(),
				router: psubs[i],
				topics: make(map[string]*pubsub.Topic),
			},
		}
		c.setupTimers()
		t.Cleanup(func() {
			stopTimer(c.timerPropose)
			stopTimer(c.timerPrevote)
			stopTimer(c.timerPrecommit)
			stopTimer(c.timerBlock)
			select {
			case c.stopTimerLoop <- struct{}{}:
			default:
			}
		})
		instances = append(instances, c)
	}

	switch {
	case n <= 5:
		connectAll(t, hosts)
	case n <= 10:
		denseConnect(t, hosts)
	case n > 10:
		sparseConnect(t, hosts)
	}

	return hosts, instances, memPool
}

func getNetHosts(t *testing.T, ctx context.Context, n int) []host.Host {
	var out []host.Host

	for i := 0; i < n; i++ {
		netw := swarmt.GenSwarm(t)
		h := bhost.NewBlankHost(netw)
		t.Cleanup(func() { h.Close() })
		out = append(out, h)
	}

	return out
}

func getGossipsub(ctx context.Context, h host.Host, opts ...pubsub.Option) *pubsub.PubSub {
	ps, err := pubsub.NewGossipSub(ctx, h, opts...)
	if err != nil {
		panic(err)
	}
	return ps
}

func getGossipsubs(ctx context.Context, hs []host.Host, opts ...pubsub.Option) []*pubsub.PubSub {
	var psubs []*pubsub.PubSub
	for _, h := range hs {
		psubs = append(psubs, getGossipsub(ctx, h, opts...))
	}
	return psubs
}

func connect(t *testing.T, a, b host.Host) {
	pinfo := a.Peerstore().PeerInfo(a.ID())
	err := b.Connect(context.Background(), pinfo)
	if err != nil {
		t.Fatal(err)
	}
}

func sparseConnect(t *testing.T, hosts []host.Host) {
	connectSome(t, hosts, 3)
}

func denseConnect(t *testing.T, hosts []host.Host) {
	connectSome(t, hosts, 10)
}

func connectSome(t *testing.T, hosts []host.Host, d int) {
	for i, a := range hosts {
		for j := 0; j < d; j++ {
			n := rand.Intn(len(hosts))
			if n == i {
				j--
				continue
			}

			b := hosts[n]

			connect(t, a, b)
		}
	}
}

func connectAll(t *testing.T, hosts []host.Host) {
	for i, a := range hosts {
		for j, b := range hosts {
			if i == j {
				continue
			}

			connect(t, a, b)
		}
	}
}

func setProposer(t *testing.T, instances []*Consensus, id peer.ID) {
	for _, instance := range instances {
		instance.propsalState.Proposer = id
	}
}

func startAll(t *testing.T, instances []*Consensus) {
	for _, instance := range instances {
		if err := instance.Start(); err != nil {
			t.Fatal(err)
		}
	}
}

func setF(t *testing.T, instances []*Consensus, f uint64) {
	for _, instance := range instances {
		instance.propsalState.f = f
	}
}
