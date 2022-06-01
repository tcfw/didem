package em

import (
	"context"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"github.com/tcfw/didem/pkg/did"
	"github.com/tcfw/didem/pkg/em"
	"github.com/tcfw/didem/pkg/node"
	"golang.org/x/crypto/sha3"
)

type ClientHandler struct {
	n         node.Node
	email     *em.Email
	recipient *did.PublicIdentity
	stream    network.Stream
}

func (c *ClientHandler) handle(ctx context.Context) error {
	stream, err := c.connectEndProvider(ctx)
	if err != nil {
		return err
	}
	c.stream = stream

	return nil
}

func (c *ClientHandler) connectEndProvider(ctx context.Context) (network.Stream, error) {
	s := sha3.Sum384([]byte(c.recipient.ID))
	mh, err := multihash.Encode(s[:], multihash.SHA3_384)
	if err != nil {
		return nil, err
	}

	recCID := cid.NewCidV1(cid.Raw, mh)
	addrs, err := c.n.P2P().FindProvider(ctx, recCID)
	if err != nil {
		return nil, errors.Wrap(err, "finding recipient provider")
	}

	var connected bool
	var connectedAddr peer.AddrInfo

	//try each recipient
	for _, addr := range addrs {
		if err := c.n.P2P().Connect(ctx, addr); err == nil {
			connected = true
			connectedAddr = addr
			break
		}
	}

	if !connected {
		return nil, errors.New("recipient provider located")
	}

	stream, err := c.n.P2P().Open(ctx, connectedAddr.ID, ProtocolID)
	if err != nil {
		return nil, errors.Wrap(err, "opening stream to recipient")
	}

	return stream, nil
}
