package priority

import (
    "testing"
)

func TestBasic(test *testing.T){
    heap := MakeHeap([]int{3, 2, 1})

    if heap.Size() != 3 {
        test.Errorf("Expected size 3, got %d", heap.Size())
    }

    top := heap.RemoveMin()
    if top != 1 {
        test.Errorf("Expected min 1, got %d", top)
    }

    top = heap.RemoveMin()
    if top != 2 {
        test.Errorf("Expected min 2, got %d", top)
    }

    top = heap.RemoveMin()
    if top != 3 {
        test.Errorf("Expected min 3, got %d", top)
    }

    if !heap.IsEmpty() {
        test.Errorf("Expected empty heap")
    }
}
