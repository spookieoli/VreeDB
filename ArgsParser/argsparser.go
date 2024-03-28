package ArgsParser

import (
	"os"
	"strings"
)

// ArgsParser struct
type ArgsParser struct {
	Options map[string]string
}

// NewArgsParser returns a new ArgsParser // TODO: SUBJECT TO HEAVY CHANGE -- use flags instead
func NewArgsParser() *ArgsParser {
	m := make(map[string]string)

	// Add the options
	m["IP"] = "0.0.0.0"
	m["PORT"] = "8080"
	m["SECURE"] = ""
	m["CERTFILE"] = ""
	m["KEYFILE"] = ""
	return &ArgsParser{Options: m}
}

// ParseArgs parses the arguments, options must start with -- and have a value delimited with =
func (a *ArgsParser) ParseArgs() bool {
	args := os.Args
	for i := 1; i < len(args); i++ {
		if len(args[i]) > 2 && args[i][0:2] == "--" {
			split := strings.Split(args[i][2:], "=")
			if len(split) == 2 {
				// Check if the option is valid
				if _, ok := a.Options[split[0]]; !ok {
					return false
				}
				a.Options[split[0]] = split[1]
			}
		} else {
			return false
		}
	}
	return true
}
