package consensus

import "github.com/ipfs/go-cid"

type BlockStore interface {
	getBlock(cid.Cid) (*Block, error)
	addBlock([]byte) (cid.Cid, error)
}
