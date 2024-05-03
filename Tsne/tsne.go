package Tsne

import (
	"VreeDB/Vector"
	"sync"
)

type Tsne struct {
	Vectors         *Vector.Vector
	Mutex           *sync.RWMutex
	Projection      *Vector.Vector
	Iterations      int
	TargetDimension int
}
