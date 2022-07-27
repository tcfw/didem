package did

import (
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/tcfw/didem/internal/utils/logging"
	"github.com/tcfw/didem/pkg/did/consensus"
	"github.com/tcfw/didem/pkg/node"
)

const (
	ProtocolID = protocol.ID("/tcfw/did/0.0.1")
)

type Handler struct {
	n node.Node

	consensus *consensus.Consensus
}

func NewHandler(n node.Node) *Handler {
	h := n.P2P().Host()
	p := n.P2P().PubSub()
	c, err := consensus.NewConsensus(h, p)
	if err != nil {
		logging.Entry().Panic(err)
	}

	return &Handler{n: n, consensus: c}
}

func (h *Handler) Handle(stream network.Stream) {
}
