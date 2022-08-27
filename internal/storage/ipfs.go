package storage

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/cockroachdb/pebble"
	"github.com/ipfs/go-cid"
	coreIface "github.com/ipfs/interface-go-ipfs-core"
	options "github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/ipfs/kubo/config"
	ipfsCore "github.com/ipfs/kubo/core"
	ipfsCoreApiIface "github.com/ipfs/kubo/core/coreapi"
	"github.com/ipfs/kubo/core/node/libp2p"
	"github.com/ipfs/kubo/plugin/loader"
	"github.com/ipfs/kubo/repo/fsrepo"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/tcfw/didem/internal/coinflip"
	"github.com/tcfw/didem/internal/utils/logging"
	"github.com/tcfw/didem/pkg/did/genesis"
	"github.com/tcfw/didem/pkg/did/w3cdid"
	"github.com/tcfw/didem/pkg/storage"
	"github.com/tcfw/didem/pkg/tx"
)

var (
	_           storage.Store = (*IPFSStorage)(nil)
	ErrNotFound               = errors.New("not found")

	bootstrapNodes = []string{
		// IPFS Bootstrapper nodes.
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
	}
)

const (
	cacheSize = 1 << 20 * 100

	tableSep           byte = ':'
	tableSepUpperBound      = tableSep + 1

	allTxNWorkers = 3
)

type metadataKeyType byte

const (
	txBlockTPrefix metadataKeyType = iota + 1
	blockStateTPrefix
	didTPrefix
	didHisoryTPrefix
	activeClaimTPrefix
	nodesTPrefix
	genesisTPrefix
)

func NewIPFSStorage(ctx context.Context, id config.Identity, repo string) (*IPFSStorage, error) {
	ipfsRepo := filepath.Join(repo, "ipfs")
	ipfs, close, err := ipfsStore(ctx, id, ipfsRepo)
	if err != nil {
		return nil, errors.Wrap(err, "opening ipfs block store")
	}

	metadataRepo := filepath.Join(repo, "metadata")
	m, err := metadataStore(ctx, metadataRepo)
	if err != nil {
		return nil, errors.Wrap(err, "opening metadata store")
	}

	s := &IPFSStorage{
		ipfsNode: ipfs,
		metadata: m,
		Close:    close,
	}

	return s, nil
}

var (
	ipfsSetup sync.Once
)

func ipfsStore(ctx context.Context, id config.Identity, repo string) (coreIface.CoreAPI, func() error, error) {
	var err error

	ipfsSetup.Do(func() {
		err = setupPlugins("")
	})

	if err != nil {
		return nil, nil, err
	}

	if err := createRepo(id, repo); err != nil {
		return nil, nil, errors.Wrap(err, "checking/creating ipfs repo")
	}

	cfg, err := newIpfsCfg(repo)
	if err != nil {
		return nil, nil, errors.Wrap(err, "creating ipfs config from repo")
	}

	node, err := ipfsCore.NewNode(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	iface, err := ipfsCoreApiIface.NewCoreAPI(node)
	if err != nil {
		return nil, nil, err
	}

	go func() {
		connectToPeers(ctx, iface, bootstrapNodes)
	}()

	return iface, node.Close, nil
}

func metadataStore(ctx context.Context, repo string) (*pebble.DB, error) {
	c := pebble.NewCache(cacheSize)
	tc := pebble.NewTableCache(c, 16, 100)
	defer tc.Unref()
	defer c.Unref()

	return pebble.Open(repo, &pebble.Options{Cache: c, TableCache: tc})
}

func newIpfsCfg(path string) (*ipfsCore.BuildCfg, error) {
	c := &ipfsCore.BuildCfg{
		Online:  true,
		Routing: libp2p.DHTServerOption,
	}

	repo, err := fsrepo.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "opening ipfs repo")
	}

	c.Repo = repo

	return c, nil
}

type IPFSStorage struct {
	ipfsNode coreIface.CoreAPI
	metadata *pebble.DB

	Close func() error
}

func (s *IPFSStorage) putRaw(ctx context.Context, d []byte) (cid.Cid, error) {
	hashType := options.Block.Hash(multihash.SHA2_256, multihash.DefaultLengths[multihash.SHA2_256])

	n, err := s.ipfsNode.Block().Put(ctx, bytes.NewReader(d), hashType, options.Block.Pin(true))
	if err != nil {
		return cid.Undef, err
	}

	logging.Entry().WithField("ipfs", n.Path().String()).Info("stored in IPFS")

	return n.Path().Cid(), nil
}

func (s *IPFSStorage) getRaw(ctx context.Context, id cid.Cid) ([]byte, error) {
	n, err := s.ipfsNode.Block().Get(ctx, path.IpfsPath(id))
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(n)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *IPFSStorage) PutTx(ctx context.Context, transaction *tx.Tx) (cid.Cid, error) {
	d, err := transaction.Marshal()
	if err != nil {
		return cid.Undef, err
	}

	return s.putRaw(ctx, d)
}

func (s *IPFSStorage) GetTx(ctx context.Context, id tx.TxID) (*tx.Tx, error) {
	data, err := s.getRaw(ctx, cid.Cid(id))
	if err != nil {
		return nil, err
	}

	tx := &tx.Tx{}
	if err := tx.Unmarshal(data); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *IPFSStorage) PutBlock(ctx context.Context, b *storage.Block) (cid.Cid, error) {
	d, err := msgpack.Marshal(b)
	if err != nil {
		return cid.Undef, errors.Wrap(err, "mashaling tx")
	}

	return s.putRaw(ctx, d)
}

func (s *IPFSStorage) GetBlock(ctx context.Context, id storage.BlockID) (*storage.Block, error) {
	data, err := s.getRaw(ctx, cid.Cid(id))
	if err != nil {
		return nil, err
	}

	b := &storage.Block{}
	if err := msgpack.Unmarshal(data, b); err != nil {
		return nil, errors.Wrap(err, "unmarshalling block")
	}

	return b, nil
}

func (s *IPFSStorage) PutSet(ctx context.Context, txs *storage.TxSet) (cid.Cid, error) {
	d, err := msgpack.Marshal(txs)
	if err != nil {
		return cid.Undef, errors.Wrap(err, "mashaling tx")
	}

	c, err := s.putRaw(ctx, d)
	if err != nil {
		return cid.Undef, errors.Wrap(err, "storing txset")
	}

	if coinflip.Flip() {
		if err := s.ipfsNode.Pin().Add(ctx, path.IpfsPath(c)); err != nil {
			return cid.Undef, errors.Wrap(err, "pinning txset")
		}
	}

	return c, nil
}

func (s *IPFSStorage) GetSet(ctx context.Context, id cid.Cid) (*storage.TxSet, error) {
	data, err := s.getRaw(ctx, id)
	if err != nil {
		return nil, err
	}

	txs := &storage.TxSet{}
	if err := msgpack.Unmarshal(data, txs); err != nil {
		return nil, errors.Wrap(err, "unmarshalling block")
	}

	return txs, nil
}

func (s *IPFSStorage) GetTxBlock(ctx context.Context, id tx.TxID) (*storage.Block, error) {
	bcid, d, err := s.metadata.Get(typedKey(txBlockTPrefix, id.String()))
	if err != nil {
		return nil, errors.Wrap(err, "looking up tx block")
	}
	defer d.Close()

	cid, err := cid.Cast(bcid)
	if err != nil {
		return nil, errors.Wrap(err, "casting tx block cid")
	}

	return s.GetBlock(ctx, storage.BlockID(cid))
}

func (s *IPFSStorage) MarkBlock(ctx context.Context, id storage.BlockID, state storage.BlockState) error {
	k := typedKey(blockStateTPrefix, id.String())

	cStateB, done, err := s.metadata.Get(k)
	if err != nil && err != pebble.ErrNotFound {
		return errors.Wrap(err, "looking up block state")
	}
	if err != pebble.ErrNotFound {
		defer done.Close()
	} else {
		cStateB = make([]byte, 4)
	}

	cState := storage.BlockState(binary.LittleEndian.Uint32(cStateB))

	nStateB := make([]byte, 4)
	binary.LittleEndian.PutUint32(nStateB, uint32(state))
	if err := s.metadata.Set(k, nStateB, nil); err != nil {
		return errors.Wrap(err, "setting block state")
	}

	if cState == storage.BlockStateValidated && state == storage.BlockStateAccepted {
		if err := s.indexBlock(ctx, id); err != nil {
			return errors.Wrap(err, "indexing block")
		}
	}

	return nil
}

func (s *IPFSStorage) indexBlock(ctx context.Context, bId storage.BlockID) error {
	b, err := s.GetBlock(ctx, bId)
	if err != nil {
		return errors.Wrap(err, "getting block")
	}

	txs, err := s.AllTx(ctx, b)
	if err != nil {
		return errors.Wrap(err, "getting block transactions")
	}

	if err := s.ipfsNode.Pin().Add(ctx, path.IpfsPath(cid.Cid(bId))); err != nil {
		return errors.Wrap(err, "pinning block")
	}

	batch := s.metadata.NewBatch()
	for id, t := range txs {
		if err := batch.Set(typedKey(txBlockTPrefix, cid.Cid(id).String()), cid.Cid(bId).Bytes(), nil); err != nil {
			return errors.Wrap(err, "a")
		}

		switch t.Type {
		case tx.TxType_Node:
			err = s.indexTxNode(batch, t, id)
		case tx.TxType_DID:
			err = s.indexTxDID(batch, t, id)
		case tx.TxType_VC:
			err = s.indexTxVC(batch, t, id)
		default:
			batch.Close()
			return errors.New("unsupported tx type")
		}
		if err != nil {
			batch.Close()
			return errors.Wrap(err, fmt.Sprintf("indexing tx %s", id))
		}

		//50% chance of storing the Tx
		if coinflip.Flip() {
			if err := s.ipfsNode.Pin().Add(ctx, path.IpfsPath(cid.Cid(id))); err != nil {
				return errors.Wrap(err, "pinning block")
			}
		}
	}

	if err := batch.Commit(&pebble.WriteOptions{}); err != nil {
		return errors.Wrap(err, "applying metadata batch index")
	}

	return nil
}

func (s *IPFSStorage) indexTxNode(b *pebble.Batch, ntx *tx.Tx, id tx.TxID) error {
	c := ntx.Data.(*tx.Node)
	k := typedKey(nodesTPrefix, c.Id)

	switch ntx.Action {
	case tx.TxActionAdd:
		return b.Set(k, cid.Cid(id).Bytes(), nil)
	case tx.TxActionRevoke:
		return b.Delete(k, nil)
	default:
		return fmt.Errorf("unsupported action %d", ntx.Action)
	}

}

func (s *IPFSStorage) indexTxDID(b *pebble.Batch, dtx *tx.Tx, id tx.TxID) error {
	c := dtx.Data.(*tx.DID)
	k := typedKey(didTPrefix, c.Document.ID)

	//append to did history
	ak := typedKey(didHisoryTPrefix, c.Document.ID, id.String())
	if err := b.Set(ak, cid.Cid(id).Bytes(), nil); err != nil {
		return errors.Wrap(err, "appending to did history")
	}

	switch dtx.Action {
	case tx.TxActionAdd, tx.TxActionUpdate:
		return b.Set(k, cid.Cid(id).Bytes(), nil)
	case tx.TxActionRevoke:
		return b.Delete(k, nil)
	default:
		return fmt.Errorf("unsupported action %d", dtx.Action)
	}
}

func (s *IPFSStorage) indexTxVC(b *pebble.Batch, vtx *tx.Tx, id tx.TxID) error {
	//todo append to did history

	return fmt.Errorf("not implemented")
}

func (s *IPFSStorage) AllTx(ctx context.Context, b *storage.Block) (map[tx.TxID]*tx.Tx, error) {
	txSeen := map[tx.TxID]*tx.Tx{}
	visited := cid.NewSet()
	queue := []cid.Cid{b.TxRoot}

	for len(queue) != 0 {
		//pop
		setCid := queue[0]
		queue = queue[1:]

		//just incase we encounter a loop (bad proposer?)
		if visited.Has(setCid) {
			continue
		} else {
			visited.Add(setCid)
		}

		set, err := s.GetSet(ctx, setCid)
		if err != nil {
			return nil, errors.Wrap(err, "getting root trie")
		}

		if set.Tx != nil {
			t, err := s.GetTx(ctx, tx.TxID(*set.Tx))
			if err != nil {
				return nil, errors.Wrap(err, "getting root tx")
			}

			txSeen[tx.TxID(*set.Tx)] = t

			//max check
			if len(txSeen) > storage.MaxBlockTxCount {
				return nil, errors.New("block containers too many tx")
			}
		}

		//push
		if len(set.Children) > 0 {
			queue = append(queue, set.Children...)
		}
	}

	return txSeen, nil
}

func (s *IPFSStorage) LookupDID(ctx context.Context, did string) (*w3cdid.Document, error) {
	k := typedKey(didTPrefix, did)

	txB, done, err := s.metadata.Get(k)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "finding did metadata key")
	}
	defer done.Close()

	cid, err := cid.Parse(txB)
	if err != nil {
		return nil, errors.Wrap(err, "casting did tx cid")
	}

	t, err := s.GetTx(ctx, tx.TxID(cid))
	if err != nil {
		return nil, errors.Wrap(err, "lookup id tx")
	}

	//check type
	if t.Type != tx.TxType_DID {
		return nil, errors.New("unexpected tx type")
	}

	data := t.Data.(*tx.DID)

	return data.Document, nil
}

func (s *IPFSStorage) DIDHistory(ctx context.Context, id string) ([]*tx.Tx, error) {
	hIter := s.metadata.NewIter(&pebble.IterOptions{
		LowerBound: []byte{byte(didHisoryTPrefix)},
		UpperBound: []byte{byte(didHisoryTPrefix) + 1},
	})
	hCids := cid.NewSet()

	for hIter.First(); hIter.Valid(); hIter.Next() {
		c, err := cid.Cast(hIter.Value())
		if err == nil {
			hCids.Add(c)
		}
	}
	defer hIter.Close()

	var mu sync.Mutex
	var wg sync.WaitGroup
	var errC chan error
	cidC := make(chan cid.Cid, hCids.Len())
	done := make(chan struct{})

	wg.Add(hCids.Len())

	go func() {
		for _, c := range hCids.Keys() {
			cidC <- c
		}

		wg.Wait()
		close(cidC)
		done <- struct{}{}
	}()

	txs := make([]*tx.Tx, 0, hCids.Len())

	for i := 0; i < allTxNWorkers; i++ {
		go func() {
			for c := range cidC {
				tx, err := s.GetTx(ctx, tx.TxID(c))
				if err != nil {
					wg.Done()
					errC <- err
					continue
				}

				mu.Lock()
				txs = append(txs, tx)
				mu.Unlock()

				wg.Done()
			}
		}()
	}

	select {
	case err := <-errC:
		return nil, errors.Wrap(err, "fetching tx")
	case <-done:
	}

	return txs, nil
}

func (s *IPFSStorage) Claims(ctx context.Context, did string) ([]*tx.Tx, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *IPFSStorage) Nodes() ([]string, error) {
	list := []string{}
	iter := s.metadata.NewIter(&pebble.IterOptions{
		LowerBound: []byte{byte(nodesTPrefix)},
		UpperBound: []byte{byte(nodesTPrefix) + 1},
	})
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		v := iter.Key()
		list = append(list, string(v[1:]))
	}

	return list, nil
}

func (s *IPFSStorage) HasGenesisApplied() bool {
	_, done, err := s.metadata.Get(typedKey(genesisTPrefix))
	if err != nil {
		if err != pebble.ErrNotFound {
			logging.Entry().Warn("failed reading genesis prefix")
		}
		return false
	}
	defer done.Close()
	return true
}

func (s *IPFSStorage) ApplyGenesis(g *genesis.Info) error {
	//TODO(tcfw)

	if err := s.metadata.Set(typedKey(genesisTPrefix), []byte{}, nil); err != nil {
		return errors.Wrap(err, "storing genesis prefix")
	}

	return nil
}

func (s *IPFSStorage) Node(ctx context.Context, id string) (*tx.Node, error) {
	key := typedKey(nodesTPrefix, id)
	v, done, err := s.metadata.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "looking up node key")
	}
	defer done.Close()

	cid, err := cid.Parse(v)
	if err != nil {
		return nil, errors.Wrap(err, "casting node cid")
	}

	nodeTx, err := s.GetTx(ctx, tx.TxID(cid))
	if err != nil {
		return nil, errors.Wrap(err, "getting node tx")
	}

	if nodeTx.Type != tx.TxType_Node {
		return nil, errors.New("unexpected tx type")
	}

	if nodeTx.Action == tx.TxActionRevoke {
		return nil, ErrNotFound
	}

	return nodeTx.Data.(*tx.Node), nil
}

func typedKey(kType metadataKeyType, parts ...string) []byte {
	n := 1
	for _, p := range parts {
		n += len(p) + 1 //add sep as well
	}

	k := make([]byte, 0, n)
	k = append(k, byte(kType))
	for _, p := range parts {
		k = append(k, []byte(p)...)
		k = append(k, tableSep)
	}

	return k[:len(k)-1]
}

func createRepo(identity config.Identity, path string) error {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return errors.Wrap(err, "creating repo dir")
		}
	}

	defaultConfig, err := config.InitWithIdentity(identity)
	if err != nil {
		return errors.Wrap(err, "creating default config")
	}

	err = fsrepo.Init(path, defaultConfig)
	if err != nil {
		return errors.Wrap(err, "failed to init p2p config")
	}

	return nil
}

func setupPlugins(externalPluginsPath string) error {
	// Load any external plugins if available on externalPluginsPath
	plugins, err := loader.NewPluginLoader(filepath.Join(externalPluginsPath, "plugins"))
	if err != nil {
		return errors.Wrap(err, "error loading plugins")
	}

	// Load preloaded and external plugins
	if err := plugins.Initialize(); err != nil {
		return errors.Wrap(err, "error initializing plugins")
	}

	if err := plugins.Inject(); err != nil {
		return errors.Wrap(err, "error initializing plugins")
	}

	return nil
}

func connectToPeers(ctx context.Context, ipfs coreIface.CoreAPI, peers []string) error {
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
