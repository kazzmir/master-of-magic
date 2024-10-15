package timedcache

import (
    "time"
    "sync"
)

type cacheValue[T any] struct {
    Value T
    ExpirationTime time.Time
    Lock sync.Mutex
}

func (value *cacheValue[T]) IsExpired() bool {
    value.Lock.Lock()
    defer value.Lock.Unlock()
    return time.Now().After(value.ExpirationTime)
}

type Cache[K comparable, T any] struct {
    Duration time.Duration
    values map[K]*cacheValue[T]
    lock sync.Mutex
}

func MakeCache[K comparable, T any](duration time.Duration) *Cache[K, T] {
    return &Cache[K, T]{
        Duration: duration,
        values: make(map[K]*cacheValue[T]),
    }
}

func (cache *Cache[K, T]) Size() int {
    cache.lock.Lock()
    defer cache.lock.Unlock()
    return len(cache.values)
}

func (cache *Cache[K, T]) Add(key K, value T) {
    cache.lock.Lock()
    defer cache.lock.Unlock()

    cacheValue := &cacheValue[T]{
        Value: value,
        ExpirationTime: time.Now().Add(cache.Duration),
    }

    cache.values[key] = cacheValue
    go func(){
        quit := false
        // keep trying to remove the item until we are successful
        for !quit {
            <-time.After(cache.Duration)
            cache.lock.Lock()
            newValue, ok := cache.values[key]

            // the value wasn't in the cache or was overwritten, just quit
            if !ok || newValue != cacheValue {
                quit = true
            } else if ok && newValue == cacheValue && newValue.IsExpired() {
                delete(cache.values, key)
                quit = true
            }
            cache.lock.Unlock()
        }
    }()
}

func (cache *Cache[K, T]) Get(key K) (T, bool) {
    cache.lock.Lock()
    defer cache.lock.Unlock()

    value, ok := cache.values[key]
    if !ok {
        var x T
        return x, false
    }

    value.Lock.Lock()
    value.ExpirationTime = time.Now().Add(cache.Duration)
    value.Lock.Unlock()

    return value.Value, true
}

func (cache *Cache[K, T]) Remove(key K) {
    cache.lock.Lock()
    defer cache.lock.Unlock()

    delete(cache.values, key)
}

func (cache *Cache[K, T]) Contains(key K) bool {
    cache.lock.Lock()
    defer cache.lock.Unlock()

    value, ok := cache.values[key]
    if !ok {
        return false
    }

    value.Lock.Lock()
    defer value.Lock.Unlock()
    return !value.IsExpired()
}
