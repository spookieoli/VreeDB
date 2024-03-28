package ApiKeyHandler

import (
	"VectoriaDB/Logger"
	"os"
)

// ApiKeyHandler struct
type ApiKeyHandler struct {
	ApiKeyHashes map[string]string
}

// ApiHandler is the global ApiKeyHandler
var ApiHandler *ApiKeyHandler

// init initializes the ApiKeyHandler
func init() {
	ApiHandler = &ApiKeyHandler{ApiKeyHashes: make(map[string]string)}
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
