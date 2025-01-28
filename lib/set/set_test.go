package set

import (
    "testing"
)

func TestSet(test *testing.T){
    s := MakeSet[int]()

    s.Insert(2)
    s.Insert(8)

    if !s.Contains(2) {
        test.Errorf("Set should contain 2")
    }

    if s.Contains(5) {
        test.Errorf("Set should not contain 5")
    }

    if s.Size() != 2 {
        test.Errorf("Set should have size 2")
    }

    s.Remove(2)

    if s.Contains(2) {
        test.Errorf("Set should not contain 2")
    }

    if s.Size() != 1 {
        test.Errorf("Set should have size 1")
    }

    s = NewSet(1, 2, 3)

    if s.Size() != 3 {
        test.Errorf("Set should have size 3")
    }

    s.RemoveMany(1, 3)

    if s.Size() != 1 {
        test.Errorf("Set should have size 1")
    }

    if s.Contains(1) || s.Contains(3) {
        test.Errorf("Set should not contain removed element")
    }

    if !s.Contains(2) {
        test.Errorf("Set should still contain 2")
    }
}
