package resolver

import (
	"context"

	"github.com/tcfw/didem/pkg/did/w3cdid"
)

type Resolver struct{}

func (r *Resolver) Resolve(did w3cdid.URL) (*w3cdid.Document, error) {
	return r.ResolveContext(context.Background(), did)
}

func (r *Resolver) ResolveContext(ctx context.Context, did w3cdid.URL) (*w3cdid.Document, error) {

}
