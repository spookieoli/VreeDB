package Node

import (
	"time"
)

// Node is a struct that holds a Vector and two pointers to other Nodes
type Node struct {
	Vector   *Vector
	Left     *Node
	Right    *Node
	Depth    int
	LastUsed time.Time
	Used     int
	self     *Node
}

// ACESInsert is the same like insert but takes a node as parameter
func (n *Node) ACESInsert(node *Node) {
	if n.Vector == nil {
		n.Vector = node.Vector
		n.self = node
		return
	}

	// Get the current axis
	axis := n.Depth % n.Vector.Length

	// Compare the new vector to the current vector
	if node.Vector.Data[axis] < n.Vector.Data[axis] {
		if n.Left == nil {
			n.Left = &Node{Depth: axis + 1}
		}
		n.Left.ACESInsert(node)
		return
	} else {
		if n.Right == nil {
			n.Right = &Node{Depth: axis + 1}
		}
		n.Right.ACESInsert(node)
		return
	}
}

// Insert inserts a Node into the tree
func (n *Node) Insert(newVector *Vector) {
	if n.Vector == nil {
		n.Vector = newVector
		n.Vector.Node = n
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
