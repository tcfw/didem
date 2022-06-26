package api

import (
	apipb "github.com/tcfw/didem/api"
	"google.golang.org/grpc"
)

func init() {
	reg = append(reg, &emApi{})
}

type emApi struct {
	apipb.UnimplementedEmServiceServer
	BaseHandler
}

func (em *emApi) Desc() *grpc.ServiceDesc {
	return &apipb.EmService_ServiceDesc
}
