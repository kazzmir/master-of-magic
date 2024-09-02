package coroutine

import (
    "fmt"
)

// TODO: support return values of an arbitrary type T
// Allow main function to cancel the coroutine?

type Coroutine struct {
    yieldFrom chan struct{}
    yieldTo chan struct{}
}

type YieldFunc func() error

type AcceptYieldFunc func(yield YieldFunc) error

var CoroutineFinished = fmt.Errorf("coroutine finished")

func NewCoroutine(user AcceptYieldFunc) *Coroutine {
    yieldTo := make(chan struct{})
    yieldFrom := make(chan struct{})

    go func(){
        defer func(){ close(yieldTo) }()
        <-yieldFrom
        user(func() error {
            yieldTo <- struct{}{}
            _, ok := <-yieldFrom

            if !ok {
                return fmt.Errorf("coroutine cancelled")
            }

            return nil
        })
    }()

    return &Coroutine{
        yieldFrom: yieldFrom,
        yieldTo: yieldTo,
    }
}

func (coro *Coroutine) Run() error {
    coro.yieldFrom <- struct{}{}
    _, ok := <-coro.yieldTo
    if !ok {
        close(coro.yieldFrom)
        return CoroutineFinished
    }
    return nil
}
