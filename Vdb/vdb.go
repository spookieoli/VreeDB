package Vdb

import (
	"VreeDB/Collection"
	"VreeDB/FileMapper"
	"VreeDB/Filter"
	"VreeDB/Logger"
	"VreeDB/Node"
	"VreeDB/Utils"
	"fmt"
	"sort"
)

// Vdb is the main struct of the VectorDatabase
type Vdb struct {
	Collections map[string]*Collection.Collection
	Mapper      *FileMapper.FileMapper
}

// DB is the global Vdb
var DB *Vdb

// init initializes the Vdb
func init() {
	DB = &Vdb{Mapper: FileMapper.Mapper}
	Logger.Log.Log("VectorDatabase initialized", "INFO")
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
func (v *Vdb) AddCollection(name string, vectorDimension int, distanceFunc string, aces bool) error {
	// Check if collection allready exists
	if _, ok := v.Collections[name]; ok {
		return fmt.Errorf("Collection with name %s allready exists", name)
	}
	v.Collections[name] = Collection.NewCollection(name, vectorDimension, distanceFunc, aces)
	// Add the collection to the FileMapper
	v.Mapper.AddCollection(name)
	// Write the Collection to the FS
	err := v.Collections[name].WriteConfig()
	if err != nil {
		return err
	}
	Logger.Log.Log("Collection "+name+" added", "INFO")
	return nil
}

// DeleteCollection deletes a Collection
func (v *Vdb) DeleteCollection(name string) error {
	if _, ok := v.Collections[name]; !ok {
		return fmt.Errorf("Collection with name %s does not exist", name)
	}

	// Cancel the ACES GoRoutine
	if v.Collections[name].ACES {
		v.Collections[name].ACESCancel()
	}

	// Delet the Collection from the FS
	delete(v.Collections, name)

	// Delete the Collection from the FileMapper
	v.Mapper.DelCollection(name)
	Logger.Log.Log("Collection "+name+" deleted", "INFO")
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

// DeletePoint will delete a point from a collection
func (v *Vdb) DeletePoint(collectionName string, vector []float64) error {
	// serach the point in the collection
	novector := false
	getid := true

	// Create the SearchParams
	sp := &Utils.SearchParams{
		CollectionName: collectionName,
		Target:         &Node.Vector{Data: vector},
		Queue:          Utils.NewHeapControl(1),
		Filter:         nil,
		Getvector:      &novector,
		Getid:          &getid,
	}

	// Search for the point
	result := v.Search(sp)
	if len(result) == 0 {
		return fmt.Errorf("Point with point %v not found in collection %s", vector, collectionName)
	}

	// Delete the Point when the distance is 0
	if result[0].Distance == 0 {
		// Delete the node
		err := v.Collections[collectionName].DeleteVectorByID([]string{result[0].Id})
		if err != nil {
			return err
		}
		Logger.Log.Log("Point deleted from collection ", "INFO")
		return nil
	}

	return fmt.Errorf("Point with point %v not found in collection %s", vector, collectionName)
}

// Search searches for the nearest neighbours of the given target vector
func (v *Vdb) Search(sp *Utils.SearchParams) []*Utils.ResultSet {
	v.Collections[sp.CollectionName].Mut.RLock()
	defer v.Collections[sp.CollectionName].Mut.RUnlock()

	// if the collection is empty we return an empty slice
	if v.Collections[sp.CollectionName].DiagonalLength == 0 {
		return []*Utils.ResultSet{}
	}

	// Start the Queue Thread
	sp.Queue.StartThreads()

	// Add 1 to the queue waitgroup
	sp.Queue.AddToWaitGroup()

	su := Utils.NewSearchUnit(sp.Filter, 0.1)

	// search
	su.Search(v.Collections[sp.CollectionName].Nodes, sp.Target, sp.Queue, v.Collections[sp.CollectionName].DistanceFunc, v.Collections[sp.CollectionName].DimensionDiff)

	// Close the channel and wait for the Queue to finish
	sp.Queue.CloseChannel()

	// Here we have some time to do some other stuff
	filterRes := false
	if v.Collections[sp.CollectionName].DistanceFuncName == "euclid" && sp.MaxDistancePercent > 0 {
		filterRes = true
	}

	// Create the ResultSet
	results := make([]*Utils.ResultSet, sp.Queue.MaxResults)

	// Wait for the Queue to finish
	sp.Queue.Wg.Wait()

	// Get the nodes from the queue
	data := sp.Queue.GetNodes()
	dataLen := len(data)

	// If this collection uses euclid and we have a maxDistancePercent > 0 we need to filter the results
	if filterRes {
		// If a result is greater than maxDistancePercent * DiagonalLength we remove it
		for i := 0; i < dataLen; i++ {
			if data[i].Distance > sp.MaxDistancePercent*v.Collections[sp.CollectionName].DiagonalLength {
				data = append(data[:i], data[i+1:]...)
				i--
			}
		}
	}

	// only create a new slice if the dataLen is smaller than the MaxResults
	if dataLen < sp.Queue.MaxResults {
		results = make([]*Utils.ResultSet, dataLen)
	}

	// Get the Payloads back from the Memory Map
	for i := 0; i < dataLen; i++ {
		m, err := FileMapper.Mapper.ReadPayload(data[i].Node.Vector.PayloadStart, sp.CollectionName)
		if err != nil {
			Logger.Log.Log("Error reading payload: "+err.Error(), "ERROR")
			continue
		}
		// if getvector is true we also return the vector
		var vd *[]float64
		if *sp.Getvector {
			vd = &data[i].Node.Vector.Data
		}
		// if getid is true we also return the id
		var id string
		if *sp.Getid {
			id = data[i].Node.Vector.Id
		}
		results[i] = &Utils.ResultSet{Payload: m, Distance: data[i].Distance, Vector: vd, Id: id}
	}

	// Sort the results by distance, smallest first
	sort.Slice(results, func(i, j int) bool {
		return results[i].Distance < results[j].Distance
	})
	return results
}

func (v *Vdb) IndexSearch(sp *Utils.SearchParams) []*Utils.ResultSet {
	v.Collections[sp.CollectionName].Mut.RLock()
	defer v.Collections[sp.CollectionName].Mut.RUnlock()

	// if the collection is empty we return an empty slice
	if v.Collections[sp.CollectionName].DiagonalLength == 0 {
		return []*Utils.ResultSet{}
	}

	// Start the Queue Thread
	sp.Queue.StartThreads()

	// Add 1 to the queue waitgroup
	sp.Queue.AddToWaitGroup()

	su := Utils.NewSearchUnit(sp.Filter, 0.1)

	// search
	su.Search(v.Collections[sp.CollectionName].Indexes[sp.IndexName].Entries[sp.IndexValue], sp.Target, sp.Queue, v.Collections[sp.CollectionName].DistanceFunc, v.Collections[sp.CollectionName].DimensionDiff)

	// Close the channel and wait for the Queue to finish
	sp.Queue.CloseChannel()

	// Here we have some time to do some other stuff
	filterRes := false
	if v.Collections[sp.CollectionName].DistanceFuncName == "euclid" && sp.MaxDistancePercent > 0 {
		filterRes = true
	}

	// Create the ResultSet
	results := make([]*Utils.ResultSet, sp.Queue.MaxResults)

	// Wait for the Queue to finish
	sp.Queue.Wg.Wait()

	// Get the nodes from the queue
	data := sp.Queue.GetNodes()
	dataLen := len(data)

	// If this collection uses euclid and we have a maxDistancePercent > 0 we need to filter the results
	if filterRes {
		// If a result is greater than maxDistancePercent * DiagonalLength we remove it
		for i := 0; i < dataLen; i++ {
			if data[i].Distance > sp.MaxDistancePercent*v.Collections[sp.CollectionName].DiagonalLength {
				data = append(data[:i], data[i+1:]...)
				i--
			}
		}
	}

	// only create a new slice if the dataLen is smaller than the MaxResults
	if dataLen < sp.Queue.MaxResults {
		results = make([]*Utils.ResultSet, dataLen)
	}

	// Get the Payloads back from the Memory Map
	for i := 0; i < dataLen; i++ {
		m, err := FileMapper.Mapper.ReadPayload(data[i].Node.Vector.PayloadStart, sp.CollectionName)
		if err != nil {
			Logger.Log.Log("Error reading payload: "+err.Error(), "ERROR")
			continue
		}
		// if getvector is true we also return the vector
		var vd *[]float64
		if *sp.Getvector {
			vd = &data[i].Node.Vector.Data
		}
		// if getid is true we also return the id
		var id string
		if *sp.Getid {
			id = data[i].Node.Vector.Id
		}
		results[i] = &Utils.ResultSet{Payload: m, Distance: data[i].Distance, Vector: vd, Id: id}
	}

	// Sort the results by distance, smallest first
	sort.Slice(results, func(i, j int) bool {
		return results[i].Distance < results[j].Distance
	})
	return results
}

// DeleteWithFilter deletes vectors from the specified collection based on the provided filters.
// It performs the following steps:
// - Calls the SerialDelete method of the specified collection to delete the vectors matching the filters
// - Returns an error if the SerialDelete method returns an error, otherwise returns nil
func (v *Vdb) DeleteWithFilter(col string, filters []Filter.Filter) error {
	err := v.Collections[col].SerialDelete(filters)
	if err != nil {
		return err
	}
	return nil
}
