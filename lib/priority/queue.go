package priority

/* a priority queue using an array-based min heap
 */

type Heap struct {
    heap []int
}

func MakeEmptyHeap() *Heap {
    return &Heap{}
}

func MakeHeap(values []int) *Heap {
    heap := &Heap{heap: values}
    heap.buildHeap()
    return heap
}

func (heap *Heap) buildHeap() {
    for i := len(heap.heap) / 2; i >= 0; i-- {
        heap.bubbleDown(i)
    }
}

/* push an element down the heap until it is in the correct position
 */
func (heap *Heap) bubbleDown(index int) {
    for {
        left := 2 * index + 1
        right := 2 * index + 2
        min := index
        if left < len(heap.heap) && heap.heap[left] < heap.heap[min] {
            min = left
        }
        if right < len(heap.heap) && heap.heap[right] < heap.heap[min] {
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

func (heap *Heap) Insert(value int) {
    heap.heap = append(heap.heap, value)
    heap.bubbleUp(len(heap.heap) - 1)
}

func (heap *Heap) bubbleUp(index int) {
    for {
        parent := (index - 1) / 2
        if parent < 0 {
            return
        }
        if heap.heap[parent] > heap.heap[index] {
            heap.heap[parent], heap.heap[index] = heap.heap[index], heap.heap[parent]
            index = parent
        } else {
            break
        }
    }
}

func (heap *Heap) Min() int {
    return heap.heap[0]
}

func (heap *Heap) RemoveMin() int {
    min := heap.heap[0]
    heap.heap[0] = heap.heap[len(heap.heap) - 1]
    heap.heap = heap.heap[:len(heap.heap) - 1]
    heap.bubbleDown(0)
    return min
}

func (heap *Heap) Size() int {
    return len(heap.heap)
}

func (heap *Heap) IsEmpty() bool {
    return len(heap.heap) == 0
}

