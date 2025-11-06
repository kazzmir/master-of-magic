package util

import (
    "testing"
    "slices"
)

func TestRotateSlice(test *testing.T) {
    v := []int{1, 2, 3, 4}
    RotateSlice(v, true)

    if !slices.Equal(v, []int{2, 3, 4, 1}) {
        test.Fatalf("invalid rotation")
    }

    RotateSlice(v, false)

    if !slices.Equal(v, []int{1, 2, 3, 4}) {
        test.Fatalf("invalid backward rotation")
    }
}

func TestFirst(test *testing.T) {
    empty := []int{}
    if got := First(empty, 42); got != 42 {
        test.Fatalf("expected default for empty slice, got %v", got)
    }

    nonEmpty := []int{7, 8, 9}
    if got := First(nonEmpty, 42); got != 7 {
        test.Fatalf("expected first element for non-empty slice, got %v", got)
    }
}
