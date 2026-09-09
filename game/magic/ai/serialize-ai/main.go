package main

import (
    "log"
    "math/rand/v2"
    "os"

    ailib "github.com/kazzmir/master-of-magic/game/magic/ai"
)

func makeStrategies() []ailib.Probability {
    choices := []ailib.Strategy{
        ailib.StrategyAttackEnemies,
        ailib.StrategyBuildArmy,
        ailib.StrategyAcquireMagicNode,
        ailib.StrategyBuildCities,
        ailib.StrategyDefendCities,
        ailib.StrategyIncreasePopulation,
        ailib.StrategyIncreasePower,
    }

    var out []ailib.Probability

    for i, choice := range rand.Perm(len(choices)) {
        if i >= 2 {
            return out
        }

        out = append(out, ailib.Probability{
            Value: rand.Float64(),
            Index: choice,
        })
    }

    return out
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ai := ailib.MakeEnemyNetAI()

    N := 800

    log.Printf("Create %v steps for training", N)

    for i := range N {
        strategies := makeStrategies()
        ai.Steps = append(ai.Steps, ailib.Step{
            Turn: uint64(i + 1),
            Strategies: strategies,
        })
    }

    log.Printf("Apply training")
    ai.ApplyTraining()

    outputPath := "ai.json"
    output, err := os.Create(outputPath)
    if err != nil {
        log.Fatalf("Failed to open file: %v", err)
    }
    err = ai.SaveNeuralNet(output)
    if err != nil {
        log.Fatalf("Failed to save AI: %v", err)
    }

    log.Printf("Saved AI to %v", outputPath)
}
