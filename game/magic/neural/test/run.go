package main

import (
    "log"
    "math"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/game/magic/neural"
)

func MSE(output, expected float64) float64 {
    v := output - expected
    return v * v
}

func train(net *neural.Network, input, expected []float64) {
    outputs := net.FeedForward(input)
    // log.Printf("Outputs: %v", outputs)

    costs := make([]float64, len(outputs))
    // totalError := 0.0
    for i := range outputs {
        // cost := MSE(outputs[i], expected[i])
        // totalError += MSE(outputs[i], expected[i])
        cost := outputs[i] - expected[i]
        costs[i] = cost
    }

    // log.Printf("Costs: %v", costs)
    // log.Printf("Total error: %v", totalError)

    net.BackPropagate(costs, input, 0.1)
}

func learn2() {
    net := neural.MakeNetwork(5, 4, []int{3, 3}, 2)

    input1 := []float64{0.3, 0.8, 0.1, 0.2, 0.18}
    expected1 := []float64{0.9, 0.3}

    input2 := []float64{0.1, 0.3, 0.18, 0.57, 0.93}
    expected2 := []float64{0.14, 0.67}

    previousLoss := 0.0
    minimumLoss := 1e-6
    for epoch := range 10000 {
        if rand.N(2) == 0 {
            train(net, input2, expected2)
            train(net, input1, expected1)
        } else {
            train(net, input1, expected1)
            train(net, input2, expected2)
        }

        log.Printf("Epoch %v", epoch)
        /*
        active1, total1 := net.CountHiddenActive()


        active2, total2 := net.CountHiddenActive()


        log.Printf("Active1: %v, total1: %v", active1, total1)
        log.Printf("Active2: %v, total2: %v", active2, total2)
        */

        outputs1 := net.FeedForward(input1)
        log.Printf("Outputs1: %v", outputs1)
        outputs2 := net.FeedForward(input2)
        log.Printf("Outputs2: %v", outputs2)

        totalLoss := net.ComputeLoss(input1, expected1) + net.ComputeLoss(input2, expected2)
        log.Printf("Total loss: %v", totalLoss)
        if totalLoss < minimumLoss {
            break
        }

        if epoch > 0 && math.Abs(totalLoss-previousLoss) < 1e-11 {
            log.Printf("Loss did not improve, stopping training")
            break
        }

        previousLoss = totalLoss
    }

}

func learn1() {
    net := neural.MakeNetwork(5, 4, []int{3, 3}, 2)

    bestLoss := 1e-10

    // for train := range 40000 {
    train := 0
    for {
        train += 1
        input1 := []float64{0.3, 0.8, 0.1, 0.2, 0.18}
        expected := []float64{0.9, 0.3}

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

        net.BackPropagate(costs, input1, 0.1)
    }
}

func main() {
    learn2()
}
