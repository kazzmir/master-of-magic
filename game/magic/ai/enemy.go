package ai

/* EnemyAI is the AI for enemy wizards
 */

import (
    _ "log"
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
            possibleBuildings := city.ComputePossibleBuildings()

            possibleBuildings.RemoveMany(buildinglib.BuildingTradeGoods, buildinglib.BuildingHousing)

            // choose some random building
            if possibleBuildings.Size() > 0 {
                values := possibleBuildings.Values()
                decisions = append(decisions, &playerlib.AIProduceDecision{
                    City: city,
                    Building: values[rand.N(len(values))],
                    Unit: units.UnitNone,
                })
            }
        }
    }

    return decisions
}

func (ai *EnemyAI) NewTurn() {
}
