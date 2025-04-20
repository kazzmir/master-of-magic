package game

import (
    "testing"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/hero"
)

type NoExperienceInfo struct {
}

func (noInfo *NoExperienceInfo) HasWarlord() bool {
    return false
}

func (noInfo *NoExperienceInfo) Crusade() bool {
    return false
}

func TestDamageNormal(test *testing.T) {
    testUnit := units.LizardSpearmen
    // prevent defense rolls so the damage is deterministic
    testUnit.Defense = 0

    baseUnit := units.MakeOverworldUnitFromUnit(testUnit, 1, 1, data.PlaneArcanus, data.BannerRed, &NoExperienceInfo{}, &units.NoEnchantments{})

    wrapper := UnitDamageWrapper{
        StackUnit: baseUnit,
    }

    maxHealth := testUnit.HitPoints * testUnit.Count

    if wrapper.GetHealth() != maxHealth {
        test.Errorf("expected health to be %v, got %d", maxHealth, wrapper.GetHealth())
    }

    if baseUnit.GetHealth() != maxHealth {
        test.Errorf("expected health to be %v, got %d", maxHealth, baseUnit.GetHealth())
    }

    strength := 4

    damage, _ := combat.ApplyDamage(&wrapper, strength, units.DamageRangedMagical, combat.DamageSourceSpell, combat.DamageModifiers{Magic: data.ChaosMagic})
    if damage != strength {
        test.Errorf("expected %v damage, got %d", strength, damage)
    }

    if wrapper.GetHealth() != maxHealth - strength {
        test.Errorf("expected health to be %v, got %d", maxHealth - strength, wrapper.GetHealth())
    }

    if baseUnit.GetHealth() != maxHealth - strength {
        test.Errorf("expected health to be %v, got %d", maxHealth - strength, baseUnit.GetHealth())
    }
}

func TestDamageHero(test *testing.T) {
    testUnit := units.HeroRakir
    // prevent defense rolls so the damage is deterministic
    testUnit.Defense = 0

    baseUnit := hero.MakeHero(units.MakeOverworldUnitFromUnit(testUnit, 1, 1, data.PlaneArcanus, data.BannerRed, &NoExperienceInfo{}, &units.NoEnchantments{}), hero.HeroRakir, "Rakir")

    wrapper := UnitDamageWrapper{
        StackUnit: baseUnit,
    }

    if wrapper.GetHealth() != units.HeroRakir.HitPoints {
        test.Errorf("expected health to be %v, got %d", units.HeroRakir.HitPoints, wrapper.GetHealth())
    }

    if baseUnit.GetHealth() != units.HeroRakir.HitPoints {
        test.Errorf("expected health to be %v, got %d", units.HeroRakir.HitPoints, baseUnit.GetHealth())
    }

    strength := 4

    damage, _ := combat.ApplyDamage(&wrapper, strength, units.DamageRangedMagical, combat.DamageSourceSpell, combat.DamageModifiers{Magic: data.ChaosMagic})
    if damage != strength {
        test.Errorf("expected %v damage, got %d", strength, damage)
    }

    if wrapper.GetHealth() != units.HeroRakir.HitPoints - strength {
        test.Errorf("expected health to be %v, got %d", units.HeroRakir.HitPoints - strength, wrapper.GetHealth())
    }

    if baseUnit.GetHealth() != units.HeroRakir.HitPoints - strength {
        test.Errorf("expected health to be %v, got %d", units.HeroRakir.HitPoints - strength, baseUnit.GetHealth())
    }
}
