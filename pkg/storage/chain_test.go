package storage

import (
	"sort"
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
