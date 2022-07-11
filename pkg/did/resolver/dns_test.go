package resolver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tcfw/didem/pkg/did/w3cdid"
)

func TestEmptyLookup(t *testing.T) {
	ctx := context.Background()

	r := &Resolver{}

	resp, err := r.resolveDNS(ctx, w3cdid.URL("did:dns:example.com"), 0)
	assert.Equal(t, "notFound", err.Error())
	assert.Nil(t, resp)
}

func TestKeyFragment(t *testing.T) {
	ctx := context.Background()

	r := &Resolver{}

	//resp usually resolves a did:key doc which isn't supported
	resp, err := r.resolveDNS(ctx, w3cdid.URL("did:dns:danubetech.com#key1"), 0)
	assert.Equal(t, ErrUnknownMethod, err)
	assert.Nil(t, resp)
}
