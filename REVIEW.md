## Issue 1: Domain entity depends on infrastructure

- File: domain/order.go
- Line: `func (o *Order) Complete(ctx context.Context, db *sql.DB) error`
- Severity: CRITICAL
- Problem: domain layer must have zero knowledge about db driver / queries / transaction.
- Impact: logic can't be unit tested withour real or mocked db.
- Fix: make Complete() method a pure function call, without passing any parameters.

## Issue 2: Domain entity performs logging

- File: domain/order.go
- Line: `log.Printf("Order %s completed", o.ID)`
- Severity: WARNING
- Problem: logging is infrastructure concern. Putting it in the domain method means every call to Complete() (including from tests) writes to the global logger and the domain package now depends on `log`.
- Impact: unwanted logs in tests, no ability to swap loggers or add structured fields(requestID / orderID / etc.).
- Fix: move the log line into the interactor.

## Issue 3: Usecases depends on the concrete repo struct and not the interface

- File: usecases/*/interactor.go
- Line: `repo *repo.OrderRepo` in all three Interactor structs
- Severity: WARNING
- Problem: violation of D of SOLID principles. Also, inner layers(like usecases) should not depend on outer layers(repo)
- Impact: usecases can't be unit-tested without a real database.
- Fix: Define interface(s) on which usecases will depend on.

## Issue 4: Repository exposes `*sql.DB` to callers

- File: usecases/bulk_complete/interactor.go
- Line: `order.Complete(ctx, uc.repo.DB())`
- Severity: CRITICAL
- Problem: repo.DB() exposes raw *sql.DB connection pool out of the repository, letting any usecase to run arbitrary SQL. There is no point of repository abstraction layer at all.
- Impact: Nothing stosp a future usecases from writing SQL queries, rather than using predefined one.
- Fix: Either remove DB() method or make it private (db()).

## Issue 5: Concurent writes to lastErr

- File: usecases/bulk_complete/interactor.go
- Line: `lastErr = err`
- Severity: CRITICAL
- Problem: lastErr is a single error variable in the enclosing scope, written concurently by every goroutine without synchronization.
- Impact: Either the program won't run if launched with `-race` flag or it will be runtime panic() (that should be avoided for most of the time).
- Fix: Either use error slice to store the errors and join them in return statement or use mutexes. Also, it is better to use concurent pattern like worker pool or semaphore.

## Issue 6: No context-cancellation handling

- File: usecases/bulk_complete/interactor.go
- Line: -
- Severity: WARNING
- Problem: The loop spawns one goroutine per orderID, but never checks ctx.Done().
- Impact: A large batch can exhaust DB connection pool. A cancelled request keeps consuming resources on request that was gived up.
- Fix: Use semaphore pattern with checking ctx.Err() != nil

## Issue 7: transfer_order performs 2 separate writes without transaction

- File: usecases/transfer_order/interactor.go
- Line: `uc.repo.Update(ctx, order)` followed by `uc.repo.CreateAuditLog(ctx, orderID, "TRANSFERRED", newCustomerID)`
- Severity: CRITICAL
- Problem: Update and CreateAuditLog are two separated round trips without atomicity.
- Impact: In failure scenario, if Update succeeds and CreateAuditLog not, it creates inconsistency in db schema.
- Fix: Create Tx manager or manage commits / rollbacks within the function call(but for clear architecture it is better to have Tx manager). Also, it is better to add `select ... for update` query.

## Issue 8: Money represented as float

- File: domain/order.go
- Line: `Total float64`
- Severity: CRITICAL
- Problem: float can't represent most decimal currency values exactly. Any arithmetic operation on Total will accumulate rounding error.
- Impact: float types are bad for precise calculations and can lead to potential financial / audit discrepancies at scale.
- Fix: Use decimal package or int64 to store and calculate Total.

## Issue 9: Missing import

- File: domain/order.go
- Line: `return errors.New("cannot complete")`
- Severity: CRITICAL
- Problem: Missing errors package import, which will lead to failed CI jobs and compilation errors.
- Impact: Won't compile
- Fix: Import errors package.

## Issue 10: Untyped / unwrapped error for an expected condition

- File: domain/order.go
- Line: `errors.New("cannot complete")`
- Severity: WARNING
- Problem: It's becoming very hard for handlers to map errors.
- Impact: Callers are forced to string-match error messages to branch on this condition.
- Fix: Define a package-level sentinel error, like `var ErrOrderNotCreated = errors.New("...")`, so callers can compare `errors.Is(err, domain.ErrOrderNotCreated)`

## Issue 11: No input validation

- File: usecases/*/interactor.go
- Line: Any line with method definition
- Severity: WARNING
- Problem: For orderID strings it will be less of a problem, because db (most likely) won't find order, but newCustomerID won't be represented as empty string.
- Impact: Empty newCustomerID will asign order to no one.
- Fix: validate the given strings according to some policies (like non-empty string, minimal and maximal length etc.).

## Issue 12: Status is bare string

- File: domain/order.go
- Severity: WARNING
- Problem: Status string allows any string to be attached, including typos, with no compile-time or runtime check
- Impact: Invalid statuses can leak into the db with nothing catching until the query with `where status = "?"` will start to miss rows.
- Fix: Define `type Status string` with const values(enum like) `StatusPending="PENDING"` etc. 