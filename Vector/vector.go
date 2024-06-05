package Vector

import (
	"VreeDB/FileMapper"
	"crypto/rand"
	"fmt"
	"sync"
)

// Vector is a struct that holds a slice of float64
type Vector struct {
	Id                 string
	Collection         string
	Data               []float64
	Length             int
	CLength            int
	Payload            *map[string]interface{}
	DataStart          int64
	PayloadStart       int64
	Indexed            bool
	SaveVectorPosition int64
	deleted            bool
	mut                *sync.RWMutex
}

// NewVector returns a new Vector
func NewVector(id string, data []float64, payload *map[string]interface{}, collection string) *Vector {
	if id == "" {
		// Generate a (pseudo) random UUID - salted with crypto/rand
		b := make([]byte, 16)
		_, err := rand.Read(b)
		if err != nil {
			panic(err) // TBD
		}
		id = fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	}

	if collection != "" {
		// This will write the vector to the memory mapped file
		ds, clen, err := FileMapper.Mapper.WriteVector(data, collection)
		if err != nil {
			// if we cannot write to the file we panic
			panic(err)
		}
		// Write the Payload to the memory mapped File
		ps, err := FileMapper.Mapper.WritePayload(payload, collection)
		if err != nil {
			// if we cannot write to the file we panic
			panic(err)
		}
		return &Vector{Id: id, Data: data, Length: len(data), DataStart: ds, Indexed: true, mut: &sync.RWMutex{}, Collection: collection, PayloadStart: ps, CLength: clen, SaveVectorPosition: -1}
	} else {
		return &Vector{Id: id, Data: data, Length: len(data), Payload: payload, Indexed: false, mut: &sync.RWMutex{}, Collection: collection, SaveVectorPosition: -1}
	}
}

// Unindex will read the data from the file and cache it in the Vector
func (v *Vector) Unindex() {
	// Protect the data from being written to while we read it
	v.mut.Lock()
	defer v.mut.Unlock()
	// read the data from the file
	if v.DataStart < 0 {
		return
	}
	v.Data = *FileMapper.Mapper.ReadVector(v.DataStart, v.Length, v.Collection)
	v.Indexed = false
}

// GetData will return the data of the vector
func (v *Vector) GetData() *[]float64 {
	// Protect the data from being written to while we read it
	v.mut.RLock()
	defer v.mut.RUnlock()
	if v.Indexed {
		return FileMapper.Mapper.ReadVector(v.DataStart, v.Length, v.Collection)
	}
	return &v.Data
}

// RecreateMut will recreate the mut
func (v *Vector) RecreateMut() {
	v.mut = &sync.RWMutex{}
}

// Delete marks the Vector as deleted by setting the `deleted` flag to true
func (v *Vector) Delete() {
	v.deleted = true
}
