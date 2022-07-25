package consensus

import (
	"context"
	"testing"

	peer "github.com/libp2p/go-libp2p-core/peer"

	"github.com/stretchr/testify/assert"

	"github.com/tcfw/didem/pkg/did/consensus/mocks"
)

func TestProposerFromBeacon(t *testing.T) {
	beacon := make(chan int64)

	peers := []peer.ID{
		peer.ID("peer_1"),
		peer.ID("peer_2"),
		peer.ID("peer_3"),
	}

	db := mocks.NewDb(t)
	db.On("Nodes").Return(peers, nil)

	c := &Consensus{
		beacon: beacon,
		db:     db,
	}

	prop := c.proposer()

	//testing modulus yay!
	tests := map[int]int{
		3: 0,
		2: 2,
		1: 1,
		0: 0,
	}

	go func() {
		for k := range tests {
			beacon <- int64(k)
		}
	}()

	for _, v := range tests {
		gotPeer := <-prop
		assert.Equal(t, peers[v], gotPeer)
	}
}

func TestReceiveNewRound(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, instances := newConsensusPubSubNet(t, ctx, 3)

	sub, err := instances[1].p2p.Msgs(pubsubMsgsChanName)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		instances[0].Start()
		if err := instances[0].StartRound(); err != nil {
			panic(err)
		}
	}()

	msg := <-sub

	assert.Equal(t, MsgTypeConsensus, msg.Type)
	assert.Equal(t, uint64(1), msg.Consensus.NewRound.Height)
	assert.Equal(t, uint32(1), msg.Consensus.NewRound.Round)
}
