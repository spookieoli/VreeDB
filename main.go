package main

import (
	"VectoriaDB/Server"
)

func main() {
	// Start the Server
	server := Server.NewServer("0.0.0.0", 8080)
	server.Start()
}
