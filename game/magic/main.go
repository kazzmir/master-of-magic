package main

import (
    "log"
    _ "fmt"
    // "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    introlib "github.com/kazzmir/master-of-magic/game/magic/intro"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/mainview"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

func stretchImage(screen *ebiten.Image, sprite *ebiten.Image){
    var options ebiten.DrawImageOptions
    options.GeoM.Scale(float64(data.ScreenWidth) / float64(sprite.Bounds().Dx()), float64(data.ScreenHeight) / float64(sprite.Bounds().Dy()))
    screen.DrawImage(sprite, &options)
}

type DrawFunc func(*ebiten.Image)

type MagicGame struct {
    Cache *lbx.LbxCache

    MainCoroutine *coroutine.Coroutine
    Drawer DrawFunc

    NewGameScreen *setup.NewGameScreen
    NewWizardScreen *setup.NewWizardScreen

    Game *gamelib.Game
}

func runIntro(yield coroutine.YieldFunc, game *MagicGame) {
    intro, err := introlib.MakeIntro(game.Cache, introlib.DefaultAnimationSpeed)
    if err != nil {
        log.Printf("Unable to run intro: %v", err)
        return
    }

    game.Drawer = func(screen *ebiten.Image) {
        intro.Draw(screen)
    }

    for intro.Update() == introlib.IntroStateRunning {
        yield()

        if ebiten.IsKeyPressed(ebiten.KeySpace) {
            return
        }
    }
}

func runNewGame(yield coroutine.YieldFunc, game *MagicGame) setup.NewGameSettings {
    newGame := setup.MakeNewGameScreen(game.Cache)

    game.Drawer = func(screen *ebiten.Image) {
        newGame.Draw(screen)
    }

    for newGame.Update() == setup.NewGameStateRunning {
        yield()
    }

    return newGame.Settings
}

func runNewWizard(yield coroutine.YieldFunc, game *MagicGame) setup.WizardCustom {
    newWizard := setup.MakeNewWizardScreen(game.Cache)

    game.Drawer = func(screen *ebiten.Image) {
        newWizard.Draw(screen)
    }

    for newWizard.Update() != setup.NewWizardScreenStateFinished {
        yield()
    }

    return newWizard.CustomWizard
}

func runMainMenu(yield coroutine.YieldFunc, game *MagicGame) mainview.MainScreenState {
    menu := mainview.MakeMainScreen(game.Cache)

    game.Drawer = func(screen *ebiten.Image) {
        menu.Draw(screen)
    }

    for menu.Update() == mainview.MainScreenStateRunning {
        yield()
    }

    return menu.State
}

func runGameInstance(yield coroutine.YieldFunc, magic *MagicGame, settings setup.NewGameSettings, wizard setup.WizardCustom) {
    game := gamelib.MakeGame(magic.Cache)
    game.Plane = data.PlaneArcanus
    game.Activate()

    magic.Drawer = func(screen *ebiten.Image) {
        game.Draw(screen)
    }

    player := game.AddPlayer(wizard)

    player.AddUnit(playerlib.Unit{
        Unit: units.GreatDrake,
        Plane: data.PlaneArcanus,
        Banner: wizard.Banner,
        X: 5,
        Y: 5,
    })

    player.LiftFog(4, 5, 3)

    game.DoNextTurn()

    for game.Update() != gamelib.GameStateQuit {
        keys := make([]ebiten.Key, 0)
        keys = inpututil.AppendJustPressedKeys(keys)

        for _, key := range keys {
            if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
                // return ebiten.Termination
                return
            }
        }

        yield()
    }
}

func runGame(yield coroutine.YieldFunc, game *MagicGame) error {
    var cache *lbx.LbxCache
    for cache == nil {
        cache = lbx.AutoCache()
        if cache == nil {
            yield()
        }
    }

    game.Cache = cache

    runIntro(yield, game)
    state := runMainMenu(yield, game)
    switch state {
        case mainview.MainScreenStateQuit: return ebiten.Termination
        case mainview.MainScreenStateNewGame:
            // yield so that clicks from the menu don't bleed into the next part
            yield()
            settings := runNewGame(yield, game)
            yield()
            wizard := runNewWizard(yield, game)
            yield()
            runGameInstance(yield, game, settings, wizard)
    }

    return ebiten.Termination
}

func NewMagicGame() (*MagicGame, error) {
    var game *MagicGame

    run := func(yield coroutine.YieldFunc) error {
        return runGame(yield, game)
    }

    game = &MagicGame{
        MainCoroutine: coroutine.MakeCoroutine(run),
        Drawer: nil,
    }

    return game, nil
}

func (game *MagicGame) Update() error {

    if game.MainCoroutine.Run() != nil {
        return ebiten.Termination
    }

    return nil
}

func (game *MagicGame) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return data.ScreenWidth, data.ScreenHeight
}

func (game *MagicGame) Draw(screen *ebiten.Image) {
    // screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})

    if game.Drawer != nil {
        game.Drawer(screen)
    }
}

func main() {
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(data.ScreenWidth * 5, data.ScreenHeight * 5)
    ebiten.SetWindowTitle("magic")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    audio.Initialize()

    game, err := NewMagicGame()
    
    if err != nil {
        log.Printf("Error: unable to load game: %v", err)
        return
    }

    err = ebiten.RunGame(game)
    if err != nil {
        log.Printf("Error: %v", err)
    }

}
