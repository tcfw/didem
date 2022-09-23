package node

import (
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/tcfw/didem/internal/comm"
	"github.com/tcfw/didem/internal/did"
	"github.com/tcfw/didem/internal/utils/logging"
)

type handlerSetup func(n *Node) (network.StreamHandler, interface{}, error)

var (
	streamHandlers = map[protocol.ID]handlerSetup{
		did.ProtocolID:  newDidStreamHandler,
		comm.ProtocolID: newCommStreamHandler,
	}
)

func (n *Node) setupStreamHandlers() error {
	p2p := n.p2p.host

	for id, handler := range streamHandlers {
		handler, inst, err := handler(n)
		if err != nil {
			return err
		}
		if inst == nil {
			continue
		}

		p2p.SetStreamHandler(id, handler)
		n.handlers[id] = inst
	}

	return nil
}

func newDidStreamHandler(n *Node) (network.StreamHandler, interface{}, error) {
	did := did.NewHandler(n)

	if did == nil {
		logging.Entry().Warn("skipping did consensus handler")
		return nil, nil, nil
	}

	go func() {
		for {
			if err := did.Start(); err != nil {
				logging.WithError(err).Error("running did handler")
				time.Sleep(10 * time.Second)
				continue
			}

			return
		}
	}()

	return did.Handle, did, nil
}

func newCommStreamHandler(n *Node) (network.StreamHandler, interface{}, error) {
	comm := comm.NewHandler(n)

	return comm.Handle, comm, nil
}
