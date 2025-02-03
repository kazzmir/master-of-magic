package ai

/* EnemyAI is the AI for enemy wizards
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
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
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

func (ai *EnemyAI) Update(self *playerlib.Player, enemies []*playerlib.Player, aiServices playerlib.AIServices, manaPerTurn int) []playerlib.AIDecision {
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

    // not casting a spell
    if self.CastingSpell.Invalid() && rand.N(10) == 0 {
        // just search for summoning spells for now

        summoningSpells := self.KnownSpells.GetSpellsBySection(spellbook.SectionSummoning)
        if len(summoningSpells.Spells) > 0 {
            for _, i := range rand.Perm(len(summoningSpells.Spells)) {
                chosen := summoningSpells.Spells[i]
                summonUnit := units.GetUnitByName(chosen.Name)
                // check unit.UpkeepMana to see if it is affordable
                if !summonUnit.IsNone() && manaPerTurn >= summonUnit.UpkeepMana {
                    decisions = append(decisions, &playerlib.AICastSpellDecision{
                        Spell: chosen,
                    })
                    break
                }
            }
        }

    }

    for _, city := range self.Cities {
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
            // FIXME: enter cities, lairs, nodes for combat
            // also, sometimes choose a preferred location to move to, such as a square for building a new city
            // or attacking a player's units
            if len(stack.CurrentPath) == 0 {
                // handling for settlers
                if stack.ActiveUnitsHasAbility(data.AbilityCreateOutpost) {
                    // find a location on the same continent as the stack that we can build a new outpost
                    // if we can't find a location, just move randomly
                    // if we are at a settlable location, build the outpost
                    // otherwise, find a path to the chosen location

                    candidateLocations := aiServices.FindSettlableLocations(stack.X(), stack.Y(), stack.Plane())
                    if len(candidateLocations) == 0 {

                        // check if the settler is already in a city
                        if self.FindCity(stack.X(), stack.Y(), stack.Plane()) == nil {
                            // just go back to a town?
                            var candidateCities []*citylib.City
                            for _, city := range self.Cities {
                                if city.Plane == stack.Plane() && len(aiServices.FindPath(stack.X(), stack.Y(), city.X, city.Y, stack, self.GetFog(stack.Plane()))) > 0 {
                                    candidateCities = append(candidateCities, city)
                                }
                            }

                            if len(candidateCities) > 0 {
                                // sort cities by distance
                                infinity := 999999

                                getDistance := functional.Memoize(func (city *citylib.City) int {
                                    path := aiServices.FindPath(stack.X(), stack.Y(), city.X, city.Y, stack, self.GetFog(stack.Plane()))
                                    if len(path) == 0 {
                                        return infinity
                                    }
                                    return len(path)
                                })

                                slices.SortFunc(candidateCities, func(a, b *citylib.City) int {
                                    return cmp.Compare(getDistance(a), getDistance(b))
                                })

                                path := aiServices.FindPath(stack.X(), stack.Y(), candidateCities[0].X, candidateCities[0].Y, stack, self.GetFog(stack.Plane()))
                                stack.CurrentPath = path
                            } else {
                                // do nothing
                            }
                        }
                    } else {
                        // choose a random location
                        location := candidateLocations[rand.N(len(candidateLocations))]
                        path := aiServices.FindPath(stack.X(), stack.Y(), location.X, location.Y, stack, self.GetFog(stack.Plane()))
                        if len(path) > 0 {
                            stack.CurrentPath = path
                        }
                    }
                } else if rand.N(4) == 0 {
                    // try upto 3 times to find a path
                    for range 3 {
                        newX, newY := stack.X() + rand.N(5) - 2, stack.Y() + rand.N(5) - 2
                        path := aiServices.FindPath(stack.X(), stack.Y(), newX, newY, stack, self.GetFog(stack.Plane()))
                        if len(path) != 0 {
                            stack.CurrentPath = path
                            break
                        }
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
                    ConfirmEncounter: func (encounter *maplib.ExtraEncounter) bool {
                        return true
                    },
                })
            }
        }
    }

    return decisions
}

func (ai *EnemyAI) PostUpdate(self *playerlib.Player, enemies []*playerlib.Player) {

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
