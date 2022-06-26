package api

import (
	"context"
	"net"

	"github.com/tcfw/didem/internal/node"
	"google.golang.org/grpc"
)

type APIHandler interface {
	Setup(*Api) error
	Desc() *grpc.ServiceDesc
}

var (
	reg = []APIHandler{}
)

type BaseHandler struct {
	a *Api
}

func (b *BaseHandler) Setup(a *Api) error {
	b.a = a
	return nil
}

type Api struct {
	n *node.Node
	g *grpc.Server
}

func NewAPI(n *node.Node) (*Api, error) {
	g := grpc.NewServer()

	for _, s := range reg {
		g.RegisterService(s.Desc(), s)
	}

	a := &Api{
		n: n,
		g: g,
	}

	return a, nil
}

func (a *Api) ListenAndServe(l net.Addr) error {
	lis, err := net.Listen("tcp", l.String())
	if err != nil {
		return err
	}

	return a.g.Serve(lis)
}

func (a *Api) Shutdown(ctx context.Context) error {
	a.g.GracefulStop()
	return nil
}
