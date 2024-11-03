package memory

import (
	"fmt"
	"math"
	"syscall/js"
)

// TSNE is a struct that holds the parameters for the t-SNE algorithm.
type TSNE struct {
	Perplexity             float64
	Theta                  float64
	MaxIter                int
	MaxIterWithoutProgress int
	Verbose                bool
	learningRate           float64
	Data                   *[][]float64 // Data is a pointer to a 2D array (slice) of floats
	targetDim              int
}

// NewTSNE creates a new TSNE struct with the given parameters.
func NewTSNE(perplexity float64, theta float64, maxIter int, maxIterWithoutProgress int, verbose bool, learningRate float64, targetDim int, d *js.Value) (*TSNE, error) {
	// Create the Data field from the input data
	tsne := &TSNE{}
	data, err := tsne.js2go(d)

	// Check if there was an error converting the data
	if err != nil {
		return nil, err
	}

	// Set the parameters for the t-SNE algorithm
	tsne.Perplexity = perplexity
	tsne.Theta = theta
	tsne.MaxIter = maxIter
	tsne.MaxIterWithoutProgress = maxIterWithoutProgress
	tsne.Verbose = verbose
	tsne.learningRate = learningRate
	tsne.targetDim = targetDim // TargetDim is normally 2D for VreeDB
	tsne.Data = data
	return tsne, err
}

// js2go converts a JavaScript 2D array to a Go 2D array.
func (t *TSNE) js2go(d *js.Value) (*[][]float64, error) {
	// Set the data field from the input data
	var data [][]float64
	length := d.Get("length").Int()
	llength := d.Index(0).Get("length").Int()

	// Check if the length of the vector is less than 3
	if llength <= 2 {
		return nil, fmt.Errorf("length of vector Dimension may not be under 2")
	}

	// Convert the 2D array to a Go 2D array
	for i := 0; i < length; i++ {
		var row []float64
		for j := 0; j < d.Index(i).Get("length").Int(); j++ {
			row = append(row, d.Index(i).Index(j).Float())
		}
		data = append(data, row)
	}
	return &data, nil
}

// execute will execute the t-SNE algorithm.
func (t *TSNE) execute() *[][]float64 {
	// Perform the t-SNE algorithm
	return t.Data
}

// pairwiseDistances calculates the pairwise distances between the data points.
func (t *TSNE) pairwiseDistances(data *[][]float64) (matrix *[][]float64) {
	n := len(*data)
	matrix = &[][]float64{} // initialize the matrix - matrix is a pointer and should not be nil

	// Calculate the pairwise distances
	for i := 0; i < n; i++ {
		row := make([]float64, n)
		for j := range row {
			row[j] = t.euclideanDistance((*data)[i], (*data)[j])
		}
		*matrix = append(*matrix, row)
	}
	return
}

// euclideanDistance will calculate the Euclidean distance between two points.
func (t *TSNE) euclideanDistance(x, y []float64) float64 {
	// Calculate the Euclidean distance
	sum := 0.0
	for idx, i := range x {
		sum += math.Pow(i-y[idx], 2)
	}
	return math.Sqrt(sum)
}

// gaussKernel calculates the Gaussian kernel for the t-SNE algorithm.
func (t *TSNE) gaussKernel(dist, sigma *float64) *float64 {
	r := math.Exp(-*dist * *dist / (2 * *sigma * *sigma))
	return &r
}

// kullbackLeiblerDivergence calculates the Kullback-Leibler divergence between two probability distributions.
func kullbackLeiblerDivergence(P, Q []float64) (*float64, error) {
	if len(P) != len(Q) {
		return nil, fmt.Errorf("length of P and Q should be equal")
	}

	// Calculate the Kullback-Leibler divergence
	klDiv := 0.0
	for i := range P {
		// Some checks
		if P[i] == 0 {
			continue // We will ignore the 0 values
		}
		if Q[i] == 0 {
			return nil, fmt.Errorf("Q[%d] is 0", i)
		}
		// Add the KL divergence
		klDiv += P[i] * math.Log(P[i]/Q[i])
	}

	// Check if the KL divergence is negative
	if klDiv < 0 {
		klDiv = 0
	}

	// Return the KL divergence
	return &klDiv, nil
}
