package em

import (
	"context"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/tcfw/didem/pkg/did"
)

type ServerHandler struct {
	resolver did.Resolver
	idStore  did.IdentityStore
	stream   network.Stream
}

func (s *ServerHandler) handle(ctx context.Context) error {
	return nil
}
