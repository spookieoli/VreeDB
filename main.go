package main

import (
	"VreeDB/AccessDataHUB"
	"VreeDB/ArgsParser"
	"VreeDB/Server"
)

func main() {
	// Start the Server
	server := Server.NewServer(*ArgsParser.Ap.Ip, *ArgsParser.Ap.Port, *ArgsParser.Ap.CertFile,
		*ArgsParser.Ap.KeyFile, *ArgsParser.Ap.Secure)
	// Add Systemevent to the AccessList
	AccessDataHUB.AccessList.ReadChan <- "SYSTEMEVENT"
	server.Start()
}
