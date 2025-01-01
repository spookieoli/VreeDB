package Vdb

type ACESRouter struct {
	VDB *Vdb
}

// New returns a new ACESRouter
func New(vdb *Vdb) *ACESRouter {
	return &ACESRouter{VDB: vdb}
}
