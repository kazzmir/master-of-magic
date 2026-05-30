package neural

import (
    "math"
    "math/rand/v2"
)

// create a neural network with a feed forward pass and back propagation
// the network should take in a state vector extracted from the game state
// and output a set of strategy weights that a lower tier AI can use to make decisions
// the reward values are used to train the network using back propagation, where
// the reward values are a vector, some of the values are used for multiple network outputs

type ActivationFunc func(weights []float64, inputs []float64, bias float64) float64

func Sigmoid(weights []float64, inputs []float64, bias float64) float64 {
    out := bias

    for i := range len(weights) {
        out += weights[i] * inputs[i]
    }

    return 1 / (1 + math.Exp(-out))
}

func MakeLeakyReLU(leak float64) ActivationFunc {
    return func(weights []float64, inputs []float64, bias float64) float64 {
        out := bias
    
        for i := range len(weights) {
            out += weights[i] * inputs[i]
        }

        if out > 0 {
            return 1
        }

        return leak * out
    }
}

func ReLU(weights []float64, inputs []float64, bias float64) float64 {
    out := bias
    
    for i := range len(weights) {
        out += weights[i] * inputs[i]
    }

    if out > 0 {
        return 1
    }

    return 0
}

type NeuronFuncs struct {
    Activation ActivationFunc
}

type Neuron struct {
    Bias float64
    Weights []float64
    Funcs NeuronFuncs
}

func (neuron *Neuron) Activate(inputs []float64) float64 {
    return neuron.Funcs.Activation(neuron.Weights, inputs, neuron.Bias)
}

type Network struct {
    Layers [][]Neuron

    outputs []float64
    inputs []float64
}

func randomWeights(size int) []float64 {
    weights := make([]float64, size)
    for i := range size {
        weights[i] = rand.Float64()
    }
    return weights
}

// N input layers that feed to M hidden layers, which feed to K output layers
func MakeNetwork(stateVectorSize int, inputNeurons int, hiddenLayers []int, outputNeurons int) *Network {

    // N hidden layers + 1 input layer + 1 output layer
    layers := make([][]Neuron, len(hiddenLayers) + 2)
    layers[0] = make([]Neuron, inputNeurons)
    for i := range inputNeurons {
        layers[0][i] = Neuron{
            Bias: rand.Float64(),
            Weights: randomWeights(stateVectorSize),
            Funcs: NeuronFuncs{
                Activation: ReLU,
            },
        }
    }

    for layer := range hiddenLayers {
        layerIndex := layer + 1

        layers[layerIndex] = make([]Neuron, hiddenLayers[layer])
        for i := range hiddenLayers[layer] {
            // each neuron in layer X has N weights, where N is the number of neurons in layer X-1
            layers[layerIndex][i] = Neuron{
                Bias: rand.Float64(),
                Weights: randomWeights(len(layers[layerIndex-1])),
                Funcs: NeuronFuncs{
                    Activation: ReLU,
                },
            }
        }
    }

    // final output layer
    layers[len(layers)-1] = make([]Neuron, outputNeurons)
    for i := range outputNeurons {
        layers[len(layers)-1][i] = Neuron{
            Bias: rand.Float64(),
            Weights: randomWeights(len(layers[len(layers)-2])),
            Funcs: NeuronFuncs{
                Activation: Sigmoid,
            },
        }
    }

    network := Network{
        Layers: layers,
    }

    return &network
}

func (network *Network) FeedForward(inputs []float64) []float64 {
    if cap(network.inputs) < len(inputs) {
        network.inputs = make([]float64, len(inputs))
    } else {
        network.inputs = network.inputs[:len(inputs)]
    }

    copy(network.inputs, inputs)

    // FIXME: replace this with gonum to do matrix operations on the layers
    // mat.NewDense(len(network.Layers[layerIndex]), len(network.Layers[layerIndex-1]), nil)
    // m.SetRow(i, weights for layer i)

    for layerIndex := range network.Layers {
        if cap(network.outputs) < len(network.Layers[layerIndex]) {
            network.outputs = make([]float64, len(network.Layers[layerIndex]))
        } else {
            network.outputs = network.outputs[:len(network.Layers[layerIndex])]
        }
        for i := range network.Layers[layerIndex] {
            network.outputs[i] = network.Layers[layerIndex][i].Activate(network.inputs)
        }
        if cap(network.inputs) < len(network.outputs) {
            network.inputs = make([]float64, len(network.outputs))
        }
        copy(network.inputs, network.outputs)
    }

    return network.outputs
}

func (network *Network) RandomizeWeights() {
}

func (network *Network) Rewards(vector []float64) {
    network.BackPropagate(vector)
}

func (network *Network) BackPropagate(vector []float64) {
}
