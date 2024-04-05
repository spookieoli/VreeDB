package AccessDataHUB

import (
	"sort"
	"sync"
	"time"
)

// ADPoint is a Datapoint that holds the data of accesses
type ADPoint struct {
	Type string
	Time time.Time
}

// AList is the struct that holds the AccessList
type AList struct {
	AccessList []ADPoint
	Mut        sync.RWMutex
	ReadChan   chan string
}

// IntervalSum is a struct that holds the sum of accesstypes
type IntervalSum struct {
	Type   string    `json:"type"`
	Period time.Time `json:"period"`
	Sum    int       `json:"sum"`
}

// AccessList is the global AccessList
var AccessList AList

// init initializes the AccessList
func init() {
	AccessList = AList{AccessList: make([]ADPoint, 0), Mut: sync.RWMutex{}, ReadChan: make(chan string, 1000)}
	AccessList.StartThread()
}

// StartThread will start the thread that reads from the ReadChan
func (al *AList) StartThread() {
	go al.writeFromChan()
}

// We will create a own goroutine that continuously reads from the ReadChan
func (al *AList) writeFromChan() {
	for {
		select {
		case data := <-al.ReadChan:
			al.Mut.Lock()
			al.AccessList = append(al.AccessList, ADPoint{Type: data, Time: time.Now()})
			al.Mut.Unlock()
		}
	}
}

// groupByIntervalAndType will group the ADPoints by Interval and Type
func (al *AList) groupByIntervalAndType() []IntervalSum {
	// Sortiere die ADPoints nach Zeit
	sort.Slice(al.AccessList, func(i, j int) bool {
		return al.AccessList[i].Time.Before(al.AccessList[j].Time)
	})

	intervalMap := make(map[string]map[time.Time]int)

	for _, point := range al.AccessList {
		interval := point.Time.Truncate(5 * time.Second)

		if _, ok := intervalMap[point.Type]; !ok {
			intervalMap[point.Type] = make(map[time.Time]int)
		}

		intervalMap[point.Type][interval] += 1
	}

	var result []IntervalSum

	for typ, intervals := range intervalMap {
		for interval, count := range intervals {
			result = append(result, IntervalSum{Type: typ, Period: interval, Sum: count})
		}
	}
	return result
}

// GetData will return the grouped data
func (al *AList) GetData() []IntervalSum {

	// Group the data
	al.Mut.RLock()
	defer al.Mut.RUnlock()
	data := al.groupByIntervalAndType()
	// Delete all data in the AccessList
	al.AccessList = make([]ADPoint, 0)
	// return the data
	return data
}
