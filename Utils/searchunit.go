package Utils

import (
	"VreeDB/ArgsParser"
	"VreeDB/Filter"
	"VreeDB/Node"
	"VreeDB/Vector"
	"math"
	"sync"
)

// SearchUnit represents a unit used for searching.
type SearchUnit struct {
	dimensionMultiplier float64
	Filter              *[]Filter.Filter
	Chan                chan *SearchData
	wg                  *sync.WaitGroup
}

type SearchData struct {
	Node          *Node.Node
	Target        *Vector.Vector
	Queue         *HeapControl
	DistanceFunc  func(*Vector.Vector, *Vector.Vector) (float64, error)
	DimensionDiff *Vector.Vector
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

// Start starts the SearchThreadpool
func (s *SearchUnit) Start() {
	for i := 0; i < *ArgsParser.Ap.SearchThreads; i++ {
		go func() {
			for data := range s.Chan {
				s.NearestNeighbors(data.Node, data.Target, data.Queue, data.DistanceFunc, data.DimensionDiff)
			}
		}()
	}
}

// NewSearchUnit returns a new SearchUnit
func NewSearchUnit(node *Node.Node, target *Vector.Vector, queue *HeapControl, filter *[]Filter.Filter,
	distanceFunc func(*Vector.Vector, *Vector.Vector) (float64, error),
	dimensionDiff *Vector.Vector, dimensionMultiplier float64) {
	su := SearchUnit{dimensionMultiplier: dimensionMultiplier, Filter: filter}
	su.NearestNeighbors(node, target, queue, distanceFunc, dimensionDiff)
}
