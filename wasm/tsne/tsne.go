package tsne

// TSNE is a struct that holds the parameters for the t-SNE algorithm.
type TSNE struct {
	Perplexity             float64
	Theta                  float64
	MaxIter                int
	MaxIterWithoutProgress int
	Verbose                bool
	Data                   *[][]float64 // Data is a pointer to a 2D array (slice) of floats
}

// NewTSNE creates a new TSNE struct with the given parameters.
func NewTSNE(perplexity float64, theta float64, maxIter int, maxIterWithoutProgress int, verbose bool, data *[][]float64) *TSNE {
	return &TSNE{
		Perplexity:             perplexity,
		Theta:                  theta,
		MaxIter:                maxIter,
		MaxIterWithoutProgress: maxIterWithoutProgress,
		Verbose:                verbose,
		Data:                   data,
	}
}
