package main

import (
    "log"

    "github.com/kazzmir/master-of-magic/game/magic/neural"
)

func main() {
    net := neural.MakeNetwork(5, 4, []int{3, 3}, 2)

    outputs := net.FeedForward([]float64{0.3, 0.8, 0.1, 0.2, 0.18})
    log.Printf("Outputs: %v", outputs)
}
