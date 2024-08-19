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