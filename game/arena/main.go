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
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"

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
                    // engine.Player.AddUnit(units.LizardSwordsmen)
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

type UIUpdateMagicBooks struct {
}

type UIUpdateMana struct {
}

type UIUpdateUnit struct {
    Unit units.StackUnit
}

type UIAddUnit struct {
    Unit units.StackUnit
}

type UIEventUpdate struct {
    Listeners map[uint64]func(UIEvent)
    Updates []UIEvent
    counter uint64
}

func MakeUIEventUpdate() *UIEventUpdate {
    return &UIEventUpdate{
        Listeners: make(map[uint64]func(UIEvent)),
        Updates: nil,
        counter: 0,
    }
}

func (events *UIEventUpdate) Update() {
    for _, update := range events.Updates {
        for _, listener := range events.Listeners {
            listener(update)
        }
    }
    events.Updates = nil
}

func (events *UIEventUpdate) Remove(id uint64){
    delete(events.Listeners, id)
}

// returns the id of the listener
func (events *UIEventUpdate) Add(f func(UIEvent)) uint64 {
    events.Listeners[events.counter] = f
    events.counter += 1
    return events.counter - 1
}

func AddEvent[T UIEvent](events *UIEventUpdate, f func(*T)) uint64 {
    return events.Add(func (event UIEvent) {
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
    face *text.Face
    lastBox *widget.Container
    buyUnit func(unit *units.Unit)
    imageCache *util.ImageCache
    units []*units.Unit

    SortNameDirection SortDirection
    SortCostDirection SortDirection
}

func MakeUnitIconList(description string, imageCache *util.ImageCache, face *text.Face, buyUnit func(*units.Unit)) *UnitIconList {
    var iconList UnitIconList

    iconList.imageCache = imageCache
    iconList.buyUnit = buyUnit
    iconList.face = face

    iconList.unitList = widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewGridLayout(
            widget.GridLayoutOpts.Columns(4),
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

    baseImage := ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 32, B: 32, A: 255})
    sortButtons.AddChild(widget.NewButton(
        widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        // widget.ButtonOpts.ToggleMode(),
        // widget.ButtonOpts.Image(ui.MakeButtonImage(ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 32, B: 32, A: 255}))),
        widget.ButtonOpts.Image(&widget.ButtonImage{
            Idle: baseImage,
            Hover: baseImage,
            Pressed: baseImage,
            Disabled: baseImage,
            // PressedHover: ui_image.NewNineSliceColor(color.NRGBA{R: 255, G: 64, B: 32, A: 255}),
        }),

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
        widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
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

    box.AddChild(widget.NewText(widget.TextOpts.Text(description, face, color.White)))

    box.AddChild(sortButtons)
    box.AddChild(scrollStuff)

    iconList.container = box

    return &iconList
}

func (iconList *UnitIconList) Reset() {
    iconList.units = nil
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
    unitBox = ui.VBox(
        widget.ContainerOpts.BackgroundImage(ui.BorderedImage(color.RGBA{R: 128, G: 128, B: 128, A: 255}, 1)),

        /*
        widget.ContainerOpts.WidgetOpts(
        widget.WidgetOpts.MouseButtonPressedHandler(func(args *widget.WidgetMouseButtonPressedEventArgs) {
            if iconList.lastBox != nil {
                iconList.lastBox.BackgroundImage = nil
            }
            unitBox.BackgroundImage = ui.SolidImage(96, 96, 32)
            iconList.lastBox = unitBox
        }),
    )
        */
    )

    unitBox.AddChild(ui.CenteredText(unit.Race.String(), iconList.face, color.White))
    unitBox.AddChild(ui.CenteredText(unit.Name, iconList.face, color.White))

    unitImage, err := iconList.imageCache.GetImageTransform(unit.GetCombatLbxFile(), unit.GetCombatIndex(units.FacingRight), 0, "enlarge", enlargeTransform(2))
    if err == nil {
        unitBox.AddChild(widget.NewGraphic(
            widget.GraphicOpts.Image(unitImage),
            widget.GraphicOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                Position: widget.RowLayoutPositionCenter,
            })),
        ))
    }

    // unitBox.AddChild(ui.CenteredText(fmt.Sprintf("Cost %d", getUnitCost(unit)), iconList.face, color.White))
    money := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("%d", getUnitCost(unit)), iconList.face, color.White))
    unitBox.AddChild(makeMoneyText(money, iconList.imageCache, widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
        Position: widget.RowLayoutPositionCenter,
    }))))

    unitBox.AddChild(widget.NewButton(
        widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
            Position: widget.RowLayoutPositionCenter,
        })),

        widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Image(ui.MakeButtonImage(ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 32, B: 32, A: 255}))),
        widget.ButtonOpts.Text("Buy Unit", iconList.face, &widget.ButtonTextColor{
            Idle: color.White,
            Hover: color.White,
            Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
        widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
            iconList.buyUnit(unit)
        }),
    ))

    iconList.unitList.AddChild(unitBox)
}

func (iconList *UnitIconList) AddUnit(unit *units.Unit) {
    iconList.units = append(iconList.units, unit)
    iconList.addUI(unit)
}

func (iconList *UnitIconList) GetWidget() *widget.Container {
    return iconList.container
}

func makeGraphicText(text *widget.Text, image *ebiten.Image, opts... widget.ContainerOpt) *widget.Container {
    box := ui.HBox(opts...)
    box.AddChild(widget.NewGraphic(
        widget.GraphicOpts.Image(image),
        widget.GraphicOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
            Position: widget.RowLayoutPositionCenter,
        })),
    ))

    box.AddChild(text)
    return box
}

func makeMoneyText(text *widget.Text, imageCache *util.ImageCache, opts... widget.ContainerOpt) *widget.Container {
    goldImage, err := imageCache.GetImageTransform("backgrnd.lbx", 42, 0, "enlarge", enlargeTransform(2))
    if err == nil {
        return makeGraphicText(text, goldImage, opts...)
    } else {
        return ui.HBox()
    }
}

func combineHorizontalElements(elements... widget.PreferredSizeLocateableWidget) *widget.Container {
    box := ui.HBox()
    box.AddChild(elements...)
    return box
}

func makeMagicShop(face *text.Face, imageCache *util.ImageCache, lbxCache *lbx.LbxCache, playerObj *player.Player, uiEvents *UIEventUpdate) *widget.Container {
    shop := ui.VBox()
    shop.AddChild(widget.NewText(widget.TextOpts.Text("Magic Shop", face, color.White)))

    books := ui.HBox()
    shop.AddChild(books)

    lifeBook, _ := imageCache.GetImageTransform("newgame.lbx", 24, 0, "enlarge", enlargeTransform(2))
    sorceryBook, _ := imageCache.GetImageTransform("newgame.lbx", 27, 0, "enlarge", enlargeTransform(2))
    natureBook, _ := imageCache.GetImageTransform("newgame.lbx", 30, 0, "enlarge", enlargeTransform(2))
    deathBook, _ := imageCache.GetImageTransform("newgame.lbx", 33, 0, "enlarge", enlargeTransform(2))
    chaosBook, _ := imageCache.GetImageTransform("newgame.lbx", 36, 0, "enlarge", enlargeTransform(2))

    centered := widget.WidgetOpts.LayoutData(widget.RowLayoutData{
        Position: widget.RowLayoutPositionCenter,
    })

    allSpells, err := spellbook.ReadSpellsFromCache(lbxCache)
    if err != nil {
        log.Printf("Error reading spells from cache: %v", err)
        return shop
    }

    allMagic := []data.MagicType{data.LifeMagic, data.SorceryMagic, data.NatureMagic, data.DeathMagic, data.ChaosMagic}

    setupMagic := func() {
        books.RemoveChildren()

        for _, magic := range allMagic {
            if playerObj.GetWizard().MagicLevel(magic) >= 11 {
                continue
            }

            buy := ui.VBox()

            var bookImage *ebiten.Image
            switch magic {
                case data.LifeMagic: bookImage = lifeBook
                case data.SorceryMagic: bookImage = sorceryBook
                case data.NatureMagic: bookImage = natureBook
                case data.DeathMagic: bookImage = deathBook
                case data.ChaosMagic: bookImage = chaosBook
            }

            graphic := widget.NewGraphic(widget.GraphicOpts.Image(bookImage), widget.GraphicOpts.WidgetOpts(centered))
            buy.AddChild(graphic)
            buy.AddChild(ui.CenteredText(magic.String(), face, color.White))

            makeIcon := func(image *ebiten.Image) *widget.Graphic {
                return widget.NewGraphic(widget.GraphicOpts.Image(image), widget.GraphicOpts.WidgetOpts(centered))
            }

            cost := uint64(math.Pow(10, 2.5 + float64(playerObj.GetWizard().MagicLevel(magic)) / 10))

            gold, _ := imageCache.GetImageTransform("backgrnd.lbx", 42, 0, "enlarge", enlargeTransform(2))
            costUI := combineHorizontalElements(makeIcon(gold), widget.NewText(widget.TextOpts.Text(fmt.Sprintf("%d", cost), face, color.White), widget.TextOpts.WidgetOpts(centered)))

            buy.AddChild(costUI)

            buyButton := widget.NewButton(
                widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
                widget.ButtonOpts.WidgetOpts(centered),
                widget.ButtonOpts.Image(ui.MakeButtonImage(ui.SolidImage(64, 32, 32))),
                widget.ButtonOpts.Text("Buy", face, &widget.ButtonTextColor{
                    Idle: color.White,
                    Hover: color.White,
                    Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                }),
                widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
                    if playerObj.Money >= cost {
                        playerObj.Money -= cost
                        uiEvents.AddUpdate(&UIUpdateMoney{})
                        uiEvents.AddUpdate(&UIUpdateMagicBooks{})
                        playerObj.GetWizard().AddMagicLevel(magic, 1)
                    }
                }),
            )

            buy.AddChild(buyButton)

            books.AddChild(buy)
        }
    }

    setupMagic()

    AddEvent(uiEvents, func (update *UIUpdateMoney) {
        setupMagic()
    })

    makeManaBuyButton := func(amount int) *widget.Button {
        return widget.NewButton(
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.WidgetOpts(centered),
            widget.ButtonOpts.Image(ui.MakeButtonImage(ui.SolidImage(64, 32, 32))),
            widget.ButtonOpts.Text(fmt.Sprintf("Buy %d", amount), face, &widget.ButtonTextColor{
                Idle: color.White,
                Hover: color.White,
                Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            }),
            widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
                if playerObj.Money >= uint64(amount) {
                    playerObj.Money -= uint64(amount)
                    playerObj.Mana += amount
                    uiEvents.AddUpdate(&UIUpdateMana{})
                    uiEvents.AddUpdate(&UIUpdateMoney{})
                }
            }),
        )
    }

    manaBox := ui.HBox()
    manaBox.AddChild(widget.NewText(widget.TextOpts.Text("Mana", face, color.White)))
    manaBox.AddChild(makeManaBuyButton(10), makeManaBuyButton(50), makeManaBuyButton(100))
    shop.AddChild(manaBox)

    var tabs []*widget.TabBookTab

    // how many spells of each rarity type the player can currently buy
    getCommonSpells := func(books int) int {
        if books >= 5 {
            return 10
        }

        if books <= 0 {
            return 0
        }

        return books * 2
    }

    getUncommonSpells := func(books int) int {
        books = books - 3

        if books >= 5 {
            return 10
        }

        if books <= 0 {
            return 0
        }

        return books * 2
    }

    getRareSpells := func(books int) int {
        books = books - 6

        if books >= 5 {
            return 10
        }

        if books <= 0 {
            return 0
        }

        return books * 2
    }

    getVeryRareSpells := func(books int) int {
        books = books - 9
        if books >= 5 {
            return 10
        }

        if books <= 0 {
            return 0
        }

        return books
    }

    for _, magic := range allMagic {
        tab := widget.NewTabBookTab(magic.String(), widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
        )))

        tab.AddChild(ui.CenteredText(fmt.Sprintf("%v Spells", magic), face, color.White))
        tabs = append(tabs, tab)

        containerSize := 300

        spellList := widget.NewContainer(
            widget.ContainerOpts.Layout(widget.NewGridLayout(
                widget.GridLayoutOpts.Columns(2),
            )),
        )

        commonCount := 0
        uncommonCount := 0
        rareCount := 0
        veryRareCount := 0

        for _, spell := range allSpells.GetSpellsByMagic(magic).Spells {

            if !spell.Eligibility.CanCastInCombat(false) {
                continue
            }

            var rarityCount int

            switch spell.Rarity {
                case spellbook.SpellRarityCommon:
                    commonCount += 1
                    rarityCount = commonCount
                case spellbook.SpellRarityUncommon:
                    uncommonCount += 1
                    rarityCount = uncommonCount
                case spellbook.SpellRarityRare:
                    rareCount += 1
                    rarityCount = rareCount
                case spellbook.SpellRarityVeryRare:
                    veryRareCount += 1
                    rarityCount = veryRareCount
            }

            border := ui.BorderedImage(color.RGBA{R: 128, G: 128, B: 128, A: 255}, 1)
            box := ui.VBox(
                widget.ContainerOpts.BackgroundImage(border),
            )

            var setupBox func()

            setupBox = func() {
                box.RemoveChildren()
                box.AddChild(ui.CenteredText(spell.Name, face, color.White))

                if playerObj.KnownSpells.Contains(spell) {
                    learned := ui.CenteredText("Learned", face, color.RGBA{R: 0, G: 255, B: 0, A: 255})
                    box.AddChild(learned)
                } else {

                    cost := uint64(spell.ResearchCost)

                    box.AddChild(makeMoneyText(ui.CenteredText(fmt.Sprintf("%d", cost), face, color.White), imageCache, widget.ContainerOpts.WidgetOpts(centered)))

                    buttonImage := ui.SolidImage(64, 32, 32)
                    buyTextColor := color.NRGBA{R: 255, G: 255, B: 0, A: 255}
                    buyTextIdle := color.NRGBA{R: 255, G: 255, B: 255, A: 255}

                    var canBuy bool

                    switch spell.Rarity {
                        case spellbook.SpellRarityCommon: canBuy = rarityCount <= getCommonSpells(playerObj.GetWizard().MagicLevel(magic))
                        case spellbook.SpellRarityUncommon: canBuy = rarityCount <= getUncommonSpells(playerObj.GetWizard().MagicLevel(magic))
                        case spellbook.SpellRarityRare: canBuy = rarityCount <= getRareSpells(playerObj.GetWizard().MagicLevel(magic))
                        case spellbook.SpellRarityVeryRare: canBuy = rarityCount <= getVeryRareSpells(playerObj.GetWizard().MagicLevel(magic))
                    }

                    if !canBuy {
                        buttonImage = ui.SolidImage(32, 16, 16)
                        buyTextColor = color.NRGBA{R: 128, G: 128, B: 128, A: 255}
                        buyTextIdle = color.NRGBA{R: 128, G: 128, B: 128, A: 255}
                    }

                    buy := widget.NewButton(
                        widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
                        widget.ButtonOpts.Image(ui.MakeButtonImage(buttonImage)),
                        widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                            Position: widget.RowLayoutPositionCenter,
                        })),
                        widget.ButtonOpts.Text("Buy", face, &widget.ButtonTextColor{
                            Idle: buyTextIdle,
                            Hover: color.White,
                            Pressed: buyTextColor,
                        }),
                        widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
                            if canBuy && cost <= playerObj.Money {
                                playerObj.Money -= cost
                                playerObj.KnownSpells.AddSpell(spell)
                                uiEvents.AddUpdate(&UIUpdateMoney{})
                                setupBox()
                            }
                        }),
                    )

                    box.AddChild(buy)
                }
            }

            setupBox()

            AddEvent(uiEvents, func (update *UIUpdateMagicBooks) {
                setupBox()
            })

            spellList.AddChild(box)
        }

        scroller := widget.NewScrollContainer(
            widget.ScrollContainerOpts.WidgetOpts(
                widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                    MaxHeight: containerSize,
                }),
            ),
            widget.ScrollContainerOpts.Content(spellList),
            widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
                Idle: ui.SolidImage(32, 32, 32),
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
                widget.WidgetOpts.MinSize(10, containerSize),
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

        both := ui.HBox()
        both.AddChild(scroller)
        both.AddChild(slider)
        tab.AddChild(both)
    }

    spellsTabs := widget.NewTabBook(
        widget.TabBookOpts.TabButtonImage(ui.MakeButtonImage(ui.SolidImage(64, 32, 32))),
        widget.TabBookOpts.TabButtonText(face, &widget.ButtonTextColor{
            Idle: color.White,
            Disabled: color.NRGBA{R: 32, G: 32, B: 32, A: 255},
            Hover: color.White,
            Pressed: color.White,
        }),
        widget.TabBookOpts.TabButtonSpacing(5),
        // widget.TabBookOpts.ContentPadding(widget.NewInsetsSimple(2)),
        widget.TabBookOpts.Tabs(tabs...),
    )

    shop.AddChild(spellsTabs)

    return shop
}

func makeShopUI(face *text.Face, imageCache *util.ImageCache, lbxCache *lbx.LbxCache, playerObj *player.Player, uiEvents *UIEventUpdate) *widget.Container {
    container := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewGridLayout(
            widget.GridLayoutOpts.Columns(2),
            widget.GridLayoutOpts.DefaultStretch(false, false),
            // widget.GridLayoutOpts.Stretch([]bool{false, false}, []bool{false, false}),
        )),
        /*
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(2),
            widget.RowLayoutOpts.Padding(widget.Insets{Top: 2, Bottom: 2, Left: 2, Right: 2}),
        )),
        */
        widget.ContainerOpts.BackgroundImage(ui.BorderedImage(color.RGBA{R: 0xc1, G: 0x80, B: 0x1a, A: 255}, 1)),
    )

    armyShop := ui.VBox()

    armyShop.AddChild(widget.NewText(
        widget.TextOpts.Text("Shop", face, color.White),
        widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionStart),
        widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
            Stretch: true,
        })),
    ))


    money := widget.NewText(
        // widget.TextOpts.Text(fmt.Sprintf("Money: %d", playerObj.Money), face, color.White),
        widget.TextOpts.Text(fmt.Sprintf("Money: %d", playerObj.Money), face, color.White),
    )

    AddEvent(uiEvents, func (update *UIUpdateMoney) {
        money.Label = fmt.Sprintf("Money: %d", playerObj.Money)
    })

    armyShop.AddChild(makeMoneyText(money, imageCache))

    container2 := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
            widget.RowLayoutOpts.Spacing(4),
        )),
    )

    armyShop.AddChild(container2)

    buyUnit := func(unit *units.Unit) {
        unitCost := getUnitCost(unit)
        if unitCost <= playerObj.Money {
            playerObj.Money -= unitCost
            newUnit := playerObj.AddUnit(*unit)
            uiEvents.AddUpdate(&UIAddUnit{Unit: newUnit})
            uiEvents.AddUpdate(&UIUpdateMoney{})
        }
    }

    unitList := MakeUnitIconList("All Units", imageCache, face, buyUnit)

    for _, unit := range getValidChoices(100000) {
        unitList.AddUnit(unit)
    }

    unitList.SortByName()

    filteredUnitList := MakeUnitIconList("Affordable Units", imageCache, face, buyUnit)

    setupFilteredList := func() {
        filteredUnitList.Clear()
        filteredUnitList.Reset()
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

    container.AddChild(armyShop)

    magicShop := makeMagicShop(face, imageCache, lbxCache, playerObj, uiEvents)
    container.AddChild(magicShop)

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

func makeBuyEnchantments(unit units.StackUnit, face *text.Face, playerObj *player.Player, uiEvents *UIEventUpdate, imageCache *util.ImageCache) *widget.Container {
    enchantments := []data.UnitEnchantment{
        data.UnitEnchantmentGiantStrength,
        data.UnitEnchantmentLionHeart,
        data.UnitEnchantmentHaste,
        data.UnitEnchantmentImmolation,
        data.UnitEnchantmentResistElements,
        data.UnitEnchantmentResistMagic,
        data.UnitEnchantmentElementalArmor,
        data.UnitEnchantmentBless,
        data.UnitEnchantmentRighteousness,
        data.UnitEnchantmentCloakOfFear,
        data.UnitEnchantmentTrueSight,
        data.UnitEnchantmentPathFinding,
        data.UnitEnchantmentFlight,
        data.UnitEnchantmentChaosChannelsDemonWings,
        data.UnitEnchantmentChaosChannelsDemonSkin,
        data.UnitEnchantmentChaosChannelsFireBreath,
        data.UnitEnchantmentEndurance,
        data.UnitEnchantmentHeroism,
        data.UnitEnchantmentHolyArmor,
        data.UnitEnchantmentHolyWeapon,
        data.UnitEnchantmentInvulnerability,
        data.UnitEnchantmentIronSkin,
        data.UnitEnchantmentRegeneration,
        data.UnitEnchantmentStoneSkin,
        data.UnitEnchantmentGuardianWind,
        data.UnitEnchantmentInvisibility,
        data.UnitEnchantmentMagicImmunity,
        data.UnitEnchantmentSpellLock,
        data.UnitEnchantmentEldritchWeapon,
        data.UnitEnchantmentFlameBlade,
        data.UnitEnchantmentBerserk,
        data.UnitEnchantmentBlackChannels,
        data.UnitEnchantmentWraithForm,
    }

    slices.SortFunc(enchantments, func(a, b data.UnitEnchantment) int {
        return cmp.Compare(a.Name(), b.Name())
    })

    enchantmentList := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewGridLayout(
            widget.GridLayoutOpts.Columns(2),
        )),
    )

    containerSize := 250

    scroller := widget.NewScrollContainer(
        widget.ScrollContainerOpts.WidgetOpts(
            widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                MaxHeight: containerSize,
            }),
        ),
        widget.ScrollContainerOpts.Content(enchantmentList),
        widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
            Idle: ui.SolidImage(32, 32, 32),
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
            widget.WidgetOpts.MinSize(10, containerSize),
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

    for _, enchantment := range enchantments {
        border := ui.BorderedImage(color.RGBA{R: 128, G: 128, B: 128, A: 255}, 1)
        box := ui.VBox(
            widget.ContainerOpts.BackgroundImage(border),
        )
        name := ui.CenteredText(enchantment.Name(), face, enchantment.Color())
        box.AddChild(name)
        money := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("%d", getEnchantmentCost(enchantment)), face, color.White))

        box.AddChild(makeMoneyText(money, imageCache, widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
            Position: widget.RowLayoutPositionCenter,
        }))))

        remove := func(){}
        enchantButton := widget.NewButton(
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(ui.MakeButtonImage(ui.SolidImage(64, 32, 32))),
            widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                Position: widget.RowLayoutPositionCenter,
            })),
            widget.ButtonOpts.Text("Enchant", face, &widget.ButtonTextColor{
                Idle: color.White,
                Hover: color.White,
                Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            }),
            widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
                cost := uint64(getEnchantmentCost(enchantment))
                if cost <= playerObj.Money && !unit.HasEnchantment(enchantment) {
                    playerObj.Money -= cost
                    unit.AddEnchantment(enchantment)
                    remove()
                    uiEvents.AddUpdate(&UIUpdateMoney{})
                    uiEvents.AddUpdate(&UIUpdateUnit{Unit: unit})
                }
            }),
        )
        box.AddChild(enchantButton)
        remove = enchantmentList.AddChild(box)
    }

    container := ui.HBox()
    container.AddChild(scroller)
    container.AddChild(slider)

    return container
}

func makeUnitInfoUI(face *text.Face, allUnits []units.StackUnit, playerObj *player.Player, uiEvents *UIEventUpdate, imageCache *util.ImageCache) *widget.Container {

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

    removeUnit := func() {
    }

    updateUnitSpecifics = func(unit units.StackUnit, setup func()) {
        currentName := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Name: %v", unit.GetFullName()), face, color.White))
        currentHealth := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("HP: %d/%d", unit.GetHealth(), unit.GetMaxHealth()), face, color.White))
        currentRace := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Race: %v", unit.GetRace()), face, color.White))

        removeUnit()
        unitSpecifics.RemoveChildren()
        unitSpecifics.AddChild(currentName)
        unitSpecifics.AddChild(currentHealth)
        unitSpecifics.AddChild(currentRace)
        unitSpecifics.AddChild(widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Experience: %d (%v)", unit.GetExperience(), unit.GetExperienceLevel().Name()), face, color.White)))

        // var currentHealTarget units.StackUnit

        makeHealCost := func(amount int) *widget.Container {
            gold, _ := imageCache.GetImageTransform("backgrnd.lbx", 42, 0, "enlarge", enlargeTransform(2))
            heart, _ := imageCache.GetImageTransform("unitview.lbx", 23, 0, "enlarge", enlargeTransform(2))
            healText := widget.NewText(widget.TextOpts.Text("Heal", face, color.White))
            heartText := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("%d", amount), face, color.White))
            forText := widget.NewText(widget.TextOpts.Text("for", face, color.White))
            goldText := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("%d", getHealCost(unit, amount)), face, color.White))

            centered := widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                Position: widget.RowLayoutPositionCenter,
            })

            makeIcon := func(image *ebiten.Image) *widget.Graphic {
                return widget.NewGraphic(widget.GraphicOpts.Image(image), widget.GraphicOpts.WidgetOpts(centered))
            }

            return combineHorizontalElements(healText, makeIcon(heart), heartText, forText, makeIcon(gold), goldText)
        }

        healContainer := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout()))
        healContainer.AddChild(makeHealCost(unit.GetDamage()))

        // healCost := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Heal %d hp for %d gold", unit.GetDamage(), getHealCost(unit, unit.GetDamage())), face, color.White))

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
                healContainer.RemoveChildren()
                healContainer.AddChild(makeHealCost(args.Slider.Current))
                // healCost.Label = fmt.Sprintf("Heal %d hp for %d gold", args.Slider.Current, getHealCost(unit, args.Slider.Current))
            }),
        )

        healButton := widget.NewButton(
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
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

        unitSpecifics.AddChild(healContainer)
        unitSpecifics.AddChild(healSlider)
        unitSpecifics.AddChild(healButton)

        enchantments := widget.NewList(
            widget.ListOpts.ScrollContainerImage(&widget.ScrollContainerImage{
                Idle: ui.SolidImage(64, 64, 64),
                Disabled: ui.SolidImage(32, 32, 32),
                Mask: ui.SolidImage(32, 32, 32),
            }),
            widget.ListOpts.SliderParams(&widget.SliderParams{
                TrackImage: &widget.SliderTrackImage{
                    Idle: ui.SolidImage(64, 64, 64),
                    Hover: ui.SolidImage(96, 96, 96),
                },
                HandleImage: ui.MakeButtonImage(ui.SolidImage(192, 192, 192)),
            }),
            widget.ListOpts.HideHorizontalSlider(),
            widget.ListOpts.EntryFontFace(face),
            widget.ListOpts.EntryColor(&widget.ListEntryColor{
                Selected: color.White,
                Unselected: color.White,
            }),
            widget.ListOpts.EntryLabelFunc(func (data any) string {
                return data.(string)
            }),
            widget.ListOpts.EntryTextPadding(widget.NewInsetsSimple(2)),
            widget.ListOpts.EntryTextPosition(widget.TextPositionCenter, widget.TextPositionCenter),
            widget.ListOpts.EntrySelectedHandler(func (args *widget.ListEntrySelectedEventArgs) {
            }),
        )

        for _, enchantment := range unit.GetEnchantments() {
            enchantments.AddEntry(enchantment.Name())
        }

        newEvent := func (update *UIUpdateUnit) {
            if update.Unit == unit {
                enchantments.SetEntries(nil)
                for _, enchantment := range unit.GetEnchantments() {
                    enchantments.AddEntry(enchantment.Name())
                }
            }
        }

        removeId := AddEvent(uiEvents, newEvent)

        removeUnit = func() {
            uiEvents.Remove(removeId)
        }

        enchantmentsBoxes := ui.HBox()

        showEnchantments := ui.VBox()

        showEnchantments.AddChild(widget.NewText(widget.TextOpts.Text("Enchantments", face, color.RGBA{R: 255, G: 255, B: 0, A: 255})))
        showEnchantments.AddChild(enchantments)

        enchantmentsBoxes.AddChild(showEnchantments)

        buyEnchantments := makeBuyEnchantments(unit, face, playerObj, uiEvents, imageCache)

        enchantmentsBoxes.AddChild(buyEnchantments)

        unitSpecifics.AddChild(enchantmentsBoxes)
    }

    unitList := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewGridLayout(
            widget.GridLayoutOpts.Columns(2),
            widget.GridLayoutOpts.DefaultStretch(false, false),
            // widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{false, false}),
        )),
    )

    var lastBox *widget.Container

    addUnit := func(unit units.StackUnit) {
        var unitBox *widget.Container
        var setup func()

        border := ui.BorderedImage(color.RGBA{R: 128, G: 128, B: 128, A: 255}, 1)

        unitBox = ui.VBox(widget.ContainerOpts.WidgetOpts(
            widget.WidgetOpts.MouseButtonPressedHandler(func(args *widget.WidgetMouseButtonPressedEventArgs) {
                updateUnitSpecifics(unit, setup)
                if lastBox != nil {
                    lastBox.SetBackgroundImage(border)
                }
                unitBox.SetBackgroundImage(ui.SolidImage(96, 96, 32))
                lastBox = unitBox
            }),
            ),
            widget.ContainerOpts.BackgroundImage(border),
        )

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
            widget.RowLayoutOpts.Padding(&widget.Insets{Top: 2, Bottom: 2, Left: 2, Right: 2}),
        )),
        widget.ContainerOpts.BackgroundImage(ui_image.NewNineSliceColor(color.NRGBA{R: 48, G: 48, B: 48, A: 255})),
    )

    armyInfo.AddChild(widget.NewText(
        widget.TextOpts.Text("Army", face, color.White),
        widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionStart),
        widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
            Stretch: true,
        })),
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

func makePlayerInfoUI(face *text.Face, playerObj *player.Player, events *UIEventUpdate, imageCache *util.ImageCache) *widget.Container {
    container := ui.HBox()

    name := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Name: %v", playerObj.Wizard.Name), face, color.White))
    container.AddChild(name)

    level := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Level: %d", playerObj.Level), face, color.White))
    container.AddChild(level)

    mana := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Mana: %d", playerObj.Mana), face, color.White))
    container.AddChild(mana)

    books := ui.HBox()
    container.AddChild(books)

    lifeBook, _ := imageCache.GetImageTransform("newgame.lbx", 24, 0, "enlarge", enlargeTransform(2))
    sorceryBook, _ := imageCache.GetImageTransform("newgame.lbx", 27, 0, "enlarge", enlargeTransform(2))
    natureBook, _ := imageCache.GetImageTransform("newgame.lbx", 30, 0, "enlarge", enlargeTransform(2))
    deathBook, _ := imageCache.GetImageTransform("newgame.lbx", 33, 0, "enlarge", enlargeTransform(2))
    chaosBook, _ := imageCache.GetImageTransform("newgame.lbx", 36, 0, "enlarge", enlargeTransform(2))

    setupBooks := func() {
        books.RemoveChildren()

        for _, book := range playerObj.GetWizard().Books {
            count := book.Count
            var bookImage *ebiten.Image
            switch book.Magic {
                case data.LifeMagic: bookImage = lifeBook
                case data.SorceryMagic: bookImage = sorceryBook
                case data.NatureMagic: bookImage = natureBook
                case data.DeathMagic: bookImage = deathBook
                case data.ChaosMagic: bookImage = chaosBook
            }

            for range count {
                books.AddChild(widget.NewGraphic(
                    widget.GraphicOpts.Image(bookImage),
                    widget.GraphicOpts.WidgetOpts(
                        widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                            Position: widget.RowLayoutPositionCenter,
                        }),
                    ),
                ))
            }
        }
    }

    setupBooks()

    AddEvent(events, func (update *UIUpdateMana) {
        mana.Label = fmt.Sprintf("Mana: %d", playerObj.Mana)
    })

    AddEvent(events, func (update *UIUpdateMagicBooks) {
        setupBooks()
    })

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

    var face1 text.Face = &face

    uiEvents := MakeUIEventUpdate()

    rootContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(4),
            widget.RowLayoutOpts.Padding(&widget.Insets{Top: 4, Bottom: 4, Left: 4, Right: 4}),
        )),
        widget.ContainerOpts.BackgroundImage(ui_image.NewNineSliceColor(color.NRGBA{R: 32, G: 32, B: 32, A: 255})),
    )

    newGameButton := widget.NewButton(
        widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Image(ui.MakeButtonImage(ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 32, B: 32, A: 255}))),
        widget.ButtonOpts.Text("Enter Battle", &face1, &widget.ButtonTextColor{
            Idle: color.White,
            Hover: color.White,
            Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
        widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
            if len(engine.Player.Units) == 0 {
                return
            }
            select {
                case engine.Events <- &EventNewGame{}:
                default:
            }
        }),
    )

    imageCache := util.MakeImageCache(engine.Cache)

    rootContainer.AddChild(newGameButton)

    rootContainer.AddChild(makePlayerInfoUI(&face1, engine.Player, uiEvents, &imageCache))

    unitInfoUI := makeUnitInfoUI(&face1, engine.Player.Units, engine.Player, uiEvents, &imageCache)

    rootContainer.AddChild(unitInfoUI)
    rootContainer.AddChild(makeShopUI(&face1, &imageCache, engine.Cache, engine.Player, uiEvents))

    ui := &ebitenui.UI{
        Container: rootContainer,
    }

    return ui, uiEvents, nil
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
    playerObj.Money = 30000
}

func test4(playerObj *player.Player) {
    for range 5 {
        playerObj.AddUnit(units.LizardSwordsmen)
    }
    for range 5 {
        playerObj.AddUnit(units.Warlocks)
    }
    playerObj.Money = 3000
}

func MakeEngine(cache *lbx.LbxCache) *Engine {
    playerObj := player.MakePlayer(data.BannerGreen)

    // test1(playerObj)
    // test3(playerObj)
    // test4(playerObj)

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

    ebiten.SetWindowSize(1200, 1050)
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    engine := MakeEngine(cache)

    err := ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
