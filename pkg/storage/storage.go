package storage

import (
	"context"

	"github.com/tcfw/didem/pkg/tx"
)

type Storage interface {
	PutTx(context.Context, *tx.Tx) (tx.ID, error)
	GetTx(context.Context, tx.ID) (*tx.Tx, error)

	Tips(context.Context) ([]tx.Tx, error)

	Parents(context.Context, tx.ID) ([]tx.Tx, error)
}
