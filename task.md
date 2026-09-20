# Backend Developer Assessment - Mid Level (Task 2)

## Instructions

1. Complete all tasks below
2. Push your solution to a **public GitHub repository**

---

## Task: Code Review and Concurrency Fix

You received a PR with the following code. Review it, identify all issues, and provide corrected implementations.

### Code to Review

**File 1: domain/order.go**
```go
package domain

import (
    "context"
    "database/sql"
    "log"
    "time"
)

type Order struct {
    ID         string
    CustomerID string
    Status     string
    Total      float64
    CreatedAt  time.Time
}

func (o *Order) Complete(ctx context.Context, db *sql.DB) error {
    if o.Status != "CONFIRMED" {
        return errors.New("cannot complete")
    }

    o.Status = "COMPLETED"
    log.Printf("Order %s completed", o.ID)

    _, err := db.ExecContext(ctx, "UPDATE orders SET status = ? WHERE id = ?", o.Status, o.ID)
    return err
}
```

**File 2: usecases/complete_order/interactor.go**
```go
package complete_order

import (
    "context"
    "fmt"
    "internal/app/order/repo"
)

type Interactor struct {
    repo *repo.OrderRepo
}

func (uc *Interactor) Execute(ctx context.Context, orderID string) error {
    order, err := uc.repo.Retrieve(ctx, orderID)
    if err != nil {
        return fmt.Errorf("failed: %w", err)
    }

    if err := order.Complete(ctx, uc.repo.DB()); err != nil {
        return fmt.Errorf("complete failed: %w", err)
    }

    return nil
}
```

**File 3: usecases/bulk_complete/interactor.go**
```go
package bulk_complete

import (
    "context"
    "sync"
)

type Interactor struct {
    repo *repo.OrderRepo
}

func (uc *Interactor) Execute(ctx context.Context, orderIDs []string) error {
    var wg sync.WaitGroup
    var lastErr error

    for _, id := range orderIDs {
        wg.Add(1)
        go func(orderID string) {
            defer wg.Done()

            order, err := uc.repo.Retrieve(ctx, orderID)
            if err != nil {
                lastErr = err
                return
            }

            if err := order.Complete(ctx, uc.repo.DB()); err != nil {
                lastErr = err
            }
        }(id)
    }

    wg.Wait()
    return lastErr
}
```

**File 4: usecases/transfer_order/interactor.go**
```go
package transfer_order

import "context"

type Interactor struct {
    repo *repo.OrderRepo
}

func (uc *Interactor) Execute(ctx context.Context, orderID, newCustomerID string) error {
    order, err := uc.repo.Retrieve(ctx, orderID)
    if err != nil {
        return err
    }

    order.CustomerID = newCustomerID

    if err := uc.repo.Update(ctx, order); err != nil {
        return err
    }

    if err := uc.repo.CreateAuditLog(ctx, orderID, "TRANSFERRED", newCustomerID); err != nil {
        return err
    }

    return nil
}
```

---

## Your Task

### Part 1: Create REVIEW.md

Find ALL issues in the code above. Categorize them:

**Architecture Violations:**
- List each violation of layer rules

**Concurrency Bugs:**
- There is a race condition in bulk_complete - identify it
- Explain what can go wrong in production

**Transaction Integrity Bugs:**
- There is a transaction bug in transfer_order - identify it
- Explain the failure scenario

**Other Issues:**
- Money handling, error handling, etc.

Format:
```
## Issue 1: [Title]
- File: ...
- Line: ...
- Severity: CRITICAL / WARNING
- Problem: ...
- Impact: ...
- Fix: ...
```

### Part 2: Provide Corrected Code

Fix all issues. Your corrected code must:
- Follow architecture rules (domain purity, repo returns mutations)
- Be thread-safe
- Handle transactions atomically
- Use proper error types

---

## Questions - Answer in ANSWERS.md

**Q1:** In `bulk_complete`, multiple goroutines write to `lastErr`. What SPECIFIC bug can occur? (Hint: it's not just "lost errors")

**Q2:** In `transfer_order`, if `Update` succeeds but `CreateAuditLog` fails:
- What is the state of the database?
- How would you fix this using the Plan pattern?

**Q3:** The domain method `Complete` accepts `context.Context` and `*sql.DB`. Why is this a violation? What's the correct signature?

**Q4:** This code ignores the error from Retrieve:
```go
source, _ := uc.repo.Retrieve(ctx, req.FromAccountID)
```
Besides "source could be nil", what ELSE could go wrong that this hides?

---

## Repository Structure

```
your-repo/
├── domain/
│   └── order.go           # Corrected
├── contracts/
│   └── repository.go
├── usecases/
│   ├── complete_order/
│   │   └── interactor.go  # Corrected
│   ├── bulk_complete/
│   │   └── interactor.go  # Corrected (thread-safe)
│   └── transfer_order/
│       └── interactor.go  # Corrected (atomic)
├── repo/
│   └── order_repo.go
├── REVIEW.md              # All issues found
└── ANSWERS.md             # Question answers
```

---

## Evaluation

Your submission will be evaluated against our engineering standards document. Key areas:
- Identified all architecture violations
- Identified race condition and explained impact
- Identified transaction integrity issue
- Corrected implementations follow all rules
- Thread-safe code where needed
- Proper use of mutex or atomic operations
