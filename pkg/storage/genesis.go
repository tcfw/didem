package storage

import "github.com/tcfw/didem/pkg/tx"

// GenesisInfo provides enough information to be used as the genesis block
// of the block chain. It does not provide a TxSet assuming that the block
// definition stores the root of the CID Merkle Tree and the storage engine
// needs to check if the TXs aligns with the TxSet root when applying the
// TXs with a metadata store.
// A ChainID is also provided
type GenesisInfo struct {
	ChainID string   `msgpack:"chain"`
	Block   Block    `magpack:"block"`
	Txs     []*tx.Tx `msgpack:"txs"`
}
