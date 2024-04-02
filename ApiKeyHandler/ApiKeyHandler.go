package ApiKeyHandler

import (
	"VectoriaDB/Logger"
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"os"
)

// ApiKeyHandler struct
type ApiKeyHandler struct {
	ApiKeyHashes map[string]bool
}

// ApiHandler is the global ApiKeyHandler
var ApiHandler *ApiKeyHandler

// init initializes the ApiKeyHandler
func init() {
	ApiHandler = &ApiKeyHandler{ApiKeyHashes: make(map[string]bool)}
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

	// Write the ApiKey to the file
	file, err := os.OpenFile("collections/__apikeys", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		Logger.Log.Log("Error opening file collections/__apikeys")
		return "", err
	}
	defer file.Close()

	// Write the ApiKey to the file
	_, err = file.WriteString(k + "\n")
	if err != nil {
		Logger.Log.Log("Error writing to file collections/__apikeys")
		return "", err
	}
	Logger.Log.Log("ApiKey successfully created and saved")
	return id, nil
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

	// Read the file line by line
	b := make([]byte, 64)
	for {
		_, err := file.Read(b)
		if err != nil {
			break
		}
		ap.ApiKeyHashes[string(b)] = true
	}
	return nil
}

// CheckApiKey will check if the ApiKey is valid
func (ap *ApiKeyHandler) CheckApiKey(apiKey string) bool {
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
