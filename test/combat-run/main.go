package main

import (
    "log"
    "time"
    "bytes"
    _ "embed"

    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/lib/lbx"
)

//go:embed spelldat.lbx
var spellDataLbx []byte

func LoadSpellData() (spellbook.Spells, error) {
    buffer := bytes.NewReader(spellDataLbx)
    spellLbx, err := lbx.ReadLbx(buffer)
    if err != nil {
        return spellbook.Spells{}, err
    }
    return spellbook.ReadSpells(&spellLbx, 0)
}

type noGlobalEnchantments struct {
}

func (*noGlobalEnchantments) HasEnchantment(enchantment data.Enchantment) bool {
    return false
}

func (*noGlobalEnchantments) HasRivalEnchantment(player *player.Player, enchantment data.Enchantment) bool {
    return false
}

func main() {
    log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)

    defendingPlayer := player.MakePlayer(setup.WizardCustom{
        Name: "AI-1",
        Banner: data.BannerBrown,
    }, false, 0, 0, nil, &noGlobalEnchantments{})

    attackingPlayer := player.MakePlayer(setup.WizardCustom{
        Name: "AI-2",
        Banner: data.BannerRed,
    }, false, 0, 0, nil, &noGlobalEnchantments{})

    attackingArmy := &combat.Army{
        Player: attackingPlayer,
    }

    defendingArmy := &combat.Army{
        Player: defendingPlayer,
    }

    for range 3 {
        attackingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.LizardSwordsmen, 1, 1, data.PlaneArcanus, attackingPlayer.Wizard.Banner, attackingPlayer.MakeExperienceInfo(), attackingPlayer.MakeUnitEnchantmentProvider()))
    }

    for range 4 {
        // defendingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.BarbarianSpearmen, 1, 1, data.PlaneArcanus, defendingPlayer.Wizard.Banner, defendingPlayer.MakeExperienceInfo(), defendingPlayer.MakeUnitEnchantmentProvider()))
        defendingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.BeastmenPriest, 1, 1, data.PlaneArcanus, defendingPlayer.Wizard.Banner, defendingPlayer.MakeExperienceInfo(), defendingPlayer.MakeUnitEnchantmentProvider()))
    }

    allSpells, err := LoadSpellData()
    if err != nil {
        log.Fatalf("Failed to load spell data: %v", err)
    }

    model := combat.MakeCombatModel(allSpells, defendingArmy, attackingArmy, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, data.MagicNone, 0, 0, make(chan combat.CombatEvent, 10))

    start := time.Now()
    state := combat.Run(model)
    end := time.Now()
    log.Printf("Combat simulation took %v", end.Sub(start))
    log.Printf("Final state: %+v", state)
}
