package Collection

import (
	"VectoriaDB/FileMapper"
	"VectoriaDB/Logger"
	"VectoriaDB/Node"
	"VectoriaDB/Svm"
	"VectoriaDB/Utils"
	"VectoriaDB/Vector"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Collection is a struct that holds a name, a pointer to a Node, a vector dimension and a distance function
type Collection struct {
	Name             string
	Nodes            *Node.Node
	VectorDimension  int
	DistanceFunc     func(*Vector.Vector, *Vector.Vector) (float64, error)
	Mut              sync.RWMutex
	Space            *map[string]*Vector.Vector
	MaxVector        *Vector.Vector
	MinVector        *Vector.Vector
	DimensionDiff    *Vector.Vector
	DiagonalLength   float64
	DistanceFuncName string
	Classifiers      map[string]*Svm.MultiClassSVM
}

// NewCollection returns a new Collection
func NewCollection(name string, vectorDimension int, distanceFuncName string) *Collection {
	// Vars
	var distanceFunc func(*Vector.Vector, *Vector.Vector) (float64, error)

	// Create the max,min and diff vectors
	ma := &Vector.Vector{Data: make([]float64, vectorDimension), Length: vectorDimension}
	mi := &Vector.Vector{Data: make([]float64, vectorDimension), Length: vectorDimension}
	dd := &Vector.Vector{Data: make([]float64, vectorDimension), Length: vectorDimension}

	if strings.ToLower(distanceFuncName) == "euclid" {
		distanceFunc = Utils.Utils.EuclideanDistance
	} else {
		distanceFunc = Utils.Utils.CosineDistance
	}

	return &Collection{Name: name, VectorDimension: vectorDimension, Nodes: &Node.Node{Depth: 0}, DistanceFunc: distanceFunc, Space: &map[string]*Vector.Vector{},
		MaxVector: ma, MinVector: mi, DimensionDiff: dd, DistanceFuncName: distanceFuncName, Classifiers: make(map[string]*Svm.MultiClassSVM)}
}

// Insert inserts a vector into the collection
func (c *Collection) Insert(vector *Vector.Vector) error {
	c.Mut.Lock()
	defer c.Mut.Unlock()
	if vector.Length != c.VectorDimension {
		return fmt.Errorf("Vector length is %d, expected %d", vector.Length, c.VectorDimension)
	} else if c.CheckID(vector.Id) {
		return fmt.Errorf("Vector with ID %s already exists", vector.Id)
	}

	// Insert the vector into the KD-Tree
	c.Nodes.Insert(vector)

	// Set diagonal Space
	c.SetDiaSpace(vector)

	// add it to the Space
	(*c.Space)[vector.Id] = vector

	// Save the Collection to the FS
	err := FileMapper.Mapper.SaveVectorWriter(vector.Id, vector.DataStart, vector.PayloadStart, c.Name)
	if err != nil {
		Logger.Log.Log("Error saving vector to file: " + err.Error())
		return err
	}
	return nil
}

// Delete deletes a vector from the collection
// CAUTION - Delete will not remove the vectors Data from the DB Files .bin!
func (c *Collection) Delete(id string) error {
	c.Mut.Lock()
	defer c.Mut.Unlock()
	if _, ok := (*c.Space)[id]; !ok {
		return fmt.Errorf("Vector with ID %s does not exist", id)
	}
	delete(*c.Space, id)
	// Rebuild the KD-Tree
	c.Rebuild()
	// Delete the vector from the FileMapper
	err := FileMapper.Mapper.SaveVectorDelete(id, c.Name)
	if err != nil {
		Logger.Log.Log("Error deleting vector from file: " + err.Error())
		return err
	}
	return nil
}

// SetDIaSpace will set the diagonal space of the Collection
func (c *Collection) SetDiaSpace(vector *Vector.Vector) {
	// Update the max and min vectors
	wg := sync.WaitGroup{}
	wg.Add(2)
	go Utils.Utils.GetMaxDimension(c.MaxVector, vector, &wg)
	go Utils.Utils.GetMinDimension(c.MinVector, vector, &wg)
	wg.Wait()

	// Calculate the difference between the max and min vectors
	Utils.Utils.CalculateDimensionDiff(c.VectorDimension, c.DimensionDiff, c.MaxVector, c.MinVector)

	// Calculate the DiogonalLength of the Collection
	Utils.Utils.CalculateDiogonalLength(&c.DiagonalLength, c.VectorDimension, c.DimensionDiff)
}

// GetNodeCount returns the number of points in the Collection
func (c *Collection) GetNodeCount() int64 {
	return int64(len(*c.Space))
}

// WriteConfig will write the Collection config to the file system
func (c *Collection) WriteConfig() error {
	c.Mut.RLock()
	defer c.Mut.RUnlock()

	// We need to save the CollectionConfig, this will be done via a struct that saves the important configs of the Collection
	file, err := os.Create("collections/" + c.Name + ".json")
	if err != nil {
		return err
	}
	// Save the struct to it
	err = json.NewEncoder(file).Encode(Utils.CollectionConfig{
		Name:             c.Name,
		VectorDimension:  c.VectorDimension,
		DistanceFuncName: c.DistanceFuncName,
		DiagonalLength:   c.DiagonalLength,
	})
	if err != nil {
		return err
	}
	return nil
}

// Recreate will recreate the KD-Tree from the SpaceMap
func (c *Collection) Recreate() {
	c.Mut.Lock()
	defer c.Mut.Unlock()
	c.Nodes = &Node.Node{Depth: 0}
	for _, v := range *c.Space {
		v.RecreateMut() // This needed to recreate the vector mut, it will not be saved in the gob file
		c.Nodes.Insert(v)
		c.SetDiaSpace(v)
	}
}

// Rebuild is like Recreate but it does not use the Mut and will not use the RecreateMut function
func (c *Collection) Rebuild() {
	// Mut already blocked in Delete
	c.Nodes = &Node.Node{Depth: 0}
	for _, v := range *c.Space {
		c.Nodes.Insert(v)
		c.SetDiaSpace(v)
	}
}

// CheckID will Check if the given ID is already in the Collection Space
func (c *Collection) CheckID(id string) bool {
	_, ok := (*c.Space)[id]
	return ok
}

// TrainClassifier will train a given classifier
func (c *Collection) TrainClassifier(name string, degree int, cValue float64, epochs int) error {
	c.Mut.RLock()
	defer c.Mut.RUnlock()

	// The classfier must exist
	if _, ok := c.Classifiers[name]; !ok {
		return fmt.Errorf("Classifier with name %s does not exists", name)
	}

	// Create a slice with alle vectors in the collection
	var data []*Vector.Vector
	for _, v := range *c.Space {
		data = append(data, v)
	}
	// Train the classfifier
	c.Classifiers[name].Train(data, epochs, cValue, degree)

	// Save the classifier
	err := c.SaveClassifier()
	if err != nil {
		Logger.Log.Log("Error saving classifier: " + err.Error())
		return err
	}
	return nil
}

// SaveClassifier will save all classifier to the file system using gob
func (c *Collection) SaveClassifier() error {
	c.Mut.RLock()
	defer c.Mut.RUnlock()

	// Open the file
	file, err := os.Create("collections/" + c.Name + "_classifiers.gob")
	if err != nil {
		return err
	}
	defer file.Close()

	// Register the SVM structs
	gob.Register(Svm.SVM{})
	gob.Register(Svm.MultiClassSVM{})

	// Save the classifiers
	err = gob.NewEncoder(file).Encode(c.Classifiers)
	if err != nil {
		return err
	}

	Logger.Log.Log("Successfully saved classifier")

	return nil
}

// ReadClassifiers will read all classifiers from the file system using gob
func (c *Collection) ReadClassifiers() error {
	c.Mut.Lock()
	defer c.Mut.Unlock()

	// Open the file
	if _, err := os.Stat("collections/" + c.Name + "_classifiers.gob"); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open("collections/" + c.Name + "_classifiers.gob")
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode the file
	err = gob.NewDecoder(file).Decode(&c.Classifiers)
	if err != nil {
		return err
	}
	return nil
}
