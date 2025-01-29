package ai

import (
    _ "log"
    "image"
    "math/rand/v2"

    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
)

type RaiderAI struct {
    MovedStacks map[*playerlib.UnitStack]bool
}

func MakeRaiderAI() *RaiderAI {
    return &RaiderAI{
        MovedStacks: make(map[*playerlib.UnitStack]bool),
    }
}

func (raider *RaiderAI) NewTurn(player *playerlib.Player) {
    raider.MovedStacks = make(map[*playerlib.UnitStack]bool)
}

func (raider *RaiderAI) MoveStacks(player *playerlib.Player, enemies []*playerlib.Player, pathfinder playerlib.PathFinder) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision
    for _, stack := range player.Stacks {
        _, moved := raider.MovedStacks[stack]
        if !moved && !stack.OutOfMoves() {
            fog := player.GetFog(stack.Plane())

            var currentPath pathfinding.Path

            for _, enemy := range enemies {
                for _, city := range enemy.Cities {
                    if city.Plane != stack.Plane() {
                        continue
                    }

                    // don't know about the city
                    if !fog[city.X][city.Y] {
                        continue
                    }

                    path := pathfinder.FindPath(stack.X(), stack.Y(), city.X, city.Y, stack, fog)
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
    return decisions
}

func (raider *RaiderAI) CreateUnits(player *playerlib.Player) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    if rand.N(10) == 0 {
        city := player.Cities[rand.N(len(player.Cities))]

        for range rand.N(3) + 1 {
            decisions = append(decisions, &playerlib.AICreateUnitDecision{
                Unit: units.ChooseRandomUnit(player.Wizard.Race),
                X: city.X,
                Y: city.Y,
                Plane: city.Plane,
            })
        }
    }

    return decisions
}

func (raider *RaiderAI) Update(player *playerlib.Player, enemies []*playerlib.Player, pathfinder playerlib.PathFinder) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    decisions = append(decisions, raider.MoveStacks(player, enemies, pathfinder)...)
    decisions = append(decisions, raider.CreateUnits(player)...)

    return decisions
}

func (raider *RaiderAI) PostUpdate(self *playerlib.Player, enemies []*playerlib.Player) {
}

func (raider *RaiderAI) ProducedUnit(city *citylib.City, player *playerlib.Player) {
}
