package did

import (
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/tcfw/didem/pkg/node"
)

const (
	ProtocolID = protocol.ID("/did/0.0.1")
)

type Handler struct {
	n node.Node
}

func NewHandler(n node.Node) *Handler {
	return &Handler{n}
}

func (h *Handler) Handle(stream network.Stream) {
}
