package NN

import (
	"VreeDB/FileMapper"
	"VreeDB/Logger"
	"VreeDB/Vector"
	"fmt"
	"math"
	"math/rand"
	"strings"
)

// Types *************************************

type Network struct {
	Layers         []Layer
	Loss           func([]float64, []float64) float64
	LossDerivative func([]float64, []float64) []float64
}

type Neuron struct {
	Weights []float64
	Bias    float64
	Output  float64
	Delta   float64
}

type Layer struct {
	Neurons        []Neuron
	ActivationName string
	Activation     ActivationFunc
	Derivative     DerivativeFunc
}

type ActivationFunc func(any) any
type DerivativeFunc func(any) any

// *****************************************

// NewNetwork creates a new network with the given layers
func NewNetwork(layers []Layer, lossfunction string) (*Network, error) {
	// Create Network
	n := &Network{}

	// Check every layer - set the activation function
	for i, layer := range layers {
		if strings.ToLower(layer.ActivationName) == "sigmoid" {
			layers[i].Activation = Sigmoid
			layers[i].Derivative = SigmoidDerivative
		} else if strings.ToLower(layer.ActivationName) == "tanh" {
			layers[i].Activation = Tanh
			layers[i].Derivative = TanhDerivative
		} else if strings.ToLower(layer.ActivationName) == "relu" {
			layers[i].Activation = ReLU
			layers[i].Derivative = ReLUDerivative
		} else if strings.ToLower(layer.ActivationName) == "softmax" {
			layers[i].Activation = Softmax
		} else {
			Logger.Log.Log("Unknown activation function: " + layer.ActivationName)
			return nil, fmt.Errorf("Unknown activation function: %s", layer.ActivationName)
		}
	}
	n.Layers = layers

	// Add Loss function
	switch strings.ToLower(lossfunction) {
	case "mse":
		n.Loss = n.MSE
		n.LossDerivative = n.MSEDerivative
	case "mae":
		n.Loss = n.MAE
		n.LossDerivative = n.MAEDerivative
	}
	return n, nil
}

// MSE is the mean squared error loss function
func (n *Network) MSE(outputs, targets []float64) float64 {
	sum := 0.0
	for i, output := range outputs {
		sum += math.Pow(output-targets[i], 2)
	}
	return sum / float64(len(outputs))
}

// MSEDerivative is the derivative of the mean squared error loss function
func (n *Network) MSEDerivative(outputs, targets []float64) []float64 {
	deltas := make([]float64, len(outputs))
	for i, output := range outputs {
		deltas[i] = 2 * (output - targets[i])
	}
	return deltas
}

// MAE is the mean absolute error loss function
func (n *Network) MAE(outputs, targets []float64) float64 {
	sum := 0.0
	for i, output := range outputs {
		sum += math.Abs(output - targets[i])
	}
	return sum / float64(len(outputs))
}

// MAEDerivative is the derivative of the mean absolute error loss function
func (n *Network) MAEDerivative(outputs, targets []float64) []float64 {
	deltas := make([]float64, len(outputs))
	for i, output := range outputs {
		if output > targets[i] {
			deltas[i] = 1
		} else {
			deltas[i] = -1
		}
	}
	return deltas
}

// Train - initializes the weights and biases and trains the network
func (n *Network) Train(trainingData [][]float64, targets [][]float64, epochs int, lr float64, batchSize int) {

	// Initialize the weights and biases
	for i := range n.Layers {
		inputLength := len(trainingData[0])
		if i > 0 {
			inputLength = len(n.Layers[i-1].Neurons)
		}

		for j := range n.Layers[i].Neurons {
			n.Layers[i].Neurons[j].Weights = make([]float64, inputLength)
			for k := range n.Layers[i].Neurons[j].Weights {
				n.Layers[i].Neurons[j].Weights[k] = rand.NormFloat64() * math.Sqrt(2.0/float64(inputLength)) // He initialization
			}
			n.Layers[i].Neurons[j].Bias = 0 // Initialize biases to zero
		}
	}

	// Trainloop
	for epoch := 0; epoch < epochs; epoch++ {
		// Split trainingData and targets into batches
		for i := 0; i < len(trainingData); i += batchSize {
			end := i + batchSize
			if end > len(trainingData) {
				end = len(trainingData)
			}
			batchData := trainingData[i:end]
			batchTargets := targets[i:end]

			// Train on batch
			totalLoss := 0.0
			for i, input := range batchData {
				output := n.Forward(input)
				n.Backpropagate(input, batchTargets[i], lr)
				totalLoss += n.Loss(output, batchTargets[i])
			}
			// Print loss every 10 epochs
			if epoch%10 == 0 {
				Logger.Log.Log("Epoch " + fmt.Sprint(epoch) + ", Loss: " + fmt.Sprint(totalLoss/float64(len(batchData))) + ", totalLoss: " + fmt.Sprint(totalLoss))
			}
		}
	}
}

// CreateTrainData creates the training data for the training
func (n *Network) CreateTrainData(vectors []*Vector.Vector) ([][]float64, [][]float64, error) {
	// Create Vars
	var x [][]float64
	var y [][]float64

	// Loop through the vectors
	for _, v := range vectors {
		// First we need to get the payload of the vector
		payload, err := FileMapper.Mapper.ReadPayload(v.PayloadStart, v.Collection)
		if err != nil {
			Logger.Log.Log("Error reading payload while creating NeuralNet Traindata: " + err.Error())
			return nil, nil, err
		}

		// if Label exists as key in payload
		if _, ok := (*payload)["Label"]; !ok {
			continue
		}

		// Label must be of type []float64
		if v, ok := (*payload)["Label"].([]interface{}); ok {
			//make float64 from int
			var label []float64
			for _, l := range v {
				if f, ok := l.(float64); ok {
					label = append(label, f)
				} else {
					continue
				}
			}
			y = append(y, label)
		} else {
			continue
		}

		// Add the vector to the training data
		x = append(x, v.Data)
	}

	// Check if the data is gt 0
	if len(x) == 0 {
		return nil, nil, fmt.Errorf("No NeuralNet Traindata created")
	} else {
		Logger.Log.Log("NeuralNet Traindata created successfully, x: " + fmt.Sprint(len(x)) + ", y: " + fmt.Sprint(len(y)))
	}
	return x, y, nil
}

// Predict - predicts the output for the given input
func (n *Network) Predict(input []float64) any {
	return n.Forward(input)
}

// Sigmoid Activation is used for hidden layers
func Sigmoid(x any) any {
	return 1 / (1 + math.Exp(-x.(float64)))
}

// SigmoidDerivative is the derivative of the sigmoid activation function
func SigmoidDerivative(x any) any {
	s := Sigmoid(x.(float64)).(float64)
	return s * (1 - s)
}

// Tanh Activation is used for hidden layers
func Tanh(x any) any {
	return math.Tanh(x.(float64))
}

// TanhDerivative is the derivative of the tanh activation function
func TanhDerivative(x any) any {
	return 1 - math.Pow(math.Tanh(x.(float64)), 2)
}

// RelU Activation is used for hidden layers
func ReLU(x any) any {
	if x.(float64) > 0 {
		return x
	}
	return 0
}

// ReLUDerivative is the derivative of the ReLU activation function
func ReLUDerivative(x any) any {
	if x.(float64) > 0 {
		return 1
	}
	return 0
}

// Softmax will be executed on the output layer
func Softmax(x any) any {
	max := x.([]float64)[0]
	for _, v := range x.([]float64) {
		if v > max {
			max = v
		}
	}

	sum := 0.0
	result := make([]float64, len(x.([]float64)))
	for i, v := range x.([]float64) {
		result[i] = math.Exp(v - max)
		sum += result[i]
	}

	for i := range result {
		result[i] /= sum
	}
	return result
}

// Forward pass through the layer
func (l *Layer) Forward(inputs []float64) []float64 {
	outputs := make([]float64, len(l.Neurons))
	if l.ActivationName == "softmax" {
		outputs = l.Activation(inputs).([]float64)
		for i, output := range outputs {
			l.Neurons[i].Output = output
		}
		return outputs
	} else {
		for i, neuron := range l.Neurons {
			sum := neuron.Bias
			for j, weight := range neuron.Weights {
				sum += inputs[j] * weight
			}

			var output float64

			if v, ok := l.Activation(sum).(int); ok {
				output = float64(v)
			} else if v, ok := l.Activation(sum).(float64); ok {
				output = v
			}

			l.Neurons[i].Output = output
			outputs[i] = output
		}
	}
	return outputs
}

// Forward feed forwards the inputs through the network
func (n *Network) Forward(inputs []float64) []float64 {
	for _, layer := range n.Layers {
		inputs = layer.Forward(inputs)
	}
	return inputs
}

// Backpropagate - backpropagates the error through the network
func (n *Network) Backpropagate(inputs, targets []float64, lr float64) {
	outputs := n.Forward(inputs)
	deltas := n.LossDerivative(outputs, targets)

	// Backwards pass
	for i := len(n.Layers) - 1; i >= 0; i-- {
		layer := &n.Layers[i]
		// if it is not the output layer, we need to calculate the deltas for the next layer
		var newDeltas []float64
		if i > 0 {
			newDeltas = make([]float64, len(n.Layers[i-1].Neurons))
		} else {
			newDeltas = make([]float64, len(inputs))
		}

		inputsForLayer := inputs // Inputs für die allererste Schicht
		if i > 0 {
			inputsForLayer = make([]float64, len(n.Layers[i-1].Neurons))
			for j := range n.Layers[i-1].Neurons {
				inputsForLayer[j] = n.Layers[i-1].Neurons[j].Output
			}
		}

		for j, neuron := range layer.Neurons {
			errorTerm := deltas[j]
			if layer.ActivationName != "softmax" {
				// Check if the derivative is int or float64
				if v, ok := layer.Derivative(neuron.Output).(int); ok {
					errorTerm *= float64(v)
				} else if v, ok := layer.Derivative(neuron.Output).(float64); ok {
					errorTerm *= v
				}
			}
			for k, input := range inputsForLayer {
				grad := errorTerm * input
				neuron.Weights[k] -= lr * grad
				if i > 0 { // Accumulate the gradients for the next layer's deltas
					newDeltas[k] += neuron.Weights[k] * errorTerm
				}
			}
			neuron.Bias -= lr * errorTerm
		}

		// for the first layer, we need to set the deltas
		if i > 0 {
			deltas = newDeltas
		}
	}
}
