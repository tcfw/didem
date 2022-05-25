package node

import (
	"context"

	"github.com/libp2p/go-libp2p"
	"github.com/tcfw/didem/internal/config"
)

func getIdentity(ctx context.Context, cfg *config.Config) (libp2p.Option, error) {
	//TODO(tcfw): store identities
	return libp2p.RandomIdentity, nil
}
