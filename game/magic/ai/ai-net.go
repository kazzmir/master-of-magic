package ai

import (
    // "math"
    "iter"
    "cmp"
    "log"
    "image"
    "slices"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/lib/functional"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/lib/algorithm"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
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
    // 1 if the wizard was banished or defeated this turn, 0 otherwise
    WasBanished int
    WasDefeated int
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

type Strategy int
const (
    StrategyAttackEnemies Strategy = iota
    StrategyBuildArmy
    StrategyAcquireMagicNode
    StrategyBuildCities
    StrategyDefendCities
    StrategyIncreasePopulation
    StrategyIncreasePower

    // FIXME: add Explore

    StrategyCount
)

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

    Attacking map[*playerlib.UnitStack]bool

    // current state of the player, used for reward calculation
    currentGold int
    currentMana int
    banished bool
    defeated bool
}

var _ playerlib.AIBehavior = (*EnemyNetAI)(nil)

func MakeEnemyNetAI() *EnemyNetAI {
    net := deep.NewNeural(&deep.Config{
        Inputs: len(makeFeatureVector(nil, nil)),
        Layout: []int{64, int(StrategyCount)},
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

    return ai.OperationalManager(player, services, top2)
}

// search for enemies to attack and move towards
// strength is how likely the AI is to attack even if the attacking stack is weaker than the target,
// with higher values being more likely to attack with weaker stacks
func (ai *EnemyNetAI) DoAttackEnemies(self *playerlib.Player, aiServices playerlib.AIServices, strength float64) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    // find possible enemy targets
    var possibleTarget []*playerlib.UnitStack
    var possibleCities []*citylib.City
    for _, enemyPlayer := range aiServices.GetEnemies(self) {
        // FIXME: if there is a diplomatic treaty with the enemy then do not attack them

        for _, enemyStack := range enemyPlayer.Stacks {
            if self.IsVisible(enemyStack.X(), enemyStack.Y(), enemyStack.Plane()) {
                possibleTarget = append(possibleTarget, enemyStack)
            }
        }

        for _, enemyCity := range enemyPlayer.Cities {
            // in theory we can see cities that on tiles that we have explored in the past
            if self.IsVisible(enemyCity.X, enemyCity.Y, enemyCity.Plane) {
                possibleCities = append(possibleCities, enemyCity)
            }
        }
    }

    if len(possibleTarget) > 0 || len(possibleCities) > 0 {
        for _, stack := range self.Stacks {
            stackPower := stackAttackPower(stack)

            if stackPower > 0 && stack.HasMoves() {

                var shortestPath pathfinding.Path

                for _, city := range possibleCities {
                    if city.Plane == stack.Plane() {
                        // just assume the city has some power in it
                        targetPower := 15
                        // FIXME: if the stack is 1 or 2 tiles away then we can get the exact power of the garrison in the city
                        if stackPower > targetPower - rand.N(int(strength * 20)) {
                            pathToCity, ok := aiServices.FindPath(stack.X(), stack.Y(), city.X, city.Y, self, stack, self.GetFog(stack.Plane()))
                            if ok {
                                if len(shortestPath) == 0 || len(pathToCity) < len(shortestPath) {
                                    shortestPath = pathToCity
                                }
                            }
                        }
                    }
                }

                if len(shortestPath) == 0 {
                    for _, target := range possibleTarget {
                        if target.Plane() == stack.Plane() {

                            targetPower := stackAttackPower(target)

                            if stackPower > targetPower - rand.N(int(strength * 20)) {

                                pathToEnemy, ok := aiServices.FindPath(stack.X(), stack.Y(), target.X(), target.Y(), self, stack, self.GetFog(stack.Plane()))
                                if ok {
                                    if len(shortestPath) == 0 || len(pathToEnemy) < len(shortestPath) {
                                        shortestPath = pathToEnemy
                                    }
                                }
                            }
                        }
                    }
                }

                if len(shortestPath) > 0 {
                    // log.Printf("AI %v moving stack at %v,%v to attack enemy via %v", self.Wizard.Name, stack.X(), stack.Y(), shortestPath)
                    decisions = append(decisions, &playerlib.AIMoveStackDecision{
                        Stack: stack,
                        Path: shortestPath,
                    })

                    ai.Attacking[stack] = true
                }
            }
        }
    }

    return decisions
}

func isBuildingCombatUnit(city *citylib.City) bool {
    if !city.ProducingUnit.Equals(units.UnitNone) {
        return rawUnitAttackPower(city.ProducingUnit) > 0
    }

    return false
}

// return a value between 0 and 1 representing how much of the current thing being produced is finished
func buildPercent(city *citylib.City) float64 {
    if !city.ProducingUnit.Equals(units.UnitNone) {
        cost := city.UnitProductionCost(&city.ProducingUnit)
        if cost <= 0 {
            return 0
        }
        return float64(city.Production) / float64(cost)
    }

    if city.ProducingBuilding != buildinglib.BuildingNone {
        cost := city.BuildingInfo.ProductionCost(city.ProducingBuilding)
        if cost <= 0 {
            return 0
        }

        return float64(city.Production) / float64(cost)
    }

    return 0
}

func (ai *EnemyNetAI) DoBuildArmy(self *playerlib.Player, aiServices playerlib.AIServices, strength float64) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    computeTransportUnits := functional.Memoize(func(plane data.Plane) int {
        return self.TransportUnits(plane)
    })

    for _, city := range self.Cities {
        // possibly switch away from whatever is currently being built to build combat units
        if !isMakingSomething(city) || (!isBuildingCombatUnit(city) && buildPercent(city) < strength) {

            useMap := aiServices.GetMap(city.Plane)
            buildNavy := useMap.OnShore(city.X, city.Y) && computeTransportUnits(city.Plane) < 2

            cityDecision := func() (playerlib.AIDecision, bool) {
                if buildNavy && chance(50) {
                    // build buildings towards a ship building if necessary
                    // otherwise if this city can build a ship then do so
                    possibleUnits := city.ComputePossibleUnits()
                    possibleUnits = slices.DeleteFunc(possibleUnits, func(unit units.Unit) bool {
                        return !unit.HasAbility(data.AbilityTransport)
                    })

                    // FIXME: sort the transport units by their strength. prefer warship over trieme

                    if len(possibleUnits) > 0 {
                        // log.Printf("AI %v building navy unit in city %v", self.Wizard.Name, city.Name)
                        return &playerlib.AIProduceDecision{
                            City: city,
                            Building: buildinglib.BuildingNone,
                            Unit: possibleUnits[rand.N(len(possibleUnits))],
                        }, true
                    } else {
                        transportBuildings := set.NewSet(
                            buildinglib.BuildingShipYard,
                            buildinglib.BuildingShipwrightsGuild,
                            buildinglib.BuildingMaritimeGuild,
                        )

                        // shipyard for draconian builds an airship, which does not have transport
                        if city.Race == data.RaceDraconian {
                            transportBuildings.Remove(buildinglib.BuildingShipYard)
                        }

                        // get full set of dependencies
                        for _, building := range transportBuildings.Values() {
                            dependencies := city.BuildingInfo.Dependencies(building)
                            transportBuildings.InsertMany(dependencies...)
                        }

                        // try to choose one of the transport buildings or one of its dependencies to build
                        possibleBuildings := city.ComputePossibleBuildings(true)
                        for _, building := range possibleBuildings.Values() {
                            if transportBuildings.Contains(building) {
                                // log.Printf("AI %v building transport building %v in city %v", self.Wizard.Name, building, city.Name)
                                return &playerlib.AIProduceDecision{
                                    City: city,
                                    Building: building,
                                    Unit: units.UnitNone,
                                }, true
                            }
                        }

                    }
                }

                if self.FoodPerTurn() > 0 && self.GoldPerTurn() > 0 && self.Gold > 50 && chance(30) {
                    possibleUnits := city.ComputePossibleUnits()

                    possibleUnits = slices.DeleteFunc(possibleUnits, func(unit units.Unit) bool {
                        if unit.IsSettlers() {
                            return true
                        }
                        return false
                    })

                    // bias towards stronger units
                    attacks := make([]int, 0, len(possibleUnits))
                    for _, unit := range possibleUnits {
                        attacks = append(attacks, rawUnitAttackPower(unit))
                    }

                    if len(possibleUnits) > 0 {
                        return &playerlib.AIProduceDecision{
                            City: city,
                            Building: buildinglib.BuildingNone,
                            Unit: algorithm.ChoseRandomWeightedElement(possibleUnits, attacks),
                        }, true
                    }
                }

                return nil, false
            }

            decision, ok := cityDecision()
            if ok {
                decisions = append(decisions, decision)
            }
        }
    }

    return decisions
}

func (ai *EnemyNetAI) DoBuildCities(self *playerlib.Player, aiServices playerlib.AIServices, strength float64) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    // to achieve this goal the AI should move settlers towards settlable locations
    // if a settler is at a settlable location, build an outpost
    // if there are no settlers and there is enough food, produce more settlers

    for _, stack := range self.Stacks {
        if stack.HasMoves() && stack.ActiveUnitsHasAbility(data.AbilityCreateOutpost) {
            // found a stack with a settler that is not currently moving

            findNewPath := len(stack.CurrentPath) == 0

            // if headed to some location, check that we can still settle there
            if len(stack.CurrentPath) > 0 {
                lastPoint := stack.CurrentPath[len(stack.CurrentPath) - 1]
                if !aiServices.IsSettlableLocation(lastPoint.X, lastPoint.Y, stack.Plane()) {
                    findNewPath = true
                }
            }

            if findNewPath {
                // search through all explored locations on the current continent for settlable locations
                fog := self.GetFog(stack.Plane())
                locations := aiServices.FindSettlableLocations(stack.X(), stack.Y(), stack.Plane(), fog)
                citiesOnContinent := aiServices.FindCitiesOnContinent(stack.X(), stack.Y(), stack.Plane(), self)

                // determine if there are any cities on this continent adjacent to a shore,
                // which means we can build a navy on this continent
                hasShoreCity := false
                useMap := aiServices.GetMap(stack.Plane())
                for _, city := range citiesOnContinent {
                    if useMap.OnShore(city.X, city.Y) {
                        hasShoreCity = true
                        break
                    }
                }

                type PathResult struct {
                    Path pathfinding.Path
                    Ok bool
                }

                pathTo := functional.Memoize(func(location image.Point) PathResult {
                    path, ok := aiServices.FindPath(stack.X(), stack.Y(), location.X, location.Y, self, stack, fog)
                    return PathResult{Path: path, Ok: ok}
                })

                maximumPopulation := functional.Memoize(func(location image.Point) int {
                    return aiServices.ComputeMaximumPopulation(location.X, location.Y, stack.Plane())
                })

                // filter out all locations we cannot reach
                locations = slices.DeleteFunc(locations, func(location image.Point) bool {
                    return pathTo(location).Ok == false
                })

                score := func(location image.Point) int {
                    total := maximumPopulation(location)

                    // prioritize shore locations if we don't have a city on this continent adjacent to a shore
                    if !hasShoreCity && useMap.OnShore(location.X, location.Y) {
                        total += 10
                    }

                    return total
                }

                slices.SortFunc(locations, func(a, b image.Point) int {
                    return cmp.Compare(score(b), score(a))
                })

                // log.Printf("AI possible settlable locations: %v", locations)

                if len(locations) > 0 {
                    // search through locations and either return a decision to build an output
                    // because the stack is already at that location, or return a decision to move to that location
                    getDecision := func() (playerlib.AIDecision, bool) {
                        // if standing on a settlable location, then return immediately
                        for _, location := range locations {
                            path := pathTo(location)

                            if len(path.Path) == 0 {
                                if aiServices.IsSettlableLocation(stack.X(), stack.Y(), stack.Plane()) {
                                    return &playerlib.AIBuildOutpostDecision{
                                        Stack: stack,
                                    }, true
                                }
                            }
                        }

                        // otherwise move towards the best settlable location
                        for _, location := range locations {
                            path := pathTo(location)
                            // log.Printf("AI moving settler stack at %v,%v to settlable location %v,%v via %v", stack.X(), stack.Y(), location.X, location.Y, path)
                            return &playerlib.AIMoveStackDecision{
                                Stack: stack,
                                Path: path.Path,
                            }, true
                        }

                        return nil, false
                    }

                    decision, ok := getDecision()
                    if ok {
                        decisions = append(decisions, decision)
                    }
                }
            }
        }
    }

    for _, city := range self.Cities {
        if !isMakingSomething(city) && chance(int(strength * 100)) {
            locations := aiServices.FindSettlableLocations(city.X, city.Y, city.Plane, self.GetFog(city.Plane))
            if len(locations) > 0 && self.FoodPerTurn() > 0 && chance(len(locations) * 5) {
                decisions = append(decisions, &playerlib.AIProduceDecision{
                    City: city,
                    Building: buildinglib.BuildingNone,
                    Unit: units.GetSettlerUnit(city.Race),
                })

            }
        }
    }

    return decisions
}

func (ai *EnemyNetAI) DoAcquireMagicNodes(self *playerlib.Player, aiServices playerlib.AIServices, strength float64) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    // send units towards magic nodes
    // if a magic node is conquered/empty then summon a magic spirit or guardian spirit
    // and meld with the node

    arcanusNodes := functional.Memoize0(aiServices.GetMap(data.PlaneArcanus).GetAllMagicNodeLocations)
    myrrorNodes := functional.Memoize0(aiServices.GetMap(data.PlaneMyrror).GetAllMagicNodeLocations)

    for _, stack := range self.Stacks {
        stackPower := stackAttackPower(stack)
        stackMap := aiServices.GetMap(stack.Plane())

        if stackMap.GetMagicNode(stack.X(), stack.Y()) != nil && stack.ActiveUnitsHasAbility(data.AbilityMeld) {
            decisions = append(decisions, &playerlib.AIMeldDecision{
                Stack: stack,
            })
            continue
        }

        // ignore stacks that have no power
        if stackPower == 0 {
            continue
        }

        if len(stack.CurrentPath) > 0 {
            continue
        }

        var points []image.Point
        switch stack.Plane() {
            case data.PlaneArcanus:
                points = arcanusNodes()
            case data.PlaneMyrror:
                points = myrrorNodes()
        }

        type NodePath struct {
            Point image.Point
            Path pathfinding.Path
        }

        var paths []NodePath

        for _, point := range points {
            node := stackMap.GetMagicNode(point.X, point.Y)
            // ignore nodes that we already own
            if node.MeldingWizard == self {
                continue
            }

            if self.IsExplored(point.X, point.Y, stack.Plane()) {
                path, ok := aiServices.FindPath(stack.X(), stack.Y(), point.X, point.Y, self, stack, self.GetFog(stack.Plane()))
                if ok {
                    paths = append(paths, NodePath{Point: point, Path: path})
                }
            }
        }

        if len(paths) > 0 {

            slices.SortFunc(paths, func(a, b NodePath) int {
                return cmp.Compare(len(a.Path), len(b.Path))
            })

            if stack.ActiveUnitsHasAbility(data.AbilityMeld) {
                for _, nodePath := range paths {
                    encounter := stackMap.GetEncounter(nodePath.Point.X, nodePath.Point.Y)

                    // the node is free, so we can meld it with a spirit
                    if encounter == nil {
                        if stack.ActiveUnitsHasAbility(data.AbilityMeld) {
                            // move towards node
                            decisions = append(decisions, &playerlib.AIMoveStackDecision{
                                Stack: stack,
                                Path: nodePath.Path,
                            })
                        }
                    } else if stackPower > rawUnitAttackPower(encounter.Units...) - rand.N(int(strength * 10)) {
                        decisions = append(decisions, &playerlib.AIMoveStackDecision{
                            Stack: stack,
                            Path: nodePath.Path,
                        })
                    }
                }
            }
        }
    }

    // return a count of all the nodes that are not melded by the current player, and are not protected
    // by guardians
    countUnmeldedNodes := func(plane data.Plane) int {
        var points []image.Point
        switch plane {
            case data.PlaneArcanus:
                points = arcanusNodes()
            case data.PlaneMyrror:
                points = myrrorNodes()
        }

        useMap := aiServices.GetMap(plane)

        total := 0
        for _, point := range points {
            if self.IsExplored(point.X, point.Y, plane) && useMap.GetEncounter(point.X, point.Y) == nil {
                node := useMap.GetMagicNode(point.X, point.Y)
                if node != nil && node.MeldingWizard != self {
                    total += 1
                }
            }
        }

        return total
    }

    countMelders := func(plane data.Plane) int {
        total := 0
        for _, stack := range self.Stacks {
            if stack.Plane() == plane && stack.ActiveUnitsHasAbility(data.AbilityMeld) {
                total += 1
            }
        }

        return total
    }

    // maybe summon a spirit
    if !self.CastingSpell.Valid() && (self.Mana > 100 || self.ManaPerTurn(aiServices.ComputePower(self), aiServices) > 0) {
        fortress := self.FindFortressCity()
        if fortress != nil {
            unmeldedNodes := countUnmeldedNodes(fortress.Plane)
            melders := countMelders(fortress.Plane)

            if unmeldedNodes > melders && chance(int(strength * 100)) {
                magicSpiritSpell := self.KnownSpells.FindByName("Magic Spirit")
                guardianSpiritSpell := self.KnownSpells.FindByName("Guardian Spirit")

                // wizards should always know magic spirit
                use := magicSpiritSpell
                if guardianSpiritSpell.Valid() {
                    use = guardianSpiritSpell
                }

                if use.Valid() {
                    decisions = append(decisions, &playerlib.AICastSpellDecision{
                        Spell: use,
                    })
                }
            }
        }
    }

    return decisions
}

func (ai *EnemyNetAI) DoIncreasePopulation(self *playerlib.Player, aiServices playerlib.AIServices, strength float64) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    populationBuildings := set.MakeSet[buildinglib.Building]()
    populationBuildings.Insert(buildinglib.BuildingHousing)

    infos := aiServices.GetBuildingInfos()
    for _, building := range []buildinglib.Building{buildinglib.BuildingGranary, buildinglib.BuildingFarmersMarket,
        buildinglib.BuildingBuildersHall, buildinglib.BuildingSawmill} {
            populationBuildings.InsertMany(infos.TransitiveDependencies(building)...)
    }

    producingPopulation := func(city *citylib.City) bool {
        if city.ProducingBuilding == buildinglib.BuildingHousing {
            return true
        }

        return populationBuildings.Contains(city.ProducingBuilding)
    }

    buildings := populationBuildings.Values()

    for _, city := range self.Cities {

        // whatever the city is currently doing, don't change it
        if city.Population >= city.MaximumCitySize() {
            continue
        }

        // if city is already producing something population related then don't change it
        // FIXME: if the city is producing housing then it might be worthwhile to switch
        // to one of the population buildings
        if producingPopulation(city) {
            continue
        }

        if buildPercent(city) < strength {
            buildable := city.GetBuildableBuildings()
            for _, index := range rand.Perm(len(buildings)) {
                check := buildings[index]
                if check == buildinglib.BuildingHousing || buildable.Contains(check) {
                    decisions = append(decisions, &playerlib.AIProduceDecision{
                        City: city,
                        Building: check,
                        Unit: units.UnitNone,
                    })
                    break
                }
            }
        }
    }

    return decisions
}

func (ai *EnemyNetAI) DoIncreasePower(self *playerlib.Player, aiServices playerlib.AIServices, strength float64) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    type DependencyFunc func(d data.Race, building buildinglib.Building) bool
    type BuildingCheckFunc func(building buildinglib.Building) bool

    checkDependency := func(kind BuildingCheckFunc) DependencyFunc { 

        return func(race data.Race, building buildinglib.Building) bool {
            infos := aiServices.GetBuildingInfos()

            dependencies := set.NewSet[buildinglib.Building]()
            for _, building := range buildinglib.RacialBuildings(race).Values() {
                if kind(building) {
                    dependencies.InsertMany(infos.Dependencies(building)...)
                }
            }

            return dependencies.Contains(building)
        }
    }

    isReligiousDependency := functional.Memoize2(checkDependency(buildinglib.Building.IsReligious))
    isEconomicDependency := functional.Memoize2(checkDependency(buildinglib.Building.IsEconomic))
    isFoodDependency := functional.Memoize2(checkDependency(buildinglib.Building.ProducesFood))

    // feels awkward to build buildings in cities here
    for _, city := range self.Cities {
        if !isMakingSomething(city) {
            // create housing
            switch {
                case city.Citizens() < 3:
                    decisions = append(decisions, &playerlib.AIProduceDecision{
                        City: city,
                        Building: buildinglib.BuildingHousing,
                        Unit: units.UnitNone,
                    })
                case self.GoldPerTurn() < 0 || self.Gold < 10 - self.GoldPerTurn():
                    decisions = append(decisions, &playerlib.AIProduceDecision{
                        City: city,
                        Building: buildinglib.BuildingTradeGoods,
                        Unit: units.UnitNone,
                    })
                case chance(int(strength * 100)):

                    // FIXME: if unrest is high then build a shrine/temple/etc
                    // if money production is low then build a marketplace/bank/etc
                    // if food production/population growth is low then build a granary/farmers market/etc
                    // otherwise build a random building

                    possibleBuildings := city.ComputePossibleBuildings(true)
                    values := possibleBuildings.Values()

                    goldSurplus := city.GoldSurplus()

                    needsFood := city.Citizens() < city.MaximumCitySize() / 2 || city.PopulationGrowthRate() < 30

                    weights := make([]int, 0, possibleBuildings.Size())
                    for _, building := range values {
                        weight := 1

                        if city.Rebels > 0 {
                            if building.IsReligious() {
                                weight += 2
                            } else if isReligiousDependency(city.Race, building) {
                                weight += 1
                            }
                        }

                        if goldSurplus < 0 {
                            if building.IsEconomic() {
                                weight += 2
                            } else if isEconomicDependency(city.Race, building) {
                                weight += 1
                            }
                        }

                        if needsFood {
                            if building.ProducesFood() {
                                weight += 2
                            } else if isFoodDependency(city.Race, building) {
                                weight += 1
                            }
                        }

                        // FIXME: also consider dependencies of buildings that produce religion/gold/food

                        weights = append(weights, weight)
                    }

                    if possibleBuildings.Size() > 0 {
                        // choose a random building to create
                        decisions = append(decisions, &playerlib.AIProduceDecision{
                            City: city,
                            Building: algorithm.ChoseRandomWeightedElement(values, weights),
                            Unit: units.UnitNone,
                        })
                    }
            }
        }
    }

    return decisions
}

func (ai *EnemyNetAI) DoDefendCities(self *playerlib.Player, aiServices playerlib.AIServices, strength float64) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    // units that are roaming around should move towards a reasonably close city
    // cities should build army units to defend themselves

    cityStackInfo := aiServices.ComputeCityStackInfo()

    for _, stack := range self.Stacks {
        // stack already at a city, nothing to do
        if self.FindCity(stack.X(), stack.Y(), stack.Plane()) != nil {
            continue
        }

        if len(stack.CurrentPath) > 0 {
            continue
        }

        cityPaths := make(map[*citylib.City]pathfinding.Path)

        var cityChoices []*citylib.City
        for _, city := range self.Cities {
            if city.Plane == stack.Plane() {

                stacks := cityStackInfo.ArcanusStacks
                if city.Plane == data.PlaneMyrror {
                    stacks = cityStackInfo.MyrrorStacks
                }

                cityStack, hasStack := stacks[image.Pt(city.X, city.Y)]
                // city would be too large to hold this stack, so skip the city
                if hasStack && cityStack.Size() + stack.Size() > data.MaxUnitsInStack {
                    continue
                }

                path, ok := aiServices.FindPath(stack.X(), stack.Y(), city.X, city.Y, self, stack, self.GetFog(city.Plane))
                if ok {
                    cityChoices = append(cityChoices, city)
                    cityPaths[city] = path
                }
            }
        }

        slices.SortFunc(cityChoices, func(a, b *citylib.City) int {
            scoreA := len(cityPaths[a]) + unitAttackPower(a.GetGarrison()...)
            scoreB := len(cityPaths[b]) + unitAttackPower(b.GetGarrison()...)

            return cmp.Compare(scoreA, scoreB)
        })

        if len(cityChoices) > 0 {
            choice := cityChoices[0]
            decisions = append(decisions, &playerlib.AIMoveStackDecision{
                Stack: stack,
                Path: cityPaths[choice],
            })
        }
    }

    for _, city := range self.Cities {
        garrison := city.GetGarrison()
        if unitAttackPower(garrison...) < int(strength * 40) {
            if city.ProducingUnit.Equals(units.UnitNone) {
                possibleUnits := city.ComputePossibleUnits()
                possibleUnits = slices.DeleteFunc(possibleUnits, func(unit units.Unit) bool {
                    return rawUnitAttackPower(unit) == 0
                })

                if len(possibleUnits) > 0 {
                    decisions = append(decisions, &playerlib.AIProduceDecision{
                        City: city,
                        Building: buildinglib.BuildingNone,
                        Unit: algorithm.ChoseRandomWeightedElement(possibleUnits, functional.Map(possibleUnits, func(unit units.Unit) int {
                            return rawUnitAttackPower(unit)
                        })),
                    })
                }
            }
        }
    }

    return decisions
}

// the operational manager takes in the strategy probabilities and selects specific actions to do, which are returned as AIDecisions from the Update function
func (ai *EnemyNetAI) OperationalManager(player *playerlib.Player, services playerlib.AIServices, strategies []Probability) []playerlib.AIDecision {

    var decisions []playerlib.AIDecision

    for _, strategy := range strategies {
        switch Strategy(strategy.Index) {
            case StrategyAttackEnemies:
                // if there are known enemy units/cities then choose a set of units to attack with and move towards a target
                // if strategy value is very high then attack targets even if our unit is weak

                decisions = append(decisions, ai.DoAttackEnemies(player, services, strategy.Value)...)
            case StrategyBuildArmy:
                // set cities to build military units, and move existing units to form armies

                decisions = append(decisions, ai.DoBuildArmy(player, services, strategy.Value)...)
            case StrategyAcquireMagicNode:
                // find magic nodes (either unconquered or enemy controlled) and move towards them with units that can capture them
                decisions = append(decisions, ai.DoAcquireMagicNodes(player, services, strategy.Value)...)
            case StrategyBuildCities:
                // find good locations to build new cities, and move settlers to those locations to found new cities
                // also set cities to build settlers
                decisions = append(decisions, ai.DoBuildCities(player, services, strategy.Value)...)
            case StrategyDefendCities:
                // move units to defend cities that are under attack or likely to be attacked, and set city garrisons
                decisions = append(decisions, ai.DoDefendCities(player, services, strategy.Value)...)
            case StrategyIncreasePopulation:
                // set cities to build housing or buildings that increase population
                decisions = append(decisions, ai.DoIncreasePopulation(player, services, strategy.Value)...)
            case StrategyIncreasePower:
                // set cities to build magic buildings
                decisions = append(decisions, ai.DoIncreasePower(player, services, strategy.Value)...)
        }
    }

    return decisions
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

    losses := make([]float64, int(StrategyCount))

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

func (ai *EnemyNetAI) DidBanish(self *playerlib.Player, other *playerlib.Player) {
    ai.Stats.EnemiesBanished += 1
}

func (ai *EnemyNetAI) DidDefeat(self *playerlib.Player, other *playerlib.Player) {
    ai.Stats.EnemiesDefeated += 1
}

func (ai *EnemyNetAI) DidSummonUnit(self *playerlib.Player, unit *units.OverworldUnit) {
    ai.Stats.UnitsSummoned += 1
}

func (ai *EnemyNetAI) PostUpdate(player *playerlib.Player, services playerlib.AIServices) {
    // compute rewards
    // rewards are any event that can be quantified, like how many enemy units were killed, how much gold was gained, how many cities were captured, etc

    var reward float64 = 0

    ai.Stats.GoldDelta = player.Gold - ai.currentGold
    ai.Stats.ManaDelta = player.Mana - ai.currentMana

    if !ai.banished && player.Banished {
        ai.Stats.WasBanished = 1
    }

    if !ai.defeated && player.Defeated {
        ai.Stats.WasDefeated = 1
    }

    // all these values picked on vibes. maybe a neural net can learn them?
    reward -= float64(ai.Stats.WasBanished) * 10000
    reward -= float64(ai.Stats.WasDefeated) * 5000
    reward += float64(ai.Stats.EnemiesBanished) * 1000
    reward += float64(ai.Stats.EnemiesDefeated) * 40
    reward -= float64(ai.Stats.UnitsLost)
    reward += float64(ai.Stats.UnitsCreated) * 0.8
    reward += float64(ai.Stats.UnitsSummoned) * 1.5
    reward += float64(ai.Stats.CitiesRazed) * 10
    reward += float64(ai.Stats.CitiesCaptured) * 20
    reward -= float64(ai.Stats.CitiesLost) * 20
    reward += float64(ai.Stats.MagicNodesGained) * 15
    reward -= float64(ai.Stats.MagicNodesLost) * 15
    reward += float64(ai.Stats.GoldDelta) * 0.3
    reward += float64(ai.Stats.ManaDelta) * 0.3
    reward += float64(ai.Stats.TerritoryExplored) * 0.1
    reward += float64(ai.Stats.SpellsLearned) * 5
    reward += float64(ai.Stats.HeroesGained) * 20
    reward -= float64(ai.Stats.HeroesLost) * 20
    reward += float64(ai.Stats.ArmyStrengthDelta) * 0.5
    reward += float64(ai.Stats.RoadsBuilt) * 0.2
    reward += float64(ai.Stats.EnemiesDiscovered) * 1.3
    reward += ai.Stats.SpellOfMasteryProgress * 50

    ai.Steps[len(ai.Steps)-1].Reward = reward

    log.Printf("Reward: %f", reward)

    // merge stacks that are on top of each other
    type Location struct {
        X, Y int
        Plane data.Plane
    }

    var stackLocations []Location

    for _, stack := range player.Stacks {
        stackLocations = append(stackLocations, Location{X: stack.X(), Y: stack.Y(), Plane: stack.Plane()})
    }

    for _, location := range stackLocations {
        stacks := player.FindAllStacks(location.X, location.Y, location.Plane)
        for len(stacks) > 1 {
            player.MergeStacks(stacks[0], stacks[1])
            stacks = player.FindAllStacks(location.X, location.Y, location.Plane)
        }
    }

    // make sure food is balanced at the end
    player.RebalanceFood()
}

func (ai *EnemyNetAI) PreTurn(player *playerlib.Player) {
    // reset stats
    ai.Stats = PlayerStats{}

    ai.currentGold = player.Gold
    ai.currentMana = player.Mana

    ai.banished = player.Banished
    ai.defeated = player.Defeated
}

func (ai *EnemyNetAI) NewTurn(player *playerlib.Player) {
    ai.Attacking = make(map[*playerlib.UnitStack]bool)
}

func (ai *EnemyNetAI) ProducedUnit(city *citylib.City, player *playerlib.Player) {
    city.ProducingBuilding = buildinglib.BuildingTradeGoods
    city.ProducingUnit = units.UnitNone
}

func (ai *EnemyNetAI) ConfirmRazeTown(city *citylib.City) bool {
    return false
}

func (ai *EnemyNetAI) HandleMerchantItem(self *playerlib.Player, item *artifact.Artifact, cost int) bool {
    if self.Gold >= cost {
        for _, hero := range self.Heroes {
            if hero != nil && hero.Status == herolib.StatusEmployed {
                slots := hero.GetArtifactSlots()
                for i := range hero.Equipment {
                    if hero.Equipment[i] == nil && slots[i].CompatibleWith(item.Type) {
                        hero.Equipment[i] = item
                        log.Printf("AI %v bought artifact %v for %v gold, and gave it to hero %v", self.Wizard.Name, item.Name, cost, hero.Name)
                        return true
                    }
                }
            }
        }

        for i := range self.VaultEquipment {
            // FIXME: possibly replace an artifact
            if self.VaultEquipment[i] == nil {
                self.VaultEquipment[i] = item
                self.Gold -= cost
                log.Printf("AI %v bought artifact %v for %v gold, and placed it in the vault", self.Wizard.Name, item.Name, cost)
                return true
            }
        }
    }

    return false
}

func (ai *EnemyNetAI) HandleHireHero(self *playerlib.Player, hero *herolib.Hero, cost int, atFortress bool, point data.PlanePoint) {
    goldPerTurn := self.GoldPerTurn()

    // always try to hire a hero if we can afford it
    if self.Gold >= cost && goldPerTurn >= hero.GetUpkeepGold() {
        added := false
        if atFortress {
            added = self.AddHeroToFortress(hero)
        } else {
            added = self.AddHero(hero, point.X, point.Y, point.Plane)
        }

        if added {
            log.Printf("AI %v hired hero %v for %v gold", self.Wizard.Name, hero.Name, cost)
            self.Gold -= cost
            hero.SetStatus(herolib.StatusEmployed)

            // FIXME: consider invoking this method
            // game.ResolveStackAt(hero.GetX(), hero.GetY(), hero.GetPlane())

            if self.SelectedStack == nil {
                self.SelectedStack = self.FindStack(hero.GetX(), hero.GetY(), hero.GetPlane())
            }
        }
    }
}

func (ai *EnemyNetAI) HandleHireMercenaries(self *playerlib.Player, mercenaries []*units.OverworldUnit, cost int) {
    neededGoldPerTurn := 0
    for _, mercenary := range mercenaries {
        neededGoldPerTurn += mercenary.GetUpkeepGold()
    }

    if self.Gold >= cost && self.GoldPerTurn() > neededGoldPerTurn && self.FoodPerTurn() > 0 {
        log.Printf("AI %v hired %v mercenaries for %v gold", self.Wizard.Name, len(mercenaries), cost)
        for _, unit := range mercenaries {
            self.AddUnit(unit)
            // FIXME: consider invoking this method
            // game.ResolveStackAt(unit.GetX(), unit.GetY(), unit.GetPlane())
        }
        self.Gold -= cost
    }
}

func (ai *EnemyNetAI) InvalidMove(stack *playerlib.UnitStack) {
}

func (ai *EnemyNetAI) MovedStack(stack *playerlib.UnitStack, path pathfinding.Path) pathfinding.Path {
    // after moving towards an enemy, clear the current path so a new path can be computed next turn
    // maybe the enemy is no longer there, so there is no point in moving towards it
    _, ok := ai.Attacking[stack]
    if ok {
        return nil
    }

    return path
}

func (ai *EnemyNetAI) ConfirmEncounter(stack *playerlib.UnitStack, encounter *maplib.ExtraEncounter) bool {
    return true
}
