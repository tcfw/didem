package storage

import (
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/ipfs/go-cid"
)

const (
	falsePositive = 0.01
)

// MakeBloom constructs a new bloom byte array from a given set
// of CIDs with a bloom estimation of the total number of CIDs
// in the set
func MakeBloom(tx []cid.Cid) ([]byte, error) {
	b := bloom.NewWithEstimates(MaxBlockTxCount, falsePositive)

	for _, t := range tx {
		b.Add(t.Bytes())
	}

	return b.GobEncode()
}

// BloomContains checks if the given bloom filter contains the CID
func BloomContains(b []byte, tx cid.Cid) (bool, error) {
	bloom := bloom.NewWithEstimates(MaxBlockTxCount, falsePositive)

	if err := bloom.GobDecode(b); err != nil {
		return false, err
	}

	return bloom.Test(tx.Bytes()), nil
}
