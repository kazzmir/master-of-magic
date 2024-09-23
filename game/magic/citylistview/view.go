package citylistview

import (
    "log"
    "fmt"
    "slices"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/hajimehoshi/ebiten/v2"
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
}

func MakeCityListScreen(cache *lbx.LbxCache, player *playerlib.Player) *CityListScreen {
    view := &CityListScreen{
        Cache: cache,
        Player: player,
        ImageCache: util.MakeImageCache(cache),
        State: CityListScreenStateRunning,
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

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image) {
            background, _ := view.ImageCache.GetImage("reload.lbx", 21, 0)
            var options ebiten.DrawImageOptions
            screen.DrawImage(background, &options)

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })

            bigFont.PrintCenter(screen, 160, 5, 1, ebiten.ColorScale{}, fmt.Sprintf("The Cities Of %v", view.Player.Wizard.Name))

            y := float64(17)
            x := float64(31)
            normalFont.Print(screen, x, y, 1, ebiten.ColorScale{}, "Name")
            normalFont.Print(screen, x + 57, y, 1, ebiten.ColorScale{}, "Race")
            normalFont.PrintRight(screen, x + 119, y, 1, ebiten.ColorScale{}, "Pop")
            normalFont.PrintRight(screen, x + 139, y, 1, ebiten.ColorScale{}, "Gold")
            normalFont.PrintRight(screen, x + 159, y, 1, ebiten.ColorScale{}, "Prd")
            normalFont.Print(screen, x + 165, y, 1, ebiten.ColorScale{}, "Producing")
            normalFont.PrintRight(screen, x + 258, y, 1, ebiten.ColorScale{}, "Time")
        },
    }

    var elements []*uilib.UIElement

    cities := slices.Clone(view.Player.Cities)
    slices.SortFunc(cities, func(a *citylib.City, b *citylib.City) int {
        if a.BirthTurn < b.BirthTurn {
            return 1
        }

        if a.BirthTurn > b.BirthTurn {
            return -1
        }

        return 0
    })

    y := 28
    for _, city := range cities {
        elementY := float64(y)
        elements = append(elements, &uilib.UIElement{
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                x := float64(31)
                normalFont.Print(screen, x, elementY, 1, ebiten.ColorScale{}, city.Name)
                normalFont.Print(screen, x + 57, elementY, 1, ebiten.ColorScale{}, city.Race.String())
                normalFont.PrintRight(screen, x + 119, elementY, 1, ebiten.ColorScale{}, fmt.Sprintf("%v", city.Citizens()))
                normalFont.PrintRight(screen, x + 139, elementY, 1, ebiten.ColorScale{}, fmt.Sprintf("%v", city.GoldSurplus()))
                normalFont.PrintRight(screen, x + 159, elementY, 1, ebiten.ColorScale{}, fmt.Sprintf("%v", int(city.WorkProductionRate())))
                normalFont.Print(screen, x + 165, elementY, 1, ebiten.ColorScale{}, city.ProducingString())
                normalFont.PrintRight(screen, x + 258, elementY, 1, ebiten.ColorScale{}, fmt.Sprintf("%v", city.BirthTurn))
            },
        })

        y += 14
    }

    ui.SetElementsFromArray(elements)

    return ui
}

func (view *CityListScreen) Update() CityListScreenState {
    view.UI.StandardUpdate()
    return view.State
}

func (view *CityListScreen) Draw(screen *ebiten.Image) {
    view.UI.Draw(view.UI, screen)
}
