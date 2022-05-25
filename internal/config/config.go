package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func GetConfig() (*Config, error) {
	var err error
	viper.SetConfigType("yaml")

	c := &Config{}

	c.p2p, err = buildP2PConfig()
	if err != nil {
		return nil, errors.Wrap(err, "p2p config")
	}

	return c, nil
}

type Config struct {
	p2p *P2P
}

func (c *Config) P2P() *P2P {
	return c.p2p
}
