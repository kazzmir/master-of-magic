package priority

import (
    "testing"
    "math/rand"
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

func TestInsert(test *testing.T){
    heap := MakeEmptyHeap()

    N := 1000

    for i := 0; i < N; i++ {
        heap.Insert(rand.Intn(10000))
    }

    if heap.Size() != N {
        test.Errorf("Expected size %v, got %d", N, heap.Size())
    }

    last := -1
    for i := 0; i < N; i++ {
        top := heap.RemoveMin()
        if top < last {
            test.Errorf("Expected increasing order, got %d after %d", top, last)
        }
        last = top
    }

    if !heap.IsEmpty() {
        test.Errorf("Expected empty heap")
    }
}