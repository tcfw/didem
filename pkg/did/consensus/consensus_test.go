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

	setProposer(t, instances, instances[0].id)

	go func() {
		instances[0].Start()
		if err := instances[0].StartRound(false); err != nil {
			panic(err)
		}
	}()

	//New round
	msg := <-sub
	assert.Equal(t, MsgTypeConsensus, msg.Type)
	assert.Equal(t, ConsensusMsgTypeNewRound, msg.Consensus.Type)
	assert.Equal(t, uint64(1), msg.Consensus.NewRound.Height)
	assert.Equal(t, uint32(1), msg.Consensus.NewRound.Round)

	//Propsal
	msg = <-sub
	assert.Equal(t, MsgTypeConsensus, msg.Type)
	assert.Equal(t, ConsensusMsgTypeProposal, msg.Consensus.Type)
	assert.Equal(t, uint64(1), msg.Consensus.Proposal.Height)
	assert.Equal(t, uint32(1), msg.Consensus.Proposal.Round)
	assert.NotEmpty(t, msg.Consensus.Proposal.BlockID)
	assert.NotEmpty(t, msg.Consensus.Proposal.Timestamp)
}

func TestReceivePrevote(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, instances := newConsensusPubSubNet(t, ctx, 3)

	sub, err := instances[0].p2p.Msgs(pubsubMsgsChanName)
	if err != nil {
		t.Fatal(err)
	}

	setProposer(t, instances, instances[0].id)

	startAll(t, instances[1:])

	for _, instance := range instances {
		if err := instance.StartRound(false); err != nil {
			t.Fatal(err)
		}
	}

	msg := <-sub
	assert.Equal(t, MsgTypeConsensus, msg.Type)
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)

	msg = <-sub
	assert.Equal(t, MsgTypeConsensus, msg.Type)
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
}
