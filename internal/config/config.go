package config

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	defaults = map[string]interface{}{
		"verbose": false,
	}
)

func init() {
	for k, v := range defaults {
		viper.SetDefault(k, v)
	}
}

func GetConfig() (*Config, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("didem")
	viper.AddConfigPath("/etc/didem/")
	viper.AddConfigPath("$HOME/.didem")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("DIDEM")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error
			logrus.New().Warnf("no config found")
		} else {
			return nil, errors.Wrap(err, "reading config file")
		}
	}

	c := &Config{}

	c.p2p, err = buildP2PConfig()
	if err != nil {
		return nil, errors.Wrap(err, "p2p config")
	}

	if viper.GetBool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.WithField("level", "debug").Debug("setting log level")
	}

	return c, nil
}

type Config struct {
	p2p *P2P
}

func (c *Config) P2P() *P2P {
	return c.p2p
}
