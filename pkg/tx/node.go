package tx

type Node struct {
	Id   string   `msgpack:"i"`
	Did  string   `msgpack:"did"`
	Keys [][]byte `msgpack:"k"`
}
