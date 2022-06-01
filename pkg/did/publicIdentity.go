package did

import "crypto"

type PublicIdentity struct {
	ID         string
	PublicKeys []PublicKey
}

type PublicKey struct {
	Key crypto.PublicKey
}
