package main

import (
    "embed"
    "log"
    "fmt"
    "image"
    "image/color"
    "slices"
    "cmp"
    "errors"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/util/common"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/colorm"
    "github.com/hajimehoshi/ebiten/v2/text/v2"

    "github.com/ebitenui/ebitenui"
    // "github.com/ebitenui/ebitenui/input"
    "github.com/ebitenui/ebitenui/widget"
    ui_image "github.com/ebitenui/ebitenui/image"
)

//go:embed assets/*
var assets embed.FS

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

func makeRoundedButtonImage(width int, height int, border int, col color.Color) *ebiten.Image {
    img := ebiten.NewImage(width, height)

    vector.DrawFilledRect(img, float32(border), 0, float32(width - border * 2), float32(height), col, true)
    vector.DrawFilledRect(img, 0, float32(border), float32(width), float32(height - border * 2), col, true)
    vector.DrawFilledCircle(img, float32(border), float32(border), float32(border), col, true)
    vector.DrawFilledCircle(img, float32(width-border), float32(border), float32(border), col, true)
    vector.DrawFilledCircle(img, float32(border), float32(height-border), float32(border), col, true)
    vector.DrawFilledCircle(img, float32(width-border), float32(height-border), float32(border), col, true)

    return img
}

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

func scaleImage(img *ebiten.Image, newHeight int) *ebiten.Image {
    scale := float64(newHeight) / float64(img.Bounds().Dy())
    newWidth := int(float64(img.Bounds().Dx()) * scale)

    newImage := ebiten.NewImage(newWidth, newHeight)
    var options ebiten.DrawImageOptions
    options.GeoM.Scale(scale, scale)
    newImage.DrawImage(img, &options)
    return newImage
}

func (engine *Engine) MakeUI() *ebitenui.UI {

    imageCache := util.MakeImageCache(engine.Cache)

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

    makeNineImage := func (img *ebiten.Image, border int) *ui_image.NineSlice {
        width := img.Bounds().Dx()
        return ui_image.NewNineSliceSimple(img, border, width - border * 2)
    }

    lighten := func (c color.Color, amount float64) color.Color {
        var change colorm.ColorM
        change.ChangeHSV(0, 1 - amount/100, 1 + amount/100)
        return change.Apply(c)
    }

    makeNineRoundedButtonImage := func(width int, height int, border int, col color.Color) *widget.ButtonImage {
        return &widget.ButtonImage{
            Idle: makeNineImage(makeRoundedButtonImage(width, height, border, col), border),
            Hover: makeNineImage(makeRoundedButtonImage(width, height, border, lighten(col, 50)), border),
            Pressed: makeNineImage(makeRoundedButtonImage(width, height, border, lighten(col, 20)), border),
        }
    }

    /*
    tabImageNine := makeNineImage(makeRoundedButtonImage(80, 80, 10, color.NRGBA{R: 64, G: 64, B: 64, A: 255}), 10)
    tabImageHoverNine := makeNineImage(makeRoundedButtonImage(80, 80, 10, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 10)
    tabImagePressedNine := makeNineImage(makeRoundedButtonImage(80, 80, 10, color.NRGBA{R: 96, G: 96, B: 96, A: 255}), 10)
    */

    face, _ := loadFont(19)

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
            widget.RowLayoutOpts.Spacing(12),
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

    // var raceTabs []*widget.TabBookTab

    allRaces := append(append(data.ArcanianRaces(), data.MyrranRaces()...), []data.Race{data.RaceFantastic, data.RaceHero}...)

    unitListContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(5),
        )),
    )

    var raceButtons []*widget.Button

    makeRaceButton := func (race data.Race, update func()) *widget.Button {
        raceLbx := "backgrnd.lbx"
        raceIndex := 44

        switch race {
            case data.RaceHero:
                raceLbx = "figures1.lbx"
                raceIndex = 2
            case data.RaceFantastic:
                raceLbx = "figure11.lbx"
                raceIndex = 115
            case data.RaceLizard:
                raceIndex = 55
            case data.RaceNomad:
                raceIndex = 56
            case data.RaceOrc:
                raceIndex = 57
            case data.RaceTroll:
                raceIndex = 58
            case data.RaceBarbarian:
                raceIndex = 45
            case data.RaceBeastmen:
                raceIndex = 46
            case data.RaceDarkElf:
                raceIndex = 47
            case data.RaceDraconian:
                raceIndex = 48
            case data.RaceDwarf:
                raceIndex = 49
            case data.RaceGnoll:
                raceIndex = 50
            case data.RaceHalfling:
                raceIndex = 51
            case data.RaceHighElf:
                raceIndex = 52
            case data.RaceHighMen:
                raceIndex = 53
            case data.RaceKlackon:
                raceIndex = 54
        }

        raceImage, _ := imageCache.GetImageTransform(raceLbx, raceIndex, 0, "enlarge", enlargeTransform(2))
        rescaled := scaleImage(raceImage, 30)

        return widget.NewButton(
            widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
            widget.ButtonOpts.TextAndImage(race.String(), face, &widget.ButtonImageImage{Idle: rescaled, Disabled: raceImage}, &widget.ButtonTextColor{
                Idle: color.White,
                Hover: color.White,
                Pressed: color.White,
            }),
            widget.ButtonOpts.PressedHandler(func (args *widget.ButtonPressedEventArgs) {
                update()
            }),
        )
    }

    for _, race := range allRaces {
        /*
        tab := widget.NewTabBookTab(
            race.String(),
            widget.ContainerOpts.Layout(widget.NewRowLayout(
                widget.RowLayoutOpts.Direction(widget.DirectionVertical),
                widget.RowLayoutOpts.Spacing(5),
            )),
        )
        */

        clickTimer := make(map[string]uint64)

        unitGraphic := widget.NewGraphic()

        updateGraphic := func (unit units.Unit) {
            unitImage, err := imageCache.GetImageTransform(unit.CombatLbxFile, unit.CombatIndex + 2, 0, "enlarge", enlargeTransform(4))
            if err == nil {
                unitGraphic.Image = unitImage
            }
        }

        unitList := widget.NewList(
            widget.ListOpts.EntryFontFace(face),
            widget.ListOpts.SliderOpts(
                widget.SliderOpts.Images(
                    &widget.SliderTrackImage{
                        Idle: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                        Hover: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                    },
                    makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xad, G: 0x8d, B: 0x55, A: 0xff}),
                ),
            ),

            widget.ListOpts.HideHorizontalSlider(),

            widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(
                widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                    MaxHeight: 200,
                }),
                widget.WidgetOpts.MinSize(0, 200),
            )),

            /*
            widget.ListOpts.ContainerOpts(widget.ContainerOpts.BackgroundImage(
                ui_image.NewNineSliceColor(color.NRGBA{R: 128, G: 64, B: 64, A: 255}),
            )),
            */

            widget.ListOpts.EntryLabelFunc(
                func (e any) string {
                    item := e.(*UnitItem)
                    return fmt.Sprintf("%v %v", item.Race, item.Unit.Name)
                },
            ),

            widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
                entry := args.Entry.(*UnitItem)

                updateGraphic(entry.Unit)

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

        space := widget.NewGraphic(widget.GraphicOpts.Image(ebiten.NewImage(60, 30)))

        raceRow := makeRow(5, space, unitGraphic, unitList)

        // tab.AddChild(makeRow(5, space, unitGraphic, unitList))

        for _, unit := range slices.SortedFunc(slices.Values(units.UnitsByRace(race)), func (a units.Unit, b units.Unit) int {
            return cmp.Compare(a.Name, b.Name)
        }) {
            unitList.AddEntry(&UnitItem{
                Race: race,
                Unit: unit,
            })
        }

        moreButtonsRow := makeRow(5,
            space,
            widget.NewButton(
                widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
                widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
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
            ),
            widget.NewButton(
                widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
                widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
                widget.ButtonOpts.Text("Add Random", face, &widget.ButtonTextColor{
                    Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                    Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                    Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
                }),
                widget.ButtonOpts.PressedHandler(func (args *widget.ButtonPressedEventArgs) {
                    choices := units.UnitsByRace(race)
                    use := choices[rand.N(len(choices))]
                    newItem := UnitItem{
                        Race: race,
                        Unit: use,
                    }
                    armyList.AddEntry(&newItem)
                    armyCount.Label = fmt.Sprintf("%v", len(armyList.Entries()))
                }),
            ),
        )

        // raceTabs = append(raceTabs, tab)

        update := func(){
            unitListContainer.RemoveChildren()
            unitListContainer.AddChild(raceRow)
            unitListContainer.AddChild(moreButtonsRow)
        }

        raceButtons = append(raceButtons, makeRaceButton(race, update))
    }

    defendingArmyCount := widget.NewText(widget.TextOpts.Text("0", face, color.NRGBA{R: 255, G: 255, B: 255, A: 255}))

    makeArmyList := func() *widget.List {
        return widget.NewList(
            widget.ListOpts.EntryFontFace(face),

            widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(
                widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                    MaxHeight: 300,
                }),
                widget.WidgetOpts.MinSize(0, 300),
            )),

            widget.ListOpts.SliderOpts(
                widget.SliderOpts.Images(&widget.SliderTrackImage{
                        Idle: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                        Hover: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                    },
                    makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xad, G: 0x8d, B: 0x55, A: 0xff}),
                ),
            ),

            widget.ListOpts.HideHorizontalSlider(),

            widget.ListOpts.EntryLabelFunc(
                func (e any) string {
                    item := e.(*UnitItem)

                    if item.Race == data.RaceFantastic || item.Race == data.RaceAll {
                        return item.Unit.Name
                    }

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

    raceRows := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(5),
        )),
    )

    buttonsPerRow := 8
    for i := 0; i < len(raceButtons); i += buttonsPerRow {
        max := i + buttonsPerRow
        if max >= len(raceButtons) {
            max = len(raceButtons)
        }

        var widgets []widget.PreferredSizeLocateableWidget
        for j := i; j < max; j++ {
            widgets = append(widgets, raceButtons[j])
        }

        raceRows.AddChild(makeRow(5, widgets...))
    }

    raceButtons[0].Click()

    rootContainer.AddChild(raceRows)

    // greyish := color.NRGBA{R: 128, G: 128, B: 128, A: 255}

    rootContainer.AddChild(unitListContainer)

    /*
    unitsTabs := widget.NewTabBook(
        widget.TabBookOpts.TabButtonImage(&widget.ButtonImage{
            Idle: tabImageNine,
            Hover: tabImageHoverNine,
            Pressed: tabImagePressedNine,
        }),
        widget.TabBookOpts.TabButtonSpacing(3),
        widget.TabBookOpts.TabButtonText(face, &widget.ButtonTextColor{Idle: color.White, Disabled: greyish}),
        widget.TabBookOpts.Tabs(raceTabs...),
        widget.TabBookOpts.TabButtonOpts(widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5})),
    )

    rootContainer.AddChild(unitsTabs)
    */

    rootContainer.AddChild(widget.NewButton(
        widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
        widget.ButtonOpts.Text("Add Random Unit", face, &widget.ButtonTextColor{
            Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
            Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
        }),
        widget.ButtonOpts.PressedHandler(func (args *widget.ButtonPressedEventArgs) {
            unit := units.AllUnits[rand.N(len(units.AllUnits))]
            newItem := UnitItem{
                Race: unit.Race,
                Unit: unit,
            }

            armyList.AddEntry(&newItem)
            armyCount.Label = fmt.Sprintf("%v", len(armyList.Entries()))
        }),
    ))

    var armyButtons []*widget.Button
    armyButtons = append(armyButtons, widget.NewButton(
        widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
        widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
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
        widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
        widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
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
    defendingArmyContainer.AddChild(makeRow(4,
        widget.NewButton(
            widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
            widget.ButtonOpts.Text("Clear Defending Army", face, &widget.ButtonTextColor{
                Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            }),
            widget.ButtonOpts.PressedHandler(func (args *widget.ButtonPressedEventArgs) {
                defendingArmyList.SetEntries(nil)
                defendingArmyCount.Label = "0"
            }),
        ),
        widget.NewButton(
            widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
            widget.ButtonOpts.Text("Remove Unit", face, &widget.ButtonTextColor{
                Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            }),
            widget.ButtonOpts.PressedHandler(func (args *widget.ButtonPressedEventArgs) {
                selected := defendingArmyList.SelectedEntry()
                if selected != nil {
                    defendingArmyList.RemoveEntry(selected)
                    defendingArmyCount.Label = fmt.Sprintf("%v", len(defendingArmyList.Entries()))
                }
            }),
        ),
    ))

    armyContainer.AddChild(defendingArmyContainer)

    attackingArmyContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
        )),
    )

    attackingArmyContainer.AddChild(makeRow(4, attackingArmyName, attackingArmyCount))

    attackingArmyContainer.AddChild(attackingArmyList)
    attackingArmyContainer.AddChild(makeRow(4,
        widget.NewButton(
            widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
            widget.ButtonOpts.Text("Clear Attacking Army", face, &widget.ButtonTextColor{
                Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            }),
            widget.ButtonOpts.PressedHandler(func (args *widget.ButtonPressedEventArgs) {
                attackingArmyList.SetEntries(nil)
                attackingArmyCount.Label = "0"
            }),
        ),
        widget.NewButton(
            widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
            widget.ButtonOpts.Text("Remove Unit", face, &widget.ButtonTextColor{
                Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            }),
            widget.ButtonOpts.PressedHandler(func (args *widget.ButtonPressedEventArgs) {
                selected := attackingArmyList.SelectedEntry()
                if selected != nil {
                    attackingArmyList.RemoveEntry(selected)
                    attackingArmyCount.Label = fmt.Sprintf("%v", len(attackingArmyList.Entries()))
                }
            }),
        ),
    ))

    armyContainer.AddChild(attackingArmyContainer)

    rootContainer.AddChild(armyContainer)

    combatPicture, _ := imageCache.GetImageTransform("special.lbx", 29, 0, "enlarge", enlargeTransform(2))
    rootContainer.AddChild(widget.NewButton(
        widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
        widget.ButtonOpts.TextAndImage("Enter Combat!", face, &widget.ButtonImageImage{Idle: combatPicture, Disabled: combatPicture}, &widget.ButtonTextColor{
            Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
            Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
        }),

        /*
        widget.ButtonOpts.Text("Enter Combat!", face, &widget.ButtonTextColor{
            Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
            Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
        }),
        widget.ButtonOpts.Graphic(combatPicture),
        */
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
    ebiten.SetWindowSize(1200, 850)
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    err := ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
