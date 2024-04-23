package Utils

import (
	"VreeDB/Filter"
	"VreeDB/Node"
	"VreeDB/Vector"
	"math"
)

type SearchUnit struct {
	dimensionMultiplier float64
	Filter              *[]Filter.Filter
}

// NearestNeighbors returns the results nearest neighbours to the given target vector
func (s *SearchUnit) NearestNeighbors(node *Node.Node, target *Vector.Vector, queue *HeapControl,
	distanceFunc func(*Vector.Vector, *Vector.Vector) (float64, error), dimensionDiff *Vector.Vector) {
	if node == nil || node.Vector == nil {
		return
	}
	axis := node.Depth % node.Vector.Length

	// Use the vector Functions
	dist, _ := distanceFunc(node.Vector, target)
	axisDiff := math.Abs(target.Data[axis] - node.Vector.Data[axis])

	// Just push it into the queue if it is small enough it will be added
	queue.AddToWaitGroup()
	queue.In <- HeapChannelStruct{node: node, dist: dist, diff: axisDiff, Filter: s.Filter}

	var primary, secondary *Node.Node
	if target.Data[axis] < node.Vector.Data[axis] {
		primary = node.Left
		secondary = node.Right
	} else {
		primary = node.Right
		secondary = node.Left
	}
	s.NearestNeighbors(primary, target, queue, distanceFunc, dimensionDiff)

	// If the distance is smaller than the dimensionDiff we need to search the other side
	if axisDiff < dimensionDiff.Data[axis]*s.dimensionMultiplier {
		s.NearestNeighbors(secondary, target, queue, distanceFunc, dimensionDiff)
	}
}

// NewSearchUnit returns a new SearchUnit
func NewSearchUnit(node *Node.Node, target *Vector.Vector, queue *HeapControl, filter *[]Filter.Filter,
	distanceFunc func(*Vector.Vector, *Vector.Vector) (float64, error),
	dimensionDiff *Vector.Vector, dimensionMultiplier float64) {
	su := SearchUnit{dimensionMultiplier: dimensionMultiplier, Filter: filter}
	su.NearestNeighbors(node, target, queue, distanceFunc, dimensionDiff)
}
