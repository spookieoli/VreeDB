package Utils

import (
	"VreeDB/ArgsParser"
	"fmt"
)

// SearchWorker represents the worker used for searching.
type SearchWorker struct {
	schan       chan *SearchData
	WorkerCount int
}

// Searcher is package global
var Searcher *SearchWorker

// init initializes the Searcher variable with a new SearchWorker instance.
func init() {
	Searcher = &SearchWorker{schan: make(chan *SearchData, 100000), WorkerCount: 0}
	Searcher.Start()
	fmt.Println("Search Workers ready")
}

// GetChan returns the channel of the SearchWorker.
func (sw *SearchWorker) GetChan() chan *SearchData {
	return sw.schan
}

// Start starts the search by creating worker goroutines that consume jobs from the channel.
// Each worker goroutine executes the NearestNeighbors method on the data received from the channel,
// and then releases the WaitGroup of the SearchUnit.
func (sw *SearchWorker) Start() {
	for i := 0; i < *ArgsParser.Ap.SearchThreads; i++ {
		go func() {
			for data := range sw.schan {
				data.SU.NearestNeighbors(data.Node, data.Target, data.Queue, data.DistanceFunc, data.DimensionDiff)
				data.SU.releaseWaitGroup()
			}
		}()
	}
}
