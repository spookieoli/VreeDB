package Utils

import (
	"VectoriaDB/Node"
	"container/heap"
)

// HeapControl is a struct that holds a slice of HeapItems and the maximum number of entries
type HeapControl struct {
	Heap       Heap
	maxEntries int
}

// The HeapItem struct is used to store a Node and its distance to the query vector
type HeapItem struct {
	Node     *Node.Node
	Distance float64
}

// Heap will be used to implement the heap interface
type Heap []*HeapItem

// Len returns the length of the heap
func (h Heap) Len() int {
	return len(h)
}

// Less compares two items in the heap > will be used to create a max heap
func (h Heap) Less(i, j int) bool {
	return h[i].Distance > h[j].Distance
}

// Swap swaps two items in the heap
func (h Heap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// Push pushes an item into the heap
func (h *Heap) Push(x interface{}) {
	*h = append(*h, x.(*HeapItem))
}

// Pop pops an item from the heap
func (h *Heap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// NewHeapControl initializes the heap with a given size
func NewHeapControl(n int) *HeapControl {
	h := &HeapControl{maxEntries: n, Heap: Heap{}}
	heap.Init(&h.Heap)
	return h
}

// Insert inserts a node into the heap
func (hc *HeapControl) Insert(node *Node.Node, distance float64) {
	heap.Push(&hc.Heap, &HeapItem{Node: node, Distance: distance})
	if hc.Heap.Len() > hc.maxEntries {
		heap.Pop(&hc.Heap)
	}
}

// GetNodes returns the nodes from the heap
func (hc *HeapControl) GetNodes() []*HeapItem {
	return hc.Heap
}
