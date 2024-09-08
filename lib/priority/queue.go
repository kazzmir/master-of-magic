package priority

/* a priority queue using an array-based min heap
 */

type PriorityQueue[T any] struct {
    Heap *Heap[T]
}

func MakePriorityQueue[T any](compare func(T,T) int) *PriorityQueue[T] {
    return &PriorityQueue[T]{Heap: MakeEmptyHeap[T](compare)}
}

func (queue *PriorityQueue[T]) Insert(value T) {
    queue.Heap.Insert(value)
}

func (queue *PriorityQueue[T]) ExtractMin() T {
    return queue.Heap.RemoveMin()
}

func (queue *PriorityQueue[T]) IsEmpty() bool {
    return queue.Heap.IsEmpty()
}

func (queue *PriorityQueue[T]) Size() int {
    return queue.Heap.Size()
}

func (queue *PriorityQueue[T]) Clear() {
    queue.Heap = MakeEmptyHeap(queue.Heap.compare)
}

func (queue *PriorityQueue[T]) Top() T {
    return queue.Heap.Min()
}
