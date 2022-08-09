package storage

import (
	"github.com/tcfw/didem/pkg/did/w3cdid"
	"github.com/tcfw/didem/pkg/tx"
)

type MetadataProvider interface {
	LookupDID(string) (*w3cdid.Document, error)
	DIDHistory(string) ([]*tx.Tx, error)

	Claims(string) ([]*tx.Tx, error) //TODO(tcfw): vc type

	Nodes() ([]string, error)
	Node(string) (*tx.Node, error)
}
