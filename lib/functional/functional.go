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

