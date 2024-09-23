package main

/*
#ifndef AVX_CHECK_H
#define AVX_CHECK_H

int check_avx_support();
int check_avx512_support();
int check_neon_support();

#endif // AVX_CHECK_H

#include <stdint.h>

#if defined(__x86_64__) || defined(_M_X64) || defined(__i386) || defined(_M_IX86)

void cpuid(int info[4], int InfoType){
    __asm__ __volatile__(
        "cpuid":
        "=a"(info[0]), "=b"(info[1]), "=c"(info[2]), "=d"(info[3]) :
        "a"(InfoType)
    );
}

int check_avx_support() {
    int info[4];
    cpuid(info, 0);
    if (info[0] < 1)
        return 0; // No AVX support

    cpuid(info, 1);
    if ((info[2] & ((int)1 << 28)) == 0)
        return 0; // No AVX support

    uint64_t xcrFeatureMask;
    __asm__ __volatile__ (
        "xgetbv" : "=a" (xcrFeatureMask) : "c" (0) : "%edx"
    );
    if ((xcrFeatureMask & 6) != 6)
        return 0; // No AVX support

    return 1; // AVX supported
}

int check_avx512_support() {
    int info[4];
    cpuid(info, 0);
    if (info[0] < 7)
        return 0; // No AVX512 support

    cpuid(info, 7);
    if ((info[1] & ((int)1 << 16)) == 0)
        return 0; // No AVX512 support

    uint64_t xcrFeatureMask;
    __asm__ __volatile__ (
        "xgetbv" : "=a" (xcrFeatureMask) : "c" (0) : "%edx"
    );
    if ((xcrFeatureMask & 0xE6) != 0xE6)
        return 0; // No AVX512 support

    return 1; // AVX512 supported
}

int check_neon_support() {
    return 0;
}

// if arm is defined - check for neon support
#elif defined(__arm__) || defined(__aarch64__)

#include <arm_neon.h>

int check_neon_support() {
    #if defined(__aarch64__)
    return 1;
    #else
    // Check for NEON support on ARM32
    uint32_t info;
    __asm__ __volatile__ (
        "mrc p15, 0, %0, c1, c0, 2"
        : "=r" (info)
    );
    return (info & (1 << 12)) != 0;
    #endif
}

int check_avx_support() {
    return 0;
}

int check_avx512_support() {
    return 0;
}

#else
int check_avx_support() {
    return 0;
}

int check_avx512_support() {
    return 0;
}

int check_neon_support() {
    return 0;
}
#endif

*/
import "C"

import (
	"VreeDB/AccessDataHUB"
	"VreeDB/ArgsParser"
	"VreeDB/Server"
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
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

	// The Server variable
	var server *Server.Server

	go func() {
		// Start the Server - to shutdown the server gracefully we need to use a go routine
		server = Server.NewServer(*ArgsParser.Ap.Ip, *ArgsParser.Ap.Port, *ArgsParser.Ap.CertFile,
			*ArgsParser.Ap.KeyFile, *ArgsParser.Ap.Secure)
		// Add Systemevent to the AccessList
		AccessDataHUB.AccessList.ReadChan <- "SYSTEMEVENT"
		server.Start()
		fmt.Println("Stopped serving clients")
	}()

	// create the signal channel
	signalChan := make(chan os.Signal, 1)

	// wait for the signal
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// get the signal
	<-signalChan

	// let the user know we are shutting down
	fmt.Println("Shutting down the server")

	// Shutdown the server
	server.Shutdown()
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
