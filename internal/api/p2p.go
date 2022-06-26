package api

import (
	apipb "github.com/tcfw/didem/api"
	"google.golang.org/grpc"
)

func init() {
	reg = append(reg, &p2pApi{})
}

type p2pApi struct {
	apipb.UnimplementedEmServiceServer
	BaseHandler
}

func (p *p2pApi) Desc() *grpc.ServiceDesc {
	return &apipb.P2PService_ServiceDesc
}
