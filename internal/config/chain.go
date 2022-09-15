package config

import (
	"encoding/base64"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/tcfw/didem/pkg/storage"
	"github.com/vmihailenco/msgpack/v5"
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

	gcfg_raw, err := base64.StdEncoding.DecodeString(gcfg)
	if err != nil {
		return nil, errors.Wrap(err, "b64 decoding genesis config")
	}

	if err := msgpack.Unmarshal(gcfg_raw, &c.Genesis); err != nil {
		return nil, errors.Wrap(err, "unmarshaling genesis info")
	}

	return c, nil
}
