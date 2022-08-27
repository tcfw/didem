package did

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/jpillora/backoff"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/pkg/errors"
	"github.com/tcfw/didem/api"
	"github.com/tcfw/didem/internal/stream"
	"github.com/tcfw/didem/internal/utils/logging"
	"github.com/tcfw/didem/pkg/did/consensus"
	"github.com/tcfw/didem/pkg/node"
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
}

func NewHandler(n node.Node) *Handler {
	h := n.P2P().Host()
	p := n.P2P().PubSub()
	c, err := consensus.NewConsensus(h, p)
	if err != nil {
		logging.Entry().Panic(err)
	}

	return &Handler{n: n, consensus: c}
}

func (h *Handler) Start() error {
	var peers []peer.ID
	bo := &backoff.Backoff{}

	for {
		peers = h.n.P2P().Peers()
		if len(peers) < tipSetPeerCountStartThreshold {
			logging.Entry().Info("waiting for more peers before querying block chain tip")
			time.Sleep(bo.Duration())
			continue
		}

		break
	}

	rand.Shuffle(len(peers), func(i, j int) { peers[i], peers[j] = peers[j], peers[i] })

	vPeers := peers[:tipSetPeerCount]

	tips, err := h.askForTip(context.Background(), vPeers)
	if err != nil {
		return errors.Wrap(err, "requesting tips")
	}

	if len(tips) != tipSetPeerCount || len(tips) != 1 {
		logging.Entry().Warn("failed to get unanimous required tips, trying again in 10s")
		time.Sleep(10 * time.Second)
		return h.Start()
	}

	tip := ""
	for t := range tips {
		tip = t
	}

	logging.Entry().Info("Got tip as ", tip)

	if tip == h.consensus.State().Block.String() {
		logging.Entry().Info("Starting consensus as matching tip")
		return h.consensus.Start()
	}

	//TODO(tcfw): Play forward chain

	return nil
}

func (h *Handler) askForTip(ctx context.Context, peers []peer.ID) (map[string]int, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

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

	for i := 0; i < 3; i++ {
		go func() {
			defer wg.Done()

			for p := range pch {
				t, err := h.ReqTip(ctx, p)
				if err != nil {
					logging.Entry().WithError(err).Error("requesting tip")
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
			}
		}()
	}

	wg.Wait()
	close(pch)

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
		logging.Entry().WithError(err).Error("reading did stream msg req")
		return
	}

	req := &api.ConsensusRequest{}
	if err := proto.Unmarshal(msg, req); err != nil {
		logging.Entry().WithError(err).Error("decoding did stream msg req")
		return
	}

	switch t := req.Request.(type) {
	case *api.ConsensusRequest_Tip:
		h.handleTipReq(ctx, s, req.GetTip())
	default:
		logging.Error("unsupported request type: %T", t)
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
		logging.Entry().WithError(err).Error("marshalling tip resp")
		return
	}

	if err := s.Write(b); err != nil {
		logging.Entry().WithError(err).Error("writing tip resp")
	}
}
