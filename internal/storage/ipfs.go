package storage

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs/config"
	ipfsCore "github.com/ipfs/go-ipfs/core"
	ipfsCoreiface "github.com/ipfs/go-ipfs/core/coreapi"
	"github.com/ipfs/go-ipfs/core/node/libp2p"
	"github.com/ipfs/go-ipfs/plugin/loader"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	coreiface "github.com/ipfs/interface-go-ipfs-core"
	options "github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/tcfw/didem/pkg/storage"
	"github.com/tcfw/didem/pkg/tx"
)

var (
	_ storage.Storage = (*IPFSStorage)(nil)
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

func (is *IPFSStorage) PutTx(ctx context.Context, transaction *tx.Tx) (tx.ID, error) {
	d, err := transaction.Marshal()
	if err != nil {
		return nil, err
	}

	hashType := options.Block.Hash(multihash.SHA2_256, multihash.DefaultLengths[multihash.SHA2_256])

	n, err := is.node.Block().Put(ctx, bytes.NewReader(d), hashType, options.Block.Pin(true))
	if err != nil {
		return nil, err
	}

	logrus.New().WithField("ipfs", n.Path().String()).Info("stored in IPFS")

	id, err := idFromMultihash(n.Path().Cid().Hash())
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (is *IPFSStorage) GetTx(ctx context.Context, id tx.ID) (*tx.Tx, error) {
	mh, err := idToMultihash(id)
	if err != nil {
		return nil, err
	}

	cid := cid.NewCidV1(multihash.SHA3_384, mh)

	n, err := is.node.Block().Get(ctx, path.IpldPath(cid))
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(n)
	if err != nil {
		return nil, err
	}

	tx := &tx.Tx{}
	if err := tx.Unmarshal(data); err != nil {
		return nil, err
	}

	return tx, nil
}

func (is *IPFSStorage) Tips(ctx context.Context) ([]tx.Tx, error) {
	return nil, errors.New("not implemeneted")
}

func (is *IPFSStorage) Parents(ctx context.Context, id tx.ID) ([]tx.Tx, error) {
	return nil, errors.New("not implemeneted")
}

func idToMultihash(id tx.ID) (multihash.Multihash, error) {
	mhBytes, err := multihash.Encode(id, multihash.SHA3_384)
	if err != nil {
		return nil, errors.Wrap(err, "encoding ID to multihash")
	}

	mh, err := multihash.Cast(mhBytes)
	if err != nil {
		return nil, errors.Wrap(err, "casting ID to multihash")
	}

	return mh, nil
}

func idFromMultihash(mh multihash.Multihash) (tx.ID, error) {
	dmh, err := multihash.Decode(mh)
	if err != nil {
		return nil, err
	}

	return tx.ID(dmh.Digest), nil
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
				log.Printf("failed to connect to %s: %s", peerInfo.ID, err)
			}
		}(peerInfo)
	}
	wg.Wait()
	return nil
}
