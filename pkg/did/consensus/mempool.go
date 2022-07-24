package consensus

import "github.com/tcfw/didem/pkg/tx"

type MemPool interface {
	AddTx(*tx.Tx) error
}

type TxMemPool struct{}
