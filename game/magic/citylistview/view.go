package citylistview

import (
    "log"
    "fmt"
    "slices"
    "strings"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/cityview"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

type CityListScreenState int

const (
    CityListScreenStateRunning CityListScreenState = iota
    CityListScreenStateDone
)

type CityListScreen struct {
    Cache *lbx.LbxCache
    Player *playerlib.Player
    ImageCache util.ImageCache
    UI *uilib.UI
    State CityListScreenState
    CurrentBuildScreen *cityview.BuildScreen
    BuildScreenUpdate func()
    DrawMinimap func(*ebiten.Image, int, int, data.Plane, uint64)
    DoSelectCity func(*citylib.City)
    FirstRow int
}

func MakeCityListScreen(cache *lbx.LbxCache, player *playerlib.Player, drawMinimap func(*ebiten.Image, int, int, data.Plane, uint64), selectCity func(*citylib.City)) *CityListScreen {
    view := &CityListScreen{
        Cache: cache,
        Player: player,
        ImageCache: util.MakeImageCache(cache),
        State: CityListScreenStateRunning,
        DrawMinimap: drawMinimap,
        DoSelectCity: selectCity,
        FirstRow: 0,
    }

    view.UI = view.MakeUI()

    return view
}

func (view *CityListScreen) MakeUI() *uilib.UI {
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

    var highlightedCity *citylib.City

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image) {
            background, _ := view.ImageCache.GetImage("reload.lbx", 21, 0)
            var options ebiten.DrawImageOptions
            scale.DrawScaled(screen, background, &options)

            ui.StandardDraw(screen)

            bigFont.PrintCenter(screen, float64(160), float64(5), scale.ScaleAmount, ebiten.ColorScale{}, fmt.Sprintf("The Cities Of %v", view.Player.Wizard.Name))

            y := float64(17)
            x := float64(31)
            normalFont.Print(screen, x, y, scale.ScaleAmount, ebiten.ColorScale{}, "Name")
            normalFont.Print(screen, x + 57, y, scale.ScaleAmount, ebiten.ColorScale{}, "Race")
            normalFont.PrintRight(screen, x + float64(119), y, scale.ScaleAmount, ebiten.ColorScale{}, "Pop")
            normalFont.PrintRight(screen, x + float64(139), y, scale.ScaleAmount, ebiten.ColorScale{}, "Gold")
            normalFont.PrintRight(screen, x + float64(159), y, scale.ScaleAmount, ebiten.ColorScale{}, "Prd")
            normalFont.Print(screen, x + float64(165), y, scale.ScaleAmount, ebiten.ColorScale{}, "Producing")
            normalFont.PrintRight(screen, x + float64(258), y, scale.ScaleAmount, ebiten.ColorScale{}, "Time")

            normalFont.Print(screen, 232, 173, scale.ScaleAmount, ebiten.ColorScale{}, fmt.Sprintf("%vGP", view.Player.Gold))
            normalFont.Print(screen, 267, 173, scale.ScaleAmount, ebiten.ColorScale{}, fmt.Sprintf("%vMP", view.Player.Mana))

            if highlightedCity != nil {
                normalFont.Print(screen, float64(99), float64(158), scale.ScaleAmount, ebiten.ColorScale{}, highlightedCity.Name)
                minimapRect := image.Rect(42, 162, 91, 195)
                minimapArea := screen.SubImage(scale.ScaleRect(minimapRect)).(*ebiten.Image)
                view.DrawMinimap(minimapArea, highlightedCity.X, highlightedCity.Y, highlightedCity.Plane, ui.Counter)
            // vector.DrawFilledRect(minimapArea, float32(minimapRect.Min.X), float32(minimapRect.Min.Y), float32(minimapRect.Bounds().Dx()), float32(minimapRect.Bounds().Dy()), util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0, B: 0, A: 128}), false)
            }
        },
    }

    var elements []*uilib.UIElement

    cities := slices.Clone(view.Player.Cities)
    slices.SortFunc(cities, func(a *citylib.City, b *citylib.City) int {
        return strings.Compare(a.Name, b.Name)
    })

    highlightColor := util.PremultiplyAlpha(color.RGBA{R: 255, G: 255, B: 255, A: 90})

    maxRows := 9
    y := 28
    rowCount := 0
    for i, city := range cities {

        if i < view.FirstRow {
            continue
        }

        if highlightedCity == nil {
            highlightedCity = city
        }

        goldSurplus := city.GoldSurplus()

        elementY := float64(y)
        elements = append(elements, &uilib.UIElement{
            Rect: image.Rect(28, int(elementY), 296, int(elementY) + 14),
            LeftClickRelease: func(element *uilib.UIElement){
                view.DoSelectCity(city)
                view.State = CityListScreenStateDone
            },
            RightClick: func(element *uilib.UIElement){
                buildScreen := cityview.MakeBuildScreen(view.Cache, city)
                view.CurrentBuildScreen = buildScreen
                view.BuildScreenUpdate = func(){
                    city.ProducingBuilding = buildScreen.ProducingBuilding
                    city.ProducingUnit = buildScreen.ProducingUnit
                }
            },
            Inside: func(element *uilib.UIElement, x int, y int){
                highlightedCity = city
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                x := float64(31)

                if highlightedCity == city {
                    vector.FillRect(screen, scale.Scale(float32((x-1))), scale.Scale(float32(elementY - 3)), scale.Scale(float32(52)), scale.Scale(float32(10)), highlightColor, false)
                    vector.FillRect(screen, scale.Scale(float32((x-1+57))), scale.Scale(float32(elementY-3)), scale.Scale(float32(44)), scale.Scale(float32(10)), highlightColor, false)
                    vector.FillRect(screen, scale.Scale(float32((x-1+119-14))), scale.Scale(float32(elementY-3)), scale.Scale(float32(16)), scale.Scale(float32(10)), highlightColor, false)
                    vector.FillRect(screen, scale.Scale(float32((x-1+139-14))), scale.Scale(float32(elementY-3)), scale.Scale(float32(16)), scale.Scale(float32(10)), highlightColor, false)
                    vector.FillRect(screen, scale.Scale(float32((x-1+159-14))), scale.Scale(float32(elementY-3)), scale.Scale(float32(16)), scale.Scale(float32(10)), highlightColor, false)
                    vector.FillRect(screen, scale.Scale(float32((x-1+165))), scale.Scale(float32(elementY-3)), scale.Scale(float32(76)), scale.Scale(float32(10)), highlightColor, false)
                    vector.FillRect(screen, scale.Scale(float32((x-1+258-13))), scale.Scale(float32(elementY-3)), scale.Scale(float32(15)), scale.Scale(float32(10)), highlightColor, false)
                }

                normalFont.Print(screen, x, elementY, scale.ScaleAmount, ebiten.ColorScale{}, city.Name)
                normalFont.Print(screen, (x + 57), elementY, scale.ScaleAmount, ebiten.ColorScale{}, city.Race.String())
                normalFont.PrintRight(screen, (x + 119), elementY, scale.ScaleAmount, ebiten.ColorScale{}, fmt.Sprintf("%v", city.Citizens()))
                normalFont.PrintRight(screen, (x + 139), elementY, scale.ScaleAmount, ebiten.ColorScale{}, fmt.Sprintf("%v", goldSurplus))
                normalFont.PrintRight(screen, (x + 159), elementY, scale.ScaleAmount, ebiten.ColorScale{}, fmt.Sprintf("%v", int(city.WorkProductionRate())))
                normalFont.Print(screen, (x + 165), elementY, scale.ScaleAmount, ebiten.ColorScale{}, city.ProducingString())
                normalFont.PrintRight(screen, (x + 258), elementY, scale.ScaleAmount, ebiten.ColorScale{}, fmt.Sprintf("%v", city.ProducingTurnsLeft()))
            },
        })

        y += 14

        rowCount += 1
        if rowCount >= maxRows {
            break
        }
    }

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
                options.GeoM.Translate(float64(this.Rect.Min.X), float64(this.Rect.Min.Y))
                use := normal
                if clicked {
                    use = clickImage
                }
                scale.DrawScaled(screen, use, &options)
            },
        }
    }

    okButtons, _ := view.ImageCache.GetImages("reload.lbx", 22)
    elements = append(elements, makeButton(239, 183, okButtons[0], okButtons[1], func(){
        view.State = CityListScreenStateDone
    }))

    upArrows, _ := view.ImageCache.GetImages("armylist.lbx", 1)
    downArrows, _ := view.ImageCache.GetImages("armylist.lbx", 2)

    scrollUpFunc := func(){
        if view.FirstRow > 0 {
            view.FirstRow -= 1
            view.UI = view.MakeUI()
        }
    }

    scrollDownFunc := func(){
        if view.FirstRow < len(cities) - maxRows {
            view.FirstRow += 1
            view.UI = view.MakeUI()
        }
    }

    elements = append(elements, makeButton(11, 27, upArrows[0], upArrows[1], scrollUpFunc))
    elements = append(elements, makeButton(299, 27, upArrows[0], upArrows[1], scrollUpFunc))

    elements = append(elements, makeButton(11, 138, downArrows[0], downArrows[1], scrollDownFunc))
    elements = append(elements, makeButton(299, 138, downArrows[0], downArrows[1], scrollDownFunc))

    ui.SetElementsFromArray(elements)

    return ui
}

func (view *CityListScreen) Update() CityListScreenState {
    if view.CurrentBuildScreen != nil {
        switch view.CurrentBuildScreen.Update() {
            case cityview.BuildScreenRunning:
            case cityview.BuildScreenOk:
                view.BuildScreenUpdate()
                view.CurrentBuildScreen = nil
            case cityview.BuildScreenCanceled:
                view.CurrentBuildScreen = nil
        }
    } else {
        view.UI.StandardUpdate()
    }
    return view.State
}

func (view *CityListScreen) Draw(screen *ebiten.Image) {
    view.UI.Draw(view.UI, screen)
    if view.CurrentBuildScreen != nil {
        view.CurrentBuildScreen.Draw(screen)
    }
}
