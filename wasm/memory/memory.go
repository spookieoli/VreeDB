package memory

import "syscall/js"

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
func NewTSNE(perplexity float64, theta float64, maxIter int, maxIterWithoutProgress int, verbose bool, d *js.Value) *TSNE {
	tsne := &TSNE{}
	data := tsne.js2go(d)
	tsne.Perplexity = perplexity
	tsne.Theta = theta
	tsne.MaxIter = maxIter
	tsne.MaxIterWithoutProgress = maxIterWithoutProgress
	tsne.Verbose = verbose
	tsne.Data = data
	return tsne
}

// js2go converts a JavaScript 2D array to a Go 2D array.
func (t *TSNE) js2go(d *js.Value) *[][]float64 {
	var data [][]float64
	length := d.Get("length").Int()
	for i := 0; i < length; i++ {
		var row []float64
		for j := 0; j < d.Index(i).Get("length").Int(); j++ {
			row = append(row, d.Index(i).Index(j).Float())
		}
		data = append(data, row)
	}
	return &data
}
