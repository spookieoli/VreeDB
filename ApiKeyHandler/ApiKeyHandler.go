package ApiKeyHandler

import (
	"VreeDB/Logger"
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"encoding/gob"
	"fmt"
	"os"
	"sync"
)

// ApiKeyHandler struct
type ApiKeyHandler struct {
	ApiKeyHashes map[string]bool
	Mut          sync.RWMutex
}

// ApiHandler is the global ApiKeyHandler
var ApiHandler *ApiKeyHandler

// init initializes the ApiKeyHandler
func init() {
	ApiHandler = &ApiKeyHandler{ApiKeyHashes: make(map[string]bool), Mut: sync.RWMutex{}}
	if ApiHandler.CheckActive() {
		err := ApiHandler.CreateApiKeyFile()
		if err != nil {
			Logger.Log.Log("Error creating file collections/__apikeys")
			panic(err) // we cannot create the file - kill the server
		}
	}
	Logger.Log.Log("ApiKeyHandler initialized")
}

// CheckActive will check if an ApiKey was already created
func (ap *ApiKeyHandler) CheckActive() bool {
	// We check if the file collections/__apikeys exists
	if _, err := os.Stat("collections/__apikeys"); err != nil {
		return true
	}
	return false
}

// CreateApiKeyFile will create the binary file where alle hashed ApiKeys are stored
func (ap *ApiKeyHandler) CreateApiKeyFile() error {
	// Create the file collections/__apikeys
	file, err := os.Create("collections/__apikeys")
	if err != nil {
		Logger.Log.Log("Error creating file collections/__apikeys")
		return err
	}
	defer file.Close()
	Logger.Log.Log("File collections/__apikeys created")
	return nil
}

// CreateApiKey will add a new ApiKey to the ApiKeyHandler
func (ap *ApiKeyHandler) CreateApiKey() (string, error) {
	ap.Mut.Lock()
	defer ap.Mut.Unlock()
	// Generate a (pseudo) random STRING - salted with crypto/rand
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("%X%X%X%X%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	// hash the ApiKey
	h := sha512.New()
	h.Write([]byte(id))
	k := fmt.Sprintf("%x", h.Sum(nil))
	ap.ApiKeyHashes[k] = true

	// Write the changes to the file using gob
	file, err := os.OpenFile("collections/__apikeys", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		Logger.Log.Log("Error opening file collections/__apikeys")
		return "", err
	}
	defer file.Close()

	// Gob encode the map
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(ap.ApiKeyHashes)
	if err != nil {
		Logger.Log.Log("Error encoding map to file")
		return "", err
	}

	// Write the map to the file
	_, err = file.Write(buf.Bytes())
	if err != nil {
		Logger.Log.Log("Error writing map to file")
		return "", err
	}

	return id, nil
}

// DeleteApiKey will delete an ApiKey from the ApiKeyHandler
func (ap *ApiKeyHandler) DeleteApiKey(apiKey string) error {
	ap.Mut.Lock()
	defer ap.Mut.Unlock()
	// hash the ApiKey
	h := sha512.New()
	h.Write([]byte(apiKey))
	k := fmt.Sprintf("%x", h.Sum(nil))
	delete(ap.ApiKeyHashes, k)

	// Write the changes to the file using gob
	file, err := os.OpenFile("collections/__apikeys", os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		Logger.Log.Log("Error opening file collections/__apikeys")
		return err
	}
	defer file.Close()

	// Gob encode the map
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(ap.ApiKeyHashes)
	if err != nil {
		Logger.Log.Log("Error encoding map to file")
		return err
	}

	// Write the map to the file
	_, err = file.Write(buf.Bytes())
	if err != nil {
		Logger.Log.Log("Error writing map to file")
		return err
	}

	return nil
}

// LoadApiKeys will load all ApiKeys Hashes from the file
func (ap *ApiKeyHandler) LoadApiKeys() error {
	// Open the file collections/__apikeys
	file, err := os.Open("collections/__apikeys")
	if err != nil {
		Logger.Log.Log("Error opening file collections/__apikeys")
		return err
	}
	defer file.Close()

	// Read the file using gob decoding
	dec := gob.NewDecoder(file)
	err = dec.Decode(&ap.ApiKeyHashes)
	if err != nil {
		Logger.Log.Log("Error decoding file collections/__apikeys - EOF Error OK on first run!")
		return err
	}
	return nil
}

// CheckApiKey will check if the ApiKey is valid
func (ap *ApiKeyHandler) CheckApiKey(apiKey string) bool {
	ap.Mut.RLock()
	defer ap.Mut.RUnlock()
	// If the map is empty we return true
	if len(ap.ApiKeyHashes) == 0 {
		return true
	}

	// hash the ApiKey
	h := sha512.New()
	h.Write([]byte(apiKey))
	k := fmt.Sprintf("%x", h.Sum(nil))
	if _, ok := ap.ApiKeyHashes[k]; ok {
		return true
	}
	return false
}
