package protocol

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
)

const DefaultBudget uint64 = 1000
const BookCost uint64 = 100
const ManaCost uint64 = 20
const MaxArmy = 20
const MaxMana uint64 = 1_000_000

// Catalog metadata is shared with the headless server; no game assets or graphics
// packages are required for authoritative pricing/validation. Regenerate using
// go run ./game/network-combat/protocol/internal/generate -data magic-data -out game/network-combat/protocol/catalog.json
// (the generator may need xvfb-run on headless Linux).
//
//go:embed catalog.json
var catalogJSON []byte

type UnitID struct {
	LbxFile string `json:"lbx_file"`
	Index   int    `json:"index"`
}
type UnitChoice struct {
	ID   UnitID `json:"id"`
	Name string `json:"name"`
	Cost uint64 `json:"cost"`
}
type SpellChoice struct {
	Name  string `json:"name"`
	Cost  uint64 `json:"cost"`
	Magic string `json:"magic"`
}
type Catalog struct {
	Units  []UnitChoice  `json:"units"`
	Spells []SpellChoice `json:"spells"`
}

var catalog Catalog
var unitChoices = map[UnitID]UnitChoice{}
var spellChoices = map[string]SpellChoice{}

var PortraitNames = []string{"Merlin", "Raven", "Sharee", "Lo Pan", "Jafar", "Oberic", "Rjak", "Sss'ra", "Tauron", "Freya", "Horus", "Ariel", "Tlaloc", "Kali"}
var MagicNames = []string{"Life", "Sorcery", "Nature", "Death", "Chaos"}

func init() {
	if err := json.Unmarshal(catalogJSON, &catalog); err != nil {
		panic(err)
	}
	for _, unit := range catalog.Units {
		unitChoices[unit.ID] = unit
	}
	for _, spell := range catalog.Spells {
		if spell.Magic != "arcane" && MagicBookIndex(spell.Magic) < 0 {
			panic(fmt.Errorf("unknown spell magic %q", spell.Magic))
		}
		spellChoices[spell.Name] = spell
	}
}

// MagicBookIndex maps a named realm to its position in Loadout.Books.
// Arcane magic has no corresponding book.
func MagicBookIndex(magic string) int {
	for i, name := range MagicNames {
		if strings.EqualFold(name, magic) {
			return i
		}
	}
	return -1
}

func UnitsCatalog() []UnitChoice                { return slices.Clone(catalog.Units) }
func SpellsCatalog() []SpellChoice              { return slices.Clone(catalog.Spells) }
func FindUnit(id UnitID) (UnitChoice, bool)     { value, ok := unitChoices[id]; return value, ok }
func FindSpell(name string) (SpellChoice, bool) { value, ok := spellChoices[name]; return value, ok }

type Loadout struct {
	Army     []UnitID  `json:"army"`
	Books    [5]uint64 `json:"books"`
	Mana     uint64    `json:"mana"`
	Spells   []string  `json:"spells"`
	Portrait int       `json:"portrait"`
}

func (loadout Loadout) Clone() Loadout {
	loadout.Army = slices.Clone(loadout.Army)
	loadout.Spells = slices.Clone(loadout.Spells)
	return loadout
}
func (loadout Loadout) Equal(other Loadout) bool {
	return loadout.Books == other.Books && loadout.Mana == other.Mana && loadout.Portrait == other.Portrait && slices.Equal(loadout.Army, other.Army) && slices.Equal(loadout.Spells, other.Spells)
}

func DefaultLoadout(host bool) Loadout {
	id := UnitID{LbxFile: "units2.lbx", Index: 0}
	if host {
		id.Index = 17
	}
	return Loadout{Army: []UnitID{id, id}, Portrait: rand.IntN(len(PortraitNames))}
}

// ValidateLoadout computes spending with subtraction to avoid uint64 overflow.
// Portraits are free; units cost production/casting cost, books cost 100 each,
// mana costs 20 each, and known combat spells cost research cost / 10 (rounded down).
func ValidateLoadout(loadout Loadout, budget uint64, requireArmy bool) (uint64, error) {
	remaining := budget
	spend := func(cost uint64) error {
		if cost > remaining {
			return fmt.Errorf("loadout exceeds budget %d", budget)
		}
		remaining -= cost
		return nil
	}
	if loadout.Portrait < 0 || loadout.Portrait >= len(PortraitNames) {
		return 0, fmt.Errorf("invalid wizard portrait")
	}
	if len(loadout.Army) > MaxArmy || requireArmy && len(loadout.Army) == 0 {
		return 0, fmt.Errorf("choose 1–%d army units", MaxArmy)
	}
	for _, id := range loadout.Army {
		unit, ok := FindUnit(id)
		if !ok {
			return 0, fmt.Errorf("unknown army unit")
		}
		if err := spend(unit.Cost); err != nil {
			return 0, err
		}
	}
	var books uint64
	for _, count := range loadout.Books {
		if count > 11 {
			return 0, fmt.Errorf("at most 11 magic books")
		}
		books += count
	}
	if books > 11 {
		return 0, fmt.Errorf("at most 11 magic books total")
	}
	if err := spend(books * BookCost); err != nil {
		return 0, err
	}
	if loadout.Mana > MaxMana {
		return 0, fmt.Errorf("mana cannot exceed %d", MaxMana)
	}
	if err := spend(loadout.Mana * ManaCost); err != nil {
		return 0, err
	}
	if len(loadout.Spells) > len(catalog.Spells) {
		return 0, fmt.Errorf("too many known spells")
	}
	seen := make(map[string]bool)
	for _, name := range loadout.Spells {
		spell, ok := FindSpell(name)
		if !ok || seen[name] {
			return 0, fmt.Errorf("invalid or duplicate spell %q", name)
		}
		seen[name] = true
		if index := MagicBookIndex(spell.Magic); index >= 0 && loadout.Books[index] == 0 {
			return 0, fmt.Errorf("%s requires a %s book", name, MagicNames[index])
		}
		if err := spend(spell.Cost); err != nil {
			return 0, err
		}
	}
	return budget - remaining, nil
}

// CombatSetup is exchanged directly between peers over the encrypted relay.
type CombatSetup struct {
	Type       string  `json:"type"`
	GameID     uint64  `json:"game_id"`
	ClientID   uint64  `json:"client_id"`
	Name       string  `json:"name"`
	Budget     uint64  `json:"budget"`
	AttackerID uint64  `json:"attacker_id"`
	Revision   uint64  `json:"revision"`
	Loadout    Loadout `json:"loadout"`
}
