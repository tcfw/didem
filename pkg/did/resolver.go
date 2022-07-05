package did

import "github.com/tcfw/didem/pkg/did/w3cdid"

type Resolver interface {
	ResolvePI(did string) (*PublicIdentity, error)
	Resolve(did w3cdid.URL) (*w3cdid.Document, error)
}
