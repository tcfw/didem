package storage

import (
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/ipfs/go-cid"
)

const (
	falsePositive = 0.01
)

func MakeBloom(tx []cid.Cid) ([]byte, error) {
	b := bloom.NewWithEstimates(MaxBlockTxCount, falsePositive)

	for _, t := range tx {
		b.Add(t.Bytes())
	}

	return b.GobEncode()
}

func BloomContains(b []byte, tx cid.Cid) (bool, error) {
	bloom := bloom.NewWithEstimates(MaxBlockTxCount, falsePositive)

	if err := bloom.GobDecode(b); err != nil {
		return false, err
	}

	return bloom.Test(tx.Bytes()), nil
}
