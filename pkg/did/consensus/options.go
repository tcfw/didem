package consensus

import (
	"github.com/tcfw/didem/pkg/storage"
	"go.dedis.ch/kyber/v3"
)

type Option func(*Consensus) error

func WithSigningKey(priv kyber.Scalar) Option {
	return func(c *Consensus) error {
		c.signingKey = priv
		return nil
	}
}

func WithBlockStore(s storage.Store) Option {
	return func(c *Consensus) error {
		c.blockStore = s
		return nil
	}
}

func WithDB(db Db) Option {
	return func(c *Consensus) error {
		c.db = db
		return nil
	}
}

func WithBeaconSource(s <-chan int64) Option {
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
