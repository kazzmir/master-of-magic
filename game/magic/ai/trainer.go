package ai

import (
    // "fmt"
    // "math"
	"github.com/kazzmir/master-of-magic/lib/deep"
	deep_train "github.com/kazzmir/master-of-magic/lib/deep/training"
)

type RewardTrainer struct {
    *internal
    solver deep_train.Solver
}

func NewRewardTrainer(solver deep_train.Solver) *RewardTrainer {
	return &RewardTrainer{
		solver:    solver,
	}
}

type internal struct {
	deltas [][]float64
}

func newTraining(layers []*deep.Layer) *internal {
	deltas := make([][]float64, len(layers))
	for i, l := range layers {
		deltas[i] = make([]float64, len(l.Neurons))
	}
	return &internal{
		deltas: deltas,
	}
}

// Train trains n
func (t *RewardTrainer) Train(net *deep.Neural, losses []float64, index int) {
    if t.internal == nil {
        t.internal = newTraining(net.Layers)
    }

    // losses := make([]float64, len(net.Layers[len(net.Layers)-1].Neurons))
    /*
    for i := range losses {
        losses[i] = -reward * math.Log(net.Layers[len(net.Layers)-1].Neurons[i].Value)
    }
    */
    // losses[action] = -reward * math.Log(net.Layers[len(net.Layers)-1].Neurons[action].Value)
    // losses[action] = loss

    // fmt.Printf("reward: %.2f, losses: %v\n", reward, losses)

    t.calculateDeltas(net, losses)
    t.update(net, index)
}

func (t *RewardTrainer) calculateDeltas(n *deep.Neural, loss []float64) {
	for i := range n.Layers[len(n.Layers)-1].Neurons {
		t.deltas[len(n.Layers)-1][i] = loss[i]
        /*deep.GetLoss(n.Config.Loss).Df(
			neuron.Value,
			ideal[i],
			neuron.DActivate(neuron.Value))
            */
	}

	for i := len(n.Layers) - 2; i >= 0; i-- {
		for j, neuron := range n.Layers[i].Neurons {
			var sum float64
			for k, s := range neuron.Out {
				sum += s.Weight * t.deltas[i+1][k]
			}
			t.deltas[i][j] = neuron.DActivate(neuron.Value) * sum
		}
	}
}

func (t *RewardTrainer) update(n *deep.Neural, it int) {

    // fmt.Printf("deltas: %v\n", t.deltas)

	var idx int
	for i, l := range n.Layers {
		for j := range l.Neurons {
			for k := range l.Neurons[j].In {
				update := t.solver.Update(l.Neurons[j].In[k].Weight,
					t.deltas[i][j]*l.Neurons[j].In[k].In,
					it,
					idx)
                // fmt.Printf("update layer=%d neuron=%d: %f\n", i, j, update)
				l.Neurons[j].In[k].Weight += update
				idx++
			}
		}
	}
}
