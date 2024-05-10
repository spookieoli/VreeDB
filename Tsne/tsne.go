package Tsne

import (
	"VreeDB/Logger"
	"VreeDB/Utils"
	"VreeDB/Vector"
	"math"
	"math/rand"
	"runtime"
	"sync"
)

// TSNE is a struct that represents the t-Distributed Stochastic Neighbor Embedding algorithm.
// It is used for dimensionality reduction and visualization of high-dimensional data.
//
// The TSNE struct has the following fields:
// - perplexity: The perplexity hyperparameter for the TSNE algorithm.
// - learningRate: The learning rate or step size for the TSNE algorithm.
// - maxIterations: The maximum number of iterations for the TSNE algorithm.
type TSNE struct {
	perplexity    float64
	learningRate  float64
	dimensions    int // represent the number of dimensions in the output space
	maxIterations int
	embeddings    []*Vector.Vector
	Collection    string
	Chan          chan *ThreadpoolData
}

// ThreadpoolData carries the vars to calculate Gradients
type ThreadpoolData struct {
	embedding1, embedding2 *Vector.Vector
	dist, sum              *float64
	wg                     *sync.WaitGroup
}

// NewTSNE is a function that creates a new instance of the TSNE struct with the provided parameters.
// It takes the perplexity, learning rate, maximum iterations, and dimensions as inputs.
// It returns a pointer to the newly created TSNE struct.
//
// Example usage:
// tsne := NewTSNE(30.0, 0.1, 1000, 2)
// tsne.PerformTSNE()
//
// Parameters:
// - perplexity: The perplexity hyperparameter for the TSNE algorithm.
// - learningRate: The learning rate or step size for the TSNE algorithm.
// - maxIterations: The maximum number of iterations for the TSNE algorithm.
// - dimensions: The number of dimensions in the output space.
//
// Returns:
// A pointer to a new TSNE instance.
func NewTSNE(learninrate float64, maxiterations, dimensions int, collection string) *TSNE {
	return &TSNE{learningRate: learninrate, maxIterations: maxiterations,
		dimensions: dimensions, Collection: collection, Chan: make(chan *ThreadpoolData, 100)}
}

// PerformTSNE performs the t-SNE algorithm.
// It updates the state of the TSNE struct based on the input data.
// It returns a Vector that represents the dimensionality-reduced data.
// The returned Vector contains the data points in the output space.
func (t *TSNE) PerformTSNE(data []*Vector.Vector) ([]*Vector.Vector, error) {
	// Create random embeddings
	embeddings := make([]*Vector.Vector, len(data))
	for i := range embeddings {
		embeddings[i] = &Vector.Vector{}
		embeddings[i].Data = make([]float64, t.dimensions)
		for k := 0; k < t.dimensions; k++ {
			embeddings[i].Data[k] = rand.Float64()
		}
	}
	t.embeddings = embeddings

	// TSNE Iterations
	for i := 0; i < t.maxIterations; i++ {
		gradients, err := t.computeGradients(data)
		if err != nil {
			// handle error
			return nil, err
		}

		// update Embeddings
		t.updateEmbeddings(t.embeddings, gradients)
	}
	// CLose the
	close(t.Chan)
	return t.embeddings, nil
}

// computeGradients computes the gradients for the TSNE algorithm based on the input data.
// It takes a slice of Vector data as input and returns a 2D slice of floats representing the gradients.
// The returned gradients represent the update values for each dimension of each embedding vector.
// The algorithm performs the following steps:
//   - It initializes a slice of embeddings with random values.
//   - It calculates the pairwise distances and affinities between the embeddings.
//   - It computes the gradients based on the differences in affinities and updates the gradients slice.
//
// The function returns an error if there is any error in calculating distances or affinities.
// Otherwise, it returns the calculated gradients slice.
func (t *TSNE) computeGradients(data []*Vector.Vector) ([][]float64, error) {
	n := len(data)

	// Create the dist and gradients
	gradients := make([][]float64, n)
	dist := make([][]float64, n)
	sum := make([]float64, n)

	// Use only one loop for the creation of the slices
	for i := range gradients {
		gradients[i] = make([]float64, t.dimensions)
		dist[i] = make([]float64, n)
	}

	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i != j && j < len(data[i].Data) {
				// Threaded calculation
				wg.Add(1)
				t.Chan <- &ThreadpoolData{dist: &dist[i][j], embedding1: t.embeddings[i], embedding2: t.embeddings[j], sum: &sum[i]}
			}
		}
	}
	// Wait for calculations to be done
	wg.Done()

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i != j && j < len(data[i].Data) {
				// Calculate distance and affine weighted probability
				num := 1.0 / (1.0 + math.Pow(dist[i][j], 2))
				pij := data[i].Data[j]
				qij := num / sum[i]

				// calculate gradients
				for d := 0; d < t.dimensions; d++ {
					gradients[i][d] += 4.0 * (pij - qij) * (t.embeddings[i].Data[d] - t.embeddings[j].Data[d])
				}
			}
		}
	}
	return gradients, nil
}

// Threadpool starts the Threadpool
func (t *TSNE) Threadpool() {
	for i := 0; i < runtime.NumCPU()/2; i++ {
		go func() {
			for data := range t.Chan {
				t.CalculateSums(data)
				data.wg.Done()
			}
		}()
	}
}

// CalculateSums calculates the sums for the CragdientUpdate
func (t *TSNE) CalculateSums(data *ThreadpoolData) {
	var err error
	*data.dist, err = Utils.Utils.EuclideanDistance(data.embedding1, data.embedding2)
	if err != nil {
		Logger.Log.Log(err.Error())
		return
	}
	*data.sum += 1.0 / (1.0 + math.Pow(*data.dist, 2))
}

// updateEmbeddings updates the embedding vectors based on the calculated gradients.
// It takes a slice of embedding vectors and a 2D slice of gradients as input.
// For each embedding vector, it updates each dimension of the vector using the learning rate and corresponding gradient.
// The updated embedding vectors are modified in-place.
func (t *TSNE) updateEmbeddings(embeddings []*Vector.Vector, gradients [][]float64) {
	for i := range embeddings {
		// optimize Memory access by first get data to scope
		embedding := embeddings[i].Data
		gradient := gradients[i]
		for d := 0; d < t.dimensions; d++ {
			embedding[d] += t.learningRate * gradient[d]
		}
	}
}
