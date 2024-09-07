package spellbook

import (
    "math/rand"
    "strings"
    "fmt"
    "bytes"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

type Spell struct {
    Name string
    Index int
    AiGroup int
    AiValue int
    SpellType int
    Section Section
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
    Magic data.MagicType
    Rarity SpellRarity
}

type Section int
const (
    SectionSpecial Section = 0
    SectionSummoning Section = 1
    SectionUnitSpell Section = 2
    SectionCitySpell Section = 3
    SectionEnchantment Section = 4
    SectionCombatSpell Section = 5
)

func (section Section) Name() string {
    switch section {
        case SectionSpecial: return "Special Spells"
        case SectionSummoning: return "Summoning"
        case SectionEnchantment: return "Enchantment"
        case SectionCitySpell: return "City Spells"
        case SectionUnitSpell: return "Unit Spells"
        case SectionCombatSpell: return "Combat Spells"
    }

    return "unknown"
}

func (section Section) String() string {
    return fmt.Sprintf("%v (%d)", section.Name(), int(section))
}

// true if there is a section after this one
func (section Section) HasNext() bool {
    return section.NextSection() != section
}

func (section Section) HasPrevious() bool {
    return section.PreviousSection() != section
}

func (section Section) NextSection() Section {
    switch section {
        case SectionSummoning: return SectionSpecial
        case SectionSpecial: return SectionCitySpell
        case SectionCitySpell: return SectionEnchantment
        case SectionEnchantment: return SectionUnitSpell
        case SectionUnitSpell: return SectionCombatSpell
        case SectionCombatSpell: return SectionCombatSpell
    }

    return section
}

func (section Section) PreviousSection() Section {
    switch section {
        case SectionSummoning: return SectionSummoning
        case SectionSpecial: return SectionSummoning
        case SectionCitySpell: return SectionSpecial
        case SectionEnchantment: return SectionCitySpell
        case SectionUnitSpell: return SectionEnchantment
        case SectionCombatSpell: return SectionUnitSpell
    }

    return section
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

func (spells Spells) GetSpellsBySection(section Section) Spells {
    var out []Spell

    for _, spell := range spells.Spells {
        if spell.Section == section {
            out = append(out, spell)
        }
    }

    return SpellsFromArray(out)
}

func (spells Spells) GetSpellsByMagic(magic data.MagicType) Spells {
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

func ReadSpellsFromCache(cache *lbx.LbxCache) (Spells, error) {
    file, err := cache.GetLbxFile("spelldat.lbx")
    if err != nil {
        return Spells{}, err
    }
    return ReadSpells(file, 0)
}

// pass in spelldat.lbx and 0
func ReadSpells(lbxFile *lbx.LbxFile, entry int) (Spells, error) {
    if entry < 0 || entry >= len(lbxFile.Data) {
        return Spells{}, fmt.Errorf("invalid lbx index %v, must be between 0 and %v", entry, len(lbxFile.Data) - 1)
    }

    reader := bytes.NewReader(lbxFile.Data[entry])

    numEntries, err := lbx.ReadUint16(reader)
    if err != nil {
        return Spells{}, err
    }

    entrySize, err := lbx.ReadUint16(reader)
    if err != nil {
        return Spells{}, err
    }

    var spells Spells

    type MagicData struct {
        Magic data.MagicType
        Rarity SpellRarity
    }

    // FIXME: turn this into a go 1.23 iterator
    spellMagicIterator := (func() chan MagicData {
        out := make(chan MagicData)

        go func() {
            defer close(out)

            out <- MagicData{Magic: data.MagicNone}
            order := []data.MagicType{data.NatureMagic, data.SorceryMagic, data.ChaosMagic, data.LifeMagic, data.DeathMagic}
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
                out <- MagicData{Magic: data.ArcaneMagic, Rarity: SpellRarityCommon}
            }

            for i := 0; i < 5; i++ {
                out <- MagicData{Magic: data.ArcaneMagic, Rarity: SpellRarityUncommon}
            }

            for i := 0; i < 4; i++ {
                out <- MagicData{Magic: data.ArcaneMagic, Rarity: SpellRarityRare}
            }

            out <- MagicData{Magic: data.ArcaneMagic, Rarity: SpellRarityVeryRare}
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

        nameData := buffer.Next(19)
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

        castCost, err := lbx.ReadUint16(buffer)
        if err != nil {
            return Spells{}, err
        }

        // fmt.Printf("  Casting Cost: %v\n", castCost)

        researchCost, err := lbx.ReadUint16(buffer)
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
            Index: i,
            AiGroup: int(aiGroup),
            AiValue: int(aiValue),
            SpellType: int(spellType),
            Section: Section(section),
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

func ReadSpellDescriptionsFromCache(cache *lbx.LbxCache) ([]string, error) {
    file, err := cache.GetLbxFile("desc.lbx")
    if err != nil {
        return nil, err
    }
    return ReadSpellDescriptions(file)
}

// pass in desc.lbx
func ReadSpellDescriptions(file *lbx.LbxFile) ([]string, error) {
    entries, err := file.RawData(0)
    if err != nil {
        return nil, err
    }

    reader := bytes.NewReader(entries)

    count, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    if count > 10000 {
        return nil, fmt.Errorf("Spell count was too high: %v", count)
    }

    size, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    if size > 10000 {
        return nil, fmt.Errorf("Size of each spell entry was too high: %v", size)
    }

    var descriptions []string

    for i := 0; i < int(count); i++ {
        data := make([]byte, size)
        _, err := reader.Read(data)

        if err != nil {
            break
        }

        nullByte := bytes.IndexByte(data, 0)
        if nullByte != -1 {
            descriptions = append(descriptions, string(data[0:nullByte]))
        } else {
            descriptions = append(descriptions, string(data))
        }
    }

    return descriptions, nil
}
