package cartographer

import (
    "log"
    "image"
    "image/color"
    "math"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    // "github.com/kazzmir/master-of-magic/lib/functional"
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
        Title: loader(fontslib.BigBlack),
        BannerFonts: cityViewFonts.BannerFonts,
    }
}

func MakeCartographer(cache *lbx.LbxCache, cities []*citylib.City, stacks []*playerlib.UnitStack, arcanusMap *maplib.Map, arcanusFog data.FogMap, myrrorMap *maplib.Map, myrrorFog data.FogMap) (coroutine.AcceptYieldFunc, func (*ebiten.Image)) {
    quit := false

    fonts := makeFonts(cache)

    imageCache := util.MakeImageCache(cache)

    counter := uint64(0)

    getAlpha := util.MakeFadeIn(7, &counter)

    currentPlane := data.PlaneArcanus

    /*
    getTileColor := functional.Memoize(func (kind terrain.TerrainType) color.RGBA {
        switch kind {
            case terrain.Ocean, terrain.River: return color.RGBA{R: 88, G: 68, B: 54, A: 255}
            case terrain.Mountain: return color.RGBA{R: 173, G: 138, B: 114, A: 255}
            case terrain.Desert: return color.RGBA{R: 172, G: 133, B: 107, A: 255}
            case terrain.SorceryNode: return color.RGBA{R: 170, G: 146, B: 129, A: 255}
            / *
            case terrain.Shore:
            case terrain.Hill
            case terrain.Grass
            case terrain.Swamp
            case terrain.Forest
            case terrain.Tundra
            case terrain.Volcano
            case terrain.Lake
            case terrain.NatureNode
            case terrain.ChaosNode
            * /
        }

        return color.RGBA{R: 47, G: 30, B: 12, A: 255}
    })
    */

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
        matrix.Scale(0.7, 0.5, 0.3, 1)
        for x := range useFog {
            for y := range useFog[x] {
                if /*x%2 == 0 && y%2 == 0 &&*/ useFog[x][y] != data.FogTypeUnexplored {
                    // tile := useMap.GetTile(x, y)
                    // tileColor := getTileColor(tile.Tile.TerrainType())
                    tileImage, err := useMap.GetTileImage(x, y, 0)
                    if err == nil {
                        options.GeoM.Reset()
                        options.GeoM.Translate(float64(x*tileImage.Bounds().Dx()), float64(y*tileImage.Bounds().Dy()))
                        options.GeoM.Scale(scaleX, scaleY)
                        colorm.DrawImage(showMap, tileImage, matrix, &options)

                        /*
                        var options2 ebiten.DrawImageOptions
                        options2.GeoM = options.GeoM
                        showMap.DrawImage(tileImage, &options2)
                        */

                        /*
                        x1, y1 := options.GeoM.Apply(1, 1)
                        x2, y2 := options.GeoM.Apply(float64(tileImage.Bounds().Dx()-1), float64(tileImage.Bounds().Dy()-1))
                        vector.StrokeRect(showMap, float32(x1), float32(y1), float32(x2-x1), float32(y2-y1), 1, color.RGBA{R: 255, G: 0, B: 0, A: 255}, false)
                        */
                    }
                    // vector.DrawFilledRect(showMap, float32(x*2), float32(y*2), 2, 2, tileColor, false)
                }
            }
        }

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

                // why is -60 so large? it seems like we should use -10 or so
                x1, y1 := options.GeoM.Apply(0, -65)

                fontUse, ok := fonts.BannerFonts[drawCityName.GetBanner()]

                if ok {
                    fontUse.PrintOptions(screen, x1 + float64(offsetX), y1 + float64(offsetY), font.FontOptions{Scale: scale.ScaleAmount, Options: &options, Justify: font.FontJustifyCenter, DropShadow: true}, cityName)
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
            util.DrawRect(screen, scale.ScaleRect(pageTurnRect), color.RGBA{R: 255, A: 255})
        },
        LeftClick: func (element *uilib.UIElement) {
            currentPlane = currentPlane.Opposite()
        },
        NotLeftClicked: func(element *uilib.UIElement){
            quit = true
        },
    })

    /*
    abs := func (a int) int {
        if a < 0 {
            return -a
        }

        return a
    }
    */

    logic := func (yield coroutine.YieldFunc) error {
        for !quit {
            counter += 1

            mouseX, mouseY = ebiten.CursorPosition()
            mouseX, mouseY = scale.Unscale2(mouseX, mouseY)

            usePlane := data.PlaneArcanus
            if currentPlane == data.PlaneMyrror {
                usePlane = data.PlaneMyrror
            }

            var geom ebiten.GeoM
            geom.Scale(float64(tileImage0.Bounds().Dx()), float64(tileImage0.Bounds().Dy()))
            geom.Scale(scaleX, scaleY)
            geom.Invert()

            mx, my := geom.Apply(float64(mouseX - offsetX), float64(mouseY - offsetY))
            // log.Printf("converted mouse coordinates: %v %v", mx, my)

            drawCityName = nil
            // FIXME: use a kd-tree or some spatial datastructure for faster look ups
            maxDistance := 1.0
            for _, city := range cities {
                if city.Plane == usePlane {
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
