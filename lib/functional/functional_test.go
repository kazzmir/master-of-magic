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
}
