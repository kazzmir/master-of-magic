package memo

type Memo[T any, Arg any] struct {
    saved T
    valid bool
    f func(...Arg) T
}

func MakeMemo[T any, Arg any](f func(...Arg) T) *Memo[T, Arg] {
    return &Memo[T, Arg]{
        f: f,
        valid: false,
    }
}

func (memo *Memo[T, Arg]) Get(args... Arg) T {
    if !memo.valid {
        memo.saved = memo.f(args...)
        memo.valid = true
    }
    return memo.saved
}

func (memo *Memo[T, Arg]) Invalidate() {
    memo.valid = false
}
