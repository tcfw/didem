package consensus

import (
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/tcfw/didem/pkg/storage"
)

type Option func(*Consensus) error

func WithIdentity(priv crypto.PrivKey) Option {
	return func(c *Consensus) error {
		c.priv = priv
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