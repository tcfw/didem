package storage

import "github.com/tcfw/didem/pkg/tx"

type GenesisInfo struct {
	ChainID string
	Block   Block
	Txs     []tx.Tx
}
