package api

import (
	"context"

	apipb "github.com/tcfw/didem/api"
	"google.golang.org/grpc"
)

func init() {
	reg = append(reg, &p2pApi{})
}

type p2pApi struct {
	apipb.UnimplementedP2PServiceServer
	BaseHandler
}

func (p *p2pApi) Desc() *grpc.ServiceDesc {
	return &apipb.P2PService_ServiceDesc
}

func (p *p2pApi) Peers(ctx context.Context, _ *apipb.PeersRequest) (*apipb.PeersResponse, error) {
	peers := p.a.n.P2P().Peers()

	list := make([]string, 0, len(peers))
	for _, p := range peers {
		list = append(list, p.String())
	}

	return &apipb.PeersResponse{Peers: list}, nil
}
