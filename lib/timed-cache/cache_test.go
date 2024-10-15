package timedcache

import (
    "testing"
    "time"
)

func TestCache1(test *testing.T){
    cache := MakeCache[string, int](time.Millisecond * 50)

    cache.Add("x", 1)
    value, ok := cache.Get("x")

    if !ok || value != 1 {
        test.Errorf("Expected value to be present")
    }

    time.Sleep(time.Millisecond * 100)

    value, ok = cache.Get("x")
    if ok {
        test.Errorf("Expected value to be expired")
    }

    cache.Add("y", 2)
    for i := 0; i < 3; i++ {
        time.Sleep(time.Millisecond * 40)
        // reset the deadline
        y, ok := cache.Get("y")
        if !ok || y != 2 {
            test.Errorf("Expected value to be present")
        }
    }

    time.Sleep(time.Millisecond * 100)
    _, ok = cache.Get("y")
    if ok {
        test.Errorf("Expected value to be expired")
    }
}
