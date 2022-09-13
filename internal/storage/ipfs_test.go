package storage

import (
	"context"
	"encoding/hex"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/kubo/config"
	"github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/assert"
	"github.com/tcfw/didem/pkg/did/w3cdid"
	"github.com/tcfw/didem/pkg/storage"
	"github.com/tcfw/didem/pkg/tx"
)

func TestTypedKey(t *testing.T) {
	k := typedKey(txBlockTPrefix, "a", "a")
	e := []byte{byte(txBlockTPrefix), 'a', tableSep, 'a'}
	assert.Equal(t, e, k)
}

func tempIPFSStorage(t *testing.T, ctx context.Context) *IPFSStorage {
	bootstrapNodes = []string{}
	repoPath, err := ioutil.TempDir("", "ipfs-shell")
	if err != nil {
		t.Fatal(err)
	}

	id, err := config.CreateIdentity(ioutil.Discard, []options.KeyGenerateOption{options.Key.Type(options.Ed25519Key)})
	if err != nil {
		t.Fatal(err)
	}

	ipfs, err := NewIPFSStorage(ctx, id, repoPath)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		os.RemoveAll(repoPath)
		ipfs.Close()
	})

	return ipfs
}

func TestIPFSAdd(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ipfs := tempIPFSStorage(t, ctx)

	tX := &tx.Tx{
		Version: tx.Version1,
		Ts:      time.Now().Unix(),
		Type:    tx.TxType_Node,
		Data: &tx.Node{
			Id:  "1111",
			Did: "did:example:abcdefghijklmnopqrstuvwxyz0123456789",
			Key: nil,
		},
	}

	cid, err := ipfs.PutTx(ctx, tX)
	if err != nil {
		t.Fatal(err)
	}

	txb, _ := tX.Marshal()
	txbh, err := multihash.Sum(txb[:], multihash.SHA2_256, multihash.DefaultLengths[multihash.SHA2_256])
	if err != nil {
		t.Fatal(err)
	}

	expected := hex.EncodeToString(txbh[:])
	idhex := hex.EncodeToString(cid.Bytes())

	assert.Equal(t, expected, idhex[4:])

	txrb, err := ipfs.GetTx(ctx, tx.TxID(cid))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, tX, txrb)
}

func TestTxSetPutGet(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ipfs := tempIPFSStorage(t, ctx)

	tX := &tx.Tx{
		Version: tx.Version1,
		Ts:      time.Now().Unix(),
		Type:    tx.TxType_Node,
		Data: &tx.Node{
			Id:  "1111",
			Did: "did:example:abcdefghijklmnopqrstuvwxyz0123456789",
			Key: nil,
		},
	}

	txCid, err := ipfs.PutTx(ctx, tX)
	if err != nil {
		t.Fatal(err)
	}

	root, err := storage.NewTxSet(ipfs, []cid.Cid{txCid})
	if err != nil {
		t.Fatal(err)
	}

	txrb, err := ipfs.AllTx(ctx, &storage.Block{TxRoot: root.Cid()})
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, txrb, 1)
	assert.Equal(t, tX, txrb[tx.TxID(txCid)])
}

func TestNodeIndex(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ipfs := tempIPFSStorage(t, ctx)

	//Check no nodes exist

	n, err := ipfs.Nodes()
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, n, 0)

	tX := &tx.Tx{
		Version: tx.Version1,
		Ts:      time.Now().Unix(),
		Type:    tx.TxType_Node,
		Action:  tx.TxActionAdd,
		Data: &tx.Node{
			Id:  "1111",
			Did: "did:example:abcdefghijklmnopqrstuvwxyz0123456789",
			Key: nil,
		},
	}

	txcid, err := ipfs.PutTx(ctx, tX)
	if err != nil {
		t.Fatal(err)
	}

	txs, err := storage.NewTxSet(ipfs, []cid.Cid{txcid})
	if err != nil {
		t.Fatal(err)
	}

	b := &storage.Block{
		TxRoot: txs.Cid(),
	}

	bcid, err := ipfs.PutBlock(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	bid := storage.BlockID(bcid)

	if err := ipfs.MarkBlock(ctx, bid, storage.BlockStateValidated); err != nil {
		t.Fatal(err)
	}

	if err := ipfs.MarkBlock(ctx, bid, storage.BlockStateAccepted); err != nil {
		t.Fatal(err)
	}

	//Check new node exists

	n, err = ipfs.Nodes()
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, n, 1)
	assert.Equal(t, tX.Data.(*tx.Node).Id, n[0])

	ntx, err := ipfs.Node(ctx, n[0])
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, tX.Data, ntx)

	//Remove node

	tX = &tx.Tx{
		Version: tx.Version1,
		Ts:      time.Now().Unix(),
		Type:    tx.TxType_Node,
		Action:  tx.TxActionRevoke,
		Data: &tx.Node{
			Id:  "1111",
			Did: "did:example:abcdefghijklmnopqrstuvwxyz0123456789",
			Key: nil,
		},
	}

	txcid, err = ipfs.PutTx(ctx, tX)
	if err != nil {
		t.Fatal(err)
	}

	txs, err = storage.NewTxSet(ipfs, []cid.Cid{txcid})
	if err != nil {
		t.Fatal(err)
	}

	b = &storage.Block{
		TxRoot: txs.Cid(),
	}

	bcid, err = ipfs.PutBlock(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	bid = storage.BlockID(bcid)

	if err := ipfs.MarkBlock(ctx, bid, storage.BlockStateValidated); err != nil {
		t.Fatal(err)
	}

	if err := ipfs.MarkBlock(ctx, bid, storage.BlockStateAccepted); err != nil {
		t.Fatal(err)
	}

	//Check no nodes exist

	n, err = ipfs.Nodes()
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, n, 0)
}

func TestDidIndex(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ipfs := tempIPFSStorage(t, ctx)

	did := "did:example:abc"

	//Assert empty
	doc, err := ipfs.LookupDID(ctx, did)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	assert.Nil(t, doc)

	h, err := ipfs.DIDHistory(ctx, did)
	assert.NoError(t, err)
	assert.Len(t, h, 0)

	diddoc := &w3cdid.Document{
		ID: did,
	}

	//Add DID
	tX := &tx.Tx{
		Version: tx.Version1,
		Ts:      time.Now().Unix(),
		Type:    tx.TxType_DID,
		Action:  tx.TxActionAdd,
		Data:    &tx.DID{Document: diddoc},
	}

	txcid, err := ipfs.PutTx(ctx, tX)
	if err != nil {
		t.Fatal(err)
	}

	txs, err := storage.NewTxSet(ipfs, []cid.Cid{txcid})
	if err != nil {
		t.Fatal(err)
	}

	b := &storage.Block{
		TxRoot: txs.Cid(),
	}

	bcid, err := ipfs.PutBlock(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	bid := storage.BlockID(bcid)

	if err := ipfs.MarkBlock(ctx, bid, storage.BlockStateValidated); err != nil {
		t.Fatal(err)
	}

	if err := ipfs.MarkBlock(ctx, bid, storage.BlockStateAccepted); err != nil {
		t.Fatal(err)
	}

	//Check for DID
	doc, err = ipfs.LookupDID(ctx, did)
	assert.NoError(t, err)
	assert.Equal(t, diddoc, doc)

	h, err = ipfs.DIDHistory(ctx, did)
	assert.NoError(t, err)
	assert.Len(t, h, 1)
	assert.Equal(t, tX, h[0])

	//Revoke

	tX = &tx.Tx{
		Version: tx.Version1,
		Ts:      time.Now().Unix(),
		Type:    tx.TxType_DID,
		Action:  tx.TxActionRevoke,
		Data:    &tx.DID{Document: diddoc},
	}

	txcid, err = ipfs.PutTx(ctx, tX)
	if err != nil {
		t.Fatal(err)
	}

	txs, err = storage.NewTxSet(ipfs, []cid.Cid{txcid})
	if err != nil {
		t.Fatal(err)
	}

	b = &storage.Block{
		TxRoot: txs.Cid(),
	}

	bcid, err = ipfs.PutBlock(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	bid = storage.BlockID(bcid)

	if err := ipfs.MarkBlock(ctx, bid, storage.BlockStateValidated); err != nil {
		t.Fatal(err)
	}

	if err := ipfs.MarkBlock(ctx, bid, storage.BlockStateAccepted); err != nil {
		t.Fatal(err)
	}

	//Check for DID revoke
	doc, err = ipfs.LookupDID(ctx, did)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	assert.Nil(t, doc)

	h, err = ipfs.DIDHistory(ctx, did)
	assert.NoError(t, err)
	assert.Len(t, h, 2)

}

func TestApplyGenesis(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Use the mem store to calculate IDs for IPFS

	mem := storage.NewMemStore()
	ipfs := tempIPFSStorage(t, ctx)

	tx1 := &tx.Tx{
		Version: tx.Version1,
		Ts:      time.Now().Unix(),
		From:    "did:example:1234",
		Type:    tx.TxType_Node,
		Action:  tx.TxActionAdd,
		Data: &tx.Node{
			Id:  "tt",
			Did: "did:example:1234",
		},
	}

	txcid, err := mem.PutTx(ctx, tx1)
	if err != nil {
		t.Fatal(err)
	}

	txs := []cid.Cid{txcid}

	bl, err := storage.NewBlock(ctx, mem, storage.BlockID(cid.Undef), txs)
	if err != nil {
		t.Fatal(err)
	}

	g := &storage.GenesisInfo{
		ChainID: "test",
		Block:   *bl,
		Txs:     []*tx.Tx{tx1},
	}

	err = ipfs.ApplyGenesis(g)
	assert.NoError(t, err)

	n, err := ipfs.Nodes()
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, n, 1)
}
