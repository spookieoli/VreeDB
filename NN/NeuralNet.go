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
func NewNetwork(layers []Layer) (*Network, error) {
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
	return &Network{Layers: layers}, nil
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

// Train - initializes the weights and biases and trains the network
func (n *Network) Train(trainingData [][]float64, targets [][]float64, epochs int, lr float64) {

	// Initialize the weights and biases
	for i := range n.Layers {
		for j := range n.Layers[i].Neurons {
			n.Layers[i].Neurons[j].Weights = make([]float64, len(trainingData[0]))
			for k := range n.Layers[i].Neurons[j].Weights {
				n.Layers[i].Neurons[j].Weights[k] = rand.Float64()
			}
			n.Layers[i].Neurons[j].Bias = rand.Float64()
		}
	}

	// Trainloop
	for epoch := 0; epoch < epochs; epoch++ {
		totalLoss := 0.0
		for i, input := range trainingData {
			output := n.Forward(input)
			n.Backpropagate(input, targets[i], lr)
			totalLoss += n.MSE(output, targets[i])
		}
		if epoch%1000 == 0 {
			fmt.Printf("Epoch %d, Loss: %.4f\n", epoch, totalLoss/float64(len(trainingData)))
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
		switch v := (*payload)["Label"].(type) {
		case []float64:
			// Add the vector to the target data
			y = append(y, v)
		default:
			continue
		}

		// Add the vector to the training data
		x = append(x, v.Data)
	}

	// Check if the data is gt 0
	if len(x) == 0 {
		Logger.Log.Log("No NeuralNet Traindata created")
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
			output := l.Activation(sum)
			l.Neurons[i].Output = output.(float64)
			outputs[i] = output.(float64)
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

// Backpropagate backpropages the error through the network
func (n *Network) Backpropagate(inputs, targets []float64, lr float64) {
	outputs := n.Forward(inputs)
	deltas := n.LossDerivative(outputs, targets)

	// Backwards pass
	for i := len(n.Layers) - 1; i >= 0; i-- {
		layer := &n.Layers[i]
		inputs := inputs
		if i > 0 {
			inputs = make([]float64, len(n.Layers[i-1].Neurons))
			for j := range n.Layers[i-1].Neurons {
				inputs[j] = n.Layers[i-1].Neurons[j].Output
			}
		}
		for j, neuron := range layer.Neurons {
			errors := deltas[j]
			if layer.ActivationName != "softmax" {
				errors *= layer.Derivative(neuron.Output).(float64)
			}
			for k := range neuron.Weights {
				neuron.Weights[k] -= lr * errors * inputs[k]
				deltas[k] += errors * neuron.Weights[k]
			}
			neuron.Bias -= lr * errors
		}
	}
}

// TODO: an routing anbauen, Function um vectoren und payload (label) zu übergeben