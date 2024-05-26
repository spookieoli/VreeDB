package Collection

import (
	"VreeDB/ArgsParser"
	"VreeDB/FileMapper"
	"VreeDB/Logger"
	"VreeDB/NN"
	"VreeDB/Node"
	"VreeDB/Svm"
	"VreeDB/Tsne"
	"VreeDB/Utils"
	"VreeDB/Vector"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Collection is a struct that holds a name, a pointer to a Node, a vector dimension and a distance function
type Collection struct {
	Name               string
	Nodes              *Node.Node
	VectorDimension    int
	DistanceFunc       func(*Vector.Vector, *Vector.Vector) (float64, error)
	Mut                sync.RWMutex
	Space              *map[string]*Vector.Vector
	MaxVector          *Vector.Vector
	MinVector          *Vector.Vector
	DimensionDiff      *Vector.Vector
	DiagonalLength     float64
	DistanceFuncName   string
	Classifiers        map[string]Classifier
	ClassifierReady    bool
	Indexes            map[string]*Index
	ClassifierTraining map[string]Classifier
	TSNE_Dimensions    []*Vector.Vector
	TSNE_Train         bool
}

// Interface for the Classifier
type Classifier interface {
	Predict([]float64) any
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
		MaxVector: ma, MinVector: mi, DimensionDiff: dd, DistanceFuncName: distanceFuncName, Classifiers: make(map[string]Classifier),
		ClassifierReady: false, ClassifierTraining: make(map[string]Classifier), Indexes: make(map[string]*Index)}
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
	pos, err := FileMapper.Mapper.SaveVectorWriter(vector.Id, vector.DataStart, vector.PayloadStart, c.Name)
	if err != nil {
		Logger.Log.Log("Error saving vector to file: " + err.Error())
		return err
	}
	// Save the position of the vector in the SaveVectorWriter
	vector.SaveVectorPosition = pos

	// Set classifier ready to true
	c.ClassifierReady = true

	// Check if there is an Index with a key from the Payload - if so add the vector to the Index
	go c.CheckIndex(vector)
	return nil
}

// Delete deletes a vector from the collection
// CAUTION - Delete will not remove the vectors Data from the DB Files .bin! - it will only flag the vector as deleted
// The vector will be removed from the KD-Tree and the Space and will not be loaded into the KD-Tree again
func (c *Collection) DeleteVectorByID(ids []string) error {
	c.Mut.Lock()
	defer c.Mut.Unlock()

	// Check if the vector exists
	for _, id := range ids {
		if _, ok := (*c.Space)[id]; !ok {
			return fmt.Errorf("Vector with ID %s does not exist", id)
		}
		// set the datasatrt in SaveVector to -1
		err := FileMapper.Mapper.SaveVectorWriteAt(-1, -1, c.Name, (*c.Space)[id].SaveVectorPosition)
		if err != nil {
			return err
		}
		// Set the datastart to -1
		(*c.Space)[id].DataStart = -1
		// Delete the vector from the Space
		delete(*c.Space, id)
	}
	// Rebuild the KD-Tree
	c.Rebuild()
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
	file, err := os.Create(*ArgsParser.Ap.FileStore + c.Name + ".json")
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

// AddClassifier adds a classifier to the Collection
func (c *Collection) AddClassifier(name string, typ string, loss string, architecture *[]NN.LayerJSON) error {
	c.Mut.Lock()
	defer c.Mut.Unlock()

	// Add the classifier to the Collection
	switch strings.ToLower(typ) {
	case "svm":
		c.Classifiers[name] = Svm.NewMultiClassSVM(name, c.Name)
	case "nn":
		// Check if architecture is nil
		if architecture == nil {
			return fmt.Errorf("no architecture given")
		}
		// create the network
		n, err := NN.NewNetwork(architecture, loss)
		if err != nil {
			return err
		}
		c.Classifiers[name] = n
	}
	return nil
}

// DeleteClassifier deletes a classifier from the Collection
func (c *Collection) DeleteClassifier(name string) error {
	c.Mut.Lock()
	defer c.Mut.Unlock()

	// Delete the classifier from the Collection
	delete(c.Classifiers, name)

	// Delete the Classifiers again to make sure it is not saved
	err := c.SaveClassifier()
	if err != nil {
		Logger.Log.Log(err.Error())
		return err
	}
	return nil
}

// DeleteAllClassifiers deletes all classifiers from the Collection
func (c *Collection) DeleteAllClassifiers() {
	c.Mut.Lock()
	defer c.Mut.Unlock()
	c.Classifiers = make(map[string]Classifier)
}

// TrainClassifier will train a given classifier
func (c *Collection) TrainClassifier(name string, degree int, lr float64, epochs int, batchsize int) error {
	c.Mut.RLock()
	defer c.Mut.RUnlock()

	// The classfier must exist
	if _, ok := c.Classifiers[name]; !ok {
		return fmt.Errorf("Classifier with name %s does not exists", name)
	}

	// Create a slice with all vectors in the collection
	var data []*Vector.Vector
	for _, v := range *c.Space {
		data = append(data, v)
	}
	// Train the classfifier
	switch v := c.Classifiers[name].(type) {
	case *Svm.MultiClassSVM:
		v.Train(data, epochs, lr, degree)
	case *NN.Network:
		// Neural Network
		x, y, err := v.CreateTrainData(data)
		if err != nil {
			return err
		}
		c.ClassifierTraining[name] = v
		v.Train(x, y, epochs, lr, batchsize)
	}

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
	file, err := os.Create(*ArgsParser.Ap.FileStore + c.Name + "_classifiers.gob")
	if err != nil {
		return err
	}
	defer file.Close()

	// Register the SVM structs
	gob.RegisterName("VreeDb/SVM.SVM", &Svm.SVM{})
	gob.RegisterName("VreeDb/SVM.MultiClassSVM", &Svm.MultiClassSVM{})
	gob.RegisterName("VreeDB/NN.Network", &NN.Network{})

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
	if _, err := os.Stat(*ArgsParser.Ap.FileStore + c.Name + "_classifiers.gob"); os.IsNotExist(err) {
		return nil
	}

	// Register the SVM structs
	gob.RegisterName("VreeDb/SVM.SVM", &Svm.SVM{})
	gob.RegisterName("VreeDb/SVM.MultiClassSVM", &Svm.MultiClassSVM{})
	gob.RegisterName("VreeDB/NN.Network", &NN.Network{})

	file, err := os.Open(*ArgsParser.Ap.FileStore + c.Name + "_classifiers.gob")
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

// ClassifierToSlice will return a slice of all Classifiernames in this collection
func (c *Collection) ClassifierToSlice() []string {
	c.Mut.RLock()
	defer c.Mut.RUnlock()
	var slice []string
	for k := range c.Classifiers {
		slice = append(slice, k)
	}
	return slice
}

// CreateIndex will create a new Index
func (c *Collection) CreateIndex(name, key string) error {
	c.Mut.Lock()
	defer c.Mut.Unlock()

	// Check if the Index already exists
	if _, ok := c.Indexes[name]; ok {
		return fmt.Errorf("Index with name %s already exists", name)
	}

	// Create the index
	index, err := NewIndex(key, c.Space, c.Name)
	if err != nil {
		return err
	}
	// Add the index to the Collection
	c.Indexes[name] = index
	return nil
}

// CheckIndex Check if a specific Index exists
func (c *Collection) CheckIndex(vector *Vector.Vector) error {
	// First check if there is an Index
	if len(c.Indexes) == 0 {
		return nil
	}

	// the result slice
	var result []string

	// Get the Payload from the hdd
	payload, err := FileMapper.Mapper.ReadPayload(vector.PayloadStart, c.Name)
	if err != nil {
		return err
	}

	// check if an Index Key is in the Payload
	for k := range c.Indexes {
		c.Indexes[k].mut.RLock()
		if _, ok := (*payload)[c.Indexes[k].Key]; ok {
			result = append(result, k)
		}
		c.Indexes[k].mut.RUnlock()
	}

	// If there is a result, add the vector to the Index
	if len(result) > 0 {
		err = c.addVectorToIndexes(result, vector)
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveIndexes saves the indexes of the collection to a file in the file store directory.
func (c *Collection) SaveIndexes() error {
	file, err := os.Create(*ArgsParser.Ap.FileStore + c.Name + "_indexes.gob")
	if err != nil {
		return err
	}
	defer file.Close()

	// make empty string slice
	indexes := make([]string, 0)

	// Get all the indexnames from  then c.Indexes map
	for indexName := range c.Indexes {
		indexes = append(indexes, indexName)
	}

	// Create Encoder
	enc := gob.NewEncoder(file)

	// The indexname is the index Field in the payload
	err = enc.Encode(indexes)
	if err != nil {
		return err
	}
	return nil
}

// RebuildIndex index will rebuild the indexes
func (c *Collection) RebuildIndex() error {
	// Open the file
	file, err := os.Open(*ArgsParser.Ap.FileStore + c.Name + "_indexes.gob")
	if err != nil {
		return err
	}
	defer file.Close()

	// Make empty string slice
	var indexes []string

	// Create Decoder
	dec := gob.NewDecoder(file)
	dec.Decode(&indexes)

	// Create c.Indexes map
	c.Indexes = make(map[string]*Index)

	// Now loop over the indexes and recreate them
	for _, indexName := range indexes {
		index, err := NewIndex(indexName, c.Space, c.Name)
		if err != nil {
			return err
		}
		c.Indexes[indexName] = index
	}
	return nil
}

// addVectorToIndexes to Add a vector to the Index(es)
func (c *Collection) addVectorToIndexes(keys []string, vector *Vector.Vector) error {
	c.Mut.RLock()
	defer c.Mut.RUnlock()

	// Add the vector to the Indexes
	for _, k := range keys {
		if index, ok := c.Indexes[k]; ok {
			c.Indexes[k].mut.Lock()
			err := index.AddToIndex(vector)
			c.Indexes[k].mut.Unlock()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetClassifierTrainingPhase will return the training phase of a classifier
func (c *Collection) GetClassifierTrainingPhase(name string) (*NN.TrainProgress, error) {

	// Check if the classifier exists
	if _, ok := c.ClassifierTraining[name]; !ok {
		return nil, fmt.Errorf("Classifier with name %s does not train", name)
	}

	// We have Neural Network and SVM
	switch v := c.ClassifierTraining[name].(type) {
	case *NN.Network:
		phase := v.GetTrainPhase()
		return &phase[len(phase)-1], nil
	default:
		return nil, fmt.Errorf("Classifier with name %s has no progress yet", name) // TODO: add for SVM
	}
}

// CreateTSNE creates a t-SNE object and performs t-SNE dimensionality reduction
// on the vectors in the collection's space. It takes the dimensions of the output
// space, the number of iterations, and the learning rate as input parameters.
// It returns an error if there was an issue performing the t-SNE dimensionality reduction.
func (c *Collection) CreateTSNE(dimensions, iterations int, learningrate float64) error {
	// Lock for reading
	c.Mut.RLock()

	// If TSNE_Train is true return
	if c.TSNE_Train {
		c.Mut.RUnlock()
		return fmt.Errorf("training already in progress")
	}

	if c.TSNE_Dimensions != nil {
		fmt.Errorf("TSNE already created")
	}

	// Set Train to true
	c.TSNE_Train = true

	// Create tsne object
	tsne := Tsne.NewTSNE(learningrate, iterations, dimensions, c.Name)
	// Get all the vectors in c.Space as slice
	data := make([]*Vector.Vector, 0, len(*c.Space))

	// Create slice from Map
	for _, v := range *c.Space {
		data = append(data, v)
	}

	// perform the training
	dim, err := tsne.PerformTSNE(data)
	if err != nil {
		return fmt.Errorf("Error creating TSNE: %v", err)
	}
	c.Mut.RUnlock()

	// Lock fpr writing
	c.Mut.Lock()
	c.TSNE_Dimensions = dim
	c.TSNE_Train = false
	c.Mut.Unlock()

	return nil
}

// GetTSNEDimensions returns the TSNE_Dimensions slice of Vector pointers from the collection.
func (c *Collection) GetTSNEDimensions() []*Vector.Vector {
	return c.TSNE_Dimensions
}
