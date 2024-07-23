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

	// if AVX is true than we check if the CPU supports AVX
	if *ArgsParser.Ap.AVX {
		if C.check_avx_support() == 0 {
			panic("CPU does not support AVX")
		} else {
			if *ArgsParser.Ap.AVX && C.check_avx512_support() == 1 {
				fmt.Println("CPU supports AVX256 and AVX512 - using AVX512")
				*ArgsParser.Ap.AVX512 = true
			} else {
				fmt.Println("CPU supports AVX256 - using AVX256")
				*ArgsParser.Ap.AVX256 = true
			}
		}
	}

	// Start the Server
	server := Server.NewServer(*ArgsParser.Ap.Ip, *ArgsParser.Ap.Port, *ArgsParser.Ap.CertFile,
		*ArgsParser.Ap.KeyFile, *ArgsParser.Ap.Secure)
	// Add Systemevent to the AccessList
	AccessDataHUB.AccessList.ReadChan <- "SYSTEMEVENT"
	server.Start()
}
