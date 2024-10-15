package use

import (
    "time"
    "testing"

    "github.com/kazzmir/master-of-magic/lib/timed-cache"
)

func TestCache(test *testing.T){
    cache := timedcache.MakeCache[string, int](time.Second * 1)
    cache.Add("y", 3)
    v, ok := cache.Get("y")
    if !ok || v != 3 {
        test.Errorf("Expected 3, got %v", v)
    }
}
