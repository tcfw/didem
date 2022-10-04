package consensus

type Tracer interface {
	OnMsg(*Msg)
	OnSendMsg(*Msg)
}
