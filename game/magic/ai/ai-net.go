package ai

import (
    // "math"
    "iter"
    "cmp"
    "log"
    "slices"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/lib/deep"
    deep_train "github.com/kazzmir/master-of-magic/lib/deep/training"
)

// this AI uses two layers:
//  layer 1: a neural network that takes in a state extraction vector of inputs and outputs a vector of strategy probabilities
//  layer 2: an 'operational manager' that acceps the strategy probabilities and selects specific actions to do
// the neural network uses reinforcement learning by using a set of reward signals as the loss/cost function to optimize the network weights

// values that are updated since the last turn, used for reward calculation
type PlayerStats struct {
    // banished when the city containing the wizards tower is defeated
    EnemiesBanished int
    // defeated when all cities owned by the wizard are defeated
    EnemiesDefeated int
    UnitsLost int
    // normal units
    UnitsCreated int
    // fantastic units via spells
    UnitsSummoned int
    CitiesRazed int
    CitiesCaptured int
    CitiesLost int
    MagicNodesGained int
    MagicNodesLost int
    GoldDelta int
    ManaDelta int
    TerritoryExplored int
    SpellsLearned int
    HeroesGained int
    HeroesLost int
    ArmyStrengthDelta int
    RoadsBuilt int
    EnemiesDiscovered int
    // value of 0 to 1
    SpellOfMasteryProgress float64
}

type StrategyProbabilities struct {
    AttackEnemies float64
    AcquireMagicNode float64
    BuildCities float64
    DefendCities float64
    IncreasePopulation float64
    IncreaseMagic float64
}

func (probabilities StrategyProbabilities) Count() int {
    return 6
}

type Step struct {
    Turn uint64
    Strategies []Probability
    Reward float64
    Return float64
}
 
type EnemyNetAI struct {
    Stats PlayerStats
    NeuralNet *deep.Neural

    Steps []Step
}

var _ playerlib.AIBehavior = (*EnemyNetAI)(nil)

func MakeEnemyNetAI() *EnemyNetAI {
    net := deep.NewNeural(&deep.Config{
        Inputs: len(makeFeatureVector(nil, nil)),
        Layout: []int{64, StrategyProbabilities{}.Count()},
        // final output layer is sigmoid, which is essentially a probability between 0 and 1 for each strategy,
        // and we can select strategies based on those probabilities
        Activation: deep.ActivationSigmoid,
        Mode: deep.ModeMultiLabel,
        Weight: deep.NewUniform(0.5, 0, nil),
        Bias: true,
    })

    return &EnemyNetAI{
        NeuralNet: net,
    }
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

func countSettlers(units iter.Seq[units.StackUnit]) int {
    count := 0
    for unit := range units {
        if unit.HasAbility(data.AbilityCreateOutpost) {
            count += 1
        }
    }
    return count
}

func totalPopulation(player *playerlib.Player) int {
    population := 0
    for _, city := range player.Cities {
        population += city.Population
    }

    return population
}

func countExploredTiles(player *playerlib.Player, services playerlib.AIServices, plane data.Plane) int {
    map_ := services.GetMap(plane)
    count := 0
    for x := range map_.Width() {
        for y := range map_.Height() {
            if player.IsExplored(x, y, plane) {
                count += 1
            }
        }
    }
    return count
}

func countVisibleEnemyUnits(player *playerlib.Player, services playerlib.AIServices, plane data.Plane) int {

    map_ := services.GetMap(plane)
    count := 0
    for x := range map_.Width() {
        for y := range map_.Height() {
            if player.IsVisible(x, y, plane) {
                stack, owner := services.FindStack(x, y, plane)
                if stack != nil && owner != nil && owner != player {
                    count += stack.Size()
                }
            }
        }
    }

    return count
}

func countMagicNodes(player *playerlib.Player, services playerlib.AIServices, plane data.Plane) int {
    map_ := services.GetMap(plane)
    return len(map_.GetMeldedNodes(player))
}

type FeatureFunction func() float64

func makeFeatureVector(player *playerlib.Player, services playerlib.AIServices) []FeatureFunction {
    return []FeatureFunction{
        func () float64 {
            return float64(services.GetTurnNumber()) / 2000
        },
        func () float64 {
            return float64(player.Gold) / 50000
        },
        func () float64 {
            return float64(player.Mana) / 20000
        },
        func () float64 {
            return float64(countArcanusCities(player)) / 30
        },
        func () float64 {
            return float64(countMyrrorCities(player)) / 30
        },
        func () float64 {
            return float64(len(player.AliveHeroes())) / 6
        },
        func () float64 {
            return float64(countSettlers(player.Units()) / 100)
        },
        func () float64 {
            return float64(countArcanusUnits(player.Units()) / 100)
        },
        func () float64 {
            return float64(countMyrrorUnits(player.Units()) / 100)
        },
        func () float64 {
            return float64(totalPopulation(player)) / 1000
        },
        func () float64 {
            return float64(player.TotalUnitUpkeepGold()) / 1000
        },
        func () float64 {
            return float64(player.TotalUnitUpkeepMana()) / 1000
        },
        func () float64 {
            return float64(player.TotalUnitUpkeepFood()) / 1000
        },
        func () float64 {
            return float64(countExploredTiles(player, services, data.PlaneArcanus)) / 5000
        },
        func () float64 {
            return float64(countExploredTiles(player, services, data.PlaneMyrror)) / 5000
        },
        func () float64 {
            return float64(countVisibleEnemyUnits(player, services, data.PlaneArcanus)) / 100
        },
        func () float64 {
            return float64(countVisibleEnemyUnits(player, services, data.PlaneMyrror)) / 100
        },
        func () float64 {
            return float64(countMagicNodes(player, services, data.PlaneArcanus)) / 20
        },
        func () float64 {
            return float64(countMagicNodes(player, services, data.PlaneMyrror)) / 20
        },
        func() float64 {
            if player.ResearchingSpell.Valid() {
                return float64(player.ResearchProgress) / float64(player.ResearchingSpell.ResearchCost)
            } else {
                return 0
            }
        },
        func() float64 {
            if player.CastingSpell.Valid() {
                spellCost := player.ComputeEffectiveSpellCost(player.CastingSpell, true)
                return float64(player.CastingSpellProgress) / float64(spellCost)
            } else {
                return 0
            }
        },
        func () float64 {
            return float64(len(player.KnownSpells.GetSpellsByRarity(spellbook.SpellRarityCommon).Spells)) / 100.0
        },
        func () float64 {
            return float64(len(player.KnownSpells.GetSpellsByRarity(spellbook.SpellRarityUncommon).Spells)) / 100.0
        },
        func () float64 {
            return float64(len(player.KnownSpells.GetSpellsByRarity(spellbook.SpellRarityRare).Spells)) / 100.0
        },
        func () float64 {
            return float64(len(player.KnownSpells.GetSpellsByRarity(spellbook.SpellRarityVeryRare).Spells)) / 100.0
        },
        func() float64 {
            if player.ResearchCandidateSpells.Contains(spellbook.SpellOfMastery) {
                return 1
            }
            return 0
        },
        func() float64 {
            if player.CastingSpell.Valid() && player.CastingSpell.Name == spellbook.SpellOfMastery.Name {
                return 1
            }
            return 0
        },
        func() float64 {
            knownPlayers := player.GetKnownPlayers()
            count := 0
            for _, known := range knownPlayers {
                if !known.Defeated && !known.Banished {
                    count += 1
                }
            }

            // max 3 opponents
            return float64(count) / 3.0
        },
        func() float64 {
            if player.Banished {
                return 1
            }
            return 0
        },
        func () float64 {
            return float64(player.Fame) / 1000
        },
        func () float64 {
            return player.TaxRate.ToFloat() / 5
        },
        func () float64 {
            return player.PowerDistribution.Mana
        },
        func () float64 {
            return player.PowerDistribution.Skill
        },
        func () float64 {
            return player.PowerDistribution.Research
        },
        func () float64 {
            return float64(player.CastingSkillPower) / 100000
        },
    }
}

func featureExtraction(player *playerlib.Player, services playerlib.AIServices) []float64 {
    // each feature should be a normalized value in the range [0, 1] representing some aspect of the game state relevant to decision making
    functions := makeFeatureVector(player, services)
    features := make([]float64, len(functions))

    for i, f := range functions {
        features[i] = min(1, max(-1, f()))
    }

    return features
}

type Probability struct {
    Value float64
    Index int
}

func pickTopN(probabilities []float64, n int) []Probability {
    all := make([]Probability, 0, len(probabilities))

    for i, p := range probabilities {
        all = append(all, Probability{
            Value: p,
            Index: i,
        })
    }

    all = slices.SortedFunc(slices.Values(all), func(a, b Probability) int {
        return cmp.Compare(a.Value, b.Value)
    })

    if len(all) > n {
        return all[len(all)-n:]
    }

    return all
}

func (ai *EnemyNetAI) Update(player *playerlib.Player, services playerlib.AIServices) []playerlib.AIDecision {
    // create feature vector from game state, feed it to neural network, get strategy probabilities
    // use strategy probabilities to select specific actions to take this turn
    // get reward signals, feed back into neural network for training

    features := featureExtraction(player, services)
    strategies := ai.NeuralNet.Predict(features)

    top2 := pickTopN(strategies, 2)

    ai.Steps = append(ai.Steps, Step{
        Turn: services.GetTurnNumber(),
        Strategies: top2,
    })

    return nil
}

func (ai *EnemyNetAI) ApplyTraining() {
    gamma := 0.991

    g := 0.0

    // go backwards in time to assign reward values to each step
    // this allows big positive rewards in the future to propagate backward to earlier steps that led to that reward
    for i := len(ai.Steps) - 1; i >= 0; i-- {
        g = ai.Steps[i].Reward + gamma * g
        ai.Steps[i].Return = g
    }

    baseline := 0.0
    beta := 0.97

    losses := make([]float64, StrategyProbabilities{}.Count())

    solver := deep_train.NewAdam(0.002, 0.9, 0.999, 1e-8)
    trainer := NewRewardTrainer(solver)
    solver.Init(ai.NeuralNet.NumWeights())

    for _, step := range ai.Steps {
        baseline = beta * baseline + (1-beta) * float64(step.Return)
        advantage := float64(step.Return) - baseline

        // compute losses for each strategy, which either encourages or discourages the strategy based on whether the advantage is positive or negative,
        // and whether the strategy was selected or not
        for i := range losses {
            base := 0
            for _, strategy := range step.Strategies {
                if strategy.Index == i {
                    base = 1
                }
            }

            losses[i] = -advantage * (float64(base) - step.Strategies[i].Value)
        }

        // back propagate the losses to update the neural network weights, using the turn number as the index for the training step
        trainer.Train(ai.NeuralNet, losses, int(step.Turn))
    }
}

func (ai *EnemyNetAI) PostUpdate(player *playerlib.Player, services playerlib.AIServices) {
    // compute rewards
    // rewards are any event that can be quantified, like how many enemy units were killed, how much gold was gained, how many cities were captured, etc

    var reward float64 = 0

    reward += float64(ai.Stats.EnemiesBanished) * 1000

    ai.Steps[len(ai.Steps)-1].Reward = reward

    /*
    // defeated when all cities owned by the wizard are defeated
    EnemiesDefeated int
    UnitsLost int
    // normal units
    UnitsCreated int
    // fantastic units via spells
    UnitsSummoned int
    CitiesRazed int
    CitiesCaptured int
    CitiesLost int
    MagicNodesGained int
    MagicNodesLost int
    GoldDelta int
    ManaDelta int
    TerritoryExplored int
    SpellsLearned int
    HeroesGained int
    HeroesLost int
    ArmyStrengthDelta int
    RoadsBuilt int
    EnemiesDiscovered int
    // value of 0 to 1
    SpellOfMasteryProgress float64
    */

    log.Printf("Reward: %f", reward)
}

func (ai *EnemyNetAI) NewTurn(player *playerlib.Player) {
    // reset stats
    ai.Stats = PlayerStats{}
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

