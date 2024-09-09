package priority

type Heap[T any] struct {
    heap []T
    compare func(a T, b T) int
}

func MakeEmptyHeap[T any](compare func(a T, b T) int) *Heap[T] {
    return &Heap[T]{compare: compare}
}

func MakeHeap[T any](values []T, compare func(a T, b T) int) *Heap[T] {
    heap := &Heap[T]{heap: values, compare: compare}
    heap.buildHeap()
    return heap
}

func (heap *Heap[T]) buildHeap() {
    for i := len(heap.heap) / 2; i >= 0; i-- {
        heap.bubbleDown(i)
    }
}

/* push an element down the heap until it is in the correct position
 */
func (heap *Heap[T]) bubbleDown(index int) {
    for {
        left := 2 * index + 1
        right := 2 * index + 2
        min := index
        if left < len(heap.heap) && heap.compare(heap.heap[left], heap.heap[min]) < 0 {
            min = left
        }
        if right < len(heap.heap) && heap.compare(heap.heap[right], heap.heap[min]) < 0 {
            min = right
        }
        if min != index {
            heap.heap[min], heap.heap[index] = heap.heap[index], heap.heap[min]
            index = min
        } else {
            break
        }
    }
}

func (heap *Heap[T]) Insert(value T) {
    heap.heap = append(heap.heap, value)
    heap.bubbleUp(len(heap.heap) - 1)
}

func (heap *Heap[T]) bubbleUp(index int) {
    for {
        parent := (index - 1) / 2
        if parent < 0 {
            return
        }
        if heap.compare(heap.heap[parent], heap.heap[index]) > 0 {
            heap.heap[parent], heap.heap[index] = heap.heap[index], heap.heap[parent]
            index = parent
        } else {
            break
        }
    }
}

func (heap *Heap[T]) Min() T {
    return heap.heap[0]
}

func (heap *Heap[T]) RemoveMin() T {
    min := heap.heap[0]
    heap.heap[0] = heap.heap[len(heap.heap) - 1]
    heap.heap = heap.heap[:len(heap.heap) - 1]
    heap.bubbleDown(0)
    return min
}

func (heap *Heap[T]) Size() int {
    return len(heap.heap)
}

func (heap *Heap[T]) IsEmpty() bool {
    return len(heap.heap) == 0
}

