package ai

/* EnemyAI is the AI for enemy wizards
 */

import (
    "log"
    "slices"
    "math/rand/v2"

    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    "github.com/kazzmir/master-of-magic/game/magic/units"
)

type EnemyAI struct {
}

func MakeEnemyAI() *EnemyAI {
    return &EnemyAI{}
}

// true if the city is producing some other than trade goods or housing
func isMakingSomething(city *citylib.City) bool {
    if !city.ProducingUnit.Equals(units.UnitNone) {
        return true
    }

    switch city.ProducingBuilding {
        case buildinglib.BuildingHousing, buildinglib.BuildingTradeGoods, buildinglib.BuildingNone: return false
        default: return true
    }
}

// stop producing that unit
func (ai *EnemyAI) ProducedUnit(city *citylib.City, player *playerlib.Player) {
    city.ProducingBuilding = buildinglib.BuildingTradeGoods
    city.ProducingUnit = units.UnitNone
}

func (ai *EnemyAI) Update(self *playerlib.Player, enemies []*playerlib.Player, pathfinder playerlib.PathFinder) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    // FIXME: research spells, cast spells

    for _, city := range self.Cities {
        // city can make something
        if !isMakingSomething(city) {
            possibleUnits := city.ComputePossibleUnits()
            possibleUnits = slices.DeleteFunc(possibleUnits, func(unit units.Unit) bool {
                if unit.IsSettlers() {
                    return true
                }
                return false
            })

            possibleBuildings := city.ComputePossibleBuildings()

            possibleBuildings.RemoveMany(buildinglib.BuildingTradeGoods, buildinglib.BuildingHousing)

            type Choice int
            const ChooseUnit Choice = 0
            const ChooseBuilding Choice = 1

            var choices []Choice
            if len(possibleUnits) > 0 {
                stack := self.FindStack(city.X, city.Y, city.Plane)
                if stack == nil || len(stack.Units()) < 9 {
                    choices = append(choices, ChooseUnit)
                }
            }

            if possibleBuildings.Size() > 0 {
                choices = append(choices, ChooseBuilding)
            }

            if len(choices) > 0 {
                switch choices[rand.N(len(choices))] {
                    case ChooseUnit:
                        unit := possibleUnits[rand.N(len(possibleUnits))]
                        decisions = append(decisions, &playerlib.AIProduceDecision{
                            City: city,
                            Building: buildinglib.BuildingNone,
                            Unit: unit,
                        })
                    case ChooseBuilding:
                        // choose some random building
                        values := possibleBuildings.Values()
                        decisions = append(decisions, &playerlib.AIProduceDecision{
                            City: city,
                            Building: values[rand.N(len(values))],
                            Unit: units.UnitNone,
                        })
                }
            }
        }
    }

    for _, stack := range self.Stacks {
        if stack.HasMoves() {
            // FIXME: enter cities, lairs, nodes for combat
            if len(stack.CurrentPath) == 0 {
                if rand.N(4) == 0 {
                    newX, newY := stack.X() + rand.N(5) - 2, stack.Y() + rand.N(5) - 2
                    path := pathfinder.FindPath(stack.X(), stack.Y(), newX, newY, stack, self.GetFog(stack.Plane()))
                    if len(path) != 0 {
                        stack.CurrentPath = path
                    }
                }
            }

            if len(stack.CurrentPath) > 0 {
                nextMove := stack.CurrentPath[0]
                stack.CurrentPath = stack.CurrentPath[1:]
                decisions = append(decisions, &playerlib.AIMoveStackDecision{
                    Stack: stack,
                    Location: nextMove,
                    Invalid: func(){
                        stack.CurrentPath = nil
                    },
                })
            }
        }
    }

    return decisions
}

func (ai *EnemyAI) NewTurn(player *playerlib.Player) {
    // make sure cities have enough farmers
    for _, city := range player.Cities {
        stack := player.FindStack(city.X, city.Y, city.Plane)
        var units []units.StackUnit
        if stack != nil {
            units = stack.Units()
        }
        city.ResetCitizens(units)
    }

    // keep going as long as there is more food available
    moreFood := true
    for moreFood && player.FoodPerTurn() < 0 {
        // try to update farmers in cities

        moreFood = false
        for _, city := range player.Cities {
            if player.FoodPerTurn() >= 0 {
                break
            }

            if city.Workers > 0 {
                moreFood = true
                city.Farmers += 1
                city.Workers -= 1
            }
        }
    }

    for _, city := range player.Cities {
        log.Printf("ai %v city %v farmer=%v worker=%v rebel=%v", player.Wizard.Name, city.Name, city.Farmers, city.Workers, city.Rebels)
    }
}
