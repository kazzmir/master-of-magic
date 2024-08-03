package lbx

import (
    "math/rand"
)

type Spell struct {
    Name string
    AiGroup int
    AiValue int
    SpellType int
    Section int
    Realm int
    Eligibility int
    CastCost int
    ResearchCost int
    Sound int
    Summoned int
    Flag1 int
    Flag2 int
    Flag3 int

    // which book of magic this spell is a part of
    Magic SpellMagic
    Rarity SpellRarity
}

type SpellRarity int

const (
    SpellRarityCommon SpellRarity = iota
    SpellRarityUncommon
    SpellRarityRare
    SpellRarityVeryRare
)

func (rarity SpellRarity) String() string {
    switch rarity {
        case SpellRarityCommon: return "Common"
        case SpellRarityUncommon: return "Uncommon"
        case SpellRarityRare: return "Rare"
        case SpellRarityVeryRare: return "Very Rare"
        default: return "Unknown"
    }
}

type SpellMagic int

const (
    SpellMagicNone SpellMagic = iota
    SpellMagicNature
    SpellMagicChaos
    SpellMagicDeath
    SpellMagicLife
    SpellMagicSorcery
    SpellMagicArcane
)

func (magic SpellMagic) String() string {
    switch magic {
        case SpellMagicNone: return "None"
        case SpellMagicNature: return "Nature"
        case SpellMagicChaos: return "Chaos"
        case SpellMagicDeath: return "Death"
        case SpellMagicLife: return "Life"
        case SpellMagicSorcery: return "Sorcery"
        case SpellMagicArcane: return "Arcane"
        default: return "Unknown"
    }
}

type Spells struct {
    Spells []Spell
}

func (spells *Spells) AddSpell(spell Spell) {
    spells.Spells = append(spells.Spells, spell)
}

func (spells Spells) GetSpellsByMagic(magic SpellMagic) Spells {
    var out []Spell

    for _, spell := range spells.Spells {
        if spell.Magic == magic {
            out = append(out, spell)
        }
    }

    return SpellsFromArray(out)
}

func (spells Spells) GetSpellsByRarity(rarity SpellRarity) Spells {
    var out []Spell

    for _, spell := range spells.Spells {
        if spell.Rarity == rarity {
            out = append(out, spell)
        }
    }

    return SpellsFromArray(out)
}

func (spells *Spells) ShuffleSpells(){
    rand.Shuffle(len(spells.Spells), func(i, j int) {
        spells.Spells[i], spells.Spells[j] = spells.Spells[j], spells.Spells[i]
    })
}

func SpellsFromArray(spells []Spell) Spells {
    return Spells{
        Spells: spells,
    }
}
