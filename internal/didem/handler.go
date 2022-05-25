package didem

import (
	"github.com/libp2p/go-libp2p-core/network"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(stream network.Stream) {
}
