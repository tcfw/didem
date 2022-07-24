package consensus

import "github.com/ipfs/go-cid"

type BlockID string

type Block struct {
	ID        BlockID
	Parent    BlockID
	Height    uint64
	CreatedAt uint64
	Proposer  string
	Signers   uint32
	Signature []byte
	Nonce     []byte
	TxRoot    cid.Cid
}

type TxTrie struct {
	Children []cid.Cid
	Tx       cid.Cid
}

type Tx struct{}
