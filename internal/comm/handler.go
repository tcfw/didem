package comm

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/pkg/errors"
	"github.com/tcfw/didem/pkg/comm"
	"github.com/tcfw/didem/pkg/node"
)

const (
	ProtocolID      = protocol.ID("/did/comm/0.0.1")
	helloLimit      = 1 << 10
	helloTimeWindow = 5 * time.Second
)

type Handler struct {
	n              node.Node
	receiveTimeout time.Duration
	store          comm.Store
}

func NewHandler(n node.Node) *Handler {
	return &Handler{n: n}
}

func (h *Handler) Handle(stream network.Stream) {
	srv := NewServerHandler(stream, h.store, h.n.ID())

	ctx, cancel := context.WithTimeout(context.Background(), h.receiveTimeout)
	defer cancel()

	srv.handle(ctx)
}

func (h *Handler) Send(ctx context.Context, email *comm.Message, opts ...comm.SendOption) error {
	cfg := comm.NewDefaultSendConfig()

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return errors.Wrap(err, "building config")
		}
	}

	return h.send(ctx, email, cfg)
}

func (h *Handler) send(ctx context.Context, email *comm.Message, cfg *comm.SendConfig) error {
	if cfg.Id == nil {
		return errors.New("sender identity missing")
	}

	if email.To == nil {
		return errors.New("recipient identity missing")
	}

	if cfg.PublicResolver == nil {
		return errors.New("no resolver configure")
	}

	rDID, err := cfg.PublicResolver.ResolvePI(email.To[0].ID)
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
