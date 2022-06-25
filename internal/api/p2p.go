package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func init() {
	reg = append(reg, &p2pApi{})
}

type p2pApi struct {
	a *Api
}

func (p *p2pApi) Setup(a *Api, r *mux.Router) error {
	p.a = a

	r.HandleFunc("/p2p/peers", p.peers).Methods("GET")

	return nil
}

func (p *p2pApi) peers(w http.ResponseWriter, _ *http.Request) {
	json.NewEncoder(w).Encode(p.a.n.P2P().Peers())
}
