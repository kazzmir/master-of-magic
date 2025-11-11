package game

import (
    "fmt"
    "slices"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/lib/lbx"
)

type TreasureItem interface {
    fmt.Stringer
}

type TreasureGold struct {
    Amount int
}

func (gold *TreasureGold) String() string {
    return fmt.Sprintf("%v gold coins", gold.Amount)
}

type TreasureMana struct {
    Amount int
}

func (mana *TreasureMana) String() string {
    return fmt.Sprintf("%v mana crystals", mana.Amount)
}

type TreasureMagicalItem struct {
    Artifact *artifact.Artifact
}

func (item *TreasureMagicalItem) String() string {
    return fmt.Sprintf("a %v", item.Artifact.Name)
}

type TreasurePrisonerHero struct {
    Hero *herolib.Hero
}

func (prisoner *TreasurePrisonerHero) String() string {
    return fmt.Sprintf("%v, a prisoner hero", prisoner.Hero.Name)
}

type TreasureSpell struct {
    Spell spellbook.Spell
}

func (spell *TreasureSpell) String() string {
    return fmt.Sprintf("the spell %v", spell.Spell.Name)
}

// always gives one spellbook of the specified magic type
type TreasureSpellbook struct {
    Magic data.MagicType
    Count int
}

func (spellbook *TreasureSpellbook) String() string {
    return fmt.Sprintf("a spellbook of %v magic", spellbook.Magic)
}

type TreasureRetort struct {
    Retort data.Retort
}

func (retort *TreasureRetort) String() string {
    return fmt.Sprintf("the retort %v", retort.Retort)
}

type Treasure struct {
    Treasures []TreasureItem
}

func combineAnd[T fmt.Stringer](pieces []T) string {
    // this could be simplified by iterating backwards through the array and
    // preprending 'and' or ', ' for each element, depending on its index

    if len(pieces) == 0 {
        return ""
    }

    // x
    if len(pieces) == 1 {
        return pieces[0].String()
    }

    // x and y
    if len(pieces) == 2 {
        return fmt.Sprintf("%v and %v", pieces[0], pieces[1])
    }

    // x, y, and z
    out := ""
    for i := range len(pieces) - 2 {
        out += fmt.Sprintf("%v, ", pieces[i])
    }
    out += fmt.Sprintf("%v and %v", pieces[len(pieces) - 2], pieces[len(pieces) - 1])
    return out
}

func (treasure Treasure) String() string {
    if len(treasure.Treasures) == 0 {
        return "Inside you find absolutely nothing."
    }

    return "Inside you find " + combineAnd(treasure.Treasures)
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
func makeTreasure(cache *lbx.LbxCache, encounterType maplib.EncounterType, budget int, wizard setup.WizardCustom, knownSpells spellbook.Spells, allSpells spellbook.Spells, heroes []*herolib.Hero, getPremadeArtifacts func() []*artifact.Artifact) Treasure {
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

    chooseRetort := func () (data.Retort, bool) {
        choices := []data.Retort{
            data.RetortAlchemy,
            data.RetortWarlord,
            data.RetortChanneler,
            data.RetortArchmage,
            data.RetortArtificer,
            data.RetortConjurer,
            data.RetortSageMaster,
            data.RetortDivinePower,
            data.RetortFamous,
            data.RetortRunemaster,
            data.RetortCharismatic,
            data.RetortChaosMastery,
            data.RetortNatureMastery,
            data.RetortSorceryMastery,
            data.RetortInfernalPower,
            data.RetortManaFocusing,
            data.RetortNodeMastery,
        }

        choices = slices.DeleteFunc(choices, func (ability data.Retort) bool {
            return wizard.RetortEnabled(ability)
        })

        if len(choices) > 0 {
            return choices[rand.N(len(choices))], true
        }

        return data.RetortNone, false
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
    // specials are spellbooks and retorts
    specialsRemaining := 2
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
        if specialsRemaining > 0 && budget >= spellBookSpend {
            choices[TreasureTypeSpellbook] = 2
        }
        if specialsRemaining > 0 && budget >= retortSpend {
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

                added := false
                for _, item := range items {
                    if gold, ok := item.(*TreasureGold); ok {
                        coins += gold.Amount
                        added = true
                    }
                }

                if !added {
                    items = append(items, &TreasureGold{Amount: coins})
                }

                budget -= coins
            case TreasureTypeMana:
                mana := rand.N(200) + 1
                mana = min(mana, budget)

                added := false
                for _, item := range items {
                    if manaItem, ok := item.(*TreasureMana); ok {
                        mana += manaItem.Amount
                        added = true
                    }
                }

                if !added {
                    items = append(items, &TreasureMana{Amount: mana})
                }

                budget -= mana
            case TreasureTypeMagicalItem:
                artifacts := getPremadeArtifacts()
                if len(artifacts) > 0 {

                    for _, choice := range rand.Perm(len(artifacts)) {
                        use := artifacts[choice]
                        if budget >= use.Cost && canUseArtifact(use, wizard) {
                            magicItemRemaining -= 1
                            budget -= use.Cost
                            items = append(items, &TreasureMagicalItem{Artifact: use})
                        }
                    }
                } else {
                    randomArtifact := artifact.MakeRandomArtifact(cache)
                    if budget >= randomArtifact.Cost && canUseArtifact(&randomArtifact, wizard) {
                        magicItemRemaining -= 1
                        budget -= randomArtifact.Cost
                        items = append(items, &TreasureMagicalItem{Artifact: &randomArtifact})
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

                // a wizard with life cannot receive death, and vice versa
                if wizard.MagicLevel(data.LifeMagic) > 0 {
                    books = slices.DeleteFunc(books, func(magic data.MagicType) bool {
                        return magic == data.DeathMagic
                    })
                } else if wizard.MagicLevel(data.DeathMagic) > 0 {
                    books = slices.DeleteFunc(books, func(magic data.MagicType) bool {
                        return magic == data.LifeMagic
                    })
                }

                if len(books) > 0 {
                    specialsRemaining -= 1

                    budget -= spellBookSpend
                    count := 1

                    // update the existing spellbook so that there is only one spellbook item, not two
                    added := false
                    for _, item := range items {
                        if spellbook, ok := item.(*TreasureSpellbook); ok {
                            if spellbook.Magic == books[0] {
                                spellbook.Count += 1
                                added = true
                            }
                        }
                    }

                    if !added {
                        items = append(items, &TreasureSpellbook{Magic: books[rand.N(len(books))], Count: count})
                    }
                }
            case TreasureTypeRetort:
                retort, ok := chooseRetort()
                if ok {
                    specialsRemaining -= 1
                    items = append(items, &TreasureRetort{Retort: retort})
                    budget -= retortSpend
                }
        }
    }

    return Treasure{Treasures: items}
}
