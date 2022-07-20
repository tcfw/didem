package consensus

import (
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

	prop := c.Proposer()

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
