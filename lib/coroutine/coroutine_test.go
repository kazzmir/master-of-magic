package coroutine

import (
    "testing"
    "slices"
    "fmt"
)

func TestCoroutine1(testing *testing.T) {
    var data []int

    v1 := func(yield YieldFunc) error {
        // fmt.Printf("v1 run 1\n")
        data = append(data, 1)
        // fmt.Printf("v1 yield 1\n")
        yield()
        // fmt.Printf("v1 run 2\n")
        data = append(data, 10, 11)
        return nil
    }

    coro := MakeCoroutine(v1)

    // fmt.Printf("coroutine main run 1\n")
    coro.Run()
    // fmt.Printf("coroutine main run 2\n")
    if !slices.Equal(data, []int{1}) {
        testing.Error("data should be [1]")
    }
    data = append(data, 5)
    coro.Run()
    // fmt.Printf("coroutine main run 3\n")
    if !slices.Equal(data, []int{1, 5, 10, 11}) {
        testing.Error("data should be [1, 5, 10, 11]")
    }
}

func TestCoroutine2(testing *testing.T){
    z := 0
    v1 := func(yield YieldFunc) error {
        for z < 10 {
            z += 1
            yield()
        }

        return nil
    }

    coro := MakeCoroutine(v1)
    for {
        err := coro.Run()
        if err != nil {
            break
        }
    }

    if z != 10 {
        testing.Error("z should be 10")
    }
}

func TestCoroutineError(test *testing.T) {
    myError := fmt.Errorf("my error")
    v1 := func(yield YieldFunc) error {
        yield()
        return myError
    }

    coro := MakeCoroutine(v1)
    var err error
    for range 10 {
        err = coro.Run()
        if err != nil {
            break
        }
    }

    if err != myError {
        test.Error("error should have been myError")
    }
}
