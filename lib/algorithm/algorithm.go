package algorithm

import (
    "math/rand/v2"
)

func ChooseRandomElement[T any](values []T) T {
    index := rand.IntN(len(values))
    return values[index]
}

func ChoseRandomWeightedElement[T any](values []T, weights []int) T {
    totalWeight := 0
    for _, weight := range weights {
        totalWeight += weight
    }

    totalWeight = rand.IntN(totalWeight)

    for index, value := range values {
        weight := weights[index]
        if totalWeight < weight {
            return value
        }
        totalWeight -= weight
    }

    return values[0]
}
