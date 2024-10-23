package ai

import (
    "image"

    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

type RaiderAI struct {
    MovedStacks map[*playerlib.UnitStack]bool
}

func MakeRaiderAI() *RaiderAI {
    return &RaiderAI{
        MovedStacks: make(map[*playerlib.UnitStack]bool),
    }
}

func (raider *RaiderAI) NewTurn() {
    raider.MovedStacks = make(map[*playerlib.UnitStack]bool)
}

func (raider *RaiderAI) Update(player *playerlib.Player, enemies []*playerlib.Player, pathfinder playerlib.PathFinder) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision
    if len(player.Stacks) > 0 {
        stack := player.Stacks[0]
        _, moved := raider.MovedStacks[stack]
        if !moved {
            enemy1 := enemies[0]
            city1 := enemy1.Cities[0]

            path := pathfinder.FindPath(stack.X(), stack.Y(), city1.X, city1.Y, stack, player.GetFog(data.PlaneArcanus))
            if path != nil {
                decisions = append(decisions, &playerlib.AIMoveStackDecision{
                    Stack: stack,
                    Location: image.Pt(path[0].X, path[0].Y),
                })
            }

            raider.MovedStacks[stack] = true
        }
    }

    return decisions
}
