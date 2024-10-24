package ai

import (
    _ "log"
    "image"
    "math/rand/v2"

    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
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

        for _, stack := range player.Stacks {
            _, moved := raider.MovedStacks[stack]
            if !moved && !stack.OutOfMoves() {

                var currentPath pathfinding.Path

                for _, enemy := range enemies {
                    for _, city := range enemy.Cities {
                        path := pathfinder.FindPath(stack.X(), stack.Y(), city.X, city.Y, stack, player.GetFog(data.PlaneArcanus))
                        if path != nil {
                            if currentPath == nil {
                                currentPath = path
                            } else {
                                if len(path) < len(currentPath) {
                                    currentPath = path
                                }
                            }
                        }
                    }
                }

                if len(currentPath) == 0 {
                    // just move randomly
                    whereX := stack.X() + rand.N(5) - 2
                    whereY := stack.Y() + rand.N(5) - 2
                    currentPath = pathfinder.FindPath(stack.X(), stack.Y(), whereX, whereY, stack, player.GetFog(data.PlaneArcanus))
                }

                if len(currentPath) > 0 {
                    decisions = append(decisions, &playerlib.AIMoveStackDecision{
                        Stack: stack,
                        Location: image.Pt(currentPath[0].X, currentPath[0].Y),
                    })

                } else {
                    stack.ExhaustMoves()
                }

                raider.MovedStacks[stack] = true
            }
        }
    }

    return decisions
}
