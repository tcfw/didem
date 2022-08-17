package storage

import (
	"context"
	"encoding/hex"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/kubo/config"
	"github.com/stretchr/testify/assert"
	"github.com/tcfw/didem/pkg/tx"
	"golang.org/x/crypto/sha3"
)

func TestTypedKey(t *testing.T) {
	k := typedKey(txBlockTPrefix, "a", "a")
	e := []byte{byte(txBlockTPrefix), 'a', tableSep, 'a'}
	assert.Equal(t, e, k)
}

func TestIPFSAdd(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repoPath, err := ioutil.TempDir("", "ipfs-shell")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		os.RemoveAll(repoPath)
	})

	id, err := config.CreateIdentity(ioutil.Discard, []options.KeyGenerateOption{options.Key.Type(options.Ed25519Key)})
	if err != nil {
		t.Fatal(err)
	}

	ipfs, err := NewIPFSStorage(ctx, id, repoPath)
	if err != nil {
		t.Fatal(err)
	}

	tX := &tx.Tx{}

	cid, err := ipfs.PutTx(ctx, tX)
	if err != nil {
		t.Fatal(err)
	}

	txb, _ := tX.Marshal()
	txbh := sha3.Sum384(txb)

	expected := hex.EncodeToString(txbh[:])

	idhex := hex.EncodeToString(cid.Bytes())

	assert.Equal(t, expected, idhex)

	txrb, err := ipfs.GetTx(ctx, tx.TxID(cid))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, tX, txrb)
}
