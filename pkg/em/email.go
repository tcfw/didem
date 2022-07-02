package em

import "github.com/tcfw/didem/pkg/did"

type Email struct {
	Time      int64               `msgpack:"e"`
	From      *did.PublicIdentity `msgpack:"f"`
	To        *did.PublicIdentity `msgpack:"t"`
	Nonce     [32]byte            `msgpack:"n"`
	Signature []byte              `msgpack:"s"`
	Headers   map[string]string   `msgpack:"h"`
	Parts     []EmailPart         `msgpack:"p"`
}

type EmailPart struct {
	Mime string `msgpack:"m"`
	Data []byte `msgpack:"d"`
}

func (e *Email) Copy() *Email {
	ne := &Email{
		Time:    e.Time,
		From:    e.From,
		To:      e.To,
		Headers: make(map[string]string, len(e.Headers)),
		Parts:   make([]EmailPart, len(e.Parts)),
	}

	for k, v := range e.Headers {
		ne.Headers[k] = v
	}

	for i, p := range e.Parts {
		ne.Parts[i].Mime = p.Mime
		ne.Parts[i].Data = make([]byte, len(p.Data))
		copy(ne.Parts[i].Data, p.Data)
	}

	return ne
}
