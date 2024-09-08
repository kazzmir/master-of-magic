package priority

/* a priority queue using an array-based min heap
 */

type PriorityQueue struct {
    Heap *Heap
}

func MakePriorityQueue() *PriorityQueue {
    return &PriorityQueue{MakeEmptyHeap()}
}

func (queue *PriorityQueue) Insert(value int) {
    queue.Heap.Insert(value)
}

func (queue *PriorityQueue) ExtractMin() int {
    return queue.Heap.RemoveMin()
}

func (queue *PriorityQueue) IsEmpty() bool {
    return queue.Heap.IsEmpty()
}

func (queue *PriorityQueue) Size() int {
    return queue.Heap.Size()
}

func (queue *PriorityQueue) Clear() {
    queue.Heap = MakeEmptyHeap()
}

func (queue *PriorityQueue) Top() int {
    return queue.Heap.Min()
}
