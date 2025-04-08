package functional

// higher order function that takes a function 'f' and returns a new function that caches the results of 'f'
func Memoize[Key comparable, Value any](f func(Key) Value) func(Key) Value {
    cache := make(map[Key]Value)
    return func(key Key) Value {
        if value, ok := cache[key]; ok {
            return value
        }

        result := f(key)
        cache[key] = result
        return result
    }
}

// memoize but with two key arguments (that are merged into a single key type)
func Memoize2[Key1 comparable, Key2 comparable, Value any](f func(Key1, Key2) Value) func(Key1, Key2) Value {
    type Key struct {
        k1 Key1
        k2 Key2
    }

    cache := make(map[Key]Value)
    return func(key1 Key1, key2 Key2) Value {
        key := Key{k1: key1, k2: key2}
        if value, ok := cache[key]; ok {
            return value
        }

        result := f(key1, key2)
        cache[key] = result
        return result
    }
}

func Memoize3[Key1 comparable, Key2 comparable, Key3 comparable, Value any](f func(Key1, Key2, Key3) Value) func(Key1, Key2, Key3) Value {
    type Key struct {
        k1 Key1
        k2 Key2
        k3 Key3
    }

    cache := make(map[Key]Value)
    return func(key1 Key1, key2 Key2, key3 Key3) Value {
        key := Key{k1: key1, k2: key2, k3: key3}
        if value, ok := cache[key]; ok {
            return value
        }

        result := f(key1, key2, key3)
        cache[key] = result
        return result
    }
}
