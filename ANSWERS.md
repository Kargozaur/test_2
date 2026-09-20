## Q1

Data race occures when accessing shared `lastErr` value from multiple goroutines. To fix it, we can use mutexes:
```go
mu.Lock()
lastErr = err
mu.Unlock()
```
But usually it is better to store all errors inside []errs slice and join them if len(errs) > 0

## Q2

We get inconsistent state. To fix the problem, we should wrap the whole implementation in transaction.
I'm not familiar with the Plan pattern, but I'm assuming that we would have a slice of Steps which will be running inside the WithinTx method in for loop and in case of failure, the whole thing will rollback.

## Q3

`Complete()` method must only to change state of the object itself and those object should not depend on the infrastructure. Additionally, in the initital implementation it breakes single responsibility principle

## Q4

The problem is that we're missing errors like timeouts, no rows found, context cancellation, sentinel errors -> we won't be able to understand why `source` will be nil.