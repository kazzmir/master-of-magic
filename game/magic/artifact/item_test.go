package artifact

import (
    "testing"

    "github.com/kazzmir/master-of-magic/game/magic/data"
)

func TestArtifactTypeName(test *testing.T) {
    if ArtifactTypeSword.Name() != "Sword" {
        test.Errorf("expected Sword but got %v", ArtifactTypeSword.Name())
    }

    if ArtifactTypeNone.Name() != "" {
        test.Errorf("expected empty name for ArtifactTypeNone but got %v", ArtifactTypeNone.Name())
    }
}

func TestArtifactSlotCompatibleWith(test *testing.T) {
    cases := []struct {
        slot ArtifactSlot
        kind ArtifactType
        compatible bool
    }{
        {ArtifactSlotMeleeWeapon, ArtifactTypeSword, true},
        {ArtifactSlotMeleeWeapon, ArtifactTypeBow, false},
        {ArtifactSlotRangedWeapon, ArtifactTypeBow, true},
        {ArtifactSlotMagicWeapon, ArtifactTypeStaff, true},
        {ArtifactSlotMagicWeapon, ArtifactTypeSword, false},
        {ArtifactSlotArmor, ArtifactTypePlate, true},
        {ArtifactSlotJewelry, ArtifactTypeMisc, true},
        {ArtifactSlotJewelry, ArtifactTypeShield, false},
    }

    for _, testCase := range cases {
        if got := testCase.slot.CompatibleWith(testCase.kind); got != testCase.compatible {
            test.Errorf("slot %v CompatibleWith(%v) = %v, want %v", testCase.slot, testCase.kind.Name(), got, testCase.compatible)
        }
    }
}

func makeAbilityPower(ability data.ItemAbility) Power {
    return Power{Type: PowerTypeAbility1, Ability: ability}
}

func TestFirstAndLastAbility(test *testing.T) {
    empty := &Artifact{}
    if empty.FirstAbility() != data.ItemAbilityNone {
        test.Errorf("expected ItemAbilityNone on an artifact with no powers")
    }

    item := &Artifact{
        Powers: []Power{
            makeAbilityPower(data.ItemAbilityHaste),
            {Type: PowerTypeAttack, Amount: 5},
            makeAbilityPower(data.ItemAbilityFlaming),
        },
    }

    if item.FirstAbility() != data.ItemAbilityHaste {
        test.Errorf("expected FirstAbility to be Haste but got %v", item.FirstAbility())
    }

    if item.LastAbility() != data.ItemAbilityFlaming {
        test.Errorf("expected LastAbility to be Flaming but got %v", item.LastAbility())
    }

    if !item.HasItemAbility(data.ItemAbilityHaste) {
        test.Errorf("expected item to have the Haste ability")
    }

    if item.HasItemAbility(data.ItemAbilityBless) {
        test.Errorf("expected item not to have the Bless ability")
    }

    if !item.HasAbilities() {
        test.Errorf("expected HasAbilities to be true")
    }

    if empty.HasAbilities() {
        test.Errorf("expected HasAbilities to be false on an artifact with no powers")
    }
}

func TestHasAbilityLargeShieldSpecialCase(test *testing.T) {
    shield := &Artifact{Type: ArtifactTypeShield}
    if !shield.HasAbility(data.AbilityLargeShield) {
        test.Errorf("a shield-type artifact should always have AbilityLargeShield")
    }

    sword := &Artifact{Type: ArtifactTypeSword}
    if sword.HasAbility(data.AbilityLargeShield) {
        test.Errorf("a sword-type artifact should not have AbilityLargeShield")
    }
}

func TestHasAbilityViaItemAbilityMapping(test *testing.T) {
    item := &Artifact{
        Powers: []Power{makeAbilityPower(data.ItemAbilityCloakOfFear)},
    }

    if !item.HasAbility(data.AbilityCauseFear) {
        test.Errorf("expected Cloak Of Fear power to confer AbilityCauseFear")
    }

    if item.HasAbility(data.AbilityLucky) {
        test.Errorf("did not expect Cloak Of Fear power to confer AbilityLucky")
    }
}

func TestHasEnchantmentAndGetEnchantments(test *testing.T) {
    item := &Artifact{
        Powers: []Power{
            makeAbilityPower(data.ItemAbilityBless),
            {Type: PowerTypeAttack, Amount: 3},
        },
    }

    if !item.HasEnchantment(data.UnitEnchantmentBless) {
        test.Errorf("expected item to confer the Bless enchantment")
    }

    enchantments := item.GetEnchantments()
    if len(enchantments) != 1 || enchantments[0] != data.UnitEnchantmentBless {
        test.Errorf("expected exactly one Bless enchantment but got %v", enchantments)
    }
}

func TestAddAndRemovePower(test *testing.T) {
    item := &Artifact{}

    power := Power{Type: PowerTypeAttack, Amount: 5, Index: 1}
    item.AddPower(power)

    if len(item.Powers) != 1 {
        test.Fatalf("expected 1 power after AddPower but got %v", len(item.Powers))
    }

    item.RemovePower(power)

    if len(item.Powers) != 0 {
        test.Errorf("expected 0 powers after RemovePower but got %v", len(item.Powers))
    }
}

func TestDamageBonusesRespectArtifactType(test *testing.T) {
    attackPower := Power{Type: PowerTypeAttack, Amount: 10}

    sword := &Artifact{Type: ArtifactTypeSword, Powers: []Power{attackPower}}
    if sword.MeleeBonus() != 10 {
        test.Errorf("expected sword MeleeBonus to be 10 but was %v", sword.MeleeBonus())
    }
    if sword.RangedAttackBonus() != 0 {
        test.Errorf("expected sword RangedAttackBonus to be 0 but was %v", sword.RangedAttackBonus())
    }

    bow := &Artifact{Type: ArtifactTypeBow, Powers: []Power{attackPower}}
    if bow.RangedAttackBonus() != 10 {
        test.Errorf("expected bow RangedAttackBonus to be 10 but was %v", bow.RangedAttackBonus())
    }
    if bow.MeleeBonus() != 0 {
        test.Errorf("expected bow MeleeBonus to be 0 but was %v", bow.MeleeBonus())
    }

    wand := &Artifact{Type: ArtifactTypeWand, Powers: []Power{attackPower}}
    if wand.MagicAttackBonus() != 10 {
        test.Errorf("expected wand MagicAttackBonus to be 10 but was %v", wand.MagicAttackBonus())
    }

    // misc items can carry any of the three attack bonus types
    misc := &Artifact{Type: ArtifactTypeMisc, Powers: []Power{attackPower}}
    if misc.MeleeBonus() != 10 || misc.RangedAttackBonus() != 10 || misc.MagicAttackBonus() != 10 {
        test.Errorf("expected a misc item to apply its attack power to all three bonus types")
    }
}

func TestDefenseBonusIncludesArmorTypeBase(test *testing.T) {
    defensePower := Power{Type: PowerTypeDefense, Amount: 3}

    chain := &Artifact{Type: ArtifactTypeChain, Powers: []Power{defensePower}}
    if chain.DefenseBonus() != 4 {
        test.Errorf("expected chain DefenseBonus to be 3+1=4 but was %v", chain.DefenseBonus())
    }

    plate := &Artifact{Type: ArtifactTypePlate, Powers: []Power{defensePower}}
    if plate.DefenseBonus() != 5 {
        test.Errorf("expected plate DefenseBonus to be 3+2=5 but was %v", plate.DefenseBonus())
    }

    misc := &Artifact{Type: ArtifactTypeMisc, Powers: []Power{defensePower}}
    if misc.DefenseBonus() != 3 {
        test.Errorf("expected misc DefenseBonus to be just the power amount, 3, but was %v", misc.DefenseBonus())
    }
}

func TestSimplePowerBonuses(test *testing.T) {
    item := &Artifact{
        Powers: []Power{
            {Type: PowerTypeToHit, Amount: 15},
            {Type: PowerTypeSpellSkill, Amount: 5},
            {Type: PowerTypeSpellSave, Amount: 2},
            {Type: PowerTypeResistance, Amount: 4},
            {Type: PowerTypeMovement, Amount: 1},
        },
    }

    if !item.HasToHitPower() || item.ToHitBonus() != 15 {
        test.Errorf("expected ToHitBonus 15 but got %v", item.ToHitBonus())
    }

    if !item.HasSpellSkillPower() || item.SpellSkillBonus() != 5 {
        test.Errorf("expected SpellSkillBonus 5 but got %v", item.SpellSkillBonus())
    }

    if !item.HasSpellSavePower() || item.SpellSaveBonus() != 2 {
        test.Errorf("expected SpellSaveBonus 2 but got %v", item.SpellSaveBonus())
    }

    if !item.HasResistancePower() || item.ResistanceBonus() != 4 {
        test.Errorf("expected ResistanceBonus 4 but got %v", item.ResistanceBonus())
    }

    if !item.HasMovementPower() || item.MovementBonus() != 1 {
        test.Errorf("expected MovementBonus 1 but got %v", item.MovementBonus())
    }

    empty := &Artifact{}
    if empty.HasToHitPower() || empty.HasSpellSkillPower() || empty.HasSpellSavePower() || empty.HasResistancePower() || empty.HasMovementPower() || empty.HasDefensePower() || empty.HasAbilityPower() {
        test.Errorf("an artifact with no powers should report false for every Has*Power check")
    }
}
