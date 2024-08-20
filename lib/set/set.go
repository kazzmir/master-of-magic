package set

type Set[T comparable] struct {
    data map[T]bool
}

func MakeSet[T comparable]() *Set[T] {
    return &Set[T]{
        data: make(map[T]bool),
    }
}

func (set *Set[T]) Insert(v T){
    set.data[v] = true
}

func (set *Set[T]) Contains(v T) bool {
    _, ok := set.data[v]
    return ok
}

func (set *Set[T]) Size() int {
    return len(set.data)
}

func (set *Set[T]) Remove(v T) {
    delete(set.data, v)
}
