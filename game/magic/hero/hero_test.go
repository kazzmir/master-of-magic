package hero

import (
    "testing"
    "github.com/kazzmir/master-of-magic/game/magic/units"
)

func TestHero(test *testing.T){
    zaldron := MakeHeroSimple(HeroZaldron)

    if len(zaldron.GetAbilities()) != 2 {
        test.Errorf("Zaldron should have two ability")
    }

    if !zaldron.HasAbility(units.AbilitySage) {
        test.Errorf("Zaldron should have the Sage ability")
    }

    taki := MakeHeroSimple(HeroTaki)
    if len(taki.GetAbilities()) != 1 && taki.GetAbilities()[0].Ability != units.AbilitySuperAgility {
        test.Errorf("taki should have the super agility ability but was %v", taki.GetAbilities())
    }

    taki.SetExtraAbilities()

    if len(taki.GetAbilities()) != 2 {
        test.Errorf("taki should have 2 abilities")
    }

    theria := MakeHeroSimple(HeroTheria)
    if len(theria.GetAbilities()) != 2 {
        test.Errorf("Theria should have 2 abilities")
    }

    // theria already has charmed, cant add it twice
    if theria.AddAbility(units.AbilityCharmed) {
        test.Errorf("Theria should not be able to add the Charmed ability")
    }

    // can add noble once
    if !theria.AddAbility(units.AbilityNoble) {
        test.Errorf("Theria should be able to add the Noble ability")
    }

    // but not again
    if theria.AddAbility(units.AbilityNoble) {
        test.Errorf("Theria should not be able to add the Noble ability again")
    }
}
