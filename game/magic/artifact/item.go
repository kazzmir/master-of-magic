package artifact

import (
	"bytes"
	"fmt"
	_ "log"

	"github.com/kazzmir/master-of-magic/game/magic/data"
	"github.com/kazzmir/master-of-magic/lib/lbx"
)

// TODO: use Artifact
type Item struct {
	Name string
	Image int
	Slot ArtifactSlot
	Type ArtifactType
	Powers []Power
	Abilities []data.Ability
}

func ReadItems(cache *lbx.LbxCache) ([]Item, error) {
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

	var out []Item

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
		_, err = lbx.ReadUint16(reader) // TODO: manaCost
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
			powers = append(powers, &PowerAttack{Amount: int(attack)})
		}

		toHit, err := lbx.ReadByte(reader)
		if err != nil {
			return nil, fmt.Errorf("read error: %v", err)
		}
		if toHit != 0 {
			powers = append(powers, &PowerToHit{Amount: int(toHit)})
		}

		defense, err := lbx.ReadByte(reader)
		if err != nil {
			return nil, fmt.Errorf("read error: %v", err)
		}
		if defense != 0 {
			powers = append(powers, &PowerDefense{Amount: int(defense)})
		}

		movement, err := lbx.ReadByte(reader)
		if err != nil {
			return nil, fmt.Errorf("read error: %v", err)
		}
		if movement != 0 {
			powers = append(powers, &PowerMovement{Amount: int(movement)})
		}

		resistance, err := lbx.ReadByte(reader)
		if err != nil {
			return nil, fmt.Errorf("read error: %v", err)
		}
		if resistance != 0 {
			powers = append(powers, &PowerResistance{Amount: int(resistance)})
		}

		spellSkill, err := lbx.ReadByte(reader)
		if err != nil {
			return nil, fmt.Errorf("read error: %v", err)
		}
		if spellSkill != 0 {
			powers = append(powers, &PowerSpellSkill{Amount: int(spellSkill)})
		}

		spellSave, err := lbx.ReadByte(reader)
		if err != nil {
			return nil, fmt.Errorf("read error: %v", err)
		}
		if spellSave != 0 {
			powers = append(powers, &PowerSpellSave{Amount: int(spellSave)})
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

		item := Item{
			Name: string(name),
			Image: int(image),
			Slot: slot,
			Type: artifactType,
			Powers: powers,
			Abilities: abilities,
		}
		out = append(out, item)
	}

	return out, nil
}
