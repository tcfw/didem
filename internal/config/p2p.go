package config

import (
	"github.com/spf13/viper"
)

type P2P struct {
	Connections struct {
		PeersCountHigh int
		PeersCountLow  int
	}
	BootstrapPeers []string
	ListenAddrs    []string
	Relay          bool
}

const (
	Cfg_p2p_connections_peerCountLow  = "p2p.connections.peerCountLow"
	Cfg_p2p_connections_peerCountHigh = "p2p.connections.peerCountHigh"
	Cfg_p2p_bootstartPeers            = "p2p.bootstartPeers"
	Cfg_p2p_listeningAddrs            = "p2p.listeningAddrs"
	Cfg_p2p_enableRelay               = "p2p.enableRelay"
)

var (
	p2pDefaults = map[string]interface{}{
		Cfg_p2p_connections_peerCountLow:  162,
		Cfg_p2p_connections_peerCountHigh: 192,
		Cfg_p2p_bootstartPeers:            []string{},
		Cfg_p2p_listeningAddrs: []string{
			"/ip4/0.0.0.0/udp/8712/quic",
			"/ip6/::0/udp/8712/quic",
		},
		Cfg_p2p_enableRelay: false,
	}
)

func init() {
	for k, v := range p2pDefaults {
		viper.SetDefault(k, v)
	}
}

func buildP2PConfig() (*P2P, error) {
	c := &P2P{}

	c.Connections.PeersCountLow = viper.GetInt(Cfg_p2p_connections_peerCountHigh)
	c.Connections.PeersCountHigh = viper.GetInt(Cfg_p2p_connections_peerCountHigh)
	c.BootstrapPeers = viper.GetStringSlice(Cfg_p2p_bootstartPeers)
	c.ListenAddrs = viper.GetStringSlice(Cfg_p2p_listeningAddrs)

	return c, nil
}
