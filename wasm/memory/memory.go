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
}

// NewTSNE creates a new TSNE struct with the given parameters.
func NewTSNE(perplexity float64, theta float64, maxIter int, maxIterWithoutProgress int, verbose bool, learningRate float64, d *js.Value) (error, *TSNE) {
	// Create the Data field from the input data
	tsne := &TSNE{}
	err, data := tsne.js2go(d)

	// Check if there was an error converting the data
	if err != nil {
		return err, nil
	}

	// Set the parameters for the t-SNE algorithm
	tsne.Perplexity = perplexity
	tsne.Theta = theta
	tsne.MaxIter = maxIter
	tsne.MaxIterWithoutProgress = maxIterWithoutProgress
	tsne.Verbose = verbose
	tsne.learningRate = learningRate
	tsne.Data = data
	return nil, tsne
}

// js2go converts a JavaScript 2D array to a Go 2D array.
func (t *TSNE) js2go(d *js.Value) (error, *[][]float64) {
	// Set the data field from the input data
	var data [][]float64
	length := d.Get("length").Int()
	llength := d.Index(0).Get("length").Int()

	// Check if the length of the vector is less than 3
	if llength <= 3 {
		return fmt.Errorf("length of vector Dimension may not be under 3"), nil
	}

	// Convert the 2D array to a Go 2D array
	for i := 0; i < length; i++ {
		var row []float64
		for j := 0; j < d.Index(i).Get("length").Int(); j++ {
			row = append(row, d.Index(i).Index(j).Float())
		}
		data = append(data, row)
	}
	return nil, &data
}

// execute will execute the t-SNE algorithm.
func (t *TSNE) execute() *[][]float64 {
	// Perform the t-SNE algorithm
	return t.Data
}

// kullbackLeiblerDivergence calculates the Kullback-Leibler divergence between two probability distributions.
func kullbackLeiblerDivergence(P, Q []float64) (error, *float64) {
	if len(P) != len(Q) {
		return fmt.Errorf("length of P and Q should be equal"), nil
	}

	// Calculate the Kullback-Leibler divergence
	klDiv := 0.0
	for i := range P {
		if P[i] == 0 {
			continue // We will ignore the 0 values
		}
		if Q[i] == 0 {
			return fmt.Errorf("Q[%d] is 0", i), nil
		}
		klDiv += P[i] * math.Log(P[i]/Q[i])
	}
	return nil, &klDiv
}
