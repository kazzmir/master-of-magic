package main

import (
    "log"
    "math"
    "math/rand/v2"
    "sync"
    // "context"

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

func learn(net *neural.Network, inputs [][]float64, expected [][]float64) int {
    previousLoss := 0.0
    minimumLoss := 1e-6
    maxEpochs := 50000
    for epoch := range maxEpochs {
        for i := range rand.Perm(len(inputs)) {
            train(net, inputs[i], expected[i])
        }

        // log.Printf("Epoch %v", epoch)
        /*
        active1, total1 := net.CountHiddenActive()


        active2, total2 := net.CountHiddenActive()


        log.Printf("Active1: %v, total1: %v", active1, total1)
        log.Printf("Active2: %v, total2: %v", active2, total2)
        */

        totalLoss := 0.0
        for i := range inputs {
            // outputs := net.FeedForward(inputs[i])
            // log.Printf("Outputs for input %v: %v", i, outputs)
            totalLoss += net.ComputeLoss(inputs[i], expected[i])
        }

        if totalLoss < minimumLoss {
            return epoch
        }

        if epoch > 0 && math.Abs(totalLoss-previousLoss) < 1e-11 {
            return epoch
        }

        previousLoss = totalLoss
    }

    return maxEpochs
}

func learnN() {

    inputs := [][]float64{
        {0.3, 0.8, 0.1, 0.2, 0.18},
        {0.1, 0.3, 0.18, 0.57, 0.93},
        {0.8, 0.4, 0.35, 0.76, 0.12},
        {0.42, 0.66, 0.92, 0.28, 0.34},
    }

    outputs := [][]float64{
        {0.9, 0.3},
        {0.14, 0.67},
        {0.39, 0.61},
        {0.08, 0.72},
    }

    goodCount := 0
    badCount := 0

    results := make(chan bool, 10)

    trainer := func(runs int) {
        for range runs {
            net := neural.MakeNetwork(5, 4, []int{30, 20}, 2)
            learn(net, inputs, outputs)

            // log.Printf("Learned in %v epochs", epochs)

            good := true
            for i := range inputs {
                // output := net.FeedForward(inputs[i])
                loss := net.ComputeLoss(inputs[i], outputs[i])
                // log.Printf("Inputs: %v, expected: %v, got: %v loss: %v", inputs[i], outputs[i], output, loss)
                good = good && loss < 1e-5
            }
            if good {
                // log.Printf("Learned successfully")
                results <- true
            } else {
                // log.Printf("Did not learn successfully")
                results <- false
            }
        }

        // log.Printf("Trainer done with %v runs", N)
    }

    /*
    for range 500 {
        net := neural.MakeNetwork(5, 4, []int{30, 20}, 2)
        learn(net, inputs, outputs)

        // log.Printf("Learned in %v epochs", epochs)

        good := true
        for i := range inputs {
            // output := net.FeedForward(inputs[i])
            loss := net.ComputeLoss(inputs[i], outputs[i])
            // log.Printf("Inputs: %v, expected: %v, got: %v loss: %v", inputs[i], outputs[i], output, loss)
            good = good && loss < 1e-5
        }
        if good {
            // log.Printf("Learned successfully")
            goodCount += 1
        } else {
            // log.Printf("Did not learn successfully")
            badCount += 1
        }

        log.Printf("Good count: %v, bad count: %v percent good = %.2f", goodCount, badCount, float64(goodCount)/float64(goodCount+badCount))
    }
    */

    var wg sync.WaitGroup
    for range 5 {
        wg.Go(func(){ trainer(100) })
    }

    // quit, cancel := context.WithCancel(context.Background())

    done := make(chan struct{})

    go func(){
        defer close(done)
        for result := range results {
            if result {
                goodCount += 1
            } else {
                badCount += 1
            }
            log.Printf("Good count: %v, bad count: %v percent good = %.2f", goodCount, badCount, float64(goodCount)/float64(goodCount+badCount))
        }
    }()

    wg.Wait()
    close(results)

    <-done

    log.Printf("Good count: %v, bad count: %v percent good = %.2f", goodCount, badCount, float64(goodCount)/float64(goodCount+badCount))
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
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)
    learnN()
}
