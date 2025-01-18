package game

import (
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
)

type TreasureItem interface {
}

type TreasureGold struct {
    Amount int
}

type TreasureMana struct {
    Amount int
}

type TreasureMagicalItem struct {
    Artifact artifact.Artifact
}

type TreasurePrisonerHero struct {
    Name string
}

type TreasureSpell struct {
    Spell spellbook.Spell
}

// always gives one spellbook of the specified magic type
type TreasureSpellbook struct {
    Magic data.MagicType
}

type TreasureRetort struct {
    Retort setup.WizardAbility
}

type Treasure struct {
    Treasures []TreasureItem
}

// stolen from maplib/map.go
func chooseValue[T comparable](choices map[T]int) T {
    total := 0
    for _, value := range choices {
        total += value
    }

    pick := rand.N(total)
    for key, value := range choices {
        if pick < value {
            return key
        }

        pick -= value
    }

    var out T
    return out
}

// given some budget, keep choosing a treasure type within the budget and add it to the treasure
func makeTreasure(encounterType maplib.EncounterType, budget int, wizard setup.WizardCustom, knownSpells spellbook.Spells, allSpells spellbook.Spells) *Treasure {
    type TreasureType int
    const (
        TreasureTypeGold TreasureType = iota
        TreasureTypeMana
        TreasureTypeMagicalItem
        TreasureTypePrisonerHero
        TreasureTypeCommonSpell
        TreasureTypeUncommonSpell
        TreasureTypeRareSpell
        TreasureTypeVeryRareSpell
        TreasureTypeSpellbook
        TreasureTypeRetort
    )

    // treasure cannot contain more than these values of each type
    spellsRemaining := 1
    spellBookRemaining := 1
    retortRemaining := 1
    prisonerRemaining := 1
    magicItemRemaining := 3

    const prisonerSpend = 1000
    const spellCommonSpend = 50
    const spellUncommonSpend = 200
    const spellRareSpend = 450
    const spellVeryRareSpend = 800
    const spellBookSpend = 1000
    const retortSpend = 1000
    const magicItemSpend = 400

    var items []TreasureItem

    for budget > 0 {
        choices := make(map[TreasureType]int)
        choices[TreasureTypeGold] = 2
        choices[TreasureTypeMana] = 2
        if spellsRemaining > 0 {
            if budget >= spellCommonSpend {
                choices[TreasureTypeCommonSpell] = 1
            }
            if budget >= spellUncommonSpend {
                choices[TreasureTypeUncommonSpell] = 1
            }
            if budget >= spellRareSpend {
                choices[TreasureTypeRareSpell] = 1
            }
            if budget >= spellVeryRareSpend {
                choices[TreasureTypeVeryRareSpell] = 1
            }
        }
        if spellBookRemaining > 0 && budget >= spellBookSpend {
            choices[TreasureTypeSpellbook] = 2
        }
        if retortRemaining > 0 && budget >= retortSpend {
            choices[TreasureTypeRetort] = 2
        }
        if prisonerRemaining > 0 && budget >= prisonerSpend {
            choices[TreasureTypePrisonerHero] = 1
        }
        if magicItemRemaining > 0 && budget >= magicItemSpend {
            choices[TreasureTypeMagicalItem] = 5
        }

        choice := chooseValue(choices)

        switch choice {
            case TreasureTypeGold:
                coins := rand.N(200) + 1
                coins = min(coins, budget)

                items = append(items, &TreasureGold{Amount: coins})

                budget -= coins
            case TreasureTypeMana:
                mana := rand.N(200) + 1
                mana = min(mana, budget)

                items = append(items, &TreasureMana{Amount: mana})

                budget -= mana
            case TreasureTypeMagicalItem:
                magicItemRemaining -= 1
                // FIXME: give a premade magical item, or create a new one. Take the wizard's spell books into account
                budget -= 400
            case TreasureTypePrisonerHero:
                prisonerRemaining -= 1
                // FIXME: find an alive and unemployed non-champion hero
                budget -= prisonerSpend
            case TreasureTypeCommonSpell:
                common := allSpells.GetSpellsByRarity(spellbook.SpellRarityCommon)
                common.RemoveSpells(knownSpells)

                if len(common.Spells) > 0 {
                    // FIXME: filter by wizard's spell books
                    spell := common.Spells[rand.N(len(common.Spells))]
                    spellsRemaining -= 1
                    // FIXME: choose some unknown common spell given the wizard's spell books

                    items = append(items, &TreasureSpell{Spell: spell})

                    budget -= spellCommonSpend
                }
            case TreasureTypeUncommonSpell:
                spellsRemaining -= 1
                // FIXME: choose some unknown uncommon spell given the wizard's spell books
                budget -= spellUncommonSpend
            case TreasureTypeRareSpell:
                spellsRemaining -= 1
                // FIXME: choose some unknown rare spell given the wizard's spell books
                budget -= spellRareSpend
            case TreasureTypeVeryRareSpell:
                spellsRemaining -= 1
                // FIXME: choose some unknown very rare spell given the wizard's spell books
                budget -= spellVeryRareSpend
            case TreasureTypeSpellbook:
                spellBookRemaining -= 1

                // FIXME: depends on the type of encounter/node
                books := []data.MagicType{data.LifeMagic, data.SorceryMagic, data.ChaosMagic, data.NatureMagic, data.DeathMagic}
                items = append(items, &TreasureSpellbook{Magic: books[rand.N(len(books))]})

                budget = 0
            case TreasureTypeRetort:
                retortRemaining -= 1
                // FIXME: give a retort of any type (maybe not myrror though?)
                budget = 0
        }
    }

    return &Treasure{Treasures: items}
}
