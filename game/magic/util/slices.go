package util

// RotateSlice rotates the elements of slice in-place.
// If forward is true the slice is rotated left (first element becomes last),
// otherwise it is rotated right (last element becomes first).
func RotateSlice[T any](slice []T, forward bool) {
    n := len(slice)
    if n <= 1 {
        return
    }

    last := n - 1
    if forward {
        // shift left by one: copy elements [1:] into [0:]
        first := slice[0]
        copy(slice[0:last], slice[1:])
        slice[last] = first
    } else {
        // shift right by one: copy elements [:last] into [1:]
        lastv := slice[last]
        copy(slice[1:], slice[0:last])
        slice[0] = lastv
    }
}

// First returns the first element of slice, or default_ when slice is empty.
func First[T any](slice []T, default_ T) T {
    if len(slice) == 0 {
        return default_
    }
    return slice[0]
}
