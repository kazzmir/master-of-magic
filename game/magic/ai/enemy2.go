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

    "github.com/kazzmir/master-of-magic/lib/functional"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
)

type Enemy2AI struct {
}

func MakeEnemy2AI() *Enemy2AI {
    return &Enemy2AI{}
}

type GoalType int

const (
    GoalNone GoalType = iota
    GoalDefeatEnemies // defeat enemy wizards
    GoalExpandTerritory // build new cities
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

func (ai *Enemy2AI) ComputeGoals(self *playerlib.Player, aiServices playerlib.AIServices) []EnemyGoal {
    var goals []EnemyGoal

    exploreGoal := EnemyGoal{
        Goal: GoalExploreTerritory,
    }

    goals = []EnemyGoal{
        EnemyGoal{
            Goal: GoalDefeatEnemies,
            SubGoals: []EnemyGoal{
                exploreGoal,
            },
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

// the decisions to make for this goal
func (ai *Enemy2AI) GoalDecisions(self *playerlib.Player, aiServices playerlib.AIServices, goal EnemyGoal) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    // recursively satisfy subgoals first
    for _, subGoal := range goal.SubGoals {
        decisions = append(decisions, ai.GoalDecisions(self, aiServices, subGoal)...)
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
                                    pathToEnemy := aiServices.FindPath(stack.X(), stack.Y(), target.X(), target.Y(), self, stack, self.GetFog(stack.Plane()))
                                    if len(pathToEnemy) > 0 {
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

                            path = aiServices.FindPath(stack.X(), stack.Y(), newX, newY, self, stack, fog)
                            if len(path) != 0 {
                                log.Printf("Explore new location %v,%v via %v", newX, newY, path)
                                break
                            }
                        }

                        // just go somewhere random
                        if len(path) == 0 {
                            newX, newY := stack.X() + rand.N(5) - 2, stack.Y() + rand.N(5) - 2
                            path = aiServices.FindPath(stack.X(), stack.Y(), newX, newY, self, stack, fog)
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

            foodPerTurn := self.FoodPerTurn()

            for _, city := range self.Cities {
                if !isMakingSomething(city) && foodPerTurn > 0 {
                    possibleUnits := city.ComputePossibleUnits()

                    possibleUnits = slices.DeleteFunc(possibleUnits, func(unit units.Unit) bool {
                        if unit.IsSettlers() {
                            return true
                        }
                        return false
                    })

                    if len(possibleUnits) > 0 {
                        decisions = append(decisions, &playerlib.AIProduceDecision{
                            City: city,
                            Building: buildinglib.BuildingNone,
                            Unit: possibleUnits[rand.N(len(possibleUnits))],
                        })
                    }
                }
            }
    }

    return decisions
}

func (ai *Enemy2AI) Update(self *playerlib.Player, aiServices playerlib.AIServices) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision
    goals := ai.ComputeGoals(self, aiServices)

    for _, goal := range goals {
        decisions = append(decisions, ai.GoalDecisions(self, aiServices, goal)...)
    }

    return decisions
}

func (ai *Enemy2AI) ConfirmEncounter(stack *playerlib.UnitStack, encounter *maplib.ExtraEncounter) bool {
    return false
}

func (ai *Enemy2AI) InvalidMove(stack *playerlib.UnitStack) {
}

func (ai *Enemy2AI) MovedStack(stack *playerlib.UnitStack) {
}

func (ai *Enemy2AI) Update2(self *playerlib.Player, aiServices playerlib.AIServices) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    // FIXME: create settlers, build cities

    if self.ResearchingSpell.Invalid() {
        if len(self.ResearchCandidateSpells.Spells) > 0 {
            // choose cheapest research cost spell
            choice := self.ResearchCandidateSpells.Spells[0]
            for _, spell := range self.ResearchCandidateSpells.Spells {
                if spell.ResearchCost < choice.ResearchCost {
                    choice = spell
                }
            }
            decisions = append(decisions, &playerlib.AIResearchSpellDecision{
                Spell: choice,
            })
        }
    }

    manaPerTurn := functional.Memoize0(func() int {
        return self.ManaPerTurn(aiServices.ComputePower(self), aiServices)
    })

    // not casting a spell
    if self.CastingSpell.Invalid() && rand.N(10) == 0 {
        // just search for summoning spells for now

        summoningSpells := self.KnownSpells.GetSpellsBySection(spellbook.SectionSummoning)
        if len(summoningSpells.Spells) > 0 {
            for _, i := range rand.Perm(len(summoningSpells.Spells)) {
                chosen := summoningSpells.Spells[i]
                summonUnit := units.GetUnitByName(chosen.Name)
                // check unit.UpkeepMana to see if it is affordable
                if !summonUnit.IsNone() && manaPerTurn() >= summonUnit.UpkeepMana {
                    decisions = append(decisions, &playerlib.AICastSpellDecision{
                        Spell: chosen,
                    })
                    break
                }
            }
        }

    }

    for _, city := range self.Cities {
        // outpost can't do anything
        if city.Outpost {
            continue
        }

        // city can make something
        if !isMakingSomething(city) {
            possibleUnits := city.ComputePossibleUnits()
            settlers := units.UnitNone

            for _, unit := range possibleUnits {
                if unit.IsSettlers() {
                    settlers = unit
                    break
                }
            }

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
            const ChooseSettlers Choice = 3

            var choices []Choice
            if len(possibleUnits) > 0 {
                stack := self.FindStack(city.X, city.Y, city.Plane)
                if stack == nil || len(stack.Units()) < 9 {
                    choices = append(choices, ChooseUnit)
                }
            }

            if !settlers.IsNone() && city.Citizens() > 5 {
                choices = append(choices, ChooseSettlers)
            }

            if possibleBuildings.Size() > 0 {
                choices = append(choices, ChooseBuilding)
            }

            if len(choices) > 0 {
                switch choices[rand.N(len(choices))] {
                    case ChooseSettlers:
                        decisions = append(decisions, &playerlib.AIProduceDecision{
                            City: city,
                            Building: buildinglib.BuildingNone,
                            Unit: settlers,
                        })
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
        // don't move if this would leave a city undefended, otherwise try to split the stack and move part of it
        if stack.HasMoves() {

            if len(stack.CurrentPath) > 0 {
                decisions = append(decisions, &playerlib.AIMoveStackDecision{
                    Stack: stack,
                    Path: stack.CurrentPath,
                })
                continue
            } else {
                // FIXME: enter cities, lairs, nodes for combat
                // also, sometimes choose a preferred location to move to, such as a square for building a new city
                // or attacking a player's units

                // a stack of only settlers shouldn't move
                nonSettlers := false
                for _, unit := range stack.ActiveUnits() {
                    if !unit.HasAbility(data.AbilityCreateOutpost) {
                        nonSettlers = true
                        break
                    }
                }

                var path pathfinding.Path

                // handling for settlers
                if stack.ActiveUnitsHasAbility(data.AbilityCreateOutpost) {
                    // find a location on the same continent as the stack that we can build a new outpost
                    // if we can't find a location, just move randomly
                    // if we are at a settlable location, build the outpost
                    // otherwise, find a path to the chosen location

                    if aiServices.IsSettlableLocation(stack.X(), stack.Y(), stack.Plane()) {
                        decisions = append(decisions, &playerlib.AIBuildOutpostDecision{
                            Stack: stack,
                        })
                        continue
                    }

                    candidateLocations := aiServices.FindSettlableLocations(stack.X(), stack.Y(), stack.Plane(), self.GetFog(stack.Plane()))
                    if len(candidateLocations) == 0 {

                        // check if the settler is already in a city
                        if self.FindCity(stack.X(), stack.Y(), stack.Plane()) == nil && !nonSettlers {
                            // just go back to a town?
                            var candidateCities []*citylib.City
                            for _, city := range self.Cities {
                                if city.Plane == stack.Plane() && len(aiServices.FindPath(stack.X(), stack.Y(), city.X, city.Y, self, stack, self.GetFog(stack.Plane()))) > 0 {
                                    candidateCities = append(candidateCities, city)
                                }
                            }

                            if len(candidateCities) > 0 {
                                // sort cities by distance
                                infinity := 999999

                                getDistance := functional.Memoize(func (city *citylib.City) int {
                                    path := aiServices.FindPath(stack.X(), stack.Y(), city.X, city.Y, self, stack, self.GetFog(stack.Plane()))
                                    if len(path) == 0 {
                                        return infinity
                                    }
                                    return len(path)
                                })

                                slices.SortFunc(candidateCities, func(a, b *citylib.City) int {
                                    return cmp.Compare(getDistance(a), getDistance(b))
                                })

                                path = aiServices.FindPath(stack.X(), stack.Y(), candidateCities[0].X, candidateCities[0].Y, self, stack, self.GetFog(stack.Plane()))
                            } else {
                                // do nothing
                            }
                        }
                    } else {
                        // FIXME: choose a location with a high population maximum and near bonuses. Possibly also near a shore so we can build water units
                        // choose a random location
                        location := candidateLocations[rand.N(len(candidateLocations))]
                        path = aiServices.FindPath(stack.X(), stack.Y(), location.X, location.Y, self, stack, self.GetFog(stack.Plane()))
                        if len(path) > 0 {
                            log.Printf("Settler going to %v, %v via %v", location.X, location.Y, path)
                        }
                    }
                }

                if nonSettlers && rand.N(2) == 0 && len(stack.CurrentPath) == 0 {
                    // try upto 3 times to find a path
                    for range 3 {
                        newX, newY := stack.X() + rand.N(5) - 2, stack.Y() + rand.N(5) - 2
                        path = aiServices.FindPath(stack.X(), stack.Y(), newX, newY, self, stack, self.GetFog(stack.Plane()))
                        if len(path) != 0 {
                            break
                        }
                    }
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

    return decisions
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
}

func (ai *Enemy2AI) NewTurn(player *playerlib.Player) {
    // make sure cities have enough farmers
    for _, city := range player.Cities {
        city.ResetCitizens()
    }

    player.RebalanceFood()

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
