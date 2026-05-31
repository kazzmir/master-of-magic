package main

import (
    "log"

    "github.com/kazzmir/master-of-magic/game/magic/neural"
)

func MSE(output, expected float64) float64 {
    v := output - expected
    return v * v
}

func main() {
    net := neural.MakeNetwork(5, 4, []int{3, 3}, 2)

    bestLoss := 1e-10

    // for train := range 40000 {
    train := 0
    for {
        train += 1
        input1 := []float64{0.3, 0.8, 0.1, 0.2, 0.18}
        expected := []float64{0.7, 0.3}

        log.Printf("Run %v", train)

        outputs := net.FeedForward(input1)
        log.Printf("Outputs: %v", outputs)

        // FIXME: for a reward based system, use -reward * log(probability) as loss

        costs := make([]float64, len(outputs))
        totalError := 0.0
        for i := range outputs {
            // cost := MSE(outputs[i], expected[i])
            totalError += MSE(outputs[i], expected[i])
            cost := outputs[i] - expected[i]
            costs[i] = cost
        }

        log.Printf("Costs: %v", costs)
        log.Printf("Total error: %v", totalError)

        if totalError < bestLoss {
            break
        }

        net.BackPropagate(costs, input1, 0.001)
    }
}
