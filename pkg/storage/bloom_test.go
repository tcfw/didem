package storage

import (
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/assert"
)

func TestBloom(t *testing.T) {
	mh1, _ := multihash.Sum([]byte{1}, multihash.SHA3_384, multihash.DefaultLengths[multihash.SHA3_384])
	mh2, _ := multihash.Sum([]byte{2}, multihash.SHA3_384, multihash.DefaultLengths[multihash.SHA3_384])

	txCids := []cid.Cid{
		cid.NewCidV1(CIDEncoding, mh1),
		cid.NewCidV1(CIDEncoding, mh2),
	}

	b, err := MakeBloom(txCids)
	if err != nil {
		t.Fatal(err)
	}

	yes, err := BloomContains(b, txCids[0])
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, yes)

	no, _ := BloomContains(b, cid.Undef)
	if err != nil {
		t.Fatal(err)
	}

	assert.False(t, no)
}
