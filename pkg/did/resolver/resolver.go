package resolver

import (
	"context"
	"errors"

	"github.com/tcfw/didem/pkg/did/w3cdid"
)

var (
	ErrUnknownMethod = errors.New("unknown did method")
)

type Resolver struct{}

// Resolve resolves a public DID ID via any supported DID method
func (r *Resolver) Resolve(did w3cdid.URL) (*w3cdid.Document, error) {
	return r.ResolveContext(context.Background(), did)
}

func (r *Resolver) ResolveContext(ctx context.Context, did w3cdid.URL) (*w3cdid.Document, error) {
	switch did.Method() {
	case "dns":
		return r.resolveDNS(ctx, did, 0)
	default:
		return nil, ErrUnknownMethod
	}
}
