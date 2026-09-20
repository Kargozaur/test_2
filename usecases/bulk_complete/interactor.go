package bulk_complete

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"t/contracts"
)

type Semaphore struct {
	ch chan struct{}
}

func NewSemaphore(lim int) *Semaphore {
	return &Semaphore{
		ch: make(chan struct{}, lim),
	}
}

func (s *Semaphore) Acquire() {
	s.ch <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.ch
}

type Interactor struct {
	repo      contracts.Repository
	semaphore *Semaphore
}

func New(repo contracts.Repository, maxExecutions int) *Interactor {
	return &Interactor{
		repo:      repo,
		semaphore: NewSemaphore(maxExecutions),
	}
}

func (i *Interactor) Execute(ctx context.Context, orderIDs []string) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	ids := removeDuplicates(orderIDs)
	errs := make([]error, 0, len(ids))
	addErr := func(err error) {
		mu.Lock()
		defer mu.Unlock()
		errs = append(errs, err)
	}
	for _, id := range ids {
		if ctx.Err() != nil {
			errs = append(errs, ctx.Err())
			continue
		}
		wg.Add(1)
		i.semaphore.Acquire()
		go func(orderID string) {
			defer wg.Done()
			defer i.semaphore.Release()
			err := i.repo.WithinTx(ctx, func(ctx context.Context, tx contracts.Tx) error {
				order, err := i.repo.RetrieveForUpdate(ctx, tx, orderID)
				if err != nil {
					return fmt.Errorf("order id: %s; retrieve: %w.", orderID, err)
				}
				if err := order.Complete(); err != nil {
					return fmt.Errorf("order id: %s; complete: %w.", orderID, err)
				}
				return i.repo.Update(ctx, tx, order)
			})
			if err != nil {
				addErr(err)
			}
		}(id)
	}
	wg.Wait()
	if len(errs) != 0 {
		return errors.Join(errs...)
	}
	return nil
}

func removeDuplicates(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
