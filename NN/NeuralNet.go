package NN

import (
	"fmt"
	"math"
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

// Activationfunctions
func Sigmoid(x any) any {
	return 1 / (1 + math.Exp(-x.(float64)))
}

func SigmoidDerivative(x any) any {
	s := Sigmoid(x.(float64)).(float64)
	return s * (1 - s)
}

func Tanh(x any) any {
	return math.Tanh(x.(float64))
}

func TanhDerivative(x any) any {
	return 1 - math.Pow(math.Tanh(x.(float64)), 2)
}

func ReLU(x any) any {
	if x.(float64) > 0 {
		return x
	}
	return 0
}

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

func (net *Network) Forward(inputs []float64) []float64 {
	for _, layer := range net.Layers {
		inputs = layer.Forward(inputs)
	}
	return inputs
}

func (net *Network) Backpropagate(inputs, targets []float64, lr float64) {
	outputs := net.Forward(inputs)
	deltas := net.LossDerivative(outputs, targets)

	// Rückwärts durch das Netz
	for i := len(net.Layers) - 1; i >= 0; i-- {
		layer := &net.Layers[i]
		inputs := inputs
		if i > 0 {
			inputs = make([]float64, len(net.Layers[i-1].Neurons))
			for j := range net.Layers[i-1].Neurons {
				inputs[j] = net.Layers[i-1].Neurons[j].Output
			}
		}
		for j, neuron := range layer.Neurons {
			error := deltas[j]
			if layer.ActivationName != "softmax" {
				error *= layer.Derivative(neuron.Output).(float64)
			}
			for k := range neuron.Weights {
				neuron.Weights[k] -= lr * error * inputs[k]
				deltas[k] += error * neuron.Weights[k]
			}
			neuron.Bias -= lr * error
		}
	}
}

func NewNetwork(layers []Layer) *Network {
	return &Network{Layers: layers}
}

func MSE(outputs, targets []float64) float64 {
	sum := 0.0
	for i, output := range outputs {
		sum += math.Pow(output-targets[i], 2)
	}
	return sum / float64(len(outputs))
}

func MSEDerivative(outputs, targets []float64) []float64 {
	deltas := make([]float64, len(outputs))
	for i, output := range outputs {
		deltas[i] = 2 * (output - targets[i])
	}
	return deltas
}

// Sample
func test() {
	layers := []Layer{
		{Neurons: make([]Neuron, 10), ActivationName: "sigmoid", Activation: Sigmoid, Derivative: SigmoidDerivative},
		{Neurons: make([]Neuron, 3), ActivationName: "softmax", Activation: Softmax, Derivative: nil}, // This want work - Softmax has []float64
	}
	network := NewNetwork(layers)
	fmt.Println(network)
	//...
}
