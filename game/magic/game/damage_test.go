package game

import (
    "testing"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
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
    testUnit.Defense = 0

    baseUnit := units.MakeOverworldUnitFromUnit(testUnit, 1, 1, data.PlaneArcanus, data.BannerRed, &NoExperienceInfo{})

    wrapper := UnitDamageWrapper{
        Unit: baseUnit,
    }

    if wrapper.GetHealth() != 16 {
        test.Errorf("expected health to be 16, got %d", wrapper.GetHealth())
    }

    if baseUnit.GetHealth() != 16 {
        test.Errorf("expected health to be 16, got %d", baseUnit.GetHealth())
    }

    damage := combat.ApplyDamage(&wrapper, 4, units.DamageRangedMagical, combat.DamageSourceSpell, combat.DamageModifiers{Magic: data.ChaosMagic})
    if damage != 4 {
        test.Errorf("expected 4 damage, got %d", damage)
    }

    if wrapper.GetHealth() != 12 {
        test.Errorf("expected health to be 12, got %d", wrapper.GetHealth())
    }

    if baseUnit.GetHealth() != 12 {
        test.Errorf("expected health to be 12, got %d", baseUnit.GetHealth())
    }
}
