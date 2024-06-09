package Utils

import (
	"VreeDB/Filter"
	"VreeDB/Logger"
	"VreeDB/Node"
	"container/heap"
	"sync"
)

// HeapChannelStruct is a struct that holds a channel and a heap
type HeapChannelStruct struct {
	node   *Node.Node
	dist   float64
	diff   float64
	Filter *[]Filter.Filter
}

// HeapControl is a struct that holds a slice of HeapItems and the maximum number of entries
type HeapControl struct {
	Heap       Heap
	maxEntries int
	In         chan HeapChannelStruct
	MaxDiff    float64
	Wg         sync.WaitGroup
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
	h := &HeapControl{maxEntries: n, Heap: Heap{}, In: make(chan HeapChannelStruct, 100000), MaxDiff: 0}
	heap.Init(&h.Heap)
	return h
}

// StartThreads starts the threads for the heap
func (hc *HeapControl) StartThreads() {
	go hc.worker()
}

// Worker will work on the threads
func (hc *HeapControl) worker() {
	defer hc.Wg.Done()
	for item := range hc.In {
		// Validate the filters
		if ok, err := hc.validateFilters(&item); !ok {
			// If the filters are not valid log the possible error
			if err != nil {
				Logger.Log.Log("Error validating filters: "+err.Error(), "ERROR")
			}
		} else if item.node.Vector.IsDeleted() {
			continue
		} else {
			// Insert the item into the heap
			hc.Insert(item.node, item.dist, item.diff)
		}
	}
}

// ValidateFilters will validate the filters on a given Node
func (hc *HeapControl) validateFilters(hcs *HeapChannelStruct) (bool, error) {
	// Dont do anything if there are no filters
	if hcs.Filter == nil {
		return true, nil
	}
	// Validate the filters
	for _, f := range *hcs.Filter {
		if ok, err := f.ValidateFilter(hcs.node.Vector); !ok {
			return false, err
		}
	}
	return true, nil
}

// Insert inserts a node into the heap
func (hc *HeapControl) Insert(node *Node.Node, distance, diff float64) {
	heap.Push(&hc.Heap, &HeapItem{Node: node, Distance: distance, Diff: diff})
	if hc.Heap.Len() > hc.maxEntries {
		heap.Pop(&hc.Heap)
	}
}

// AddToWaitGroup adds a new item to the waitgroup
func (hc *HeapControl) AddToWaitGroup() {
	hc.Wg.Add(1)
}

// CloseChannel closes the channel
func (hc *HeapControl) CloseChannel() {
	close(hc.In)
}

// GetNodes returns the nodes from the heap
func (hc *HeapControl) GetNodes() []*HeapItem {
	return hc.Heap
}
