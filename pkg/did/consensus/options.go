package consensus

import (
	"github.com/tcfw/didem/pkg/cryptography"
	"github.com/tcfw/didem/pkg/storage"
)

type Option func(*Consensus) error

func WithSigningKey(priv *cryptography.Bls12381PrivateKey) Option {
	return func(c *Consensus) error {
		c.signingKey = priv
		return nil
	}
}

func WithBlockStore(s storage.Store) Option {
	return func(c *Consensus) error {
		c.store = s
		return nil
	}
}

func WithBeaconSource(s <-chan uint64) Option {
	return func(c *Consensus) error {
		c.beacon = s
		return nil
	}
}

func WithValidator(v storage.Validator) Option {
	return func(c *Consensus) error {
		c.validator = v
		return nil
	}
}

func WithGenesis(g *storage.GenesisInfo) Option {
	return func(c *Consensus) error {
		c.chain = []byte(g.ChainID)
		c.genesis = g

		return nil
	}
}
