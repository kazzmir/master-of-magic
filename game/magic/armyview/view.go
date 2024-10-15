package armyview

import (
    "log"
    "fmt"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
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
    DrawMinimap func(*ebiten.Image, int, int, [][]bool, uint64)
}

func MakeArmyScreen(cache *lbx.LbxCache, player *playerlib.Player, drawMinimap func(*ebiten.Image, int, int, [][]bool, uint64), showVault func()) *ArmyScreen {
    view := &ArmyScreen{
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        Player: player,
        State: ArmyScreenStateRunning,
        ShowVault: showVault,
        DrawMinimap: drawMinimap,
        FirstRow: 0,
    }

    view.UI = view.MakeUI()

    return view
}

func (view *ArmyScreen) MakeUI() *uilib.UI {
    var highlightedUnit *units.OverworldUnit

    fontLbx, err := view.Cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

    normalColor := util.Lighten(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, -30)
    normalPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        normalColor, normalColor, normalColor,
        normalColor, normalColor, normalColor,
    }
    normalFont := font.MakeOptimizedFontWithPalette(fonts[1], normalPalette)

    // red := color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}

    yellow := color.RGBA{R: 0xf9, G: 0xdb, B: 0x4c, A: 0xff}
    bigPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        util.RotateHue(yellow, -0.50),
        util.RotateHue(yellow, -0.30),
        util.RotateHue(yellow, -0.10),
        yellow,
    }
    bigFont := font.MakeOptimizedFontWithPalette(fonts[4], bigPalette)

    upArrows, _ := view.ImageCache.GetImages("armylist.lbx", 1)
    downArrows, _ := view.ImageCache.GetImages("armylist.lbx", 2)

    ui := &uilib.UI{
        Draw: func(this *uilib.UI, screen *ebiten.Image) {
            background, _ := view.ImageCache.GetImage("armylist.lbx", 0, 0)
            var options ebiten.DrawImageOptions
            screen.DrawImage(background, &options)

            this.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })

            bigFont.PrintCenter(screen, 160, 10, 1, options.ColorScale, fmt.Sprintf("The Armies Of %v", view.Player.Wizard.Name))

            if highlightedUnit != nil {
                raceName := highlightedUnit.Unit.Race.String()
                normalFont.PrintCenter(screen, 190, 162, 1, options.ColorScale, fmt.Sprintf("%v %v", raceName, highlightedUnit.Unit.Name))

                normalFont.PrintCenter(screen, 30, 162, 1, options.ColorScale, "UPKEEP")

                normalFont.PrintCenter(screen, 45, 170, 1, options.ColorScale, fmt.Sprintf("%v", highlightedUnit.Unit.UpkeepGold))
                normalFont.PrintCenter(screen, 45, 180, 1, options.ColorScale, fmt.Sprintf("%v", highlightedUnit.Unit.UpkeepMana))
                normalFont.PrintCenter(screen, 45, 190, 1, options.ColorScale, fmt.Sprintf("%v", highlightedUnit.Unit.UpkeepFood))
            }

            minimapRect := image.Rect(85, 163, 135, 197)
            minimapArea := screen.SubImage(minimapRect).(*ebiten.Image)

            if highlightedUnit != nil {
                view.DrawMinimap(minimapArea, highlightedUnit.X, highlightedUnit.Y, view.Player.GetFog(highlightedUnit.Plane), this.Counter)
            } else {
                // just choose random point
                view.DrawMinimap(minimapArea, 10, 10, view.Player.GetFog(data.PlaneArcanus), this.Counter)
            }

            // vector.DrawFilledRect(minimapArea, float32(minimapRect.Min.X), float32(minimapRect.Min.Y), float32(minimapRect.Bounds().Dx()), float32(minimapRect.Bounds().Dy()), util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0, B: 0, A: 128}), false)
        },
    }

    var elements []*uilib.UIElement

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
                screen.DrawImage(use, &options)
            },
        }
    }

    itemButtons, _ := view.ImageCache.GetImages("armylist.lbx", 3)
    elements = append(elements, makeButton(273, 163, itemButtons[0], itemButtons[1], func(){
        view.ShowVault()
    }))

    okButtons, _ := view.ImageCache.GetImages("armylist.lbx", 4)
    elements = append(elements, makeButton(273, 183, okButtons[0], okButtons[1], func(){
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

    elements = append(elements, makeButton(60, 26, upArrows[0], upArrows[1], scrollUnitsUp))
    elements = append(elements, makeButton(250, 26, upArrows[0], upArrows[1], scrollUnitsUp))

    elements = append(elements, makeButton(60, 139, downArrows[0], downArrows[1], scrollUnitsDown))
    elements = append(elements, makeButton(250, 139, downArrows[0], downArrows[1], scrollUnitsDown))

    // FIXME: show 6 heroes, 3 on either side

    // row := view.FirstRow
    rowY := 25
    rowCount := 0

    highlightColor := util.PremultiplyAlpha(color.RGBA{R: 255, G: 255, B: 255, A: 90})

    banner := view.Player.Wizard.Banner

    for i, stack := range view.Player.Stacks {
        if i < view.FirstRow {
            continue
        }

        x := 78

        for _, unit := range stack.Units() {
            elementX := float64(x)
            elementY := float64(rowY)

            if highlightedUnit == nil {
                highlightedUnit = unit
            }
            pic, _ := view.ImageCache.GetImageTransform(unit.Unit.LbxFile, unit.Unit.Index, 0, banner.String(), units.MakeUpdateUnitColorsFunc(banner))
            if pic != nil {
                elements = append(elements, &uilib.UIElement{
                    Rect: util.ImageRect(int(elementX), int(elementY), pic),
                    LeftClick: func (this *uilib.UIElement){
                        view.Player.SelectedStack = stack
                        view.State = ArmyScreenStateDone
                    },
                    Inside: func (this *uilib.UIElement, x, y int){
                        highlightedUnit = unit
                    },
                    Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                        var options colorm.DrawImageOptions
                        var matrix colorm.ColorM
                        options.GeoM.Translate(elementX, elementY)

                        if highlightedUnit == unit {
                            x, y := options.GeoM.Apply(0, 0)
                            vector.DrawFilledRect(screen, float32(x), float32(y+1), float32(pic.Bounds().Dx()), float32(pic.Bounds().Dy())-1, highlightColor, false)
                        }

                        if unit.Patrol {
                            matrix.ChangeHSV(0, 0, 1)
                        }

                        colorm.DrawImage(screen, pic, matrix, &options)
                    },
                })
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

    ui.SetElementsFromArray(elements)

    return ui
}

func (view *ArmyScreen) Update() ArmyScreenState {
    view.UI.StandardUpdate()
    return view.State
}

func (view *ArmyScreen) Draw(screen *ebiten.Image) {
    view.UI.Draw(view.UI, screen)

}
