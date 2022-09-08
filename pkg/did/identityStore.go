package did

// IdentityStore to provide a means of looking up private identities and use
// them in consensus and DID/VC creation
type IdentityStore interface {
	Find(id string) (PrivateIdentity, error)
	List() ([]PrivateIdentity, error)
}
