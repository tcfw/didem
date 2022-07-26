package storage

import (
	"context"

	"github.com/ipfs/go-cid"
	"github.com/tcfw/didem/pkg/tx"
)

type Storage interface {
	PutTx(context.Context, *tx.Tx) (cid.Cid, error)
	GetTx(context.Context, cid.Cid) (*tx.Tx, error)

	PutBlock(context.Context, *Block) (cid.Cid, error)
	GetBlock(context.Context, cid.Cid) (*Block, error)

	PutTrie(context.Context, *TxTrie) (cid.Cid, error)
	GetTrie(context.Context, cid.Cid) (*TxTrie, error)
}
