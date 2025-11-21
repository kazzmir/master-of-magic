package game

import (
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/ai"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
)

type GameModel struct {
    ArcanusMap *maplib.Map
    MyrrorMap *maplib.Map
    Players []*playerlib.Player
    Plane data.Plane

    heroNames map[int]map[herolib.HeroType]string
    allSpells spellbook.Spells
}

func MakeGameModel(terrainData *terrain.TerrainData, settings setup.NewGameSettings, startingPlane data.Plane, cityProvider maplib.CityProvider, heroNames map[int]map[herolib.HeroType]string, allSpells spellbook.Spells) *GameModel {

    planeTowers := maplib.GeneratePlaneTowerPositions(settings.LandSize, 6)

    return &GameModel{
        ArcanusMap: maplib.MakeMap(terrainData, settings.LandSize, settings.Magic, settings.Difficulty, data.PlaneArcanus, cityProvider, planeTowers),
        MyrrorMap: maplib.MakeMap(terrainData, settings.LandSize, settings.Magic, settings.Difficulty, data.PlaneMyrror, cityProvider, planeTowers),
        heroNames: heroNames,
        allSpells: allSpells,
        Plane: startingPlane,
    }
}

func (model *GameModel) CurrentMap() *maplib.Map {
    if model.Plane == data.PlaneArcanus {
        return model.ArcanusMap
    }

    return model.MyrrorMap
}

func (model *GameModel) SwitchPlane() {
    switch model.Plane {
        case data.PlaneArcanus: model.Plane = data.PlaneMyrror
        case data.PlaneMyrror: model.Plane = data.PlaneArcanus
    }
}

func (model *GameModel) AddPlayer(wizard setup.WizardCustom, human bool) *playerlib.Player {
    useNames := model.heroNames[len(model.Players)]
    if useNames == nil {
        useNames = make(map[herolib.HeroType]string)
    }

    newPlayer := playerlib.MakePlayer(wizard, human, model.CurrentMap().Width(), model.CurrentMap().Height(), useNames, model)

    if !human {
        newPlayer.AIBehavior = ai.MakeEnemyAI()
        newPlayer.StrategicCombat = true
    }

    startingSpells := []string{"Magic Spirit", "Spell of Return"}
    if wizard.RetortEnabled(data.RetortArtificer) {
        startingSpells = append(startingSpells, "Enchant Item", "Create Artifact")
    }

    newPlayer.ResearchPoolSpells = wizard.StartingSpells.Copy()

    // not sure its necessary to add the starting spells to the research pool
    for _, spell := range startingSpells {
        newPlayer.ResearchPoolSpells.AddSpell(model.allSpells.FindByName(spell))
    }

    // every wizard gets all arcane spells by default
    newPlayer.ResearchPoolSpells.AddAllSpells(model.allSpells.GetSpellsByMagic(data.ArcaneMagic))

    newPlayer.KnownSpells = wizard.StartingSpells.Copy()
    for _, spell := range startingSpells {
        newPlayer.KnownSpells.AddSpell(model.allSpells.FindByName(spell))
    }
    newPlayer.CastingSkillPower = computeInitialCastingSkillPower(newPlayer.Wizard.Books)

    newPlayer.InitializeResearchableSpells(&model.allSpells)
    newPlayer.UpdateResearchCandidates()

    // log.Printf("Research spells: %v", newPlayer.ResearchPoolSpells)

    // famous wizards get a head start of 10 fame
    if wizard.RetortEnabled(data.RetortFamous) {
        newPlayer.Fame += 10
    }

    model.Players = append(model.Players, newPlayer)
    return newPlayer
}

// true if any alive player has the given enchantment enabled
func (model *GameModel) HasEnchantment(enchantment data.Enchantment) bool {
    for _, player := range model.Players {
        if !player.Defeated && player.HasEnchantment(enchantment) {
            return true
        }
    }

    return false
}

// true if any alive player that is not the given one has the given enchantment enabled
func (model *GameModel) HasRivalEnchantment(original *playerlib.Player, enchantment data.Enchantment) bool {
    for _, player := range model.Players {
        if !player.Defeated && player != original && player.HasEnchantment(enchantment) {
            return true
        }
    }

    return false
}

