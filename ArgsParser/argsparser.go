package ArgsParser

import (
	"flag"
	"runtime"
)

// ArgsParser struct
type ArgsParser struct {
	Ip            *string
	Port          *int
	SearchThreads *int
	Secure        *bool
	CertFile      *string
	KeyFile       *string
	CreateApiKey  *bool
	Loglocation   *string
	FileStore     *string
	LogLevel      *string
	PGOCollect    *bool
	AVX           *bool
	AVX256        *bool
	//AVX512        *bool
}

// Ap is a global ArgsParser
var Ap *ArgsParser

func init() {
	// Create a new ArgsParser
	Ap = &ArgsParser{}

	// Get the flags
	Ap.Ip = flag.String("ip", "0.0.0.0", "The IP to bind the server to")
	Ap.Loglocation = flag.String("loglocation", "log.txt", "The location of the log file")
	Ap.FileStore = flag.String("filestore", "collections/", "The directory of the file store")
	Ap.Port = flag.Int("port", 8080, "The port to bind the server to")
	Ap.Secure = flag.Bool("secure", false, "Use HTTPS")
	Ap.CertFile = flag.String("certfile", "", "The path to the certificate file")
	Ap.KeyFile = flag.String("keyfile", "", "The path to the key file")
	Ap.CreateApiKey = flag.Bool("createapikey", false, "Create a new API key")
	Ap.SearchThreads = flag.Int("searchthreads", runtime.NumCPU()/2, "The number of search threads")
	Ap.LogLevel = flag.String("loglevel", "INFO", "The log level")
	Ap.PGOCollect = flag.Bool("pgocollect", false, "Collect PGO data")
	Ap.AVX256 = flag.Bool("avx256", false, "Use AVX256")
	// Ap.AVX512 = flag.Bool("avx512", false, "Use AVX512")

	// Parse
	flag.Parse()

	// Check if SearchThreads is gt 0
	if *Ap.SearchThreads <= 0 {
		// Exit
		panic("SearchThreads must be greater than 0")
	}

	// Check if Ap.FileStore ends with a slash
	if (*Ap.FileStore)[len(*Ap.FileStore)-1] != '/' {
		*Ap.FileStore += "/"
	}
}
