package cartographer

import (
    "log"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"

    "github.com/hajimehoshi/ebiten/v2"
)

type Fonts struct {
    Name *font.Font
    Title *font.Font
}

func makeFonts(cache *lbx.LbxCache) Fonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Error: could not load fonts: %v", err)
        return Fonts{}
    }

    return Fonts{
        Name: loader(fontslib.SmallBlack),
        Title: loader(fontslib.BigBlack),
    }
}

func MakeCartographer(cache *lbx.LbxCache, cities []*citylib.City, arcanusMap *maplib.Map, arcanusFog data.FogMap, myrrorMap *maplib.Map, myrrorFog data.FogMap) (coroutine.AcceptYieldFunc, func (*ebiten.Image)) {
    quit := false

    fonts := makeFonts(cache)

    imageCache := util.MakeImageCache(cache)

    counter := uint64(0)

    getAlpha := util.MakeFadeIn(7, &counter)

    currentPlane := data.PlaneArcanus

    ui := &uilib.UI{
        Draw: func (ui *uilib.UI, screen *ebiten.Image) {
            background, _ := imageCache.GetImage("reload.lbx", 2, 0)
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, background, &options)

            planeName := "Arcanus Plane"
            if currentPlane == data.PlaneMyrror {
                planeName = "Myrror Plane"
            }
            fonts.Title.PrintOptions(screen, float64(background.Bounds().Dx() / 2), 10, font.FontOptions{Scale: scale.ScaleAmount, Options: &options, Justify: font.FontJustifyCenter}, planeName)

            ui.StandardDraw(screen)
        },
    }

    ui.SetElementsFromArray(nil)

    pageTurnRect := image.Rect(283, 173, 315, 198)
    ui.AddElement(&uilib.UIElement{
        Rect: pageTurnRect,
        Draw: func (element *uilib.UIElement, screen *ebiten.Image) {
            util.DrawRect(screen, scale.ScaleRect(pageTurnRect), color.RGBA{R: 255, A: 255})
        },
        LeftClick: func (element *uilib.UIElement) {
            currentPlane = currentPlane.Opposite()
        },
        NotLeftClicked: func(element *uilib.UIElement){
            quit = true
        },
    })

    logic := func (yield coroutine.YieldFunc) error {
        for !quit {
            counter += 1

            ui.StandardUpdate()

            err := yield()
            if err != nil {
                return err
            }
        }

        getAlpha = util.MakeFadeOut(7, &counter)
        for range 7 {
            counter += 1
            yield()
        }

        return nil
    }

    draw := func (screen *ebiten.Image) {
        ui.Draw(ui, screen)
    }

    return logic, draw
}
