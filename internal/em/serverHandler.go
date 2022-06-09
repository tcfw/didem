package em

import (
	"context"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/pkg/errors"
	"github.com/tcfw/didem/pkg/did"
)

type ServerHandler struct {
	resolver did.Resolver
	idStore  did.IdentityStore
	stream   network.Stream
}

func (s *ServerHandler) handle(ctx context.Context) error {
	if s.resolver == nil {
		return errors.New("no identity resolver set to validate recipients")
	}

	if s.idStore == nil {
		return errors.New("no identity store set to receive mail")
	}

	return nil
}
