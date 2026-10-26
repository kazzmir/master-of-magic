package protocol

import (
	"fmt"
	"math"
	"testing"
)

func TestDefaultPortraitRange(t *testing.T) {
	for range 100 {
		for _, host := range []bool{true, false} {
			loadout := DefaultLoadout(host)
			if loadout.Portrait < 0 || loadout.Portrait >= len(PortraitNames) {
				t.Fatal("default portrait out of range", loadout.Portrait)
			}
			if _, err := ValidateLoadout(loadout, DefaultBudget, true); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestSpellCostsResearchDividedByTen(t *testing.T) {
	spell, ok := FindSpell("Fire Bolt") // Research cost: 60.
	if !ok || spell.Cost != 6 || spell.Mana != 5 || spell.Magic != "chaos" {
		t.Fatal("incorrect spell price", spell)
	}
	loadout := Loadout{Spells: []string{"Fire Bolt"}}
	loadout.Books[4] = 1
	if spent, err := ValidateLoadout(loadout, BookCost+6, false); err != nil || spent != BookCost+6 {
		t.Fatal("incorrect spell spending", spent, err)
	}
	if _, err := ValidateLoadout(loadout, BookCost+5, false); err == nil {
		t.Fatal("spell bought below its price")
	}
}

func TestTwentyUnitArmyLimit(t *testing.T) {
	loadout := DefaultLoadout(true)
	unit := loadout.Army[0]
	loadout.Army = nil
	for range 20 {
		loadout.Army = append(loadout.Army, unit)
	}
	if _, err := ValidateLoadout(loadout, math.MaxUint64, true); err != nil {
		t.Fatal("20-unit army rejected", err)
	}
	loadout.Army = append(loadout.Army, unit)
	if _, err := ValidateLoadout(loadout, math.MaxUint64, true); err == nil {
		t.Fatal("21-unit army accepted")
	}
}

func TestManaCostsTen(t *testing.T) {
	loadout := Loadout{Mana: 1}
	if spent, err := ValidateLoadout(loadout, 10, false); err != nil || spent != 10 {
		t.Fatal("one mana must cost exactly 10 budget", spent, err)
	}
	if _, err := ValidateLoadout(loadout, 9, false); err == nil {
		t.Fatal("one mana purchased with less than 10 budget")
	}
}

func TestLoadoutBudget(t *testing.T) {
	loadout := DefaultLoadout(true)
	loadout.Mana = 123
	loadout.Books[4] = 1
	for _, spell := range SpellsCatalog() {
		if spell.Magic == "chaos" {
			loadout.Spells = []string{spell.Name}
			break
		}
	}
	if len(loadout.Spells) == 0 {
		t.Fatal("missing combat spell catalog")
	}
	unit, _ := FindUnit(loadout.Army[0])
	spell, _ := FindSpell(loadout.Spells[0])
	want := 2*unit.Cost + 123*ManaCost + BookCost + spell.Cost
	spent, err := ValidateLoadout(loadout, math.MaxUint64, true)
	if err != nil || spent != want {
		t.Fatal(spent, want, err)
	}
	budget := uint64(1) // Exceed the budget early, but still count the whole loadout.
	if spent, err := ValidateLoadout(loadout, budget, true); err == nil || spent != want || err.Error() != fmt.Sprintf("loadout exceeds budget: used %d / %d", want, budget) {
		t.Fatal("overspending error missing full spending total/budget", spent, err)
	}
	if _, err := ValidateLoadout(loadout, want, true); err != nil {
		t.Fatal(err)
	}
	loadout.Portrait = 13
	if spent, err := ValidateLoadout(loadout, want, true); err != nil || spent != want {
		t.Fatal("portrait changed spending", err)
	}
}

func TestLoadoutRejectsInvalidSelections(t *testing.T) {
	for _, kind := range []string{"unit", "army", "portrait", "books", "mana", "spell", "duplicate", "realm"} {
		t.Run(kind, func(t *testing.T) {
			loadout := DefaultLoadout(false)
			switch kind {
			case "unit":
				loadout.Army[0] = UnitID{LbxFile: "invalid", Index: 999}
			case "army":
				loadout.Army = make([]UnitID, 21)
			case "portrait":
				loadout.Portrait = 14
			case "books":
				loadout.Books[0] = math.MaxUint64
			case "mana":
				loadout.Mana = math.MaxUint64
			case "spell":
				loadout.Spells = []string{"not a spell"}
			case "duplicate":
				loadout.Spells = []string{SpellsCatalog()[0].Name, SpellsCatalog()[0].Name}
				for i := range loadout.Books {
					loadout.Books[i] = 1
				}
			case "realm":
				for _, spell := range SpellsCatalog() {
					if MagicBookIndex(spell.Magic) >= 0 {
						loadout.Spells = []string{spell.Name}
						break
					}
				}
			}
			if _, err := ValidateLoadout(loadout, math.MaxUint64, true); err == nil {
				t.Fatal("invalid selection accepted")
			}
		})
	}
	if _, err := ValidateLoadout(Loadout{}, DefaultBudget, false); err != nil {
		t.Fatal("empty draft rejected", err)
	}
	if _, err := ValidateLoadout(Loadout{}, DefaultBudget, true); err == nil {
		t.Fatal("empty army ready accepted")
	}
}
