package Boot

import (
	"VreeDB/Collection"
	"VreeDB/FileMapper"
	"VreeDB/Logger"
	"VreeDB/Utils"
	"VreeDB/Vector"
	"encoding/json"
	"os"
	"strings"
)

// BootUp will boot the application, restore all existing collections and will check for integrity
type BootUp struct {
}

// NewBootUp returns a new BootUp
func NewBootUp() *BootUp {
	return &BootUp{}
}

// Boot boots the application
func (b *BootUp) Boot() map[string]*Collection.Collection {
	return b.RestoreCollections()
}

// RestoreCollections restores the collection
func (b *BootUp) RestoreCollections() map[string]*Collection.Collection {
	collections := make(map[string]*Collection.Collection)

	// open directory collections ans list all *json files
	entries, err := os.ReadDir("collections")
	if err != nil {
		panic(err) // Panic if there is an error - without the collection directory the application cannot work
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") {
			// Open the file
			file, err := os.Open("collections/" + entry.Name())
			if err != nil {
				Logger.Log.Log("Error opening file: "+err.Error(), "ERROR")
				continue
			}
			// create a new CollectionConfig
			c := Utils.CollectionConfig{}
			// Decode the file
			err = json.NewDecoder(file).Decode(&c)
			if err != nil {
				Logger.Log.Log("Error decoding file: "+err.Error(), "ERROR")
				continue
			}
			// Enter the collection into the map
			collections[strings.Split(entry.Name(), ".")[0]] = Collection.NewCollection(c.Name, c.VectorDimension, c.DistanceFuncName, c.Aces)

			// Set the DiagonalLength
			collections[c.Name].DiagonalLength = c.DiagonalLength

			// Create the collection in the Filemapper
			FileMapper.Mapper.AddCollection(c.Name)

			// Restore vectors (if any)
			vectors, err := b.RestoreVectors(c.Name, collections[c.Name].VectorDimension)
			if err != nil {
				Logger.Log.Log("Error restoring vectors: "+err.Error(), "ERROR")
				continue
			}
			// Set the vectors
			collections[c.Name].Space = vectors

			// Recreate the KD-Tree
			collections[c.Name].Recreate()

			// Set ClassifierReady
			collections[c.Name].ClassifierReady = true

			// Restore Indexes
			err = collections[c.Name].RebuildIndex()
			Logger.Log.Log("Collection "+c.Name+" indexes restored", "INFO")

			// recreate the SVMs (if present)
			err = collections[c.Name].ReadClassifiers()
			if err != nil {
				Logger.Log.Log("Error reading SVMs: "+err.Error(), "ERROR")
			}
			Logger.Log.Log("Collection "+c.Name+" classifiers restored", "INFO")

			// Close the file
			file.Close()
			Logger.Log.Log("Collection "+c.Name+" restored", "INFO")

		}
	}
	// Log that we are done
	Logger.Log.Log("VreeDB Bootup complete", "INFO")
	return collections
}

// RestoreVectors restores the vectors of a given collection.
// It reads the saved vectors from a file, skips deleted vectors,
// and creates new Vector instances for each vector.
// The restored vectors are then returned as a map where the key is
// the vector ID and the value is the corresponding Vector instance.
// The restored vectors are also unindexed and their properties,
// such as Collection, DataStart, PayloadStart, Length, and SaveVectorPosition,
// are set based on the read data.
func (b *BootUp) RestoreVectors(collection string, dimension int) (*map[string]*Vector.Vector, error) {
	vectors := make(map[string]*Vector.Vector)
	m, err := FileMapper.Mapper.SaveVectorRead(collection)
	if err != nil {
		Logger.Log.Log("Error reading SaveVector: "+err.Error(), "ERROR")
		return nil, err
	}

	for _, v := range *m {
		// Dont restore deleted vectors
		if v.DataStart < 0 {
			continue
		}
		vectors[v.VectorID] = Vector.NewVector(v.VectorID, nil, nil, "")
		vectors[v.VectorID].Collection = collection
		vectors[v.VectorID].DataStart = v.DataStart
		vectors[v.VectorID].PayloadStart = v.PayloadStart
		vectors[v.VectorID].Length = dimension
		vectors[v.VectorID].SaveVectorPosition = v.SaveVectorPosition
		vectors[v.VectorID].Unindex()
	}
	return &vectors, nil
}
