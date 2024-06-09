package Logger

import (
	"VreeDB/ArgsParser"
	"os"
	"strings"
	"time"
)

type Level string

const (
	INFO    Level = "INFO"    // INFO is the default log level
	DEBUG   Level = "DEBUG"   // DEBUG is a debug log level
	WARNING Level = "WARNING" // WARNING is a warning log level
	ERROR   Level = "ERROR"   // ERROR is an error log level
)

// Logger struct
type Logger struct {
	Logfile  *os.File
	In       chan *LogMessage
	Quit     chan bool
	LOGLEVEL Level
}

// LogMessage struct represents a log message and its associated log level.
type LogMessage struct {
	Message string
	Level   string
}

// Log is a singleton
var Log *Logger

// init initializes the Logger - Log is singleton
func init() {
	// open the Log file for write access
	f, err := os.OpenFile(*ArgsParser.Ap.Loglocation, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// Panic if there is an error - logfile is critical
	if err != nil {
		panic(err)
	}
	Log = &Logger{Logfile: f, In: make(chan *LogMessage, 100), Quit: make(chan bool), LOGLEVEL: Level(*ArgsParser.Ap.LogLevel)}
}

// Start will start the LoggerService
func (l *Logger) Start() {
	go func() {
		for {
			select {
			case msg := <-l.In:
				// write date:time message to the log file
				if Level(msg.Level) == l.LOGLEVEL {
					date := time.Now()
					var sb strings.Builder
					sb.WriteString(date.Format("2006-01-02T15:04:05Z07:00"))
					sb.WriteString(" [")
					sb.WriteString(msg.Level)
					sb.WriteString("] ")
					sb.WriteString(msg.Message)
					sb.WriteString("\n")
					l.LogIt(msg.Message)
				}
			case <-l.Quit:
				return
			}

		}
	}()
}

// LogIt writes a string to the log file
func (l *Logger) LogIt(s string) {
	// message to the file
	_, err := l.Logfile.WriteString(s)
	// panic if there is an error - logfile is critical
	if err != nil {
		panic(err)
	}
}

// Log will log a message with a given level to the Logger's input channel.
// The message and level are wrapped in a LogMessage struct and sent to the channel.
// It does not return any value.
func (l *Logger) Log(message, level string) {
	l.In <- &LogMessage{Message: message, Level: level}
}

// Stop will stop the LoggerService
func (l *Logger) Stop() {
	l.Quit <- true
}
