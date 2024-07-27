package Utils

/*
#cgo CFLAGS: -O3 -mavx
#cgo LDFLAGS: -lm

#include <stdint.h>

#if defined(__x86_64__) || defined(_M_X64) || defined(__i386) || defined(_M_IX86)
#include <immintrin.h>
#include <math.h>

double euclidean_distance_avx(const double* a, const double* b, int n) {
    __m256d sum = _mm256_setzero_pd(); // Setzt sum auf 0
    int i;

    // Unroll the loop to process 16 elements at a time if possible
    for (i = 0; i <= n - 16; i += 16) {
        __m256d va1 = _mm256_loadu_pd(&a[i]);
        __m256d vb1 = _mm256_loadu_pd(&b[i]);
        __m256d diff1 = _mm256_sub_pd(va1, vb1);
        __m256d sq1 = _mm256_mul_pd(diff1, diff1);

        __m256d va2 = _mm256_loadu_pd(&a[i + 4]);
        __m256d vb2 = _mm256_loadu_pd(&b[i + 4]);
        __m256d diff2 = _mm256_sub_pd(va2, vb2);
        __m256d sq2 = _mm256_mul_pd(diff2, diff2);

        __m256d va3 = _mm256_loadu_pd(&a[i + 8]);
        __m256d vb3 = _mm256_loadu_pd(&b[i + 8]);
        __m256d diff3 = _mm256_sub_pd(va3, vb3);
        __m256d sq3 = _mm256_mul_pd(diff3, diff3);

        __m256d va4 = _mm256_loadu_pd(&a[i + 12]);
        __m256d vb4 = _mm256_loadu_pd(&b[i + 12]);
        __m256d diff4 = _mm256_sub_pd(va4, vb4);
        __m256d sq4 = _mm256_mul_pd(diff4, diff4);

        sum = _mm256_add_pd(sum, sq1);
        sum = _mm256_add_pd(sum, sq2);
        sum = _mm256_add_pd(sum, sq3);
        sum = _mm256_add_pd(sum, sq4);
    }

    // Handle the remaining elements (if any) in chunks of 4
    for (; i <= n - 4; i += 4) {
        __m256d va = _mm256_loadu_pd(&a[i]);
        __m256d vb = _mm256_loadu_pd(&b[i]);
        __m256d diff = _mm256_sub_pd(va, vb);
        __m256d sq = _mm256_mul_pd(diff, diff);
        sum = _mm256_add_pd(sum, sq);
    }

    // Sum the elements in the __m256d register
    __m256d temp = _mm256_hadd_pd(sum, sum);
    temp = _mm256_hadd_pd(temp, temp);
    __m128d sum_high = _mm256_extractf128_pd(temp, 1);
    __m128d result = _mm_add_pd(_mm256_castpd256_pd128(temp), sum_high);
    double final_sum = _mm_cvtsd_f64(_mm_hadd_pd(result, result));

    // Handle the remaining elements (if any) one by one
    for (; i < n; i++) {
        double diff = a[i] - b[i];
        final_sum += diff * diff;
    }

    return sqrt(final_sum);
}

double cosine_distance_avx(const double* a, const double* b, int n) {
    __m256d sum_a = _mm256_setzero_pd();
    __m256d sum_b = _mm256_setzero_pd();
    __m256d sum_ab = _mm256_setzero_pd();
    int i;

    // Unroll the loop to process 16 elements at a time if possible
    for (i = 0; i <= n - 16; i += 16) {
        __m256d va1 = _mm256_loadu_pd(&a[i]);
        __m256d vb1 = _mm256_loadu_pd(&b[i]);
        sum_ab = _mm256_add_pd(sum_ab, _mm256_mul_pd(va1, vb1));
        sum_a = _mm256_add_pd(sum_a, _mm256_mul_pd(va1, va1));
        sum_b = _mm256_add_pd(sum_b, _mm256_mul_pd(vb1, vb1));

        __m256d va2 = _mm256_loadu_pd(&a[i + 4]);
        __m256d vb2 = _mm256_loadu_pd(&b[i + 4]);
        sum_ab = _mm256_add_pd(sum_ab, _mm256_mul_pd(va2, vb2));
        sum_a = _mm256_add_pd(sum_a, _mm256_mul_pd(va2, va2));
        sum_b = _mm256_add_pd(sum_b, _mm256_mul_pd(vb2, vb2));

        __m256d va3 = _mm256_loadu_pd(&a[i + 8]);
        __m256d vb3 = _mm256_loadu_pd(&b[i + 8]);
        sum_ab = _mm256_add_pd(sum_ab, _mm256_mul_pd(va3, vb3));
        sum_a = _mm256_add_pd(sum_a, _mm256_mul_pd(va3, va3));
        sum_b = _mm256_add_pd(sum_b, _mm256_mul_pd(vb3, vb3));

        __m256d va4 = _mm256_loadu_pd(&a[i + 12]);
        __m256d vb4 = _mm256_loadu_pd(&b[i + 12]);
        sum_ab = _mm256_add_pd(sum_ab, _mm256_mul_pd(va4, vb4));
        sum_a = _mm256_add_pd(sum_a, _mm256_mul_pd(va4, va4));
        sum_b = _mm256_add_pd(sum_b, _mm256_mul_pd(vb4, vb4));
    }

    // Handle the remaining elements (if any) in chunks of 4
    for (; i <= n - 4; i += 4) {
        __m256d va = _mm256_loadu_pd(&a[i]);
        __m256d vb = _mm256_loadu_pd(&b[i]);
        sum_ab = _mm256_add_pd(sum_ab, _mm256_mul_pd(va, vb));
        sum_a = _mm256_add_pd(sum_a, _mm256_mul_pd(va, va));
        sum_b = _mm256_add_pd(sum_b, _mm256_mul_pd(vb, vb));
    }

    // Sum the elements in the __m256d registers
    __m256d temp_ab = _mm256_add_pd(sum_ab, _mm256_permute2f128_pd(sum_ab, sum_ab, 1));
    temp_ab = _mm256_add_pd(temp_ab, _mm256_permute_pd(temp_ab, 0x5));
    double final_sum_ab = _mm_cvtsd_f64(_mm256_castpd256_pd128(temp_ab)) + _mm_cvtsd_f64(_mm256_extractf128_pd(temp_ab, 1));

    __m256d temp_a = _mm256_add_pd(sum_a, _mm256_permute2f128_pd(sum_a, sum_a, 1));
    temp_a = _mm256_add_pd(temp_a, _mm256_permute_pd(temp_a, 0x5));
    double final_sum_a = _mm_cvtsd_f64(_mm256_castpd256_pd128(temp_a)) + _mm_cvtsd_f64(_mm256_extractf128_pd(temp_a, 1));

    __m256d temp_b = _mm256_add_pd(sum_b, _mm256_permute2f128_pd(sum_b, sum_b, 1));
    temp_b = _mm256_add_pd(temp_b, _mm256_permute_pd(temp_b, 0x5));
    double final_sum_b = _mm_cvtsd_f64(_mm256_castpd256_pd128(temp_b)) + _mm_cvtsd_f64(_mm256_extractf128_pd(temp_b, 1));

    // Handle the remaining elements (if any) one by one
    for (; i < n; i++) {
        double va = a[i];
        double vb = b[i];
        final_sum_ab += va * vb;
        final_sum_a += va * va;
        final_sum_b += vb * vb;
    }

    return 1.0 - (final_sum_ab / (sqrt(final_sum_a) * sqrt(final_sum_b)));
}

// dummy for x86 x64
double euclidean_distance_neon(double *array1, double *array2, int len){
	return 0;
}

double cosine_distance_neon(double *array1, double *array2, int len) {
	return 0;
}

#elif defined(__arm__) || defined(__aarch64__)

#include <arm_neon.h>
#include <math.h>

double euclidean_distance_neon(double *array1, double *array2, int len) {
	double result = 0;
	float64x2_t a, b, resultNeon = vdupq_n_f64(0.0);

    // Loop over full 2-value chunks of the arrays
	for(int i = 0; i < len - 1; i+=2) {
		a = vld1q_f64(array1 + i);
		b = vld1q_f64(array2 + i);

        a = vsubq_f64(a, b); // a = a - b
		resultNeon = vmlaq_f64(resultNeon, a, a); // resultNeon = resultNeon + a * a
	}

    // Add results of vector computation back into a standard double variable
    double resultArray[2];
	vst1q_f64(resultArray, resultNeon);
    result += resultArray[0] + resultArray[1];

	// If the array length is not even, we have one remaining value to process
	if(len % 2 != 0) {
		double diff = array1[len-1] - array2[len-1];
		result += diff * diff;
	}

	return sqrt(result);
}

double cosine_distance_neon(double *array1, double *array2, int len) {
    double dot_product = 0.0, norm_a = 0.0, norm_b = 0.0;
    float64x2_t a, b, dp_vec = vdupq_n_f64(0.0), norm_a_vec = vdupq_n_f64(0.0), norm_b_vec = vdupq_n_f64(0.0);

    // Loop over full 2-value chunks of the arrays
    for(int i = 0; i < len - 1; i+=2) {
        a = vld1q_f64(array1 + i);
        b = vld1q_f64(array2 + i);

        dp_vec = vmlaq_f64(dp_vec, a, b); // dp_vec += a * b
        norm_a_vec = vmlaq_f64(norm_a_vec, a, a); // norm_a_vec += a * a
        norm_b_vec = vmlaq_f64(norm_b_vec, b, b); // norm_b_vec += b * b
    }

    // Add results of vector computation back into standard double variables
    double dp_arr[2], norm_a_arr[2], norm_b_arr[2];
    vst1q_f64(dp_arr, dp_vec);
    vst1q_f64(norm_a_arr, norm_a_vec);
    vst1q_f64(norm_b_arr, norm_b_vec);

    dot_product += dp_arr[0] + dp_arr[1];
    norm_a += norm_a_arr[0] + norm_a_arr[1];
    norm_b += norm_b_arr[0] + norm_b_arr[1];

    // If the array length is not even, we have one remaining value to process
    if(len % 2 != 0) {
        double a_val = array1[len-1];
        double b_val = array2[len-1];

        dot_product += a_val * b_val;
        norm_a += a_val * a_val;
        norm_b += b_val * b_val;
    }

    // Cosine similarity is dot product divided by product of norms (Lengths of array1 and array2)
    double cos_sim = dot_product / (sqrt(norm_a) * sqrt(norm_b));
    // Cosine distance is 1 - cosine similarity
    return 1.0 - cos_sim;
}

double euclidean_distance_avx(const double* a, const double* b, int n) {
	return 0;
}

double cosine_distance_avx(const double* a, const double* b, int n){
	return 0;
}

#else

double euclidean_distance_avx(const double* a, const double* b, int n) {
	return 0;
}

double cosine_distance_avx(const double* a, const double* b, int n){
	return 0;
}

double euclidean_distance_neon(double *array1, double *array2, int len){
	return 0;
}

double cosine_distance_neon(double *array1, double *array2, int len) {
	return 0;
}

#endif

*/
import "C"

import (
	"VreeDB/Vector"
	"crypto/rand"
	"fmt"
	"math"
	"runtime"
	"sync"
	"unsafe"
)

type Util struct {
}

// CollectionConfig is a struct to hold the configuration of a Collection
type CollectionConfig struct {
	Name             string
	VectorDimension  int
	DistanceFuncName string
	DiagonalLength   float64
}

// ResultSet is the result of a search
type ResultSet struct {
	Payload  *map[string]interface{}
	Distance float64
	Vector   *[]float64
	Id       string
}

// Utils is the main struct of the Utils
var Utils *Util

// init initializes the Util
func init() {
	Utils = &Util{}
}

// EuclideanDistance function calculates the Euclidean distance between two vectors
func (u *Util) EuclideanDistance(vector1, vector2 *Vector.Vector) (float64, error) {
	var sum float64
	for i := 0; i < vector1.Length; i++ {
		diff := vector1.Data[i] - vector2.Data[i]
		sum += diff * diff
	}
	return math.Sqrt(sum), nil
}

// EuclideanDistanceAVX256 calculates the Euclidean distance between two vectors using AVX256
func (u *Util) EuclideanDistanceAVX256(vector1, vector2 *Vector.Vector) (float64, error) {
	return float64(C.euclidean_distance_avx((*C.double)(unsafe.Pointer(&vector1.Data[0])), (*C.double)(unsafe.Pointer(&vector2.Data[0])), C.int(vector1.Length))), nil
}

// EuclideanDistanceNEON calculates the Euclidean distance between two vectors using ARM/NEON
func (u *Util) EuclideanDistanceNEON(vector1, vector2 *Vector.Vector) (float64, error) {
	return float64(C.euclidean_distance_neon((*C.double)(unsafe.Pointer(&vector1.Data[0])), (*C.double)(unsafe.Pointer(&vector2.Data[0])), C.int(vector1.Length))), nil
}

// CosineDistance function calculates the Cosine distance between two vectors
func (u *Util) CosineDistance(vector1, vector2 *Vector.Vector) (float64, error) {
	var sum, sum1, sum2 float64

	for _, value := range vector1.Data {
		sum1 += value * value
	}

	for _, value := range vector2.Data {
		sum2 += value * value
	}

	for i, value := range vector1.Data {
		sum += value * vector2.Data[i]
	}

	return 1 - (sum / (math.Sqrt(sum1) * math.Sqrt(sum2))), nil
}

// CosineDistanceAVX256 calculates the Cosine distance between two vectors using AVX256.
func (u *Util) CosineDistanceAVX256(vector1, vector2 *Vector.Vector) (float64, error) {
	return float64(C.cosine_distance_avx((*C.double)(unsafe.Pointer(&vector1.Data[0])), (*C.double)(unsafe.Pointer(&vector2.Data[0])), C.int(vector1.Length))), nil
}

// CosineDistanceNEON calculates the cosine distance between two vectors using ARM/NEON.
func (u *Util) CosineDistanceNEON(vector1, vector2 *Vector.Vector) (float64, error) {
	return float64(C.cosine_distance_neon((*C.double)(unsafe.Pointer(&vector1.Data[0])), (*C.double)(unsafe.Pointer(&vector2.Data[0])), C.int(vector1.Length))), nil
}

// FastSqrt is a faster implementation of the Sqrt function
func (u *Util) FastSqrt(x float64) float64 {
	i := math.Float64bits(x)
	i = 0x5fe6eb50c7b537a9 - (i >> 1)
	y := math.Float64frombits(i)
	return 1 / (y * (1.5 - (x*0.5)*y*y))
}

// GetMaxDimension returns the maximum value of two vectors
func (u *Util) GetMaxDimension(vector1, vector2 *Vector.Vector, wg *sync.WaitGroup) {
	defer wg.Done()
	for idx := range vector1.Data {
		if vector2.Data[idx] > vector1.Data[idx] {
			vector1.Data[idx] = vector2.Data[idx]
		}
	}
}

// GetMinDimension returns the minimum value of two vectors
func (u *Util) GetMinDimension(vector1, vector2 *Vector.Vector, wg *sync.WaitGroup) {
	defer wg.Done()
	for idx := range vector1.Data {
		if vector2.Data[idx] < vector1.Data[idx] {
			vector1.Data[idx] = vector2.Data[idx]
		}
	}
}

// CalculateDimensionDiff will calculate the difference between the max and min vectors
func (u *Util) CalculateDimensionDiff(dimension int, dimensionDiff, maxVector, minVector *Vector.Vector) {
	for i := 0; i < dimension; i++ {
		(*dimensionDiff).Data[i] = (*maxVector).Data[i] - (*minVector).Data[i]
	}
}

// Calculate the DiogonalLength of the Collection
func (u *Util) CalculateDiogonalLength(diagonalLength *float64, dimension int, dimensionDiff *Vector.Vector) {
	*diagonalLength = 0
	for i := 0; i < dimension; i++ {
		*diagonalLength += (*dimensionDiff).Data[i] * (*dimensionDiff).Data[i]
	}
}

// GetMemoryUsage returns the memory usage of the application
func (u *Util) GetMemoryUsage() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1024 / 1024
}

// GetAvailableRAM returns the available RAM
func (u *Util) GetAvailableRAM() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Sys) / 1024 / 1024
}

// Create a pseudo random UUID
func (u *Util) CreateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
