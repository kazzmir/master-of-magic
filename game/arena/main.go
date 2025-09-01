package main

import (
    "log"
    "errors"
    "math"
    "math/rand/v2"
    "image/color"
    "image"
    "fmt"
    "slices"
    "cmp"

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
    "github.com/hajimehoshi/ebiten/v2/vector"

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
                } else {
                    for _, unit := range engine.Player.Units {
                        unit.AddExperience(20)
                    }
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

type SortDirection int

const (
    SortDirectionAscending SortDirection = iota
    SortDirectionDescending
)

func (sort SortDirection) Next() SortDirection {
    switch sort {
        case SortDirectionAscending: return SortDirectionDescending
        case SortDirectionDescending: return SortDirectionAscending
    }
    return SortDirectionAscending
}

type UnitIconList struct {
    unitList *widget.Container
    container *widget.Container
    face *text.GoTextFace
    lastBox *widget.Container
    SelectedUnit func(unit *units.Unit)
    imageCache *util.ImageCache
    units []*units.Unit

    SortNameDirection SortDirection
    SortCostDirection SortDirection
}

func MakeUnitIconList(imageCache *util.ImageCache, face *text.GoTextFace, selectedUnit func(*units.Unit)) *UnitIconList {
    var iconList UnitIconList

    iconList.imageCache = imageCache
    iconList.SelectedUnit = selectedUnit
    iconList.face = face

    iconList.unitList = widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewGridLayout(
            widget.GridLayoutOpts.Columns(3),
            widget.GridLayoutOpts.Stretch([]bool{true, true, true}, []bool{false, false, false}),
        )),
    )

    scroller := widget.NewScrollContainer(
        widget.ScrollContainerOpts.WidgetOpts(
            widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                MaxHeight: 400,
            }),
        ),
        widget.ScrollContainerOpts.Content(iconList.unitList),
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
        widget.SliderOpts.PageSizeFunc(func() int {
            return 20
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

    scroller.GetWidget().ScrolledEvent.AddHandler(func (args any) {
        eventArgs := args.(*widget.WidgetScrolledEventArgs)
        slider.Current -= int(math.Round(eventArgs.Y * 8))
    })

    scrollStuff := ui.HBox()
    scrollStuff.AddChild(scroller)
    scrollStuff.AddChild(slider)

    box := ui.VBox()

    sortButtons := ui.HBox()

    // only change how we sort if the same button is pressed twice in a row
    lastSort := 0

    sortButtons.AddChild(widget.NewButton(
        widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Image(ui.MakeButtonImage(ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 32, B: 32, A: 255}))),
        widget.ButtonOpts.Text("Sort by Name", face, &widget.ButtonTextColor{
            Idle: color.White,
            Hover: color.White,
            Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
        widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
            if lastSort == 0 {
                iconList.SortNameDirection = iconList.SortNameDirection.Next()
            }
            lastSort = 0
            iconList.SortByName()
        }),
    ))

    sortButtons.AddChild(widget.NewButton(
        widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Image(ui.MakeButtonImage(ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 32, B: 32, A: 255}))),
        widget.ButtonOpts.Text("Sort by Cost", face, &widget.ButtonTextColor{
            Idle: color.White,
            Hover: color.White,
            Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
        widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
            if lastSort == 1 {
                iconList.SortCostDirection = iconList.SortCostDirection.Next()
            }
            lastSort = 1
            iconList.SortByCost()
        }),
    ))


    box.AddChild(sortButtons)
    box.AddChild(scrollStuff)

    iconList.container = box

    return &iconList
}

func (iconList *UnitIconList) Clear() {
    iconList.unitList.RemoveChildren()
    iconList.lastBox = nil
}

func (iconList *UnitIconList) SortByName() {
    var sortFunc func(a, b *units.Unit) int

    switch iconList.SortNameDirection {
        case SortDirectionAscending:
            sortFunc = func(a, b *units.Unit) int {
                return cmp.Compare(fmt.Sprintf("%v %v", a.Race, a.Name), fmt.Sprintf("%v %v", b.Race, b.Name))
            }
        case SortDirectionDescending:
            sortFunc = func(a, b *units.Unit) int {
                return cmp.Compare(fmt.Sprintf("%v %v", b.Race, b.Name), fmt.Sprintf("%v %v", a.Race, a.Name))
            }
    }

    slices.SortFunc(iconList.units, sortFunc)

    iconList.Clear()
    for _, unit := range iconList.units {
        iconList.addUI(unit)
    }
}

func (iconList *UnitIconList) SortByCost() {
    var sortFunc func(a, b *units.Unit) int

    switch iconList.SortCostDirection {
        case SortDirectionAscending:
            sortFunc = func(a, b *units.Unit) int {
                return cmp.Compare(getUnitCost(a), getUnitCost(b))
            }
        case SortDirectionDescending:
            sortFunc = func(a, b *units.Unit) int {
                return cmp.Compare(getUnitCost(b), getUnitCost(a))
            }
    }

    slices.SortFunc(iconList.units, sortFunc)
    iconList.Clear()
    for _, unit := range iconList.units {
        iconList.addUI(unit)
    }
}

func (iconList *UnitIconList) addUI(unit *units.Unit) {
    var unitBox *widget.Container
    unitBox = ui.VBox(widget.ContainerOpts.WidgetOpts(
        widget.WidgetOpts.MouseButtonPressedHandler(func(args *widget.WidgetMouseButtonPressedEventArgs) {
            iconList.SelectedUnit(unit)
            if iconList.lastBox != nil {
                iconList.lastBox.BackgroundImage = nil
            }
            unitBox.BackgroundImage = ui.SolidImage(96, 96, 32)
            iconList.lastBox = unitBox
        }),
    ))

    unitBox.AddChild(widget.NewText(
        widget.TextOpts.Text(fmt.Sprintf("%v %v", unit.Race, unit.Name), iconList.face, color.White)),
    )

    unitImage, err := iconList.imageCache.GetImageTransform(unit.GetCombatLbxFile(), unit.GetCombatIndex(units.FacingRight), 0, "enlarge", enlargeTransform(2))
    if err == nil {
        box1 := ui.HBox()
        box1.AddChild(widget.NewGraphic(widget.GraphicOpts.Image(unitImage)))
        box1.AddChild(widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Cost %d", getUnitCost(unit)), iconList.face, color.White)))
        unitBox.AddChild(box1)
    }

    iconList.unitList.AddChild(unitBox)
}

func (iconList *UnitIconList) AddUnit(unit *units.Unit) {
    iconList.units = append(iconList.units, unit)
    iconList.addUI(unit)
}

func (iconList *UnitIconList) GetWidget() *widget.Container {
    return iconList.container
}

func makeShopUI(face *text.GoTextFace, imageCache *util.ImageCache, playerObj *player.Player, uiEvents *UIEventUpdate) *widget.Container {
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

    unitList := MakeUnitIconList(imageCache, face, func(unit *units.Unit) {
        unitName.Label = fmt.Sprintf("Name: %v", unit.Name)
        unitCost.Label = fmt.Sprintf("Cost: %d", getUnitCost(unit))
        selected = unit
    })

    // container2.AddChild(unitList)

    for _, unit := range getValidChoices(100000) {
        unitList.AddUnit(unit)
    }

    unitList.SortByName()

    filteredUnitList := MakeUnitIconList(imageCache, face, func(unit *units.Unit) {
        unitName.Label = fmt.Sprintf("Name: %v", unit.Name)
        unitCost.Label = fmt.Sprintf("Cost: %d", getUnitCost(unit))
        selected = unit
    })

    setupFilteredList := func() {
        filteredUnitList.Clear()
        for _, unit := range getValidChoices(playerObj.Money) {
            filteredUnitList.AddUnit(unit)
        }

        filteredUnitList.SortByName()
    }

    setupFilteredList()

    AddEvent(uiEvents, func (update *UIUpdateMoney) {
        setupFilteredList()
    })

    tabAll := widget.NewTabBookTab("All", widget.ContainerOpts.Layout(widget.NewGridLayout(
        widget.GridLayoutOpts.Columns(1),
        widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false}),
    )))
    tabAll.AddChild(unitList.GetWidget())
    tabAffordable := widget.NewTabBookTab("Affordable", widget.ContainerOpts.Layout(
        widget.NewGridLayout(
        widget.GridLayoutOpts.Columns(1),
        widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false}),
    )))
    tabAffordable.AddChild(filteredUnitList.GetWidget())

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

    var updateUnitSpecifics func(unit units.StackUnit, setup func())

    updateUnitSpecifics = func(unit units.StackUnit, setup func()) {
        currentName := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Name: %v", unit.GetFullName()), face, color.White))
        currentHealth := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("HP: %d/%d", unit.GetHealth(), unit.GetMaxHealth()), face, color.White))
        currentRace := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Race: %v", unit.GetRace()), face, color.White))

        unitSpecifics.RemoveChildren()
        unitSpecifics.AddChild(currentName)
        unitSpecifics.AddChild(currentHealth)
        unitSpecifics.AddChild(currentRace)
        unitSpecifics.AddChild(widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Experience: %d (%v)", unit.GetExperience(), unit.GetExperienceLevel().Name()), face, color.White)))

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
                updated := false

                healCost := uint64(getHealCost(unit, 1))

                for range healSlider.Current {
                    if healCost > playerObj.Money || unit.GetHealth() >= unit.GetMaxHealth() {
                        break
                    }

                    unit.AdjustHealth(1)
                    playerObj.Money -= healCost
                    updated = true
                }

                if updated {
                    updateUnitSpecifics(unit, setup)
                    uiEvents.AddUpdate(&UIUpdateMoney{})
                    setup()
                }
            }),
        )

        unitSpecifics.AddChild(healCost)
        unitSpecifics.AddChild(healSlider)
        unitSpecifics.AddChild(healButton)
    }

    unitList := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewGridLayout(
            widget.GridLayoutOpts.Columns(2),
            widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{false, false}),
        )),
    )

    var lastBox *widget.Container

    addUnit := func(unit units.StackUnit) {
        var unitBox *widget.Container
        var setup func()
        unitBox = ui.VBox(widget.ContainerOpts.WidgetOpts(
            widget.WidgetOpts.MouseButtonPressedHandler(func(args *widget.WidgetMouseButtonPressedEventArgs) {
                updateUnitSpecifics(unit, setup)
                if lastBox != nil {
                    lastBox.BackgroundImage = nil
                }
                unitBox.BackgroundImage = ui.SolidImage(96, 96, 32)
                lastBox = unitBox
            }),
        ))

        setup = func(){
            unitBox.RemoveChildren()
            unitBox.AddChild(widget.NewText(
                widget.TextOpts.Text(fmt.Sprintf("%v %v", unit.GetRace(), unit.GetFullName()), face, color.White)),
            )

            unitImage, err := imageCache.GetImageTransform(unit.GetLbxFile(), unit.GetLbxIndex(), 0, "enlarge", enlargeTransform(2))
            if err == nil {

                box1 := ui.HBox()
                box2 := ui.VBox()
                box1.AddChild(box2)
                box2.AddChild(widget.NewGraphic(widget.GraphicOpts.Image(unitImage)))

                badges := ebiten.NewImage(40, 20)
                badges.Fill(color.RGBA{})

                badgeInfo := units.GetExperienceBadge(unit)

                var badgeOptions ebiten.DrawImageOptions

                badgeOptions.GeoM.Translate(1, 1)
                for range badgeInfo.Count {
                    pic, _ := imageCache.GetImage("main.lbx", badgeInfo.Badge.IconLbxIndex(), 0)
                    scale.DrawScaled(badges, pic, &badgeOptions)
                    badgeOptions.GeoM.Translate(4, 0)
                }

                box2.AddChild(widget.NewGraphic(widget.GraphicOpts.Image(badges)))

                highHealth := color.RGBA{R: 0, G: 0xff, B: 0, A: 0xff}
                mediumHealth := color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff}
                lowHealth := color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}

                healthColor := highHealth
                percent := float32(unit.GetHealth()) / float32(unit.GetMaxHealth())

                if percent < 0.33 {
                    healthColor = lowHealth
                } else if percent < 0.66 {
                    healthColor = mediumHealth
                }

                healthImage := ebiten.NewImage(80, 20)
                length := float32(healthImage.Bounds().Dx()) * percent
                if length < 1 {
                    length = 1
                }

                vector.DrawFilledRect(healthImage, 0, 10, float32(healthImage.Bounds().Dx()), 4, color.RGBA{R: 0, G: 0, B: 0, A: 255}, true)
                vector.DrawFilledRect(healthImage, 0, 10, length, 4, healthColor, true)
                box1.AddChild(widget.NewGraphic(widget.GraphicOpts.Image(healthImage)))

                unitBox.AddChild(box1)
            }
        }

        setup()

        unitList.AddChild(unitBox)
    }

    for _, unit := range allUnits {
        addUnit(unit)
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
        widget.SliderOpts.PageSizeFunc(func() int {
            return 20
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

    scroller.GetWidget().ScrolledEvent.AddHandler(func (args any) {
        eventArgs := args.(*widget.WidgetScrolledEventArgs)
        slider.Current -= int(math.Round(eventArgs.Y * 8))
    })

    AddEvent(uiEvents, func (update *UIAddUnit) {
        addUnit(update.Unit)
    })

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
    rootContainer.AddChild(makeShopUI(&face, &imageCache, engine.Player, &uiEvents))

    ui := &ebitenui.UI{
        Container: rootContainer,
    }

    return ui, &uiEvents, nil
}

func test1(playerObj *player.Player) {
    playerObj.AddUnit(units.LizardSwordsmen)
}

func test2(playerObj *player.Player) {
    v := playerObj.AddUnit(units.LizardSwordsmen)
    v.AddExperience(100)
}

func test3(playerObj *player.Player) {
    v := playerObj.AddUnit(units.LizardSwordsmen)
    v.AdjustHealth(-10)
    playerObj.Money = 30
}

func test4(playerObj *player.Player) {
    for range 5 {
        playerObj.AddUnit(units.LizardSwordsmen)
    }
    for range 5 {
        playerObj.AddUnit(units.Warlocks)
    }
}

func MakeEngine(cache *lbx.LbxCache) *Engine {
    playerObj := player.MakePlayer(data.BannerGreen)

    // test1(playerObj)
    // test3(playerObj)
    test4(playerObj)

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

    ebiten.SetWindowSize(1200, 1000)
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    engine := MakeEngine(cache)

    err := ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
