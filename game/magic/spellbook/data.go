package spellbook

import (
    "slices"
    "math/rand/v2"
    "strings"
    "fmt"
    "bytes"
    // "sort"

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
    Realm int // FIXME: Realm is equal to LBX magic realm constant, but is unused. This var may be removed, as the Magic var below is the only one being used
    Eligibility EligibilityType
    CastCost int
    OverrideCost int
    ResearchCost int
    Sound int
    Summoned int
    Flag1 int
    Flag2 int
    Flag3 int

    // which book of magic this spell is a part of
    // FIXME: MagicType as a value is not equal to LBX magic realm constant. This var may be renamed to Realm, and the Realm var above can be removed
    Magic data.MagicType
    Rarity SpellRarity
}

func (spell Spell) IsVariableCost() bool {
    return spell.SpellType >= 18
}

func (spell Spell) Invalid() bool {
    return !spell.Valid()
}

func (spell Spell) Valid() bool {
    return spell.Name != "" && spell.Name != "None"
}

func (spell Spell) IsOfRealm(realm data.MagicType) bool {
    return spell.Magic == realm
}

func (spell Spell) Cost(overland bool) int {
    if spell.OverrideCost != 0 {
        return spell.OverrideCost
    }

    return spell.BaseCost(overland)
}

func (spell Spell) IsSummoning() bool {
    return spell.Section == SectionSummoning
}

// the unit enchantment this spell would apply to a unit if any, or UnitEnchantmentNone if none
func (spell Spell) GetUnitEnchantment() data.UnitEnchantment {
    switch spell.Name {
        case "Giant Strength": return data.UnitEnchantmentGiantStrength
        case "Lionheart": return data.UnitEnchantmentLionHeart
        case "Haste": return data.UnitEnchantmentHaste
        case "Immolation": return data.UnitEnchantmentImmolation
        case "Resist Elements": return data.UnitEnchantmentResistElements
        case "Resist Magic": return data.UnitEnchantmentResistMagic
        case "Elemental Armor": return data.UnitEnchantmentElementalArmor
        case "Bless": return data.UnitEnchantmentBless
        /*
    UnitEnchantmentRighteousness
    UnitEnchantmentCloakOfFear
    UnitEnchantmentTrueSight
    UnitEnchantmentPathFinding
    UnitEnchantmentFlight
    UnitEnchantmentChaosChannelsDemonWings
    UnitEnchantmentChaosChannelsDemonSkin
    UnitEnchantmentChaosChannelsFireBreath
    UnitEnchantmentEndurance
    UnitEnchantmentHeroism
    UnitEnchantmentHolyArmor
    UnitEnchantmentHolyWeapon
    UnitEnchantmentInvulnerability
    UnitEnchantmentPlanarTravel
    UnitEnchantmentIronSkin
    UnitEnchantmentRegeneration
    UnitEnchantmentStoneSkin
    UnitEnchantmentWaterWalking
    UnitEnchantmentGuardianWind
    UnitEnchantmentInvisibility
    UnitEnchantmentMagicImmunity
    UnitEnchantmentSpellLock
    UnitEnchantmentWindWalking
    UnitEnchantmentEldritchWeapon
    UnitEnchantmentFlameBlade
    UnitEnchantmentBerserk
    UnitEnchantmentBlackChannels
    UnitEnchantmentWraithForm
    */

    }

    return data.UnitEnchantmentNone
}

// the curse that this spell would apply to a unit
func (spell Spell) GetUnitCurse() data.UnitEnchantment {
    switch spell.Name {
        /*
UnitCurseConfusion
    UnitCurseCreatureBinding
    UnitCurseMindStorm
    UnitCurseVertigo
    UnitCurseShatter
    UnitCurseWarpCreatureMelee
    UnitCurseWarpCreatureDefense
    UnitCurseWarpCreatureResistance
    UnitCurseBlackSleep
    UnitCursePossession
    UnitCurseWeakness
    UnitCurseWeb
    */

    }

    return data.UnitEnchantmentNone
}

// overland=true if casting in overland, otherwise casting in combat
// this does not include any additional costs for the spell
func (spell Spell) BaseCost(overland bool) int {

    if overland {
        switch spell.Eligibility {
            case EligibilityBoth, EligibilityBoth2: return spell.CastCost * 5
            case EligibilityOverlandOnly: return spell.CastCost
            case EligibilityBothFriendlyCity: return spell.CastCost * 5
            case EligibilityBothSameCost: return spell.CastCost
            case EligibilityOverlandWhileBanished: return spell.CastCost
        }
    }

    return spell.CastCost
}

func (spell Spell) SpentAdditionalCost(overland bool) int {
    if spell.OverrideCost != 0 {
        return spell.OverrideCost - spell.BaseCost(overland)
    }

    return 0
}

type EligibilityType int
const (
    EligibilityCombatOnly EligibilityType = 0xff
    EligibilityCombatOnly2 EligibilityType = 0xfe
    EligibilityBoth EligibilityType = 0x0
    EligibilityOverlandOnly EligibilityType = 0x1
    EligibilityBoth2 EligibilityType = 0x2
    EligibilityBothFriendlyCity EligibilityType = 0x3
    EligibilityBothSameCost = 0x4
    EligibilityOverlandWhileBanished = 0x5
)

func (eligibility EligibilityType) CanCastInCombat(defendingCity bool) bool {
    switch eligibility {
        case EligibilityCombatOnly, EligibilityCombatOnly2, EligibilityBoth, EligibilityBoth2, EligibilityBothSameCost:
            return true
        case EligibilityBothFriendlyCity:
            return defendingCity
        default:
            return false
    }
}

func (eligibility EligibilityType) CanCastInOverland() bool {
    switch eligibility {
        case EligibilityBoth, EligibilityOverlandOnly, EligibilityBoth2,
             EligibilityBothFriendlyCity, EligibilityBothSameCost,
             EligibilityOverlandWhileBanished:
            return true
        default:
            return false
    }
}

type Section int
const (
    SectionSpecial Section = 0
    SectionSummoning Section = 1
    SectionEnchantment Section = 2
    SectionCitySpell Section = 3
    SectionUnitSpell Section = 4
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

func (spells *Spells) Copy() Spells {
    out := make([]Spell, len(spells.Spells))
    copy(out, spells.Spells)
    return SpellsFromArray(out)
}

func (spells *Spells) Sub(min int, max int) Spells {
    if min < 0 {
        min = 0
    }

    if max > len(spells.Spells) {
        max = len(spells.Spells)
    }

    if min >= len(spells.Spells) {
        return Spells{}
    }

    return SpellsFromArray(spells.Spells[min:max])
}

func (spells *Spells) AddAllSpells(more Spells) {
    for _, spell := range more.Spells {
        spells.AddSpell(spell)
    }
}

func (spells *Spells) RemoveSpells(toRemove Spells){
    for _, spell := range toRemove.Spells {
        spells.RemoveSpell(spell)
    }
}

func (spells *Spells) RemoveSpellsByMagic(magic data.MagicType){
    spells.RemoveSpells(spells.GetSpellsByMagic(magic))
}

// returns true if the spell was added, false if it was not
func (spells *Spells) AddSpell(spell Spell) bool {
    if spell.Invalid(){
        return false
    }

    for _, check := range spells.Spells {
        if check.Name == spell.Name {
            return false
        }
    }
    spells.Spells = append(spells.Spells, spell)
    return true
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

func (spells *Spells) Contains(spell Spell) bool {
    return spells.FindByName(spell.Name).Name == spell.Name
}

func (spells *Spells) FindById(id int) Spell {
    if id >= 0 && id < len(spells.Spells) {
        candidate := spells.Spells[id]
        if candidate.Index == id {
            return candidate
        }
    }

    for _, spell := range spells.Spells {
        if spell.Index == id {
            return spell
        }
    }

    return Spell{}
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

    /*
    sort.Slice(out, func(i, j int) bool {
        if out[i].Magic == out[j].Magic {
            return out[i].CastCost < out[j].CastCost
        }
        return out[i].Magic < out[j].Magic
    })
    */

    return SpellsFromArray(out)
}

/* the subset of spells that can be cast on the overworld */
func (spells Spells) OverlandSpells() Spells {
    var out []Spell

    for _, spell := range spells.Spells {
        if spell.Name != "None" && spell.Eligibility.CanCastInOverland() {
            out = append(out, spell)
        }
    }

    return SpellsFromArray(out)
}

/* the subset of spells that can be cast in combat
 * pass defendingCity=true if the spell is being cast in a city that the caster is defending
 */
func (spells Spells) CombatSpells(defendingCity bool) Spells {
    var out []Spell

    for _, spell := range spells.Spells {
        if spell.Name != "None" && spell.Eligibility.CanCastInCombat(defendingCity) {
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

func (spells *Spells) SortByRarity(){
    slices.SortFunc(spells.Spells, func(a Spell, b Spell) int {
        if a.Rarity < b.Rarity {
            return -1
        }

        if a.Rarity == b.Rarity {
            return 0
        }

        return 1
    })
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
            Eligibility: EligibilityType(eligibility),
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
