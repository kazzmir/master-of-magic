package engine

import (
	"slices"
	"testing"

	"github.com/kazzmir/master-of-magic/game/arena/player"
	"github.com/kazzmir/master-of-magic/game/magic/data"
	"github.com/kazzmir/master-of-magic/game/magic/units"
)

func TestCloneCost(t *testing.T) {
	for _, test := range []struct {
		bonus data.WeaponBonus
		cost  uint64
	}{
		{data.WeaponNone, 0},
		{data.WeaponMagic, 100},
		{data.WeaponMythril, 400},
		{data.WeaponAdamantium, 1000},
	} {
		p := player.MakePlayer(data.BannerGreen)
		u := p.AddUnit(units.LizardSwordsmen)
		u.SetWeaponBonus(test.bonus)
		u.AddExperience(100)
		u.AdjustHealth(-1)
		if got, want := getCloneCost(u), uint64(150)+test.cost; got != want {
			t.Fatalf("weapon %v: cost = %d, want %d", test.bonus, got, want)
		}
		u.AddEnchantment(data.UnitEnchantmentBless)
		u.AddEnchantment(data.UnitEnchantmentGiantStrength)
		want := uint64(150) + test.cost + uint64(getEnchantmentCost(data.UnitEnchantmentBless)+getEnchantmentCost(data.UnitEnchantmentGiantStrength))
		if got := getCloneCost(u); got != want {
			t.Fatalf("enchanted weapon %v: cost = %d, want %d", test.bonus, got, want)
		}
	}
}

func TestCloneArmyUnit(t *testing.T) {
	for _, raw := range []units.Unit{units.LizardSwordsmen, units.GreatDrake} {
		p := player.MakePlayer(data.BannerGreen)
		source := p.AddUnit(raw)
		source.AddExperience(100)
		source.AddEnchantment(data.UnitEnchantmentGiantStrength)
		source.AddEnchantment(data.UnitEnchantmentHolyArmor)
		if raw.Race != data.RaceFantastic {
			source.SetWeaponBonus(data.WeaponAdamantium)
		}
		source.AdjustHealth(-1)
		experience := source.GetExperience()
		p.Money = getCloneCost(source)
		clone := cloneArmyUnit(p, source)
		if clone == nil || clone == source || len(p.Units) != 2 || p.Money != 0 {
			t.Fatal("clone should add a distinct unit and deduct the exact price")
		}
		cloneRaw := clone.GetRawUnit()
		if !cloneRaw.Equals(raw) || clone.GetWeaponBonus() != source.GetWeaponBonus() || !slices.Equal(clone.GetEnchantments(), source.GetEnchantments()) {
			t.Fatal("clone did not preserve unit type and upgrades")
		}
		if clone.GetExperience() != 0 || clone.GetHealth() != clone.GetMaxHealth() {
			t.Fatal("clone should start at full health and zero experience")
		}
		clone.AddEnchantment(data.UnitEnchantmentBless)
		clone.AdjustHealth(-1)
		if source.HasEnchantment(data.UnitEnchantmentBless) || source.GetDamage() != 1 || source.GetExperience() != experience {
			t.Fatal("cloning or modifying a clone affected its source")
		}
	}
}

func TestCloneArmyUnitCannotAfford(t *testing.T) {
	p := player.MakePlayer(data.BannerGreen)
	source := p.AddUnit(units.LizardSwordsmen)
	p.Money = getCloneCost(source) - 1
	before := p.Money
	if cloneArmyUnit(p, source) != nil || p.Money != before || len(p.Units) != 1 {
		t.Fatal("unaffordable clone must not change the army or gold")
	}
	p.RemoveUnit(source)
	p.Money = 10000
	if cloneArmyUnit(p, source) != nil || cloneArmyUnit(p, nil) != nil || p.Money != 10000 || len(p.Units) != 0 {
		t.Fatal("a missing or sold unit must not be cloned")
	}
}
