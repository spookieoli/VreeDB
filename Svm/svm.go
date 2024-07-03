package Svm

import (
	"VreeDB/FileMapper"
	"VreeDB/Logger"
	"VreeDB/Vector"
	"context"
	"fmt"
	"math"
	"runtime"
	"sync"
)

// epsilon due to the way floats are represented in memory, it is not possible to compare two floats directly
const epsilon float64 = 0.00000001

// The SVM struct will represent the SVM
type SVM struct {
	Alpha   []float64
	Bias    float64
	Data    []*Vector.Vector
	Kernel  func([]float64, []float64) float64
	Degree  int
	idxChan chan IdxChan
	Ctx     context.Context
	Cancel  context.CancelFunc
}

// MultiClassSVM is a struct that holds multiple SVMs
type MultiClassSVM struct {
	Classifiers map[int]*SVM
	Name        string
	Training    bool
	Collection  string
}

// Struct to communicate with the go routines of training
type IdxChan struct {
	Idx    int
	IdxCol int
	Sum    *float64
	Wg     *sync.WaitGroup
	data   *Vector.Vector
	Mut    *sync.Mutex
}

// polynomialKernel is a function that calculates the polynomial kernel
func polynomialKernel(x, y []float64, degree int) float64 {
	dot := 0.0
	for i := 0; i < len(x); i++ {
		dot += x[i] * y[i]
	}
	return math.Pow(dot+1, float64(degree))
}

// Train is a function that trains the SVM
func (svm *SVM) Train(data []*Vector.Vector, epochs int, C float64, degree int) {
	svm.Data = data
	svm.Degree = degree
	n := len(data)
	svm.Alpha = make([]float64, n)
	svm.Kernel = func(x, y []float64) float64 {
		return polynomialKernel(x, y, svm.Degree)
	}

	mut := sync.Mutex{}

	// Train the SVM
	for epoch := 0; epoch < epochs; epoch++ {
		Logger.Log.Log("Epoch "+fmt.Sprint(epoch), "INFO")
		for i := 0; i < n; i++ {
			sumArr := make([]float64, n)
			// Create a wait group to wait for all go routines to finish
			wg := sync.WaitGroup{}
			for j := 0; j < n; j++ {
				wg.Add(1)
				svm.idxChan <- IdxChan{Idx: i, IdxCol: j, Sum: &sumArr[i], Wg: &wg, Mut: &mut}
			}
			// wait for the go routines to finish
			wg.Wait()

			// summarize the sumArr
			sum := 0.0
			for _, sum := range sumArr {
				sum += sum
			}

			// Update the Alpha
			if float64((*svm.Data[i].Payload)["Label"].(int))*(sum+svm.Bias) < 1-epsilon {
				svm.Alpha[i] += C
			}
		}
		// Delete the Mutex
		Logger.Log.Log("Epoch "+fmt.Sprint(epoch)+" done", "INFO")
	}

	// End the go routines
	svm.Cancel()

	// Calculate the bias
	sum := 0.0
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			sum += svm.Alpha[j] * float64((*svm.Data[j].Payload)["Label"].(int)) * svm.Kernel(svm.Data[i].Data, svm.Data[j].Data)
		}
	}
	svm.Bias = float64((*svm.Data[0].Payload)["Label"].(int)) - sum
}

// decisionFunction is a function that calculates the decision function
func (svm *SVM) decisionFunction(x []float64, collection string) float64 {
	sum := 0.0
	for i := 0; i < len(svm.Data); i++ {
		sum += svm.Alpha[i] * float64((*svm.Data[i].Payload)["Label"].(int)) * svm.Kernel(x, svm.Data[i].Data)
	}
	return sum
}

// Train is a function that trains the MultiClassSVM
func (mcs *MultiClassSVM) Train(data []*Vector.Vector, epochs int, C float64, degree int) {
	mcs.Classifiers = make(map[int]*SVM)
	classes := make(map[int]bool)
	for _, point := range data {
		m, err := FileMapper.Mapper.ReadPayload(point.PayloadStart, mcs.Collection)
		if err != nil {
			Logger.Log.Log("Error reading payload: "+err.Error(), "INFO")
			return
		}
		point.Payload = m
		classes[int((*point.Payload)["Label"].(float64))] = true
	}

	for class := range classes {
		// create a new SVM
		svm := &SVM{idxChan: make(chan IdxChan, len(data))}
		ctx, cancel := context.WithCancel(context.Background())
		svm.Ctx = ctx
		svm.Cancel = cancel
		svm.StartThreads() // Will start the go routines

		// Create a new slice with the modified data
		modifiedData := make([]*Vector.Vector, len(data))
		for i, point := range data {
			if int((*point.Payload)["Label"].(float64)) == class {
				modifiedData[i] = &Vector.Vector{Data: point.Data, Payload: &map[string]interface{}{"Label": 1}, PayloadStart: point.PayloadStart}
			} else {
				modifiedData[i] = &Vector.Vector{Data: point.Data, Payload: &map[string]interface{}{"Label": -1}, PayloadStart: point.PayloadStart}
			}
		}
		Logger.Log.Log("Training SVM for class "+fmt.Sprint(class), "INFO")
		svm.Train(modifiedData, epochs, C, degree)
		Logger.Log.Log("Training done for class "+fmt.Sprint(class), "INFO")
		mcs.Classifiers[class] = svm
		// Clear Context and Cancel
		svm.Ctx = nil
		svm.Cancel = nil
	}

	// Set all the data items Payload to nil
	for _, point := range data {
		point.Payload = nil
	}

	// Log that the training is done
	mcs.Training = false
	Logger.Log.Log("Training of MultiClassSVM "+mcs.Name+" in Collection "+mcs.Collection+" done", "INFO")
}

// Predict is a function that predicts the class of a given vector
func (mcs *MultiClassSVM) Predict(features []float64) any {
	var maxClass int
	maxScore := math.Inf(-1)

	for class, svm := range mcs.Classifiers {
		// if not present - set the kernel to the polynomial kernel
		if svm.Kernel == nil {
			svm.Kernel = func(x, y []float64) float64 {
				return polynomialKernel(x, y, svm.Degree)
			}
		}
		// calculate the score
		score := svm.decisionFunction(features, mcs.Collection)
		if score > maxScore {
			maxScore = score
			maxClass = class
		}
	}
	return maxClass
}

// StartThreads is a function that starts the go routines
func (svm *SVM) StartThreads() {
	// Start cpu cores -1 go routines
	for i := 0; i < runtime.NumCPU()/2; i++ {
		go func() {
			for {
				select {
				case idxchan := <-svm.idxChan:
					sum := 0.0
					sum = svm.Alpha[idxchan.IdxCol] * float64((*svm.Data[idxchan.IdxCol].Payload)["Label"].(int)) * svm.Kernel(svm.Data[idxchan.Idx].Data, svm.Data[idxchan.IdxCol].Data)
					idxchan.Mut.Lock()
					*idxchan.Sum += sum
					idxchan.Mut.Unlock()
					idxchan.Wg.Done()
				case <-svm.Ctx.Done():
					return
				}
			}
		}()
	}
}

// NewMultiClassSVM Creates new MultiClassSVM
func NewMultiClassSVM(name string, collection string) *MultiClassSVM {
	return &MultiClassSVM{Name: name, Collection: collection}
}
