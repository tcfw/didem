package mock

import (
	"context"

	"github.com/tcfw/didem/pkg/storage"
	"github.com/tcfw/didem/pkg/tx"
)

type MockValidator struct {
}

func (m *MockValidator) IsBlockValid(_ context.Context, _ *storage.Block, _ bool) error {
	return nil
}

func (m *MockValidator) IsTxValid(_ context.Context, _ *tx.Tx) error {
	return nil
}

func (m *MockValidator) ApplyFromTip(_ context.Context, _ storage.BlockID) error {
	return nil
}
