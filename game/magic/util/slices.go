package util

func RotateSlice[T any](slice []T, forward bool){
    if len(slice) <= 1 {
        return
    }

    last := len(slice) - 1

    if forward {
        v := slice[0]

        for i := 0; i < len(slice) - 1; i++ {
            slice[i] = slice[i+1]
        }
        slice[last] = v
    } else {
        v := slice[last]
        for i := last; i > 0; i-- {
            slice[i] = slice[i-1]
        }
        slice[0] = v
    }
}

func First[T any](slice []T, default_ T) T {
    if len(slice) == 0 {
        return default_
    }

    return slice[0]
}
