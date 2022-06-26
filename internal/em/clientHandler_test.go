package em

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/tcfw/didem/pkg/did"
	"github.com/tcfw/didem/pkg/em"
	emPkg "github.com/tcfw/didem/pkg/em"
)

func TestSendClientHello(t *testing.T) {
	h1, err := libp2p.New()
	if err != nil {
		t.Fatal(err)
	}

	h2, err := libp2p.New()
	if err != nil {
		t.Fatal(err)
	}

	var received *em.Email
	done := make(chan struct{}, 1)

	h1.SetStreamHandler(ProtocolID, func(s network.Stream) {
		sh := &ServerHandler{rw: NewStreamRW(s)}

		if err := sh.readClientHello(); err != nil {
			t.Fatal(err)
		}

		received = sh.clientHello
		s.Close()
		done <- struct{}{}
	})

	h2.Connect(context.Background(), peer.AddrInfo{
		ID:    h1.ID(),
		Addrs: h1.Addrs(),
	})

	s, err := h2.NewStream(context.Background(), h1.ID(), ProtocolID)
	if err != nil {
		t.Fatal(err)
	}

	toId, err := did.GenerateEd25519Identity(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	toIdPub, err := toId.PublicIdentity()
	if err != nil {
		t.Fatal(err)
	}

	fromId, err := did.GenerateEd25519Identity(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	fromIdPub, err := fromId.PublicIdentity()
	if err != nil {
		t.Fatal(err)
	}

	email := &emPkg.Email{
		Time:    1,
		To:      toIdPub,
		From:    fromIdPub,
		Headers: map[string]string{"subject": "test"},
		Parts: []em.EmailPart{
			{Mime: "text/plain", Data: []byte("test")},
		},
	}

	handle := &ClientHandler{
		email:     email,
		identity:  fromId,
		recipient: toIdPub,
	}

	stream := &ClientStream{
		ch:     handle,
		stream: s,
		rw:     NewStreamRW(s),
	}

	if err := stream.sendHello(); err != nil {
		t.Fatal(err)
	}

	<-done
	assert.NotEmpty(t, received)
}
