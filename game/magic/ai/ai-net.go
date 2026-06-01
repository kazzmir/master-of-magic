package ai

import (
    // "math"
    "iter"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
)

// this AI uses two layers:
//  layer 1: a neural network that takes in a state extraction vector of inputs and outputs a vector of strategy probabilities
//  layer 2: an 'operational manager' that acceps the strategy probabilities and selects specific actions to do
// the neural network uses reinforcement learning by using a set of reward signals as the loss/cost function to optimize the network weights
 
type EnemyNetAI struct {
}

var _ playerlib.AIBehavior = (*EnemyNetAI)(nil)

func MakeEnemyNetAI() *EnemyNetAI {
    return &EnemyNetAI{}
}

func countArcanusCities(player *playerlib.Player) int {
    count := 0
    for _, city := range player.Cities {
        if city.Plane == data.PlaneArcanus {
            count += 1
        }
    }

    return count
}

func countMyrrorCities(player *playerlib.Player) int {
    count := 0
    for _, city := range player.Cities {
        if city.Plane == data.PlaneMyrror {
            count += 1
        }
    }

    return count
}

func countUnits(units iter.Seq[units.StackUnit], plane data.Plane) int {
    count := 0
    for unit := range units {
        if unit.GetPlane() == plane {
            count += 1
        }
    }

    return count
}

func countArcanusUnits(units iter.Seq[units.StackUnit]) int {
    return countUnits(units, data.PlaneArcanus)
}

func countMyrrorUnits(units iter.Seq[units.StackUnit]) int {
    return countUnits(units, data.PlaneMyrror)
}

func featureExtraction(player *playerlib.Player, services playerlib.AIServices) []float64 {
    var features []float64

    // each feature should be a normalized value in the range [0, 1] representing some aspect of the game state relevant to decision making

    features = append(features, min(1, float64(services.GetTurnNumber()) / 1000))
    features = append(features, min(1, float64(player.Gold) / 20000))
    features = append(features, min(1, float64(player.Mana) / 20000))
    features = append(features, min(1, float64(countArcanusCities(player)) / 30))
    features = append(features, min(1, float64(countMyrrorCities(player)) / 30))
    features = append(features, min(1, float64(len(player.AliveHeroes())) / 6))
    features = append(features, min(1, float64(countArcanusUnits(player.Units()) / 100)))
    features = append(features, min(1, float64(countMyrrorUnits(player.Units()) / 100)))

    if player.ResearchingSpell.Valid() {
        features = append(features, min(1, float64(player.ResearchProgress) / float64(player.ResearchingSpell.ResearchCost)))
    } else {
        features = append(features, 0)
    }
    if player.CastingSpell.Valid() {
        spellCost := player.ComputeEffectiveSpellCost(player.CastingSpell, true)
        features = append(features, min(1, float64(player.CastingSpellProgress) / float64(spellCost)))
    } else {
        features = append(features, 0)
    }

    return features
}

func (ai *EnemyNetAI) Update(player *playerlib.Player, services playerlib.AIServices) []playerlib.AIDecision {
    return nil
}

func (ai *EnemyNetAI) PostUpdate(player *playerlib.Player, services playerlib.AIServices) {
}

func (ai *EnemyNetAI) NewTurn(player *playerlib.Player) {
}

func (ai *EnemyNetAI) ProducedUnit(city *citylib.City, player *playerlib.Player) {
}

func (ai *EnemyNetAI) ConfirmRazeTown(city *citylib.City) bool {
    return false
}

func (ai *EnemyNetAI) HandleMerchantItem(player *playerlib.Player, artifact *artifact.Artifact, cost int) bool {
    return false
}

func (ai *EnemyNetAI) HandleHireHero(player *playerlib.Player, hero *herolib.Hero, x int, b bool, point data.PlanePoint) {
}

func (ai *EnemyNetAI) HandleHireMercenaries(player *playerlib.Player, guys []*units.OverworldUnit, cost int) {
}

func (ai *EnemyNetAI) InvalidMove(stack *playerlib.UnitStack) {
}

func (ai *EnemyNetAI) MovedStack(stack *playerlib.UnitStack, path pathfinding.Path) pathfinding.Path {
    return nil
}

func (ai *EnemyNetAI) ConfirmEncounter(stack *playerlib.UnitStack, encounter *maplib.ExtraEncounter) bool {
    return false
}

