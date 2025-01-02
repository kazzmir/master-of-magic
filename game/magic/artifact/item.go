package artifact

import (
    "bytes"
    "fmt"
    "slices"
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
    }

    return "unknown"
}

type Power struct {
    Type PowerType
    Amount int
    Name string
    Spell spellbook.Spell
}

type Artifact struct {
    Type ArtifactType
    Image int
    Name string
	Cost int
    Powers []Power
    Abilities []data.Ability
}

func (artifact *Artifact) HasAbility(ability data.AbilityType) bool {
    switch ability {
        case data.AbilityLargeShield: return artifact.Type == ArtifactTypeShield
    }

    for _, check := range artifact.Abilities {
        if check.Ability == ability {
            return true
        }
    }

    return false
}

func (artifact *Artifact) AddPower(power Power) {
    artifact.Powers = append(artifact.Powers, power)
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

func ReadArtifacts(cache *lbx.LbxCache) ([]Artifact, error) {
    itemData, err := cache.GetLbxFile("itemdata.lbx")
    if err != nil {
        return nil, fmt.Errorf("unable to read itemdata.lbx: %v", err)
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

    abilityMap := map[uint32]data.Ability{
        1 << 0:  data.MakeAbility(data.AbilityVampiric),
        1 << 1:  data.MakeAbility(data.AbilityGuardianWind),
        1 << 2:  data.MakeAbility(data.AbilityLightning),
        1 << 3:  data.MakeAbility(data.AbilityCloakOfFear),
        1 << 4:  data.MakeAbility(data.AbilityDestruction),
        1 << 5:  data.MakeAbility(data.AbilityWraithform),
        1 << 6:  data.MakeAbility(data.AbilityRegeneration),
        1 << 7:  data.MakeAbility(data.AbilityPathfinding),
        1 << 8:  data.MakeAbility(data.AbilityWaterWalking),
        1 << 9:  data.MakeAbility(data.AbilityResistElements),
        1 << 10: data.MakeAbility(data.AbilityElementalArmor),
        1 << 11: data.MakeAbility(data.AbilityChaos),
        1 << 12: data.MakeAbility(data.AbilityStoning),
        1 << 13: data.MakeAbility(data.AbilityEndurance),
        1 << 14: data.MakeAbility(data.AbilityHaste),
        1 << 15: data.MakeAbility(data.AbilityInvisibility),
        1 << 16: data.MakeAbility(data.AbilityDeath),
        1 << 17: data.MakeAbility(data.AbilityFlight),
        1 << 18: data.MakeAbility(data.AbilityResistMagic),
        1 << 19: data.MakeAbility(data.AbilityMagicImmunity),
        1 << 20: data.MakeAbility(data.AbilityFlaming),
        1 << 21: data.MakeAbility(data.AbilityHolyAvenger),
        1 << 22: data.MakeAbility(data.AbilityTrueSight),
        1 << 23: data.MakeAbility(data.AbilityPhantasmal),
        1 << 24: data.MakeAbility(data.AbilityPowerDrain),
        1 << 25: data.MakeAbility(data.AbilityBless),
        1 << 26: data.MakeAbility(data.AbilityLionHeart),
        1 << 27: data.MakeAbility(data.AbilityGiantStrength),
        1 << 28: data.MakeAbility(data.AbilityPlanarTravel),
        1 << 29: data.MakeAbility(data.AbilityMerging),
        1 << 30: data.MakeAbility(data.AbilityRighteousness),
        1 << 31: data.MakeAbility(data.AbilityInvulnerability),
    }

    for i := 0; i < int(numEntries); i++ {
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

        // TODO: use this somehow? seems unnecessary
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
            powers = append(powers, Power{Type: PowerTypeSpellSave, Amount: int(spellSave), Name: fmt.Sprintf("+%v Spell Save", spellSave)})
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
            // TODO: map spells, e.g.
            // 42 Dispel True
            // 46 Confusion
            // 50 Psionic Blast
            // 60 Phantom beast
            // 62 Invisibility
            // 67 Mind Storm
            // 73 Creature Binding
            // 112 Disintegrate
            // 124 Holy Weapon
            // 151 High Prayer
            // powers = append(powers, &PowerSpellCharges{Spell: spell, Charges: int(charges)})
        }

        // Abilities
        abilitiesValue, err := lbx.ReadUint32(reader)
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }

        var abilities []data.Ability
        for mask, ability := range abilityMap {
            if abilitiesValue&mask != 0 {
                abilities = append(abilities, ability)
            }
        }

        // Requirements
        _, err = lbx.ReadByte(reader) // TODO: natureRanksNeeded
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }

        _, err = lbx.ReadByte(reader) // TODO: sorceryRanksNeeded
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }

        _, err = lbx.ReadByte(reader) // TODO: chaosRanksNeeded
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }

        _, err = lbx.ReadByte(reader) // TODO: lifeRanksNeeded
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
        }

        _, err = lbx.ReadByte(reader) // TODO: deathRanksNeeded
        if err != nil {
            return nil, fmt.Errorf("read error: %v", err)
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
            // Slot: slot,
            Type: artifactType,
            Powers: powers,
            Abilities: abilities,
        }
        out = append(out, artifact)
    }

    return out, nil
}
