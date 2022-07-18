package comm

import "github.com/tcfw/didem/pkg/did"

type Message struct {
	Type        string                `msgpack:"type"`
	Id          string                `msgpack:"i"`
	ThreadId    string                `msgpack:"th"`
	CreatedTime int64                 `msgpack:"e"`
	ExpiresTime int64                 `msgpack:"ee"`
	From        *did.PublicIdentity   `msgpack:"f"`
	To          []*did.PublicIdentity `msgpack:"t"`
	Nonce       [32]byte              `msgpack:"n"`
	Signature   []byte                `msgpack:"s"`
	Body        map[string]string     `msgpack:"h"`
	Attachments []MessageAttachment   `msgpack:"p"`
}

type MessageAttachment struct {
	Id        string `msgpack:"i"`
	MediaType string `msgpack:"m"`
	Data      []byte `msgpack:"d"`
}

func (e *Message) Copy() *Message {
	ne := &Message{
		CreatedTime: e.CreatedTime,
		From:        e.From,
		To:          e.To,
		Body:        make(map[string]string, len(e.Body)),
		Attachments: make([]MessageAttachment, len(e.Attachments)),
	}

	for k, v := range e.Body {
		ne.Body[k] = v
	}

	for i, p := range e.Attachments {
		ne.Attachments[i].MediaType = p.MediaType
		ne.Attachments[i].Data = make([]byte, len(p.Data))
		copy(ne.Attachments[i].Data, p.Data)
	}

	return ne
}
