package consensus

import (
	"context"
	"testing"
	"time"

	peer "github.com/libp2p/go-libp2p-core/peer"

	"github.com/stretchr/testify/assert"

	"github.com/tcfw/didem/pkg/did/consensus/mocks"
	"github.com/tcfw/didem/pkg/did/w3cdid"
	"github.com/tcfw/didem/pkg/tx"
)

func TestProposerFromBeacon(t *testing.T) {
	beacon := make(chan int64)

	peers := []peer.ID{
		peer.ID("peer_1"),
		peer.ID("peer_2"),
		peer.ID("peer_3"),
	}

	db := mocks.NewDb(t)
	db.On("Nodes").Maybe().Return(peers, nil)

	c := &Consensus{
		beacon: beacon,
		db:     db,
	}

	prop := c.proposer()

	//testing modulus yay!
	tests := []struct {
		k int
		v int
	}{
		{3, 0},
		{2, 2},
		{1, 1},
		{0, 0},
	}

	go func() {
		for _, k := range tests {
			beacon <- int64(k.k)
		}
	}()

	for _, v := range tests {
		gotPeer := <-prop
		assert.Equal(t, peers[v.v], gotPeer)
	}
}

func TestReceiveNewRound(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, instances, _ := newConsensusPubSubNet(t, ctx, 3)

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

	_, instances, memPool := newConsensusPubSubNet(t, ctx, 3)

	prop := instances[0]

	sub, err := prop.p2p.Msgs(pubsubMsgsChanName)
	if err != nil {
		t.Fatal(err)
	}

	setProposer(t, instances, prop.id)

	tx := &tx.Tx{
		Version: 1,
		Ts:      time.Now().Unix(),
		Type:    tx.TxType_PKPublish,
		Data:    w3cdid.Document{},
	}
	if err := memPool.AddTx(tx, 0); err != nil {
		t.Fatal(err)
	}

	startAll(t, instances[1:])

	for _, instance := range instances {
		if err := instance.StartRound(false); err != nil {
			t.Fatal(err)
		}
	}

	//2x votes from each node

	msg := <-sub
	assert.Equal(t, MsgTypeConsensus, msg.Type)
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)

	msg = <-sub
	assert.Equal(t, MsgTypeConsensus, msg.Type)
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
}

func TestSuccessfulRound(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, instances, _ := newConsensusPubSubNet(t, ctx, 3)

	prop := instances[0]

	sub, err := prop.p2p.Msgs(pubsubMsgsChanName)
	if err != nil {
		t.Fatal(err)
	}

	setProposer(t, instances, prop.id)

	startAll(t, instances[1:])

	for _, instance := range instances {
		if err := instance.StartRound(false); err != nil {
			t.Fatal(err)
		}
	}

	msg := <-sub
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
	assert.Equal(t, msg.Consensus.Vote.Type, VoteTypePreVote)
	prop.onVote(msg, msg.From)

	msg = <-sub
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
	assert.Equal(t, msg.Consensus.Vote.Type, VoteTypePreVote)
	prop.onVote(msg, msg.From)

	msg = <-sub
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
	assert.Equal(t, msg.Consensus.Vote.Type, VoteTypePreCommit)
	prop.onVote(msg, msg.From)

	msg = <-sub
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
	assert.Equal(t, msg.Consensus.Vote.Type, VoteTypePreCommit)
	prop.onVote(msg, msg.From)
}

func TestNewRound(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, instances, _ := newConsensusPubSubNet(t, ctx, 3)

	prop := instances[0]

	sub, err := prop.p2p.Msgs(pubsubMsgsChanName)
	if err != nil {
		t.Fatal(err)
	}

	setProposer(t, instances, prop.id)

	startAll(t, instances[1:])

	for _, instance := range instances {
		if err := instance.StartRound(false); err != nil {
			t.Fatal(err)
		}
	}

	msg := <-sub
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
	assert.Equal(t, VoteTypePreVote, msg.Consensus.Vote.Type)
	assert.Equal(t, uint32(1), msg.Consensus.Vote.Round)
	prop.onVote(msg, msg.From)

	msg = <-sub
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
	assert.Equal(t, VoteTypePreVote, msg.Consensus.Vote.Type)
	assert.Equal(t, uint32(1), msg.Consensus.Vote.Round)
	prop.onVote(msg, msg.From)

	msg = <-sub
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
	assert.Equal(t, msg.Consensus.Vote.Type, VoteTypePreCommit)
	assert.Equal(t, uint32(1), msg.Consensus.Vote.Round)
	prop.onVote(msg, msg.From)

	msg = <-sub
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
	assert.Equal(t, msg.Consensus.Vote.Type, VoteTypePreCommit)
	assert.Equal(t, uint32(1), msg.Consensus.Vote.Round)
	prop.onVote(msg, msg.From)

	for i := uint32(2); i < 5; i++ {
		if err := prop.StartRound(true); err != nil {
			t.Fatal(err)
		}

		msg = <-sub
		assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
		assert.Equal(t, msg.Consensus.Vote.Type, VoteTypePreVote)
		assert.Equal(t, i, msg.Consensus.Vote.Round)
		prop.onVote(msg, msg.From)

		msg = <-sub
		assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
		assert.Equal(t, msg.Consensus.Vote.Type, VoteTypePreVote)
		assert.Equal(t, i, msg.Consensus.Vote.Round)
		prop.onVote(msg, msg.From)

		msg = <-sub
		assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
		assert.Equal(t, msg.Consensus.Vote.Type, VoteTypePreCommit)
		assert.Equal(t, i, msg.Consensus.Vote.Round)
		prop.onVote(msg, msg.From)

		msg = <-sub
		assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
		assert.Equal(t, msg.Consensus.Vote.Type, VoteTypePreCommit)
		assert.Equal(t, i, msg.Consensus.Vote.Round)
		prop.onVote(msg, msg.From)
	}
}
