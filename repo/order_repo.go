package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"t/contracts"
	"t/domain"
)

type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) *Repo {
	return &Repo{
		db: db,
	}
}

type txManager struct {
	tx *sql.Tx
}

func (t *txManager) Commit() error {
	return t.tx.Commit()
}

func (t *txManager) Rollback() error {
	return t.Rollback()
}

func (r *Repo) WithinTx(ctx context.Context, fn func(ctx context.Context, tx contracts.Tx) error) error {
	t, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	wrapped := &txManager{tx: t}
	defer wrapped.Rollback()
	if err := fn(ctx, wrapped); err != nil {
		return fmt.Errorf("fn tx: failed to execute")
	}
	if err := wrapped.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (r *Repo) Retrieve(ctx context.Context, id string) (*domain.Order, error) {
	row := r.db.QueryRowContext(ctx,
		`select id, customer_id, status, total, created_at from orders
	where id = ?`, id)
	return r.scanOrder(row)
}

func (r *Repo) RetrieveForUpdate(ctx context.Context, tx contracts.Tx, id string) (*domain.Order, error) {
	t, err := unwrapTx(tx)
	if err != nil {
		return nil, err
	}
	row := t.QueryRowContext(ctx, `
	select id, customer_id, status, total, created_at from orders
	where id = ? for update`, id)
	return r.scanOrder(row)
}

func (r *Repo) Update(ctx context.Context, tx contracts.Tx, order *domain.Order) error {
	t, err := unwrapTx(tx)
	if err != nil {
		return err
	}
	_, err = t.ExecContext(ctx,
		`update orders set customer_id = ?, status = ?, total_cents = ? where id = ?`,
		order.CustomerID, order.Status, order.Total, order.ID)
	return err
}

func (r *Repo) CreateAuditLog(ctx context.Context, tx contracts.Tx, orderID, action, detail string) error {
	t, err := unwrapTx(tx)
	if err != nil {
		return err
	}
	_, err = t.ExecContext(ctx,
		`insert into order_audit_log (order_id, action, detail, created_at) VALUES (?, ?, ?, now())`,
		orderID, action, detail)
	return err
}

func (r *Repo) scanOrder(row *sql.Row) (*domain.Order, error) {
	var o domain.Order
	var status string
	if err := row.Scan(&o.ID, &o.CustomerID, &status, &o.Total, &o.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("order %s not found: %w", o.ID, err)
		}
		return nil, err
	}
	o.Status = domain.Status(status)
	return &o, nil
}

func unwrapTx(tx contracts.Tx) (*sql.Tx, error) {
	t, ok := tx.(*txManager)
	if !ok {
		return nil, fmt.Errorf("unexpected tx implementation")
	}
	return t.tx, nil
}
