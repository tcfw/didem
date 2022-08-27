package storage

import (
	"context"

	"github.com/tcfw/didem/pkg/did/genesis"
	"github.com/tcfw/didem/pkg/did/w3cdid"
	"github.com/tcfw/didem/pkg/tx"
)

type MetadataProvider interface {
	LookupDID(context.Context, string) (*w3cdid.Document, error)
	DIDHistory(context.Context, string) ([]*tx.Tx, error)

	Claims(context.Context, string) ([]*tx.Tx, error) //TODO(tcfw): vc type

	Nodes() ([]string, error)
	Node(context.Context, string) (*tx.Node, error)

	HasGenesisApplied() bool
	ApplyGenesis(*genesis.Info) error
}
