package artifact

import (
    "testing"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/lib/set"
)

func sampleSword(name string, attack int) Artifact {
    return Artifact{
        Type: ArtifactTypeSword,
        Name: name,
        Image: 0,
        Cost: 150,
        Powers: []Power{
            {Type: PowerTypeAttack, Amount: attack, Name: "+1 Attack"},
        },
    }
}

func TestMakeCatalogIndexesEverySlot(test *testing.T) {
    catalog := MakeCatalog([]Artifact{
        sampleSword("First", 1),
        sampleSword("Second", 2),
    })

    if catalog.Len() != 2 {
        test.Fatalf("expected 2 slots, got %v", catalog.Len())
    }

    if catalog.Slots[0].CatalogIndex != 0 || catalog.Slots[0].Name != "First" {
        test.Errorf("slot 0 = %+v", catalog.Slots[0])
    }

    if catalog.Slots[1].CatalogIndex != 1 || catalog.Slots[1].Name != "Second" {
        test.Errorf("slot 1 = %+v", catalog.Slots[1])
    }

    if !catalog.IsAvailable(0) || !catalog.IsAvailable(1) {
        test.Errorf("fresh catalog slots should all be available")
    }
}

func TestAwardLeavesSlotUnavailable(test *testing.T) {
    catalog := MakeCatalog([]Artifact{
        sampleSword("First", 1),
        sampleSword("Second", 2),
    })

    first := catalog.Slots[0]
    catalog.Award(first)

    if catalog.IsAvailable(0) {
        test.Errorf("awarding slot 0 should mark it unavailable")
    }

    if !catalog.IsAvailable(1) {
        test.Errorf("awarding slot 0 should not touch slot 1")
    }

    available := catalog.AvailableArtifacts()
    if len(available) != 1 || available[0].Name != "Second" {
        test.Errorf("expected only Second to remain available, got %v", available)
    }
}

func TestReplaceDoesNotResurrectAwardedSlot(test *testing.T) {
    catalog := MakeCatalog([]Artifact{
        sampleSword("First", 1),
        sampleSword("Second", 2),
    })

    catalog.Award(catalog.Slots[0])

    edited := sampleSword("Ultimate Defense", 3)
    catalog.Replace(0, &edited)

    if catalog.IsAvailable(0) {
        test.Errorf("editing an awarded slot must not make it available again")
    }

    if catalog.Slots[0].Name != "Ultimate Defense" {
        test.Errorf("expected slot 0 name to update, got %v", catalog.Slots[0].Name)
    }

    if catalog.Slots[0].CatalogIndex != 0 {
        test.Errorf("Replace should keep the slot index, got %v", catalog.Slots[0].CatalogIndex)
    }
}

func TestReplaceMutatesAwardedPointer(test *testing.T) {
    catalog := MakeCatalog([]Artifact{
        sampleSword("First", 1),
        sampleSword("Second", 2),
    })

    held := catalog.Slots[0]
    catalog.Award(held)

    edited := sampleSword("Ultimate Defense", 3)
    catalog.Replace(0, &edited)

    if held != catalog.Slots[0] {
        test.Errorf("Replace must keep the awarded pointer so equipped copies update")
    }

    if held.Name != "Ultimate Defense" || held.Powers[0].Amount != 3 {
        test.Errorf("awarded item should show the edit, got %+v", held)
    }
}

func TestReplaceUpdatesUnusedSlotForLaterAward(test *testing.T) {
    catalog := MakeCatalog([]Artifact{
        sampleSword("First", 1),
        sampleSword("Second", 2),
    })

    edited := sampleSword("Ultimate Defense", 3)
    catalog.Replace(1, &edited)

    remaining := catalog.AvailableArtifacts()
    if len(remaining) != 2 {
        test.Fatalf("editing an unused slot should leave it available, got %v items", len(remaining))
    }

    found := catalog.FindByName("Ultimate Defense")
    if found == nil || found != catalog.Slots[1] {
        test.Fatalf("edited slot should be findable by its new name")
    }

    catalog.Award(found)

    if catalog.IsAvailable(1) {
        test.Errorf("awarding the edited slot should mark it used")
    }

    if catalog.Slots[1].Powers[0].Amount != 3 {
        test.Errorf("awarded item should still carry the edited powers")
    }
}

func TestAwardByNameOnlyConsumesAvailableSlot(test *testing.T) {
    catalog := MakeCatalog([]Artifact{
        sampleSword("Shared", 1),
        sampleSword("Other", 2),
    })

    catalog.AwardByName("Shared")

    if catalog.IsAvailable(0) {
        test.Errorf("AwardByName should consume the named available slot")
    }

    catalog.AwardByName("Shared")
    if !catalog.IsAvailable(1) {
        test.Errorf("AwardByName must not consume a different slot")
    }
}

func TestWithAvailableNamesRestoresOldSaveList(test *testing.T) {
    catalog := MakeCatalog([]Artifact{
        sampleSword("First", 1),
        sampleSword("Second", 2),
        sampleSword("Third", 3),
    })

    restored := catalog.WithAvailableNames([]string{"Second", "Third"})

    if restored.IsAvailable(0) {
        test.Errorf("First was not in the remaining-name list")
    }

    if !restored.IsAvailable(1) || !restored.IsAvailable(2) {
        test.Errorf("Second and Third should still be available")
    }
}

func TestSerializeRoundTripPreservesEdits(test *testing.T) {
    catalog := MakeCatalog([]Artifact{
        sampleSword("First", 1),
        sampleSword("Second", 2),
    })

    edited := sampleSword("Ultimate Defense", 3)
    catalog.Replace(1, &edited)
    catalog.Award(catalog.Slots[0])

    loaded := ReconstructCatalog(catalog.Serialize(), catalog.AvailableMask(), spellbook.Spells{})

    if loaded.Len() != 2 {
        test.Fatalf("expected 2 slots after round trip, got %v", loaded.Len())
    }

    if loaded.IsAvailable(0) {
        test.Errorf("awarded slot 0 should stay unavailable")
    }

    if !loaded.IsAvailable(1) {
        test.Errorf("unused slot 1 should stay available")
    }

    if loaded.Slots[1].Name != "Ultimate Defense" || loaded.Slots[1].Powers[0].Amount != 3 {
        test.Errorf("edited slot did not survive the round trip: %+v", loaded.Slots[1])
    }
}

func TestRequirementsFromPowersUsesRealmNotSlotOrder(test *testing.T) {
    // official ITEMMAKE would have written Nature=1, Sorcery=4 because
    // +1 Attack (no books) is skipped and Elemental Armor is the "second"
    // property. Correct behavior is Nature=4 from the armor's own realm.
    powers := []Power{
        {Type: PowerTypeAttack, Amount: 1, Name: "+1 Attack"},
        {Type: PowerTypeAbility1, Amount: 4, Magic: data.NatureMagic, Ability: data.ItemAbilityElementalArmor},
        {Type: PowerTypeAbility2, Amount: 2, Magic: data.LifeMagic, Ability: data.ItemAbilityHolyAvenger},
    }

    got := RequirementsFromPowers(powers)

    if len(got) != 2 {
        test.Fatalf("expected Nature + Life requirements, got %+v", got)
    }

    if got[0].MagicType != data.NatureMagic || got[0].Amount != 4 {
        test.Errorf("expected Nature 4 first, got %+v", got[0])
    }

    if got[1].MagicType != data.LifeMagic || got[1].Amount != 2 {
        test.Errorf("expected Life 2 second, got %+v", got[1])
    }
}

func TestSamePowerMatchesCatalogAbilityStoredAsAbility1(test *testing.T) {
    catalogPower := Power{Type: PowerTypeAbility1, Ability: data.ItemAbilityFlaming}
    itempowPower := Power{Type: PowerTypeAbility2, Ability: data.ItemAbilityFlaming}

    if !samePower(catalogPower, itempowPower) {
        test.Errorf("Flaming stored as Ability1 should match the itempow Ability2 entry")
    }

    if samePower(catalogPower, Power{Type: PowerTypeAbility1, Ability: data.ItemAbilityBless}) {
        test.Errorf("different abilities should not match")
    }
}

func TestFilterPowersForTypeDropsIncompatible(test *testing.T) {
    flaming := Power{Type: PowerTypeAbility1, Ability: data.ItemAbilityFlaming, Name: "Flaming"}
    item := &Artifact{
        Type: ArtifactTypeSword,
        Powers: []Power{
            {Type: PowerTypeAttack, Amount: 3, Name: "+3 Attack"},
            flaming,
            {Type: PowerTypeSpellCharges, Amount: 2, Name: "Fireball x2"},
        },
    }

    compat := map[Power]set.Set[ArtifactType]{
        flaming: *set.NewSet(ArtifactTypeSword, ArtifactTypeAxe),
    }

    FilterPowersForType(item, ArtifactTypeShield, compat)

    if item.Type != ArtifactTypeShield {
        test.Errorf("expected type to change to shield")
    }

    if len(item.Powers) != 1 || item.Powers[0].Type != PowerTypeAttack {
        test.Errorf("expected only the attack bonus to remain, got %+v", item.Powers)
    }
}
