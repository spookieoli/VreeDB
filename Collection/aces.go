package Collection

import (
	"VreeDB/ArgsParser"
	"VreeDB/Node"
	"sync"
)

// Ac is an advanced cluster - this cluster is build from a small subset of collection Nodes
type Ac struct {
	Nodes        *Node.Node
	Mut          *sync.RWMutex
	Collection   *Collection
	Count        int64
	Distribution float64
}

// NewAc returns a new Ac
func NewAc(collection *Collection) *Ac {
	// Create a new Ac
	return &Ac{Nodes: &Node.Node{Depth: 0}, Mut: &sync.RWMutex{}, Collection: collection, Count: 0, Distribution: *ArgsParser.Ap.ACDistribution}
}

// TODO: Implement Add, Remove, Renew functions
