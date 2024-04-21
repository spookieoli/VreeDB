package FileMapper

import (
	"VreeDB/Logger"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"io"
	"math"
	"os"
	"sync"
	"syscall"
)

type SaveVector struct {
	VectorID     string
	DataStart    int64
	PayloadStart int64
}

type FileMapper struct {
	CollectionNames []string
	FileName        map[string]string
	File            map[string]*os.File
	Mut             map[string]*sync.RWMutex
	MappedData      map[string][]byte
	Mapped          map[string]bool
}

// the filemapper is a singleton
var Mapper *FileMapper

// NewFileMapper initializes the FileMapper (Mapper)
func init() {
	// Create Mapper singleton
	Mapper = &FileMapper{}
	// init the maps
	Mapper.FileName = make(map[string]string)
	Mapper.File = make(map[string]*os.File)
	Mapper.Mut = make(map[string]*sync.RWMutex)
	Mapper.MappedData = make(map[string][]byte)
	Mapper.Mapped = make(map[string]bool)
}

func (f *FileMapper) Start(collections []string) {
	// Loop over all Collections
	for _, name := range collections {
		// check if data.bin file exists
		_, err := os.Stat("collections/" + name + ".bin")
		if err != nil {
			// if not create it
			file, err := os.Create("collections/" + name + ".bin")
			if err != nil {
				Logger.Log.Log("Error creating file: " + err.Error())
				panic(err)
			}
			file.Close()
		}
		Mapper.FileName[name] = "collections/" + name + ".bin"
		Mapper.Mut[name] = &sync.RWMutex{}
		Mapper.CollectionNames = append(Mapper.CollectionNames, name)
		Mapper.MapFile(name)
	}
}

// WriteVector will write data to the file
func (f *FileMapper) WriteVector(arr []float64, collection string) (int64, int, error) {
	// Lock the file for writing
	f.Mut[collection].Lock()
	defer f.Mut[collection].Unlock()
	// Unmap the file from memory
	f.Unmap(collection)
	// open the file again for writing und append
	file, err := os.OpenFile(f.FileName[collection], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		Logger.Log.Log("Error opening file: " + err.Error())
		// Here we panic because we can't continue without the file
		panic(err)
	}
	f.File[collection] = file
	defer f.File[collection].Close()

	// Get the Start position
	start, err := f.File[collection].Seek(0, io.SeekEnd)
	if err != nil {
		Logger.Log.Log("Error seeking to end of file: " + err.Error())
		// Here we panic because we can't continue without the start position
		panic(err)
	}

	// Array in die Datei schreiben
	for _, value := range arr {
		err := binary.Write(f.File[collection], binary.LittleEndian, value)
		if err != nil {
			panic(err)
		}
	}

	// Close the file and reopen it
	err = f.File[collection].Close()

	// Map the file again
	f.MapFile(collection)

	// Return the start position and the length of the array
	return start, len(arr), err
}

// ReadVector will read data from the file
func (f *FileMapper) ReadVector(start int64, length int, collection string) *[]float64 {
	// Lock the file for reading
	f.Mut[collection].RLock()
	defer f.Mut[collection].RUnlock()
	// if not mapped we map it
	if !f.Mapped[collection] {
		f.MapFile(collection)
	}
	// Create the array
	arr := make([]float64, length)
	// Check if the slice is not empty
	if len(f.MappedData[collection]) > 0 {
		// Read the data from the file
		for i := 0; i < length; i++ {
			arr[i] = math.Float64frombits(binary.LittleEndian.Uint64(f.MappedData[collection][start+int64(i)*8 : start+int64(i)*8+8]))
		}
	}
	return &arr
}

// WritePayload will write the payload to the file
func (f *FileMapper) WritePayload(payload *map[string]interface{}, collection string) (int64, error) {
	// Lock the file for writing
	f.Mut[collection].Lock()
	defer f.Mut[collection].Unlock()

	// Unmap the file from memory
	f.Unmap(collection)

	// Map in einen Byte-Slice serialisieren
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	// Register types
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})

	// Encode the payload
	err := enc.Encode(payload)
	if err != nil {
		Logger.Log.Log("Error encoding payload: " + err.Error())
		return 0, err
	}
	encodedBytes := buf.Bytes()

	// open the file for writing und append
	file, err := os.OpenFile(f.FileName[collection], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Here we panic because we can't continue without the file
		panic(err)
	}
	f.File[collection] = file
	defer f.File[collection].Close()

	// Sicherstellen, dass die Datei groß genug ist
	fileInfo, err := f.File[collection].Stat()
	if err != nil {
		Logger.Log.Log("Error getting file info: " + err.Error())
		return 0, err
	}

	// Get the actual File size
	fileSize := fileInfo.Size()

	// We need to get the required size
	requiredSize := fileSize + int64(len(encodedBytes))
	if requiredSize >= fileSize {
		err = file.Truncate(requiredSize)
		if err != nil {
			Logger.Log.Log("Error truncating file: " + err.Error())
			return 0, err
		}
	}

	// Zur letzten Position in der Datei springen
	offset, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		Logger.Log.Log("Error seeking to end of file: " + err.Error())
		return 0, err
	}

	// Serialisierte Daten in die Datei schreiben
	_, err = file.Write(encodedBytes)
	if err != nil {
		Logger.Log.Log("Error writing to file: " + err.Error())
		return 0, err
	}
	// Map the file again
	f.MapFile(collection)
	// Return the offset
	return offset, nil
}

// ReadPayload will read the payload from the file
func (f *FileMapper) ReadPayload(offset int64, collection string) (*map[string]interface{}, error) {
	// Lock the file for reading
	f.Mut[collection].RLock()
	defer f.Mut[collection].RUnlock()
	// Bytes-Slice ab der gegebenen Position erstellen
	data := f.MappedData[collection][offset:]

	// Gob register types
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})

	// Daten deserialisieren
	var m map[string]interface{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&m)
	if err != nil {
		Logger.Log.Log("Error decoding payload: " + err.Error())
		return nil, err
	}
	return &m, nil
}

// MapFile will map the file to memory
func (f *FileMapper) MapFile(collection string) {
	// Open the file
	file, err := os.Open(f.FileName[collection])
	f.File[collection] = file
	// get the file info
	fileInfo, err := f.File[collection].Stat()
	if err != nil {
		Logger.Log.Log("Error getting file info: " + err.Error())
		// We panic here because we can't continue without the file info
		panic(err)
	}

	// if the file is not empty we map it to memory
	if fileInfo.Size() != 0 {
		mappedData, err := syscall.Mmap(int(f.File[collection].Fd()), 0, int(fileInfo.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
		if err != nil {
			Logger.Log.Log("Error mapping file: " + err.Error())
			// We panic here because we can't continue without the mapped data
			panic(err)
		}
		f.MappedData[collection] = mappedData
		f.Mapped[collection] = true
	} else {
		f.MappedData[collection] = nil
	}
}

// Unmap will unmap the file from memory
func (f *FileMapper) Unmap(collection string) {
	// Check if the file is mapped
	if f.MappedData[collection] != nil && f.Mapped[collection] {
		err := syscall.Munmap(f.MappedData[collection])
		if err != nil {
			// We panic here because we can't continue without the mapped data
			Logger.Log.Log("Error unmapping file: " + err.Error())
			panic(err)
		}
		f.Mapped[collection] = false
		// Close the file
		err = f.File[collection].Close()
		if err != nil {
			// We panic here because we can't continue without the file
			panic(err)
		}
	}
}

// AddCollection adds a collection to the FileMapper
func (f *FileMapper) AddCollection(collection string) {
	// Check if data.cin file exists
	_, err := os.Stat("collections/" + collection + ".bin")
	if err != nil {
		// if not create it
		file, err := os.Create("collections/" + collection + ".bin")
		if err != nil {
			panic(err)
		}
		file.Close()
	}
	f.FileName[collection] = "collections/" + collection + ".bin"
	f.Mut[collection] = &sync.RWMutex{}
	f.CollectionNames = append(f.CollectionNames, collection)
	f.MapFile(collection)
}

// DelCollection deletes a collection from the FileMapper
func (f *FileMapper) DelCollection(collection string) {
	// Unmap the file from memory
	f.Unmap(collection)
	// Delete the file
	err := os.Remove(f.FileName[collection])
	if err != nil {
		// We panic here because we can't continue without the file
		panic(err)
	}
	// if meta file exists delete it
	_, err = os.Stat("collections/" + collection + "_meta.bin")
	if err == nil {
		err = os.Remove("collections/" + collection + "_meta.bin")
		if err != nil {
			Logger.Log.Log("Error deleting meta file: " + err.Error())
		}
	}
	// Remove the collection.json if exists
	_, err = os.Stat("collections/" + collection + ".json")
	if err == nil {
		err = os.Remove("collections/" + collection + ".json")
		if err != nil {
			Logger.Log.Log("Error deleting collection config file: " + err.Error())
		}
	}
	// Remove the collection from the CollectionNames
	for i, col := range f.CollectionNames {
		if col == collection {
			f.CollectionNames = append(f.CollectionNames[:i], f.CollectionNames[i+1:]...)
		}
	}
}

// SaveVectorWriter will write the vector.ID, vector.DataStart, vector.PayloadStart to the file system
func (w *FileMapper) SaveVectorWriter(id string, datastart, payloadstart int64, collection string) error {
	// Lock the Wal
	w.Mut[collection].Lock()
	defer w.Mut[collection].Unlock()

	// Open the file "collection"_meta.bin
	file, err := os.OpenFile("collections/"+collection+"_meta.bin", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		Logger.Log.Log("Error opening meta.json file: " + err.Error())
		return err
	}
	defer file.Close()

	// Create the SaveVector
	sv := SaveVector{VectorID: id, DataStart: datastart, PayloadStart: payloadstart}

	// use json to encode the SaveVector
	encoder := json.NewEncoder(file)
	err = encoder.Encode(sv)
	if err != nil {
		Logger.Log.Log("Error encoding SaveVector: " + err.Error())
		return err
	}
	return nil
}

// SaveVectorRead will read the vector.ID, vector.DataStart, vector.PayloadStart from the file system and returns a map of vectors
func (w *FileMapper) SaveVectorRead(collection string) (*map[string]SaveVector, error) {
	// Lock the Wal - we use a write lock because here will be no memory mapped file
	w.Mut[collection].Lock()
	defer w.Mut[collection].Unlock()

	// Create the map
	vectors := make(map[string]SaveVector)
	// Open the file "collection"_meta.bin if existing
	_, err := os.Stat("collections/" + collection + "_meta.bin")
	if err == nil {
		file, err := os.Open("collections/" + collection + "_meta.bin")
		if err != nil {
			Logger.Log.Log("Error opening meta.json file: " + err.Error())
			return nil, err
		}
		defer file.Close()

		// Use json to decode the SaveVector
		decoder := json.NewDecoder(file)
		for {
			var sv SaveVector
			if err := decoder.Decode(&sv); err == io.EOF {
				break
			} else if err != nil {
				Logger.Log.Log("Error decoding SaveVector: " + err.Error())
				return nil, err
			}
			vectors[sv.VectorID] = sv
		}
	}
	return &vectors, nil
}

// SaveVectorDelete will delete the vector.ID, vector.DataStart, vector.PayloadStart from the file system // TODO: This is shit
func (w *FileMapper) SaveVectorDelete(id string, collection string) error {
	// Lock the mutex for reading
	w.Mut[collection].Lock()
	defer w.Mut[collection].Unlock()

	// Open the file "collection"_meta.bin
	file, err := os.Open("collections/" + collection + "_meta.bin")
	if err != nil {
		Logger.Log.Log("Error opening meta.json file: " + err.Error())
		return err
	}

	// Decode all JSON objects into a slice of SaveVector
	var vectors []SaveVector
	decoder := json.NewDecoder(file)
	for {
		var sv SaveVector
		if err := decoder.Decode(&sv); err == io.EOF {
			break
		} else if err != nil {
			Logger.Log.Log("Error decoding SaveVector: " + err.Error())
			return err
		}
		vectors = append(vectors, sv)
	}

	file.Close()
	w.Mut[collection].Unlock()

	// Iterate over the slice and if the VectorID matches the given ID, remove that element from the slice
	for i, vector := range vectors {
		if vector.VectorID == id {
			vectors = append(vectors[:i], vectors[i+1:]...)
			break
		}
	}

	// Lock the mutex for writing
	w.Mut[collection].Lock()

	// Open the file in write mode, this will clear the file content
	file, err = os.OpenFile("collections/"+collection+"_meta.bin", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		Logger.Log.Log("Error opening meta.json file: " + err.Error())
		return err
	}
	defer file.Close()

	// Encode the modified slice of SaveVector back into the file
	encoder := json.NewEncoder(file)
	for _, vector := range vectors {
		err = encoder.Encode(vector)
		if err != nil {
			Logger.Log.Log("Error encoding SaveVector: " + err.Error())
			return err
		}
	}

	return nil
}
