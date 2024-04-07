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
	Entries        map[string]*Node.Node
	CollectionName string
}

// NewIndex returns a new Index
func NewIndex(payloadkey string, space *map[string]*Vector.Vector, collection string) (*Index, error) {
	// Create the Indexstruct
	index := &Index{Entries: make(map[string]*Node.Node), CollectionName: collection}
	// Create a vectorMap as startinpoint to create the subtrees
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
		index.Entries[(*vectors[0].Payload)[payloadkey].(string)] = n
	}
	return index, nil
}

// getVectorFromPayloadIndex returns a map for a specific payload
func (i *Index) getVectorFromPayloadIndex(payloadkey string, space *map[string]*Vector.Vector) (*map[string][]*Vector.Vector, error) {
	// Create the map
	vectorMap := make(map[string][]*Vector.Vector)

	// Loop over all the entries
	for _, vector := range *space {
		// Check if key is in the Payload
		if _, ok := (*vector.Payload)[payloadkey]; ok {
			// Load the payload from the hdd
			payload, err := FileMapper.Mapper.ReadPayload(vector.PayloadStart, i.CollectionName)
			if err != nil {
				return nil, err
			}
			//Check if the value is a string
			if _, ok := (*payload)[payloadkey].(string); !ok {
				return nil, fmt.Errorf("payload %s is not a string - this is not supported yet", payloadkey)
			}
			// Check if the key is in the map
			if _, ok := vectorMap[(*vector.Payload)[payloadkey].(string)]; !ok {
				// add it to the map
				vectorMap[(*vector.Payload)[payloadkey].(string)] = []*Vector.Vector{}
			}
			// add the vector to the map
			vectorMap[(*vector.Payload)[payloadkey].(string)] = append(vectorMap[(*vector.Payload)[payloadkey].(string)], vector)
		} else {
			continue
		}
	}
	return &vectorMap, nil
}
