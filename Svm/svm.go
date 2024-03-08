package Svm

import (
	"VectoriaDB/Logger"
	"VectoriaDB/Vector"
	"math"
)

// The SVM struct will represent the SVM
type SVM struct {
	Alpha  []float64
	Bias   float64
	Data   []*Vector.Vector
	Kernel func([]float64, []float64) float64
	Degree int
}

// MultiClassSVM is a struct that holds multiple SVMs
type MultiClassSVM struct {
	Classifiers map[int]*SVM
	Name        string
	Training    bool
	Collection  string
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

	// Train the SVM
	for epoch := 0; epoch < epochs; epoch++ {
		for i := 0; i < n; i++ {
			sum := 0.0
			for j := 0; j < n; j++ {
				sum += svm.Alpha[j] * float64((*svm.Data[j].Payload)["Label"].(int)) * svm.Kernel(svm.Data[i].Data, svm.Data[j].Data)
			}
			if float64((*svm.Data[i].Payload)["Label"].(int))*(sum+svm.Bias) < 1 {
				svm.Alpha[i] += C
			}
		}
	}

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
func (svm *SVM) decisionFunction(x []float64) float64 {
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
		classes[(*point.Payload)["Label"].(int)] = true
	}

	for class := range classes {
		svm := &SVM{}
		modifiedData := make([]*Vector.Vector, len(data))
		for i, point := range data {
			if (*point.Payload)["Label"].(int) == class {
				modifiedData[i] = &Vector.Vector{Data: point.Data, Payload: &map[string]interface{}{"Label": 1}}
			} else {
				modifiedData[i] = &Vector.Vector{Data: point.Data, Payload: &map[string]interface{}{"Label": -1}}
			}
		}
		svm.Train(modifiedData, epochs, C, degree)
		mcs.Classifiers[class] = svm
	}
	// Log that the training is done
	mcs.Training = false
	Logger.Log.Log("Training of MultiClassSVM " + mcs.Name + " in Collection " + mcs.Collection + " done")
}

// Predict is a function that predicts the class of a given vector
func (mcs *MultiClassSVM) Predict(features []float64) int {
	var maxClass int
	maxScore := math.Inf(-1)

	for class, svm := range mcs.Classifiers {
		score := svm.decisionFunction(features)
		if score > maxScore {
			maxScore = score
			maxClass = class
		}
	}
	return maxClass
}

// NewMultiClassSVM Creates new MultiClassSVM
func NewMultiClassSVM(name string, collection string) *MultiClassSVM {
	return &MultiClassSVM{Name: name, Collection: collection}
}

/*
	var mcs MultiClassSVM
	mcs.Train(data, 10, 1.0, 3)

	testPoint := Point{Features: []float64{0.4, 0.6}}
	prediction := mcs.Predict(testPoint.Features)
	fmt.Printf("Vorhersage für Punkt (%.2f, %.2f): Klasse %d\n", testPoint.Features[0], testPoint.Features[1], prediction)
*/
