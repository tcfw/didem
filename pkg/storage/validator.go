package storage

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tcfw/didem/internal/utils/logging"
	"github.com/tcfw/didem/pkg/tx"
)

type Validator interface {
	IsBlockValid(context.Context, *Block, bool) error
	IsTxValid(context.Context, *tx.Tx) error
	ApplyFromTip(context.Context, BlockID) error
}

type TxValidator struct {
	s Store
}

func NewTxValidator(s Store) *TxValidator {
	return &TxValidator{s}
}

func (v *TxValidator) ApplyFromTip(ctx context.Context, id BlockID) error {
	lastApplied, err := v.s.GetLastApplied(ctx)
	if err != nil {
		return errors.Wrap(err, "getting last applied block")
	}

	currentTip, err := v.s.GetBlock(ctx, id)
	if err != nil {
		return errors.Wrap(err, "getting tip block")
	}

	logging.Entry().Warn("Fetching new blocks")

	//Fetch all older blocks and queue up play forward
	queue := []BlockID{currentTip.ID}

	for currentTip.ID != lastApplied.ID {
		currentTip, err = v.s.GetBlock(ctx, currentTip.Parent)
		if err != nil {
			return errors.Wrap(err, "getting block")
		}
		queue = append(queue, currentTip.ID)
		logging.Entry().Infof("%d to process\r", len(queue))
	}

	return nil
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
	/*
		checks:
		->add
		- is signed by genesis key &&
		- node does not already exist &&
		- node has registered did

		->revoke
		- is signed by node || signed by genesis key

		->update: not supported
	*/
	return nil
}

func (v *TxValidator) isDIDTxValid(ctx context.Context, t *tx.Tx) error {
	//TODO(tcfw)
	/*
		checks:
		->add
		- did does not exist &&
		- tx is signed by did key &&
		- is valid did structure

		->update
		- did exists
		- tx signed by a key in did

		->revoke
		- did exists
		- tx signed by a key in did
	*/
	return nil
}
func (v *TxValidator) isVCTxValid(ctx context.Context, t *tx.Tx) error {
	//TODO(tcfw)
	/*
		checks:
		->add
		- has proof from controller &&
		- is valid vc structure

		->update
		->revoke
	*/
	return nil
}
