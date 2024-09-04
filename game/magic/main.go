package main

import (
    "log"
    _ "fmt"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    introlib "github.com/kazzmir/master-of-magic/game/magic/intro"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    // "github.com/kazzmir/master-of-magic/game/magic/units"
    // playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"

    "github.com/hajimehoshi/ebiten/v2"
    // "github.com/hajimehoshi/ebiten/v2/inpututil"
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

func runNewGame(yield coroutine.YieldFunc, game *MagicGame) {
    newGame := setup.MakeNewGameScreen(game.Cache)

    game.Drawer = func(screen *ebiten.Image) {
        newGame.Draw(screen)
    }

    for newGame.Update() == setup.NewGameStateRunning {
        yield()
    }
}

func runGame(yield coroutine.YieldFunc, game *MagicGame) error {
    runIntro(yield, game)
    runNewGame(yield, game)

    return ebiten.Termination
}

func NewMagicGame() (*MagicGame, error) {
    var game *MagicGame

    run := func(yield coroutine.YieldFunc) error {
        return runGame(yield, game)
    }

    cache := lbx.AutoCache()
    game = &MagicGame{
        Cache: cache,
        MainCoroutine: coroutine.MakeCoroutine(run),
        Drawer: nil,
    }

    /*
    err := game.NewGameScreen.Load(game.LbxCache)
    if err != nil {
        return nil, err
    }
    game.NewGameScreen.Activate()
    */

    /*
    err := game.NewWizardScreen.Load(game.LbxCache)
    if err != nil {
        return nil, err
    }
    game.NewWizardScreen.Activate()
    */

    /*
    wizard := setup.WizardCustom{
        Banner: data.BannerBlue,
    }

    game.Game = gamelib.MakeGame(game.LbxCache)
    game.Game.Plane = data.PlaneArcanus
    game.Game.Activate()

    player := game.Game.AddPlayer(wizard)

    player.AddUnit(playerlib.Unit{
        Unit: units.GreatDrake,
        Plane: data.PlaneArcanus,
        Banner: wizard.Banner,
        X: 5,
        Y: 5,
    })

    player.LiftFog(4, 5, 3)

    game.Game.DoNextTurn()
    */

    return game, nil
}

func (game *MagicGame) Update() error {

    if game.MainCoroutine.Run() != nil {
        return ebiten.Termination
    }

    /*
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    if game.NewGameScreen != nil && game.NewGameScreen.IsActive() {
        switch game.NewGameScreen.Update() {
            case setup.NewGameStateRunning:
            case setup.NewGameStateOk:
                game.NewGameScreen.Deactivate()
                err := game.NewWizardScreen.Load(game.LbxCache)
                if err != nil {
                    return err
                }
                game.NewWizardScreen.Activate()
            case setup.NewGameStateCancel:
                return ebiten.Termination
        }
    }

    if game.NewWizardScreen != nil && game.NewWizardScreen.IsActive() {
        switch game.NewWizardScreen.Update() {
            case setup.NewWizardScreenStateFinished:
                game.NewWizardScreen.Deactivate()
                wizard := game.NewWizardScreen.CustomWizard
                log.Printf("Launch game with wizard: %+v\n", wizard)
                game.Game = gamelib.MakeGame(game.LbxCache)
                game.Game.AddPlayer(wizard)
                game.Game.Activate()
                game.NewWizardScreen = nil
        }
    }

    if game.Game != nil && game.Game.IsActive() {
        game.Game.Update()
    }
    */

    return nil
}

func (game *MagicGame) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return data.ScreenWidth, data.ScreenHeight
}

func (game *MagicGame) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})

    if game.Drawer != nil {
        game.Drawer(screen)
    }

    /*

    if game.NewGameScreen.IsActive() {
        game.NewGameScreen.Draw(screen)
    }

    if game.NewWizardScreen != nil && game.NewWizardScreen.IsActive() {
        game.NewWizardScreen.Draw(screen)
    }

    if game.Game != nil && game.Game.IsActive() {
        game.Game.Draw(screen)
    }
    */
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
