package neural

import (
    "math"
    "math/rand/v2"
    // "log"
)

// create a neural network with a feed forward pass and back propagation
// the network should take in a state vector extracted from the game state
// and output a set of strategy weights that a lower tier AI can use to make decisions
// the reward values are used to train the network using back propagation, where
// the reward values are a vector, some of the values are used for multiple network outputs

type ActivationFunc func(z float64) float64

func Sigmoid(v float64) float64 {
    return 1 / (1 + math.Exp(-v))
}

func SigmoidDerivative(v float64) float64 {
    s := Sigmoid(v)
    return s * (1 - s)
}

func SigmoidDerivativeFromOutput(output float64) float64 {
    return output * (1 - output)
}

func MakeLeakyReLU(leak float64) ActivationFunc {
    return func(z float64) float64 {
        if z > 0 {
            return z
        }

        return leak * z
    }
}

func MakeLeakyReLUDerivative(leak float64) ActivationFunc {
    return func(z float64) float64 {
        if z > 0 {
            return 1
        }

        return leak
    }
}

func ReLU(z float64) float64 {
    return max(0, z)
}

func ReLUDerivative(z float64) float64 {
    if z > 0 {
        return 1
    }
    
    return 0
}

type NeuronFuncs struct {
    Activation ActivationFunc
    Derivative ActivationFunc
}

type Neuron struct {
    Bias float64
    Weights []float64
    Funcs NeuronFuncs

    // last computed value during forward pass
    z float64
    activation float64
}

func (neuron *Neuron) Activate(inputs []float64) float64 {
    z := neuron.Bias
    for i := range len(neuron.Weights) {
        z += neuron.Weights[i] * inputs[i]
    }

    neuron.z = z
    neuron.activation = neuron.Funcs.Activation(neuron.z)

    return neuron.activation
}

type Network struct {
    Layers [][]Neuron

    outputs []float64
    inputs []float64
    costs []float64
    nextLayerCosts []float64
}

func randomWeights(size int) []float64 {
    weights := make([]float64, size)
    for i := range size {
        weights[i] = rand.Float64() * 2 - 1
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
            // Bias: rand.Float64(),
            Bias: rand.NormFloat64() * 0.01,
            Weights: randomWeights(stateVectorSize),
            Funcs: NeuronFuncs{
                Activation: ReLU,
                Derivative: ReLUDerivative,
            },
        }
    }

    for layer := range hiddenLayers {
        layerIndex := layer + 1

        layers[layerIndex] = make([]Neuron, hiddenLayers[layer])
        for i := range hiddenLayers[layer] {
            // each neuron in layer X has N weights, where N is the number of neurons in layer X-1
            layers[layerIndex][i] = Neuron{
                // Bias: rand.Float64() * 2 - 1,
                Bias: rand.NormFloat64() * 0.01,
                Weights: randomWeights(len(layers[layerIndex-1])),
                Funcs: NeuronFuncs{
                    Activation: ReLU,
                    Derivative: ReLUDerivative,
                },
            }
        }
    }

    // final output layer
    layers[len(layers)-1] = make([]Neuron, outputNeurons)
    for i := range outputNeurons {
        layers[len(layers)-1][i] = Neuron{
            Bias: rand.Float64() * 2 - 1,
            Weights: randomWeights(len(layers[len(layers)-2])),
            Funcs: NeuronFuncs{
                Activation: Sigmoid,
                Derivative: SigmoidDerivativeFromOutput,
            },
        }
    }

    network := Network{
        Layers: layers,
    }

    return &network
}

func (network *Network) CountActive() (int, int) {
    active := 0
    total := 0
    for layerIndex := range network.Layers {
        for i := range network.Layers[layerIndex] {
            total += 1
            if network.Layers[layerIndex][i].activation > 0 {
                active += 1
            }
        }
    }

    return active, total
}

func (network *Network) CountHiddenActive() (int, int) {
    active := 0
    total := 0
    for i := 1; i < len(network.Layers) - 1; i++ {
        for j := range network.Layers[i] {
            total += 1
            if network.Layers[i][j].activation > 0 {
                active += 1
            }
        }
    }

    return active, total
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
        network.inputs = network.inputs[:len(network.outputs)]
        copy(network.inputs, network.outputs)
    }

    return network.outputs
}

func (network *Network) RandomizeWeights() {
}

func (network *Network) Rewards(vector []float64) {
    network.BackPropagate(vector, nil, 0.01)
}

func (network *Network) BackPropagate(costs []float64, inputs []float64, learningRate float64) {
    if cap(network.costs) < len(costs) {
        network.costs = make([]float64, len(costs))
    } else {
        network.costs = network.costs[:len(costs)]
    }

    copy(network.costs, costs)

    for layerIndex := len(network.Layers) - 1; layerIndex >= 0; layerIndex-- {
        layer := network.Layers[layerIndex]

        if cap(network.nextLayerCosts) < len(layer[0].Weights) {
            network.nextLayerCosts = make([]float64, len(layer[0].Weights))
        } else {
            network.nextLayerCosts = network.nextLayerCosts[:len(layer[0].Weights)]
        }

        for i := range network.nextLayerCosts {
            network.nextLayerCosts[i] = 0
        }

        // log.Printf("Back propagate layer %v with costs %v", layerIndex, network.costs)

        for i := range layer {
            neuron := &layer[i]
            dy_dz := neuron.Funcs.Derivative(neuron.activation)
            delta_y := network.costs[i] * dy_dz

            for weight := range neuron.Weights {
                var dz_dw0 float64
                if layerIndex > 0 {
                    dz_dw0 = network.Layers[layerIndex - 1][weight].activation
                } else {
                    dz_dw0 = inputs[weight]
                }

                network.nextLayerCosts[weight] += neuron.Weights[weight] * delta_y

                neuron.Weights[weight] -= learningRate * dz_dw0 * delta_y
            }
            neuron.Bias -= learningRate * delta_y
        }

        if cap(network.costs) < len(network.nextLayerCosts) {
            network.costs = make([]float64, len(network.nextLayerCosts))
        } else {
            network.costs = network.costs[:len(network.nextLayerCosts)]
        }

        copy(network.costs, network.nextLayerCosts)
    }

    /*
    for i := range network.Layers[len(network.Layers)-1] {
        neuron := &network.Layers[len(network.Layers)-1][i]
        log.Printf("Output %v: z=%v output=%v", i, neuron.z, neuron.activation)
    }
    */
}

func MSE(output, expected float64) float64 {
    v := output - expected
    return v * v
}

func (network *Network) ComputeLoss(inputs []float64, expected []float64) float64 {
    outputs := network.FeedForward(inputs)
    totalError := 0.0
    for i := range outputs {
        totalError += MSE(outputs[i], expected[i])
    }
    return totalError
}
