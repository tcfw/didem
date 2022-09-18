package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multibase"
	"github.com/tcfw/didem/pkg/cryptography"
	"github.com/tcfw/didem/pkg/did"
	"github.com/tcfw/didem/pkg/did/w3cdid"
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

	sk, err := did.GenerateBls12381G2Identity()
	if err != nil {
		panic(err)
	}
	blsSk := sk.PrivateKey().(*cryptography.Bls12381PrivateKey)

	pk, err := sk.PublicIdentity()
	if err != nil {
		panic(err)
	}

	pkmb, err := pk.PublicKeys[0].AsMultibase()
	if err != nil {
		panic(err)
	}

	txs := []*tx.Tx{
		{
			Version: tx.Version1,
			Ts:      time.Now().Unix(),
			From:    pk.ID,
			Type:    tx.TxType_DID,
			Action:  tx.TxActionAdd,
			Data: &tx.DID{
				Document: &w3cdid.Document{
					ID: "did:didem:" + pk.ID,
					Authentication: []cryptography.VerificationMethod{
						{
							ID:                 "did:didem:" + pk.ID,
							Type:               cryptography.Bls12381G2Key2020,
							PublicKeyMultibase: pkmb,
						},
					},
				},
			},
		},
		{
			Version: tx.Version1,
			Ts:      time.Now().Unix(),
			From:    pk.ID,
			Type:    tx.TxType_Node,
			Action:  tx.TxActionAdd,
			Data: &tx.Node{
				Id:  pk.ID,
				Did: "did:didem:" + pk.ID,
				Key: []byte(pkmb),
			},
		},
	}
	txCids := []cid.Cid{}

	memstore := storage.NewMemStore()

	for _, t := range txs {
		msg, err := t.Marshal()
		if err != nil {
			panic(err)
		}
		t.Signature, err = blsSk.Sign(nil, msg, nil)
		if err != nil {
			panic(err)
		}

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

	blMsg, err := msgpack.Marshal(block)
	if err != nil {
		panic(err)
	}
	block.Signature, err = blsSk.Sign(nil, blMsg, nil)
	if err != nil {
		panic(err)
	}

	config.Block = *block
	config.Txs = txs

	b, err := msgpack.Marshal(config)
	if err != nil {
		panic(err)
	}

	b64, err := multibase.Encode(multibase.Base58BTC, b)
	if err != nil {
		panic(err)
	}

	blsSkB, err := blsSk.Bytes()
	if err != nil {
		panic(err)
	}

	s, err := multibase.Encode(multibase.Base58BTC, blsSkB)
	if err != nil {
		panic(err)
	}

	fmt.Printf("SK: %s\n", s)

	fmt.Printf("Genesis Config:\n%s", b64)
}
