package Utils

import (
	"VectoriaDB/Node"
	"VectoriaDB/Vector"
	"math"
	"sync/atomic"
)

type SearchUnit struct {
}

// NearestNeighbors returns the results nearest neighbours to the given target vector
func (s *SearchUnit) NearestNeighbors(node *Node.Node, target *Vector.Vector, queue *HeapControl, distanceFunc func(*Vector.Vector, *Vector.Vector) (float64, error)) {
	if node == nil || node.Vector == nil {
		return
	}
	atomic.AddUint64(&Utils.Searched, 1)
	axis := node.Depth % node.Vector.Length

	// Use the vector Functions
	dist, _ := distanceFunc(node.Vector, target)
	axisDiff := math.Abs(target.Data[axis] - node.Vector.Data[axis])

	// Just push it into the queue if it is small enough it will be added
	queue.AddToWaitGroup()
	queue.In <- HeapChannelStruct{node: node, dist: dist, diff: axisDiff}

	var primary, secondary *Node.Node
	if target.Data[axis] < node.Vector.Data[axis] {
		primary, secondary = node.Left, node.Right
	} else {
		primary, secondary = node.Right, node.Left
	}

	s.NearestNeighbors(primary, target, queue, distanceFunc)

	// Check if we should go further
	if target.Data[axis] > node.Vector.Data[axis] {
		s.NearestNeighbors(secondary, target, queue, distanceFunc)
	}
}

// NewSearchUnit returns a new SearchUnit
func NewSearchUnit(node *Node.Node, target *Vector.Vector, queue *HeapControl, distanceFunc func(*Vector.Vector, *Vector.Vector) (float64, error)) {
	su := SearchUnit{}
	su.NearestNeighbors(node, target, queue, distanceFunc)
}
