package main

import (
    "os"
    "fmt"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    mouselib "github.com/kazzmir/master-of-magic/lib/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/load"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    Game *gamelib.Game
    Coroutine *coroutine.Coroutine
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    engine.Game.Draw(screen)
    mouse.Mouse.Draw(screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
}

func (engine *Engine) Update() error {

    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    inputmanager.Update()

    if engine.Coroutine.Run() != nil {
        return ebiten.Termination
    }

    return nil
}

func createGame(cache *lbx.LbxCache, saveGame *load.SaveGame) *gamelib.Game {
    game := gamelib.MakeGame(cache, saveGame.ToSettings())

    // load data
    game.ArcanusMap = saveGame.ToMap(game.ArcanusMap.Data, data.PlaneArcanus, nil)
    game.MyrrorMap = saveGame.ToMap(game.MyrrorMap.Data, data.PlaneMyrror, nil)
    game.TurnNumber = uint64(saveGame.Turn)
    // FIXME: game.ArtifactPool
    // FIXME: game.RandomEvents
    // FIXME: game.RoadWorkArcanus
    // FIXME: game.RoadWorkMyrror
    // FIXME: game.PurifyWorkArcanus
    // FIXME: game.PurifyWorkMyrror
    // FIXME: game.Players

    wizard := saveGame.ToWizard(0)

    player := game.AddPlayer(wizard, true)
    player.LiftFog(20, 20, 50, data.PlaneArcanus)
    player.LiftFog(20, 20, 50, data.PlaneMyrror)
    player.ArcanusFog = saveGame.ToFogMap(data.PlaneArcanus)
    player.MyrrorFog = saveGame.ToFogMap(data.PlaneMyrror)
    player.UpdateFogVisibility()
    player.Cities = saveGame.ToCities(player, 0, game)

    for i := 1; i < int(saveGame.NumPlayers); i++ {
        wizard := saveGame.ToWizard(i)
        enemy := game.AddPlayer(wizard, false)
        enemy.Cities = saveGame.ToCities(enemy, int8(i), game)
        player.AwarePlayer(enemy)
    }

    game.Camera.Center(20, 20)
    if len(player.Cities) > 0 {
        game.Camera.Center(player.Cities[0].X, player.Cities[0].Y)
    }

    // player.Admin = true
    // player.LiftFog(20, 20, 50, data.PlaneArcanus)

    return game
}

func NewEngine(saveGame *load.SaveGame) (*Engine, error) {
    cache := lbx.AutoCache()

    game := createGame(cache, saveGame)
    game.DoNextTurn()

    run := func(yield coroutine.YieldFunc) error {
        for game.Update(yield) != gamelib.GameStateQuit {
            yield()
        }

        return ebiten.Termination
    }

    normalMouse, err := mouselib.GetMouseNormal(cache, &game.ImageCache)
    if err == nil {
        mouse.Mouse.SetImage(normalMouse)
    }

    return &Engine{
        LbxCache: cache,
        Coroutine: coroutine.MakeCoroutine(run),
        Game: game,
    }, nil
}

func main(){
    if len(os.Args) < 2 {
        fmt.Printf("Give a GAM file to load\n")
        return
    }

    reader, err := os.Open(os.Args[1])
    if err != nil {
        fmt.Printf("Error opening file: %v\n", err)
        return
    }
    defer reader.Close()

    saveGame, err := load.LoadSaveGame(reader)
    if err != nil {
        fmt.Printf("Error loading save game: %v\n", err)
        return
    }

    monitorWidth, _ := ebiten.Monitor().Size()
    size := monitorWidth / 390

    ebiten.SetWindowSize(data.ScreenWidth * size, data.ScreenHeight * size)
    ebiten.SetWindowTitle("new screen")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    audio.Initialize()
    mouse.Initialize()

    engine, err := NewEngine(saveGame)
    if err != nil {
        fmt.Printf("Error: unable to load engine: %v", err)
        return
    }

    err = ebiten.RunGame(engine)
    if err != nil {
        fmt.Printf("Error: %v", err)
    }

    engine.Game.Shutdown()
}
