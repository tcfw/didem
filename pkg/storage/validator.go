package storage

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tcfw/didem/pkg/tx"
)

type Validator interface {
	IsBlockValid(context.Context, *Block, bool) error
	IsTxValid(context.Context, *tx.Tx) error
}

type TxValidator struct {
	s Store
}

func NewTxValidator(s Store) *TxValidator {
	return &TxValidator{s}
}

func (v *TxValidator) IsBlockValid(ctx context.Context, b *Block, isNewBlock bool) error {
	if err := v.s.MarkBlock(ctx, b.ID, BlockStateUnvalidated); err != nil {
		return errors.Wrap(err, "marking block unvalidated")
	}

	txs, err := v.s.AllTx(ctx, b)
	if err != nil {
		return errors.Wrap(err, "getting block txs")
	}

	for id, tx := range txs {
		if isNewBlock {
			block, err := v.s.GetTxBlock(ctx, id)
			if err != nil && !errors.Is(err, ErrNotFound) {
				return errors.Wrap(err, "checking for preexisting tx")
			}
			if block != nil {
				return errors.New("tx already in previous block")
			}
		}

		if err := v.IsTxValid(ctx, tx); err != nil {
			return errors.Wrap(err, "invalid tx in block")
		}
	}

	if err := v.s.MarkBlock(ctx, b.ID, BlockStateValidated); err != nil {
		return errors.Wrap(err, "marking block validated")
	}

	return nil
}

func (v *TxValidator) IsTxValid(ctx context.Context, t *tx.Tx) error {
	switch t.Type {
	case tx.TxType_DID:
		return v.isDIDTxValid(ctx, t)
	case tx.TxType_VC:
		return v.isVCTxValid(ctx, t)
	case tx.TxType_Node:
		return v.isNodeTxValid(ctx, t)
	default:
		return errors.New("unknown Tx type")
	}
}

func (v *TxValidator) isNodeTxValid(ctx context.Context, t *tx.Tx) error {
	//TODO(tcfw)
	return nil
}

func (v *TxValidator) isDIDTxValid(ctx context.Context, t *tx.Tx) error {
	//TODO(tcfw)
	return nil
}
func (v *TxValidator) isVCTxValid(ctx context.Context, t *tx.Tx) error {
	//TODO(tcfw)
	return nil
}
