package algorithm_test

import (
    "github.com/kazzmir/master-of-magic/lib/algorithm"
    "testing"
)

func TestRandomWeight(test *testing.T) {
    c := 0
    for range 100 {
        if algorithm.ChoseRandomWeightedElement([]string{"a", "b", "c"}, []int{1, 2, 80000}) == "c" {
            c += 1
        }
    }
    if c < 90 {
        test.Errorf("Expected to get 'c' at least 8 times, got %d", c)
    }

    a := 0
    b := 0
    for range 1000 {
        choice := algorithm.ChoseRandomWeightedElement([]string{"a", "b"}, []int{5, 5})
        if choice == "a" {
            a += 1
        } else if choice == "b" {
            b += 1
        }
    }

    if a < 400 || b < 400 || a > 600 || b > 600 {
        test.Errorf("Expected to get 'a' and 'b' about equal times, got a: %d, b: %d", a, b)
    }
}
