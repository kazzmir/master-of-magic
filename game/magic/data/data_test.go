package data

import "testing"

func TestRomanNumeral(test *testing.T) {
    cases := map[int]string{
        1: "I",
        4: "IV",
        9: "IX",
        10: "X",
        11: "11", // falls back to plain digits above the known range
    }

    for value, expected := range cases {
        if got := romanNumeral(value); got != expected {
            test.Errorf("romanNumeral(%v) = %q, want %q", value, got, expected)
        }
    }
}

func TestAllAbilitiesHaveNames(test *testing.T) {
    seen := make(map[string]bool)

    for _, abilityType := range AllAbilities() {
        ability := MakeAbility(abilityType)
        name := ability.Name()

        if name == "" {
            test.Errorf("ability %v has an empty Name()", int(abilityType))
        }

        if seen[name] {
            test.Errorf("ability name %q is used by more than one AbilityType", name)
        }
        seen[name] = true
    }
}

func TestIsHeroAbility(test *testing.T) {
    if !MakeAbility(AbilityLucky).IsHeroAbility() {
        test.Errorf("Lucky should be a hero ability")
    }

    if MakeAbility(AbilityFireBreath).IsHeroAbility() {
        test.Errorf("Fire Breath should not be a hero ability")
    }
}

func TestAbilityValueRoundTrip(test *testing.T) {
    ability := MakeAbilityValue(AbilityCaster, 2.5)

    if ability.Ability != AbilityCaster {
        test.Errorf("expected Ability to be AbilityCaster")
    }

    if ability.Value != 2.5 {
        test.Errorf("expected Value to be 2.5 but was %v", ability.Value)
    }
}

func TestItemAbilityMagicType(test *testing.T) {
    // spot-check a few representative item abilities rather than every one -
    // this is meant as baseline coverage of the mapping logic, not exhaustive
    cases := map[ItemAbility]MagicType{
        ItemAbilityHaste: SorceryMagic,
        ItemAbilityFlaming: ChaosMagic,
        ItemAbilityBless: LifeMagic,
    }

    for item, expected := range cases {
        if got := item.MagicType(); got != expected {
            test.Errorf("%v.MagicType() = %v, want %v", item.Name(), got, expected)
        }
    }
}

func TestItemAbilityAbilityType(test *testing.T) {
    if got := ItemAbilityCloakOfFear.AbilityType(); got != AbilityCauseFear {
        test.Errorf("ItemAbilityCloakOfFear.AbilityType() = %v, want AbilityCauseFear", got)
    }

    // documents the current (buggy) behavior noted in a FIXME above this
    // case in abilities.go: Stoning should carry a value (AbilityValue(
    // AbilityStoningTouch, 1)) but currently returns the bare ability type
    // with no value attached. If this is fixed, update this test rather
    // than deleting it.
    if got := ItemAbilityStoning.AbilityType(); got != AbilityStoningTouch {
        test.Errorf("ItemAbilityStoning.AbilityType() = %v, want AbilityStoningTouch", got)
    }

    // an item ability with no explicit case should fall back to AbilityNone
    if got := ItemAbilityHaste.AbilityType(); got != AbilityNone {
        test.Errorf("ItemAbilityHaste.AbilityType() = %v, want AbilityNone (no explicit mapping)", got)
    }
}

func TestItemAbilityEnchantment(test *testing.T) {
    if got := ItemAbilityBless.Enchantment(); got != UnitEnchantmentBless {
        test.Errorf("ItemAbilityBless.Enchantment() = %v, want UnitEnchantmentBless", got)
    }
}

func TestItemAbilityName(test *testing.T) {
    if ItemAbilityStoning.Name() == "" {
        test.Errorf("ItemAbilityStoning should have a non-empty Name()")
    }

    if ItemAbilityNone.Name() == ItemAbilityStoning.Name() {
        test.Errorf("ItemAbilityNone and ItemAbilityStoning should not share a name")
    }
}
