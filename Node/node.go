package Node

import (
	"VreeDB/Vector"
	"time"
)

// Node is a struct that holds a Vector and two pointers to other Nodes
type Node struct {
	Vector   *Vector.Vector
	Left     *Node
	Right    *Node
	Depth    int
	LastUsed time.Time
	Used     int
}

// Insert inserts a Node into the tree // TBD: Will be in the Collection package
func (n *Node) Insert(newVector *Vector.Vector) {
	if n.Vector == nil {
		n.Vector = newVector
		return
	}

	// Get the current axis
	axis := n.Depth % n.Vector.Length

	// Checking if the vector is new or will be readded bei recreation
	if newVector.Collection != "" {
		newVector.Unindex()
	}

	// Compare the new vector to the current vector
	if newVector.Data[axis] < n.Vector.Data[axis] {
		if n.Left == nil {
			n.Left = &Node{Depth: axis + 1}
		}
		n.Left.Insert(newVector)
		return
	} else {
		if n.Right == nil {
			n.Right = &Node{Depth: axis + 1}
		}
		n.Right.Insert(newVector)
		return
	}
}
