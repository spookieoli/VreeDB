package ApiKeyHandler

import (
	"VreeDB/ArgsParser"
	"VreeDB/Logger"
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"encoding/gob"
	"fmt"
	"golang.org/x/crypto/argon2"
	"os"
	"sync"
)

// Params represents configuration parameters for cryptographic operations including memory usage, iterations, and other settings.
type Params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

type ApiKey struct {
	Salt []byte
}

// ApiKeyHandler struct
type ApiKeyHandler struct {
	ApiKeyHashes map[string]ApiKey
	argonParams  *Params
	Mut          sync.RWMutex
}

// ApiHandler is the global ApiKeyHandler
var ApiHandler *ApiKeyHandler

// init initializes the ApiKeyHandler
func init() {

	// If collections directory does not exist, create it
	if _, err := os.Stat(*ArgsParser.Ap.FileStore); os.IsNotExist(err) {
		err := os.Mkdir(*ArgsParser.Ap.FileStore, 0755)
		if err != nil {
			panic(err)
		}
	}

	// Set the parameters for the Argon2id algorithm
	p := &Params{memory: 64 * 1024, iterations: 4, parallelism: 2, saltLength: 16, keyLength: 32}

	ApiHandler = &ApiKeyHandler{ApiKeyHashes: make(map[string]ApiKey), Mut: sync.RWMutex{}, argonParams: p}
	if ApiHandler.CheckActive() {
		err := ApiHandler.CreateApiKeyFile()
		if err != nil {
			Logger.Log.Log("Error creating file collections/__apikeys", "ERROR")
			panic(err) // we cannot create the file - kill the server
		}
	}
	ApiHandler.LoadApiKeys()
	Logger.Log.Log("ApiKeyHandler initialized", "INFO")

	// Argument Createapikey is set - create a new ApiKey
	if *ArgsParser.Ap.CreateApiKey {
		apiKey, err := ApiHandler.CreateApiKey()
		if err != nil {
			Logger.Log.Log("Error creating ApiKey", "ERROR")
			panic(err)
		}
		fmt.Println("New ApiKey created (PLEASE NOTE THIS ONE!): " + apiKey)
	}
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
		Logger.Log.Log("Error creating file collections/__apikeys", "ERROR")
		return err
	}
	defer file.Close()
	Logger.Log.Log("File collections/__apikeys created", "INFO")
	return nil
}

// salt creates a random salt
func (ap *ApiKeyHandler) salt() ([]byte, error) {
	salt := make([]byte, ap.argonParams.saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

// createRandomString creates a random string
func (ap *ApiKeyHandler) createRandomString() (string, error) {
	wrd := make([]byte, 32)
	_, err := rand.Read(wrd)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%X%X%X%X%X", wrd[0:4], wrd[4:6], wrd[6:8], wrd[8:10], wrd[10:]), nil
}

// generateFromPassword generates a hash from a password and a salt
func (ap *ApiKeyHandler) generateFromPassword(password string, salt []byte) string {
	// Create the hash
	hash := argon2.IDKey([]byte(password), salt, ap.argonParams.iterations, ap.argonParams.memory, ap.argonParams.parallelism, ap.argonParams.keyLength)

	// Generate String from hash
	return fmt.Sprintf("%x", hash)
}

// CreateApiKey will add a new ApiKey to the ApiKeyHandler
func (ap *ApiKeyHandler) CreateApiKey() (string, error) {
	ap.Mut.Lock()
	defer ap.Mut.Unlock()
	// Generate a (pseudo) random STRING - salted with crypto/rand
	id, err := ap.createRandomString()
	if err != nil {
		Logger.Log.Log("Error creating random string", "ERROR")
		return "", err
	}

	// Create a salt
	salt, err := ap.salt()
	if err != nil {
		Logger.Log.Log("Error creating salt", "ERROR")
		return "", err
	}

	// hash the ApiKey
	k := ap.generateFromPassword(id, salt)
	ap.ApiKeyHashes[k] = ApiKey{Salt: salt}

	// Write the changes to the file using gob
	file, err := os.OpenFile("collections/__apikeys", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		Logger.Log.Log("Error opening file collections/__apikeys", "ERROR")
		return "", err
	}
	defer file.Close()

	// Gob encode the map
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(ap.ApiKeyHashes)
	if err != nil {
		Logger.Log.Log("Error encoding map to file", "ERROR")
		return "", err
	}

	// Write the map to the file
	_, err = file.Write(buf.Bytes())
	if err != nil {
		Logger.Log.Log("Error writing map to file", "ERROR")
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
	file, err := os.OpenFile(*ArgsParser.Ap.FileStore+"/__apikeys", os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		Logger.Log.Log("Error opening file "+(*ArgsParser.Ap.FileStore)+"/__apikeys", "ERROR")
		return err
	}
	defer file.Close()

	// Gob encode the map
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(ap.ApiKeyHashes)
	if err != nil {
		Logger.Log.Log("Error encoding map to file", "ERROR")
		return err
	}

	// Write the map to the file
	_, err = file.Write(buf.Bytes())
	if err != nil {
		Logger.Log.Log("Error writing map to file", "ERROR")
		return err
	}

	return nil
}

// LoadApiKeys will load all ApiKeys Hashes from the file
func (ap *ApiKeyHandler) LoadApiKeys() error {
	// Open the file collections/__apikeys
	file, err := os.Open("collections/__apikeys")
	if err != nil {
		Logger.Log.Log("Error opening file collections/__apikeys", "ERROR")
		return err
	}
	defer file.Close()

	// Read the file using gob decoding
	dec := gob.NewDecoder(file)
	err = dec.Decode(&ap.ApiKeyHashes)
	if err != nil {
		Logger.Log.Log("Error decoding file collections/__apikeys - EOF Error OK if no APIKEY created!", "INFO")
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

	// ok wie got some ApiKeys - lets check
	for k := range ap.ApiKeyHashes {
		h := ap.generateFromPassword(apiKey, ap.ApiKeyHashes[k].Salt)
		if _, ok := ap.ApiKeyHashes[h]; ok {
			return true
		}
	}
	return false
}

// CheckIfEmpty checks if there are any ApiKeys stored
func (ap *ApiKeyHandler) CheckIfEmpty() bool {
	ap.Mut.RLock()
	defer ap.Mut.RUnlock()
	if len(ap.ApiKeyHashes) == 0 {
		return true
	}
	return false
}
