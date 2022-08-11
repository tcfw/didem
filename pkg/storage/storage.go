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
	MetadataProvider

	PutTx(context.Context, *tx.Tx) (cid.Cid, error)
	GetTx(context.Context, tx.TxID) (*tx.Tx, error)

	PutBlock(context.Context, *Block) (cid.Cid, error)
	GetBlock(context.Context, BlockID) (*Block, error)

	AllTx(context.Context, *Block) (map[tx.TxID]*tx.Tx, error)

	PutSet(context.Context, *TxSet) (cid.Cid, error)
	GetSet(context.Context, cid.Cid) (*TxSet, error)

	GetTxBlock(context.Context, tx.TxID) (*Block, error)
	MarkBlock(context.Context, BlockID, BlockState) error
}
