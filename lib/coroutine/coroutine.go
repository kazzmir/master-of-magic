package coroutine

import (
    "fmt"
)

// TODO: support return values of an arbitrary type T
// Allow main function to cancel the coroutine?

type Coroutine struct {
    yieldFrom chan struct{}
    yieldTo chan struct{}
    errorOut *error
}

type YieldFunc func() error

type AcceptYieldFunc func(yield YieldFunc) error

var CoroutineFinished = fmt.Errorf("coroutine finished")

func MakeCoroutine(user AcceptYieldFunc) *Coroutine {
    yieldTo := make(chan struct{})
    yieldFrom := make(chan struct{})

    coroutineError := CoroutineFinished

    go func(){
        defer func(){ close(yieldTo) }()
        <-yieldFrom
        err := user(func() error {
            yieldTo <- struct{}{}
            _, ok := <-yieldFrom

            if !ok {
                return fmt.Errorf("coroutine cancelled")
            }

            return nil
        })
        if err != nil {
            coroutineError = err
        }
    }()

    return &Coroutine{
        yieldFrom: yieldFrom,
        yieldTo: yieldTo,
        errorOut: &coroutineError,
    }
}

/* nil return means the coroutine is still running.
 * CoroutineFinished means the coroutine has finished.
 * any other non-nil error is an error from the user's function
 */
func (coro *Coroutine) Run() error {
    coro.yieldFrom <- struct{}{}
    _, ok := <-coro.yieldTo
    if !ok {
        close(coro.yieldFrom)
        return *coro.errorOut
    }
    return nil
}
