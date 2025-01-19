package game

import (
    "slices"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
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
    Hero *herolib.Hero
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
func makeTreasure(encounterType maplib.EncounterType, budget int, wizard setup.WizardCustom, knownSpells spellbook.Spells, allSpells spellbook.Spells, heroes []*herolib.Hero, getPremadeArtifacts func() []artifact.Artifact) *Treasure {
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

    chooseRetort := func () (setup.WizardAbility, bool) {
        choices := []setup.WizardAbility{
            setup.AbilityAlchemy,
            setup.AbilityWarlord,
            setup.AbilityChanneler,
            setup.AbilityArchmage,
            setup.AbilityArtificer,
            setup.AbilityConjurer,
            setup.AbilitySageMaster,
            setup.AbilityDivinePower,
            setup.AbilityFamous,
            setup.AbilityRunemaster,
            setup.AbilityCharismatic,
            setup.AbilityChaosMastery,
            setup.AbilityNatureMastery,
            setup.AbilitySorceryMastery,
            setup.AbilityInfernalPower,
            setup.AbilityManaFocusing,
            setup.AbilityNodeMastery,
        }

        choices = slices.DeleteFunc(choices, func (ability setup.WizardAbility) bool {
            return wizard.AbilityEnabled(ability)
        })

        if len(choices) > 0 {
            return choices[rand.N(len(choices))], true
        }

        return setup.AbilityNone, false
    }

    chooseSpell := func (rarity spellbook.SpellRarity) (spellbook.Spell, bool) {
        spells := allSpells.GetSpellsByRarity(rarity)
        spells.RemoveSpells(knownSpells)

        var possible spellbook.Spells

        for _, book := range wizard.Books {
            possible.AddAllSpells(spells.GetSpellsByMagic(book.Magic))
        }

        // arcane spells are always available I guess?
        possible.AddAllSpells(spells.GetSpellsByMagic(data.ArcaneMagic))

        if len(possible.Spells) > 0 {
            spell := possible.Spells[rand.N(len(possible.Spells))]
            return spell, true
        }

        return spellbook.Spell{}, false
    }

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
        if prisonerRemaining > 0 && budget >= prisonerSpend && len(heroes) > 0 {
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
                artifacts := getPremadeArtifacts()
                // FIXME: if there are no premade artifacts left, then generate a random new one
                if len(artifacts) > 0 {
                    choice := artifacts[rand.N(len(artifacts))]
                    if budget >= choice.Cost {
                        magicItemRemaining -= 1
                        budget -= choice.Cost
                        items = append(items, &TreasureMagicalItem{Artifact: choice})
                    }
                }
            case TreasureTypePrisonerHero:
                if len(heroes) > 0 {
                    prisonerRemaining -= 1
                    budget -= prisonerSpend
                    hero := heroes[rand.N(len(heroes))]
                    items = append(items, &TreasurePrisonerHero{Hero: hero})
                }
            case TreasureTypeCommonSpell:
                spell, ok := chooseSpell(spellbook.SpellRarityCommon)

                if ok {
                    spellsRemaining -= 1
                    items = append(items, &TreasureSpell{Spell: spell})
                    budget -= spellCommonSpend
                }
            case TreasureTypeUncommonSpell:
                spell, ok := chooseSpell(spellbook.SpellRarityUncommon)

                if ok {
                    spellsRemaining -= 1
                    items = append(items, &TreasureSpell{Spell: spell})
                    budget -= spellUncommonSpend
                }
            case TreasureTypeRareSpell:
                spell, ok := chooseSpell(spellbook.SpellRarityRare)

                if ok {
                    spellsRemaining -= 1
                    items = append(items, &TreasureSpell{Spell: spell})
                    budget -= spellRareSpend
                }
            case TreasureTypeVeryRareSpell:
                spell, ok := chooseSpell(spellbook.SpellRarityVeryRare)

                if ok {
                    spellsRemaining -= 1
                    items = append(items, &TreasureSpell{Spell: spell})
                    budget -= spellVeryRareSpend
                }
            case TreasureTypeSpellbook:

                var books []data.MagicType
                switch encounterType {
                    case maplib.EncounterTypeLair, maplib.EncounterTypeCave, maplib.EncounterTypePlaneTower:
                        books = []data.MagicType{data.LifeMagic, data.SorceryMagic, data.ChaosMagic, data.NatureMagic, data.DeathMagic}

                    case maplib.EncounterTypeAncientTemple, maplib.EncounterTypeFallenTemple:
                        books = []data.MagicType{data.LifeMagic}

                    case maplib.EncounterTypeRuins, maplib.EncounterTypeAbandonedKeep, maplib.EncounterTypeDungeon:
                        books = []data.MagicType{data.DeathMagic}

                    case maplib.EncounterTypeChaosNode:
                        books = []data.MagicType{data.ChaosMagic}

                    case maplib.EncounterTypeNatureNode:
                        books = []data.MagicType{data.NatureMagic}

                    case maplib.EncounterTypeSorceryNode:
                        books = []data.MagicType{data.SorceryMagic}
                }

                if len(books) > 0 {
                    spellBookRemaining -= 1
                    items = append(items, &TreasureSpellbook{Magic: books[rand.N(len(books))]})
                    budget = 0
                }
            case TreasureTypeRetort:
                retort, ok := chooseRetort()
                if ok {
                    retortRemaining -= 1
                    budget = 0
                    items = append(items, &TreasureRetort{Retort: retort})
                }
        }
    }

    return &Treasure{Treasures: items}
}
