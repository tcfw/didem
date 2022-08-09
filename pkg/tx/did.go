package tx

import "github.com/tcfw/didem/pkg/did/w3cdid"

type DID struct {
	Document *w3cdid.Document `msgpack:"d"`
}
