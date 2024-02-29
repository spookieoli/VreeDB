package Logger

import (
	"os"
	"time"
)

type Logger struct {
	Logfile *os.File
	In      chan string
	Quit    chan bool
}

// Log is a singleton
var Log *Logger

// init initializes the Logger - Log is singleton
func init() {
	// open the Log file for write access
	f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// Panic if there is an error - logfile is critical
	if err != nil {
		panic(err)
	}
	Log = &Logger{Logfile: f, In: make(chan string, 100), Quit: make(chan bool)}
}

// Start will start the LoggerService
func (l *Logger) Start() {
	go func() {
		for {
			select {
			case msg := <-l.In:
				// write date:time message to the log file
				l.Log(msg + "\n")
			case <-l.Quit:
				return
			}

		}
	}()
}

// Log writes a string to the log file
func (l *Logger) Log(s string) {
	// write the current date and the time to the log file + the string
	date := time.Now()
	s = date.Format("2006-01-02T15:04:05Z07:00") + " " + s + "\n"
	_, err := l.Logfile.WriteString(s)
	// panic if there is an error - logfile is critical
	if err != nil {
		panic(err)
	}
}

// Stop will stop the LoggerService
func (l *Logger) Stop() {
	l.Quit <- true
}
