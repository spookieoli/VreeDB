package Collection

import (
	"VreeDB/FileMapper"
	"VreeDB/Node"
	"VreeDB/Vector"
	"fmt"
)

// Index is the type to index specific vector payloads
type Index struct {
	// Indexes are sub kd trees
	Entries        map[any]*Node.Node
	CollectionName string
	Key            string
}

// NewIndex returns a new Index
func NewIndex(payloadkey string, space *map[string]*Vector.Vector, collection string) (*Index, error) {
	// Create the Indexstruct
	index := &Index{Entries: make(map[any]*Node.Node), CollectionName: collection, Key: payloadkey}

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

		// Insert the Node into the Index
		switch v := (*vectors[0].Payload)[payloadkey].(type) {
		case int:
		case float64:
		case string:
			index.Entries[v] = n
		default:
			return nil, fmt.Errorf("only string, float64 and int are allowed")
		}
	}
	return index, nil
}

// getVectorFromPayloadIndex returns a map for a specific payload
func (i *Index) getVectorFromPayloadIndex(payloadkey string, space *map[string]*Vector.Vector) (*map[any][]*Vector.Vector, error) {
	// Create the map
	vectorMap := make(map[any][]*Vector.Vector)

	// Loop over all the entries
	for _, vector := range *space {

		// Check if key is in the Payload
		if _, ok := (*vector.Payload)[payloadkey]; ok {

			// Load the payload from the hdd
			payload, err := FileMapper.Mapper.ReadPayload(vector.PayloadStart, i.CollectionName)
			if err != nil {
				return nil, err
			}

			// only string, int and float64 are allowed
			switch v := (*payload)[payloadkey].(type) {
			case int:
			case float64:
			case string:
				if _, ok := vectorMap[v]; !ok {
					vectorMap[v] = []*Vector.Vector{}
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
