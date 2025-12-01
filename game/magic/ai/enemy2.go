package ai

/* EnemyAI is the AI for enemy wizards
 * Actions an enemy can take:
 *  research new spell
 *  cast spell
 *  each city can be producing something, or housing/trade goods
 *  each unit can patrol to defend an area, move to attack an enemy, or move to explore
 *  diplomacy with other wizards
 *
 * given a goal, create a plan. each turn the AI will try to execute part of the plan
 * after exploration, the AI may need to replan
 */

import (
    "log"
    "slices"
    "cmp"
    "math/rand/v2"
    "image"

    "github.com/kazzmir/master-of-magic/lib/functional"
    "github.com/kazzmir/master-of-magic/lib/set"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    // "github.com/kazzmir/master-of-magic/game/magic/spellbook"
)

type Enemy2AI struct {
    // units that are currently moving towards an enemy
    Attacking map[*playerlib.UnitStack]bool
}

func MakeEnemy2AI() *Enemy2AI {
    return &Enemy2AI{}
}

type GoalType int

const (
    GoalNone GoalType = iota
    GoalDefeatEnemies // defeat enemy wizards
    GoalBuildCities
    GoalExploreTerritory // explore the map
    GoalResearchMagic // research spells
    GoalCastSpellOfMastery
    GoalDefendTerritory
    GoalBuildArmy
    GoalIncreasePower
    GoalMeldNodes
)

type EnemyGoal struct {
    Goal GoalType
    Weight float32 // normalized between 0 and 1

    // subgoals that must be satisfied for the current goal before an action
    // can be taken towards the current goal
    SubGoals []EnemyGoal
}

// weight of this goal, higher weight means higher priority
func (goal *EnemyGoal) GetWeight() float32 {
    return goal.Weight
}

func chance(percent int) bool {
    return rand.N(100) < percent
}

func (ai *Enemy2AI) ComputeGoals(self *playerlib.Player, aiServices playerlib.AIServices) []EnemyGoal {
    var goals []EnemyGoal

    exploreGoal := EnemyGoal{
        Goal: GoalExploreTerritory,
    }

    buildArmyGoal := EnemyGoal{
        Goal: GoalBuildArmy,
    }

    // only build an army if we have few stacks compared to the turn number
    if uint64(len(self.Stacks)) < max(5, aiServices.GetTurnNumber() / 5) {
        exploreGoal.SubGoals = append(exploreGoal.SubGoals, buildArmyGoal)
    }

    goals = []EnemyGoal{
        EnemyGoal{
            Goal: GoalDefeatEnemies,
            SubGoals: []EnemyGoal{
                exploreGoal,
            },
        },
        EnemyGoal{
            Goal: GoalBuildCities,
        },
        EnemyGoal{
            Goal: GoalIncreasePower,
        },
    }

    // goals = append(goals, exploreGoal)

    /*
    goals = append(goals, EnemyGoal{
        Goal: GoalDefeatEnemies,
        SubGoals: []EnemyGoal{
            exploreGoal,
            EnemyGoal{
                Goal: GoalBuildArmy,
            },
        },
    })

    goals = append(goals, EnemyGoal{
        Goal: GoalCastSpellOfMastery,
        SubGoals: []EnemyGoal{
            EnemyGoal{
                Goal: GoalResearchMagic,
            },
            EnemyGoal{
                Goal: GoalIncreasePower,
                SubGoals: []EnemyGoal{
                    EnemyGoal{
                        Goal: GoalMeldNodes,
                    },
                },
            },
        },
    })
    */

    return goals
}

// stop producing that unit
func (ai *Enemy2AI) ProducedUnit(city *citylib.City, player *playerlib.Player) {
    city.ProducingBuilding = buildinglib.BuildingTradeGoods
    city.ProducingUnit = units.UnitNone
}

type AIData struct {
    FoodPerTurn func() int
    GoldPerTurn func() int
}

// the decisions to make for this goal
// seenGoals is a set used to avoid computing the same goal twice
func (ai *Enemy2AI) GoalDecisions(self *playerlib.Player, aiServices playerlib.AIServices, goal EnemyGoal, seenGoals *set.Set[GoalType], aiData *AIData) []playerlib.AIDecision {
    // if we've seen this goal then just return nil
    if seenGoals.Contains(goal.Goal) {
        return nil
    }

    seenGoals.Insert(goal.Goal)

    var decisions []playerlib.AIDecision

    // recursively satisfy subgoals first
    for _, subGoal := range goal.SubGoals {
        decisions = append(decisions, ai.GoalDecisions(self, aiServices, subGoal, seenGoals, aiData)...)
    }

    switch goal.Goal {
        case GoalDefeatEnemies:
            // find possible enemy targets
            for _, enemyPlayer := range aiServices.GetEnemies(self) {
                var possibleTarget []*playerlib.UnitStack
                for _, enemyStack := range enemyPlayer.Stacks {
                    if self.IsVisible(enemyStack.X(), enemyStack.Y(), enemyStack.Plane()) {
                        possibleTarget = append(possibleTarget, enemyStack)
                    }
                }

                if len(possibleTarget) > 0 {
                    for _, stack := range self.Stacks {
                        if stack.HasMoves() {

                            var shortestPath pathfinding.Path
                            for _, target := range possibleTarget {
                                if target.Plane() == stack.Plane() {
                                    pathToEnemy, ok := aiServices.FindPath(stack.X(), stack.Y(), target.X(), target.Y(), self, stack, self.GetFog(stack.Plane()))
                                    if ok {
                                        if len(shortestPath) == 0 || len(pathToEnemy) < len(shortestPath) {
                                            shortestPath = pathToEnemy
                                        }
                                    }
                                }
                            }

                            if len(shortestPath) > 0 {
                                log.Printf("AI %v moving stack at %v,%v to attack enemy via %v", self.Wizard.Name, stack.X(), stack.Y(), shortestPath)
                                decisions = append(decisions, &playerlib.AIMoveStackDecision{
                                    Stack: stack,
                                    Path: shortestPath,
                                })

                                ai.Attacking[stack] = true
                            }
                        }
                    }
                }

                for _, enemyCity := range enemyPlayer.Cities {
                    // in theory we can see cities that on tiles that we have explored in the past
                    if self.IsVisible(enemyCity.X, enemyCity.Y, enemyCity.Plane) {
                    }
                }
            }

        case GoalBuildCities:
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

                        // FIXME: also prioritize locations on shore in case we need to build a navy
                        // if there are no cities on this continent already adjacent to a shore
                        // then possibly prefer shore locations

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
                                    log.Printf("AI moving settler stack at %v,%v to settlable location %v,%v via %v", stack.X(), stack.Y(), location.X, location.Y, path)
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
                if !isMakingSomething(city) && chance(60) {
                    locations := aiServices.FindSettlableLocations(city.X, city.Y, city.Plane, self.GetFog(city.Plane))
                    if len(locations) > 0 && aiData.FoodPerTurn() > 0 && chance(len(locations) * 5) {
                        decisions = append(decisions, &playerlib.AIProduceDecision{
                            City: city,
                            Building: buildinglib.BuildingNone,
                            Unit: units.GetSettlerUnit(city.Race),
                        })

                    }
                }
            }

        case GoalExploreTerritory:
            // in order to explore territory we need units available that are not busy

            // if there are units that are not busy and not moving, then move them to some unexplored location nearby
            // if there are no units available, then see if we can produce more units
            // if we can't sustain more units then cities should wait to grow in population via housing
            // or create more cities with settlers

            // goal 1 might want to take one action for a city, but goal 2 might want to take a conflicting action
            // choose the goal that has a higher weight

            for _, stack := range slices.Clone(self.Stacks) {
                if stack.HasMoves() {

                    if len(stack.CurrentPath) > 0 {
                        decisions = append(decisions, &playerlib.AIMoveStackDecision{
                            Stack: stack,
                            Path: stack.CurrentPath,
                        })
                        continue
                    }

                    if len(stack.CurrentPath) == 0 {

                        // split the stack in half
                        /*
                        if len(stack.Units()) > 1 && rand.N(4) == 0 {
                            stackUnits := stack.Units()
                            stack = self.SplitStack(stack, stackUnits[0:len(stackUnits) / 2])
                        }
                        */
                        // FIXME: maybe emite SplitStack decision

                        var path pathfinding.Path
                        fog := self.GetFog(stack.Plane())
                        useMap := aiServices.GetMap(stack.Plane())

                        // try upto 5 times to find a path
                        distance := 3
                        for range 5 {
                            distance += 3
                            newX, newY := stack.X() + rand.N(distance) - distance / 2, stack.Y() + rand.N(distance) - distance / 2

                            tile := useMap.GetTile(newX, newY)
                            if !tile.Valid() || self.IsExplored(newX, newY, stack.Plane()) {
                                continue
                            }

                            path, _ = aiServices.FindPath(stack.X(), stack.Y(), newX, newY, self, stack, fog)
                            if len(path) != 0 {
                                log.Printf("Explore new location %v,%v via %v", newX, newY, path)
                                break
                            }
                        }

                        // just go somewhere random
                        if len(path) == 0 {
                            newX, newY := stack.X() + rand.N(5) - 2, stack.Y() + rand.N(5) - 2
                            path, _ = aiServices.FindPath(stack.X(), stack.Y(), newX, newY, self, stack, fog)
                        }

                        if len(path) > 0 {
                            decisions = append(decisions, &playerlib.AIMoveStackDecision{
                                Stack: stack,
                                Path: path,
                            })
                        }
                    }
                }
            }

        case GoalBuildArmy:

            computeTransportUnits := functional.Memoize(func(plane data.Plane) int {
                return self.TransportUnits(plane)
            })

            for _, city := range self.Cities {
                if !isMakingSomething(city) {

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

                        if aiData.FoodPerTurn() > 0 && aiData.GoldPerTurn() > 0 && self.Gold > 50 && chance(30) {
                            possibleUnits := city.ComputePossibleUnits()

                            possibleUnits = slices.DeleteFunc(possibleUnits, func(unit units.Unit) bool {
                                if unit.IsSettlers() {
                                    return true
                                }
                                return false
                            })

                            if len(possibleUnits) > 0 {
                                return &playerlib.AIProduceDecision{
                                    City: city,
                                    Building: buildinglib.BuildingNone,
                                    Unit: possibleUnits[rand.N(len(possibleUnits))],
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
        case GoalIncreasePower:
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
                        case aiData.GoldPerTurn() < 0 || self.Gold < 100:
                            decisions = append(decisions, &playerlib.AIProduceDecision{
                                City: city,
                                Building: buildinglib.BuildingTradeGoods,
                                Unit: units.UnitNone,
                            })
                        case chance(40):
                            possibleBuildings := city.ComputePossibleBuildings(true)
                            if possibleBuildings.Size() > 0 {
                                // choose a random building to create
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

        default:
            log.Printf("WARNING: unhandled goal %v", goal.Goal)
    }

    return decisions
}

func (ai *Enemy2AI) Update(self *playerlib.Player, aiServices playerlib.AIServices) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision
    goals := ai.ComputeGoals(self, aiServices)

    // compute these values once for all goals
    aiData := AIData{
        FoodPerTurn: functional.Memoize0(func() int {
            return self.FoodPerTurn()
        }),
        GoldPerTurn: functional.Memoize0(func() int {
            return self.GoldPerTurn()
        }),
    }

    seenGoals := set.MakeSet[GoalType]()
    for _, goal := range goals {
        decisions = append(decisions, ai.GoalDecisions(self, aiServices, goal, seenGoals, &aiData)...)
    }

    return decisions
}

func (ai *Enemy2AI) ConfirmEncounter(stack *playerlib.UnitStack, encounter *maplib.ExtraEncounter) bool {
    return false
}

func (ai *Enemy2AI) InvalidMove(stack *playerlib.UnitStack) {
}

func (ai *Enemy2AI) MovedStack(stack *playerlib.UnitStack, path pathfinding.Path) pathfinding.Path {
    // after moving towards an enemy, clear the current path so a new path can be computed next turn
    // maybe the enemy is no longer there, so there is no point in moving towards it
    _, ok := ai.Attacking[stack]
    if ok {
        return nil
    }

    return path
}

func (ai *Enemy2AI) PostUpdate(self *playerlib.Player, aiServices playerlib.AIServices) {

    // merge stacks that are on top of each other
    type Location struct {
        X, Y int
        Plane data.Plane
    }

    var stackLocations []Location

    for _, stack := range self.Stacks {
        stackLocations = append(stackLocations, Location{X: stack.X(), Y: stack.Y(), Plane: stack.Plane()})
    }

    for _, location := range stackLocations {
        stacks := self.FindAllStacks(location.X, location.Y, location.Plane)
        for len(stacks) > 1 {
            self.MergeStacks(stacks[0], stacks[1])
            stacks = self.FindAllStacks(location.X, location.Y, location.Plane)
        }
    }

    // make sure food is balanced at the end
    self.RebalanceFood()
}

func (ai *Enemy2AI) NewTurn(player *playerlib.Player) {
    // make sure cities have enough farmers
    for _, city := range player.Cities {
        city.ResetCitizens()
    }

    player.RebalanceFood()

    ai.Attacking = make(map[*playerlib.UnitStack]bool)

    for _, city := range player.Cities {
        log.Printf("ai %v city %v farmer=%v worker=%v rebel=%v", player.Wizard.Name, city.Name, city.Farmers, city.Workers, city.Rebels)
    }
}

func (ai *Enemy2AI) ConfirmRazeTown(city *citylib.City) bool {
    return false
}

func (ai *Enemy2AI) HandleMerchantItem(self *playerlib.Player, item *artifact.Artifact, cost int) bool {
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
