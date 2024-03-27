package ArgsParser

import (
	"os"
	"strings"
)

// ArgsParser struct
type ArgsParser struct {
	Options map[string]string
}

// NewArgsParser returns a new ArgsParser
func NewArgsParser() *ArgsParser {
	return &ArgsParser{Options: make(map[string]string)}
}

// ParseArgs parses the arguments, options must start with -- and have a value delimited with =
func (a *ArgsParser) ParseArgs() bool {
	args := os.Args
	for i := 1; i < len(args); i++ {
		if len(args[i]) > 2 && args[i][0:2] == "--" {
			split := strings.Split(args[i][2:], "=")
			if len(split) == 2 {
				a.Options[split[0]] = split[1]
			}
		} else {
			return false
		}
	}
	return true
}
