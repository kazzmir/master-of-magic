package ai

import (
    _ "log"
    "math/rand/v2"

    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/lib/functional"
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

// return a random number between low and high inclusive
func randomRange(low int, high int) int {
    if low > high {
        return 0
    }

    if low == high {
        return low
    }

    return rand.N(high-low+1) + low
}

func (raider *RaiderAI) MoveStacks(player *playerlib.Player, enemies []*playerlib.Player, aiServices playerlib.AIServices) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision
    for _, stack := range player.Stacks {
        _, moved := raider.MovedStacks[stack]
        if !moved && !stack.OutOfMoves() {
            // FIXME: if the unit walked by a previously unknown city, they should stop their current path and possibly attack the city
            if stack.CurrentPath != nil {
                decisions = append(decisions, &playerlib.AIMoveStackDecision{
                    Stack: stack,
                    Path: stack.CurrentPath,
                    ConfirmEncounter_: func (encounter *maplib.ExtraEncounter) bool {
                        return false
                    },
                })
                raider.MovedStacks[stack] = true
                continue
            }

            fog := player.GetFog(stack.Plane())

            var currentPath pathfinding.Path

            for _, enemy := range enemies {
                for _, city := range enemy.Cities {
                    if city.Plane != stack.Plane() {
                        continue
                    }

                    // don't know about the city
                    // FIXME: how should this behave for different fog types?
                    if fog[city.X][city.Y] == data.FogTypeUnexplored {
                        continue
                    }

                    path := aiServices.FindPath(stack.X(), stack.Y(), city.X, city.Y, player, stack, fog)
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

            // choose a random unexplored tile on this continent
            if len(currentPath) == 0 {
                continent := aiServices.GetMap(stack.Plane()).GetContinentTiles(stack.X(), stack.Y())
                for _, tileIndex := range rand.Perm(len(continent)) {
                    tile := &continent[tileIndex]
                    if fog[tile.X][tile.Y] == data.FogTypeUnexplored {
                        currentPath = aiServices.FindPath(stack.X(), stack.Y(), tile.X, tile.Y, player, stack, fog)
                        if len(currentPath) > 0 {
                            break
                        }
                    }
                }

                // just move randomly because all tiles have been explored
                whereX := stack.X() + randomRange(-5, 5)
                whereY := stack.Y() + randomRange(-5, 5)
                currentPath = aiServices.FindPath(stack.X(), stack.Y(), whereX, whereY, player, stack, fog)
            }

            if len(currentPath) > 0 {
                decisions = append(decisions, &playerlib.AIMoveStackDecision{
                    Stack: stack,
                    Path: currentPath,
                    // never enter an encounter
                    ConfirmEncounter_: func (encounter *maplib.ExtraEncounter) bool {
                        return false
                    },
                })

            } else {
                stack.ExhaustMoves()
            }

            raider.MovedStacks[stack] = true
        }
    }
    return decisions
}

func (raider *RaiderAI) CreateUnits(player *playerlib.Player, aiServices playerlib.AIServices) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    // don't create too many stacks
    if len(player.Stacks) > 5 {
        return decisions
    }

    getContinent := functional.Memoize3(func(x int, y int, plane data.Plane) []maplib.FullTile {
        return aiServices.GetMap(plane).GetContinentTiles(x, y)
    })

    findEnemyCity := func (city *citylib.City) bool {
        tiles := getContinent(city.X, city.Y, city.Plane)
        for _, tile := range tiles {
            otherCity, otherPlayer := aiServices.FindCity(tile.X, tile.Y, city.Plane)
            if otherCity != nil && otherPlayer != player {
                return true
            }
        }

        return false
    }

    for _, city := range player.Cities {

        if rand.N(40) == 0 {
            if findEnemyCity(city) {
                // create a stack of N units
                for range rand.N(3) + 1 {
                    decisions = append(decisions, &playerlib.AICreateUnitDecision{
                        Unit: units.ChooseRandomUnit(player.Wizard.Race),
                        X: city.X,
                        Y: city.Y,
                        Plane: city.Plane,
                    })
                }
            }
        }
    }

    return decisions
}

func (raider *RaiderAI) Update(player *playerlib.Player, enemies []*playerlib.Player, aiServices playerlib.AIServices, manaPerTurn int) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    decisions = append(decisions, raider.MoveStacks(player, enemies, aiServices)...)
    decisions = append(decisions, raider.CreateUnits(player, aiServices)...)

    return decisions
}

func (raider *RaiderAI) PostUpdate(self *playerlib.Player, enemies []*playerlib.Player) {
}

func (raider *RaiderAI) ProducedUnit(city *citylib.City, player *playerlib.Player) {
}

func (raider *RaiderAI) ConfirmRazeTown(city *citylib.City) bool {
    return true
}
