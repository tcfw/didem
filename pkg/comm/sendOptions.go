package comm

import "github.com/tcfw/didem/pkg/did"

type SendOption func(cfg *SendConfig) error

func WithIdentity(id did.PrivateIdentity) SendOption {
	return func(cfg *SendConfig) error {
		cfg.Id = id
		return nil
	}
}

type SendConfig struct {
	Id             did.PrivateIdentity
	PublicResolver did.Resolver
}

func NewDefaultSendConfig() *SendConfig {
	return &SendConfig{}
}
