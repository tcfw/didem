package did

type Resolver interface {
	Find(id string) (*PublicIdentity, error)
}
