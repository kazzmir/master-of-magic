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

func train(net *neural.Network, input, expected []float64, learningRate float64) {
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

    net.BackPropagate(costs, input, learningRate)
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
            train(net, input2, expected2, 0.01)
            train(net, input1, expected1, 0.01)
        } else {
            train(net, input1, expected1, 0.01)
            train(net, input2, expected2, 0.01)
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

func learn(net *neural.Network, inputs [][]float64, expected [][]float64, debug bool) int {
    previousLoss := 0.0
    minimumLoss := 1e-6
    maxEpochs := 50000

    learningRate := 0.01

    currentLoss := float64(-1)

    for epoch := range maxEpochs {
        for i := range rand.Perm(len(inputs)) {
            train(net, inputs[i], expected[i], learningRate)
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

        if currentLoss < 0 || totalLoss < currentLoss / 5 {
            currentLoss = totalLoss
            learningRate /= 2
        }

        averageLoss := totalLoss / float64(len(inputs))

        if debug {
            log.Printf("Epoch %v, learning rate: %v, total loss: %v average loss: %v loss@0: %v", epoch, learningRate, totalLoss, averageLoss, net.ComputeLoss(inputs[0], expected[0]))
        }

        if averageLoss < minimumLoss {
            return epoch
        }

        if epoch > 0 && math.Abs(totalLoss-previousLoss) < 1e-11 {
            return epoch
        }

        previousLoss = totalLoss
    }

    return maxEpochs
}

func makeRandomFloatArray(size int) []float64 {
    arr := make([]float64, size)
    for i := range arr {
        // arr[i] = rand.Float64() * 2 -1 
        arr[i] = rand.Float64()
    }
    return arr
}

func learnN() {

    /*
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
    */

    samples := 500

    inputs := make([][]float64, samples)
    outputs := make([][]float64, samples)
    
    for i := range inputs {
        inputs[i] = makeRandomFloatArray(5)
        outputs[i] = makeRandomFloatArray(2)
    }

    goodCount := 0
    badCount := 0

    results := make(chan bool, 10)

    trainer := func(runs int) {
        for range runs {
            net := neural.MakeNetwork(len(inputs[0]), 100, []int{40, 20}, len(outputs[0]))
            learn(net, inputs, outputs, false)

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

func learnDebug() {
    samples := 500

    inputs := make([][]float64, samples)
    outputs := make([][]float64, samples)

    for i := range inputs {
        inputs[i] = makeRandomFloatArray(5)
        outputs[i] = makeRandomFloatArray(2)
    }

    net := neural.MakeNetwork(len(inputs[0]), 150, []int{80, 40}, len(outputs[0]))
    epochs := learn(net, inputs, outputs, true)

    log.Printf("Learned in %v epochs", epochs)

    good := true
    for i := range inputs {
        // output := net.FeedForward(inputs[i])
        loss := net.ComputeLoss(inputs[i], outputs[i])
        // log.Printf("Loss for input %v: %v", i, loss)
        // log.Printf("Inputs: %v, expected: %v, got: %v loss: %v", inputs[i], outputs[i], output, loss)
        good = good && loss < 1e-5
    }

    log.Printf("Learned successfully: %v", good)
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)
    // learnN()
    learnDebug()
}
