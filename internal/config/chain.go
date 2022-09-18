package config

import (
	"github.com/multiformats/go-multibase"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/tcfw/didem/pkg/storage"
	"github.com/vmihailenco/msgpack/v5"
)

type Chain struct {
	Genesis storage.GenesisInfo
	Key     string
}

const (
	Cfg_chain_genesisInfo = "chain.genesis"
	Cfg_chain_key         = "chain.key"
)

func buildChainConfig() (*Chain, error) {
	c := &Chain{}

	gcfg := viper.GetString(Cfg_chain_genesisInfo)

	_, gcfg_raw, err := multibase.Decode(gcfg)
	if err != nil {
		return nil, errors.Wrap(err, "b64 decoding genesis config")
	}

	if err := msgpack.Unmarshal(gcfg_raw, &c.Genesis); err != nil {
		return nil, errors.Wrap(err, "unmarshaling genesis info")
	}

	c.Key = viper.GetString(Cfg_chain_key)

	return c, nil
}
