package deep

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_RestoreFromDump(t *testing.T) {
	n := NewNeural(&Config{
		Inputs:     1,
		Layout:     []int{5, 3, 1},
		Activation: ActivationSigmoid,
		Weight:     NewUniform(0.5, 0, rand.New(rand.NewPCG(0, 0))),
		Bias:       true,
	})

	dump := n.Dump()
	new := FromDump(dump)

	for i, biases := range n.Biases {
		for j, bias := range biases {
			assert.Equal(t, bias.Weight, new.Biases[i][j].Weight)
		}
	}
	assert.Equal(t, n.String(), new.String())
	assert.Equal(t, n.Predict([]float64{0}), new.Predict([]float64{0}))
}

func Test_Marshal(t *testing.T) {
	n := NewNeural(&Config{
		Inputs:     1,
		Layout:     []int{3, 3, 1},
		Activation: ActivationSigmoid,
		Weight:     NewUniform(0.5, 0, rand.New(rand.NewPCG(0, 0))),
		Bias:       true,
	})

	dump, err := n.Marshal()
	assert.Nil(t, err)

	new, err := Unmarshal(dump)
	assert.Nil(t, err)

	for i, biases := range n.Biases {
		for j, bias := range biases {
			assert.Equal(t, bias.Weight, new.Biases[i][j].Weight)
		}
	}
	assert.Equal(t, n.String(), new.String())
	assert.Equal(t, n.Predict([]float64{0}), new.Predict([]float64{0}))
}
