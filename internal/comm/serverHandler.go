package comm

import (
	"context"
	"crypto/rand"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/pkg/errors"
	"github.com/tcfw/didem/internal/stream"
	"github.com/tcfw/didem/pkg/comm"
	"github.com/tcfw/didem/pkg/did"
	"github.com/vmihailenco/msgpack/v5"
)

type ServerHandler struct {
	resolver did.Resolver

	//stores
	idStore did.IdentityStore
	emStore comm.Store

	//stream
	stream network.Stream
	rw     *stream.RW

	//handshake state
	clientHello *comm.Message
	identity    did.PrivateIdentity
	received    *comm.Message
}

func NewServerHandler(nstream network.Stream, store comm.Store, idStore did.IdentityStore) *ServerHandler {
	s := &ServerHandler{
		stream:  nstream,
		emStore: store,
		idStore: idStore,
		rw:      stream.NewRW(nstream),
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

	s.clientHello = &comm.Message{}
	if err := msgpack.Unmarshal(b, s.clientHello); err != nil {
		return errors.Wrap(err, "parsing client hello")
	}

	return nil
}

func (s *ServerHandler) validateClientHello() error {
	if len(s.clientHello.Body) != 0 {
		return errors.New("unexpected body in client hello")
	}

	if len(s.clientHello.Attachments) != 0 {
		return errors.New("unexpected attachments in client hello")
	}

	pi, err := s.idStore.Find(s.clientHello.To[0].ID)
	if err != nil {
		return errors.Wrap(err, "finding matching identity")
	}

	s.identity = pi

	pk, err := s.resolver.ResolvePI(s.clientHello.From.ID)
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
	for _, pk := range s.clientHello.To[0].PublicKeys {
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
	hello, err := s.makeHello()
	if err != nil {
		return err
	}

	if err := s.sign(hello); err != nil {
		return errors.Wrap(err, "signing client hello")
	}

	b, err := msgpack.Marshal(hello)
	if err != nil {
		return errors.Wrap(err, "mashalling client hello")
	}

	return s.rw.Write(b)
}

func (c *ServerHandler) makeHello() (*comm.Message, error) {
	//swap from/to
	e := &comm.Message{
		CreatedTime: c.clientHello.CreatedTime,
		From:        c.clientHello.To[0],
		To:          []*did.PublicIdentity{c.clientHello.From},
		Nonce:       c.clientHello.Nonce,
	}

	if _, err := rand.Read(e.Nonce[:]); err != nil {
		return nil, errors.Wrap(err, "reading nonce")
	}

	return e, nil
}

func (s *ServerHandler) readEmail() error {
	b, err := s.rw.Read()
	if err != nil {
		return err
	}

	s.received = &comm.Message{}
	if err := msgpack.Unmarshal(b, s.received); err != nil {
		return errors.Wrap(err, "parsing email")
	}

	return s.validateEmail()
}

func (s *ServerHandler) validateEmail() error {
	if !s.clientHello.From.Matches(s.received.From) {
		return errors.New("received sender does not match client hello")
	}

	if !s.clientHello.To[0].Matches(s.received.To[0]) {
		return errors.New("received sender does not match client hello")
	}

	hd := s.clientHello
	hd.Signature = nil
	b, err := msgpack.Marshal(hd)
	if err != nil {
		return errors.Wrap(err, "rebuilding signature data")
	}
	var hasMatchingSignature bool
	for _, pk := range s.clientHello.To[0].PublicKeys {
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

func (c *ServerHandler) sign(e *comm.Message) error {
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
