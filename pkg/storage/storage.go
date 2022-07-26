package storage

import (
	"context"

	"github.com/ipfs/go-cid"
	"github.com/tcfw/didem/pkg/tx"
)

type Storage interface {
	PutTx(context.Context, *tx.Tx) (tx.ID, error)
	GetTx(context.Context, tx.ID) (*tx.Tx, error)

	PutBlock(context.Context, *Block) (cid.Cid, error)
	GetBlock(context.Context, cid.Cid) (*Block, error)
}
