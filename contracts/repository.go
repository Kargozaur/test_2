package contracts

import (
	"context"
	"t/domain"
)

type Tx interface {
	Commit() error
	Rollback() error
}

type Repository interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error
	Retrieve(ctx context.Context, id string) (*domain.Order, error)
	RetrieveForUpdate(ctx context.Context, tx Tx, id string) (*domain.Order, error)
	Update(ctx context.Context, tx Tx, order *domain.Order) error
	CreateAuditLog(ctx context.Context, tx Tx, orderID, action, detail string) error
}
