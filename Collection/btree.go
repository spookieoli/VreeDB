package Collection

// TreeNode represents a node in a tree data structure.
// It contains an array of keys, an array of child nodes, and a flag indicating
// if the node is a leaf node.
type TreeNode struct {
	keys     []int
	children []*TreeNode
	isLeaf   bool
}

// Tree represents a tree data structure.
// It contains a pointer to the root node of the tree.
type Tree struct {
	root *TreeNode
}
