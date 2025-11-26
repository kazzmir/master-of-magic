package hero

import (
    "testing"
    "math"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
)

func floatEqual(a, b float32) bool {
    return math.Abs(float64(a - b)) < 0.0001
}

func TestHero(test *testing.T){
    zaldron := MakeHeroSimple(HeroZaldron)

    if len(zaldron.GetAbilities()) != 2 {
        test.Errorf("Zaldron should have two ability")
    }

    if !zaldron.HasAbility(data.AbilitySage) {
        test.Errorf("Zaldron should have the Sage ability")
    }

    taki := MakeHeroSimple(HeroTaki)
    if len(taki.GetAbilities()) != 1 && taki.GetAbilities()[0].Ability != data.AbilitySuperAgility {
        test.Errorf("taki should have the super agility ability but was %v", taki.GetAbilities())
    }

    taki.SetExtraAbilities()

    if len(taki.GetAbilities()) != 2 {
        test.Errorf("taki should have 2 abilities")
    }

    // 5 base + 1 from super agility
    if taki.GetDefense() != 6 {
        test.Errorf("taki should have 6 defense but was %v", taki.GetDefense())
    }

    taki.AddExperience(units.ExperienceChampionHero.ExperienceRequired(false, false))

    // 5 base, 2 from champion, 7 from super agility
    if taki.GetDefense() != 5 + 2 + 7 {
        test.Errorf("taki should have 14 defense but was %v", taki.GetDefense())
    }

    theria := MakeHeroSimple(HeroTheria)
    if len(theria.GetAbilities()) != 2 {
        test.Errorf("Theria should have 2 abilities")
    }

    // theria already has charmed, cant add it twice
    if theria.AddAbility(data.AbilityCharmed) {
        test.Errorf("Theria should not be able to add the Charmed ability")
    }

    // can add noble once
    if !theria.AddAbility(data.AbilityNoble) {
        test.Errorf("Theria should be able to add the Noble ability")
    }

    // but not again
    if theria.AddAbility(data.AbilityNoble) {
        test.Errorf("Theria should not be able to add the Noble ability again")
    }

    if int(theria.GetAbilityValue(data.AbilityCaster)) != 0 {
        test.Errorf("Theria should not have the Caster ability")
    }

    if !theria.AddAbility(data.AbilityCaster) {
        test.Errorf("Theria should be able to add the Caster ability")
    }

    if !floatEqual(theria.GetAbilityValue(data.AbilityCaster), 2.5) {
        test.Errorf("Theria should have 2.5 Caster ability but was %v", theria.GetAbilityValue(data.AbilityCaster))
    }

    if !theria.AddAbility(data.AbilityCaster) {
        test.Errorf("Theria should be able to add the Caster ability")
    }

    if !floatEqual(theria.GetAbilityValue(data.AbilityCaster), 5) {
        test.Errorf("Theria should have 5 Caster ability")
    }

    if len(theria.GetBaseProgression()) != 0 {
        test.Errorf("There should be no progression yet")
    }

    theria.GainLevel(units.ExperienceMyrmidon)

    if len(theria.GetBaseProgression()) != 4 {
        test.Errorf("There should be progression yet")
    }
}

func TestHeroProgression(test *testing.T) {
    zaldron := MakeHeroSimple(HeroZaldron)
    if !zaldron.AddAbility(data.AbilityAgility) {
        test.Errorf("unable to add agility")
    }

    zaldron.AddExperience(units.ExperienceChampionHero.ExperienceRequired(false, false))
    if zaldron.GetAbilityDefense() != 5 {
        test.Errorf("Expected defense for champion with agility to be 5 but was %v", zaldron.GetAbilityDefense())
    }

    // zaldron starts with 7.5, at champion level, caster should be 37
    if int(zaldron.GetAbilityValue(data.AbilityCaster)) != 37 {
        test.Errorf("Expected caster level to be 37 but was %v", zaldron.GetAbilityValue(data.AbilityCaster))
    }
}
