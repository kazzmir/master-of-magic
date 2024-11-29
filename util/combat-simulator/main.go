package main

import (
    "embed"
    "log"
    "fmt"
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
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/text/v2"

    "github.com/ebitenui/ebitenui"
    // "github.com/ebitenui/ebitenui/input"
    "github.com/ebitenui/ebitenui/widget"
    ui_image "github.com/ebitenui/ebitenui/image"
)

//go:embed assets/*
var assets embed.FS

const EngineWidth = 1024
const EngineHeight = 768

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
    Counter uint64
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
    engine.Counter += 1

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
            return outsideWidth, outsideHeight
        case EngineModeCombat:
            return data.ScreenWidth, data.ScreenHeight
    }

    return 0, 0
}

func (engine *Engine) EnterCombat(defenderUnits []units.Unit, attackerUnits []units.Unit) {
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
    for _, unit := range defenderUnits {
        made := units.MakeOverworldUnitFromUnit(unit, 1, 1, data.PlaneArcanus, cpuPlayer.Wizard.Banner, cpuPlayer.MakeExperienceInfo())
        defendingArmy.AddUnit(made)
    }
    // warlock := units.MakeOverworldUnitFromUnit(units.Warlocks, 1, 1, data.PlaneArcanus, cpuPlayer.Wizard.Banner, cpuPlayer.MakeExperienceInfo())
    // defendingArmy.AddUnit(warlock)

    defendingArmy.LayoutUnits(combat.TeamDefender)

    attackingArmy := combat.Army{
        Player: humanPlayer,
    }

    /*
    for range 2 {
        attackingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.HighMenBowmen, 1, 1, data.PlaneArcanus, humanPlayer.Wizard.Banner, humanPlayer.MakeExperienceInfo()))
    }
    */

    for _, unit := range attackerUnits {
        made := units.MakeOverworldUnitFromUnit(unit, 1, 1, data.PlaneArcanus, humanPlayer.Wizard.Banner, humanPlayer.MakeExperienceInfo())
        attackingArmy.AddUnit(made)
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

    backgroundImageRaw, _, err := ebitenutil.NewImageFromFileSystem(assets, "assets/box.png")
    if err != nil {
        log.Printf("Could not load box.png: %v", err)
        return nil
    }

    backgroundImageNine := ui_image.NewNineSliceSimple(backgroundImageRaw, 1, 39)

    _ = backgroundImageNine

    backgroundImage := ui_image.NewNineSliceColor(color.NRGBA{R: 32, G: 32, B: 32, A: 255})

    rootContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(10),
            widget.RowLayoutOpts.Padding(widget.Insets{Top: 10, Left: 10, Right: 10}),
        )),
        widget.ContainerOpts.BackgroundImage(backgroundImage),
        // widget.ContainerOpts.BackgroundImage(backgroundImageNine),
    )

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

    fakeImage := ui_image.NewNineSliceColor(color.NRGBA{R: 32, G: 32, B: 32, A: 255})
    buttonImage := ui_image.NewNineSliceColor(color.NRGBA{R: 0, G: 0, B: 128, A: 255})
    // buttonImageHover := ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 64, B: 64, A: 255})
    // buttonImageIdle := ui_image.NewNineSliceColor(color.NRGBA{R: 80, G: 80, B: 0, A: 255})

    /*
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
    */

    type UnitItem struct {
        Race data.Race
        Unit units.Unit
    }

    var armyList *widget.List
    var armyCount *widget.Text

    var raceTabs []*widget.TabBookTab

    allRaces := append(append(data.ArcanianRaces(), data.MyrranRaces()...), data.RaceFantastic)

    for _, race := range allRaces {
        tab := widget.NewTabBookTab(
            race.String(),
            widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
        )

        clickTimer := make(map[string]uint64)

        unitList := widget.NewList(
            widget.ListOpts.EntryFontFace(face),
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

            widget.ListOpts.EntryLabelFunc(
                func (e any) string {
                    item := e.(*UnitItem)
                    return fmt.Sprintf("%v %v", item.Race, item.Unit.Name)
                },
            ),

            widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
                entry := args.Entry.(*UnitItem)

                lastTime, ok := clickTimer[entry.Unit.Name]
                // log.Printf("Entry %v lastTime %v counter %v ok %v", entry, lastTime, engine.Counter, ok)
                if ok && engine.Counter - lastTime < 30 {
                    // log.Printf("  adding %v to defending army", entry)
                    newItem := *entry
                    armyList.AddEntry(&newItem)
                    armyCount.Label = fmt.Sprintf("%v", len(armyList.Entries()))
                    clickTimer[entry.Unit.Name] = engine.Counter + 30
                } else {
                    clickTimer[entry.Unit.Name] = engine.Counter
                }
            }),

            widget.ListOpts.EntryColor(&widget.ListEntryColor{
                Selected: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                Unselected: color.NRGBA{R: 128, G: 128, B: 128, A: 255},
            }),

            widget.ListOpts.ScrollContainerOpts(
                widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
                    Idle: fakeImage,
                    Disabled: fakeImage,
                    Mask: fakeImage,
                }),
            ),

            widget.ListOpts.AllowReselect(),
        )

        tab.AddChild(unitList)

        for _, unit := range units.UnitsByRace(race) {
            unitList.AddEntry(&UnitItem{
                Race: race,
                Unit: unit,
            })
        }

        tab.AddChild(widget.NewButton(
            widget.ButtonOpts.Image(&widget.ButtonImage{
                Idle: buttonImage,
                Hover: buttonImage,
                Pressed: buttonImage,
            }),
            widget.ButtonOpts.Text("Add", face, &widget.ButtonTextColor{
                Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
            }),
            widget.ButtonOpts.PressedHandler(func (args *widget.ButtonPressedEventArgs) {
                entry := unitList.SelectedEntry()
                if entry != nil {
                    entry := entry.(*UnitItem)
                    newItem := *entry
                    armyList.AddEntry(&newItem)
                    armyCount.Label = fmt.Sprintf("%v", len(armyList.Entries()))
                }
            }),
        ))

        raceTabs = append(raceTabs, tab)
    }

    defendingArmyCount := widget.NewText(widget.TextOpts.Text("0", face, color.NRGBA{R: 255, G: 255, B: 255, A: 255}))

    makeArmyList := func() *widget.List {
        return widget.NewList(
            widget.ListOpts.EntryFontFace(face),

            widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                MaxHeight: 300,
            }))),

            widget.ListOpts.SliderOpts(
                widget.SliderOpts.Images(&widget.SliderTrackImage{
                    Idle: fakeImage,
                    Hover: fakeImage,
                }, &widget.ButtonImage{
                    Idle: ui_image.NewNineSliceColor(color.NRGBA{R: 128, G: 0, B: 0, A: 255}),
                    Hover: ui_image.NewNineSliceColor(color.NRGBA{R: 160, G: 0, B: 0, A: 255}),
                    Pressed: ui_image.NewNineSliceColor(color.NRGBA{R: 160, G: 0, B: 0, A: 255}),
                }),
            ),

            widget.ListOpts.EntryLabelFunc(
                func (e any) string {
                    item := e.(*UnitItem)
                    return fmt.Sprintf("%v %v", item.Race, item.Unit.Name)
                },
            ),

            widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
                /*
                entry := args.Entry.(*ListItem)
                fmt.Println("Entry Selected: ", entry.Name)
                */
            }),

            widget.ListOpts.EntryColor(&widget.ListEntryColor{
                Selected: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
                Unselected: color.NRGBA{R: 0, G: 255, B: 0, A: 255},
            }),

            widget.ListOpts.ScrollContainerOpts(
                widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
                    Idle: ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 64, B: 64, A: 255}),
                    Disabled: fakeImage,
                    Mask: fakeImage,
                }),
            ),
        )
    }

    defendingArmyList := makeArmyList()
    defendingArmyName := widget.NewText(
        widget.TextOpts.Text("Defending Army", face, color.NRGBA{R: 255, G: 0, B: 0, A: 255}),
    )

    attackingArmyCount := widget.NewText(widget.TextOpts.Text("0", face, color.NRGBA{R: 255, G: 255, B: 255, A: 255}))
    attackingArmyList := makeArmyList()
    attackingArmyName := widget.NewText(
        widget.TextOpts.Text("Attacking Army", face, color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
    )

    armyList = defendingArmyList
    armyCount = defendingArmyCount

    greyish := color.NRGBA{R: 128, G: 128, B: 128, A: 255}

    unitsTabs := widget.NewTabBook(
        widget.TabBookOpts.TabButtonImage(&widget.ButtonImage{
            Idle: backgroundImage,
            Hover: ui_image.NewNineSliceColor(color.NRGBA{R: 128, G: 128, B: 128, A: 255}),
            Pressed: ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 64, B: 64, A: 255}),
        }),
        widget.TabBookOpts.TabButtonSpacing(3),
        widget.TabBookOpts.TabButtonText(face, &widget.ButtonTextColor{Idle: color.White, Disabled: greyish}),
        widget.TabBookOpts.Tabs(raceTabs...),
    )

    rootContainer.AddChild(unitsTabs)

    var armyButtons []*widget.Button
    armyButtons = append(armyButtons, widget.NewButton(
        widget.ButtonOpts.Image(&widget.ButtonImage{
            Idle: buttonImage,
            Hover: buttonImage,
            Pressed: buttonImage,
        }),
        widget.ButtonOpts.Text("Defending Army", face, &widget.ButtonTextColor{
            Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
            Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
        widget.ButtonOpts.PressedHandler(func (args *widget.ButtonPressedEventArgs) {
            armyList = defendingArmyList
            armyCount = defendingArmyCount
            defendingArmyName.Color = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
            attackingArmyName.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
        }),
    ))

    armyButtons = append(armyButtons, widget.NewButton(
        widget.ButtonOpts.Image(&widget.ButtonImage{
            Idle: buttonImage,
            Hover: buttonImage,
            Pressed: buttonImage,
        }),
        widget.ButtonOpts.Text("Attacking Army", face, &widget.ButtonTextColor{
            Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
            Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
        widget.ButtonOpts.PressedHandler(func (args *widget.ButtonPressedEventArgs) {
            armyList = attackingArmyList
            armyCount = attackingArmyCount
            attackingArmyName.Color = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
            defendingArmyName.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
        }),
    ))

    var armyRadioElements []widget.RadioGroupElement
    for _, button := range armyButtons {
        armyRadioElements = append(armyRadioElements, button)
    }

    widget.NewRadioGroup(widget.RadioGroupOpts.Elements(armyRadioElements...))

    makeRow := func(spacing int, children ...widget.PreferredSizeLocateableWidget) *widget.Container {
        container := widget.NewContainer(
            widget.ContainerOpts.Layout(widget.NewRowLayout(
                widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
                widget.RowLayoutOpts.Spacing(spacing),
            )),
        )

        for _, child := range children {
            container.AddChild(child)
        }

        return container
    }

    rootContainer.AddChild(makeRow(10, armyButtons[0], armyButtons[1]))

    armyContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
            widget.RowLayoutOpts.Spacing(30),
        )),
    )

    defendingArmyContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
        )),
    )

    defendingArmyContainer.AddChild(makeRow(4, defendingArmyName, defendingArmyCount))

    defendingArmyContainer.AddChild(defendingArmyList)
    defendingArmyContainer.AddChild(widget.NewButton(
        widget.ButtonOpts.Image(&widget.ButtonImage{
            Idle: buttonImage,
            Hover: buttonImage,
            Pressed: buttonImage,
        }),
        widget.ButtonOpts.Text("Clear Defending Army", face, &widget.ButtonTextColor{
            Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
            Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
        widget.ButtonOpts.PressedHandler(func (args *widget.ButtonPressedEventArgs) {
            defendingArmyList.SetEntries(nil)
            defendingArmyCount.Label = "0"
        }),
    ))

    armyContainer.AddChild(defendingArmyContainer)

    attackingArmyContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
        )),
    )

    attackingArmyContainer.AddChild(makeRow(4, attackingArmyName, attackingArmyCount))

    attackingArmyContainer.AddChild(attackingArmyList)
    attackingArmyContainer.AddChild(widget.NewButton(
        widget.ButtonOpts.Image(&widget.ButtonImage{
            Idle: buttonImage,
            Hover: buttonImage,
            Pressed: buttonImage,
        }),
        widget.ButtonOpts.Text("Clear Attacking Army", face, &widget.ButtonTextColor{
            Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
            Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
        widget.ButtonOpts.PressedHandler(func (args *widget.ButtonPressedEventArgs) {
            attackingArmyList.SetEntries(nil)
            attackingArmyCount.Label = "0"
        }),
    ))

    armyContainer.AddChild(attackingArmyContainer)

    rootContainer.AddChild(armyContainer)

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
            var defenders []units.Unit

            for _, entry := range defendingArmyList.Entries() {
                defenders = append(defenders, entry.(*UnitItem).Unit)
            }

            var attackers []units.Unit

            for _, entry := range attackingArmyList.Entries() {
                attackers = append(attackers, entry.(*UnitItem).Unit)
            }

            if len(defenders) > 0 && len(attackers) > 0 {
                engine.EnterCombat(defenders, attackers)
            }
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
    ebiten.SetWindowSize(1200, 768)
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    err := ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
