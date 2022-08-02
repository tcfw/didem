package storage

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"

	"github.com/ipfs/go-cid"
	coreiface "github.com/ipfs/interface-go-ipfs-core"
	options "github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/ipfs/kubo/config"
	ipfsCore "github.com/ipfs/kubo/core"
	ipfsCoreiface "github.com/ipfs/kubo/core/coreapi"
	"github.com/ipfs/kubo/core/node/libp2p"
	"github.com/ipfs/kubo/plugin/loader"
	"github.com/ipfs/kubo/repo/fsrepo"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/tcfw/didem/internal/utils/logging"
	"github.com/tcfw/didem/pkg/storage"
	"github.com/tcfw/didem/pkg/tx"
)

var (
	_ storage.Store = (*IPFSStorage)(nil)
)

func NewIPFSStorage(ctx context.Context) (*IPFSStorage, error) {
	if err := setupPlugins(""); err != nil {
		return nil, err
	}

	cfg := newIpfsCfg()

	node, err := ipfsCore.NewNode(ctx, cfg)
	if err != nil {
		return nil, err
	}

	iface, err := ipfsCoreiface.NewCoreAPI(node)
	if err != nil {
		return nil, err
	}

	go func() {
		bootstrapNodes := []string{
			// IPFS Bootstrapper nodes.
			"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
			"/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
			"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
			"/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
		}

		connectToPeers(ctx, iface, bootstrapNodes)
	}()

	s := &IPFSStorage{
		node: iface,
	}

	return s, nil
}

func newIpfsCfg() *ipfsCore.BuildCfg {
	c := &ipfsCore.BuildCfg{}

	c.NilRepo = true
	c.Online = true
	c.Routing = libp2p.DHTOption

	// repoPath, err := createTempRepo()
	// if err != nil {
	// 	panic(err)
	// }

	// logrus.New().WithField("repo", repoPath).Infof("IPFS repo")

	// repo, err := fsrepo.Open(repoPath)
	// if err != nil {
	// 	panic(err)
	// }

	// c.Repo = repo

	return c
}

type IPFSStorage struct {
	node coreiface.CoreAPI
}

func (is *IPFSStorage) putRaw(ctx context.Context, d []byte) (cid.Cid, error) {
	hashType := options.Block.Hash(multihash.SHA2_256, multihash.DefaultLengths[multihash.SHA2_256])

	n, err := is.node.Block().Put(ctx, bytes.NewReader(d), hashType, options.Block.Pin(true))
	if err != nil {
		return cid.Undef, err
	}

	logging.Entry().WithField("ipfs", n.Path().String()).Info("stored in IPFS")

	return n.Path().Cid(), nil
}

func (is *IPFSStorage) getRaw(ctx context.Context, id cid.Cid) ([]byte, error) {
	n, err := is.node.Block().Get(ctx, path.IpldPath(id))
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(n)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (is *IPFSStorage) PutTx(ctx context.Context, transaction *tx.Tx) (cid.Cid, error) {
	d, err := transaction.Marshal()
	if err != nil {
		return cid.Undef, err
	}

	return is.putRaw(ctx, d)
}

func (is *IPFSStorage) GetTx(ctx context.Context, id tx.TxID) (*tx.Tx, error) {
	data, err := is.getRaw(ctx, cid.Cid(id))
	if err != nil {
		return nil, err
	}

	tx := &tx.Tx{}
	if err := tx.Unmarshal(data); err != nil {
		return nil, err
	}

	return tx, nil
}

func (is *IPFSStorage) PutBlock(ctx context.Context, b *storage.Block) (cid.Cid, error) {
	d, err := msgpack.Marshal(b)
	if err != nil {
		return cid.Undef, errors.Wrap(err, "mashaling tx")
	}

	return is.putRaw(ctx, d)
}

func (is *IPFSStorage) GetBlock(ctx context.Context, id storage.BlockID) (*storage.Block, error) {
	data, err := is.getRaw(ctx, cid.Cid(id))
	if err != nil {
		return nil, err
	}

	b := &storage.Block{}
	if err := msgpack.Unmarshal(data, b); err != nil {
		return nil, errors.Wrap(err, "unmarshalling block")
	}

	return b, nil
}

func (is *IPFSStorage) PutSet(ctx context.Context, txs *storage.TxSet) (cid.Cid, error) {
	d, err := msgpack.Marshal(txs)
	if err != nil {
		return cid.Undef, errors.Wrap(err, "mashaling tx")
	}

	return is.putRaw(ctx, d)
}

func (is *IPFSStorage) GetSet(ctx context.Context, id cid.Cid) (*storage.TxSet, error) {

	data, err := is.getRaw(ctx, id)
	if err != nil {
		return nil, err
	}

	txs := &storage.TxSet{}
	if err := msgpack.Unmarshal(data, txs); err != nil {
		return nil, errors.Wrap(err, "unmarshalling block")
	}

	return txs, nil
}

func createTempRepo() (string, error) {
	repoPath, err := ioutil.TempDir("", "ipfs-shell")
	if err != nil {
		return "", fmt.Errorf("failed to get temp dir: %s", err)
	}

	// Create a config with default options and a 2048 bit key
	cfg, err := config.Init(ioutil.Discard, 2048)
	if err != nil {
		return "", err
	}

	// Create the repo with the config
	err = fsrepo.Init(repoPath, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to init ephemeral node: %s", err)
	}

	return repoPath, nil
}

func setupPlugins(externalPluginsPath string) error {
	// Load any external plugins if available on externalPluginsPath
	plugins, err := loader.NewPluginLoader(filepath.Join(externalPluginsPath, "plugins"))
	if err != nil {
		return fmt.Errorf("error loading plugins: %s", err)
	}

	// Load preloaded and external plugins
	if err := plugins.Initialize(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	if err := plugins.Inject(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	return nil
}

func connectToPeers(ctx context.Context, ipfs coreiface.CoreAPI, peers []string) error {
	var wg sync.WaitGroup
	peerInfos := make(map[peer.ID]*peer.AddrInfo, len(peers))
	for _, addrStr := range peers {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			return err
		}
		pii, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return err
		}
		pi, ok := peerInfos[pii.ID]
		if !ok {
			pi = &peer.AddrInfo{ID: pii.ID}
			peerInfos[pi.ID] = pi
		}
		pi.Addrs = append(pi.Addrs, pii.Addrs...)
	}

	wg.Add(len(peerInfos))
	for _, peerInfo := range peerInfos {
		go func(peerInfo *peer.AddrInfo) {
			defer wg.Done()
			err := ipfs.Swarm().Connect(ctx, *peerInfo)
			if err != nil {
				logging.Entry().Printf("failed to connect to %s: %s", peerInfo.ID, err)
			}
		}(peerInfo)
	}
	wg.Wait()
	return nil
}
