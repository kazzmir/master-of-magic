package main

import (
    "log"
    "errors"
    "math"
    "math/rand/v2"
    "image/color"
    "image"
    "fmt"

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
    "github.com/kazzmir/master-of-magic/game/magic/util"

    "github.com/kazzmir/master-of-magic/game/arena/player"
    "github.com/kazzmir/master-of-magic/game/arena/ui"

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
    CurrentBattleReward uint64

    UI *ebitenui.UI
    UIUpdates *UIEventUpdate
}

var CombatDoneErr = errors.New("combat done")

func getValidChoices(budget uint64) []*units.Unit {
    var choices []*units.Unit

    for i := range units.AllUnits {
        choice := &units.AllUnits[i]
        if choice.Race == data.RaceHero || choice.Name == "Settlers" {
            continue
        }
        if getUnitCost(choice) > budget {
            continue
        }
        if choice.Sailing {
            continue
        }
        if choice.HasAbility(data.AbilityTransport) {
            continue
        }

        choices = append(choices, choice)
    }

    return choices
}

func (engine *Engine) MakeBattleFunc() coroutine.AcceptYieldFunc {
    defendingArmy := combat.Army {
        Player: engine.Player,
    }

    for _, unit := range engine.Player.Units {
        defendingArmy.AddUnit(unit)
    }

    defendingArmy.LayoutUnits(combat.TeamDefender)

    enemyPlayer := player.MakeAIPlayer(data.BannerRed)

    budget := uint64(100 * math.Pow(1.8, float64(engine.Player.Level)))
    engine.CurrentBattleReward = 0

    for budget > 0 {
        choices := getValidChoices(budget)

        if len(choices) == 0 {
            break
        }

        choice := choices[rand.N(len(choices))]

        enemyPlayer.AddUnit(*choice)
        unitCost := getUnitCost(choice)
        budget -= unitCost
        engine.CurrentBattleReward += unitCost
    }

    /*
    for range engine.Player.Level {
        choice := units.AllUnits[rand.N(len(units.AllUnits))]
        if choice.Race == data.RaceHero || choice.Name == "Settlers" {
            continue
        }
        enemyPlayer.AddUnit(choice)
    }
    */

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
            engine.UIUpdates.Update()

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
                engine.Player.Money += engine.CurrentBattleReward

                var aliveUnits []units.StackUnit
                for _, unit := range engine.Player.Units {
                    if unit.GetHealth() > 0 {
                        aliveUnits = append(aliveUnits, unit)
                    }
                }

                engine.Player.Units = aliveUnits
                if len(engine.Player.Units) == 0 {
                    log.Printf("All units lost, starting new game")
                    engine.Player = player.MakePlayer(data.BannerGreen)
                    engine.Player.AddUnit(units.LizardSwordsmen)
                }

                engine.UI, engine.UIUpdates, err = engine.MakeUI()
                if err != nil {
                    log.Printf("Error creating UI: %v", err)
                }
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

type UIEvent interface {
}

type UIUpdateMoney struct {
}

type UIAddUnit struct {
    Unit units.StackUnit
}

type UIEventUpdate struct {
    Listeners []func(UIEvent)
    Updates []UIEvent
}

func (events *UIEventUpdate) Update() {
    for _, update := range events.Updates {
        for _, listener := range events.Listeners {
            listener(update)
        }
    }
    events.Updates = nil
}

func (events *UIEventUpdate) Add(f func(UIEvent)) {
    events.Listeners = append(events.Listeners, f)
}

func AddEvent[T UIEvent](events *UIEventUpdate, f func(*T)) {
    events.Add(func (event UIEvent) {
        if update, ok := event.(*T); ok {
            f(update)
        }
    })
}

func (events *UIEventUpdate) AddUpdate(event UIEvent) {
    events.Updates = append(events.Updates, event)
}

func makeShopUI(face *text.GoTextFace, playerObj *player.Player, uiEvents *UIEventUpdate) *widget.Container {
    container := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(2),
            widget.RowLayoutOpts.Padding(widget.Insets{Top: 2, Bottom: 2, Left: 2, Right: 2}),
        )),
    )

    container.AddChild(widget.NewText(
        widget.TextOpts.Text("Shop", face, color.White),
    ))

    money := widget.NewText(
        widget.TextOpts.Text(fmt.Sprintf("Money: %d", playerObj.Money), face, color.White),
    )

    AddEvent(uiEvents, func (update *UIUpdateMoney) {
        money.Label = fmt.Sprintf("Money: %d", playerObj.Money)
    })

    container.AddChild(money)

    unitName := widget.NewText(
        widget.TextOpts.Text("Name: ", face, color.White),
    )

    unitCost := widget.NewText(
        widget.TextOpts.Text("Cost: ", face, color.White),
    )

    container2 := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
            widget.RowLayoutOpts.Spacing(4),
        )),
    )

    container.AddChild(container2)

    var selected *units.Unit

    makeList := func() *widget.List {
        return widget.NewList(
            widget.ListOpts.EntryFontFace(face),
            widget.ListOpts.SliderOpts(
                widget.SliderOpts.Images(
                    &widget.SliderTrackImage{
                        Idle: ui.SolidImage(64, 64, 64),
                        Hover: ui.SolidImage(96, 96, 96),
                    },
                    ui.MakeButtonImage(ui.SolidImage(192, 192, 192)),
                ),
            ),
            widget.ListOpts.HideHorizontalSlider(),
            widget.ListOpts.ContainerOpts(
                widget.ContainerOpts.WidgetOpts(
                    widget.WidgetOpts.LayoutData(widget.GridLayoutData{
                        MaxHeight: 500,
                    }),
                    widget.WidgetOpts.MinSize(0, 200),
                ),
            ),
            widget.ListOpts.EntryLabelFunc(
                func (e any) string {
                    // unit := e.(units.StackUnit)
                    // return unit.GetName()
                    unit := e.(*units.Unit)

                    return fmt.Sprintf("%v %v", unit.Race, unit.Name)
                },
            ),
            widget.ListOpts.EntrySelectedHandler(func (args *widget.ListEntrySelectedEventArgs) {
                // log.Printf("Selected unit: %v", args.Entry)
                unit := args.Entry.(*units.Unit)

                unitName.Label = fmt.Sprintf("Name: %v", unit.Name)
                unitCost.Label = fmt.Sprintf("Cost: %d", getUnitCost(unit))
                selected = unit
            }),
            widget.ListOpts.EntryColor(&widget.ListEntryColor{
                Selected: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                Unselected: color.NRGBA{R: 128, G: 128, B: 128, A: 255},
            }),
            widget.ListOpts.ScrollContainerOpts(
                widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
                    Idle: ui.SolidImage(64, 64, 64),
                    Disabled: ui.SolidImage(32, 32, 32),
                    Mask: ui.SolidImage(32, 32, 32),
                }),
            ),
        )
    }

    unitList := makeList()

    // container2.AddChild(unitList)

    for _, unit := range getValidChoices(100000) {
        unitList.AddEntry(unit)
    }

    filteredUnitList := makeList()

    setupFilteredList := func() {
        filteredUnitList.SetEntries(nil)
        for _, unit := range getValidChoices(playerObj.Money) {
            filteredUnitList.AddEntry(unit)
        }
    }

    setupFilteredList()

    AddEvent(uiEvents, func (update *UIUpdateMoney) {
        setupFilteredList()
    })

    tabAll := widget.NewTabBookTab("All", widget.ContainerOpts.Layout(widget.NewGridLayout(
        widget.GridLayoutOpts.Columns(1),
        widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false}),
    )))
    tabAll.AddChild(unitList)
    tabAffordable := widget.NewTabBookTab("Affordable", widget.ContainerOpts.Layout(
        widget.NewGridLayout(
        widget.GridLayoutOpts.Columns(1),
        widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false}),
    )))
    tabAffordable.AddChild(filteredUnitList)

    tabs := widget.NewTabBook(
        widget.TabBookOpts.TabButtonImage(ui.MakeButtonImage(ui.SolidImage(64, 32, 32))),
        widget.TabBookOpts.TabButtonText(face, &widget.ButtonTextColor{
            Idle: color.White,
            Disabled: color.NRGBA{R: 32, G: 32, B: 32, A: 255},
            Hover: color.White,
            Pressed: color.White,
        }),
        widget.TabBookOpts.TabButtonSpacing(5),
        // widget.TabBookOpts.ContentPadding(widget.NewInsetsSimple(2)),
        widget.TabBookOpts.Tabs(tabAll, tabAffordable),
    )

    container2.AddChild(tabs)

    infoContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(4),
        )),
    )

    infoContainer.AddChild(unitName)
    infoContainer.AddChild(unitCost)

    infoContainer.AddChild(widget.NewButton(
        widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Image(ui.MakeButtonImage(ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 32, B: 32, A: 255}))),
        widget.ButtonOpts.Text("Buy Unit", face, &widget.ButtonTextColor{
            Idle: color.White,
            Hover: color.White,
            Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
        widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
            if selected == nil {
                return
            }
            unitCost := getUnitCost(selected)
            if unitCost <= playerObj.Money {
                playerObj.Money -= unitCost
                newUnit := playerObj.AddUnit(*selected)
                uiEvents.AddUpdate(&UIAddUnit{Unit: newUnit})
                uiEvents.AddUpdate(&UIUpdateMoney{})
            }

        }),
    ))

    container2.AddChild(infoContainer)

    return container
}

func getHealCost(unit units.StackUnit, amount int) int {
    raw := unit.GetRawUnit()
    return int(float64(getUnitCost(&raw)) * 0.8 * float64(amount) / float64(unit.GetMaxHealth()))
}

func enlargeTransform(factor int) util.ImageTransformFunc {
    var f util.ImageTransformFunc

    f = func (original *image.Paletted) image.Image {
        newImage := image.NewPaletted(image.Rect(0, 0, original.Bounds().Dx() * factor, original.Bounds().Dy() * factor), original.Palette)

        for y := 0; y < original.Bounds().Dy(); y++ {
            for x := 0; x < original.Bounds().Dx(); x++ {
                colorIndex := original.ColorIndexAt(x, y)

                for dy := 0; dy < factor; dy++ {
                    for dx := 0; dx < factor; dx++ {
                        newImage.SetColorIndex(x * factor + dx, y * factor + dy, colorIndex)
                    }
                }
            }
        }

        return newImage
    }

    return f 
}

func makeUnitInfoUI(face *text.GoTextFace, allUnits []units.StackUnit, playerObj *player.Player, uiEvents *UIEventUpdate, imageCache *util.ImageCache) *widget.Container {

    unitSpecifics := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(2),
        )),
    )

    unitSpecifics.AddChild(widget.NewText(
        widget.TextOpts.Text("Unit Specifics", face, color.White),
    ))

    var updateUnitSpecifics func(unit units.StackUnit)

    updateUnitSpecifics = func(unit units.StackUnit) {
        currentName := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Name: %v", unit.GetFullName()), face, color.White))
        currentHealth := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("HP: %d/%d", unit.GetHealth(), unit.GetMaxHealth()), face, color.White))
        currentRace := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Race: %v", unit.GetRace()), face, color.White))

        unitSpecifics.RemoveChildren()
        unitSpecifics.AddChild(currentName)
        unitSpecifics.AddChild(currentHealth)
        unitSpecifics.AddChild(currentRace)

        // var currentHealTarget units.StackUnit
        healCost := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Heal %d hp for %d gold", unit.GetDamage(), getHealCost(unit, unit.GetDamage())), face, color.White))

        healSlider := widget.NewSlider(
            widget.SliderOpts.Direction(widget.DirectionHorizontal),
            widget.SliderOpts.MinMax(0, unit.GetDamage()),
            widget.SliderOpts.InitialCurrent(unit.GetDamage()),
            widget.SliderOpts.WidgetOpts(
                widget.WidgetOpts.MinSize(200, 10),
            ),
            widget.SliderOpts.Images(
                &widget.SliderTrackImage{
                    Idle: ui.SolidImage(64, 64, 64),
                    Hover: ui.SolidImage(96, 96, 96),
                },
                &widget.ButtonImage{
                    Idle: ui.SolidImage(192, 192, 192),
                    Hover: ui.SolidImage(255, 255, 0),
                    Pressed: ui.SolidImage(255, 128, 0),
                },
            ),
            widget.SliderOpts.FixedHandleSize(6),
            widget.SliderOpts.TrackOffset(0),
            widget.SliderOpts.PageSizeFunc(func() int {
                return 3
            }),
            widget.SliderOpts.ChangedHandler(func (args *widget.SliderChangedEventArgs) {
                healCost.Label = fmt.Sprintf("Heal %d hp for %d gold", args.Slider.Current, getHealCost(unit, args.Slider.Current))
            }),
        )

        healButton := widget.NewButton(
            widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(ui.MakeButtonImage(ui.SolidImage(64, 32, 32))),
            widget.ButtonOpts.Text("Heal Unit", face, &widget.ButtonTextColor{
                Idle: color.White,
                Hover: color.White,
                Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            }),
            widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
                cost := uint64(getHealCost(unit, healSlider.Current))

                if cost <= playerObj.Money {
                    unit.AdjustHealth(healSlider.Current)
                    updateUnitSpecifics(unit)
                    playerObj.Money -= uint64(cost)
                    uiEvents.AddUpdate(&UIUpdateMoney{})
                }
            }),
        )

        unitSpecifics.AddChild(healCost)
        unitSpecifics.AddChild(healSlider)
        unitSpecifics.AddChild(healButton)
    }

    unitList := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewGridLayout(
            widget.GridLayoutOpts.Columns(1),
            widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false}),
        )),
    )

    for _, unit := range allUnits {
        unitBox := ui.VBox()
        unitBox.AddChild(widget.NewText(
            widget.TextOpts.Text(fmt.Sprintf("%v %v", unit.GetRace(), unit.GetFullName()), face, color.White)),
        )

        unitImage, err := imageCache.GetImageTransform(unit.GetLbxFile(), unit.GetLbxIndex(), 0, "enlarge", enlargeTransform(2))
        if err == nil {
            unitBox.AddChild(widget.NewGraphic(widget.GraphicOpts.Image(unitImage)))
        }

        unitList.AddChild(unitBox)
    }

    scroller := widget.NewScrollContainer(
        widget.ScrollContainerOpts.WidgetOpts(
            widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                MaxHeight: 400,
            }),
        ),
        widget.ScrollContainerOpts.Content(unitList),
        widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
            Idle: ui.SolidImage(64, 64, 64),
            Mask: ui.SolidImage(32, 32, 32),
        }),
    )

    slider := widget.NewSlider(
        widget.SliderOpts.Direction(widget.DirectionVertical),
        widget.SliderOpts.MinMax(0, 100),
        widget.SliderOpts.InitialCurrent(0),
        widget.SliderOpts.ChangedHandler(func (args *widget.SliderChangedEventArgs) {
            scroller.ScrollTop = float64(args.Slider.Current) / 100
        }),
        widget.SliderOpts.WidgetOpts(
            widget.WidgetOpts.MinSize(10, 400),
        ),
        widget.SliderOpts.Images(
            &widget.SliderTrackImage{
                Idle: ui.SolidImage(64, 64, 64),
                Hover: ui.SolidImage(96, 96, 96),
            },
            &widget.ButtonImage{
                Idle: ui.SolidImage(192, 192, 192),
                Hover: ui.SolidImage(255, 255, 0),
                Pressed: ui.SolidImage(255, 128, 0),
            },
        ),
    )

    /*
    unitList := widget.NewList(
        widget.ListOpts.EntryFontFace(face),
        widget.ListOpts.SliderOpts(
            widget.SliderOpts.Images(
                &widget.SliderTrackImage{
                    Idle: ui.SolidImage(64, 64, 64),
                    Hover: ui.SolidImage(96, 96, 96),
                },
                ui.MakeButtonImage(ui.SolidImage(192, 192, 192)),
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
                return fmt.Sprintf("%v %v", unit.GetRace(), unit.GetName())
            },
        ),
        widget.ListOpts.EntrySelectedHandler(func (args *widget.ListEntrySelectedEventArgs) {
            // log.Printf("Selected unit: %v", args.Entry)

            unit := args.Entry.(units.StackUnit)

            updateUnitSpecifics(unit)
        }),
        widget.ListOpts.EntryColor(&widget.ListEntryColor{
            Selected: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            Unselected: color.NRGBA{R: 128, G: 128, B: 128, A: 255},
        }),
        widget.ListOpts.ScrollContainerOpts(
            widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
                Idle: ui.SolidImage(64, 64, 64),
                Disabled: ui.SolidImage(32, 32, 32),
                Mask: ui.SolidImage(32, 32, 32),
            }),
        ),
    )
    */

    /*
    for _, unit := range allUnits {
        unitList.AddEntry(unit)
    }

    AddEvent(uiEvents, func (update *UIAddUnit) {
        unitList.AddEntry(update.Unit)
    })
    */

    armyInfo := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(2),
            widget.RowLayoutOpts.Padding(widget.Insets{Top: 2, Bottom: 2, Left: 2, Right: 2}),
        )),
        widget.ContainerOpts.BackgroundImage(ui_image.NewNineSliceColor(color.NRGBA{R: 48, G: 48, B: 48, A: 255})),
    )

    armyInfo.AddChild(widget.NewText(
        widget.TextOpts.Text("Army", face, color.White),
    ))

    scrollStuff := ui.HBox()
    scrollStuff.AddChild(scroller)
    scrollStuff.AddChild(slider)

    armyInfo.AddChild(scrollStuff)

    unitInfoContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
            widget.RowLayoutOpts.Spacing(4),
            // widget.RowLayoutOpts.Padding(widget.Insets{Top: 4, Bottom: 4, Left: 4, Right: 4}),
        )),
    )

    unitInfoContainer.AddChild(armyInfo)

    unitInfoContainer.AddChild(unitSpecifics)

    return unitInfoContainer
}

func makePlayerInfoUI(face *text.GoTextFace, playerObj *player.Player) *widget.Container {
    container := ui.HBox()

    name := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Name: %v", playerObj.Wizard.Name), face, color.White))
    container.AddChild(name)

    level := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Level: %d", playerObj.Level), face, color.White))
    container.AddChild(level)

    return container
}

func (engine *Engine) MakeUI() (*ebitenui.UI, *UIEventUpdate, error) {
    font, err := console.LoadFont()
    if err != nil {
        return nil, nil, err
    }

    face := text.GoTextFace{
        Source: font,
        Size: 18,
    }

    var uiEvents UIEventUpdate

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
        widget.ButtonOpts.Image(ui.MakeButtonImage(ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 32, B: 32, A: 255}))),
        widget.ButtonOpts.Text("Enter Battle", &face, &widget.ButtonTextColor{
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

    rootContainer.AddChild(makePlayerInfoUI(&face, engine.Player))

    imageCache := util.MakeImageCache(engine.Cache)

    unitInfoUI := makeUnitInfoUI(&face, engine.Player.Units, engine.Player, &uiEvents, &imageCache)

    rootContainer.AddChild(unitInfoUI)
    rootContainer.AddChild(makeShopUI(&face, engine.Player, &uiEvents))

    ui := &ebitenui.UI{
        Container: rootContainer,
    }

    return ui, &uiEvents, nil
}

func MakeEngine(cache *lbx.LbxCache) *Engine {
    playerObj := player.MakePlayer(data.BannerGreen)

    for range 20 {
        playerObj.AddUnit(units.LizardSwordsmen)
    }

    engine := Engine{
        GameMode: GameModeUI,
        Player: playerObj,
        Cache: cache,
        Events: make(chan EngineEvents, 10),
    }

    var err error
    engine.UI, engine.UIUpdates, err = engine.MakeUI()
    if err != nil {
        log.Printf("Error creating UI: %v", err)
    }
    return &engine
}

func showCost() {
    for _, unit := range units.AllUnits {
        cost := getUnitCost(&unit)
        log.Printf("Unit %v %v costs %v", unit.Race, unit.Name, cost)
    }
}

func main() {
    log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)

    showCost()

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
