package did

import "crypto"

type PublicIdentity struct {
	ID         string      `msgpack:"i"`
	PublicKeys []PublicKey `msgpack:"p"`
}

type PublicKey struct {
	Key crypto.PublicKey `msgpack:"k"`
}

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
