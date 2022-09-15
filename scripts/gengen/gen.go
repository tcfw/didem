package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/tcfw/didem/pkg/storage"
	"github.com/tcfw/didem/pkg/tx"
	"github.com/vmihailenco/msgpack/v5"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	config := &storage.GenesisInfo{
		ChainID: "testnet",
	}

	txs := []*tx.Tx{
		{
			Version: tx.Version1,
			Ts:      time.Now().Unix(),
			From:    "",
			Type:    tx.TxType_Node,
			Action:  tx.TxActionAdd,
			Data: &tx.Node{
				Id:  "",
				Did: "",
			},
		},
	}
	txCids := []cid.Cid{}

	memstore := storage.NewMemStore()

	for _, t := range txs {
		cid, err := memstore.PutTx(ctx, t)
		if err != nil {
			panic(err)
		}
		txCids = append(txCids, cid)
	}

	block, err := storage.NewBlock(ctx, memstore, storage.BlockID(cid.Undef), txCids)
	if err != nil {
		panic(err)
	}

	config.Block = *block
	config.Txs = txs

	b, err := msgpack.Marshal(config)
	if err != nil {
		panic(err)
	}

	b64 := base64.StdEncoding.EncodeToString(b)

	fmt.Printf("Genesis Config:\n%s", b64)
}
