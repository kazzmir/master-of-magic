package optional

type Optional[T any] struct {
    Value T
    Present bool
}

func Of[T any](value T) Optional[T] {
    return Optional[T]{Value: value, Present: true}
}

func Empty[T any]() Optional[T] {
    return Optional[T]{Present: false}
}

// invoke some function with the value if present
func (o *Optional[T]) With(f func(v T)) {
    if o.Present {
        f(o.Value)
    }
}
