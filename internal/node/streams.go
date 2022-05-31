package node

import (
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/tcfw/didem/internal/did"
	"github.com/tcfw/didem/internal/em"
)

type handlerSetup func(n *Node) (network.StreamHandler, interface{}, error)

var (
	streamHandlers = map[protocol.ID]handlerSetup{
		did.ProtocolID: newDidStreamHandler,
		em.ProtocolID:  newEmStreamHandler,
	}
)

func (n *Node) setupStreamHandlers() error {
	p2p := n.p2p.host

	for id, handler := range streamHandlers {
		handler, inst, err := handler(n)
		if err != nil {
			return err
		}

		p2p.SetStreamHandler(id, handler)
		n.handlers[id] = inst
	}

	return nil
}

func newDidStreamHandler(n *Node) (network.StreamHandler, interface{}, error) {
	did := did.NewHandler(n)

	return did.Handle, did, nil
}

func newEmStreamHandler(n *Node) (network.StreamHandler, interface{}, error) {
	em := em.NewHandler(n)

	return em.Handle, em, nil
}
