package main

import (
    "log"
    "errors"
    "math/rand/v2"
    "image/color"
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

    budget := uint64(engine.Player.Level) * 200
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

func (events *UIEventUpdate) AddUpdate(event UIEvent) {
    events.Updates = append(events.Updates, event)
}

func makeShopUI(face *text.GoTextFace, playerObj *player.Player, buyCallback func(units.StackUnit), uiEvents *UIEventUpdate) *widget.Container {
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

    uiEvents.Add(func (event UIEvent) {
        switch event.(type) {
            case *UIUpdateMoney:
                money.Label = fmt.Sprintf("Money: %d", playerObj.Money)
        }
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
                widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                    MaxHeight: 400,
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

    container2.AddChild(unitList)

    for _, unit := range getValidChoices(100000) {
        unitList.AddEntry(unit)
    }

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
            selected := unitList.SelectedEntry()

            unit := selected.(*units.Unit)

            unitCost := getUnitCost(unit)
            if unitCost <= playerObj.Money {
                playerObj.Money -= unitCost
                newUnit := playerObj.AddUnit(*unit)
                buyCallback(newUnit)

                money.Label = fmt.Sprintf("Money: %d", playerObj.Money)
            }

        }),
    ))

    container2.AddChild(infoContainer)

    return container
}

func makeUnitInfoUI(face *text.GoTextFace, allUnits []units.StackUnit, playerObj *player.Player, uiEvents *UIEventUpdate) (*widget.Container, func(units.StackUnit)) {

    currentName := widget.NewText(widget.TextOpts.Text("", face, color.White))
    currentHealth := widget.NewText(widget.TextOpts.Text("", face, color.White))
    currentRace := widget.NewText(widget.TextOpts.Text("", face, color.White))

    updateHealth := func(unit units.StackUnit) {
        currentHealth.Label = fmt.Sprintf("HP: %d/%d", unit.GetHealth(), unit.GetMaxHealth())
    }

    var currentHealTarget units.StackUnit
    healCost := widget.NewText(widget.TextOpts.Text("", face, color.White))

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

            currentName.Label = fmt.Sprintf("Name: %v", unit.GetFullName())
            updateHealth(unit)
            currentRace.Label = fmt.Sprintf("Race: %v", unit.GetRace())
            currentHealTarget = unit

            healCost.Label = fmt.Sprintf("Heal Cost %d", 20 * currentHealTarget.GetDamage())
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

    for _, unit := range allUnits {
        unitList.AddEntry(unit)
    }

    buyCallback := func(unit units.StackUnit) {
        unitList.AddEntry(unit)
    }

    armyInfo.AddChild(unitList)

    unitInfoContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
            widget.RowLayoutOpts.Spacing(4),
            // widget.RowLayoutOpts.Padding(widget.Insets{Top: 4, Bottom: 4, Left: 4, Right: 4}),
        )),
    )

    unitInfoContainer.AddChild(armyInfo)

    unitSpecifics := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(2),
        )),
    )

    unitSpecifics.AddChild(widget.NewText(
        widget.TextOpts.Text("Unit Specifics", face, color.White),
    ))
    unitSpecifics.AddChild(currentName)
    unitSpecifics.AddChild(currentRace)
    unitSpecifics.AddChild(currentHealth)

    healButton := widget.NewButton(
        widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Image(ui.MakeButtonImage(ui.SolidImage(64, 32, 32))),
        widget.ButtonOpts.Text("Heal Unit", face, &widget.ButtonTextColor{
            Idle: color.White,
            Hover: color.White,
            Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
        widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
            if currentHealTarget == nil {
                return
            }

            cost := uint64(20 * currentHealTarget.GetDamage())

            if cost <= playerObj.Money {
                currentHealTarget.AdjustHealth(currentHealTarget.GetDamage())
                updateHealth(currentHealTarget)
                playerObj.Money -= uint64(cost)
                uiEvents.AddUpdate(&UIUpdateMoney{})
            }
        }),
    )

    healBox := ui.HBox()

    healBox.AddChild(healButton)
    healBox.AddChild(healCost)

    unitSpecifics.AddChild(healBox)

    unitInfoContainer.AddChild(unitSpecifics)

    return unitInfoContainer, buyCallback
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

    unitInfoUI, buyCallback := makeUnitInfoUI(&face, engine.Player.Units, engine.Player, &uiEvents)

    rootContainer.AddChild(unitInfoUI)
    rootContainer.AddChild(makeShopUI(&face, engine.Player, buyCallback, &uiEvents))

    ui := &ebitenui.UI{
        Container: rootContainer,
    }

    return ui, &uiEvents, nil
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
