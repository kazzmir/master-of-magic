package functional

import (
    "testing"
)

func TestMemoize(test *testing.T) {
    count := 0
    f := func(x int) int {
        count += 1
        return x * x
    }

    memoized := Memoize(f)
    if memoized(3) != 9 {
        test.Errorf("Expected 9")
    }
    if memoized(4) != 16 {
        test.Errorf("Expected 16")
    }
    if memoized(3) != 9 {
        test.Errorf("Expected 9")
    }

    if count != 2 {
        test.Errorf("Expected 2 calls to f, got %d", count)
    }

    count = 0
    f2 := func(x, y int) int {
        count += 1
        return x * y
    }

    memoized2 := Memoize2(f2)

    if memoized2(3, 4) != 12 {
        test.Errorf("Expected 12")
    }

    if memoized2(5, 6) != 30 {
        test.Errorf("Expected 30")
    }

    if memoized2(3, 4) != 12 {
        test.Errorf("Expected 12")
    }

    if count != 2 {
        test.Errorf("Expected 2 calls to f2, got %d", count)
    }

    f0_count := 0
    f0 := func() int {
        f0_count += 1
        return 8
    }

    f0_v1 := f0()
    f0_v2 := f0()
    if f0_v1 != 8 || f0_v2 != 8 {
        test.Errorf("Expected 8")
    }

    if f0_count != 2 {
        test.Errorf("Expected 2 calls to f0, got %d", f0_count)
    }
}

func TestCurry(test *testing.T) {
    f := func(x, y int) int {
        return x + y
    }

    curried := Curry2(f)
    f1 := curried(3)
    if f1(4) != 7 {
        test.Errorf("Expected 7")
    }

    f2 := curried(5)
    if f2(6) != 11 {
        test.Errorf("Expected 11")
    }
}
