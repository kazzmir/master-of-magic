package settings

import (
    "context"
    "fmt"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    musiclib "github.com/kazzmir/master-of-magic/game/magic/music"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

func MakeSettingsUI(cache *lbx.LbxCache, imageCache *util.ImageCache, music *musiclib.Music) (*uilib.UIElementGroup, context.Context) {
    fonts := fontslib.MakeSettingsFonts(cache)

    group := uilib.MakeGroup()
    quit, cancel := context.WithCancel(context.Background())

    background, _ := imageCache.GetImage("load.lbx", 11, 0)

    getAlpha := group.MakeFadeIn(7)

    group.AddElement(&uilib.UIElement{
        Layer: 4,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var backgroundOptions ebiten.DrawImageOptions
            backgroundOptions.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, background, &backgroundOptions)
        },
    })

    ok, _ := imageCache.GetImage("load.lbx", 4, 0)

    settingsLayer := uilib.UILayer(5)

    group.AddElement(&uilib.UIElement{
        Layer: settingsLayer,
        Rect: util.ImageRect(266, 176, ok),
        LeftClick: func(element *uilib.UIElement){
            getAlpha = group.MakeFadeOut(7)
            group.AddDelay(7, func(){
                cancel()
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, ok, &options)
        },
    })

    group.AddElement(&uilib.UIElement{
        Layer: settingsLayer,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            fonts.OptionFont.PrintOptions(screen, 30, 40, font.FontOptions{Scale: scale.ScaleAmount, DropShadow: true, Options: &options}, fmt.Sprintf("Volume: %02d%%", int(music.GetVolume() * 100)))
        },
    })

    slider, _ := imageCache.GetImage("spellscr.lbx", 3, 0)

    volumeClicked := false
    group.AddElement(&uilib.UIElement{
        Layer: settingsLayer,
        Rect: image.Rect(30, 50, 30 + 80, 50 + slider.Bounds().Dy()),
        Inside: func(this *uilib.UIElement, x int, y int){
            if volumeClicked {
                music.SetVolume(min(1, float64(x) / float64(this.Rect.Dx() - 1)))
            }
        },
        LeftClick: func(element *uilib.UIElement){
            volumeClicked = true
        },
        LeftClickRelease: func(element *uilib.UIElement){
            volumeClicked = false
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            backgroundRect := element.Rect
            backgroundRect.Max.X += 5
            backgroundRect.Min.X -= 1
            backgroundRect.Min.Y -= 1

            vector.FillRect(screen, float32(scale.Scale(backgroundRect.Min.X)), float32(scale.Scale(backgroundRect.Min.Y)), float32(scale.Scale(backgroundRect.Dx())), float32(scale.Scale(backgroundRect.Dy())), color.NRGBA{R: 32, G: 32, B: 32, A: uint8(200 * getAlpha())}, false)
            util.DrawRect(screen, scale.ScaleRect(backgroundRect), color.NRGBA{R: 255, G: 255, B: 255, A: uint8(200 * getAlpha())})

            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            options.GeoM.Translate(float64(element.Rect.Min.X) + float64(element.Rect.Dx()) * music.GetVolume(), float64(element.Rect.Min.Y))
            options.GeoM.Translate(float64(-slider.Bounds().Dx()/2), 0)
            scale.DrawScaled(screen, slider, &options)

            // util.DrawRect(screen, scale.ScaleRect(element.Rect), color.RGBA{R: 255, A: 255})
        },
    })

    return group, quit

    /*
    var elements []*uilib.UIElement

    var makeElements func()

    makeElements = func() {
        *background, _ = imageCache.GetImage("load.lbx", 11, 0)
        ok, _ := imageCache.GetImage("load.lbx", 4, 0)
        ui.RemoveElements(elements)
        elements = nil

        elements = append(elements, &uilib.UIElement{
            Rect: util.ImageRect(266, 176, ok),
            LeftClick: func(element *uilib.UIElement){
                onOk()
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
                scale.DrawScaled(screen, ok, &options)
            },
        })

        resolutionBackground, _ := imageCache.GetImage("load.lbx", 5, 0)

        elements = append(elements, &uilib.UIElement{
            Rect: util.ImageRect(20, 40, resolutionBackground),
            LeftClick: func(element *uilib.UIElement){
                selected := func(name string, scale int, algorithm scale.ScaleAlgorithm) string {
                    / *
                    if data.ScreenScale == scale && data.ScreenScaleAlgorithm == algorithm {
                        return name + "*"
                    }
                    * /
                    return name
                }

                update := func(scale int, algorithm scale.ScaleAlgorithm){
                    / *
                    data.ScreenScale = scale
                    data.ScreenScaleAlgorithm = algorithm
                    data.ScreenWidth = data.ScreenWidthOriginal * scale
                    data.ScreenHeight = data.ScreenHeightOriginal * scale
                    game.UpdateImages()
                    *imageCache = util.MakeImageCache(game.Cache)
                    makeElements()
                    * /
                }

                makeChoices := func (name string, scales []int, algorithm scale.ScaleAlgorithm) []uilib.Selection {
                    var out []uilib.Selection
                    for _, value := range scales {
                        out = append(out, uilib.Selection{
                            Name: selected(fmt.Sprintf("%v %vx", name, value), value, algorithm),
                            Action: func(){
                                update(value, algorithm)
                            },
                        })
                    }
                    return out
                }

                normalChoices := makeChoices("Normal", []int{1, 2, 3, 4}, scale.ScaleAlgorithmNormal)
                scaleChoices := makeChoices("Scale", []int{2, 3, 4}, scale.ScaleAlgorithmScale)
                xbrChoices := makeChoices("XBR", []int{2, 3, 4}, scale.ScaleAlgorithmXbr)

                choices := append(append(normalChoices, scaleChoices...), xbrChoices...)

                ui.AddElements(uilib.MakeSelectionUI(ui, game.Cache, imageCache, 40, 10, "Resolution", choices, true))
            },
            Draw: func (element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
                scale.DrawScaled(screen, resolutionBackground, &options)

                x, y := options.GeoM.Apply(float64(3), float64(3))
                fonts.OptionFont.Print(screen, x, y, scale.ScaleAmount, options.ColorScale, "Screen")
            },
        })

        ui.AddElements(elements)
    }

    makeElements()
    */
}
