package ArgsParser

import (
	"VreeDB/ApiKeyHandler"
	"flag"
)

// ArgsParser struct
type ArgsParser struct {
	Ip           *string
	Port         *int
	Secure       *bool
	CertFile     *string
	KeyFile      *string
	CreateApiKey *bool
}

// Ap is a global ArgsParser
var Ap *ArgsParser

func init() {
	// Create a new ArgsParser
	Ap = &ArgsParser{}

	// Get the flags
	Ap.Ip = flag.String("ip", "0.0.0.0", "The IP to bind the server to")
	Ap.Port = flag.Int("port", 8080, "The port to bind the server to")
	Ap.Secure = flag.Bool("secure", false, "Use HTTPS")
	Ap.CertFile = flag.String("certfile", "", "The path to the certificate file")
	Ap.KeyFile = flag.String("keyfile", "", "The path to the key file")
	Ap.CreateApiKey = flag.Bool("createapikey", false, "Create a new API key")

	// Parse
	flag.Parse()

	// if CreateApiKey is true, then the program will create a new API and show it in the console
	if *Ap.CreateApiKey {
		// Create a new API key
		if len(ApiKeyHandler.ApiHandler.ApiKeyHashes) == 0 {
			apiKey, err := ApiKeyHandler.ApiHandler.CreateApiKey()
			if err != nil {
				panic(err)
			} else {
				// Show the API key
				println("API Key: " + apiKey)
			}
		}
	}
}
