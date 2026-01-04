package artifact

import (
    "bytes"
    "fmt"
    "slices"
    "cmp"
    "log"
    "math/rand/v2"
    _ "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
)

type ArtifactType int
const (
    ArtifactTypeNone ArtifactType = iota
    ArtifactTypeSword
    ArtifactTypeMace
    ArtifactTypeAxe
    ArtifactTypeBow
    ArtifactTypeStaff
    ArtifactTypeWand
    ArtifactTypeMisc
    ArtifactTypeShield
    ArtifactTypeChain
    ArtifactTypePlate
)

func (a ArtifactType) String() string {
    return a.Name()
}

func (a ArtifactType) Name() string {
    switch a {
        case ArtifactTypeSword: return "Sword"
        case ArtifactTypeMace: return "Mace"
        case ArtifactTypeAxe: return "Axe"
        case ArtifactTypeBow: return "Bow"
        case ArtifactTypeStaff: return "Staff"
        case ArtifactTypeWand: return "Wand"
        case ArtifactTypeMisc: return "Misc"
        case ArtifactTypeShield: return "Shield"
        case ArtifactTypeChain: return "Chain"
        case ArtifactTypePlate: return "Plate"
    }

    return ""
}

type ArtifactSlot int
const (
    ArtifactSlotMeleeWeapon ArtifactSlot = iota
    ArtifactSlotRangedWeapon
    ArtifactSlotMagicWeapon
    ArtifactSlotAnyWeapon
    ArtifactSlotArmor
    ArtifactSlotJewelry
)

// the index in itemisc.lbx for this slot
func (slot ArtifactSlot) ImageIndex() int {
    switch slot {
        case ArtifactSlotMeleeWeapon: return 19
        case ArtifactSlotRangedWeapon: return 20
        case ArtifactSlotMagicWeapon: return 22
        case ArtifactSlotAnyWeapon: return 21
        case ArtifactSlotArmor: return 24
        case ArtifactSlotJewelry: return 23
    }

    return -1
}

func (slot ArtifactSlot) CompatibleWith(kind ArtifactType) bool {
    switch slot {
        case ArtifactSlotMeleeWeapon:
            return kind == ArtifactTypeSword || kind == ArtifactTypeMace || kind == ArtifactTypeAxe
        case ArtifactSlotRangedWeapon:
            return kind == ArtifactTypeSword || kind == ArtifactTypeMace || kind == ArtifactTypeAxe || kind == ArtifactTypeBow
        case ArtifactSlotMagicWeapon:
            return kind == ArtifactTypeStaff || kind == ArtifactTypeWand
        case ArtifactSlotAnyWeapon:
            return kind == ArtifactTypeStaff || kind == ArtifactTypeWand || kind == ArtifactTypeSword || kind == ArtifactTypeMace || kind == ArtifactTypeAxe
        case ArtifactSlotArmor:
            return kind == ArtifactTypeShield || kind == ArtifactTypeChain || kind == ArtifactTypePlate
        case ArtifactSlotJewelry:
            return kind == ArtifactTypeMisc
    }

    return false
}

type PowerType int
const (
    PowerTypeNone PowerType = iota
    PowerTypeAttack
    PowerTypeDefense
    PowerTypeToHit
    PowerTypeSpellSkill
    PowerTypeSpellSave
    PowerTypeMovement
    PowerTypeResistance
    PowerTypeSpellCharges

    PowerTypeAbility1
    PowerTypeAbility2
    PowerTypeAbility3
)

func (section PowerType) String() string {
    switch section {
        case PowerTypeAttack: return "Attack"
        case PowerTypeDefense: return "Defense"
        case PowerTypeToHit: return "To Hit"
        case PowerTypeSpellSkill: return "Spell Skill"
        case PowerTypeSpellSave: return "Spell Save"
        case PowerTypeMovement: return "Movement"
        case PowerTypeResistance: return "Resistance"
        case PowerTypeSpellCharges: return "Spell Charges"

        case PowerTypeAbility1: return "Ability 1"
        case PowerTypeAbility2: return "Ability 2"
        case PowerTypeAbility3: return "Ability 3"
    }

    return "unknown"
}

type Power struct {
    Type PowerType
    Amount int // for an ability this is the number of books of the Magic needed
    Name string
    Ability data.ItemAbility
    Magic data.MagicType // for abilities

    Spell spellbook.Spell
    SpellCharges int

    // powers are sorted by how they are defined in itempow.lbx, so we just use that number here
    // this field has no utility other than sorting
    Index int
}

type Requirement struct {
    MagicType data.MagicType
    Amount int
}

type Artifact struct {
    Type ArtifactType
    Image int
    Name string
    Cost int
    Powers []Power
    Requirements []Requirement
}

// if true then generally this artifact should be rendered with a glow around it
func (artifact *Artifact) HasAbilities() bool {
    return artifact.HasAbilityPower()
}

func (artifact *Artifact) FirstAbility() data.ItemAbility {
    for _, power := range artifact.Powers {
        if power.Type == PowerTypeAbility1 || power.Type == PowerTypeAbility2 || power.Type == PowerTypeAbility3 {
            return power.Ability
        }
    }

    return data.ItemAbilityNone
}

func (artifact *Artifact) LastAbility() data.ItemAbility {
    for i := len(artifact.Powers) - 1; i >= 0; i-- {
        power := artifact.Powers[i]
        if power.Type == PowerTypeAbility1 || power.Type == PowerTypeAbility2 || power.Type == PowerTypeAbility3 {
            return power.Ability
        }
    }

    return data.ItemAbilityNone
}

func (artifact *Artifact) HasItemAbility(ability data.ItemAbility) bool {
    return slices.ContainsFunc(artifact.Powers, func (power Power) bool {
        if power.Type == PowerTypeAbility1 || power.Type == PowerTypeAbility2 || power.Type == PowerTypeAbility3 {
            return power.Ability == ability
        }

        return false
    })
}

func (artifact *Artifact) HasEnchantment(enchantment data.UnitEnchantment) bool {
    for _, check := range artifact.Powers {
        isAbility := check.Type == PowerTypeAbility1 || check.Type == PowerTypeAbility2 || check.Type == PowerTypeAbility3
        if isAbility && check.Ability.Enchantment() == enchantment {
            return true
        }
    }

    return false

}

func (artifact *Artifact) HasAbility(ability data.AbilityType) bool {
    switch ability {
        case data.AbilityLargeShield: return artifact.Type == ArtifactTypeShield
    }

    for _, check := range artifact.Powers {
        isAbility := check.Type == PowerTypeAbility1 || check.Type == PowerTypeAbility2 || check.Type == PowerTypeAbility3
        if isAbility {
            if check.Ability.AbilityType() == ability {
                return true
            }

            // an item power might confer an enchantment that confers an ability
            for _, enchantmentAbility := range check.Ability.Enchantment().Abilities() {
                if enchantmentAbility.Ability == ability {
                    return true
                }
            }
        }
    }

    return false
}

func (artifact *Artifact) GetEnchantments() []data.UnitEnchantment {
    var out []data.UnitEnchantment
    for _, check := range artifact.Powers {
        isAbility := check.Type == PowerTypeAbility1 || check.Type == PowerTypeAbility2 || check.Type == PowerTypeAbility3
        if isAbility {
            enchantment := check.Ability.Enchantment()
            if enchantment != data.UnitEnchantmentNone {
                out = append(out, enchantment)
            }
        }
    }

    return out
}

func (artifact *Artifact) AddPower(power Power) {
    artifact.Powers = append(artifact.Powers, power)
    slices.SortFunc(artifact.Powers, func (a, b Power) int {
        return cmp.Compare(a.Index, b.Index)
    })
}

func (artifact *Artifact) RemovePower(remove Power) {
    artifact.Powers = slices.DeleteFunc(artifact.Powers, func (power Power) bool {
        return remove == power
    })
}

func hasPower(powerType PowerType, powers []Power) bool {
    for _, power := range powers {
        if power.Type == powerType {
            return true
        }
    }

    return false
}

func addPowers(powerType PowerType, powers []Power) int {
    amount := 0
    for _, power := range powers {
        if power.Type == powerType {
            amount += power.Amount
        }
    }

    return amount
}

func (artifact *Artifact) MeleeBonus() int {
    switch artifact.Type {
        case ArtifactTypeSword, ArtifactTypeMace, ArtifactTypeAxe, ArtifactTypeMisc:
            return addPowers(PowerTypeAttack, artifact.Powers)
        default:
            return 0
    }
}

func (artifact *Artifact) RangedAttackBonus() int {
    switch artifact.Type {
        case ArtifactTypeBow, ArtifactTypeMisc:
            return addPowers(PowerTypeAttack, artifact.Powers)
        default:
            return 0
    }
}

func (artifact *Artifact) MagicAttackBonus() int {
    switch artifact.Type {
        case ArtifactTypeWand, ArtifactTypeStaff, ArtifactTypeMisc:
            return addPowers(PowerTypeAttack, artifact.Powers)
        default:
            return 0
    }
}

func (artifact *Artifact) DefenseBonus() int {
    base := addPowers(PowerTypeDefense, artifact.Powers)
    switch artifact.Type {
        case ArtifactTypeChain:
            base += 1
        case ArtifactTypePlate:
            base += 2
    }

    return base
}

// returns the spell and how many charges it has
func (artifact *Artifact) GetSpellCharge() (spellbook.Spell, int) {
    for _, power := range artifact.Powers {
        if power.Type == PowerTypeSpellCharges {
            return power.Spell, power.Amount
        }
    }

    return spellbook.Spell{}, 0
}

func (artifact *Artifact) HasSpellCharges() bool {
    return hasPower(PowerTypeSpellCharges, artifact.Powers)
}

func (artifact *Artifact) HasDefensePower() bool {
    return hasPower(PowerTypeDefense, artifact.Powers)
}

func (artifact *Artifact) HasSpellSavePower() bool {
    return hasPower(PowerTypeSpellSave, artifact.Powers)
}

func (artifact *Artifact) HasSpellSkillPower() bool {
    return hasPower(PowerTypeSpellSkill, artifact.Powers)
}

func (artifact *Artifact) HasResistancePower() bool {
    return hasPower(PowerTypeResistance, artifact.Powers)
}

func (artifact *Artifact) HasMovementPower() bool {
    return hasPower(PowerTypeMovement, artifact.Powers)
}

func (artifact *Artifact) HasToHitPower() bool {
    return hasPower(PowerTypeToHit, artifact.Powers)
}

func (artifact *Artifact) HasAbilityPower() bool {
    return hasPower(PowerTypeAbility1, artifact.Powers) || hasPower(PowerTypeAbility2, artifact.Powers) || hasPower(PowerTypeAbility3, artifact.Powers)
}

func (artifact *Artifact) ToHitBonus() int {
    return addPowers(PowerTypeToHit, artifact.Powers)
}

func (artifact *Artifact) SpellSkillBonus() int {
    return addPowers(PowerTypeSpellSkill, artifact.Powers)
}

func (artifact *Artifact) SpellSaveBonus() int {
    return addPowers(PowerTypeSpellSave, artifact.Powers)
}

func (artifact *Artifact) ResistanceBonus() int {
    return addPowers(PowerTypeResistance, artifact.Powers)
}

func (artifact *Artifact) MovementBonus() int {
    return addPowers(PowerTypeMovement, artifact.Powers)
}

// generate an artifact with random properties
func MakeRandomArtifact(cache *lbx.LbxCache) Artifact {

    // random number between min and max inclusive
    randRange := func(min, max int) int {
        return rand.N(max - min + 1) + min
    }

    chooseImage := func(kind ArtifactType) int {
        switch kind {
            case ArtifactTypeSword: return randRange(0, 8)
            case ArtifactTypeMace: return randRange(9, 19)
            case ArtifactTypeAxe: return randRange(20, 28)
            case ArtifactTypeBow: return randRange(29, 37)
            case ArtifactTypeStaff: return randRange(38, 46)
            case ArtifactTypeWand: return randRange(107, 115)
            case ArtifactTypeMisc: return randRange(72, 106)
            case ArtifactTypeShield: return randRange(62, 71)
            case ArtifactTypeChain: return randRange(47, 54)
            case ArtifactTypePlate: return randRange(55, 61)
            default: return 0
        }
    }

    types := []ArtifactType{ArtifactTypeSword, ArtifactTypeMace, ArtifactTypeAxe, ArtifactTypeBow,
                            ArtifactTypeStaff, ArtifactTypeWand, ArtifactTypeMisc, ArtifactTypeShield,
                            ArtifactTypeChain, ArtifactTypePlate }

    artifact := Artifact{
        Type: types[rand.N(len(types))],
    }

    _, costs, compatibilities, err := ReadPowers(cache)
    if err != nil {
        return Artifact{}
    }

    var powers []Power

    for power, types := range compatibilities {
        // ignore spell charges for now
        if power.Type == PowerTypeSpellCharges {
            continue
        }
        if types.Contains(artifact.Type) {
            powers = append(powers, power)
        }
    }

    // it would be very bad if there are no powers
    if len(powers) > 0 {
        numPowers := min(rand.N(4) + 1, len(powers))

        for _, index := range rand.Perm(len(powers))[:numPowers] {
            artifact.Powers = append(artifact.Powers, powers[index])
        }
    }

    artifact.Image = chooseImage(artifact.Type)

    artifact.Cost = calculateCost(&artifact, costs)
    artifact.Name = getName(&artifact, "")

    return artifact
}

func GetItemConversionMaps() (map[byte]ArtifactSlot, map[byte]ArtifactType, map[uint32]data.ItemAbility) {
    slotMap := map[byte]ArtifactSlot{
        1: ArtifactSlotMeleeWeapon,
        2: ArtifactSlotRangedWeapon,
        4: ArtifactSlotMagicWeapon,
        6: ArtifactSlotJewelry,
        5: ArtifactSlotArmor,
    }

    typeMap := map[byte]ArtifactType{
        0: ArtifactTypeSword,
        1: ArtifactTypeMace,
        2: ArtifactTypeAxe,
        3: ArtifactTypeBow,
        4: ArtifactTypeStaff,
        5: ArtifactTypeWand,
        6: ArtifactTypeMisc,
        7: ArtifactTypeShield,
        8: ArtifactTypeChain,
        9: ArtifactTypePlate,
    }

    abilityMap := map[uint32]data.ItemAbility {
        1 << 0:  data.ItemAbilityVampiric,
        1 << 1:  data.ItemAbilityGuardianWind,
        1 << 2:  data.ItemAbilityLightning,
        1 << 3:  data.ItemAbilityCloakOfFear,
        1 << 4:  data.ItemAbilityDestruction,
        1 << 5:  data.ItemAbilityWraithform,
        1 << 6:  data.ItemAbilityRegeneration,
        1 << 7:  data.ItemAbilityPathfinding,
        1 << 8:  data.ItemAbilityWaterWalking,
        1 << 9:  data.ItemAbilityResistElements,
        1 << 10: data.ItemAbilityElementalArmor,
        1 << 11: data.ItemAbilityChaos,
        1 << 12: data.ItemAbilityStoning,
        1 << 13: data.ItemAbilityEndurance,
        1 << 14: data.ItemAbilityHaste,
        1 << 15: data.ItemAbilityInvisibility,
        1 << 16: data.ItemAbilityDeath,
        1 << 17: data.ItemAbilityFlight,
        1 << 18: data.ItemAbilityResistMagic,
        1 << 19: data.ItemAbilityMagicImmunity,
        1 << 20: data.ItemAbilityFlaming,
        1 << 21: data.ItemAbilityHolyAvenger,
        1 << 22: data.ItemAbilityTrueSight,
        1 << 23: data.ItemAbilityPhantasmal,
        1 << 24: data.ItemAbilityPowerDrain,
        1 << 25: data.ItemAbilityBless,
        1 << 26: data.ItemAbilityLionHeart,
        1 << 27: data.ItemAbilityGiantStrength,
        1 << 28: data.ItemAbilityPlanarTravel,
        1 << 29: data.ItemAbilityMerging,
        1 << 30: data.ItemAbilityRighteousness,
        1 << 31: data.ItemAbilityInvulnerability,
    }

    return slotMap, typeMap, abilityMap
}

func ReadArtifacts(cache *lbx.LbxCache) ([]Artifact, error) {
    itemData, err := cache.GetLbxFile("itemdata.lbx")
    if err != nil {
        return nil, fmt.Errorf("unable to read itemdata.lbx: %v", err)
    }

    spells, err := spellbook.ReadSpellsFromCache(cache)
    if err != nil {
        // return nil, fmt.Errorf("unable to read spells: %v", err)
        log.Printf("Warning: could not load spells: %v", err)
    }

    reader, err := itemData.GetReader(0)
    if err != nil {
        return nil, fmt.Errorf("unable to read entry 0 in itemdata.lbx: %v", err)
    }

    numEntries, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil, fmt.Errorf("read error: %v", err)
    }

    entrySize, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil, fmt.Errorf("read error: %v", err)
    }
    if entrySize != 56 {
        return nil, fmt.Errorf("unsupported itemdata.lbx")
    }

    var out []Artifact

    slotMap, typeMap, abilityMap := GetItemConversionMaps()

    for i := range numEntries {
        // Name
        name := make([]byte, 30)
        n, err := reader.Read(name)
        if err != nil || n != int(30) {
            return nil, fmt.Errorf("unable to read item name %v: %v", i, err)
        }
        name = bytes.Trim(name, "\x00")

        // Image index to items.lbx
        image, err := lbx.ReadUint16(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }

        // Slot
        slotValue, err := lbx.ReadByte(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }

        slot, exists := slotMap[slotValue]
        if !exists {
            return nil, fmt.Errorf("Invalid slot type %v", slotValue)
        }
        _ = slot

        // Type
        typeValue, err := lbx.ReadByte(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }

        artifactType, exists := typeMap[typeValue]
        if !exists {
            return nil, fmt.Errorf("Invalid artifact type %v", typeValue)
        }

        // Cost
        cost, err := lbx.ReadUint16(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }

        // Modifiers
        var powers []Power
        attack, err := lbx.ReadByte(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }
        if attack != 0 {
            powers = append(powers, Power{Type: PowerTypeAttack, Amount: int(attack), Name: fmt.Sprintf("+%v Attack", attack)})
        }

        toHit, err := lbx.ReadByte(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }
        if toHit != 0 {
            powers = append(powers, Power{Type: PowerTypeToHit, Amount: int(toHit), Name: fmt.Sprintf("+%v To Hit", toHit)})
        }

        defense, err := lbx.ReadByte(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }
        if defense != 0 {
            powers = append(powers, Power{Type: PowerTypeDefense, Amount: int(defense), Name: fmt.Sprintf("+%v Defense", defense)})
        }

        movement, err := lbx.ReadByte(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }
        if movement != 0 {
            powers = append(powers, Power{Type: PowerTypeMovement, Amount: int(movement), Name: fmt.Sprintf("+%v Movement", movement)})
        }

        resistance, err := lbx.ReadByte(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }
        if resistance != 0 {
            powers = append(powers, Power{Type: PowerTypeResistance, Amount: int(resistance), Name: fmt.Sprintf("+%v Resistance", resistance)})
        }

        spellSkill, err := lbx.ReadByte(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }
        if spellSkill != 0 {
            powers = append(powers, Power{Type: PowerTypeSpellSkill, Amount: int(spellSkill), Name: fmt.Sprintf("+%v Spell Skill", spellSkill)})
        }

        spellSave, err := lbx.ReadByte(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }
        if spellSave != 0 {
            powers = append(powers, Power{Type: PowerTypeSpellSave, Amount: int(spellSave), Name: fmt.Sprintf("-%v Spell Save", spellSave)})
        }

        // Spells
        spell, err := lbx.ReadByte(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }
        charges, err := lbx.ReadUint16(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }
        if spell != 0 && charges != 0 {
            useSpell := spells.FindById(int(spell))
            powers = append(powers, Power{
                Type: PowerTypeSpellCharges,
                Amount: int(charges),
                Spell: useSpell,
                SpellCharges: int(charges),
                Name: fmt.Sprintf("%v Charges of %v", charges, useSpell.Name),
            })
        }

        // Abilities
        abilitiesValue, err := lbx.ReadUint32(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }

        for mask, ability := range abilityMap {
            if abilitiesValue&mask != 0 {
                powers = append(powers, Power{Type: PowerTypeAbility1, Amount: 0, Name: ability.Name(), Ability: ability})
            }
        }

        // Requirements
        var requirements []Requirement
        natureRanksNeeded, err := lbx.ReadByte(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }
        if natureRanksNeeded != 0 {
            requirements = append(requirements, Requirement{MagicType: data.NatureMagic, Amount: int(natureRanksNeeded)})
        }

        sorceryRanksNeeded, err := lbx.ReadByte(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }
        if sorceryRanksNeeded != 0 {
            requirements = append(requirements, Requirement{MagicType: data.SorceryMagic, Amount: int(sorceryRanksNeeded)})
        }

        chaosRanksNeeded, err := lbx.ReadByte(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }
        if chaosRanksNeeded != 0 {
            requirements = append(requirements, Requirement{MagicType: data.ChaosMagic, Amount: int(chaosRanksNeeded)})
        }

        lifeRanksNeeded, err := lbx.ReadByte(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }
        if lifeRanksNeeded != 0 {
            requirements = append(requirements, Requirement{MagicType: data.LifeMagic, Amount: int(lifeRanksNeeded)})
        }

        deathRanksNeeded, err := lbx.ReadByte(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }
        if deathRanksNeeded != 0 {
            requirements = append(requirements, Requirement{MagicType: data.DeathMagic, Amount: int(deathRanksNeeded)})
        }

        // The last byte seems to be some sort of flag
        _, err = lbx.ReadByte(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }

        artifact := Artifact{
            Name: string(name),
            Image: int(image),
            Cost: int(cost),
            Type: artifactType,
            Powers: powers,
            Requirements: requirements,
        }
        out = append(out, artifact)
    }

    return out, nil
}
