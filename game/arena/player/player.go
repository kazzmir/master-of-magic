package player

import (
    "slices"
    "image"

    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
)

type Player struct {
    Money uint64
    Wizard setup.WizardCustom
    Level int
    AI bool
    Mana int
    OriginalMana int

    KnownSpells spellbook.Spells

    Units []units.StackUnit
}

func MakePlayer(banner data.BannerType) *Player {
    return &Player{
        Money: 200,
        Wizard: setup.WizardCustom{
            Name: "Player",
            Banner: banner,
        },
        Level: 1,
    }
}

func MakeAIPlayer(banner data.BannerType) *Player {
    return &Player{
        Money: 1000,
        AI: true,
        Wizard: setup.WizardCustom{
            Name: "Enemy",
            Banner: banner,
        },
    }
}

func (player *Player) RemoveUnit(unit units.StackUnit) {
    player.Units = slices.DeleteFunc(player.Units, func(u units.StackUnit) bool {
        return u == unit
    })
}

func (player *Player) AddUnit(unit units.Unit) units.StackUnit {
    newUnit := units.MakeOverworldUnitFromUnit(unit, 0, 0, data.PlaneArcanus, player.Wizard.Banner, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    player.Units = append(player.Units, newUnit)
    return newUnit
}

func (player *Player) GetKnownSpells() spellbook.Spells {
    return player.KnownSpells
}

func (player *Player) GetWizard() *setup.WizardCustom {
    return &player.Wizard
}

// need this to compute tile distance so that in combat the range of spell costs can be computed
type BasicCatchmentProvider struct {
}

func (provider *BasicCatchmentProvider) GetCatchmentArea(x int, y int) map[image.Point]maplib.FullTile {
    return nil
}

func (provider *BasicCatchmentProvider) GetGoldBonus(x int, y int) int {
    return 0
}

func (provider *BasicCatchmentProvider) OnShore(x int, y int) bool {
    return false
}

func (provider *BasicCatchmentProvider) ByRiver(x int, y int) bool {
    return false
}

func abs(a int) int {
    if a < 0 {
        return -a
    }
    return a
}

func (provider *BasicCatchmentProvider) TileDistance(x1 int, y1 int, x2 int, y2 int) int {
    return abs(x1 - x2) + abs(y1 - y2)
}

func (player *Player) FindFortressCity() *citylib.City {
    return &citylib.City{
        Plane: data.PlaneArcanus,
        // just far enough away to have range be 1 for magic
        X: 5,
        Y: 0,
        CatchmentProvider: &BasicCatchmentProvider{},
    }
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
    return player.Mana
}

func (player *Player) UseMana(mana int) {
    player.Mana -= mana
}

func (player *Player) ComputeCastingSkill() int {
    return player.OriginalMana
}

func (player *Player) IsHuman() bool {
    return !player.AI
}

func (player *Player) IsAI() bool {
    return player.AI
}

func (player *Player) ComputeEffectiveResearchPerTurn(cost float64, spell spellbook.Spell) int {
    return 0
}

func (player *Player) ComputeEffectiveSpellCost(spell spellbook.Spell, overland bool) int {
    return spellbook.ComputeSpellCost(&player.Wizard, spell, overland, false)
}
