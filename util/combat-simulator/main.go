package main

import (
    "log"
    "image/color"
    "errors"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/util/common"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/text/v2"

    "github.com/ebitenui/ebitenui"
    // "github.com/ebitenui/ebitenui/input"
    "github.com/ebitenui/ebitenui/widget"
    ui_image "github.com/ebitenui/ebitenui/image"
)

const EngineWidth = 800
const EngineHeight = 600

type EngineMode int
const (
    EngineModeMenu EngineMode = iota
    EngineModeCombat
)

type HoverData struct {
    OnHover func()
    OnUnhover func()
}

type Engine struct {
    Cache *lbx.LbxCache
    Mode EngineMode
    UI *ebitenui.UI
    Combat *combat.CombatScreen
    Coroutine *coroutine.Coroutine
}

func MakeEngine(cache *lbx.LbxCache) *Engine {
    engine := Engine{
        Cache: cache,
    }

    engine.UI = engine.MakeUI()

    return &engine
}

var CombatDoneErr = errors.New("combat done")

func (engine *Engine) Update() error {
    keys := inpututil.AppendJustPressedKeys(nil)

    for _, key := range keys {
        switch key {
            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                return ebiten.Termination
        }
    }

    switch engine.Mode {
        case EngineModeMenu:
            engine.UI.Update()
        case EngineModeCombat:
            inputmanager.Update()
            err := engine.Coroutine.Run()
            if errors.Is(err, CombatDoneErr) {
                engine.Combat = nil
                engine.Coroutine = nil
                engine.Mode = EngineModeMenu
            }
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    switch engine.Mode {
        case EngineModeMenu:
            engine.UI.Draw(screen)
        case EngineModeCombat:
            engine.Combat.Draw(screen)
            mouse.Mouse.Draw(screen)
    }
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    switch engine.Mode {
        case EngineModeMenu:
            return EngineWidth, EngineHeight
        case EngineModeCombat:
            return data.ScreenWidth, data.ScreenHeight
    }

    return 0, 0
}

func (engine *Engine) EnterCombat() {
    engine.Mode = EngineModeCombat

    cpuPlayer := playerlib.MakePlayer(setup.WizardCustom{
        Name: "CPU",
        Banner: data.BannerRed,
    }, false, nil, nil)

    humanPlayer := playerlib.MakePlayer(setup.WizardCustom{
        Name: "Human",
        Banner: data.BannerGreen,
    }, true, nil, nil)

    defendingArmy := combat.Army{
        Player: cpuPlayer,
    }
    warlock := units.MakeOverworldUnitFromUnit(units.Warlocks, 1, 1, data.PlaneArcanus, cpuPlayer.Wizard.Banner, cpuPlayer.MakeExperienceInfo())
    defendingArmy.AddUnit(warlock)

    defendingArmy.LayoutUnits(combat.TeamDefender)

    attackingArmy := combat.Army{
        Player: humanPlayer,
    }

    for range 2 {
        attackingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, 1, 1, data.PlaneArcanus, humanPlayer.Wizard.Banner, humanPlayer.MakeExperienceInfo()))
    }

    attackingArmy.LayoutUnits(combat.TeamAttacker)

    combatScreen := combat.MakeCombatScreen(engine.Cache, &defendingArmy, &attackingArmy, humanPlayer, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{})
    engine.Combat = combatScreen

    run := func(yield coroutine.YieldFunc) error {
        for combatScreen.Update(yield) == combat.CombatStateRunning {
            yield()
        }

        return CombatDoneErr
    }

    engine.Coroutine = coroutine.MakeCoroutine(run)
}

func (engine *Engine) MakeUI() *ebitenui.UI {
    face, _ := loadFont(18)

    rootContainer := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))))

    var label1 *widget.Text

    label1 = widget.NewText(
        widget.TextOpts.Text("Master of Magic Combat Simulator", face, color.White),
        widget.TextOpts.WidgetOpts(
            widget.WidgetOpts.CursorEnterHandler(func(args *widget.WidgetCursorEnterEventArgs) {
                label1.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
            }),
            widget.WidgetOpts.CursorExitHandler(func(args *widget.WidgetCursorExitEventArgs) {
                label1.Color = color.White
            }),
        ),
    )

    rootContainer.AddChild(label1)

    fakeImage := ui_image.NewNineSliceColor(color.NRGBA{R: 255, G: 0, B: 0, A: 255})
    buttonImage := ui_image.NewNineSliceColor(color.NRGBA{R: 150, G: 150, B: 0, A: 255})

    unitList1 := widget.NewListComboButton(
        widget.ListComboButtonOpts.SelectComboButtonOpts(
            widget.SelectComboButtonOpts.ComboButtonOpts(
                widget.ComboButtonOpts.MaxContentHeight(200),
                widget.ComboButtonOpts.ButtonOpts(
                    widget.ButtonOpts.Image(&widget.ButtonImage{
                        Idle: buttonImage,
                        Hover: buttonImage,
                        Pressed: buttonImage,
                    }),
                    widget.ButtonOpts.Text("Select Unit", face, &widget.ButtonTextColor{
                        Idle: color.White,
                        Disabled: color.White,
                    }),
                ),
            ),
        ),
        widget.ListComboButtonOpts.EntryLabelFunc(
            func (e any) string {
                return "Button " + e.(string)
            },
            func (e any) string {
                return "List " + e.(string)
            },
        ),
        widget.ListComboButtonOpts.ListOpts(
            widget.ListOpts.EntryFontFace(face),
            widget.ListOpts.Entries([]any{"x", "y", "z", "a", "b"}),
            widget.ListOpts.EntryColor(&widget.ListEntryColor{
                Selected: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
                Unselected: color.NRGBA{R: 0, G: 255, B: 0, A: 255},
            }),
            widget.ListOpts.SliderOpts(
                widget.SliderOpts.Images(&widget.SliderTrackImage{
                    Idle: fakeImage,
                    Hover: fakeImage,
                }, &widget.ButtonImage{
                    Idle: fakeImage,
                    Hover: fakeImage,
                    Pressed: fakeImage,
                }),
            ),
            widget.ListOpts.ScrollContainerOpts(
                widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
                    Idle: fakeImage,
                    Disabled: fakeImage,
                    Mask: fakeImage,
                }),
            ),
        ),
    )
    rootContainer.AddChild(unitList1)

    unitsHighMenTab := widget.NewTabBookTab(
        "High Men",
        widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
    )

    unitsHighMenTab.AddChild(widget.NewText(
        widget.TextOpts.Text("Bowmen", face, color.White),
    ))

    unitsKlackonTab := widget.NewTabBookTab(
        "Klackon",
        widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
    )

    unitsKlackonTab.AddChild(widget.NewText(
        widget.TextOpts.Text("Stag Beetle", face, color.White),
    ))

    unitsTabs := widget.NewTabBook(
        widget.TabBookOpts.TabButtonImage(&widget.ButtonImage{
            Idle: buttonImage,
            Hover: buttonImage,
            Pressed: buttonImage,
        }),
        widget.TabBookOpts.TabButtonText(face, &widget.ButtonTextColor{Idle: color.White, Disabled: color.White}),
        widget.TabBookOpts.Tabs(unitsHighMenTab, unitsKlackonTab),
    )

    rootContainer.AddChild(unitsTabs)

    rootContainer.AddChild(widget.NewButton(
        widget.ButtonOpts.Image(&widget.ButtonImage{
            Idle: buttonImage,
            Hover: buttonImage,
            Pressed: buttonImage,
        }),
        widget.ButtonOpts.Text("Enter Combat!", face, &widget.ButtonTextColor{
            Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
            Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
        }),
        widget.ButtonOpts.PressedHandler(func (args *widget.ButtonPressedEventArgs) {
            engine.EnterCombat()
        }),
    ))

    ui := ebitenui.UI{
        Container: rootContainer,
    }

    return &ui
}

func loadFont(size float64) (text.Face, error) {
    source, err := common.LoadFont()

    if err != nil {
        log.Fatal(err)
        return nil, err
    }

    return &text.GoTextFace{
        Source: source,
        Size:   size,
    }, nil
}

func main(){
    cache := lbx.AutoCache()

    audio.Initialize()
    mouse.Initialize()

    engine := MakeEngine(cache)
    ebiten.SetWindowSize(800, 600)
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    err := ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
