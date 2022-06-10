package em

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"github.com/tcfw/didem/pkg/did"
	"github.com/tcfw/didem/pkg/em"
	"github.com/tcfw/didem/pkg/node"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/crypto/sha3"
)

type ClientHandler struct {
	n node.Node

	//send info
	email     *em.Email
	identity  did.PrivateIdentity
	recipient *did.PublicIdentity

	//stream
	stream network.Stream
	rw     *streamRW

	//state
	serverHello *em.Email
	resolver    did.Resolver
}

func (c *ClientHandler) handle(ctx context.Context) error {
	if c.resolver == nil {
		return errors.New("no identity store set to validate recipients")
	}

	if c.identity == nil {
		return errors.New("no sender identity set")
	}

	c.email.Time = time.Now().Unix()

	s, err := c.connectEndProvider(ctx)
	if err != nil {
		return err
	}
	c.stream = s
	c.rw = NewStreamRW(s)
	defer c.stream.Close()

	c.email.Time = time.Now().Unix()
	rand.Read(c.email.Nonce[:])

	if err := c.handshake(); err != nil {
		return errors.Wrap(err, "client handshake")
	}

	return c.send()
}

func (c *ClientHandler) handshake() error {
	if err := c.sendHello(); err != nil {
		return err
	}

	if err := c.readServerHello(); err != nil {
		return err
	}

	if err := c.validateServerHello(); err != nil {
		return err
	}

	return nil
}

func (c *ClientHandler) sendHello() error {
	hello := c.makeHello()
	if err := c.sign(hello); err != nil {
		return errors.Wrap(err, "signing client hello")
	}

	b, err := msgpack.Marshal(hello)
	if err != nil {
		return errors.Wrap(err, "mashalling client hello")
	}

	return c.rw.Write(b)
}

func (c *ClientHandler) readServerHello() error {
	b, err := c.rw.Read()
	if err != nil {
		return errors.Wrap(err, "reading server hello")
	}

	serverHello := &em.Email{}
	if err := msgpack.Unmarshal(b, serverHello); err != nil {
		return errors.Wrap(err, "unmarshalling server hello")
	}
	c.serverHello = serverHello

	return nil
}

func (c *ClientHandler) validateServerHello() error {
	h := c.serverHello

	//Check time window
	helloTime := time.Unix(h.Time, 0)
	if helloTime.Before(time.Now().Add(-helloTimeWindow)) || helloTime.After(time.Now().Add(helloTimeWindow)) {
		return errors.New("hello too old")
	}

	//Check nonce
	nsum := 0
	for _, np := range h.Nonce {
		nsum += int(np)
	}
	if nsum == 0 {
		return errors.New("empty nonce")
	}

	//Check recipient identity
	if h.To.ID == "" || len(h.To.PublicKeys) == 0 {
		return errors.New("no identity provided")
	}
	if h.From.ID != c.recipient.ID {
		return errors.New("unexpected recipient ID")
	}

	if !c.recipient.Matches(h.From) {
		return errors.New("public identity mismatch")
	}

	//Check sender identity
	if h.To.ID != c.email.From.ID {
		return errors.New("unexpected change in sender ID")
	}
	//TODO(tcfw) deep compare sender pubkeys

	//Check public identity
	knownId, err := c.resolver.Find(h.From.ID)
	if err != nil {
		return errors.New("validating publickly known identity of recipient")
	}

	//Check signature
	hd := h
	hd.Signature = nil
	b, err := msgpack.Marshal(hd)
	if err != nil {
		return errors.Wrap(err, "rebuilding signature data")
	}
	var hasMatchingSignature bool
	for _, pk := range knownId.PublicKeys {
		if err := verify(pk.Key, b, hd.Signature); err == nil {
			hasMatchingSignature = true
			break
		}
	}

	if !hasMatchingSignature {
		return errors.New("no matching singature found")
	}

	return nil
}

func (c *ClientHandler) sign(e *em.Email) error {
	b, err := msgpack.Marshal(e)
	if err != nil {
		return errors.Wrap(err, "mashalling client hello")
	}

	s, err := sign(c.identity.PrivateKey(), b)
	if err != nil {
		return err
	}
	e.Signature = s

	return nil
}

func (c *ClientHandler) makeHello() *em.Email {
	e := &em.Email{
		Time:  c.email.Time,
		From:  c.email.From,
		To:    c.email.To,
		Nonce: c.email.Nonce,
	}

	rand.Read(e.Nonce[:])

	return e
}

func (c *ClientHandler) send() error {
	b, err := msgpack.Marshal(c.email)
	if err != nil {
		return errors.Wrap(err, "marshaling email")
	}

	if _, err = c.stream.Write(b); err != nil {
		return errors.Wrap(err, "transmitting email")
	}

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
