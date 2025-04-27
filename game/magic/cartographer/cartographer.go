package cartographer

import (
    "log"
    "image"
    "image/color"
    "math"
    "maps"
    "slices"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/functional"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    // "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/colorm"
)

type Fonts struct {
    Name *font.Font
    Title *font.Font

    BannerFonts map[data.BannerType]*font.Font
}

func makeFonts(cache *lbx.LbxCache) Fonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Error: could not load fonts: %v", err)
        return Fonts{}
    }

    cityViewFonts, err := fontslib.MakeCityViewFonts(cache)
    if err != nil {
        return Fonts{}
    }

    return Fonts{
        Name: loader(fontslib.SmallBlack),
        Title: loader(fontslib.MassiveBlack),
        BannerFonts: cityViewFonts.BannerFonts,
    }
}

func MakeCartographer(cache *lbx.LbxCache, cities []*citylib.City, stacks []*playerlib.UnitStack, players []*playerlib.Player, arcanusMap *maplib.Map, arcanusFog data.FogMap, myrrorMap *maplib.Map, myrrorFog data.FogMap) (coroutine.AcceptYieldFunc, func (*ebiten.Image)) {
    quit := false

    fonts := makeFonts(cache)

    imageCache := util.MakeImageCache(cache)

    counter := uint64(0)

    getAlpha := util.MakeFadeIn(7, &counter)

    currentPlane := data.PlaneArcanus

    playerName := functional.Memoize(func (banner data.BannerType) string {
        for _, player := range players {
            if player.GetBanner() == banner {
                return player.Wizard.Name
            }
        }

        return ""
    })

    usedBanners := make(map[data.BannerType]string)
    for _, city := range cities {
        usedBanners[city.GetBanner()] = playerName(city.GetBanner())
    }

    bannerList := slices.Sorted(maps.Keys(usedBanners))

    getFlag := func (banner data.BannerType) *ebiten.Image {
        index := 3
        switch banner {
            case data.BannerBlue: index = 3
            case data.BannerGreen: index = 4
            case data.BannerPurple: index = 5
            case data.BannerRed: index = 6
            case data.BannerYellow: index = 7
            case data.BannerBrown: index = 8
        }

        flag, _ := imageCache.GetImage("reload.lbx", index, 0)
        return flag
    }

    bannerColor := func (banner data.BannerType) color.RGBA {
        switch banner {
            case data.BannerBlue: return color.RGBA{R: 0x00, G: 0x00, B: 0xff, A: 0xff}
            case data.BannerGreen: return color.RGBA{R: 0x00, G: 0xf0, B: 0x00, A: 0xff}
            case data.BannerPurple: return color.RGBA{R: 0x8f, G: 0x30, B: 0xff, A: 0xff}
            case data.BannerRed: return color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}
            case data.BannerYellow: return color.RGBA{R: 0xff, G: 0xff, B: 0x00, A: 0xff}
            case data.BannerBrown: return color.RGBA{R: 0xce, G: 0x65, B: 0x00, A: 0xff}
        }

        return color.RGBA{}
    }

    tileImage0, _ := arcanusMap.GetTileImage(0, 0, 0)

    renderMap := func (plane data.Plane) *ebiten.Image {
        showMap := ebiten.NewImage(20*11, 18*9)
        showMap.Fill(color.RGBA{A: 0})
        // showMap.Fill(color.RGBA{R: 32, G: 32, B: 32, A: 255})

        useMap := arcanusMap
        useFog := arcanusFog
        if plane == data.PlaneMyrror {
            useMap = myrrorMap
            useFog = myrrorFog
        }

        // log.Printf("tile width: %v height: %v", tileImage0.Bounds().Dx(), tileImage0.Bounds().Dy())

        scaleX := float64(showMap.Bounds().Dx()) / float64(useMap.Width() * tileImage0.Bounds().Dx())
        scaleY := float64(showMap.Bounds().Dy()) / float64(useMap.Height() * tileImage0.Bounds().Dy())

        var options colorm.DrawImageOptions
        var matrix colorm.ColorM
        // matrix.ScaleWithColor(color.RGBA{R: 217, G: 112, B: 61, A: 255})
        matrix.ChangeHSV(0, 0, 1.8)
        // matrix.Translate(155/255.0, 80/255.0, 44/255.0, 0)
        // a brownish color
        matrix.Scale(0.7, 0.5, 0.3, 1)
        for x := range useFog {
            for y := range useFog[x] {
                if useFog[x][y] != data.FogTypeUnexplored {
                    tileImage, err := useMap.GetTileImage(x, y, 0)
                    if err == nil {
                        options.GeoM.Reset()
                        options.GeoM.Translate(float64(x*tileImage.Bounds().Dx()), float64(y*tileImage.Bounds().Dy()))
                        options.GeoM.Scale(scaleX, scaleY)
                        colorm.DrawImage(showMap, tileImage, matrix, &options)
                    }
                }
            }
        }

        // draw squares for cities
        for _, city := range cities {
            if city.Plane == plane {
                if useFog[city.X][city.Y] != data.FogTypeUnexplored {
                    options.GeoM.Reset()
                    options.GeoM.Translate(float64(city.X*tileImage0.Bounds().Dx()), float64(city.Y*tileImage0.Bounds().Dy()))
                    options.GeoM.Scale(scaleX, scaleY)

                    size := 20.0

                    x1, y1 := options.GeoM.Apply(0, 0)
                    x2, y2 := options.GeoM.Apply(size, size)

                    offset := 7.0
                    shadow_x1, shadow_y1 := options.GeoM.Apply(offset, offset)
                    shadow_x2, shadow_y2 := options.GeoM.Apply(offset + size, offset + size)

                    vector.DrawFilledRect(showMap, float32(shadow_x1), float32(shadow_y1), float32(shadow_x2 - shadow_x1), float32(shadow_y2 - shadow_y1), color.RGBA{A:255}, false)
                    vector.DrawFilledRect(showMap, float32(x1), float32(y1), float32(x2 - x1), float32(y2 - y1), bannerColor(city.GetBanner()), false)
                }
            }
        }

        // draw single pixel for unit stack
        for _, stack := range stacks {
            if stack.Plane() == plane {
                if useFog[stack.X()][stack.Y()] != data.FogTypeUnexplored {
                    options.GeoM.Reset()
                    options.GeoM.Translate(float64(stack.X()*tileImage0.Bounds().Dx()), float64(stack.Y()*tileImage0.Bounds().Dy()))
                    options.GeoM.Scale(scaleX, scaleY)

                    x1, y1 := options.GeoM.Apply(3, 3)

                    vector.DrawFilledRect(showMap, float32(x1), float32(y1), 1, 1, bannerColor(stack.GetBanner()), false)
                }
            }
        }


        return showMap
    }

    arcanusRender := renderMap(data.PlaneArcanus)
    myrrorRender := renderMap(data.PlaneMyrror)

    mouseX, mouseY := 0, 0
    var drawCityName *citylib.City

    offsetX := 25
    offsetY := 30

    scaleX := float64(arcanusRender.Bounds().Dx()) / float64(arcanusMap.Width() * tileImage0.Bounds().Dx())
    scaleY := float64(arcanusRender.Bounds().Dy()) / float64(arcanusMap.Height() * tileImage0.Bounds().Dy())

    ui := &uilib.UI{
        Draw: func (ui *uilib.UI, screen *ebiten.Image) {
            background, _ := imageCache.GetImage("reload.lbx", 2, 0)
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, background, &options)

            render := arcanusRender
            planeName := "Arcanus Plane"
            if currentPlane == data.PlaneMyrror {
                planeName = "Myrror Plane"
                render = myrrorRender
            }
            fonts.Title.PrintOptions(screen, float64(background.Bounds().Dx() / 2), 10, font.FontOptions{Scale: scale.ScaleAmount, Options: &options, Justify: font.FontJustifyCenter}, planeName)

            options.GeoM.Translate(float64(offsetX), float64(offsetY))
            scale.DrawScaled(screen, render, &options)

            if drawCityName != nil {
                cityName := drawCityName.Name

                options.GeoM.Reset()
                options.GeoM.Translate(float64(drawCityName.X*tileImage0.Bounds().Dx()), float64(drawCityName.Y*tileImage0.Bounds().Dy()))
                options.GeoM.Scale(scaleX, scaleY)

                // why is -60 so large? it seems like we should use -10 or so. A: this is because of the scale in the geom
                x1, y1 := options.GeoM.Apply(0, -65)

                fontUse, ok := fonts.BannerFonts[drawCityName.GetBanner()]

                if ok {
                    fontUse.PrintOptions(screen, x1 + float64(offsetX), y1 + float64(offsetY), font.FontOptions{Scale: scale.ScaleAmount, Options: &options, Justify: font.FontJustifyCenter, DropShadow: true}, cityName)
                }
            }

            bannerY := 80
            for _, banner := range bannerList {
                name := usedBanners[banner]
                if name != "" {
                    options.GeoM.Reset()
                    options.GeoM.Translate(260, float64(bannerY))
                    flag := getFlag(banner)
                    scale.DrawScaled(screen, flag, &options)
                    x, _ := options.GeoM.Apply(float64(flag.Bounds().Dx() + 2), 0)
                    fonts.Name.PrintOptions(screen, x, float64(bannerY), font.FontOptions{Scale: scale.ScaleAmount, Options: &options, Justify: font.FontJustifyLeft}, name)
                    bannerY += flag.Bounds().Dy() + 2
                }
            }

            ui.StandardDraw(screen)
        },
    }

    ui.SetElementsFromArray(nil)

    pageTurnRect := image.Rect(283, 173, 315, 198)
    ui.AddElement(&uilib.UIElement{
        Rect: pageTurnRect,
        Draw: func (element *uilib.UIElement, screen *ebiten.Image) {
            // util.DrawRect(screen, scale.ScaleRect(pageTurnRect), color.RGBA{R: 255, A: 255})
        },
        LeftClick: func (element *uilib.UIElement) {
            currentPlane = currentPlane.Opposite()
        },
        NotLeftClicked: func(element *uilib.UIElement){
            quit = true
        },
    })

    logic := func (yield coroutine.YieldFunc) error {
        var geom ebiten.GeoM
        geom.Scale(float64(tileImage0.Bounds().Dx()), float64(tileImage0.Bounds().Dy()))
        geom.Scale(scaleX, scaleY)
        geom.Invert()

        for !quit {
            counter += 1

            mouseX, mouseY = ebiten.CursorPosition()
            mouseX, mouseY = scale.Unscale2(mouseX, mouseY)

            usePlane := data.PlaneArcanus
            useFog := arcanusFog
            if currentPlane == data.PlaneMyrror {
                usePlane = data.PlaneMyrror
                useFog = myrrorFog
            }

            mx, my := geom.Apply(float64(mouseX - offsetX), float64(mouseY - offsetY))
            // log.Printf("converted mouse coordinates: %v %v", mx, my)

            drawCityName = nil
            // FIXME: use a kd-tree or some spatial datastructure for faster look ups
            maxDistance := 1.0
            for _, city := range cities {
                if city.Plane == usePlane && useFog[city.X][city.Y] != data.FogTypeUnexplored {
                    if math.Abs(mx - float64(city.X)) < maxDistance && math.Abs(my - float64(city.Y)) < maxDistance {
                        drawCityName = city
                        break
                    }
                }
            }

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
