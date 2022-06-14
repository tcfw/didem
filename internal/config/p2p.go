package config

import (
	"os/user"
	"path/filepath"
	"strings"

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
	IdentityFile   string
}

const (
	Cfg_p2p_connections_peerCountLow  = "p2p.connections.peerCountLow"
	Cfg_p2p_connections_peerCountHigh = "p2p.connections.peerCountHigh"
	Cfg_p2p_bootstartPeers            = "p2p.bootstartPeers"
	Cfg_p2p_listeningAddrs            = "p2p.listeningAddrs"
	Cfg_p2p_enableRelay               = "p2p.enableRelay"
	Cfg_p2p_identityFile              = "p2p.identityFile"
)

var (
	p2pDefaults = map[string]interface{}{
		Cfg_p2p_connections_peerCountLow:  162,
		Cfg_p2p_connections_peerCountHigh: 192,
		Cfg_p2p_bootstartPeers: []string{
			"/dnsaddr/bootstrap.didem.tcfw.com.au/p2p/12D3KooWBaKMRyqvN4VKKggSW9up159ugtEYYSTry1wtZKEzyHcX",
		},
		Cfg_p2p_listeningAddrs: []string{
			"/ip4/0.0.0.0/udp/8712/quic",
			"/ip6/::0/udp/8712/quic",
		},
		Cfg_p2p_enableRelay:  false,
		Cfg_p2p_identityFile: "~/.didem/p2p_identity",
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
	c.Relay = viper.GetBool(Cfg_p2p_enableRelay)
	c.IdentityFile = expandPath(viper.GetString(Cfg_p2p_identityFile))

	return c, nil
}

func expandPath(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if path == "~" {
		// In case of "~", which won't be caught by the "else if"
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		path = filepath.Join(dir, path[2:])
	}

	return path
}
