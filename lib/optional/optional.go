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
