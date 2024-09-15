package main

import (
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    // "github.com/kazzmir/master-of-magic/game/magic/terrain"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    Game *gamelib.Game
    Coroutine *coroutine.Coroutine
}

func createScenario1(cache *lbx.LbxCache) *gamelib.Game {
    wizard := setup.WizardCustom{
        Name: "player",
        Banner: data.BannerBlue,
        Abilities: []setup.WizardAbility{
            setup.AbilityAlchemy,
            setup.AbilitySageMaster,
        },
        Books: []data.WizardBook{
            data.WizardBook{
                Magic: data.LifeMagic,
                Count: 3,
            },
            data.WizardBook{
                Magic: data.SorceryMagic,
                Count: 8,
            },
        },
    }

    game := gamelib.MakeGame(cache)

    game.Plane = data.PlaneArcanus

    player := game.AddPlayer(wizard)

    player.AddCity(citylib.City{
        Population: 6000,
        Name: "Test City",
        Plane: data.PlaneArcanus,
        ProducingBuilding: citylib.BuildingHousing,
        ProducingUnit: units.UnitNone,
        Race: data.RaceHighElf,
        Wall: false,
        X: 4,
        Y: 5,
    })

    player.Gold = 83
    player.Mana = 26

    // game.Map.Map.Terrain[3][6] = terrain.TileNatureForest.Index

    player.LiftFog(4, 5, 3)

    drake := player.AddUnit(playerlib.Unit{
        Unit: units.GreatDrake,
        Plane: data.PlaneArcanus,
        Banner: wizard.Banner,
        X: 5,
        Y: 5,
    })

    for i := 0; i < 5; i++ {
        fireElemental := player.AddUnit(playerlib.Unit{
            Unit: units.FireElemental,
            Plane: data.PlaneArcanus,
            Banner: wizard.Banner,
            X: 5,
            Y: 5,
        })
        _ = fireElemental
    }

    stack := player.FindStackByUnit(drake)
    player.SetSelectedStack(stack)

    player.LiftFog(5, 5, 2)

    enemy1 := game.AddPlayer(setup.WizardCustom{
        Name: "dingus",
        Banner: data.BannerRed,
    })

    enemy1.AddUnit(playerlib.Unit{
        Unit: units.Warlocks,
        Plane: data.PlaneArcanus,
        Banner: enemy1.Wizard.Banner,
        X: 6,
        Y: 6,
    })

    enemy1.AddUnit(playerlib.Unit{
        Unit: units.HighMenBowmen,
        Plane: data.PlaneArcanus,
        Banner: enemy1.Wizard.Banner,
        X: 6,
        Y: 6,
    })

    return game
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()

    game := createScenario1(cache)

    game.DoNextTurn()

    run := func(yield coroutine.YieldFunc) error {
        for game.Update(yield) != gamelib.GameStateQuit {
            yield()
        }

        return ebiten.Termination
    }

    return &Engine{
        LbxCache: cache,
        Coroutine: coroutine.MakeCoroutine(run),
        Game: game,
    }, nil
}

func (engine *Engine) Update() error {

    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    /*
    switch engine.Game.Update() {
        case gamelib.GameStateRunning:
    }
    */
    if engine.Coroutine.Run() != nil {
        return ebiten.Termination
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    engine.Game.Draw(screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return data.ScreenWidth, data.ScreenHeight
}

func main(){

    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    monitorWidth, _ := ebiten.Monitor().Size()

    size := monitorWidth / 390

    ebiten.SetWindowSize(data.ScreenWidth * size, data.ScreenHeight * size)
    ebiten.SetWindowTitle("new screen")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    engine, err := NewEngine()

    audio.Initialize()

    if err != nil {
        log.Printf("Error: unable to load engine: %v", err)
        return
    }

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }

}
