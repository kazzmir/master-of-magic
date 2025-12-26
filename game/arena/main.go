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
    musiclib "github.com/kazzmir/master-of-magic/game/magic/music"

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
    GameModeNewGameUI
    GameModeBattle
)

type EngineEvents interface {
}

type EventNewGame struct {
    Level Difficulty
}

type EventEnterBattle struct {
}

type Difficulty int
const (
    DifficultyEasy Difficulty = iota
    DifficultyNormal
    DifficultyHard
    DifficultyImpossible
)

func (d Difficulty) String() string {
    switch d {
        case DifficultyEasy: return "Easy"
        case DifficultyNormal: return "Normal"
        case DifficultyHard: return "Hard"
        case DifficultyImpossible: return "Impossible"
    }

    return "?"
}

type PlaySound struct {
    Maker audio.MakePlayerFunc
}

func (play *PlaySound) Play() {
    if play.Maker != nil {
        sound := play.Maker()
        sound.Play()
    }
}

type Engine struct {
    GameMode GameMode
    Player *player.Player
    Cache *lbx.LbxCache

    Events chan EngineEvents
    Music *musiclib.Music

    Drawers []func(screen *ebiten.Image)

    Difficulty Difficulty
    CombatCoroutine *coroutine.Coroutine
    CombatScreen *combat.CombatScreen
    CurrentBattleReward uint64

    LeftClickSound *PlaySound

    UI *ebitenui.UI
    NewGameUI *ebitenui.UI
    UIUpdates *UIEventUpdate
}

var CombatDoneErr = errors.New("combat done")

func getValidChoices(budget uint64) []*units.Unit {
    var choices []*units.Unit

    for i := range units.AllUnits {
        choice := &units.AllUnits[i]
        if choice.Race == data.RaceHero || choice.IsSettlers() {
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

func randomChoose[T any](choices ...T) T {
    return choices[rand.N(len(choices))]
}

func getValidUnitEnchantments() []data.UnitEnchantment {
    return []data.UnitEnchantment{
        data.UnitEnchantmentGiantStrength,
        data.UnitEnchantmentLionHeart,
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
        data.UnitEnchantmentBlackChannels,
        data.UnitEnchantmentWraithForm,
    }
}

func (engine *Engine) PushDrawer(drawer func(screen *ebiten.Image)) {
    engine.Drawers = append(engine.Drawers, drawer)
}

func (engine *Engine) PopDrawer() {
    if len(engine.Drawers) > 0 {
        engine.Drawers = engine.Drawers[:len(engine.Drawers)-1]
    }
}

func setupAISpells(enemyPlayer *player.Player, lbxCache *lbx.LbxCache, budget uint64) uint64 {
    var totalCosts uint64

    allSpells, err := spellbook.ReadSpellsFromCache(lbxCache)
    if err != nil {
        return 0
    }

    doSpell := func(counter *int, numSpells func(int) int, spell spellbook.Spell){
        if *counter < numSpells(enemyPlayer.GetWizard().MagicLevel(spell.Magic)) && budget >= uint64(spell.ResearchCost) {
            enemyPlayer.KnownSpells.AddSpell(spell)
            budget -= uint64(spell.ResearchCost)
            totalCosts += uint64(spell.ResearchCost)
            *counter += 1
        }
    }

    commonCount := 0
    uncommonCount := 0
    rareCount := 0
    veryRareCount := 0

    for _, spell := range allSpells.Spells {
        if !spell.Eligibility.CanCastInCombat(false) {
            continue
        }

        if rand.N(10) > 5 {
            switch spell.Rarity {
                case spellbook.SpellRarityCommon: doSpell(&commonCount, getCommonSpells, spell)
                case spellbook.SpellRarityUncommon: doSpell(&uncommonCount, getUncommonSpells, spell)
                case spellbook.SpellRarityRare: doSpell(&rareCount, getRareSpells, spell)
                case spellbook.SpellRarityVeryRare: doSpell(&veryRareCount, getVeryRareSpells, spell)
            }
        }
    }

    return totalCosts
}

// returns the new budget and how much was spent on enchantments
func tryAddEnchantments(unit units.StackUnit, playerObj *player.Player, budget uint64) uint64 {
    enchantments := getValidUnitEnchantments()

    numEnchantments := rand.N(2 * playerObj.Level)
    used := 0
    var totalCost uint64

    for {
        if used >= numEnchantments {
            break
        }

        if len(enchantments) == 0 {
            break
        }

        choice := enchantments[rand.N(len(enchantments))]

        // log.Printf("Considering adding enchantment %v to unit %v. budget=%v cost=%v %v/%v", choice, unit.GetFullName(), budget, totalCost, used, numEnchantments)

        requirements := getEnchantmentRequirements(choice)
        cost := uint64(getEnchantmentCost(choice))

        if playerObj.GetWizard().MagicLevel(requirements.Magic) < requirements.Count || cost > budget {

            enchantments = slices.DeleteFunc(enchantments, func(e data.UnitEnchantment) bool {
                return e == choice
            })

            continue
        }

        unit.AddEnchantment(choice)

        enchantments = slices.DeleteFunc(enchantments, func(e data.UnitEnchantment) bool {
            return e == choice
        })

        used += 1
        budget -= cost
        totalCost += cost
    }

    return totalCost
}

func tryUpgradeUnit(unit units.StackUnit, budget uint64) uint64 {
    upgradeCost := uint64(0)

    if unit.GetRace() == data.RaceFantastic {
        return upgradeCost
    }

    for unit.GetWeaponBonus() != data.WeaponAdamantium {

        cost := getWeaponUpgradeCost(unit.GetWeaponBonus())

        if cost > budget || rand.N(100) > 60 {
            break
        }

        switch unit.GetWeaponBonus() {
            case data.WeaponNone:
                unit.SetWeaponBonus(data.WeaponMagic)
            case data.WeaponMagic:
                unit.SetWeaponBonus(data.WeaponMythril)
            case data.WeaponMythril:
                unit.SetWeaponBonus(data.WeaponAdamantium)
        }

        budget -= cost
        upgradeCost += cost
    }

    return upgradeCost
}

func (engine *Engine) MakeBattleFunc() coroutine.AcceptYieldFunc {
    defendingArmy := combat.Army {
        Player: engine.Player,
    }

    for _, unit := range engine.Player.Units {
        defendingArmy.AddUnit(unit)
    }

    enemyPlayer := player.MakeAIPlayer(data.BannerRed)

    enemyPlayer.Level = engine.Player.Level
    enemyPlayer.Mana = engine.Player.Level * 10 + rand.N(15) - 7
    enemyPlayer.OriginalMana = enemyPlayer.Mana

    engine.Player.Mana = engine.Player.OriginalMana

    booksMax := min(12, engine.Player.Level * 2)

    for _, magic := range []data.MagicType{data.LifeMagic, data.SorceryMagic, data.NatureMagic, data.DeathMagic, data.ChaosMagic} {
        enemyPlayer.GetWizard().AddMagicLevel(magic, rand.N(booksMax))
    }

    var baseBudget float64 = 100
    var baseExponent float64 = 1.8

    switch engine.Difficulty {
        case DifficultyEasy:
            baseBudget = 80
            baseExponent = 1.5
        case DifficultyNormal:
            baseBudget = 100
            baseExponent = 1.8
        case DifficultyHard:
            baseBudget = 120
            baseExponent = 1.95
        case DifficultyImpossible:
            baseBudget = 150
            baseExponent = 2.1
    }

    budget := uint64(baseBudget * math.Pow(baseExponent, float64(engine.Player.Level)))

    var spellCosts uint64

    log.Printf("Starting budget: %v level=%v", budget, engine.Player.Level)

    spellCosts = setupAISpells(enemyPlayer, engine.Cache, budget / 2)

    budget -= spellCosts

    log.Printf("Enemy magic: %v", enemyPlayer.GetWizard().Books)
    log.Printf("Enemy spells: %v", enemyPlayer.GetKnownSpells().Spells)
    log.Printf("Enemy mana: %v", enemyPlayer.Mana)

    engine.CurrentBattleReward = spellCosts

    // at least have some amount to buy units
    if budget < 80 {
        budget = 80
    }

    // log.Printf("Budget after spells: %v", budget)

    for budget > 0 {
        choices := getValidChoices(budget)

        if len(choices) == 0 {
            break
        }

        slices.SortFunc(choices, func(a, b *units.Unit) int {
            return cmp.Compare(getUnitCost(b), getUnitCost(a))
        })

        weight := rand.N(100)
        var choice *units.Unit
        var start, end int
        // choose from first 1/3rd of array if weight < 70
        if weight < 70 {
            start = 0
            end = len(choices) / 3
        } else if weight < 90 {
            // choose from 2nd 1/3rd of array
            start = len(choices) / 3
            end = 2 * len(choices) / 3
        } else {
            // choose from last 1/3rd of array
            start = 2 * len(choices) / 3
            end = len(choices)
        }

        if end <= start {
            end = start + 1
        }

        sub := choices[start:end]

        if len(sub) == 0 {
            sub = choices
        }

        choice = sub[rand.N(len(sub))]

        addedUnit := enemyPlayer.AddUnit(*choice)

        unitCost := getUnitCost(choice)
        budget -= unitCost

        var enchantmentCost uint64
        enchantmentCost = tryAddEnchantments(addedUnit, enemyPlayer, budget / 2)

        budget -= enchantmentCost

        upgradeCost := tryUpgradeUnit(addedUnit, budget / 2)

        budget -= upgradeCost

        engine.CurrentBattleReward += enchantmentCost
        engine.CurrentBattleReward += unitCost
        engine.CurrentBattleReward += upgradeCost

        if budget > 10000000 {
            panic("budget overflow")
        }
    }

    switch engine.Difficulty {
        case DifficultyEasy: engine.CurrentBattleReward = engine.CurrentBattleReward * 2 + 200
        case DifficultyNormal: engine.CurrentBattleReward = (engine.CurrentBattleReward * 3) / 2 + 100
        case DifficultyHard: engine.CurrentBattleReward = (engine.CurrentBattleReward * 5) / 4
        case DifficultyImpossible: engine.CurrentBattleReward = (engine.CurrentBattleReward * 9) / 10
    }

    attackingArmy := combat.Army {
        Player: enemyPlayer,
    }

    for _, unit := range enemyPlayer.Units {
        attackingArmy.AddUnit(unit)
    }

    landscape := randomChoose(combat.CombatLandscapeGrass, combat.CombatLandscapeDesert, combat.CombatLandscapeMountain, combat.CombatLandscapeTundra)

    screen := combat.MakeCombatScreen(engine.Cache, &defendingArmy, &attackingArmy, engine.Player, landscape, data.PlaneArcanus, combat.ZoneType{}, data.MagicNone, 0, 0)
    engine.CombatScreen = screen

    return func(yield coroutine.YieldFunc) error {
        engine.Music.PushSong(randomChoose(musiclib.SongCombat1, musiclib.SongCombat2))
        defer engine.Music.PopSong()

        for screen.Update(yield) == combat.CombatStateRunning {
            yield()
        }

        var endScreen *combat.CombatEndScreen

        lastState := screen.Update(yield)
        if lastState == combat.CombatStateAttackerWin || lastState == combat.CombatStateDefenderFlee {
            endScreen = combat.MakeCombatEndScreen(engine.Cache, combat.CombatEndScreenResultLose, 0, 0, 0, 0)
            engine.Music.PushSong(musiclib.SongYouLose)
        } else if lastState == combat.CombatStateDefenderWin {
            endScreen = combat.MakeCombatEndScreen(engine.Cache, combat.CombatEndScreenResultWin, 0, 0, 0, 0)
            engine.Music.PushSong(musiclib.SongYouWin)
        }

        defer engine.Music.PopSong()

        engine.PushDrawer(func(screen *ebiten.Image) {
            endScreen.Draw(screen)
        })

        defer engine.PopDrawer()

        for endScreen.Update() == combat.CombatEndScreenRunning {
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
        case GameModeNewGameUI:
            engine.NewGameUI.Update()

            select {
                case event := <-engine.Events:
                    switch use := event.(type) {
                        case *EventNewGame:
                            engine.GameMode = GameModeUI
                            engine.Difficulty = use.Level
                            engine.Player = player.MakePlayer(data.BannerGreen)
                            // test3(engine.Player)

                            if engine.Difficulty == DifficultyEasy {
                                engine.Player.Money = 500
                            }

                            var err error
                            engine.UI, engine.UIUpdates, err = engine.MakeUI()
                            if err != nil {
                                log.Printf("Error creating UI: %v", err)
                            }

                    }
                default:
            }

        case GameModeUI:
            engine.UI.Update()
            engine.UIUpdates.Update()

            select {
                case event := <-engine.Events:
                    switch event.(type) {
                        case *EventEnterBattle:
                            engine.GameMode = GameModeBattle
                            engine.CombatCoroutine = coroutine.MakeCoroutine(engine.MakeBattleFunc())
                    }
                default:
            }
        case GameModeBattle:
            err := engine.CombatCoroutine.Run()
            if errors.Is(err, CombatDoneErr) {
                lastState := engine.CombatScreen.Model.FinishState

                engine.CombatCoroutine = nil
                engine.CombatScreen = nil

                engine.Player.Level += 1
                engine.Player.Money += engine.CurrentBattleReward

                var aliveUnits []units.StackUnit
                for _, unit := range engine.Player.Units {
                    if unit.GetHealth() > 0 {
                        aliveUnits = append(aliveUnits, unit)
                    }
                }

                engine.Player.Units = aliveUnits
                if lastState != combat.CombatStateDefenderWin {
                    log.Printf("All units lost, starting new game")
                    engine.GameMode = GameModeNewGameUI
                    engine.Player = player.MakePlayer(data.BannerGreen)
                    // engine.Player.AddUnit(units.LizardSwordsmen)
                } else {
                    for _, unit := range engine.Player.Units {
                        unit.AddExperience(20)
                    }
                    engine.GameMode = GameModeUI
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
    if len(engine.Drawers) > 0 {
        last := engine.Drawers[len(engine.Drawers)-1]
        last(screen)
    }
}

func (engine *Engine) DefaultDraw(screen *ebiten.Image) {
    switch engine.GameMode {
        case GameModeNewGameUI:
            engine.DrawUI(screen)
            vector.FillRect(screen, 0, 0, float32(screen.Bounds().Dx()), float32(screen.Bounds().Dy()), color.NRGBA{R: 0, G: 0, B: 0, A: 180}, true)
            engine.NewGameUI.Draw(screen)
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

func resizeImage(old *ebiten.Image, amount float64) *ebiten.Image {
    width, height := old.Bounds().Dx(), old.Bounds().Dy()
    newW, newH := int(float64(width) * amount), int(float64(height) * amount)
    newImage := ebiten.NewImage(newW, newH)
    op := &ebiten.DrawImageOptions{}
    op.GeoM.Scale(float64(newW)/float64(width), float64(newH)/float64(height))
    newImage.DrawImage(old, op)
    return newImage
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

type UIRemoveUnit struct {
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
    playSound *PlaySound

    SortNameDirection SortDirection
    SortCostDirection SortDirection
}

func standardButtonImage() *widget.ButtonImage {
    body := color.NRGBA{R: 64, G: 32, B: 32, A: 255}
    border := color.NRGBA{R: 32, G: 16, B: 16, A: 255}
    return ui.MakeButtonImage(ui_image.NewBorderedNineSliceColor(body, border, 1))
}

func MakeUnitIconList(description string, imageCache *util.ImageCache, face *text.Face, buyUnit func(*units.Unit), playSound *PlaySound) *UnitIconList {
    var iconList UnitIconList

    iconList.imageCache = imageCache
    iconList.buyUnit = buyUnit
    iconList.face = face
    iconList.playSound = playSound

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
        widget.SliderOpts.Orientation(widget.DirectionVertical),
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

    upArrow, _ := imageCache.GetImageTransform("resource.lbx", 32, 0, "enlarge", enlargeTransform(2))
    downArrow, _ := imageCache.GetImageTransform("resource.lbx", 33, 0, "enlarge", enlargeTransform(2))

    sortButtons := ui.HBox()

    const SortByName = 0
    const SortByCost = 1

    // only change how we sort if the same button is pressed twice in a row
    lastSort := SortByName

    var sortNameNone *widget.Button
    var sortNameUp *widget.Button
    var sortNameDown *widget.Button

    var sortCostNone *widget.Button
    var sortCostUp *widget.Button
    var sortCostDown *widget.Button

    var currentSortNameButton *widget.Button
    var currentSortCostButton *widget.Button

    // swap the button for the given sort kind to the one based on the current direction
    // the other sort kind is set to the button without an arrow
    updateSortButton := func(sortKind int) {
        var newButton *widget.Button
        switch sortKind {
            case SortByName:
                switch iconList.SortNameDirection {
                    case SortDirectionAscending: newButton = sortNameUp
                    case SortDirectionDescending: newButton = sortNameDown
                    default:
                        panic("sort button error")
                }

                sortButtons.ReplaceChild(currentSortNameButton, newButton)
                currentSortNameButton = newButton

                sortButtons.ReplaceChild(currentSortCostButton, sortCostNone)
                currentSortCostButton = sortCostNone

            case SortByCost:
                switch iconList.SortCostDirection {
                    case SortDirectionAscending: newButton = sortCostUp
                    case SortDirectionDescending: newButton = sortCostDown
                    default:
                        panic("sort button error")
                }

                sortButtons.ReplaceChild(currentSortCostButton, newButton)
                currentSortCostButton = newButton

                sortButtons.ReplaceChild(currentSortNameButton, sortNameNone)
                currentSortNameButton = sortNameNone
        }
    }

    makeSortButton := func(text string, image *ebiten.Image, clicked func()) *widget.Button {
        opts := []widget.ButtonOpt{
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(standardButtonImage()),
            widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
                playSound.Play()
                clicked()
            }),
        }

        buttonColor := widget.ButtonTextColor{
            Idle: color.White,
            Hover: color.White,
            Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }

        if image != nil {
            opts = append(opts, widget.ButtonOpts.TextAndImage(text, face, &widget.GraphicImage{Idle: image, Disabled: image}, &buttonColor))
        } else {
            opts = append(opts, widget.ButtonOpts.Text(text, face, &buttonColor))
        }

        return widget.NewButton(opts...)
    }

    makeSortNameButton := func(image *ebiten.Image) *widget.Button {
        return makeSortButton("Sort by Name", image, func() {
            if lastSort == SortByName {
                iconList.SortNameDirection = iconList.SortNameDirection.Next()
            }
            lastSort = SortByName
            iconList.SortByName()
            updateSortButton(SortByName)
        })
    }

    makeSortCostButton := func(image *ebiten.Image) *widget.Button {
        return makeSortButton("Sort by Cost", image, func() {
            if lastSort == SortByCost {
                iconList.SortCostDirection = iconList.SortCostDirection.Next()
            }
            lastSort = SortByCost
            iconList.SortByCost()
            updateSortButton(SortByCost)
        })
    }

    sortNameNone = makeSortNameButton(nil)
    sortNameUp = makeSortNameButton(upArrow)
    sortNameDown = makeSortNameButton(downArrow)

    currentSortNameButton = sortNameNone

    sortCostNone = makeSortCostButton(nil)
    sortCostUp = makeSortCostButton(upArrow)
    sortCostDown = makeSortCostButton(downArrow)

    currentSortCostButton = sortCostNone

    sortButtons.AddChild(currentSortNameButton)
    sortButtons.AddChild(currentSortCostButton)

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

    heart, _ := iconList.imageCache.GetImageTransform("unitview.lbx", 23, 0, "enlarge", enlargeTransform(2))

    meleeImage, _ := iconList.imageCache.GetImageTransform("unitview.lbx", 13, 0, "enlarge", enlargeTransform(2))
    rangeMagicImage, _ := iconList.imageCache.GetImageTransform("unitview.lbx", 14, 0, "enlarge", enlargeTransform(2))
    rangeBowImage, _ := iconList.imageCache.GetImageTransform("unitview.lbx", 18, 0, "enlarge", enlargeTransform(2))
    rangeBoulderImage, _ := iconList.imageCache.GetImageTransform("unitview.lbx", 19, 0, "enlarge", enlargeTransform(2))
    defenseImage, _ := iconList.imageCache.GetImageTransform("unitview.lbx", 22, 0, "enlarge", enlargeTransform(2))

    walkingImage, _ := iconList.imageCache.GetImageTransform("unitview.lbx", 24, 0, "enlarge", enlargeTransform(2))
    flyingImage, _ := iconList.imageCache.GetImageTransform("unitview.lbx", 25, 0, "enlarge", enlargeTransform(2))
    swimmingImage, _ := iconList.imageCache.GetImageTransform("unitview.lbx", 26, 0, "enlarge", enlargeTransform(2))

    resistanceImage, _ := iconList.imageCache.GetImageTransform("unitview.lbx", 27, 0, "enlarge", enlargeTransform(2))

    unitImage, err := iconList.imageCache.GetImageTransform(unit.GetCombatLbxFile(), unit.GetCombatIndex(units.FacingRight), 0, "enlarge", enlargeTransform(2))
    if err == nil {
        unitDetails := ui.HBox(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
            Position: widget.RowLayoutPositionCenter,
        })))

        unitBox.AddChild(unitDetails)

        unitDetails.AddChild(widget.NewGraphic(
            widget.GraphicOpts.Image(unitImage),
            /*
            widget.GraphicOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                Position: widget.RowLayoutPositionStart,
            })),
            */
        ))

        stats := widget.NewContainer(
            widget.ContainerOpts.Layout(widget.NewGridLayout(
                widget.GridLayoutOpts.Columns(2),
                widget.GridLayoutOpts.Spacing(4, 2),
            )),
        )

        unitDetails.AddChild(stats)

        centered := widget.WidgetOpts.LayoutData(widget.RowLayoutData{
            Position: widget.RowLayoutPositionCenter,
        })

        makeIcon := func(image *ebiten.Image) *widget.Graphic {
            return widget.NewGraphic(widget.GraphicOpts.Image(resizeImage(image, 0.90)), widget.GraphicOpts.WidgetOpts(centered))
        }

        makeText := func(text string) *widget.Text {
            return widget.NewText(widget.TextOpts.Text(text, iconList.face, color.White))
        }

        stats.AddChild(combineHorizontalElements(makeIcon(heart), makeText(fmt.Sprintf("%d", unit.GetHitPoints()))))

        moves := unit.GetMovementSpeed(true)
        moveImage := walkingImage
        if unit.Flying {
            moveImage = flyingImage
        } else if unit.Swimming {
            moveImage = swimmingImage
        }
        stats.AddChild(combineHorizontalElements(makeIcon(moveImage), makeText(fmt.Sprintf("%d", moves))))

        stats.AddChild(combineHorizontalElements(makeIcon(defenseImage), makeText(fmt.Sprintf("%d", unit.GetDefense()))))
        stats.AddChild(combineHorizontalElements(makeIcon(meleeImage), makeText(fmt.Sprintf("%d", unit.GetMeleeAttackPower()))))

        if unit.GetRangedAttackPower() > 0 {
            var rangedImage *ebiten.Image
            switch unit.GetRangedAttackDamageType() {
                case units.DamageNone:
                case units.DamageRangedMagical:
                    rangedImage = rangeMagicImage
                case units.DamageRangedPhysical:
                    rangedImage = rangeBowImage
                case units.DamageRangedBoulder:
                    rangedImage = rangeBoulderImage
            }

            stats.AddChild(combineHorizontalElements(makeIcon(rangedImage), makeText(fmt.Sprintf("%d", unit.GetRangedAttackPower()))))
        }

        stats.AddChild(combineHorizontalElements(makeIcon(resistanceImage), makeText(fmt.Sprintf("%d", unit.GetResistance()))))
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
        widget.ButtonOpts.Image(standardButtonImage()),
        widget.ButtonOpts.Text("Buy Unit", iconList.face, &widget.ButtonTextColor{
            Idle: color.White,
            Hover: color.White,
            Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
        widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
            iconList.buyUnit(unit)
            iconList.playSound.Play()
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

func combineHorizontalElementsCentered(elements... widget.PreferredSizeLocateableWidget) *widget.Container {
    box := ui.HBox(widget.ContainerOpts.WidgetOpts(
        widget.WidgetOpts.LayoutData(widget.RowLayoutData{
            Position: widget.RowLayoutPositionCenter,
        }),
    ))
    box.AddChild(elements...)
    return box
}

func makeMagicShop(face *text.Face, imageCache *util.ImageCache, lbxCache *lbx.LbxCache, playerObj *player.Player, uiEvents *UIEventUpdate, playSound *PlaySound) *widget.Container {
    shop := ui.VBox()
    shop.AddChild(widget.NewText(widget.TextOpts.Text("Magic Shop", face, color.White)))

    books := ui.HBox()
    shop.AddChild(books)

    lifeBook, _ := imageCache.GetImageTransform("newgame.lbx", 24, 0, "enlarge", enlargeTransform(2))
    sorceryBook, _ := imageCache.GetImageTransform("newgame.lbx", 27, 0, "enlarge", enlargeTransform(2))
    natureBook, _ := imageCache.GetImageTransform("newgame.lbx", 30, 0, "enlarge", enlargeTransform(2))
    deathBook, _ := imageCache.GetImageTransform("newgame.lbx", 33, 0, "enlarge", enlargeTransform(2))
    chaosBook, _ := imageCache.GetImageTransform("newgame.lbx", 36, 0, "enlarge", enlargeTransform(2))

    manaImage, _ := imageCache.GetImageTransform("backgrnd.lbx", 43, 0, "enlarge", enlargeTransform(2))

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

            cost := uint64(math.Pow(10, 2.1 + float64(playerObj.GetWizard().MagicLevel(magic)) / 10))

            gold, _ := imageCache.GetImageTransform("backgrnd.lbx", 42, 0, "enlarge", enlargeTransform(2))
            costUI := combineHorizontalElements(makeIcon(gold), widget.NewText(widget.TextOpts.Text(fmt.Sprintf("%d", cost), face, color.White), widget.TextOpts.WidgetOpts(centered)))

            buy.AddChild(costUI)

            buyButton := widget.NewButton(
                widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
                widget.ButtonOpts.WidgetOpts(centered),
                widget.ButtonOpts.Image(standardButtonImage()),
                widget.ButtonOpts.Text("Buy", face, &widget.ButtonTextColor{
                    Idle: color.White,
                    Hover: color.White,
                    Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                }),
                widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
                    if playerObj.Money >= cost {
                        playSound.Play()
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

    makeManaBuyButton := func(amount int) *widget.Button {
        var manaCost uint64
        for i := range amount {
            manaCost += computeManaCost(playerObj.OriginalMana + i + 1)
        }

        return widget.NewButton(
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.WidgetOpts(centered),
            widget.ButtonOpts.Image(standardButtonImage()),
            widget.ButtonOpts.Text(fmt.Sprintf("Buy %d for %d", amount, manaCost), face, &widget.ButtonTextColor{
                Idle: color.White,
                Hover: color.White,
                Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            }),
            widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
                if playerObj.Money >= uint64(manaCost) {
                    playSound.Play()
                    playerObj.Money -= uint64(manaCost)
                    playerObj.OriginalMana += amount
                    uiEvents.AddUpdate(&UIUpdateMana{})
                    uiEvents.AddUpdate(&UIUpdateMoney{})
                }
            }),
        )
    }

    manaBox := ui.HBox()

    setupMana := func() {
        manaBox.RemoveChildren()
        manaBox.AddChild(widget.NewText(widget.TextOpts.Text("Mana", face, color.White)))
        manaBox.AddChild(makeManaBuyButton(1), makeManaBuyButton(5))
    }

    setupMana()

    AddEvent(uiEvents, func (update *UIUpdateMoney) {
        setupMagic()
        setupMana()
    })

    shop.AddChild(manaBox)

    var tabs []*widget.TabBookTab

    getMagicIcon := func(magic data.MagicType) *widget.GraphicImage {
        makeIcon := func(img *ebiten.Image) *widget.GraphicImage {
            return &widget.GraphicImage{
                Idle: img,
                Disabled: img,
            }
        }

        index := -1
        switch magic {
            case data.LifeMagic: index = 7
            case data.SorceryMagic: index = 5
            case data.NatureMagic: index = 4
            case data.DeathMagic: index = 8
            case data.ChaosMagic: index = 6
        }

        if index > 0 {
            icon, _ := imageCache.GetImageTransform("spells.lbx", index, 0, "enlarge", enlargeTransform(2))
            return makeIcon(icon)
        }

        return nil
    }

    for _, magic := range allMagic {
        tab := widget.NewTabBookTab(
            widget.TabBookTabOpts.Label(magic.String()),
            widget.TabBookTabOpts.Image(getMagicIcon(magic)),
            widget.TabBookTabOpts.ContainerOpts(
            widget.ContainerOpts.Layout(widget.NewRowLayout(
                widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            ))),
        )

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

                manaAmount := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("%d", playerObj.ComputeEffectiveSpellCost(spell, false)), face, color.White))
                box.AddChild(combineHorizontalElementsCentered(ui.CenteredText(spell.Name, face, color.White), widget.NewGraphic(widget.GraphicOpts.Image(manaImage), widget.GraphicOpts.WidgetOpts(centered)), manaAmount))

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

                    booksNeeded := 0
                    needed:
                    for {
                        switch spell.Rarity {
                            case spellbook.SpellRarityCommon:
                                if rarityCount <= getCommonSpells(booksNeeded) {
                                    break needed
                                }
                            case spellbook.SpellRarityUncommon:
                                if rarityCount <= getUncommonSpells(booksNeeded) {
                                    break needed
                                }
                            case spellbook.SpellRarityRare:
                                if rarityCount <= getRareSpells(booksNeeded) {
                                    break needed
                                }
                            case spellbook.SpellRarityVeryRare:
                                if rarityCount <= getVeryRareSpells(booksNeeded) {
                                    break needed
                                }
                        }

                        booksNeeded += 1
                    }

                    if booksNeeded > 0 {
                        var use *ebiten.Image
                        switch magic {
                            case data.LifeMagic: use = lifeBook
                            case data.SorceryMagic: use = sorceryBook
                            case data.NatureMagic: use = natureBook
                            case data.DeathMagic: use = deathBook
                            case data.ChaosMagic: use = chaosBook
                        }

                        final := ebiten.NewImage(use.Bounds().Dx() * booksNeeded, use.Bounds().Dy())

                        for x := range booksNeeded {
                            var ops ebiten.DrawImageOptions
                            ops.GeoM.Translate(float64(x * use.Bounds().Dx()), 0)

                            if x >= playerObj.GetWizard().MagicLevel(magic) {
                                ops.ColorScale.ScaleWithColor(color.NRGBA{R: 90, G: 90, B: 90, A: 255})
                            }

                            final.DrawImage(use, &ops)
                        }

                        books := ui.HBox(
                            widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                                Position: widget.RowLayoutPositionCenter,
                            })),
                        )

                        books.AddChild(widget.NewGraphic(
                            widget.GraphicOpts.Image(final),
                            widget.GraphicOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                                Position: widget.RowLayoutPositionCenter,
                            })),
                        ))

                        box.AddChild(books)
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
                                playSound.Play()
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
            widget.SliderOpts.Orientation(widget.DirectionVertical),
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
        widget.TabBookOpts.TabButtonImage(standardButtonImage()),
        widget.TabBookOpts.TabButtonText(face, &widget.ButtonTextColor{
            Idle: color.White,
            Disabled: color.NRGBA{R: 32, G: 32, B: 32, A: 255},
            Hover: color.White,
            Pressed: color.White,
        }),
        widget.TabBookOpts.TabButtonSpacing(8),
        // widget.TabBookOpts.ContentPadding(widget.NewInsetsSimple(2)),
        widget.TabBookOpts.Tabs(tabs...),
    )

    shop.AddChild(spellsTabs)

    return shop
}

func makeArmyShop(face *text.Face, imageCache *util.ImageCache, playerObj *player.Player, uiEvents *UIEventUpdate, playSound *PlaySound) *widget.Container {
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
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
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

    unitList := MakeUnitIconList("All Units", imageCache, face, buyUnit, playSound)

    for _, unit := range getValidChoices(100000) {
        unitList.AddUnit(unit)
    }

    unitList.SortByName()

    filteredUnitList := MakeUnitIconList("Affordable Units", imageCache, face, buyUnit, playSound)

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

    allButton := widget.NewButton(
        widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Image(standardButtonImage()),
        widget.ButtonOpts.Text("All Units", face, &widget.ButtonTextColor{
            Idle: color.White,
            Hover: color.White,
            Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
    )

    affordableButton := widget.NewButton(
        widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Image(standardButtonImage()),
        widget.ButtonOpts.Text("Affordable Units", face, &widget.ButtonTextColor{
            Idle: color.White,
            Hover: color.White,
            Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
    )

    buttons := ui.HBox()
    buttons.AddChild(allButton, affordableButton)

    listContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
        )),
    )

    first := true

    widget.NewRadioGroup(
        widget.RadioGroupOpts.Elements(allButton, affordableButton),
        widget.RadioGroupOpts.InitialElement(allButton),
        widget.RadioGroupOpts.ChangedHandler(func(args *widget.RadioGroupChangedEventArgs) {
            // the handler is called when the radio group is created, so don't play sound the first time
            if !first {
                playSound.Play()
            }
            first = false
            listContainer.RemoveChildren()
            if args.Active == allButton {
                listContainer.AddChild(unitList.GetWidget())
            } else {
                listContainer.AddChild(filteredUnitList.GetWidget())
            }
        }),
    )

    container2.AddChild(buttons)

    listContainer.AddChild(unitList.GetWidget())

    container2.AddChild(listContainer)

    return armyShop
}

func makeShopUI(face *text.Face, imageCache *util.ImageCache, lbxCache *lbx.LbxCache, playerObj *player.Player, uiEvents *UIEventUpdate, playSound *PlaySound) *widget.Container {
    container := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewGridLayout(
            widget.GridLayoutOpts.Columns(2),
            widget.GridLayoutOpts.DefaultStretch(false, false),
            widget.GridLayoutOpts.Padding(&widget.Insets{Top: 4, Bottom: 4, Left: 4, Right: 4}),
            widget.GridLayoutOpts.Spacing(20, 0),
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

    armyShop := makeArmyShop(face, imageCache, playerObj, uiEvents, playSound)

    container.AddChild(armyShop)

    magicShop := makeMagicShop(face, imageCache, lbxCache, playerObj, uiEvents, playSound)
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

func makeBuyEnchantments(unit units.StackUnit, face *text.Face, playerObj *player.Player, uiEvents *UIEventUpdate, imageCache *util.ImageCache, playSound *PlaySound) *widget.Container {
    // remove any enchantments the unit already has
    enchantments := slices.DeleteFunc(getValidUnitEnchantments(), unit.HasEnchantment)

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
        widget.SliderOpts.Orientation(widget.DirectionVertical),
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

    lifeBook, _ := imageCache.GetImageTransform("newgame.lbx", 24, 0, "enlarge", enlargeTransform(2))
    sorceryBook, _ := imageCache.GetImageTransform("newgame.lbx", 27, 0, "enlarge", enlargeTransform(2))
    natureBook, _ := imageCache.GetImageTransform("newgame.lbx", 30, 0, "enlarge", enlargeTransform(2))
    deathBook, _ := imageCache.GetImageTransform("newgame.lbx", 33, 0, "enlarge", enlargeTransform(2))
    chaosBook, _ := imageCache.GetImageTransform("newgame.lbx", 36, 0, "enlarge", enlargeTransform(2))

    setupEnchantments := func() {
        for _, enchantment := range enchantments {
            border := ui.BorderedImage(color.RGBA{R: 128, G: 128, B: 128, A: 255}, 1)
            box := ui.VBox(
                widget.ContainerOpts.BackgroundImage(border),
            )
            name := ui.CenteredText(enchantment.Name(), face, enchantment.Color())
            box.AddChild(name)

            books := ui.HBox(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                Position: widget.RowLayoutPositionCenter,
            })))

            requirements := getEnchantmentRequirements(enchantment)
            var use *ebiten.Image
            switch requirements.Magic {
                case data.LifeMagic: use = lifeBook
                case data.SorceryMagic: use = sorceryBook
                case data.NatureMagic: use = natureBook
                case data.DeathMagic: use = deathBook
                case data.ChaosMagic: use = chaosBook
            }

            final := ebiten.NewImage(use.Bounds().Dx() * requirements.Count, use.Bounds().Dy())

            for x := range requirements.Count {
                var ops ebiten.DrawImageOptions
                ops.GeoM.Translate(float64(x * use.Bounds().Dx()), 0)

                if x >= playerObj.GetWizard().MagicLevel(requirements.Magic) {
                    ops.ColorScale.ScaleWithColor(color.NRGBA{R: 90, G: 90, B: 90, A: 255})
                }

                final.DrawImage(use, &ops)
            }

            books.AddChild(widget.NewGraphic(widget.GraphicOpts.Image(final), widget.GraphicOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                Position: widget.RowLayoutPositionCenter,
            }))))

            box.AddChild(books)

            money := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("%d", getEnchantmentCost(enchantment)), face, color.White))

            box.AddChild(makeMoneyText(money, imageCache, widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                Position: widget.RowLayoutPositionCenter,
            }))))

            enchantTextColor := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
            canBuy := playerObj.GetWizard().MagicLevel(requirements.Magic) >= requirements.Count
            if !canBuy {
                enchantTextColor = color.NRGBA{R: 128, G: 128, B: 128, A: 255}
            }

            remove := func(){}
            enchantButton := widget.NewButton(
                widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
                widget.ButtonOpts.Image(standardButtonImage()),
                widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                    Position: widget.RowLayoutPositionCenter,
                })),
                widget.ButtonOpts.Text("Enchant", face, &widget.ButtonTextColor{
                    Idle: enchantTextColor,
                    Hover: enchantTextColor,
                    Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                }),
                widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
                    cost := uint64(getEnchantmentCost(enchantment))
                    if canBuy && cost <= playerObj.Money && !unit.HasEnchantment(enchantment) {
                        playSound.Play()
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
    }

    setupEnchantments()

    AddEvent(uiEvents, func (update *UIUpdateMagicBooks) {
        enchantmentList.RemoveChildren()
        setupEnchantments()
    })

    container := ui.HBox()
    container.AddChild(scroller)
    container.AddChild(slider)

    return container
}

func sum[T any](items []T, f func(T) int) int {
    total := 0
    for _, item := range items {
        total += f(item)
    }
    return total
}

func makeUnitInfoUI(face *text.Face, allUnits []units.StackUnit, playerObj *player.Player, uiEvents *UIEventUpdate, imageCache *util.ImageCache, playSound *PlaySound) *widget.Container {

    unitContainer := ui.HBox()

    var updateUnitSpecifics func(unit units.StackUnit, setup func())

    removeUnit := func() {
    }

    gold, _ := imageCache.GetImageTransform("backgrnd.lbx", 42, 0, "enlarge", enlargeTransform(2))
    moneyImage := &widget.GraphicImage{
        Idle: gold,
        Disabled: gold,
    }

    updateUnitSpecifics = func(unit units.StackUnit, setup func()) {
        unitContainer.RemoveChildren()

        removeUnit()

        if unit == nil {
            return
        }

        unitSpecifics := widget.NewContainer(
            widget.ContainerOpts.Layout(widget.NewRowLayout(
                widget.RowLayoutOpts.Direction(widget.DirectionVertical),
                widget.RowLayoutOpts.Spacing(2),
            )),
        )

        unitContainer.AddChild(unitSpecifics)

        unitSpecifics.AddChild(widget.NewText(
            widget.TextOpts.Text("Unit Specifics", face, color.White),
        ))

        currentName := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Name: %v", unit.GetFullName()), face, color.White))
        currentHealth := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("HP: %d/%d", unit.GetHealth(), unit.GetMaxHealth()), face, color.White))
        currentRace := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Race: %v", unit.GetRace()), face, color.White))

        rawUnit := unit.GetRawUnit()
        var removeButton *widget.Button

        makeRemoveButton := func() *widget.Button {
            sellCost := getUnitCost(&rawUnit) + uint64(sum(unit.GetEnchantments(), getEnchantmentCost))
            sellCost = (sellCost * 3) / 4
            return widget.NewButton(
                widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
                widget.ButtonOpts.Image(standardButtonImage()),
                widget.ButtonOpts.TextAndImage(fmt.Sprintf("Sell %v for %v", unit.GetFullName(), sellCost), face, moneyImage, &widget.ButtonTextColor{
                    Idle: color.White,
                    Hover: color.White,
                    Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                }),
                widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
                    log.Printf("Selling unit %v", unit.GetFullName())
                    playSound.Play()
                    playerObj.Money += sellCost
                    playerObj.RemoveUnit(unit)
                    updateUnitSpecifics(nil, func(){})

                    uiEvents.AddUpdate(&UIUpdateMoney{})
                    uiEvents.AddUpdate(&UIRemoveUnit{Unit: unit})
                }),
            )
        }

        removeButton = makeRemoveButton()

        updateSellCost := func() {
            newButton := makeRemoveButton()
            unitContainer.ReplaceChild(removeButton, newButton)
            removeButton = newButton
        }

        unitContainer.AddChild(removeButton)

        unitSpecifics.RemoveChildren()
        unitSpecifics.AddChild(currentName)
        unitSpecifics.AddChild(currentHealth)
        unitSpecifics.AddChild(currentRace)
        if unit.GetRace() != data.RaceFantastic {
            unitSpecifics.AddChild(widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Experience: %d (%v)", unit.GetExperience(), unit.GetExperienceLevel().Name()), face, color.White)))
            var makeMeleeBox func() *widget.Container

            makeMeleeBox = func() *widget.Container {
                meleeBox := ui.HBox()
                meleeIcon, _ := imageCache.GetImageTransform("unitview.lbx", 13, 0, "enlarge", enlargeTransform(2))

                switch unit.GetWeaponBonus() {
                    case data.WeaponMagic:
                        meleeIcon, _ = imageCache.GetImageTransform("unitview.lbx", 16, 0, "enlarge", enlargeTransform(2))
                    case data.WeaponMythril:
                        meleeIcon, _ = imageCache.GetImageTransform("unitview.lbx", 15, 0, "enlarge", enlargeTransform(2))
                    case data.WeaponAdamantium:
                        meleeIcon, _ = imageCache.GetImageTransform("unitview.lbx", 17, 0, "enlarge", enlargeTransform(2))
                }

                meleeBox.AddChild(combineHorizontalElements(
                    widget.NewText(widget.TextOpts.Text("Melee:", face, color.White)),
                    widget.NewGraphic(widget.GraphicOpts.Image(meleeIcon)),
                ))

                if unit.GetWeaponBonus() != data.WeaponAdamantium {
                    upgradeCost := getWeaponUpgradeCost(unit.GetWeaponBonus())
                    meleeBox.AddChild(widget.NewButton(
                        widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
                        widget.ButtonOpts.Image(standardButtonImage()),
                        widget.ButtonOpts.TextAndImage(fmt.Sprintf("Upgrade %v", upgradeCost), face, moneyImage, &widget.ButtonTextColor{
                            Idle: color.White,
                            Hover: color.White,
                            Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                        }),
                        widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
                            if upgradeCost > playerObj.Money {
                                return
                            }

                            playSound.Play()

                            switch unit.GetWeaponBonus() {
                                case data.WeaponNone:
                                    unit.SetWeaponBonus(data.WeaponMagic)
                                case data.WeaponMagic:
                                    unit.SetWeaponBonus(data.WeaponMythril)
                                case data.WeaponMythril:
                                    unit.SetWeaponBonus(data.WeaponAdamantium)
                            }

                            newMeleeBox := makeMeleeBox()
                            unitSpecifics.ReplaceChild(meleeBox, newMeleeBox)
                            playerObj.Money -= upgradeCost
                            uiEvents.AddUpdate(&UIUpdateMoney{})
                        }),
                    ))
                }

                return meleeBox
            }

            unitSpecifics.AddChild(makeMeleeBox())
        }

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
            widget.SliderOpts.Orientation(widget.DirectionHorizontal),
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
            widget.ButtonOpts.Image(standardButtonImage()),
            widget.ButtonOpts.Text("Heal Unit", face, &widget.ButtonTextColor{
                Idle: color.White,
                Hover: color.White,
                Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            }),
            widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
                playSound.Play()
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
            widget.ListOpts.ContainerOpts(
                widget.ContainerOpts.WidgetOpts(
                    widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                        MaxHeight: 230,
                    }),
                ),
            ),

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

                updateSellCost()
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

        buyEnchantments := makeBuyEnchantments(unit, face, playerObj, uiEvents, imageCache, playSound)

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

    unitMap := make(map[units.StackUnit]*widget.Container)

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

                vector.FillRect(healthImage, 0, 10, float32(healthImage.Bounds().Dx()), 4, color.RGBA{R: 0, G: 0, B: 0, A: 255}, true)
                vector.FillRect(healthImage, 0, 10, length, 4, healthColor, true)
                box1.AddChild(widget.NewGraphic(widget.GraphicOpts.Image(healthImage)))

                unitBox.AddChild(box1)
            }
        }

        setup()

        unitList.AddChild(unitBox)

        unitMap[unit] = unitBox
    }

    removeUnitUI := func(unit units.StackUnit) {
        box, ok := unitMap[unit]
        if ok {
            unitList.RemoveChild(box)
            delete(unitMap, unit)
        }
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
        widget.SliderOpts.Orientation(widget.DirectionVertical),
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

    AddEvent(uiEvents, func (update *UIRemoveUnit) {
        removeUnitUI(update.Unit)
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

    unitInfoContainer.AddChild(unitContainer)

    return unitInfoContainer
}

func makePlayerInfoUI(face *text.Face, playerObj *player.Player, events *UIEventUpdate, imageCache *util.ImageCache) *widget.Container {
    container := ui.HBox()

    name := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Name: %v", playerObj.Wizard.Name), face, color.White))
    container.AddChild(name)

    level := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Level: %d", playerObj.Level), face, color.White))
    container.AddChild(level)

    mana := widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Mana: %d", playerObj.OriginalMana), face, color.White))
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
        mana.Label = fmt.Sprintf("Mana: %d", playerObj.OriginalMana)
    })

    AddEvent(events, func (update *UIUpdateMagicBooks) {
        setupBooks()
    })

    return container
}

func (engine *Engine) MakeNewGameUI() (*ebitenui.UI, error) {
    rootContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewAnchorLayout(
        ),
    ))

    font, err := console.LoadFont()
    if err != nil {
        return nil, err
    }

    face := text.GoTextFace{
        Source: font,
        Size: 24,
    }

    var face1 text.Face = &face

    buttons := ui.VBox(
        widget.ContainerOpts.BackgroundImage(ui_image.NewBorderedNineSliceColor(color.NRGBA{R: 32, G: 32, B: 32, A: 255}, color.NRGBA{R: 200, G: 200, B: 200, A: 255}, 1)),
        widget.ContainerOpts.WidgetOpts(
            widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
                HorizontalPosition: widget.AnchorLayoutPositionCenter,
                VerticalPosition: widget.AnchorLayoutPositionCenter,
            }),
        ),
    )

    for _, difficulty := range []Difficulty{DifficultyEasy, DifficultyNormal, DifficultyHard, DifficultyImpossible} {
        buttons.AddChild(widget.NewButton(
            widget.ButtonOpts.WidgetOpts(
                widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                    Position: widget.RowLayoutPositionCenter,
                }),
            ),
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 4, Bottom: 4, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(standardButtonImage()),
            widget.ButtonOpts.Text(fmt.Sprintf("Start %v Mode", difficulty), &face1, &widget.ButtonTextColor{
                Idle: color.White,
                Hover: color.White,
                Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            }),
            widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
                engine.LeftClickSound.Play()
                select {
                    case engine.Events <- &EventNewGame{Level: difficulty}:
                    default:
                }
            }),
        ))
    }

    rootContainer.AddChild(buttons)

    ui := &ebitenui.UI{
        Container: rootContainer,
    }

    return ui, nil
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

    var face2 text.Face = &text.GoTextFace{
        Source: font,
        Size: 20,
    }

    uiEvents := MakeUIEventUpdate()

    rootContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(4),
            widget.RowLayoutOpts.Padding(&widget.Insets{Top: 4, Bottom: 4, Left: 4, Right: 4}),
        )),
        widget.ContainerOpts.BackgroundImage(ui_image.NewNineSliceColor(color.NRGBA{R: 32, G: 32, B: 32, A: 255})),
    )

    enterBattleButton := widget.NewButton(
        widget.ButtonOpts.WidgetOpts(
            widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                Position: widget.RowLayoutPositionCenter,
            }),
        ),
        widget.ButtonOpts.TextPadding(&widget.Insets{Top: 4, Bottom: 4, Left: 5, Right: 5}),
        widget.ButtonOpts.Image(standardButtonImage()),
        widget.ButtonOpts.Text("Enter Battle", &face2, &widget.ButtonTextColor{
            Idle: color.White,
            Hover: color.White,
            Pressed: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
        widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
            if len(engine.Player.Units) == 0 {
                return
            }
            engine.LeftClickSound.Play()
            select {
                case engine.Events <- &EventEnterBattle{}:
                default:
            }
        }),
    )

    imageCache := util.MakeImageCache(engine.Cache)

    rootContainer.AddChild(enterBattleButton)

    rootContainer.AddChild(makePlayerInfoUI(&face1, engine.Player, uiEvents, &imageCache))

    unitInfoUI := makeUnitInfoUI(&face1, engine.Player.Units, engine.Player, uiEvents, &imageCache, engine.LeftClickSound)

    rootContainer.AddChild(unitInfoUI)
    rootContainer.AddChild(makeShopUI(&face1, &imageCache, engine.Cache, engine.Player, uiEvents, engine.LeftClickSound))

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
    playerObj.Money = 300000
    playerObj.Level = 8
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

    music := musiclib.MakeMusic(cache)

    clickMaker, err := audio.LoadSoundMaker(cache, audio.SoundClick)
    if err != nil {
        log.Printf("Error loading click sound: %v", err)
        clickMaker = nil
    }

    engine := Engine{
        GameMode: GameModeNewGameUI,
        Player: playerObj,
        Cache: cache,
        Events: make(chan EngineEvents, 10),
        Music: music,
        LeftClickSound: &PlaySound{Maker: clickMaker},
    }

    engine.PushDrawer(engine.DefaultDraw)

    engine.Music.PushSongs(musiclib.SongBackground1, musiclib.SongBackground2, musiclib.SongBackground3)

    engine.UI, engine.UIUpdates, err = engine.MakeUI()
    if err != nil {
        log.Printf("Error creating UI: %v", err)
    }

    engine.NewGameUI, err = engine.MakeNewGameUI()

    return &engine
}

func (engine *Engine) Shutdown() {
    engine.Music.Stop()
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

    engine.Shutdown()
}
