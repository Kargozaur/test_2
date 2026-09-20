package transfer_order

import (
	"context"
	"fmt"
	"t/contracts"
)

const (
	TRANSFERRED = "TRANSFERRED"
)

type Interactor struct {
	repo contracts.Repository
}

func New(repo contracts.Repository) *Interactor {
	return &Interactor{
		repo: repo,
	}
}

func (i *Interactor) Execute(ctx context.Context, orderID, newCustomerID string) error {
	if orderID == "" {
		return fmt.Errorf("transfer_order: orderID is required")
	}
	if newCustomerID == "" {
		return fmt.Errorf("transfer_order: newCustomerID is required")
	}
	err := i.repo.WithinTx(ctx, func(ctx context.Context, tx contracts.Tx) error {
		order, err := i.repo.RetrieveForUpdate(ctx, tx, orderID)
		if err != nil {
			return fmt.Errorf("retrieve order: %s, %w", orderID, err)
		}
		prevCustomerID := order.CustomerID
		order.CustomerID = newCustomerID
		detail := fmt.Sprintf("from=%s to=%s", prevCustomerID, order.CustomerID)
		if err := i.repo.Update(ctx, tx, order); err != nil {
			return fmt.Errorf("update order: %s, %w", orderID, err)
		}
		if err := i.repo.CreateAuditLog(ctx, tx, orderID, TRANSFERRED, detail); err != nil {
			return fmt.Errorf("audit log for order: %s, %w", orderID, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("transfer_order: %w", err)
	}
	return nil
}
