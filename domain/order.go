package domain

import (
	"errors"
	"time"
)

type Status string

const (
	CONFIRMED Status = "CONFIRMED"
	COMPLETED Status = "COMPLETED"
)

var ErrNotConfirmed = errors.New("can not complete an order that is not confirmed")

type Order struct {
	ID         string
	CustomerID string
	Status     Status
	Total      int64
	CreatedAt  time.Time
}

func (o *Order) Complete() error {
	if o.Status != CONFIRMED {
		return ErrNotConfirmed
	}
	o.Status = COMPLETED
	return nil
}
