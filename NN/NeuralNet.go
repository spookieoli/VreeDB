package NN

import (
	"math"
)

type ActivationFunc func(float64) any
type DerivativeFunc func(float64) any

// Activationfunctions
func Sigmoid(x float64) float64 {
	return 1 / (1 + math.Exp(-x))
}

func SigmoidDerivative(x float64) float64 {
	s := Sigmoid(x)
	return s * (1 - s)
}

func Tanh(x float64) float64 {
	return math.Tanh(x)
}

func TanhDerivative(x float64) float64 {
	return 1 - math.Pow(math.Tanh(x), 2)
}

func ReLU(x float64) float64 {
	if x > 0 {
		return x
	}
	return 0
}

func ReLUDerivative(x float64) float64 {
	if x > 0 {
		return 1
	}
	return 0
}

// Softmax will be executed on the output layer
func Softmax(x []float64) []float64 {
	max := x[0]
	for _, v := range x {
		if v > max {
			max = v
		}
	}

	sum := 0.0
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = math.Exp(v - max)
		sum += result[i]
	}

	for i := range result {
		result[i] /= sum
	}
	return result
}

type Neuron struct {
	Weights []float64
	Bias    float64
	Output  float64
	Delta   float64
}

type Layer struct {
	Neurons    []Neuron
	Activation ActivationFunc
	Derivative DerivativeFunc
}

func (l *Layer) Forward(inputs []float64) []float64 {
	outputs := make([]float64, len(l.Neurons))
	for i, neuron := range l.Neurons {
		sum := neuron.Bias
		for j, weight := range neuron.Weights {
			sum += inputs[j] * weight
		}
		output := l.Activation(sum)
		l.Neurons[i].Output = output
		outputs[i] = output
	}
	return outputs
}

type Network struct {
	Layers         []Layer
	Loss           func([]float64, []float64) float64
	LossDerivative func([]float64, []float64) []float64
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
			if layer.Activation != Softmax {
				error *= layer.Derivative(neuron.Output)
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

// Hier kann eine quadratische Verlustfunktion implementiert werden
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
		{Neurons: make([]Neuron, 10), Activation: Sigmoid, Derivative: SigmoidDerivative},
		{Neurons: make([]Neuron, 3), Activation: Softmax, Derivative: nil}, // This want work - Softmax has []float64
	}
	network := NewNetwork(layers)
	//...
}
