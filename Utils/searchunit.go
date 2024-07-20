package Utils

import (
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
	SU            *SearchUnit
}

// NearestNeighbors returns the results nearest neighbours to the given target vector.
// It calculates the axis for the given node and computes the distance and axis difference
// between the node vector and the target vector. It then pushes the node into the queue
// by invoking the `In` channel of `queue`. It also decides whether to search the left or
// right child node based on the target vector values. Finally, it recursively calls
// `NearestNeighbors` on the primary and secondary child nodes.
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
	s.AddToWaitGroup()
	// We put this in a goroutine to prevent a possible deadlock
	go func() {
		s.Chan <- &SearchData{Node: primary, Target: target, Queue: queue, DistanceFunc: distanceFunc, DimensionDiff: dimensionDiff, SU: s}
	}()

	// If the distance is smaller than the dimensionDiff we need to search the other side
	if axisDiff < dimensionDiff.Data[axis]*s.dimensionMultiplier {
		s.AddToWaitGroup()
		// We put this in a goroutine to prevent a possible deadlock
		go func() {
			s.Chan <- &SearchData{secondary, target, queue, distanceFunc, dimensionDiff, s}
		}()
	}
}

// NewSearchUnit returns a new SearchUnit
func NewSearchUnit(filter *[]Filter.Filter, dimensionMultiplier float64) *SearchUnit {
	return &SearchUnit{dimensionMultiplier: dimensionMultiplier, Filter: filter, Chan: Searcher.GetChan(), wg: &sync.WaitGroup{}}
}

// AddToWaitGroup blocks until the SearchUnit is finished
func (s *SearchUnit) AddToWaitGroup() {
	s.wg.Add(1)
}

// ReleaseWaitGroup releases the WaitGroup
func (s *SearchUnit) releaseWaitGroup() {
	s.wg.Done()
}

// Search starts the search
func (s *SearchUnit) Search(node *Node.Node, target *Vector.Vector, queue *HeapControl,
	distanceFunc func(*Vector.Vector, *Vector.Vector) (float64, error), dimensionDiff *Vector.Vector) {
	s.AddToWaitGroup()
	s.Chan <- &SearchData{Node: node, Target: target, Queue: queue, DistanceFunc: distanceFunc, DimensionDiff: dimensionDiff, SU: s}
	s.wg.Wait()
}
