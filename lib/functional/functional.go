package functional

func Memoize0[Value any](f func() Value) func() Value {
    var cached bool
    var cache Value
    return func() Value {
        if cached {
            return cache
        }

        cache = f()
        cached = true
        return cache
    }
}

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

// curry a 2-argument function to produce a 1-argument function
func Curry2[Arg1 any, Arg2 any, Result any](f func(Arg1, Arg2) Result) func(Arg1) func(Arg2) Result {
    return func(arg1 Arg1) func(Arg2) Result {
        return func(arg2 Arg2) Result {
            return f(arg1, arg2)
        }
    }
}

// curry a 1-argument function to produce a 0-argument function
func Curry1[Arg1 any, Result any](f func(Arg1) Result) func(Arg1) func () Result {
    return func(arg Arg1) func () Result {
        return func() Result {
            return f(arg)
        }
    }
}
