package armyview

import (
    "fmt"
    "log"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/colorm"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

type ArmyScreenState int

const (
    ArmyScreenStateRunning ArmyScreenState = iota
    ArmyScreenStateDone
)

type ArmyScreen struct {
    Cache *lbx.LbxCache
    ImageCache util.ImageCache
    Player *playerlib.Player
    State ArmyScreenState
    ShowVault func()
    FirstRow int
    UI *uilib.UI
    DrawMinimap func(*ebiten.Image, int, int, data.FogMap, data.Plane, uint64)

    ShowArcanus bool
    ShowMyrror bool

    arcanusCounter float32
    myrrorCounter float32
}

func MakeArmyScreen(cache *lbx.LbxCache, player *playerlib.Player, drawMinimap func(*ebiten.Image, int, int, data.FogMap, data.Plane, uint64), showVault func()) *ArmyScreen {
    view := &ArmyScreen{
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        Player: player,
        State: ArmyScreenStateRunning,
        ShowVault: showVault,
        DrawMinimap: drawMinimap,
        FirstRow: 0,
        ShowArcanus: true,
        ShowMyrror: true,
    }

    view.UI = view.MakeUI()

    return view
}

type ArmyViewFonts struct {
    NormalFont *font.Font
    SmallerFont *font.Font
    BigFont *font.Font
}

func MakeArmyViewFonts(cache *lbx.LbxCache) *ArmyViewFonts {
    use, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Error loading army view fonts: %v", err)
        return nil
    }

    return &ArmyViewFonts{
        NormalFont: use(fontslib.NormalFont),
        SmallerFont: use(fontslib.SmallerFont),
        BigFont: use(fontslib.BigFont),
    }
}

func (view *ArmyScreen) MakeUI() *uilib.UI {
    var highlightedUnit units.StackUnit

    fonts := MakeArmyViewFonts(view.Cache)

    upArrows, _ := view.ImageCache.GetImages("armylist.lbx", 1)
    downArrows, _ := view.ImageCache.GetImages("armylist.lbx", 2)

    upkeepGold := view.Player.TotalUnitUpkeepGold()
    upkeepFood := view.Player.TotalUnitUpkeepFood()
    upkeepMana := view.Player.TotalUnitUpkeepMana()

    ui := &uilib.UI{
        Draw: func(this *uilib.UI, screen *ebiten.Image) {
            background, _ := view.ImageCache.GetImage("armylist.lbx", 0, 0)
            var options ebiten.DrawImageOptions
            scale.DrawScaled(screen, background, &options)

            fonts.BigFont.PrintOptions(screen, float64(160), float64(8), font.FontOptions{Justify: font.FontJustifyCenter, Options: &options, Scale: scale.ScaleAmount}, fmt.Sprintf("The Armies Of %v", view.Player.Wizard.Name))

            if highlightedUnit != nil {
                raceName := highlightedUnit.GetRace().String()
                fonts.NormalFont.PrintOptions(screen, float64(190), float64(162), font.FontOptions{Justify: font.FontJustifyCenter, Options: &options, Scale: scale.ScaleAmount}, fmt.Sprintf("%v %v", raceName, highlightedUnit.GetName()))

            }

            shadow := font.FontOptions{Justify: font.FontJustifyCenter, DropShadow: true, Scale: scale.ScaleAmount}
            fonts.NormalFont.PrintOptions(screen, float64(30), float64(162), shadow, "UPKEEP")
            fonts.NormalFont.PrintOptions(screen, float64(45), float64(172), shadow, fmt.Sprintf("%v", upkeepGold))
            fonts.NormalFont.PrintOptions(screen, float64(45), float64(182), shadow, fmt.Sprintf("%v", upkeepMana))
            fonts.NormalFont.PrintOptions(screen, float64(45), float64(192), shadow, fmt.Sprintf("%v", upkeepFood))

            minimapRect := image.Rect(85, 163, 135, 197)
            minimapArea := screen.SubImage(scale.ScaleRect(minimapRect)).(*ebiten.Image)

            if highlightedUnit != nil {
                view.DrawMinimap(minimapArea, highlightedUnit.GetX(), highlightedUnit.GetY(), view.Player.GetFog(highlightedUnit.GetPlane()), highlightedUnit.GetPlane(), this.Counter)
            } else {
                // just choose random point
                view.DrawMinimap(minimapArea, 10, 10, view.Player.GetFog(data.PlaneArcanus), data.PlaneArcanus, this.Counter)
            }

            this.StandardDraw(screen)

            // vector.DrawFilledRect(minimapArea, float32(minimapRect.Min.X), float32(minimapRect.Min.Y), float32(minimapRect.Bounds().Dx()), float32(minimapRect.Bounds().Dy()), util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0, B: 0, A: 128}), false)
        },
    }

    ui.SetElementsFromArray(nil)

    // arcanus filter
    arcanusRect := image.Rect(0, 0, int(fonts.SmallerFont.MeasureTextWidth("Arcanus", 1)), fonts.SmallerFont.Height()).Add(image.Pt(77, 18))
    ui.AddElement(&uilib.UIElement{
        Rect: arcanusRect,
        LeftClick: func (this *uilib.UIElement){
            view.ShowArcanus = !view.ShowArcanus
            view.UI = view.MakeUI()
        },
        Inside: func (this *uilib.UIElement, x, y int){
            view.arcanusCounter = min(view.arcanusCounter + 0.03, 0.5)
        },
        NotInside: func (this *uilib.UIElement) {
            view.arcanusCounter = max(view.arcanusCounter - 0.03, 0)
        },
        Draw: func(this *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            if !view.ShowArcanus {
                options.ColorScale.Scale(0.5 + view.arcanusCounter, 0.5 + view.arcanusCounter, 0.5, 1)
            } else {
                options.ColorScale.SetR(1 + view.arcanusCounter)
                options.ColorScale.SetG(1 + view.arcanusCounter)
            }
            fonts.SmallerFont.PrintOptions(screen, float64(arcanusRect.Min.X), float64(arcanusRect.Min.Y), font.FontOptions{DropShadow: true, Scale: scale.ScaleAmount, Options: &options}, "Arcanus")
        },
    })

    // myrror filter
    myrrorRect := image.Rect(0, 0, int(fonts.SmallerFont.MeasureTextWidth("Myrror", 1)), fonts.SmallerFont.Height()).Add(image.Pt(arcanusRect.Max.X + 10, arcanusRect.Min.Y))
    ui.AddElement(&uilib.UIElement{
        Rect: myrrorRect,
        LeftClick: func (this *uilib.UIElement){
            view.ShowMyrror = !view.ShowMyrror
            view.UI = view.MakeUI()
        },
        Inside: func (this *uilib.UIElement, x, y int){
            view.myrrorCounter = min(view.myrrorCounter + 0.03, 0.5)
        },
        NotInside: func (this *uilib.UIElement) {
            view.myrrorCounter = max(view.myrrorCounter - 0.03, 0)
        },
        Draw: func(this *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            if !view.ShowMyrror {
                options.ColorScale.Scale(0.5 + view.myrrorCounter, 0.5 + view.myrrorCounter, 0.5, 1)
            } else {
                options.ColorScale.SetR(1 + view.myrrorCounter)
                options.ColorScale.SetG(1 + view.myrrorCounter)
            }
            fonts.SmallerFont.PrintOptions(screen, float64(myrrorRect.Min.X), float64(myrrorRect.Min.Y), font.FontOptions{DropShadow: true, Scale: scale.ScaleAmount, Options: &options}, "Myrror")
        },
    })

    makeButton := func (x int, y int, normal *ebiten.Image, clickImage *ebiten.Image, action func()) *uilib.UIElement {

        clicked := false

        return &uilib.UIElement{
            Rect: util.ImageRect(x, y, normal),
            LeftClick: func (this *uilib.UIElement){
                clicked = true
            },
            LeftClickRelease: func (this *uilib.UIElement){
                action()
                clicked = false
            },
            Draw: func(this *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(x), float64(y))
                use := normal
                if clicked {
                    use = clickImage
                }
                scale.DrawScaled(screen, use, &options)
            },
        }
    }

    var resetUnits func()

    itemButtons, _ := view.ImageCache.GetImages("armylist.lbx", 3)
    ui.AddElement(makeButton(273, 163, itemButtons[0], itemButtons[1], func(){
        view.ShowVault()
        resetUnits()
    }))

    okButtons, _ := view.ImageCache.GetImages("armylist.lbx", 4)
    ui.AddElement(makeButton(273, 183, okButtons[0], okButtons[1], func(){
        view.State = ArmyScreenStateDone
    }))

    scrollUnitsUp := func(){
        if view.FirstRow > 0 {
            view.FirstRow -= 1
            view.UI = view.MakeUI()
        }
    }

    scrollUnitsDown := func(){
        totalStacks := len(view.Player.Stacks)
        if view.FirstRow < totalStacks - 6 {
            view.FirstRow += 1
            view.UI = view.MakeUI()
        }
    }

    ui.AddElement(makeButton(60, 26, upArrows[0], upArrows[1], scrollUnitsUp))
    ui.AddElement(makeButton(250, 26, upArrows[0], upArrows[1], scrollUnitsUp))

    ui.AddElement(makeButton(60, 139, downArrows[0], downArrows[1], scrollUnitsDown))
    ui.AddElement(makeButton(250, 139, downArrows[0], downArrows[1], scrollUnitsDown))

    highlightColor := util.PremultiplyAlpha(color.RGBA{R: 255, G: 255, B: 255, A: 90})

    banner := view.Player.Wizard.Banner

    var unitElements []*uilib.UIElement
    resetUnits = func(){
        ui.RemoveElements(unitElements)

        // row := view.FirstRow
        rowY := 25
        rowCount := 0

        // recompute upkeep values
        upkeepGold = view.Player.TotalUnitUpkeepGold()
        upkeepFood = view.Player.TotalUnitUpkeepFood()
        upkeepMana = view.Player.TotalUnitUpkeepMana()

        for i, hero := range view.Player.AliveHeroes() {
            x := (12 + (i % 2) * 265)
            y := (5 + (i / 2) * 51)

            portraitLbx, portraitIndex := hero.GetPortraitLbxInfo()
            pic, _ := view.ImageCache.GetImage(portraitLbx, portraitIndex, 0)

            rect := util.ImageRect(x, y, pic)

            var disband func()

            heroElement := &uilib.UIElement{
                Rect: rect,
                RightClick: func (this *uilib.UIElement){
                    ui.AddGroup(unitview.MakeUnitContextMenu(view.Cache, ui, hero, disband))
                },
                LeftClick: func (this *uilib.UIElement){
                    stack := view.Player.FindStackByUnit(hero)
                    if stack != nil {
                        view.Player.SelectedStack = stack
                        view.State = ArmyScreenStateDone
                    }
                },
                Inside: func (this *uilib.UIElement, x, y int){
                    highlightedUnit = hero
                },
                Draw: func(this *uilib.UIElement, screen *ebiten.Image){
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(x), float64(y))
                    scale.DrawScaled(screen, pic, &options)

                    options.GeoM.Translate(0, float64(pic.Bounds().Dy()))

                    nameX, nameY := options.GeoM.Apply(0, 0)

                    fonts.SmallerFont.PrintOptions(screen, nameX + float64(15), nameY + float64(6), font.FontOptions{Justify: font.FontJustifyCenter, Options: &options, Scale: scale.ScaleAmount}, hero.GetName())
                },
            }

            disband = func(){
                ui.RemoveElement(heroElement)
                view.Player.RemoveUnit(hero)
                resetUnits()
            }

            ui.AddElement(heroElement)
            unitElements = append(unitElements, heroElement)
        }

        for i, stack := range view.Player.Stacks {
            if i < view.FirstRow {
                continue
            }

            if stack.Plane() == data.PlaneArcanus && !view.ShowArcanus {
                continue
            }

            if stack.Plane() == data.PlaneMyrror && !view.ShowMyrror {
                continue
            }

            x := 78

            for _, unit := range stack.Units() {
                elementX := float64(x)
                elementY := float64(rowY)

                if highlightedUnit == nil {
                    highlightedUnit = unit
                }
                pic, _ := view.ImageCache.GetImageTransform(unit.GetLbxFile(), unit.GetLbxIndex(), 0, banner.String(), units.MakeUpdateUnitColorsFunc(banner))
                if pic != nil {
                    element := &uilib.UIElement{
                        Rect: util.ImageRect(int(elementX), int(elementY), pic),
                        LeftClick: func (this *uilib.UIElement){
                            view.Player.SelectedStack = stack
                            view.State = ArmyScreenStateDone
                        },
                        RightClick: func (this *uilib.UIElement){
                            disband := func(){
                                view.Player.RemoveUnit(unit)
                                resetUnits()
                            }
                            ui.AddGroup(unitview.MakeUnitContextMenu(view.Cache, ui, unit, disband))
                        },
                        Inside: func (this *uilib.UIElement, x, y int){
                            highlightedUnit = unit
                        },
                        Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                            var options colorm.DrawImageOptions
                            var matrix colorm.ColorM
                            options.GeoM.Translate(elementX, elementY)
                            options.GeoM.Scale(scale.ScaleAmount, scale.ScaleAmount)

                            if highlightedUnit == unit {
                                x, y := options.GeoM.Apply(0, 0)
                                x2, y2 := options.GeoM.Apply(float64(pic.Bounds().Dx()), float64(pic.Bounds().Dy()))
                                vector.FillRect(screen, float32(x), float32(y+1), float32(x2-x), float32(y2-y)-1, highlightColor, false)
                            }

                            if unit.GetBusy() == units.BusyStatusPatrol || unit.GetBusy() == units.BusyStatusStasis {
                                matrix.ChangeHSV(0, 0, 1)
                            }

                            colorm.DrawImage(screen, pic, matrix, &options)

                            enchantment := util.First(unit.GetEnchantments(), data.UnitEnchantmentNone)
                            if enchantment != data.UnitEnchantmentNone {
                                util.DrawOutline(screen, &view.ImageCache, pic, options.GeoM, ebiten.ColorScale{}, ui.Counter/10, enchantment.Color())
                            }

                        },
                    }
                    ui.AddElement(element)
                    unitElements = append(unitElements, element)
                    x += pic.Bounds().Dx() + 1
                }
            }

            // there are only 6 slots to show at a time
            rowCount += 1
            if rowCount >= 6 {
                break
            }

            rowY += 22
        }
    }

    resetUnits()

    return ui
}

func (view *ArmyScreen) Update() ArmyScreenState {
    view.UI.StandardUpdate()
    return view.State
}

func (view *ArmyScreen) Draw(screen *ebiten.Image) {
    view.UI.Draw(view.UI, screen)

}
