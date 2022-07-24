package consensus

import "github.com/ipfs/go-cid"

type BlockID string

type Block struct {
	Version   uint32  `msgpack:"v"`
	ID        BlockID `magpack:"i"`
	Parent    BlockID `msgpack:"p"`
	Height    uint64  `msgpack:"h"`
	CreatedAt uint64  `msgpack:"t"`
	Proposer  string  `msgpack:"w"`
	Signers   uint32  `msgpack:"sn"`
	Signature []byte  `msgpack:"s"`
	Nonce     []byte  `msgpack:"n"`
	TxRoot    cid.Cid `msgpack:"x"`
}

type TxTrie struct {
	Children []cid.Cid `msgpack:"c"`
	Tx       cid.Cid   `msgpack:"t"`
}

type Tx struct{}
