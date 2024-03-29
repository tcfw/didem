package consensus

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/pkg/errors"
	"github.com/tcfw/didem/pkg/storage"
	"github.com/tcfw/didem/pkg/tx"
)

// makeBlock attempts to create a new block from the mempool assuming the current node
// is the current proposer to be sent to other nodes
func (c *Consensus) makeBlock() (*storage.Block, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	block := &storage.Block{
		Version:   storage.Version,
		Parent:    storage.BlockID(c.state.Block),
		Height:    c.propsalState.Height,
		CreatedAt: time.Now().Unix(),
		Proposer:  c.id.String(),
	}

	_, err := rand.Read(block.Nonce[:])
	if err != nil {
		return nil, err
	}

	n := storage.MaxBlockTxCount
	if c.memPool.Len() < n {
		n = c.memPool.Len()
	}

	if n == 0 {
		return nil, nil //nothing to propose
	}

	txs := make([]*tx.Tx, 0, n)
	for n > 0 {
		tx := c.memPool.GetTx()
		if tx == nil {
			break
		}
		txs = append(txs, tx)
		n--
	}

	//TODO(tcfw): confirm if the txs are valid if applied sequently in the same block

	txCids := make([]cid.Cid, 0, len(txs))

	for _, tx := range txs {
		c, err := c.store.PutTx(ctx, tx)
		if err != nil {
			return nil, errors.Wrap(err, "storing tx")
		}
		txCids = append(txCids, c)
	}

	block.Bloom, err = storage.MakeBloom(txCids)
	if err != nil {
		return nil, errors.Wrap(err, "creating block bloom filter")
	}

	txSet, err := storage.NewTxSet(c.store, txCids)
	if err != nil {
		return nil, errors.Wrap(err, "creating tx set")
	}

	txsCid, err := c.store.PutSet(ctx, txSet)
	if err != nil {
		return nil, errors.Wrap(err, "storing tx set")
	}

	block.TxRoot = txsCid

	return block, nil
}
