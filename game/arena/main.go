package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kazzmir/master-of-magic/game/arena/engine"
)

func main() {
	scenario := flag.Int("test", 0, "start a prepared test army (1-5, larger numbers have more units, gold, and magic)")
	flag.Parse()
	if *scenario < 0 || *scenario > engine.TestScenarioCount {
		fmt.Fprintf(os.Stderr, "-test must be between 0 and %d (0 starts a normal game)\n", engine.TestScenarioCount)
		os.Exit(2)
	}
	engine.RunWithTestScenario(*scenario)
}
