package comm

import "github.com/tcfw/didem/pkg/did"

type Message struct {
	Type        string                `msgpack:"y" json:"type"`
	Id          string                `msgpack:"i" json:"id"`
	ThreadId    string                `msgpack:"r" json:"trid,omitempty"`
	CreatedTime int64                 `msgpack:"c" json:"created_time,omitempty"`
	ExpiresTime int64                 `msgpack:"e" json:"expires_time,omitempty"`
	From        *did.PublicIdentity   `msgpack:"f" json:"from"`
	To          []*did.PublicIdentity `msgpack:"t" json:"to"`
	Nonce       [32]byte              `msgpack:"n" json:"none,omitempty"`
	Signature   []byte                `msgpack:"s" json:"signature,omitempty"`
	Body        map[string]string     `msgpack:"h" json:"body,omitempty"`
	Attachments []MessageAttachment   `msgpack:"p" json:"attachments,omitempty"`
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
