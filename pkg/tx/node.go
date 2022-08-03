package tx

type Node struct {
	Id  string `msgpack:"i"`
	Did string `msgpack:"did"`
	Key []byte `msgpack:"k"`
}
