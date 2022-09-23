package did

import (
	"context"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/jpillora/backoff"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multibase"
	"github.com/pkg/errors"
	"github.com/tcfw/didem/api"
	"github.com/tcfw/didem/internal/stream"
	"github.com/tcfw/didem/internal/utils/logging"
	"github.com/tcfw/didem/pkg/cryptography"
	"github.com/tcfw/didem/pkg/did/consensus"
	"github.com/tcfw/didem/pkg/node"
	"github.com/tcfw/didem/pkg/storage"
	"google.golang.org/protobuf/proto"
)

const (
	ProtocolID = protocol.ID("/tcfw/did/0.0.1")

	tipSetPeerCountStartThreshold = 10
	tipSetPeerCount               = 5
)

type Handler struct {
	n node.Node

	consensus *consensus.Consensus
	validator storage.Validator
}

func NewHandler(n node.Node) *Handler {
	h := n.P2P().Host()
	p := n.P2P().PubSub()

	chainCfg := n.Cfg().Chain()

	if chainCfg.Genesis.ChainID == "" {
		logging.Entry().Warn("no chain genesis loaded")
		return nil
	}

	validator := storage.NewTxValidator(n.Storage())

	var key *cryptography.Bls12381PrivateKey

	if chainCfg.Key != "" {
		_, keyRaw, err := multibase.Decode(chainCfg.Key)
		if err != nil {
			logging.WithError(err).Fatal("unable to decode chain key")
		}

		key, err = cryptography.NewBls12391PrivateKeyFromBytes(keyRaw)
		if err != nil {
			logging.WithError(err).Fatal("unable to decode BLS12381 key")
		}
	}

	opts := []consensus.Option{
		consensus.WithSigningKey(key),
		consensus.WithBlockStore(n.Storage()),
		consensus.WithBeaconSource(n.RandomSource()),
		consensus.WithValidator(validator),
	}

	if chainCfg.Genesis.ChainID != "" {
		opts = append(opts, consensus.WithGenesis(&chainCfg.Genesis))
	}

	c, err := consensus.NewConsensus(h, p, opts...)
	if err != nil {
		logging.Entry().Fatal(err)
	}

	return &Handler{n: n, consensus: c}
}

func (h *Handler) Start() error {
	var peers []peer.ID
	bo := &backoff.Backoff{
		Min: 5 * time.Second,
		Max: 5 * time.Minute,
	}

	requiredPeers := tipSetPeerCountStartThreshold

	if strings.HasPrefix(h.consensus.ChainID(), "test") {
		//allow peer threshold for all testnets
		requiredPeers = 1
	}

	for {
		cm_peers := h.n.P2P().Peers()

		//filter peers with drand tags since they won't implement the didem protocols
		for _, p := range cm_peers {
			info := h.n.P2P().Host().ConnManager().GetTagInfo(p)
			if _, ok := info.Tags["drand"]; !ok {
				peers = append(peers, p)
			}
		}

		if len(peers) < requiredPeers {
			d := bo.Duration()
			logging.Entry().
				WithField("waiting", d).
				WithField("have", len(peers)).
				WithField("need", requiredPeers).
				Info("waiting for more peers")

			time.Sleep(d)
			continue
		}

		break
	}

	rand.Shuffle(len(peers), func(i, j int) { peers[i], peers[j] = peers[j], peers[i] })

	vPeers := peers[:requiredPeers]

	logging.Entry().Debug("got enough peers to check tip")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tips, err := h.askForTip(ctx, vPeers)
	if err != nil {
		return errors.Wrap(err, "requesting tips")
	}

	if len(tips) != requiredPeers || len(tips) != 1 {
		logging.Entry().Warn("failed to get unanimous required tips, trying again in 10s")
		time.Sleep(10 * time.Second)
		return h.Start()
	}

	tip := ""
	for t := range tips {
		tip = t
	}

	logging.Entry().Debug("Got tip as ", tip)

	if tip == h.consensus.State().Block.String() {
		logging.Entry().Info("Starting consensus as matching tip")

		if err := h.consensus.Start(); err != nil {
			return errors.Wrap(err, "starting consensus")
		}

		return nil
	}

	bcid, err := cid.Parse(tip)
	if err != nil {
		logging.WithError(err).Error("parsing chain tip cid")
		return err
	}

	if err := h.validator.ApplyFromTip(context.Background(), storage.BlockID(bcid)); err != nil {
		logging.WithError(err).Error("applying updated tip")
		return err
	}

	logging.Entry().Info("Finished checking tip")

	if err := h.consensus.Start(); err != nil {
		return errors.Wrap(err, "starting consensus")
	}

	return nil
}

func (h *Handler) askForTip(ctx context.Context, peers []peer.ID) (map[string]int, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(len(peers))
	tips := map[string]int{}

	pch := make(chan peer.ID, len(peers))
	go func() {
		for _, p := range peers {
			pch <- p
		}
	}()

	defer close(pch)

	for i := 0; i < 3; i++ {
		go func() {
			for p := range pch {
				logging.Entry().WithField("peer", p.String()).Debug("asking peer for tip")

				t, err := h.ReqTip(ctx, p)
				if err != nil {
					logging.WithError(err).WithField("peer", p.String()).Error("requesting tip")
					continue
				}

				mu.Lock()
				tc, ok := tips[t]
				if !ok {
					tc = 0
				}
				tc++
				tips[t] = tc
				mu.Unlock()

				wg.Done()
			}
		}()
	}

	done := make(chan struct{})

	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return nil, errors.New("ctx timeout")
	case <-done:
	}

	return tips, nil
}

func (h *Handler) ReqTip(ctx context.Context, p peer.ID) (string, error) {
	s, err := h.n.P2P().Open(ctx, p, ProtocolID)
	if err != nil {
		return "", errors.Wrap(err, "connecting to peer")
	}
	defer s.Close()

	stream := stream.NewRW(s)

	req := &api.ConsensusRequest{Request: &api.ConsensusRequest_Tip{
		Tip: &api.TipRequest{
			ChainId: h.consensus.ChainID(),
		},
	}}

	b, err := proto.Marshal(req)
	if err != nil {
		return "", errors.Wrap(err, "marshalling req")
	}

	if err := stream.Write(b); err != nil {
		return "", errors.Wrap(err, "sending req")
	}

	br, err := stream.Read()
	if err != nil {
		return "", errors.Wrap(err, "reading resp")
	}

	resp := &api.TipResponse{}
	if err := proto.Unmarshal(br, resp); err != nil {
		return "", errors.Wrap(err, "unmarshalling resp")
	}

	if resp.ChainId != h.consensus.ChainID() {
		return "", errors.New("resp contained mismatch chain id")
	}

	return resp.Tip, nil
}

func (h *Handler) Handle(nstream network.Stream) {
	defer nstream.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s := stream.NewRW(nstream)
	msg, err := s.Read()
	if err != nil {
		logging.WithError(err).Error("reading did stream msg req")
		return
	}

	req := &api.ConsensusRequest{}
	if err := proto.Unmarshal(msg, req); err != nil {
		logging.WithError(err).Error("decoding did stream msg req")
		return
	}

	switch t := req.Request.(type) {
	case *api.ConsensusRequest_Tip:
		h.handleTipReq(ctx, s, req.GetTip())
	default:
		logging.Entry().Errorf("unsupported request type: %T", t)
		return
	}
}

func (h *Handler) handleTipReq(ctx context.Context, s *stream.RW, req *api.TipRequest) {
	if req.ChainId != h.consensus.ChainID() {
		return
	}

	state := h.consensus.State()

	resp := &api.TipResponse{
		Tip:     state.Block.String(),
		ChainId: h.consensus.ChainID(),
	}

	b, err := proto.Marshal(resp)
	if err != nil {
		logging.WithError(err).Error("marshalling tip resp")
		return
	}

	if err := s.Write(b); err != nil {
		logging.WithError(err).Error("writing tip resp")
	}
}
