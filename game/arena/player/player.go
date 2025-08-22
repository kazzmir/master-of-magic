package player

import (
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
)

type Player struct {
    Name string
    Money uint64
    Wizard setup.WizardCustom

    Units []units.StackUnit
}

func MakePlayer() *Player {
    return &Player{
        Name: "Player",
        Money: 1000,
        Units: []units.StackUnit{
            units.MakeOverworldUnitFromUnit(units.LizardSwordsmen, 0, 0, data.PlaneArcanus, data.BannerGreen, &units.NoExperienceInfo{}, &units.NoEnchantments{}),
        },
    }
}

func (player *Player) GetKnownSpells() spellbook.Spells {
    return spellbook.Spells{}
}

func (player *Player) GetWizard() *setup.WizardCustom {
    return &player.Wizard
}

func (player *Player) FindFortressCity() *citylib.City {
    return nil
}

func (player *Player) MakeExperienceInfo() units.ExperienceInfo {
    return &units.NoExperienceInfo{}
}

func (player *Player) MakeUnitEnchantmentProvider() units.GlobalEnchantmentProvider {
    return &units.NoEnchantments{}
}

func (player *Player) HasEnchantment(data.Enchantment) bool {
    return false
}

func (player *Player) GetMana() int {
    return 0
}

func (player *Player) UseMana(mana int) {
}

func (player *Player) ComputeCastingSkill() int {
    return 0
}

func (player *Player) IsHuman() bool {
    return true
}

func (player *Player) IsAI() bool {
    return false
}

func (player *Player) ComputeEffectiveResearchPerTurn(cost float64, spell spellbook.Spell) int {
    return 0
}

func (player *Player) ComputeEffectiveSpellCost(spell spellbook.Spell, overland bool) int {
    return spellbook.ComputeSpellCost(&player.Wizard, spell, overland, false)
}


