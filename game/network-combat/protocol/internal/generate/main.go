// Generate the headless multiplayer catalog from the game's canonical metadata.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/kazzmir/master-of-magic/game/magic/spellbook"
	"github.com/kazzmir/master-of-magic/game/magic/units"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
	"github.com/kazzmir/master-of-magic/lib/lbx"
)

func main() {
	data := flag.String("data", "magic-data", "LBX directory")
	out := flag.String("out", "catalog.json", "output file")
	flag.Parse()
	var catalog protocol.Catalog
	for _, unit := range units.AllUnits {
		cost := unit.ProductionCost
		if cost == 0 {
			cost = unit.CastingCost
		}
		if cost <= 0 || unit.MeleeAttackPower == 0 && unit.RangedAttackPower == 0 {
			continue
		}
		id := unit.GetId()
		catalog.Units = append(catalog.Units, protocol.UnitChoice{ID: protocol.UnitID{LbxFile: id.LbxFile, Index: id.Index}, Name: fmt.Sprintf("%s %s", unit.Race.String(), unit.Name), Cost: uint64(cost)})
	}
	cache := lbx.CacheFromPath(*data)
	if cache == nil {
		log.Fatal("no LBX data")
	}
	spells, err := spellbook.ReadSpellsFromCache(cache)
	if err != nil {
		log.Fatal(err)
	}
	for _, spell := range spells.Spells {
		if !spell.Valid() || !spell.Eligibility.CanCastInCombat(false) {
			continue
		}
		catalog.Spells = append(catalog.Spells, protocol.SpellChoice{Name: spell.Name, Cost: uint64(max(0, spell.ResearchCost)) / 10, Magic: strings.ToLower(spell.Magic.String())})
	}
	sort.Slice(catalog.Units, func(i, j int) bool { return catalog.Units[i].Name < catalog.Units[j].Name })
	sort.Slice(catalog.Spells, func(i, j int) bool { return catalog.Spells[i].Name < catalog.Spells[j].Name })
	encoded, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*out, append(encoded, '\n'), 0644); err != nil {
		log.Fatal(err)
	}
}
