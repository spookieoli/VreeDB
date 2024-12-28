package Collection

import (
	"VreeDB/ArgsParser"
	"VreeDB/Node"
	"VreeDB/Utils"
	"VreeDB/Vector"
	"sync"
)

// Ac is an advanced cluster - this cluster is build from a small subset of collection Nodes
type Ac struct {
	Nodes        *Node.Node
	Space        []*Node.Node
	Mut          *sync.RWMutex
	Collection   *Collection
	Count        int64
	Distribution float64
	DistanceFunc func(*Vector.Vector, *Vector.Vector) (float64, error)
}

// NewAc returns a new Ac
func NewAc(collection *Collection) *Ac {
	// Create a new Ac
	ac := &Ac{Nodes: &Node.Node{Depth: 0}, Mut: &sync.RWMutex{}, Collection: collection, Count: 0, Distribution: *ArgsParser.Ap.ACDistribution}

	// Get actual Distance func, depends on AVX enabled or not
	if *ArgsParser.Ap.AVX256 {
		ac.DistanceFunc = Utils.Utils.EuclideanDistanceAVX256
	} else {
		ac.DistanceFunc = Utils.Utils.EuclideanDistance
	}

	// If we are on ARM, use Neon
	if *ArgsParser.Ap.Neon {
		ac.DistanceFunc = Utils.Utils.EuclideanDistanceNEON
	}

	return ac
}

// Insert inserts a Node into the Ac
func (a *Ac) Insert(node *Node.Node) {
	// check if we have already enough ACES points
	if a.Count >= int64(float64(a.Collection.GetNodeCount())**ArgsParser.Ap.ACDistribution) {
		return
	}

	// Check if the distance is greater to all added nodes
	if a.chkDistances(node.Vector) {
		a.Nodes.ACESInsert(node)
		a.Space = append(a.Space, node)
		a.Count++
	}
}

// chkDistances checks the distances of the nodes in the Ac
func (a *Ac) chkDistances(v *Vector.Vector) bool {
	a.Mut.RLock()
	defer a.Mut.RUnlock()
	// Get the Diagonal Length of the Collection
	dl := a.Collection.DiagonalLength * *ArgsParser.Ap.ACESMindDist

	// Check every node in the Ac - if the distance to one of the Vectors is smaller than ACESMindDist, break
	// For now, we will only use euclidean distance
	for _, node := range a.Space {
		dist, err := a.DistanceFunc(node.Vector, v)
		if err != nil {
			return false
		}

		// If the distance is greater or equal than the minimum distance, return continue
		if dist >= dl {
			continue
		}
	}
	return true
}
