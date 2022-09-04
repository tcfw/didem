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

	logging.Entry().Info("Fetching new blocks")

	//Fetch all older blocks and queue up play forward
	queue := []*Block{currentTip}

	for currentTip.ID != lastApplied.ID {
		currentTip, err = v.s.GetBlock(ctx, currentTip.Parent)
		if err != nil {
			return errors.Wrap(err, "getting block")
		}
		queue = append(queue, currentTip)
		logging.Entry().Infof("%d to process\r", len(queue))
	}

	//invert queue
	for i, j := 0, len(queue)-1; i < j; i, j = i+1, j-1 {
		queue[i], queue[j] = queue[j], queue[i]
	}

	logging.Entry().Info("Applying new blocks")

	//validate and import each block
	for i, b := range queue {
		if err := v.IsBlockValid(ctx, b, true); err != nil {
			return errors.Wrap(err, "validating block")
		}

		if err := v.s.MarkBlock(ctx, b.ID, BlockStateAccepted); err != nil {
			return errors.Wrap(err, "indexing block")
		}
		logging.Entry().Infof("%d blocks processed\r", i)
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

	ogs := v.s
	v.s, err = v.s.StartTest(ctx)
	if err != nil {
		v.s = ogs
		return errors.Wrap(err, "failed to start test apply")
	}
	defer func() {
		v.s = ogs
		v.s.CompleteTest(ctx)
	}()

	for id, tx := range txs {
		if err := v.IsTxValid(ctx, tx); err != nil {
			return errors.Wrap(err, "invalid tx in block")
		}

		if isNewBlock {
			block, err := v.s.GetTxBlock(ctx, id)
			if err != nil && !errors.Is(err, ErrNotFound) {
				return errors.Wrap(err, "checking for preexisting tx")
			}
			if block != nil {
				return errors.New("tx already in previous block")
			}
		}

		if err := v.s.ApplyTx(ctx, id, tx); err != nil {
			return errors.Wrap(err, "applying test tx")
		}
	}

	if err := v.s.MarkBlock(ctx, b.ID, BlockStateValidated); err != nil {
		return errors.Wrap(err, "marking block validated")
	}

	return nil
}

func (v *TxValidator) IsTxValid(ctx context.Context, t *tx.Tx) error {
	if t.Version != tx.Version1 {
		return ErrTxVersionNotSupported
	}

	if t.Ts == 0 {
		return ErrTxMissingTimestamp
	}

	if t.From == "" {
		return ErrTxMissingFrom
	}

	if t.Data == nil {
		return ErrTxMissingData
	}

	switch t.Type {
	case tx.TxType_DID:
		return v.isDIDTxValid(ctx, t)
	case tx.TxType_VC:
		return v.isVCTxValid(ctx, t)
	case tx.TxType_Node:
		return v.isNodeTxValid(ctx, t)
	default:
		return ErrTxUnsupportedType
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

	switch t.Action {
	// case tx.TxActionAdd:
	// case tx.TxActionRevoke:
	default:
		return ErrOpNotSupported
	}

	return nil
}

func (v *TxValidator) isDIDTxValid(ctx context.Context, t *tx.Tx) error {
	did := t.Data.(*tx.DID)
	existingDid, lookupErr := v.s.LookupDID(ctx, did.Document.ID)

	//construct signature data
	signature := t.Signature
	t.Signature = nil
	sigmsg, err := t.Marshal()
	if err != nil {
		return errors.Wrap(err, "marshalling tx for signature")
	}

	if err := did.Document.IsValid(); err != nil {
		return ErrDIDInvalid
	}

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

	switch t.Action {
	case tx.TxActionAdd:
		if lookupErr != ErrNotFound {
			return ErrDIDAlreadyExists
		}

		if err := did.Document.Signed(signature, sigmsg); err != nil {
			return ErrDIDInvalidSignature
		}
	case tx.TxActionUpdate, tx.TxActionRevoke:
		if lookupErr != nil {
			return errors.Wrap(err, "tx validation")
		}
		if err := existingDid.Signed(signature, sigmsg); err != nil {
			return ErrDIDInvalidSignature
		}
	default:
		return ErrOpNotSupported
	}

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
