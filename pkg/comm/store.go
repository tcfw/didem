package comm

type Store interface {
	AddToInbox(*Message) error
}
