package ai

import (
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

