//go:build ml
package training

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SplitSize(t *testing.T) {
	e := make(Examples, 10)

	batches := e.SplitSize(2)
	assert.Len(t, batches, 5)
	for _, batch := range batches {
		assert.Equal(t, 2, len(batch))
	}
}

func Test_SplitN(t *testing.T) {
	e := make(Examples, 10)

	partitions := e.SplitN(3)
	assert.Len(t, partitions, 3)
	assert.Len(t, partitions[0], 4)
	assert.Len(t, partitions[1], 3)
	assert.Len(t, partitions[2], 3)
}

func Test_Split(t *testing.T) {
	e := make(Examples, 1000)

	a, b := e.Split(0.5)

	assert.InEpsilon(t, len(a), 500, 0.1)
	assert.InEpsilon(t, len(b), 500, 0.1)
}
