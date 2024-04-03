package main

import (
	"VreeDB/ArgsParser"
	"VreeDB/Server"
)

func main() {
	// Start the Server
	server := Server.NewServer(*ArgsParser.Ap.Ip, *ArgsParser.Ap.Port, *ArgsParser.Ap.CertFile,
		*ArgsParser.Ap.KeyFile, *ArgsParser.Ap.Secure)
	server.Start()
}
