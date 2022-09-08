package storage

import (
	"context"

	"github.com/tcfw/didem/pkg/did/w3cdid"
	"github.com/tcfw/didem/pkg/tx"
)

// MetadataProvider provides indexes and metadata for TXs that
// have been successfully validates and applied in the block chain
// The information should be maintainable and rebuildable without
// a metadata store and should be used as a performance measure for
// reading blockchain records
type MetadataProvider interface {
	LookupDID(context.Context, string) (*w3cdid.Document, error)
	DIDHistory(context.Context, string) ([]*tx.Tx, error)

	Claims(context.Context, string) ([]*tx.Tx, error) //TODO(tcfw): vc type

	Nodes() ([]string, error)
	Node(context.Context, string) (*tx.Node, error)

	HasGenesisApplied() bool
	ApplyGenesis(*GenesisInfo) error

	ApplyTx(context.Context, tx.TxID, *tx.Tx) error

	StartTest(context.Context) (Store, error)
	CompleteTest(context.Context) error
}
