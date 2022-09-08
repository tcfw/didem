package resolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/miekg/dns"
	"github.com/tcfw/didem/pkg/did/w3cdid"
)

// based on draft-mayrhofer-did-dns-03

// resolveDNS resolves a DID by comparing against well known DNS entries 
func (r *Resolver) resolveDNS(ctx context.Context, did w3cdid.URL, depth int) (*w3cdid.Document, error) {
	if depth > 5 {
		return nil, errors.New("notFound")
	}

	client, err := NewDNSDIDResolver()
	if err != nil {
		return nil, errors.Wrap(err, "constructing dns client")
	}

	rrname := fmt.Sprintf("_did.%s", did.Id())

	if did.Fragment() != "" {
		rrname = fmt.Sprintf("_%s.%s", did.Fragment(), rrname)
	}

	rrres, err := client.Resolve(ctx, rrname)
	if err != nil {
		return nil, err
	}

	var doc *w3cdid.Document

	if strings.HasPrefix(rrres, "did:dns:") {
		doc, err = r.resolveDNS(ctx, w3cdid.URL(rrres), depth+1)
	} else {
		doc, err = r.ResolveContext(ctx, w3cdid.URL(rrres))
	}
	if err != nil {
		return nil, err
	}

	doc.ID = did.Id()

	return doc, nil
}

type DNSDIDResolver struct {
	dnsc   *dns.Client
	config *dns.ClientConfig
}

func (d *DNSDIDResolver) Resolve(ctx context.Context, name string) (string, error) {
	if !strings.HasSuffix(name, ".") {
		name += "."
	}

	req := &dns.Msg{}
	req.Id = dns.Id()
	req.RecursionDesired = true
	req.Question = make([]dns.Question, 1)
	req.Question[0] = dns.Question{name, dns.TypeURI, dns.ClassINET}

	resp, _, err := d.dnsc.ExchangeContext(ctx, req, fmt.Sprintf("%s:%s", d.config.Servers[0], d.config.Port))
	if err != nil {
		return "", errors.Wrap(err, "querying dns record")
	}

	if len(resp.Answer) == 0 {
		return "", errors.New("notFound")
	}

	var resolvedReference string

	for _, r := range resp.Answer {
		uri, ok := r.(*dns.URI)
		if !ok {
			continue
		}

		if !strings.HasPrefix(uri.Target, "did:") {
			continue
		}

		resolvedReference = uri.Target
		break
	}

	if resolvedReference == "" {
		return "", errors.New("notFound")
	}

	return resolvedReference, nil
}

func NewDNSDIDResolver() (*DNSDIDResolver, error) {
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return nil, errors.Wrap(err, "building dns config")
	}

	dnsc := new(dns.Client)

	return &DNSDIDResolver{dnsc, config}, nil
}
