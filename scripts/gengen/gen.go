package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/kubo/config"
	"github.com/multiformats/go-multibase"
	istorage "github.com/tcfw/didem/internal/storage"
	"github.com/tcfw/didem/pkg/cryptography"
	"github.com/tcfw/didem/pkg/did"
	"github.com/tcfw/didem/pkg/did/w3cdid"
	"github.com/tcfw/didem/pkg/storage"
	"github.com/tcfw/didem/pkg/tx"
	"github.com/vmihailenco/msgpack/v5"
)

func tempIPFSStorage(ctx context.Context) (*istorage.IPFSStorage, func(), error) {
	repoPath, err := ioutil.TempDir("", "ipfs-shell")
	if err != nil {
		return nil, nil, err
	}

	id, err := config.CreateIdentity(ioutil.Discard, []options.KeyGenerateOption{options.Key.Type(options.Ed25519Key)})
	if err != nil {
		return nil, nil, err
	}

	ipfs, err := istorage.NewIPFSStorage(ctx, id, repoPath)
	if err != nil {
		return nil, nil, err
	}

	c := func() {
		os.RemoveAll(repoPath)
		ipfs.Close()
	}

	return ipfs, c, nil
}

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

	store, close, err := tempIPFSStorage(context.Background())
	if err != nil {
		panic(err)
	}
	defer close()

	for _, t := range txs {
		msg, err := t.Marshal()
		if err != nil {
			panic(err)
		}
		t.Signature, err = blsSk.Sign(nil, msg, nil)
		if err != nil {
			panic(err)
		}

		cid, err := store.PutTx(ctx, t)
		if err != nil {
			panic(err)
		}
		txCids = append(txCids, cid)
	}

	block, err := storage.NewBlock(ctx, store, storage.BlockID(cid.Undef), txCids)
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
