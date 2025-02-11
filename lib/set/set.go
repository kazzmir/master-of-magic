package set

type Set[T comparable] struct {
    data map[T]bool
}

func MakeSet[T comparable]() *Set[T] {
    return &Set[T]{
        data: make(map[T]bool),
    }
}

func NewSet[T comparable](values ...T) *Set[T] {
    set := MakeSet[T]()
    for _, value := range values {
        set.Insert(value)
    }
    return set
}

func (set *Set[T]) Clone() *Set[T] {
    newSet := MakeSet[T]()
    for k := range set.data {
        newSet.data[k] = true
    }
    return newSet
}

func (set *Set[T]) Insert(v T){
    set.data[v] = true
}

func (set *Set[T]) Clear() {
    set.data = make(map[T]bool)
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

func (set *Set[T]) RemoveMany(values ...T) {
    for _, value := range values {
        delete(set.data, value)
    }
}

// FIXME: turn this into an iterator
func (set *Set[T]) Values() []T {
    if set == nil {
        return nil
    }

    var out []T
    for k := range set.data {
        out = append(out, k)
    }
    return out
}
