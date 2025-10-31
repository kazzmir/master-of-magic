package priority

import (
    "cmp"
    "testing"
)

func TestOrder(test *testing.T){
    type Object struct {
        Value int
        Cost int
        Time int
    }

    compare := func(a, b Object) int {
        if a.Cost < b.Cost {
            return -1
        }

        if a.Cost > b.Cost {
            return 1
        }

        // need this to ensure order of elements that have the same cost
        return cmp.Compare(a.Time, b.Time)
    }

    p := MakePriorityQueue[Object](compare)
    for v := range 100 {
        p.Insert(Object{Value: v, Cost: 1, Time: v})
    }

    last := -1
    // the order we extract things should be exactly equal to the order we inserted them
    for !p.IsEmpty() {
        obj := p.ExtractMin()
        if obj.Value <= last {
            test.Errorf("Expected value less than %d, got %d", last, obj.Value)
        }
        last = obj.Value
    }
}
