package config

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/tcfw/didem/pkg/storage"
)

type Chain struct {
	Genesis storage.GenesisInfo
}

const (
	Cfg_chain_genesisInfo = "chain.genesis"
)

func buildChainConfig() (*Chain, error) {
	c := &Chain{}

	gcfg := viper.GetString(Cfg_chain_genesisInfo)

	if err := json.Unmarshal([]byte(gcfg), &c.Genesis); err != nil {
		return nil, errors.Wrap(err, "unmarshaling genesis info")
	}

	return c, nil
}
