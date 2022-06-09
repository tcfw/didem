package em

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/pkg/errors"
	"github.com/tcfw/didem/pkg/em"
	"github.com/tcfw/didem/pkg/node"
)

const (
	ProtocolID      = protocol.ID("em")
	helloLimit      = 1 << 10
	helloTimeWindow = 5 * time.Second
)

type Handler struct {
	n              node.Node
	receiveTimeout time.Duration
}

func NewHandler(n node.Node) *Handler {
	return &Handler{n: n}
}

func (h *Handler) Handle(stream network.Stream) {
	srv := &ServerHandler{
		stream:   stream,
		resolver: h.n.Resolver(),
		idStore: h.n.ID(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.receiveTimeout)
	defer cancel()

	srv.handle(ctx)
}

func (h *Handler) Send(ctx context.Context, email *em.Email, opts ...em.SendOption) error {
	cfg := em.NewDefaultSendConfig()

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return errors.Wrap(err, "building config")
		}
	}

	return h.send(ctx, email, cfg)
}

func (h *Handler) send(ctx context.Context, email *em.Email, cfg *em.SendConfig) error {
	if cfg.Id == nil {
		return errors.New("sender identity missing")
	}

	if email.To == nil {
		return errors.New("recipient identity missing")
	}

	if cfg.PublicResolver == nil {
		return errors.New("no resolver configure")
	}

	rDID, err := cfg.PublicResolver.Find(email.To.ID)
	if err != nil {
		return errors.Wrap(err, "resolving recipient")
	}

	client := &ClientHandler{
		n:         h.n,
		email:     email,
		recipient: rDID,
		resolver:  h.n.Resolver(),
	}

	return client.handle(ctx)
}
