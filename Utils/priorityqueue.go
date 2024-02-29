package Utils

import (
	"VectoriaDB/Node"
	"math"
	"sort"
	"sync"
)

type ResultItem struct {
	Node     *Node.Node
	Distance float64
	Diff     float64
}

type PriorityQueue struct {
	Nodes       []*ResultItem
	Depth       int
	Mut         *sync.Mutex
	MaxDistance float64
	MaxDiff     float64
}

// NewPriorityQueue returns a new PriorityQueue
func NewPriorityQueue(depth int) *PriorityQueue {
	return &PriorityQueue{Depth: depth, MaxDistance: math.MaxFloat64, Mut: &sync.Mutex{}}
}

// Push pushes a Node into the queue, they may only be Depth items in the queue
func (pq *PriorityQueue) Push(node *Node.Node, distance float64, diff float64) {
	pq.Mut.Lock()
	defer pq.Mut.Unlock()
	if len(pq.Nodes) < pq.Depth {
		pq.Nodes = append(pq.Nodes, &ResultItem{Node: node, Distance: distance, Diff: diff})
		// Update MaxDistance
		sort.Slice(pq.Nodes, func(i, j int) bool {
			return pq.Nodes[i].Diff < pq.Nodes[j].Diff
		})
		pq.MaxDiff = pq.Nodes[len(pq.Nodes)-1].Diff
		return
	}

	// is the distance given smaller than one of the distances in the queue?
	if pq.GetSomeDistance(distance) {
		// if yes, add the node to the queue
		pq.Nodes = append(pq.Nodes, &ResultItem{Node: node, Distance: distance, Diff: diff})
		// sort the queue by distance
		sort.Slice(pq.Nodes, func(i, j int) bool {
			return pq.Nodes[i].Distance < pq.Nodes[j].Distance
		})
		// and only keep the first Depth items
		pq.Nodes = pq.Nodes[:pq.Depth]
		// Update MaxDiff
		sort.Slice(pq.Nodes, func(i, j int) bool {
			return pq.Nodes[i].Diff < pq.Nodes[j].Diff
		})
		pq.MaxDiff = pq.Nodes[len(pq.Nodes)-1].Diff
		return
	}
}

// GetMaxDistanceIndex returns the index of the MaxDistance - WARNING: This is not thread safe - may only be called if from Push
func (p *PriorityQueue) getMaxDistanceIndex() int {
	for i, d := range p.Nodes {
		if d.Distance == p.MaxDistance {
			return i
		}
	}
	return -1
}

// GetMaxDiffIndex returns the index of the MaxDiff - WARNING: This is not thread safe - may only be called if from Push
func (p *PriorityQueue) getMaxDiffIndex() int {
	for i, d := range p.Nodes {
		if d.Diff == p.MaxDiff {
			return i
		}
	}
	return -1
}

// IsNotFull returns true if the queue is not full
func (p *PriorityQueue) IsNotFull() bool {
	p.Mut.Lock()
	defer p.Mut.Unlock()
	return len(p.Nodes) < p.Depth
}

// GetMaxDistance returns the MaxDistance
func (p *PriorityQueue) GetMaxDistance() float64 {
	p.Mut.Lock()
	defer p.Mut.Unlock()
	return p.MaxDistance
}

// GetMaxDiff returns the MaxDiff
func (p *PriorityQueue) GetMaxDiff() float64 {
	p.Mut.Lock()
	defer p.Mut.Unlock()
	return p.MaxDiff
}

// GetSomeDistance checks if the given Distance is smaller than one of the distances in the queue
func (p *PriorityQueue) GetSomeDistance(distance float64) bool {
	for _, d := range p.Nodes {
		if d.Distance > distance {
			return true
		}
	}
	return false
}

// GetSomeDistance checks if the given Distance is smaller than one of the distances in the queue
func (p *PriorityQueue) OuterGetSomeDistance(distance float64) bool {
	p.Mut.Lock()
	defer p.Mut.Unlock()
	for _, d := range p.Nodes {
		if d.Distance > distance {
			return true
		}
	}
	return false
}
