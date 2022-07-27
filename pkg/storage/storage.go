package storage

import (
	"context"

	"github.com/ipfs/go-cid"
	"github.com/tcfw/didem/pkg/tx"
)

const (
	CIDEncodingTx    = 0x87
	CIDEncodingBlock = 0x88
	CIDEncodingSet   = 0x89
)

type Store interface {
	PutTx(context.Context, *tx.Tx) (cid.Cid, error)
	GetTx(context.Context, cid.Cid) (*tx.Tx, error)

	PutBlock(context.Context, *Block) (cid.Cid, error)
	GetBlock(context.Context, cid.Cid) (*Block, error)

	PutSet(context.Context, *TxSet) (cid.Cid, error)
	GetSet(context.Context, cid.Cid) (*TxSet, error)
}
