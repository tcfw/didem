package did

type IdentityStore interface {
	Find(id string) (*PrivateIdentity, error)
}
