package consensus

import (
	"context"
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"

	"github.com/tcfw/didem/pkg/did/w3cdid"
	"github.com/tcfw/didem/pkg/storage"
	"github.com/tcfw/didem/pkg/tx"
)

func TestProposerFromBeacon(t *testing.T) {
	beacon := make(chan int64)

	peers := []peer.ID{
		peer.ID("peer_1"),
		peer.ID("peer_2"),
		peer.ID("peer_3"),
	}

	blockStore := storage.NewMemStore()
	txs := []cid.Cid{}

	for _, p := range peers {
		ntx := &tx.Tx{
			Version: tx.Version1,
			Ts:      time.Now().Unix(),
			Type:    tx.TxType_Node,
			Action:  tx.TxActionAdd,
			Data: &tx.Node{
				Id: string(p),
			},
		}
		cid, err := blockStore.PutTx(context.Background(), ntx)
		if err != nil {
			t.Fatal(err)
		}

		txs = append(txs, cid)
	}

	txset, err := storage.NewTxSet(blockStore, txs)
	if err != nil {
		t.Fatal(err)
	}

	b := &storage.Block{
		TxRoot: txset.Cid(),
	}

	bcid, err := blockStore.PutBlock(context.Background(), b)
	if err != nil {
		t.Fatal(err)
	}

	if err := blockStore.MarkBlock(context.Background(), storage.BlockID(bcid), storage.BlockStateValidated); err != nil {
		t.Fatal(err)
	}

	if err := blockStore.MarkBlock(context.Background(), storage.BlockID(bcid), storage.BlockStateAccepted); err != nil {
		t.Fatal(err)
	}

	c := &Consensus{
		beacon: beacon,
		store:  blockStore,
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
	assert.Equal(t, uint64(2), msg.Consensus.NewRound.Height)
	assert.Equal(t, uint32(1), msg.Consensus.NewRound.Round)

	//Propsal
	msg = <-sub
	assert.Equal(t, MsgTypeConsensus, msg.Type)
	assert.Equal(t, ConsensusMsgTypeProposal, msg.Consensus.Type)
	assert.Equal(t, uint64(2), msg.Consensus.Proposal.Height)
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
		Type:    tx.TxType_DID,
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

	blockVoteGracePeriod = 10 * time.Millisecond

	msg := <-sub
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
	assert.Equal(t, VoteTypePreVote, msg.Consensus.Vote.Type)
	prop.onVote(msg, msg.From)

	msg = <-sub
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
	assert.Equal(t, VoteTypePreVote, msg.Consensus.Vote.Type)
	prop.onVote(msg, msg.From)

	msg = <-sub
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
	assert.Equal(t, VoteTypePreCommit, msg.Consensus.Vote.Type)
	prop.onVote(msg, msg.From)

	msg = <-sub
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
	assert.Equal(t, VoteTypePreCommit, msg.Consensus.Vote.Type)
	prop.onVote(msg, msg.From)

	assert.Len(t, prop.propsalState.PreVotes, 2)
	assert.Len(t, prop.propsalState.PreCommits, 2)
	assert.Len(t, prop.propsalState.PreVotesEvidence, 2)
	assert.Len(t, prop.propsalState.PreCommitsEvidence, 2)
	assert.Equal(t, block, prop.propsalState.Step)

	time.Sleep(20 * time.Millisecond)

	assert.Equal(t, block, prop.propsalState.Step)
	assert.Equal(t, uint64(2), prop.state.Height)
	assert.Equal(t, uint32(1), prop.propsalState.Round)

	validator := instances[1]
	for validator.state.Height == 1 {
		time.Sleep(1 * time.Millisecond)
	}

	assert.Equal(t, uint64(2), prop.state.Height)
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
	assert.Equal(t, VoteTypePreCommit, msg.Consensus.Vote.Type)
	assert.Equal(t, uint32(1), msg.Consensus.Vote.Round)
	prop.onVote(msg, msg.From)

	msg = <-sub
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
	assert.Equal(t, VoteTypePreCommit, msg.Consensus.Vote.Type)
	assert.Equal(t, uint32(1), msg.Consensus.Vote.Round)
	prop.onVote(msg, msg.From)

	for i := uint32(2); i < 5; i++ {
		if err := prop.StartRound(true); err != nil {
			t.Fatal(err)
		}

		msg = <-sub
		assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
		assert.Equal(t, VoteTypePreVote, msg.Consensus.Vote.Type)
		assert.Equal(t, i, msg.Consensus.Vote.Round)
		prop.onVote(msg, msg.From)

		msg = <-sub
		assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
		assert.Equal(t, VoteTypePreVote, msg.Consensus.Vote.Type)
		assert.Equal(t, i, msg.Consensus.Vote.Round)
		prop.onVote(msg, msg.From)

		msg = <-sub
		assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
		assert.Equal(t, VoteTypePreCommit, msg.Consensus.Vote.Type)
		assert.Equal(t, i, msg.Consensus.Vote.Round)
		prop.onVote(msg, msg.From)

		msg = <-sub
		assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
		assert.Equal(t, VoteTypePreCommit, msg.Consensus.Vote.Type)
		assert.Equal(t, i, msg.Consensus.Vote.Round)
		prop.onVote(msg, msg.From)
	}
}

func TestTimeoutProposal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	timeoutPropose = 20 * time.Millisecond
	timeoutPrevote = 20 * time.Millisecond
	timeoutPrecommit = 20 * time.Millisecond

	t.Cleanup(func() {
		timeoutPropose = 1 * time.Minute
		timeoutPrevote = 1 * time.Minute
		timeoutPrecommit = 1 * time.Minute
	})

	_, instances, _ := newConsensusPubSubNet(t, ctx, 2)

	prop := instances[0]

	sub, err := prop.p2p.Msgs(pubsubMsgsChanName)
	if err != nil {
		t.Fatal(err)
	}

	setProposer(t, instances, prop.id)

	startAll(t, instances[1:])

	//fail to start proposer
	for _, instance := range instances[1:] {
		if err := instance.StartRound(false); err != nil {
			t.Fatal(err)
		}
	}

	//proposer should receive an empty vote for missing proposal
	msg := <-sub
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
	assert.Equal(t, VoteTypePreVote, msg.Consensus.Vote.Type)
	assert.Equal(t, uint32(1), msg.Consensus.Vote.Round)
	assert.Equal(t, "", msg.Consensus.Vote.BlockID)

	//proposer should receive an empty vote for missing prevotes
	msg = <-sub
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
	assert.Equal(t, VoteTypePreCommit, msg.Consensus.Vote.Type)
	assert.Equal(t, uint32(1), msg.Consensus.Vote.Round)
	assert.Equal(t, "", msg.Consensus.Vote.BlockID)

	//proposer should receive a new empty vote for missing precommits
	msg = <-sub
	assert.Equal(t, ConsensusMsgTypeVote, msg.Consensus.Type)
	assert.Equal(t, VoteTypePreVote, msg.Consensus.Vote.Type)
	assert.Equal(t, uint32(2), msg.Consensus.Vote.Round)
	assert.Equal(t, "", msg.Consensus.Vote.BlockID)
}
