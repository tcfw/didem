package genesis

import (
	"time"

	"github.com/ipfs/go-cid"
	"github.com/tcfw/didem/pkg/tx"
)

type Info struct {
	ChainID   string
	CreatedAt time.Time
	Tx        []*tx.Tx
	TxSetData map[cid.Cid][]byte
}
