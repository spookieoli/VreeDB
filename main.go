package main

import (
	"VectoriaDB/ArgsParser"
	"VectoriaDB/Server"
)

func main() {
	// Vars
	var b bool

	// Create the ArgsParser
	argsParser := ArgsParser.NewArgsParser()

	// Start the Server
	server := Server.NewServer(*argsParser.Ip, *argsParser.Port, *argsParser.CertFile,
		*argsParser.KeyFile, b, argsParser)
	server.Start()
}
