package main

import (
	"VreeDB/AccessDataHUB"
	"VreeDB/ArgsParser"
	"VreeDB/Server"
	"os"
	"runtime/pprof"
)

func main() {
	// Check if we should collect PGO data
	if *ArgsParser.Ap.PGOCollect {
		f, err := os.Create("default.pgo")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		err = pprof.StartCPUProfile(f)
		if err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	}

	// Start the Server
	server := Server.NewServer(*ArgsParser.Ap.Ip, *ArgsParser.Ap.Port, *ArgsParser.Ap.CertFile,
		*ArgsParser.Ap.KeyFile, *ArgsParser.Ap.Secure)
	// Add Systemevent to the AccessList
	AccessDataHUB.AccessList.ReadChan <- "SYSTEMEVENT"
	server.Start()
}
