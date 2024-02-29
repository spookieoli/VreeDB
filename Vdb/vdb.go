package Vdb

import (
	"VectoriaDB/Collection"
	"VectoriaDB/FileMapper"
	"VectoriaDB/Logger"
	"VectoriaDB/Utils"
	"VectoriaDB/Vector"
	"fmt"
	"sort"
	"sync"
)

type Vdb struct {
	Collections map[string]*Collection.Collection
	Mapper      *FileMapper.FileMapper
}

// DB is the global Vdb
var DB *Vdb

// init initializes the Vdb
func init() {
	DB = &Vdb{Mapper: FileMapper.Mapper}
	Logger.Log.Log("VectorDatabase initialized")
}

// InitFileMapper initializes the FileMapper
func (v *Vdb) InitFileMapper() {
	// Create a slice out of the collections map
	var collections []string
	for _, key := range v.Collections {
		collections = append(collections, key.Name)
	}
	FileMapper.Mapper.Start(collections)
}

// AddCollection creates a new Collection
func (v *Vdb) AddCollection(name string, vectorDimension int, distanceFunc string) error {
	// Check if collection allready exists
	if _, ok := v.Collections[name]; ok {
		return fmt.Errorf("Collection with name %s allready exists", name)
	}
	v.Collections[name] = Collection.NewCollection(name, vectorDimension, distanceFunc)
	// Add the collection to the FileMapper
	v.Mapper.AddCollection(name)
	// Write the Collection to the FS
	err := v.Collections[name].WriteConfig()
	if err != nil {
		return err
	}
	Logger.Log.Log("Collection " + name + " added")
	return nil
}

// DeleteCollection deletes a Collection
func (v *Vdb) DeleteCollection(name string) error {
	if _, ok := v.Collections[name]; !ok {
		return fmt.Errorf("Collection with name %s does not exist", name)
	}
	delete(v.Collections, name)
	// Delete the Collection from the FileMapper
	v.Mapper.DelCollection(name)
	Logger.Log.Log("Collection " + name + " deleted")
	return nil
}

// ListCollections returns a list of all collections names
func (v *Vdb) ListCollections() []string {
	var collections []string
	for key := range v.Collections {
		collections = append(collections, key)
	}
	return collections
}

// Search searches for the nearest neighbours of the given target vector
func (v *Vdb) Search(collectionName string, target *Vector.Vector, queue *Utils.PriorityQueue, maxDistancePercent float64) []*Utils.ResultItem {
	v.Collections[collectionName].Mut.RLock()
	defer v.Collections[collectionName].Mut.RUnlock()
	wg := sync.WaitGroup{}
	Utils.NewSearchUnit(v.Collections[collectionName].Nodes, target, queue, v.Collections[collectionName].DistanceFunc)
	wg.Wait()

	// If this collection uses euclid and we have a maxDistancePercent > 0 we need to filter the results
	if v.Collections[collectionName].DistanceFuncName == "euclid" && maxDistancePercent > 0 {
		// If a result is greater than maxDistancePercent * DiagonalLength we remove it
		for i := 0; i < len(queue.Nodes); i++ {
			if queue.Nodes[i].Distance > maxDistancePercent*v.Collections[collectionName].DiagonalLength {
				queue.Nodes = append(queue.Nodes[:i], queue.Nodes[i+1:]...)
				i--
			}
		}
	}

	// Sort the results -- TODO: USE NEW SLICE SORTING
	sort.Slice(queue.Nodes, func(i, j int) bool {
		return queue.Nodes[i].Distance < queue.Nodes[j].Distance
	})

	// Get the Payloads back from the Memory Map
	for i := 0; i < len(queue.Nodes); i++ {
		m, err := FileMapper.Mapper.ReadPayload(queue.Nodes[i].Node.Vector.PayloadStart, collectionName)
		if err != nil {
			Logger.Log.Log("Error reading payload: " + err.Error())
			continue
		}
		queue.Nodes[i].Node.Vector.Payload = m
	}

	return queue.Nodes
}
