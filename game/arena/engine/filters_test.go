package engine

import (
	"testing"

	"github.com/kazzmir/master-of-magic/game/magic/units"
)

func TestUnitShopFilters(t *testing.T) {
	for _, test := range []struct {
		name    string
		filters unitShopFilters
		unit    units.Unit
		gold    uint64
		want    bool
	}{
		{"no filters", unitShopFilters{}, units.LizardSwordsmen, 0, true},
		{"affordable", unitShopFilters{Affordable: true}, units.LizardSwordsmen, getUnitCost(&units.LizardSwordsmen), true},
		{"too expensive", unitShopFilters{Affordable: true}, units.LizardSwordsmen, 0, false},
		{"physical ranged", unitShopFilters{Ranged: true}, units.HighMenBowmen, 0, true},
		{"magical ranged", unitShopFilters{Ranged: true}, units.Warlocks, 0, true},
		{"boulder ranged", unitShopFilters{Ranged: true}, units.Catapult, 0, true},
		{"melee excluded", unitShopFilters{Ranged: true}, units.LizardSwordsmen, 0, false},
		{"flying", unitShopFilters{Flying: true}, units.SkyDrake, 0, true},
		{"ground excluded", unitShopFilters{Flying: true}, units.Warlocks, 0, false},
		{"flying and ranged", unitShopFilters{Flying: true, Ranged: true}, units.Pegasai, 0, true},
		{"ranged but not flying", unitShopFilters{Flying: true, Ranged: true}, units.Catapult, 0, false},
		{"flying but not ranged", unitShopFilters{Flying: true, Ranged: true}, units.GreatDrake, 0, false},
		{"all filters", unitShopFilters{Affordable: true, Flying: true, Ranged: true}, units.Pegasai, getUnitCost(&units.Pegasai), true},
		{"all filters too expensive", unitShopFilters{Affordable: true, Flying: true, Ranged: true}, units.Pegasai, 0, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := test.filters.matches(&test.unit, test.gold); got != test.want {
				t.Fatalf("matches = %v, want %v", got, test.want)
			}
		})
	}
}
