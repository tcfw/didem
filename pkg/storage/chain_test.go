package storage

import (
	"sort"
	"strconv"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/assert"
)

func TestCidListSort(t *testing.T) {
	mh, _ := multihash.Sum([]byte("aaa"), multihash.SHA3_256, multihash.DefaultLengths[multihash.SHA3_256])
	c := cid.NewCidV1(CIDEncodingBlock, mh)
	txs := []cid.Cid{c, cid.Undef}
	sort.Sort(cidList(txs))

	assert.Equal(t, cid.Undef, txs[0])
	assert.Equal(t, c, txs[1])
}

func TestNewTxSet(t *testing.T) {
	set := []cid.Cid{}

	for i := 0; i < MaxBlockTxCount; i++ {
		mh, _ := multihash.Sum([]byte(strconv.Itoa(i)), multihash.SHA3_256, multihash.DefaultLengths[multihash.SHA3_256])
		c := cid.NewCidV1(CIDEncodingBlock, mh)
		set = append(set, c)
	}

	s := NewMemStore()

	root, err := NewTxSet(s, set)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotEmpty(t, root)
}
