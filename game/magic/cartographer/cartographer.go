package cartographer

import (
    "log"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    // "github.com/kazzmir/master-of-magic/lib/functional"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
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

    renderMap := func (plane data.Plane) *ebiten.Image {
        showMap := ebiten.NewImage(240, 140)
        // showMap.Fill(color.RGBA{R: 32, G: 32, B: 32, A: 255})

        useMap := arcanusMap
        useFog := arcanusFog
        if plane == data.PlaneMyrror {
            useMap = myrrorMap
            useFog = myrrorFog
        }

        var options colorm.DrawImageOptions
        var matrix colorm.ColorM
        // matrix.ScaleWithColor(color.RGBA{R: 217, G: 112, B: 61, A: 255})
        matrix.ChangeHSV(0, 0, 1.5)
        // matrix.Translate(155/255.0, 80/255.0, 44/255.0, 0)
        matrix.Scale(0.7, 0.5, 0.3, 1)
        for x := range useFog {
            for y := range useFog[x] {
                if useFog[x][y] != data.FogTypeUnexplored {
                    // tile := useMap.GetTile(x, y)
                    // tileColor := getTileColor(tile.Tile.TerrainType())
                    tileImage, err := useMap.GetTileImage(x, y, 0)
                    if err == nil {
                        options.GeoM.Reset()
                        options.GeoM.Translate(float64(x*tileImage.Bounds().Dx()), float64(y*tileImage.Bounds().Dy()))
                        options.GeoM.Scale(0.2, 0.2)
                        colorm.DrawImage(showMap, tileImage, matrix, &options)
                    }
                    // vector.DrawFilledRect(showMap, float32(x*2), float32(y*2), 2, 2, tileColor, false)
                }
            }
        }

        for _, city := range cities {
            if city.Plane == plane {
                if useFog[city.X][city.Y] != data.FogTypeUnexplored {
                    vector.DrawFilledRect(showMap, float32(city.X*2), float32(city.Y*2), 2, 2, bannerColor(city.GetBanner()), false)
                }
            }
        }


        return showMap
    }

    arcanusRender := renderMap(data.PlaneArcanus)
    myrrorRender := renderMap(data.PlaneMyrror)

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

            options.GeoM.Translate(15, 40)
            scale.DrawScaled(screen, render, &options)

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
