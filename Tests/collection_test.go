// collection_test.go
package Collection

import (
	"VreeDB/Collection"
	"VreeDB/Vector"
	"testing"
)

func TestNewCollection(t *testing.T) {
	// Creating a new collection
	collection := Collection.NewCollection("test_collection", 3, "euclid")

	// Check if the collection was created successfully
	if collection.Name != "test_collection" {
		t.Errorf("Expected collection name to be 'test_collection', got %s", collection.Name)
	}
	if collection.VectorDimension != 3 {
		t.Errorf("Expected vector dimension to be 3, got %d", collection.VectorDimension)
	}
	if collection.DistanceFuncName != "euclid" {
		t.Errorf("Expected distance function name to be 'euclid', got %s", collection.DistanceFuncName)
	}
	if collection.ClassifierReady != false {
		t.Errorf("Expected classifier ready to be false, got %t", collection.ClassifierReady)
	}
}

func TestInsert(t *testing.T) {
	// Creating a new collection
	collection := Collection.NewCollection("test_collection", 3, "euclid")

	// Creating a vector to insert
	vector := &Vector.Vector{Id: "v1", Data: []float64{1, 2, 3}, Length: 3}

	// Inserting the vector
	err := collection.Insert(vector)
	if err != nil {
		t.Errorf("Inserting vector failed: %s", err)
	}

	// Check if the vector was inserted
	if _, ok := (*collection.Space)["v1"]; !ok {
		t.Errorf("Expected vector with ID 'v1' to be in the collection")
	}
}

func TestInsertDifferentDimension(t *testing.T) {
	// Creating a new collection
	collection := Collection.NewCollection("test_collection", 3, "euclid")

	// Creating a vector with a different dimension
	vector := &Vector.Vector{Id: "v1", Data: []float64{1, 2}, Length: 2}

	// Inserting the vector
	err := collection.Insert(vector)
	if err == nil {
		t.Errorf("Expected error when inserting vector with different dimension, got nil")
	}
}

func TestDeleteVectorByID(t *testing.T) {
	// Creating a new collection
	collection := Collection.NewCollection("test_collection", 3, "euclid")

	// Creating a vector to insert
	vector := &Vector.Vector{Id: "v1", Data: []float64{1, 2, 3}, Length: 3}

	// Inserting the vector
	err := collection.Insert(vector)
	if err != nil {
		t.Errorf("Inserting vector failed: %s", err)
	}

	// Deleting the vector
	err = collection.DeleteVectorByID([]string{"v1"})
	if err != nil {
		t.Errorf("Deleting vector failed: %s", err)
	}

	// Check if the vector was deleted
	if _, ok := (*collection.Space)["v1"]; ok {
		t.Errorf("Expected vector with ID 'v1' to be deleted")
	}
}
