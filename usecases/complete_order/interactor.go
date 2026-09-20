package complete_order

import (
	"context"
	"errors"
	"fmt"
	"t/contracts"
)

var (
	ErrEmptyParams = errors.New("empty parameters were provided")
)

type Interactor struct {
	repo contracts.Repository
}

func New(repo contracts.Repository) *Interactor {
	return &Interactor{
		repo: repo,
	}
}

func (i *Interactor) Execute(ctx context.Context, orderID string) error {
	if orderID == "" {
		return ErrEmptyParams
	}
	err := i.repo.WithinTx(ctx, func(ctx context.Context, tx contracts.Tx) error {
		order, err := i.repo.RetrieveForUpdate(ctx, tx, orderID)
		if err != nil {
			return fmt.Errorf("retrieve order %s: %w", orderID, err)
		}
		if err := order.Complete(); err != nil {
			return fmt.Errorf("complete order %s: %w", orderID, err)
		}
		if err := i.repo.Update(ctx, tx, order); err != nil {
			return fmt.Errorf("persist order %s: %w", orderID, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("complete_order: %w", err)
	}
	return nil
}
