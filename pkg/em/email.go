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
