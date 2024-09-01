package lbx

import (
    "math/rand"
    "strings"
    "fmt"
    "bytes"
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

func (spells *Spells) RemoveSpell(toRemove Spell){
    var out []Spell
    for _, spell := range spells.Spells {
        if spell.Name != toRemove.Name {
            out = append(out, spell)
        }
    }
    spells.Spells = out
}

func (spells *Spells) FindByName(name string) Spell {
    for _, spell := range spells.Spells {
        if strings.ToLower(spell.Name) == strings.ToLower(name) {
            return spell
        }
    }

    return Spell{}
}

func (spells *Spells) HasSpell(spell Spell) bool {
    return spells.FindByName(spell.Name).Name == spell.Name
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

func ReadSpells(lbx *LbxFile, entry int) (Spells, error) {
    if entry < 0 || entry >= len(lbx.Data) {
        return Spells{}, fmt.Errorf("invalid lbx index %v, must be between 0 and %v", entry, len(lbx.Data) - 1)
    }

    reader := bytes.NewReader(lbx.Data[entry])

    numEntries, err := ReadUint16(reader)
    if err != nil {
        return Spells{}, err
    }

    entrySize, err := ReadUint16(reader)
    if err != nil {
        return Spells{}, err
    }

    var spells Spells

    type MagicData struct {
        Magic SpellMagic
        Rarity SpellRarity
    }

    spellMagicIterator := (func() chan MagicData {
        out := make(chan MagicData)

        go func() {
            defer close(out)

            out <- MagicData{Magic: SpellMagicNone}
            order := []SpellMagic{SpellMagicNature, SpellMagicSorcery, SpellMagicChaos, SpellMagicLife, SpellMagicDeath}
            rarities := []SpellRarity{SpellRarityCommon, SpellRarityUncommon, SpellRarityRare, SpellRarityVeryRare}

            for _, magic := range order {
                // 10 types of common, uncommon, rare, very rare for each book of magic
                for _, rarity := range rarities {
                    for i := 0; i < 10; i++ {
                        out <- MagicData{Magic: magic, Rarity: rarity}
                    }
                }
            }

            // for arcane the spells are
            // common: magic spirit, dispel magic, spell of return, summoning circle
            // uncommon: detect magic, recall hero, disenchant area, enchant item, summon hero
            // rare: awareness, disjunction, create artifact, summon champion
            // very rare: spell of mastery

            for i := 0; i < 4; i++ {
                out <- MagicData{Magic: SpellMagicArcane, Rarity: SpellRarityCommon}
            }

            for i := 0; i < 5; i++ {
                out <- MagicData{Magic: SpellMagicArcane, Rarity: SpellRarityUncommon}
            }

            for i := 0; i < 4; i++ {
                out <- MagicData{Magic: SpellMagicArcane, Rarity: SpellRarityRare}
            }

            out <- MagicData{Magic: SpellMagicArcane, Rarity: SpellRarityVeryRare}
        }()

        return out
    })()

    for i := 0; i < int(numEntries); i++ {
        data := make([]byte, entrySize)
        n, err := reader.Read(data)
        if err != nil {
            return Spells{}, fmt.Errorf("Error reading help index %v: %v", i, err)
        }

        buffer := bytes.NewBuffer(data[0:n])

        nameData := buffer.Next(18)
        // fmt.Printf("Spell %v\n", i)

        name, err := bytes.NewBuffer(nameData).ReadString(0)
        if err != nil {
            name = string(nameData)
        } else {
            name = name[0:len(name)-1]
        }
        // fmt.Printf("  Name: %v\n", string(name))

        aiGroup, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }
        // fmt.Printf("  AI Group: %v\n", aiGroup)

        aiValue, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }
        // fmt.Printf("  AI Value: %v\n", aiValue)

        spellType, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Spell Type: %v\n", spellType)

        section, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Section: %v\n", section)

        realm, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Magic Realm: %v\n", realm)

        eligibility, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Caster Eligibility: %v\n", eligibility)

        buffer.Next(1) // ignore extra unused byte from 2-byte alignment

        castCost, err := readUint16Big(buffer)
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Casting Cost: %v\n", castCost)

        researchCost, err := readUint16Big(buffer)
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Research Cost: %v\n", researchCost)

        sound, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Sound effect: %v\n", sound)

        // skip extra byte due to 2-byte alignment
        buffer.ReadByte()

        summoned, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Summoned: %v\n", summoned)

        flag1, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }

        flag2, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }

        // FIXME: should this be a uint16?
        flag3, err := buffer.ReadByte()
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Flag1=%v Flag2=%v Flag3=%v\n", flag1, flag2, flag3)

        magicData := <-spellMagicIterator

        spells.AddSpell(Spell{
            Name: name,
            AiGroup: int(aiGroup),
            AiValue: int(aiValue),
            SpellType: int(spellType),
            Section: int(section),
            Realm: int(realm),
            Eligibility: int(eligibility),
            CastCost: int(castCost),
            ResearchCost: int(researchCost),
            Sound: int(sound),
            Summoned: int(summoned),
            Flag1: int(flag1),
            Flag2: int(flag2),
            Flag3: int(flag3),

            Magic: magicData.Magic,
            Rarity: magicData.Rarity,
        })
    }

    return spells, nil
}
