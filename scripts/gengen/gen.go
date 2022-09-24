package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
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

var (
	p2pIDfile = flag.String("id file", "~/.didem/p2p_identity", "p2p identiy file for nodes")
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
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// p2pidFile := expandPath(*p2pIDfile)

	// idB, err := ioutil.ReadFile(p2pidFile)
	// if err != nil {
	// 	panic(errors.Wrap(err, "reading identity file"))
	// }

	// p2ppriv, err := crypto.UnmarshalPrivateKey(idB)
	// if err != nil {
	// 	panic(errors.Wrap(err, "unmarshaling private key"))
	// }

	// nid, err := peer.IDFromPrivateKey(p2ppriv)
	// if err != nil {
	// 	panic(err)
	// }

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
				Id:  "12D3KooWHovX9xtxzV9SkRJUrHZbYHUMCfpiHJgCR2QzULtfywf3",
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

func expandPath(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if path == "~" {
		// In case of "~", which won't be caught by the "else if"
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		path = filepath.Join(dir, path[2:])
	}

	return path
}
