package Utils

import (
	"VectoriaDB/Node"
	"container/heap"
	"context"
	"fmt"
	"sync"
)

// HeapChannelStruct is a struct that holds a channel and a heap
type HeapChannelStruct struct {
	node *Node.Node
	dist float64
	diff float64
}

// HeapControl is a struct that holds a slice of HeapItems and the maximum number of entries
type HeapControl struct {
	Heap       Heap
	maxEntries int
	In         chan HeapChannelStruct
	ctx        context.Context
	abort      context.CancelFunc
	MaxDiff    float64
	Mut        sync.RWMutex
}

// The HeapItem struct is used to store a Node and its distance to the query vector
type HeapItem struct {
	Node     *Node.Node
	Distance float64
	Diff     float64
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
	ctx, cancel := context.WithCancel(context.Background())
	h := &HeapControl{maxEntries: n, Heap: Heap{}, In: make(chan HeapChannelStruct, 1), Mut: sync.RWMutex{}, ctx: ctx, abort: cancel, MaxDiff: 0}
	heap.Init(&h.Heap)
	return h
}

// StartThreads starts the threads for the heap
func (hc *HeapControl) StartThreads() {
	go func() {
		for {
			select {
			case item := <-hc.In:
				hc.Insert(item.node, item.dist, item.diff)
			case <-hc.ctx.Done():
				return
			}
		}
	}()
}

// StopThreads stops the threads for the heap
func (hc *HeapControl) StopThreads() {
	// Send a signal to the context to stop the threads
	hc.abort()
	// Close the channel
	close(hc.In)
}

// Insert inserts a node into the heap
func (hc *HeapControl) Insert(node *Node.Node, distance, diff float64) {
	heap.Push(&hc.Heap, &HeapItem{Node: node, Distance: distance, Diff: diff})
	if hc.Heap.Len() > hc.maxEntries {
		fmt.Println("Popping")
		heap.Pop(&hc.Heap)
	}
	// Set the Maxdiff
	hc.Mut.Lock()
	for i := 0; i < len(hc.Heap); i++ {
		if hc.Heap[i].Diff > hc.MaxDiff {
			hc.MaxDiff = hc.Heap[i].Diff
		}
	}
	hc.Mut.Unlock()
}

// GetMaxDiff returns the maximum difference
func (hc *HeapControl) GetMaxDiff() float64 {
	hc.Mut.RLock()
	defer hc.Mut.RUnlock()
	return hc.MaxDiff
}

// GetNodes returns the nodes from the heap
func (hc *HeapControl) GetNodes() []*HeapItem {
	return hc.Heap
}
