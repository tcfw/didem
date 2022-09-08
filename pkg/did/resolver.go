package did

import "github.com/tcfw/didem/pkg/did/w3cdid"

// Resolver allows for a DID to be resolved agnostically any given source
type Resolver interface {
	ResolvePI(did string) (*PublicIdentity, error)
	Resolve(did w3cdid.URL) (*w3cdid.Document, error)
}
