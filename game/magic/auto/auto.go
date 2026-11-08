package auto

import (
    "fmt"
    "log"
    "slices"

    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/game"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    settingslib "github.com/kazzmir/master-of-magic/game/magic/settings"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    "github.com/kazzmir/master-of-magic/lib/lbx"
)

type HeadlessHandlers struct {
}

var _ game.MovementHandler = &HeadlessHandlers{}

func (handler *HeadlessHandlers) ShowMovement(x int, y int, stack *playerlib.UnitStack, center bool) {
}

func (handler *HeadlessHandlers) DoEncounter(player *playerlib.Player, stack *playerlib.UnitStack, encounter *maplib.ExtraEncounter, map_ *maplib.Map, x int, y int) combat.CombatState {
    return combat.CombatStateDefenderWin
}

func (handler *HeadlessHandlers) DoCombat(player *playerlib.Player, stack *playerlib.UnitStack, enemy *playerlib.Player, enemyStack *playerlib.UnitStack, zone combat.ZoneType) combat.CombatState {
    return combat.CombatStateDefenderWin
}

func (handler *HeadlessHandlers) DefeatCity(player *playerlib.Player, stack *playerlib.UnitStack, enemy *playerlib.Player, city *citylib.City) (bool, int) {
    return true, 0
}

func (handler *HeadlessHandlers) DiscoverWizards() {
}

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

    gameSettings := settingslib.MakeSettings(lbxCache)

    gameModel := game.MakeGameModel(terrainData, settings, gameSettings, data.PlaneArcanus, events, heroNames, allSpells, game.CreateArtifactPool(lbxCache), buildingInfo, lbxCache)

    for range 100 {

        log.Printf("Turn %d", gameModel.TurnNumber)

        for _, player := range gameModel.Players {
            out := player.AIBehavior.Update(player, gameModel)
            gameModel.ApplyAIDecisions(player, out)

            gameModel.DoNextTurn()

            for _, stack := range slices.Clone(player.Stacks) {
                // stop moving once any unit in the stack has no moves left
                for !stack.AnyOutOfMoves() && len(stack.CurrentPath) > 0 {
                    stack.CurrentPath = gameModel.DoAiMoveUnit(&HeadlessHandlers{}, player, stack)
                }
            }

            // continue any in-progress road building and attempt to meld any node
            // a stack is now standing on (engineers/melders may have just arrived).
            for _, stack := range slices.Clone(player.Stacks) {
                for _, unit := range stack.Units() {
                    if unit.GetBusy() == units.BusyStatusBuildRoad && len(unit.GetBuildRoadPath()) > 0 {
                        gameModel.MaybeBuildRoads(stack, player)
                        break
                    }
                }
            }
            for _, stack := range slices.Clone(player.Stacks) {
                gameModel.TryMeldNode(stack, player)
            }

            player.AIBehavior.PostUpdate(player, gameModel)
        }
    }

    return nil
}
