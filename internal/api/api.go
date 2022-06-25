package api

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/tcfw/didem/internal/node"
)

type APIHandler interface {
	Setup(*Api, *mux.Router) error
}

var (
	reg = []APIHandler{}
)

type Api struct {
	n *node.Node
	s *http.Server
}

func NewAPI(n *node.Node) *Api {
	return &Api{n: n}
}

func (a *Api) ListenAndServe(l net.Addr) error {
	a.s = &http.Server{
		Addr:         l.String(),
		Handler:      a.router(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  1 * time.Minute,
	}
	return a.s.ListenAndServe()
}

func (a *Api) Shutdown(ctx context.Context) error {
	return a.s.Shutdown(ctx)
}

func (a *Api) router() http.Handler {
	r := mux.NewRouter()

	for _, h := range reg {
		if err := h.Setup(a, r); err != nil {
			panic(err)
		}
	}

	return r
}
