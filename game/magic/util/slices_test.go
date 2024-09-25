package util

import (
    "testing"
    "slices"
)

func TestRotateSlice(test *testing.T){
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
