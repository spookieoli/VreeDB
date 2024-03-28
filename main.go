package main

import (
	"VectoriaDB/ArgsParser"
	"VectoriaDB/Server"
	"log"
	"strconv"
)

func main() {
	// Vars
	var b bool

	// Create the ArgsParser
	argsParser := ArgsParser.NewArgsParser()
	if argsParser.ParseArgs() == false {
		log.Fatal("Arguments are in the wrong format")
	}

	// Check if SECURE is set
	if argsParser.Options["SECURE"] == "" {
		b = false
	} else {
		b = true
	}

	// Port must be a number so we convert it
	port, err := strconv.Atoi(argsParser.Options["PORT"])
	if err != nil {
		log.Fatal("Cannot convert Port - please check the format of the port")
	}

	// Start the Server
	server := Server.NewServer(argsParser.Options["IP"], port, argsParser.Options["CERTFILE"],
		argsParser.Options["KEYFILE"], b, argsParser)
	server.Start()
}
