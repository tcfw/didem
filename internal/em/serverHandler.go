package em

import (
	"context"
	"crypto/rand"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/pkg/errors"
	"github.com/tcfw/didem/pkg/did"
	"github.com/tcfw/didem/pkg/em"
	"github.com/vmihailenco/msgpack/v5"
)

type ServerHandler struct {
	resolver did.Resolver

	//stores
	idStore did.IdentityStore
	emStore em.Store

	//stream
	stream network.Stream
	rw     *streamRW

	//handshake state
	clientHello *em.Email
	identity    did.PrivateIdentity
	received    *em.Email
}

func NewServerHandler(stream network.Stream, store em.Store, idStore did.IdentityStore) *ServerHandler {
	s := &ServerHandler{
		stream:  stream,
		emStore: store,
		idStore: idStore,
		rw:      NewStreamRW(stream),
	}

	return s
}

func (s *ServerHandler) handle(ctx context.Context) error {
	if s.resolver == nil {
		return errors.New("no identity resolver set to validate recipients")
	}

	if s.idStore == nil {
		return errors.New("no identity store set to receive mail")
	}

	defer s.stream.Close()

	if err := s.readClientHello(); err != nil {
		return errors.Wrap(err, "reading client hello")
	}

	if err := s.validateClientHello(); err != nil {
		return errors.Wrap(err, "validating client hello")
	}

	if err := s.sendHello(); err != nil {
		return errors.Wrap(err, "sending server hello")
	}

	if err := s.readEmail(); err != nil {
		return errors.Wrap(err, "reading email")
	}

	return s.emStore.AddToInbox(s.received)
}

func (s *ServerHandler) readClientHello() error {
	b, err := s.rw.Read()
	if err != nil {
		return errors.Wrap(err, "reading client hello")
	}

	s.clientHello = &em.Email{}
	if err := msgpack.Unmarshal(b, s.clientHello); err != nil {
		return errors.Wrap(err, "parsing client hello")
	}

	return nil
}

func (s *ServerHandler) validateClientHello() error {
	if len(s.clientHello.Headers) != 0 {
		return errors.New("unexpected headers in client hello")
	}

	if len(s.clientHello.Parts) != 0 {
		return errors.New("unexpected parts in client hello")
	}

	pi, err := s.idStore.Find(s.clientHello.To.ID)
	if err != nil {
		return errors.Wrap(err, "finding matching identity")
	}

	s.identity = pi

	pk, err := s.resolver.Find(s.clientHello.From.ID)
	if err != nil {
		return errors.Wrap(err, "finding sender identity")
	}

	if !pk.Matches(s.clientHello.From) {
		return errors.New("public identity mismatch")
	}

	hd := s.clientHello
	hd.Signature = nil
	b, err := msgpack.Marshal(hd)
	if err != nil {
		return errors.Wrap(err, "rebuilding signature data")
	}
	var hasMatchingSignature bool
	for _, pk := range s.clientHello.To.PublicKeys {
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

func (s *ServerHandler) sendHello() error {
	hello := s.makeHello()
	if err := s.sign(hello); err != nil {
		return errors.Wrap(err, "signing client hello")
	}

	b, err := msgpack.Marshal(hello)
	if err != nil {
		return errors.Wrap(err, "mashalling client hello")
	}

	return s.rw.Write(b)
}

func (c *ServerHandler) makeHello() *em.Email {
	//swap from/to
	e := &em.Email{
		Time:  c.clientHello.Time,
		From:  c.clientHello.To,
		To:    c.clientHello.From,
		Nonce: c.clientHello.Nonce,
	}

	rand.Read(e.Nonce[:])

	return e
}

func (s *ServerHandler) readEmail() error {
	b, err := s.rw.Read()
	if err != nil {
		return err
	}

	s.received = &em.Email{}
	if err := msgpack.Unmarshal(b, s.received); err != nil {
		return errors.Wrap(err, "parsing email")
	}

	return s.validateEmail()
}

func (s *ServerHandler) validateEmail() error {
	if !s.clientHello.From.Matches(s.received.From) {
		return errors.New("received sender does not match client hello")
	}

	if !s.clientHello.To.Matches(s.received.To) {
		return errors.New("received sender does not match client hello")
	}

	hd := s.clientHello
	hd.Signature = nil
	b, err := msgpack.Marshal(hd)
	if err != nil {
		return errors.Wrap(err, "rebuilding signature data")
	}
	var hasMatchingSignature bool
	for _, pk := range s.clientHello.To.PublicKeys {
		if err := verify(pk.Key, b, hd.Signature); err == nil {
			hasMatchingSignature = true
			break
		}
	}

	if !hasMatchingSignature {
		return errors.New("no matching singature found on received email")
	}

	return nil
}

func (c *ServerHandler) sign(e *em.Email) error {
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
