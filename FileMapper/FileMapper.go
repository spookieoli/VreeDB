package FileMapper

import (
	"VreeDB/ArgsParser"
	"VreeDB/Logger"
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
)

type SaveVector struct {
	VectorID           string
	DataStart          int64
	PayloadStart       int64
	SaveVectorPosition int64
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
		_, err := os.Stat(*ArgsParser.Ap.FileStore + name + ".bin")
		if err != nil {
			// if not create it
			file, err := os.Create(*ArgsParser.Ap.FileStore + name + ".bin")
			if err != nil {
				Logger.Log.Log("Error creating file: "+err.Error(), "ERROR")
				panic(err)
			}
			file.Close()
		}
		Mapper.FileName[name] = *ArgsParser.Ap.FileStore + name + ".bin"
		Mapper.Mut[name] = &sync.RWMutex{}
		Mapper.CollectionNames = append(Mapper.CollectionNames, name)
		Mapper.MapFile(name)
	}
}

// GetCompressedBuffer encodes the given float64 array,
// compresses it using GzipWriter, and returns the resulting compressed buffer.
// It returns an error if there is an error encoding the array or closing GzipWriter.
func (f *FileMapper) GetCompressedBuffer(arr []float64) (bytes.Buffer, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	err := gob.NewEncoder(gz).Encode(arr)
	if err != nil {
		Logger.Log.Log("Error encoding array: "+err.Error(), "ERROR")
		// Here we panic because we can't continue without encoding the array
		return buf, fmt.Errorf("cant encode data")
	}
	err = gz.Close()
	if err != nil {
		Logger.Log.Log("Error closing GzipWriter: "+err.Error(), "ERROR")
		return buf, fmt.Errorf("Error closing GzipWriter: " + err.Error())
	}
	return buf, nil
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
		Logger.Log.Log("Error opening file: "+err.Error(), "ERROR")
		// Here we panic because we can't continue without the file
		panic(err)
	}
	f.File[collection] = file
	defer f.File[collection].Close()

	// Get the Start position
	start, err := f.File[collection].Seek(0, io.SeekEnd)
	if err != nil {
		Logger.Log.Log("Error seeking to end of file: "+err.Error(), "ERROR")
		// Here we panic because we can't continue without the start position
		panic(err)
	}

	// Compress []float64
	buf, err := f.GetCompressedBuffer(arr)
	if err != nil {
		return 0, 0, err
	}

	// Write the data to the file
	n, err := f.File[collection].Write(buf.Bytes())
	if err != nil {
		Logger.Log.Log("Error writing to file: "+err.Error(), "ERROR")
		// Here we panic because we can't continue without writing to the file
		panic(err)
	}

	// Close the file and reopen it
	err = f.File[collection].Close()

	// Map the file again
	f.MapFile(collection)

	// Return the start position and the length of the array
	return start, n, err
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
		// Extract the gzip compressed data
		compressedData := f.MappedData[collection][start:]
		buf := bytes.NewBuffer(compressedData)
		gz, err := gzip.NewReader(buf)
		if err != nil {
			panic("cannot create gzip reader")
		}
		dec := gob.NewDecoder(gz)
		err = dec.Decode(&arr)
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
		Logger.Log.Log("Error encoding payload: "+err.Error(), "ERROR")
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
		Logger.Log.Log("Error getting file info: "+err.Error(), "ERROR")
		return 0, err
	}

	// Get the actual File size
	fileSize := fileInfo.Size()

	// We need to get the required size
	requiredSize := fileSize + int64(len(encodedBytes))
	if requiredSize >= fileSize {
		err = file.Truncate(requiredSize)
		if err != nil {
			Logger.Log.Log("Error truncating file: "+err.Error(), "ERROR")
			return 0, err
		}
	}

	// Zur letzten Position in der Datei springen
	offset, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		Logger.Log.Log("Error seeking to end of file: "+err.Error(), "ERROR")
		return 0, err
	}

	// Serialisierte Daten in die Datei schreiben
	_, err = file.Write(encodedBytes)
	if err != nil {
		Logger.Log.Log("Error writing to file: "+err.Error(), "ERROR")
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
		Logger.Log.Log("Error decoding payload: "+err.Error(), "ERROR")
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
		Logger.Log.Log("Error getting file info: "+err.Error(), "ERROR")
		// We panic here because we can't continue without the file info
		panic(err)
	}

	// if the file is not empty we map it to memory
	if fileInfo.Size() != 0 {
		mappedData, err := syscall.Mmap(int(f.File[collection].Fd()), 0, int(fileInfo.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
		if err != nil {
			Logger.Log.Log("Error mapping file: "+err.Error(), "ERROR")
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
			Logger.Log.Log("Error unmapping file: "+err.Error(), "ERROR")
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
	_, err := os.Stat(*ArgsParser.Ap.FileStore + collection + ".bin")
	if err != nil {
		// if not create it
		file, err := os.Create(*ArgsParser.Ap.FileStore + collection + ".bin")
		if err != nil {
			panic(err)
		}
		file.Close()
	}
	f.FileName[collection] = *ArgsParser.Ap.FileStore + collection + ".bin"
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
	_, err = os.Stat(*ArgsParser.Ap.FileStore + collection + "_meta.bin")
	if err == nil {
		err = os.Remove(*ArgsParser.Ap.FileStore + collection + "_meta.bin")
		if err != nil {
			Logger.Log.Log("Error deleting meta file: "+err.Error(), "ERROR")
		}
	}
	// Remove the collection.json if exists
	_, err = os.Stat(*ArgsParser.Ap.FileStore + collection + ".json")
	if err == nil {
		err = os.Remove(*ArgsParser.Ap.FileStore + collection + ".json")
		if err != nil {
			Logger.Log.Log("Error deleting collection config file: "+err.Error(), "ERROR")
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
func (w *FileMapper) SaveVectorWriter(id string, datastart, payloadstart int64, collection string) (int64, error) {
	// Lock the Wal
	w.Mut[collection].Lock()
	defer w.Mut[collection].Unlock()

	// Open the file "collection"_meta.bin
	file, err := os.OpenFile(*ArgsParser.Ap.FileStore+collection+"_meta.bin", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		Logger.Log.Log("Error opening meta.json file: "+err.Error(), "ERROR")
		return 0, err
	}
	defer file.Close()

	// Get the position of the Filepointer
	pos, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		Logger.Log.Log("Error seeking to end of file: "+err.Error(), "ERROR")
		return 0, err
	}

	// Create the SaveVector
	sv := SaveVector{VectorID: id, DataStart: datastart, PayloadStart: payloadstart, SaveVectorPosition: pos}

	// use json to encode the SaveVector
	encoder := json.NewEncoder(file)
	err = encoder.Encode(sv)
	if err != nil {
		Logger.Log.Log("Error encoding SaveVector: "+err.Error(), "ERROR")
		return 0, err
	}
	return pos, nil
}

// SaveVectorRead will read the vector.ID, vector.DataStart, vector.PayloadStart from the file system and returns a map of vectors
func (w *FileMapper) SaveVectorRead(collection string) (*map[string]SaveVector, error) {
	// Lock the Wal - we use a write lock because here will be no memory mapped file
	w.Mut[collection].Lock()
	defer w.Mut[collection].Unlock()

	// Create the map
	vectors := make(map[string]SaveVector)
	// Open the file "collection"_meta.bin if existing
	_, err := os.Stat(*ArgsParser.Ap.FileStore + collection + "_meta.bin")
	if err == nil {
		file, err := os.Open(*ArgsParser.Ap.FileStore + collection + "_meta.bin")
		if err != nil {
			Logger.Log.Log("Error opening meta.json file: "+err.Error(), "ERROR")
			return nil, err
		}
		defer file.Close()

		// Use json to decode the SaveVector
		decoder := json.NewDecoder(file)
		for {
			var sv SaveVector
			// where are we in the file - save the position
			sv.SaveVectorPosition, err = file.Seek(0, io.SeekCurrent)
			if err != nil {
				Logger.Log.Log("Error getting position: "+err.Error(), "ERROR")
				return nil, err
			}
			if err := decoder.Decode(&sv); err == io.EOF {
				break
			} else if err != nil {
				Logger.Log.Log("Error decoding SaveVector: "+err.Error(), "ERROR")
				return nil, err
			}
			vectors[sv.VectorID] = sv
		}
	}
	return &vectors, nil
}

// SaveVectorWriteAt will write the vector.ID, vector.DataStart, vector.PayloadStart to the file system at a specific position
func (w *FileMapper) SaveVectorWriteAt(datastart, payloadstart int64, collection string, pos int64) error {
	// Lock the Wal
	w.Mut[collection].Lock()
	defer w.Mut[collection].Unlock()
	file, err := os.OpenFile(*ArgsParser.Ap.FileStore+collection+"_meta.bin", os.O_RDWR, 0644)
	if err != nil {
		Logger.Log.Log("Error opening meta.json file: "+err.Error(), "ERROR")
		return err
	}
	defer file.Close()

	// Set the file pointer to the position
	_, err = file.Seek(pos, io.SeekStart)
	if err != nil {
		Logger.Log.Log("Error seeking to position in file: "+err.Error(), "ERROR")
		return err
	}

	// Interpret the previous data as a SaveVector
	var prevSaveVector SaveVector
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&prevSaveVector)
	if err != nil {
		Logger.Log.Log("Error decoding previous data: "+err.Error(), "ERROR")
		return err
	}

	// Overwrite the previous element with spaces
	_, err = file.Seek(pos, io.SeekStart)
	if err != nil {
		Logger.Log.Log("Error seeking to position in file: "+err.Error(), "ERROR")
		return err
	}
	prevData, err := json.Marshal(prevSaveVector)
	if err != nil {
		Logger.Log.Log("Error marshalling previous data: "+err.Error(), "ERROR")
		return err
	}

	// Create the savevector
	sv := SaveVector{VectorID: "", DataStart: datastart, PayloadStart: payloadstart, SaveVectorPosition: pos}

	// Use json to encode the SaveVector
	data, err := json.Marshal(sv)
	if err != nil {
		Logger.Log.Log("Error encoding SaveVector: "+err.Error(), "ERROR")
		return err
	}

	// write only if the data is the same length or smaller than the previous data
	if len(data) <= len(prevData) {
		// space out the data
		spaces := make([]byte, len(prevData))
		for i := range spaces {
			spaces[i] = ' '
		}
		_, err = file.Write(spaces)
		if err != nil {
			Logger.Log.Log("Error writing to file: "+err.Error(), "ERROR")
			return err
		}
		// Write the data at the specified position
		_, err = file.WriteAt(data, pos)
		if err != nil {
			Logger.Log.Log("Error writing to file: "+err.Error(), "ERROR")
			return err
		}
		return nil
	} else {
		Logger.Log.Log("Data is larger than previous data", "INFO")
		return fmt.Errorf("Data is larger than previous data - FORBIDDEN - this is a BUG - please report!")
	}
}
