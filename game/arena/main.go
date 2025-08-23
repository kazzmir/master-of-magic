package main

import (
    "log"
    "errors"
    "math/rand/v2"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/console"

    "github.com/kazzmir/master-of-magic/game/arena/player"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/text/v2"

    "github.com/ebitenui/ebitenui"
    "github.com/ebitenui/ebitenui/widget"
    ui_image "github.com/ebitenui/ebitenui/image"
)

/*
 * Start with a small army (single unit?), fight a battle. If you win then you get money/score that you can use to buy more units and spells.
 *  1. start game, pick a wizard portrait, name, etc
 *  2. pick an army from a small set of units
 *  3. fight a battle against an equivalent foe
 *  4. use money to buy more units and spells
 *  5. repeat from step 3
 */

type GameMode int

const (
    GameModeUI GameMode = iota
    GameModeBattle
)

type EngineEvents interface {
}

type EventNewGame struct {
}

type Engine struct {
    GameMode GameMode
    Player *player.Player
    Cache *lbx.LbxCache

    Events chan EngineEvents

    CombatCoroutine *coroutine.Coroutine
    CombatScreen *combat.CombatScreen

    UI *ebitenui.UI
}

var CombatDoneErr = errors.New("combat done")

func (engine *Engine) MakeBattleFunc() coroutine.AcceptYieldFunc {
    defendingArmy := combat.Army {
        Player: engine.Player,
    }

    for _, unit := range engine.Player.Units {
        defendingArmy.AddUnit(unit)
    }

    defendingArmy.LayoutUnits(combat.TeamDefender)

    enemyPlayer := player.MakeAIPlayer(data.BannerRed)

    count := 0
    for count < engine.Player.Level + 1 {
        choice := units.AllUnits[rand.N(len(units.AllUnits))]
        if choice.Race == data.RaceHero || choice.Name == "Settlers" {
            continue
        }
        enemyPlayer.AddUnit(choice)
        count += 1
    }

    attackingArmy := combat.Army {
        Player: enemyPlayer,
    }

    for _, unit := range enemyPlayer.Units {
        attackingArmy.AddUnit(unit)
    }

    attackingArmy.LayoutUnits(combat.TeamAttacker)

    screen := combat.MakeCombatScreen(engine.Cache, &defendingArmy, &attackingArmy, engine.Player, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, data.MagicNone, 0, 0)
    engine.CombatScreen = screen

    return func(yield coroutine.YieldFunc) error {
        for screen.Update(yield) == combat.CombatStateRunning {
            yield()
        }

        return CombatDoneErr
    }
}

func (engine *Engine) Update() error {
    keys := inpututil.AppendJustPressedKeys(nil)

    for _, key := range keys {
        switch key {
            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                return ebiten.Termination
        }
    }

    inputmanager.Update()

    switch engine.GameMode {
        case GameModeUI:
            engine.UI.Update()

            select {
                case event := <-engine.Events:
                    switch event.(type) {
                        case *EventNewGame:
                            engine.GameMode = GameModeBattle
                            engine.CombatCoroutine = coroutine.MakeCoroutine(engine.MakeBattleFunc())
                    }
                default:
            }
        case GameModeBattle:
            err := engine.CombatCoroutine.Run()
            if errors.Is(err, CombatDoneErr) {
                engine.CombatCoroutine = nil
                engine.CombatScreen = nil
                engine.GameMode = GameModeUI

                engine.Player.Level += 1
            }
    }

    return nil
}

func (engine *Engine) DrawUI(screen *ebiten.Image) {
    engine.UI.Draw(screen)
}

func (engine *Engine) DrawBattle(screen *ebiten.Image) {
    engine.CombatScreen.Draw(screen)
    mouse.Mouse.Draw(screen)
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    switch engine.GameMode {
        case GameModeUI:
            engine.DrawUI(screen)
        case GameModeBattle:
            engine.DrawBattle(screen)
    }
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (int, int) {
    switch engine.GameMode {
        case GameModeUI: return outsideWidth, outsideHeight
        case GameModeBattle: return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
    }

    return outsideWidth, outsideHeight
}

func makeButtonImage(baseImage *ui_image.NineSlice) *widget.ButtonImage {
    return &widget.ButtonImage{
        Idle: baseImage,
        Hover: baseImage,
        Pressed: baseImage,
        Disabled: baseImage,
    }
}

func solidImage(r uint8, g uint8, b uint8) *ui_image.NineSlice {
    return ui_image.NewNineSliceColor(color.NRGBA{R: r, G: g, B: b, A: 255})
}

func (engine *Engine) MakeUI() (*ebitenui.UI, error) {
    font, err := console.LoadFont()
    if err != nil {
        return nil, err
    }

    face := text.GoTextFace{
        Source: font,
        Size: 18,
    }

    rootContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(4),
            widget.RowLayoutOpts.Padding(widget.Insets{Top: 4, Bottom: 4, Left: 4, Right: 4}),
        )),
        widget.ContainerOpts.BackgroundImage(ui_image.NewNineSliceColor(color.NRGBA{R: 32, G: 32, B: 32, A: 255})),
    )

    newGameButton := widget.NewButton(
        widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Image(makeButtonImage(ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 32, B: 32, A: 255}))),
        widget.ButtonOpts.Text("New Game", &face, &widget.ButtonTextColor{
            Idle: color.White,
            Hover: color.White,
            Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
        widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
            select {
                case engine.Events <- &EventNewGame{}:
                default:
            }
        }),
    )

    rootContainer.AddChild(newGameButton)

    armyInfo := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(2),
            widget.RowLayoutOpts.Padding(widget.Insets{Top: 2, Bottom: 2, Left: 2, Right: 2}),
        )),
        widget.ContainerOpts.BackgroundImage(ui_image.NewNineSliceColor(color.NRGBA{R: 48, G: 48, B: 48, A: 255})),
    )

    armyInfo.AddChild(widget.NewText(
        widget.TextOpts.Text("Army Info", &face, color.White),
    ))

    unitList := widget.NewList(
        widget.ListOpts.EntryFontFace(&face),
        widget.ListOpts.SliderOpts(
            widget.SliderOpts.Images(
                &widget.SliderTrackImage{
                    Idle: solidImage(64, 64, 64),
                    Hover: solidImage(96, 96, 96),
                },
                makeButtonImage(solidImage(192, 192, 192)),
            ),
        ),
        widget.ListOpts.HideHorizontalSlider(),
        widget.ListOpts.ContainerOpts(
            widget.ContainerOpts.WidgetOpts(
                widget.WidgetOpts.MinSize(0, 200),
            ),
        ),
        widget.ListOpts.EntryLabelFunc(
            func (e any) string {
                unit := e.(units.StackUnit)
                return unit.GetName()
            },
        ),
        widget.ListOpts.EntrySelectedHandler(func (args *widget.ListEntrySelectedEventArgs) {
            log.Printf("Selected unit: %v", args.Entry)
        }),
        widget.ListOpts.EntryColor(&widget.ListEntryColor{
            Selected: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            Unselected: color.NRGBA{R: 128, G: 128, B: 128, A: 255},
        }),
        widget.ListOpts.ScrollContainerOpts(
            widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
                Idle: solidImage(64, 64, 64),
                Disabled: solidImage(32, 32, 32),
                Mask: solidImage(32, 32, 32),
            }),
        ),
    )

    for _, unit := range engine.Player.Units {
        unitList.AddEntry(unit)
    }

    armyInfo.AddChild(unitList)

    rootContainer.AddChild(armyInfo)

    ui := &ebitenui.UI{
        Container: rootContainer,
    }

    return ui, nil
}

func MakeEngine(cache *lbx.LbxCache) *Engine {
    playerObj := player.MakePlayer(data.BannerGreen)

    playerObj.AddUnit(units.LizardSwordsmen)

    engine := Engine{
        GameMode: GameModeUI,
        Player: playerObj,
        Cache: cache,
        Events: make(chan EngineEvents, 10),
    }

    var err error
    engine.UI, err = engine.MakeUI()
    if err != nil {
        log.Printf("Error creating UI: %v", err)
    }
    return &engine
}

func main() {
    log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)

    cache := lbx.AutoCache()

    audio.Initialize()
    mouse.Initialize()

    ebiten.SetWindowSize(1200, 900)
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    engine := MakeEngine(cache)

    err := ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
