package em

type Store interface {
	AddToInbox(*Email) error
}
