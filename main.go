package main

/*
#cgo CFLAGS: -I./avx
#cgo LDFLAGS: -L./avx -lavx_check
#include "avx_check.c"
*/
import "C"

import (
	"VreeDB/AccessDataHUB"
	"VreeDB/ArgsParser"
	"VreeDB/Server"
	"fmt"
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

	// Check if the CPU supports vector acceleration
	checkVectorAcceleration()

	// Start the Server
	server := Server.NewServer(*ArgsParser.Ap.Ip, *ArgsParser.Ap.Port, *ArgsParser.Ap.CertFile,
		*ArgsParser.Ap.KeyFile, *ArgsParser.Ap.Secure)
	// Add Systemevent to the AccessList
	AccessDataHUB.AccessList.ReadChan <- "SYSTEMEVENT"
	server.Start()
}

// check_vector_acceleration checks if the CPU supports vector acceleration and prints the selected acceleration method.
// If AVX256 flag is true, it checks for AVX support and panics if not supported.
// If Neon flag is true, it checks for Neon support and panics if not supported.
// But only if the flag is true.
func checkVectorAcceleration() {
	// if AVX is true than we check if the CPU supports AVX
	if *ArgsParser.Ap.AVX256 {
		if C.check_avx_support() == 0 {
			panic("CPU does not support AVX")
		} else {
			fmt.Println("CPU supports AVX256 - using AVX256")
		}
	} else if *ArgsParser.Ap.Neon {
		if C.check_neon_support() == 0 {
			panic("CPU does not support Neon")
		} else {
			fmt.Println("CPU supports Neon - using Neon")
		}
	}
}
