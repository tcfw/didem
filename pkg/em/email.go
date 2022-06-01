package em

import "github.com/tcfw/didem/pkg/did"

type Email struct {
	From          *did.PublicIdentity
	To            *did.PublicIdentity
	Nonce         []byte
	SignatureFrom []byte
	Headers       map[string]string
	Parts         [][]byte
}
