package resolver

import (
	"context"
	"errors"

	"github.com/miekg/dns"
	"github.com/tcfw/didem/pkg/did/w3cdid"
)

func (r *Resolver) resolveDNS(ctx context.Context, did w3cdid.URL) (*w3cdid.Document, error) {
	return nil, errors.New("not implemented")
}

type DNSDIDResolver struct {
	dnsc *dns.Client
}

func NewDNSDIDResolver() *DNSDIDResolver {
	dnsc := new(dns.Client)

	return &DNSDIDResolver{dnsc}
}
