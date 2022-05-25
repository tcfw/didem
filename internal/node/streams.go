package node

import (
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/tcfw/didem/internal/didem"
)

type handlerSetup func(n *Node) (network.StreamHandler, error)

var (
	streamHandlers = map[string]handlerSetup{
		"didem": newDidemStreamHandler,
	}
)

func (n *Node) setupStreamHandlers() error {
	p2p := n.p2p.host

	for id, handler := range streamHandlers {
		handler, err := handler(n)
		if err != nil {
			return err
		}
		p2p.SetStreamHandler(protocol.ID(id), handler)
	}

	return nil
}

func newDidemStreamHandler(n *Node) (network.StreamHandler, error) {
	didem := didem.NewHandler()

	return didem.Handle, nil
}
