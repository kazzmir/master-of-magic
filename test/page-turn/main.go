package main

import (
    "log"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 800
const ScreenHeight = 600

type Engine struct {
    Cache *lbx.LbxCache
    ImageCache util.ImageCache
    Images []*ebiten.Image
    Counter uint64

    PageImage *ebiten.Image
    LeftPage *ebiten.Image
    RightPage *ebiten.Image
}

func (engine *Engine) Update() error {
    engine.Counter += 1
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    return nil
}

func (engine *Engine) Page1Distortions(page *ebiten.Image) util.Distortion {
    return util.Distortion{
        Top: image.Pt(page.Bounds().Dx()/2 + 20, 5),
        Bottom: image.Pt(page.Bounds().Dx()/2 + 20, page.Bounds().Dy() - 12),
        Segments: []util.Segment{
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 40, 0),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 40, page.Bounds().Dy() - 25),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 60, -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 60, page.Bounds().Dy() - 33),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 80, -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 80, page.Bounds().Dy() - 30),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 100, -0),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 100, page.Bounds().Dy() - 22),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 130, -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 130, page.Bounds().Dy() - 12),
            },
        },
    }
}

func (engine *Engine) Page2Distortions(page *ebiten.Image) util.Distortion {

    return util.Distortion{
        Top: image.Pt(page.Bounds().Dx()/2 + 20, 5),
        Bottom: image.Pt(page.Bounds().Dx()/2 + 20, page.Bounds().Dy() - 15),
        Segments: []util.Segment{
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 40, 0),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 40, page.Bounds().Dy() - 28),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 58, -13),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 58, page.Bounds().Dy() - 35),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 73, -20),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 73, page.Bounds().Dy() - 35),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 90, -0),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 90, page.Bounds().Dy() - 22),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 + 120, -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 + 120, page.Bounds().Dy() - 12),
            },
        },
    }
}

func (engine *Engine) DrawPage1(screen *ebiten.Image, page *ebiten.Image, options ebiten.DrawImageOptions){
    screen.DrawImage(page, &options)
    distortions := engine.Page1Distortions(page)
    util.DrawDistortion(screen, page, engine.PageImage, distortions, options)
}

func (engine *Engine) DrawPage2(screen *ebiten.Image, page *ebiten.Image, options ebiten.DrawImageOptions){
    screen.DrawImage(page, &options)
    distortions := engine.Page2Distortions(page)
    util.DrawDistortion(screen, page, engine.PageImage, distortions, options)
}

func (engine *Engine) Draw(screen *ebiten.Image){
    screen.Fill(color.RGBA{R: 0, G: 150, B: 200, A: 255})
    pages, _ := engine.ImageCache.GetImages("book.lbx", 1)
    var options ebiten.DrawImageOptions
    options.GeoM.Scale(1.2, 1.2)
    options.GeoM.Translate(10, 20)

    engine.DrawPage1(screen, pages[0], options)

    options.GeoM.Translate(float64(pages[0].Bounds().Dx() + 10), 0)
    engine.DrawPage2(screen, pages[1], options)

    options.GeoM.Reset()
    options.GeoM.Scale(1.2, 1.2)
    background, _ := engine.ImageCache.GetImage("scroll.lbx", 6, 0)
    options.GeoM.Translate(10, 300)
    screen.DrawImage(background, &options)

    options.GeoM.Translate(200, 10)
    // screen.DrawImage(engine.Images[0], &options)
    screen.DrawImage(engine.PageImage, &options)


    options.GeoM.Translate(200, -10)
    screen.DrawImage(background, &options)

    options2 := options
    options2.GeoM.Translate(15, 10)
    screen.DrawImage(engine.LeftPage, &options2)

    options2.GeoM.Translate(175, 0)
    screen.DrawImage(engine.RightPage, &options2)

    pageIndex := (engine.Counter / 10) % uint64(len(pages) + 1)
    if pageIndex == 0 {
        engine.DrawPage1(screen, pages[pageIndex], options)
    } else if pageIndex == 1 {
        engine.DrawPage2(screen, pages[pageIndex], options)
    } else if pageIndex < uint64(len(pages)) {
        screen.DrawImage(pages[pageIndex], &options)
    }
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return ScreenWidth, ScreenHeight
}

func NewEngine() (*Engine, error){
    cache := lbx.AutoCache()

    image1 := ebiten.NewImage(30, 170)
    image1.Fill(color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff})
    vector.DrawFilledCircle(image1, 15, 15, 8, color.RGBA{R: 0, G: 0xff, B: 0, A: 0xff}, true)

    image2 := ebiten.NewImage(30, 170)
    image2.Fill(color.RGBA{R: 0, G: 0, B: 0xff, A: 0xff})
    image2.Fill(color.RGBA{R: 0, G: 0, B: 0x0, A: 0x0})
    vector.DrawFilledCircle(image2, 15, 45, 8, color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff}, true)

    images := []*ebiten.Image{image1, image2}

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        return nil, err
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        return nil, err
    }

    greyLight := util.PremultiplyAlpha(color.RGBA{R: 35, G: 35, B: 35, A: 164})
    textPaletteLighter := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        util.PremultiplyAlpha(color.RGBA{R: 35, G: 35, B: 35, A: 64}),
        greyLight, greyLight, greyLight,
        greyLight, greyLight, greyLight,
    }
    
    textFont := font.MakeOptimizedFontWithPalette(fonts[0], textPaletteLighter)

    spellPage := ebiten.NewImage(150, 170)
    spellPage.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 0})

    textFont.PrintWrap(spellPage, 0, 5, 135, 1, ebiten.ColorScale{}, "This is a test of the emergency broadcast system. This is only a test. If this were a real emergency, you would be instructed to do something else.")
    // textFont.PrintWrap(spellPage, 0, 50, 135, 1, ebiten.ColorScale{}, "A sub-image returned by SubImage can be used as a rendering source and a rendering destination. If a sub-image is used as a rendering source, the image is used as if it is a small image. If a sub-image is used as a rendering destination, the region being rendered is clipped.")
    textFont.PrintWrap(spellPage, 0, 50, 135, 1, ebiten.ColorScale{}, "aaaaa aaaaaa bbbbb bbbbb bbbb ccccc ccccc ccccc ccc dddd ddddd dddd dddd eeee eeee eeee")

    vector.DrawFilledCircle(spellPage, 10, 90, 5, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, true)
    vector.DrawFilledCircle(spellPage, 45, 90, 5, color.RGBA{R: 0x0, G: 0xff, B: 0, A: 0xff}, true)
    
    vector.StrokeLine(spellPage, 10, 110, 120, 110, 3, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, true)
    vector.StrokeLine(spellPage, 10, 130, 120, 130, 3, color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff}, true)

    leftPage := ebiten.NewImage(150, 170)
    leftPage.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 0})
    textFont.PrintWrap(leftPage, 5, 5, 135, 1, ebiten.ColorScale{}, "This is some text on the left page. Note that an important logic should not rely on values returned by RGBA64At, since the returned values can include very slight differences between some machines.")
    rightPage := ebiten.NewImage(150, 170)
    textFont.PrintWrap(rightPage, 5, 5, 135, 1, ebiten.ColorScale{}, "If the shader unit is texels, one of the specified image is non-nil and its size is different from (width, height), DrawTrianglesShader panics. If one of the specified image is non-nil and is disposed, DrawTrianglesShader panics.")

    return &Engine{
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        Images: images,
        PageImage: spellPage,
        LeftPage: leftPage,
        RightPage: rightPage,
    }, nil
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(int(float64(ScreenWidth) * 1.8), int(float64(ScreenHeight) * 1.8))
    ebiten.SetWindowTitle("page turn")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    engine, err := NewEngine()

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}

// legacy stuff
func (engine *Engine) DrawPage1_old(screen *ebiten.Image, page *ebiten.Image, options ebiten.DrawImageOptions){
    screen.DrawImage(page, &options)

        /*
        imageOptions := options
        imageOptions.GeoM.Translate(400, 40)
        imageOptions.GeoM.Skew(0.05, 0)
        */

        /*

    ax0, ay0 := options.GeoM.Apply(0, 0)
    ax1, ay1 := options.GeoM.Apply(float64(page.Bounds().Dx()), float64(page.Bounds().Dy()))
    subScreen := screen.SubImage(image.Rect(int(ax0), int(ay0), int(ax1), int(ay1))).(*ebiten.Image)

    use := engine.PageImage

    x1, y1 := options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 20, 5)
    x2, y2 := options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 40, 0)
    x3, y3 := options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 40, float64(page.Bounds().Dy()) - 25)
    x4, y4 := options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 20, float64(page.Bounds().Dy()) - 12)

    sx := float32(0)
    sy := float32(engine.PageImage.Bounds().Dy())

    drawDistortedImage(subScreen, use, []ebiten.Vertex{
        ebiten.Vertex{
            DstX: float32(x1),
            DstY: float32(y1),
            SrcX: sx + 30 * 0,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x2),
            DstY: float32(y2),
            SrcX: sx + 30 * 1,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x3),
            DstY: float32(y3),
            SrcX: sx + 30 * 1,
            SrcY: sy,
        },
        ebiten.Vertex{
            DstX: float32(x4),
            DstY: float32(y4),
            SrcX: sx + 30 * 0,
            SrcY: sy,
        },
    })

    x1 = x2
    y1 = y2
    x4 = x3
    y4 = y3
    x2, y2 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 60, -10)
    x3, y3 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 60, float64(page.Bounds().Dy()) - 33)

    // drawDistortedImage(subScreen, use, []ebiten.Vertex{ebiten.Vertex{DstX: float32(x1), DstY: float32(y1)}, ebiten.Vertex{DstX: float32(x2), DstY: float32(y2)}, ebiten.Vertex{DstX: float32(x3), DstY: float32(y3)}, ebiten.Vertex{DstX: float32(x4), DstY: float32(y4)}})

    drawDistortedImage(subScreen, use, []ebiten.Vertex{
        ebiten.Vertex{
            DstX: float32(x1),
            DstY: float32(y1),
            SrcX: sx + 30 * 1,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x2),
            DstY: float32(y2),
            SrcX: sx + 30 * 2,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x3),
            DstY: float32(y3),
            SrcX: sx + 30 * 2,
            SrcY: sy,
        },
        ebiten.Vertex{
            DstX: float32(x4),
            DstY: float32(y4),
            SrcX: sx + 30 * 1,
            SrcY: sy,
        },
    })

    x1 = x2
    y1 = y2
    x4 = x3
    y4 = y3
    x2, y2 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 80, -10)
    x3, y3 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 80, float64(page.Bounds().Dy()) - 30)

    drawDistortedImage(subScreen, use, []ebiten.Vertex{
        ebiten.Vertex{
            DstX: float32(x1),
            DstY: float32(y1),
            SrcX: sx + 30 * 2,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x2),
            DstY: float32(y2),
            SrcX: sx + 30 * 3,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x3),
            DstY: float32(y3),
            SrcX: sx + 30 * 3,
            SrcY: sy,
        },
        ebiten.Vertex{
            DstX: float32(x4),
            DstY: float32(y4),
            SrcX: sx + 30 * 2,
            SrcY: sy,
        },
    })

    x1 = x2
    y1 = y2
    x4 = x3
    y4 = y3
    x2, y2 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 100, -0)
    x3, y3 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 100, float64(page.Bounds().Dy()) - 22)

    drawDistortedImage(subScreen, use, []ebiten.Vertex{
        ebiten.Vertex{
            DstX: float32(x1),
            DstY: float32(y1),
            SrcX: sx + 30 * 3,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x2),
            DstY: float32(y2),
            SrcX: sx + 30 * 4,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x3),
            DstY: float32(y3),
            SrcX: sx + 30 * 4,
            SrcY: sy,
        },
        ebiten.Vertex{
            DstX: float32(x4),
            DstY: float32(y4),
            SrcX: sx + 30 * 3,
            SrcY: sy,
        },
    })

    x1 = x2
    y1 = y2
    x4 = x3
    y4 = y3
    x2, y2 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 130, -10)
    x3, y3 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 130, float64(page.Bounds().Dy()) - 12)

    drawDistortedImage(subScreen, use, []ebiten.Vertex{
        ebiten.Vertex{
            DstX: float32(x1),
            DstY: float32(y1),
            SrcX: sx + 30 * 4,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x2),
            DstY: float32(y2),
            SrcX: sx + 30 * 5,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x3),
            DstY: float32(y3),
            SrcX: sx + 30 * 5,
            SrcY: sy,
        },
        ebiten.Vertex{
            DstX: float32(x4),
            DstY: float32(y4),
            SrcX: sx + 30 * 4,
            SrcY: sy,
        },
    })
    */

    /*
    x1 = x2
    y1 = y2
    x4 = x3
    y4 = y3
    x2, y2 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 140, 10)
    x3, y3 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 140, float64(page.Bounds().Dy()) - 3)

    drawDistortedImage(subScreen, use, []ebiten.Vertex{ebiten.Vertex{DstX: float32(x1), DstY: float32(y1)}, ebiten.Vertex{DstX: float32(x2), DstY: float32(y2)}, ebiten.Vertex{DstX: float32(x3), DstY: float32(y3)}, ebiten.Vertex{DstX: float32(x4), DstY: float32(y4)}})
    */

}

func (engine *Engine) DrawPage2_old(screen *ebiten.Image, page *ebiten.Image, options ebiten.DrawImageOptions){
    screen.DrawImage(page, &options)

    /*
    ax0, ay0 := options.GeoM.Apply(0, 0)
    ax1, ay1 := options.GeoM.Apply(float64(page.Bounds().Dx()), float64(page.Bounds().Dy()))
    subScreen := screen.SubImage(image.Rect(int(ax0), int(ay0), int(ax1), int(ay1))).(*ebiten.Image)

    use := engine.PageImage

    x1, y1 := options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 20, 5)
    x2, y2 := options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 40, 0)
    x3, y3 := options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 40, float64(page.Bounds().Dy()) - 28)
    x4, y4 := options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 20, float64(page.Bounds().Dy()) - 15)

    sx := float32(0)
    sy := float32(engine.PageImage.Bounds().Dy())

    drawDistortedImage(subScreen, use, []ebiten.Vertex{
        ebiten.Vertex{
            DstX: float32(x1),
            DstY: float32(y1),
            SrcX: sx + 30 * 0,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x2),
            DstY: float32(y2),
            SrcX: sx + 30 * 1,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x3),
            DstY: float32(y3),
            SrcX: sx + 30 * 1,
            SrcY: sy,
        },
        ebiten.Vertex{
            DstX: float32(x4),
            DstY: float32(y4),
            SrcX: sx + 30 * 0,
            SrcY: sy,
        },
    })

    x1 = x2
    y1 = y2
    x4 = x3
    y4 = y3
    x2, y2 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 58, -13)
    x3, y3 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 58, float64(page.Bounds().Dy()) - 35)

    // drawDistortedImage(subScreen, use, []ebiten.Vertex{ebiten.Vertex{DstX: float32(x1), DstY: float32(y1)}, ebiten.Vertex{DstX: float32(x2), DstY: float32(y2)}, ebiten.Vertex{DstX: float32(x3), DstY: float32(y3)}, ebiten.Vertex{DstX: float32(x4), DstY: float32(y4)}})

    drawDistortedImage(subScreen, use, []ebiten.Vertex{
        ebiten.Vertex{
            DstX: float32(x1),
            DstY: float32(y1),
            SrcX: sx + 30 * 1,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x2),
            DstY: float32(y2),
            SrcX: sx + 30 * 2,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x3),
            DstY: float32(y3),
            SrcX: sx + 30 * 2,
            SrcY: sy,
        },
        ebiten.Vertex{
            DstX: float32(x4),
            DstY: float32(y4),
            SrcX: sx + 30 * 1,
            SrcY: sy,
        },
    })

    x1 = x2
    y1 = y2
    x4 = x3
    y4 = y3
    x2, y2 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 73, -20)
    x3, y3 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 73, float64(page.Bounds().Dy()) - 35)

    drawDistortedImage(subScreen, use, []ebiten.Vertex{
        ebiten.Vertex{
            DstX: float32(x1),
            DstY: float32(y1),
            SrcX: sx + 30 * 2,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x2),
            DstY: float32(y2),
            SrcX: sx + 30 * 3,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x3),
            DstY: float32(y3),
            SrcX: sx + 30 * 3,
            SrcY: sy,
        },
        ebiten.Vertex{
            DstX: float32(x4),
            DstY: float32(y4),
            SrcX: sx + 30 * 2,
            SrcY: sy,
        },
    })

    x1 = x2
    y1 = y2
    x4 = x3
    y4 = y3
    x2, y2 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 90, -0)
    x3, y3 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 90, float64(page.Bounds().Dy()) - 22)

    drawDistortedImage(subScreen, use, []ebiten.Vertex{
        ebiten.Vertex{
            DstX: float32(x1),
            DstY: float32(y1),
            SrcX: sx + 30 * 3,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x2),
            DstY: float32(y2),
            SrcX: sx + 30 * 4,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x3),
            DstY: float32(y3),
            SrcX: sx + 30 * 4,
            SrcY: sy,
        },
        ebiten.Vertex{
            DstX: float32(x4),
            DstY: float32(y4),
            SrcX: sx + 30 * 3,
            SrcY: sy,
        },
    })

    x1 = x2
    y1 = y2
    x4 = x3
    y4 = y3
    x2, y2 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 120, -10)
    x3, y3 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 120, float64(page.Bounds().Dy()) - 12)

    drawDistortedImage(subScreen, use, []ebiten.Vertex{
        ebiten.Vertex{
            DstX: float32(x1),
            DstY: float32(y1),
            SrcX: sx + 30 * 4,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x2),
            DstY: float32(y2),
            SrcX: sx + 30 * 5,
            SrcY: 0,
        },
        ebiten.Vertex{
            DstX: float32(x3),
            DstY: float32(y3),
            SrcX: sx + 30 * 5,
            SrcY: sy,
        },
        ebiten.Vertex{
            DstX: float32(x4),
            DstY: float32(y4),
            SrcX: sx + 30 * 4,
            SrcY: sy,
        },
    })
    */

}
