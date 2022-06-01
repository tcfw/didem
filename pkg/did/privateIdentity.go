package did

import "crypto"

type PrivateIdentity interface {
	PublicID() string
	PrivateKey() crypto.PrivateKey
}
