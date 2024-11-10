package Collection

import (
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
