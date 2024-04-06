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

// DeleteGT10 will delete all the ADPoints that are older than 10 Minutes
func (al *AList) DeleteGT10() {
	now := time.Now()
	for i := 0; i < len(al.AccessList); i++ {
		if now.Sub(al.AccessList[i].Time) > 10*time.Minute {
			al.AccessList = append(al.AccessList[:i], al.AccessList[i+1:]...)
			i--
		}
	}
}

// We will create a own goroutine that continuously reads from the ReadChan
func (al *AList) writeFromChan() {
	for {
		select {
		case data := <-al.ReadChan:
			al.Mut.Lock()
			al.AccessList = append(al.AccessList, ADPoint{Type: data, Time: time.Now()})
			al.DeleteGT10()
			al.Mut.Unlock()
		}
	}
}

// groupByIntervalAndType will group the ADPoints by Interval and Type
func (al *AList) groupByIntervalAndType() []IntervalSum {
	// Vars
	var tempList []ADPoint
	// now minus 5 seconds
	now := time.Now().Add(-5 * time.Second)

	al.Mut.Lock()
	defer al.Mut.Unlock()
	// Sortiere die ADPoints nach Zeit
	sort.Slice(al.AccessList, func(i, j int) bool {
		return al.AccessList[i].Time.Before(al.AccessList[j].Time)
	})

	// create intervalmap
	intervalMap := make(map[string]map[time.Time]int)

	for _, point := range al.AccessList {
		interval := point.Time.Truncate(5 * time.Second)

		// Is the interval completed?
		if interval.After(now) {
			tempList = append(tempList, point)
			continue
		}

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

	// Clear the AccessList and insert all the points that are in the tempList
	al.AccessList = tempList
	return result
}

// GetData will return the grouped data
func (al *AList) GetData() []IntervalSum {
	// Group the data
	data := al.groupByIntervalAndType()
	// return the data
	return data
}
