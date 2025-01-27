package ai

/* EnemyAI is the AI for enemy wizards
 */

import (
    _ "log"
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
        case buildinglib.BuildingHousing, buildinglib.BuildingTradeGoods: return false
        default: return true
    }
}

func (ai *EnemyAI) Update(self *playerlib.Player, enemies []*playerlib.Player, pathfinder playerlib.PathFinder) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    for _, city := range self.Cities {
        // city can make something
        if !isMakingSomething(city) {
            possibleUnits := city.ComputePossibleUnits()
            slices.DeleteFunc(possibleUnits, func(unit units.Unit) bool {
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
                choices = append(choices, ChooseUnit)
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

    return decisions
}

func (ai *EnemyAI) NewTurn() {
}
