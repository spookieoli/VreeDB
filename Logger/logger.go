package Logger

import (
	"VreeDB/ArgsParser"
	"log"
	"os"
)

// Log is a singleton
var infoLogger *log.Logger
var errorLogger *log.Logger
var debugLogger *log.Logger

// init initializes the Logger - Log is singleton
func init() {
	// open the Log file for write access
	f, err := os.OpenFile(*ArgsParser.Ap.Loglocation, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// Panic if there is an error - logfile is critical
	if err != nil {
		panic(err)
	}
	// Create the loggers
	infoLogger = log.New(f, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(f, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	debugLogger = log.New(f, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
}
