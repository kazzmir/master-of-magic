package main

import (
    "fmt"
    // "math"
    "math/rand/v2"
    "github.com/kazzmir/master-of-magic/lib/deep"
    "github.com/kazzmir/master-of-magic/game/magic/ai"
	deep_train "github.com/kazzmir/master-of-magic/lib/deep/training"
)

func getOutput(x float64) int {
    return int(x * 10)
}

// all probabilities should sum to 1
func sample(probabilities []float64) int {
    r := rand.Float64()
    cumulative := 0.0
    for i, p := range probabilities {
        cumulative += p
        if r < cumulative {
            return i
        }
    }
    return len(probabilities) - 1 // return last index if not found due to rounding
}

func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}

func main() {
    net := deep.NewNeural(&deep.Config{
        Inputs: 3,
        Layout: []int{10, 8, 3},
        Activation: deep.ActivationSigmoid,
        Mode: deep.ModeMultiClass,
        Weight: deep.NewUniform(0.5, 0, nil),
        Bias: true,
    })

    // red gives negative score, green gives positive and blue gives zero score
    red := 0
    green := 0
    blue := 0

    solver := deep_train.NewAdam(0.002, 0.9, 0.999, 1e-8)
    trainer := ai.NewRewardTrainer(solver)
    solver.Init(net.NumWeights())

    losses := make([]float64, 3)

    baseline := 0.0

    for i := range 10000 {
        fmt.Printf("[%d] red: %d, green: %d, blue: %d\n", i, red, green, blue)

        last_red := red
        last_green := green
        last_blue := blue

        total := red + green + blue
        if total == 0 {
            total = 1
        }

        outputs := net.Predict([]float64{float64(red) / float64(total), float64(green) / float64(total), float64(blue) / float64(total)})

        // fmt.Printf("  red: %.2f, green: %.2f, blue: %.2f\n", outputs[0], outputs[1], outputs[2])

        // probabilities := deep.Softmax(outputs)
        probabilities := outputs

        fmt.Printf("  red: %.4f, green: %.4f, blue: %.4f\n", probabilities[0], probabilities[1], probabilities[2])

        if outputs[1] > 0.99 {
            fmt.Printf("  high green probability, skipping update\n")
            break
        }

        action := sample(probabilities)
        fmt.Printf("  action: %d\n", action)

        switch action {
            case 0:
                red += 1
            case 1:
                green += 1
            case 2:
                blue += 1
        }

        /*
        red += getOutput(outputs[0])
        green += getOutput(outputs[1])
        blue += getOutput(outputs[2])
        */

        /*
        if outputs[0] > 0.5 {
            red += 1
        }

        if outputs[1] > 0.5 {
            green += 1
        }
        
        if outputs[2] > 0.5 {
            blue += 1
        }
        */

        diff_red := red - last_red
        diff_green := green - last_green
        diff_blue := blue - last_blue
        score := -2 * diff_red + diff_green + diff_blue * 0

        // score := -(red * red) + green

        beta := 0.97
        baseline = beta * baseline + (1-beta) * float64(score)
        advantage := float64(score) - baseline

        for z := range probabilities {
            target := 0.0
            if z == action {
                target = 1
            }

            losses[z] = - advantage * (target - probabilities[z])
        }

        /*
        for z := range probabilities {
            if z == action {
                losses[z] = float64(-score) * (1 - probabilities[z])
            } else {
                losses[z] = float64(-score) * (0 - probabilities[z])
            }
        }
        */

        // loss := float64(-score) * math.Log(probabilities[action])

        fmt.Printf("  score: %d, losses: %v\n", score, losses)

        trainer.Train(net, losses, i + 1)
    }

}
