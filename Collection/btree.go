package Collection

// TreeNode represents a node in a tree data structure.
// It contains an array of keys, an array of child nodes, and a flag indicating
// if the node is a leaf node.
type TreeNode struct {
	keys     []int
	children []*TreeNode
	isLeaf   bool
}
