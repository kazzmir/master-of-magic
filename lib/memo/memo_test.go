package memo

import (
    "testing"
)

func TestMemo(test *testing.T) {
    count := 0
    memo1 := MakeMemo[int, int](func(x... int) int {
        count += 1
        return x[0] + 3
    })

    if memo1.Get(3) != 6 {
        test.Error("Memo returned wrong value")
    }

    if count != 1 {
        test.Error("Memo called function more than once")
    }

    if memo1.Get(3) != 6 {
        test.Error("Memo returned wrong value")
    }

    if count != 1 {
        test.Error("Memo called function more than once")
    }

    memo2 := MakeMemo[int, int](func(x... int) int {
        return 12
    })

    if memo2.Get() != 12 {
        test.Error("Memo returned wrong value")
    }
}
