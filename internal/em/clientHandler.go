package em

import (
	"context"
	"crypto/rand"
	"sync"
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
	resolver  did.Resolver
}

type ClientStream struct {
	ch *ClientHandler

	//stream
	stream network.Stream
	rw     *streamRW

	//state
	serverHello *em.Email
}

func (c *ClientHandler) handle(ctx context.Context) error {
	if c.resolver == nil {
		return errors.New("no identity store set to validate recipients")
	}

	if c.identity == nil {
		return errors.New("no sender identity set")
	}

	c.email.Time = time.Now().Unix()
	rand.Read(c.email.Nonce[:])

	ss, err := c.connectEndProviders(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errs := make(chan error, 1)
	done := make(chan struct{})

	for _, s := range ss {
		wg.Add(1)

		go func(stream network.Stream) {
			defer wg.Done()
			defer stream.Close()

			cs := ClientStream{ch: c, stream: stream, rw: NewStreamRW(stream)}

			if err := cs.handshake(); err != nil {
				errs <- errors.Wrap(err, "client handshake")
				return
			}

			if err := cs.send(); err != nil {
				errs <- err
			}
		}(s)
	}

	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

	select {
	case err := <-errs:
		return err
	case <-done:
		return nil
	}
}

func (cs *ClientStream) handshake() error {
	if err := cs.sendHello(); err != nil {
		return err
	}

	if err := cs.readServerHello(); err != nil {
		return err
	}

	if err := cs.validateServerHello(); err != nil {
		return err
	}

	return nil
}

func (cs *ClientStream) sendHello() error {
	hello := cs.makeHello()
	if err := cs.sign(hello); err != nil {
		return errors.Wrap(err, "signing client hello")
	}

	b, err := msgpack.Marshal(hello)
	if err != nil {
		return errors.Wrap(err, "mashalling client hello")
	}

	return cs.rw.Write(b)
}

func (cs *ClientStream) readServerHello() error {
	b, err := cs.rw.Read()
	if err != nil {
		return errors.Wrap(err, "reading server hello")
	}

	serverHello := &em.Email{}
	if err := msgpack.Unmarshal(b, serverHello); err != nil {
		return errors.Wrap(err, "unmarshalling server hello")
	}
	cs.serverHello = serverHello

	return nil
}

func (cs *ClientStream) validateServerHello() error {
	h := cs.serverHello

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
	if h.From.ID != cs.ch.recipient.ID {
		return errors.New("unexpected recipient ID")
	}

	if !cs.ch.recipient.Matches(h.From) {
		return errors.New("public identity mismatch")
	}

	//Check sender identity
	if h.To.ID != cs.ch.email.From.ID {
		return errors.New("unexpected change in sender ID")
	}
	//TODO(tcfw) deep compare sender pubkeys

	//Check public identity
	knownId, err := cs.ch.resolver.ResolvePI(h.From.ID)
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

func (cs *ClientStream) sign(e *em.Email) error {
	b, err := msgpack.Marshal(e)
	if err != nil {
		return errors.Wrap(err, "mashalling client hello")
	}

	s, err := sign(cs.ch.identity.PrivateKey(), b)
	if err != nil {
		return err
	}
	e.Signature = s

	return nil
}

func (cs *ClientStream) makeHello() *em.Email {
	e := &em.Email{
		Time:  cs.ch.email.Time,
		From:  cs.ch.email.From,
		To:    cs.ch.email.To,
		Nonce: cs.ch.email.Nonce,
	}

	rand.Read(e.Nonce[:])

	return e
}

func (cs *ClientStream) send() error {
	b, err := msgpack.Marshal(cs.ch.email)
	if err != nil {
		return errors.Wrap(err, "marshaling email")
	}

	if _, err = cs.stream.Write(b); err != nil {
		return errors.Wrap(err, "transmitting email")
	}

	return nil
}

func (c *ClientHandler) findProviders(ctx context.Context) ([]peer.AddrInfo, error) {
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

	return addrs, nil
}

func (c *ClientHandler) connectEndProviders(ctx context.Context) ([]network.Stream, error) {
	addrs, err := c.findProviders(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "finding recipient provider")
	}

	if len(addrs) == 0 {
		return nil, errors.New("no recipient providers")
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	connectedAddrs := []peer.AddrInfo{}

	//try each recipient
	for _, addr := range addrs {
		wg.Add(1)
		go func(addr peer.AddrInfo) {
			defer wg.Done()
			if err := c.n.P2P().Connect(ctx, addr); err == nil {
				mu.Lock()
				connectedAddrs = append(connectedAddrs, addr)
				mu.Unlock()
			}
		}(addr)
	}

	wg.Wait()

	if len(connectedAddrs) == 0 {
		return nil, errors.New("no recipient provider connected")
	}

	streams := make([]network.Stream, 0, len(connectedAddrs))

	for _, a := range connectedAddrs {
		wg.Add(1)
		go func(connAddr peer.AddrInfo) {
			defer wg.Done()
			stream, err := c.n.P2P().Open(ctx, connAddr.ID, ProtocolID)
			if err != nil {
				//TODO(tcfw) monitor for failed opens, this might be a bad actor
				return
			}
			mu.Lock()
			streams = append(streams, stream)
			mu.Unlock()
		}(a)
	}

	wg.Wait()

	return streams, nil
}
