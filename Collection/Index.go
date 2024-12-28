package Collection

import (
	"VreeDB/FileMapper"
	"VreeDB/Node"
	"fmt"
	"sync"
)

// Index is the type to index specific vector payloads
type Index struct {
	// Indexes are sub kd trees
	Entries        map[any]*Node.Node
	CollectionName string
	Key            string
	mut            *sync.RWMutex
}

// NewIndex returns a new Index
func NewIndex(payloadkey string, space *map[string]*Node.Vector, collection string) (*Index, error) {
	// Create the Indexstruct
	index := &Index{Entries: make(map[any]*Node.Node), CollectionName: collection, Key: payloadkey, mut: &sync.RWMutex{}}

	// Create a vectorMap as starting point to create the subtrees
	vectorMap, err := index.getVectorFromPayloadIndex(payloadkey, space)

	// Check for errors
	if err != nil {
		return nil, err
	}

	// Build the subtrees
	for _, vectors := range *vectorMap {
		// Create a new Node
		n := &Node.Node{Depth: 0}
		// Insert the vectors into the Node
		for _, vector := range vectors {
			n.Insert(vector)
		}

		// Get the payload from the hdd
		payload, err := FileMapper.Mapper.ReadPayload(vectors[0].PayloadStart, collection)
		if err != nil {
			return nil, err
		}

		// Insert the Node into the Index
		switch v := (*payload)[payloadkey].(type) {
		case int, float64, string:
			index.Entries[v] = n
		default:
			return nil, fmt.Errorf("only string, float64 and int are allowed")
		}
	}
	return index, nil
}

// getVectorFromPayloadIndex returns a map for a specific payload
func (i *Index) getVectorFromPayloadIndex(payloadkey string, space *map[string]*Node.Vector) (*map[any][]*Node.Vector, error) {
	// Create the map
	vectorMap := make(map[any][]*Node.Vector)

	// Loop over all the entries
	for _, vector := range *space {

		// Load the payload from the hdd
		payload, err := FileMapper.Mapper.ReadPayload(vector.PayloadStart, i.CollectionName)
		if err != nil {
			return nil, err
		}

		// Check if key is in the Payload
		if _, ok := (*payload)[payloadkey]; ok {
			// only string, int and float64 are allowed
			switch v := (*payload)[payloadkey].(type) {
			case int, float64, string:
				if _, ok := vectorMap[v]; !ok {
					vectorMap[v] = []*Node.Vector{}
				}
				// Add to the vectorMap
				vectorMap[v] = append(vectorMap[v], vector)
			default:
				return nil, fmt.Errorf("only string, float64 and int are allowed")
			}
		} else {
			continue
		}
	}
	return &vectorMap, nil
}

// AddToIndex adds a vector to the Index
func (i *Index) AddToIndex(vector *Node.Vector) error {

	// Get the Payload from the hdd
	payload, err := FileMapper.Mapper.ReadPayload(vector.PayloadStart, i.CollectionName)
	if err != nil {
		return err
	}

	// Check if the key is in the Payload
	if _, ok := i.Entries[(*payload)[i.Key]]; !ok {
		// Add the key to the Index
		i.Entries[(*payload)[i.Key]] = &Node.Node{Depth: 0}
	}

	// add it to the Node
	i.Entries[(*vector.Payload)[i.Key]].Insert(vector)
	return nil
}
