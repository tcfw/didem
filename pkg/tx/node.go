package tx

import (
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

type Node struct {
	Id   peer.ID
	Did  string
	Keys []crypto.PubKey
}
