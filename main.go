package main

import (
	"VectoriaDB/ArgsParser"
	"VectoriaDB/Server"
)

func main() {
	// Create the ArgsParser
	argsParser := ArgsParser.NewArgsParser()

	// Start the Server
	server := Server.NewServer(*argsParser.Ip, *argsParser.Port, *argsParser.CertFile,
		*argsParser.KeyFile, *argsParser.Secure, argsParser)
	server.Start()
}
