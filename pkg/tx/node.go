package tx

type Node struct {
	Id  string `msgpack:"i"`
	Did string `msgpack:"d"`
	Key []byte `msgpack:"k"`
}
