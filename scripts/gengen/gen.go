package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/kubo/config"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multibase"
	istorage "github.com/tcfw/didem/internal/storage"
	"github.com/tcfw/didem/pkg/cryptography"
	"github.com/tcfw/didem/pkg/did"
	"github.com/tcfw/didem/pkg/did/w3cdid"
	"github.com/tcfw/didem/pkg/storage"
	"github.com/tcfw/didem/pkg/tx"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	p2pIdents = flag.String("p2p_identities", "", "p2p identites for each node")
)

type identity struct {
	blsSk   *cryptography.Bls12381PrivateKey
	blsPKMB string
	pID     peer.ID
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	config := &storage.GenesisInfo{
		ChainID: "testnet",
	}

	idents := strings.Split(*p2pIdents, ",")

	if len(idents) == 0 {
		panic("no identities provided")
	}

	identities := []identity{}

	for _, i := range idents {
		sk, err := did.GenerateBls12381G2Identity()
		if err != nil {
			panic(err)
		}

		pk, err := sk.PublicIdentity()
		if err != nil {
			panic(err)
		}

		pkmb, err := pk.PublicKeys[0].AsMultibase()
		if err != nil {
			panic(err)
		}

		pid := peer.ID(i)

		identities = append(identities, identity{
			blsSk:   sk.PrivateKey().(*cryptography.Bls12381PrivateKey),
			blsPKMB: pkmb,
			pID:     pid,
		})
	}

	txs := []*tx.Tx{}

	for _, i := range identities {
		txs = append(txs,
			&tx.Tx{
				Version: tx.Version1,
				Ts:      time.Now().Unix(),
				From:    i.pID.Pretty(),
				Type:    tx.TxType_DID,
				Action:  tx.TxActionAdd,
				Data: &tx.DID{
					Document: &w3cdid.Document{
						ID: "did:didem:" + i.pID.String(),
						Authentication: []cryptography.VerificationMethod{
							{
								ID:                 "did:didem:" + i.pID.String(),
								Type:               cryptography.Bls12381G2Key2020,
								PublicKeyMultibase: i.blsPKMB,
							},
						},
					},
				},
			},
			&tx.Tx{
				Version: tx.Version1,
				Ts:      time.Now().Unix(),
				From:    i.pID.String(),
				Type:    tx.TxType_Node,
				Action:  tx.TxActionAdd,
				Data: &tx.Node{
					Id:  i.pID.String(),
					Did: "did:didem:" + i.pID.String(),
					Key: []byte(i.blsPKMB),
				},
			},
		)
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
		t.Signature, err = identities[0].blsSk.Sign(nil, msg, nil)
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
	block.Signature, err = identities[0].blsSk.Sign(nil, blMsg, nil)
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

	fmt.Printf("Genesis Config:\n%s", b64)
	fmt.Printf("\n\nGenesis Config (Hex):\n%s", hex.EncodeToString(b))

	fmt.Printf("\n\nIdentities: \n")
	for i, id := range identities {
		skb, err := id.blsSk.Bytes()
		if err != nil {
			panic(err)
		}

		skbmb, err := multibase.Encode(multibase.Base58BTC, skb)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Identity %d\n\tSK: %s\n\tID: %s\n\tPeer: %s\n\n", i, skbmb, id.pID.Pretty(), idents[i])
	}

}

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
