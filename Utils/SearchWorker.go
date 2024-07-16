package Utils

// SearchWorker represents the worker used for searching.
type SearchWorker struct {
	Chan        chan *SearchUnit
	WorkerCount int
}

// Searcher is package global
var Searcher *SearchWorker

// init initializes the Searcher variable with a new SearchWorker instance.
func init() {
	Searcher = &SearchWorker{Chan: make(chan *SearchUnit, 100000), WorkerCount: 0}
}

func


