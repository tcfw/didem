package did

import "crypto"

// PublicIdentity holds the publicly available form of any secret key
type PublicIdentity struct {
	ID         string      `msgpack:"i" json:"."`
	PublicKeys []PublicKey `msgpack:"p"`
}

// PublicKey a public key identity
type PublicKey struct {
	Key crypto.PublicKey `msgpack:"k"`
}

// Matches checks if a public identity matches another public identity by
// comparing if any public key is the same
func (pi *PublicIdentity) Matches(p *PublicIdentity) bool {
	if pi.ID != p.ID {
		return false
	}

	for _, pipk := range pi.PublicKeys {
		var f bool
		for _, ppk := range p.PublicKeys {
			if ppk.Key == pipk.Key {
				f = true
			}
		}
		if !f {
			return false
		}
	}
	return true
}
