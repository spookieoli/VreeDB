package Utils

import (
	"VreeDB/Vector"
	"crypto/rand"
	"fmt"
	"math"
	"runtime"
	"sync"
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
