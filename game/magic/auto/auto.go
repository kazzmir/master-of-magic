package auto

import (
    "fmt"
    "log"

    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/game"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    "github.com/kazzmir/master-of-magic/lib/lbx"
)

func RunAutoGame() error {
    settings := setup.NewGameSettings{
        Opponents: 3,
        Difficulty: data.DifficultyAverage,
        Magic: data.MagicSettingNormal,
        LandSize: 2,
        DisableRaiders: true,
    }

    lbxCache := lbx.AutoCache()
    if lbxCache == nil {
        return fmt.Errorf("failed to create lbx cache")
    }

    terrainLbx, err := lbxCache.GetLbxFile("terrain.lbx")
    if err != nil {
        log.Printf("Error: could not load terrain: %v", err)
        return nil
    }

    terrainData, err := terrain.ReadTerrainData(terrainLbx)
    if err != nil {
        log.Printf("Error: could not load terrain: %v", err)
        return nil
    }

    allSpells, err := spellbook.ReadSpellsFromCache(lbxCache)
    if err != nil {
        log.Printf("Could not read spells from cache: %v", err)
        return nil
    }

    heroNames := herolib.ReadNamesPerWizard(lbxCache)

    buildingInfo, err := buildinglib.ReadBuildingInfo(lbxCache)
    if err != nil {
        log.Printf("Unable to read building info: %v", err)
        return nil
    }

    events := make(chan game.GameEvent, 100)

    gameModel := game.MakeGameModel(terrainData, settings, data.PlaneArcanus, events, heroNames, allSpells, game.CreateArtifactPool(lbxCache), buildingInfo, lbxCache)

    for range 100 {

        for _, player := range gameModel.Players {
            out := player.AIBehavior.Update(player, gameModel)
            gameModel.ApplyAIDecisions(player, out)
        }

        gameModel.TurnNumber += 1
    }

    return nil
}
