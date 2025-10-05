package main

import (
    "os"
    "fmt"
    "flag"
    "bytes"
    "bufio"
    "io"
    "log"

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

func NewEngine(saveGame *load.SaveGame, admin bool) (*Engine, error) {
    cache := lbx.AutoCache()

    game := saveGame.Convert(cache)
    game.DoNextTurn()

    if admin {
        player := game.Players[0]
        player.Admin = true
        player.LiftFogAll(data.PlaneArcanus)
        player.LiftFogAll(data.PlaneMyrror)
        for _, other := range game.Players {
            if player != other {
                player.AwarePlayer(other)
            }
        }
    }

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

func compareSaveFile(path string) error {
    reader, err := os.Open(path)
    if err != nil {
        return err
    }
    defer reader.Close()

    saveGame, err := load.LoadSaveGame(reader)
    if err != nil {
        return fmt.Errorf("Unable to load saved game: %v", err)
    }

    reader2, err := os.Open(path)
    if err != nil {
        return err
    }
    defer reader2.Close()

    var outputBytes bytes.Buffer
    err = load.WriteSaveGame(saveGame, &outputBytes)
    if err != nil {
        return err
    }

    var inputBytes bytes.Buffer
    inputReader := bufio.NewReader(reader2)
    io.Copy(&inputBytes, inputReader)

    maxLength := min(outputBytes.Len(), inputBytes.Len())

    // log.Printf("Output: %v", outputBytes.Bytes()[:50])
    // log.Printf("Input:  %v", inputBytes.Bytes()[:50])

    for i := range maxLength {
        if outputBytes.Bytes()[i] != inputBytes.Bytes()[i] {
            return fmt.Errorf("save file content mismatch at byte %d: actual %d vs input %d", i, outputBytes.Bytes()[i], inputBytes.Bytes()[i])
        }
    }

    if outputBytes.Len() != inputBytes.Len() {
        return fmt.Errorf("save file size mismatch: %d vs %d", outputBytes.Len(), inputBytes.Len())
    }

    /*
    if !bytes.Equal(outputBytes.Bytes(), inputBytes.Bytes()) {
        return fmt.Errorf("save file content mismatch")
    }
    */

    return nil
}

func main(){

    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    var admin bool

    flag.BoolVar(&admin, "admin", false, "Make the player an admin (optional)")
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %v [options] filename\n\n", os.Args[0])
        fmt.Fprintln(os.Stderr, "Options:")
        flag.PrintDefaults()
        fmt.Fprintln(os.Stderr, "\nExample:")
        fmt.Fprintln(os.Stderr, "  ", os.Args[0], "--admin SAVE1.GAM")
    }

    flag.Parse()

    positionalArgs := flag.Args()

    if len(positionalArgs) < 1 {
        flag.Usage()
        return
    }

    err := compareSaveFile(positionalArgs[0])
    if err != nil {
        log.Printf("Error comparing save file: %v", err)
        return
    }

    reader, err := os.Open(positionalArgs[0])
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

    engine, err := NewEngine(saveGame, admin)
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
