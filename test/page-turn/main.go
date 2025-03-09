package main

import (
    "log"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"

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

func RightSideFlipLeftDistortions2(page *ebiten.Image) util.Distortion {
    offset := 30
    return util.Distortion{
        Top: image.Pt(page.Bounds().Dx()/2 - 130 + offset, 5),
        Bottom: image.Pt(page.Bounds().Dx()/2 - 130 + offset, page.Bounds().Dy() - 0),

        Segments: []util.Segment{
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 100 + offset, -0),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 100 + offset, page.Bounds().Dy() - 22),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 80 + offset, -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 80 + offset, page.Bounds().Dy() - 30),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 60 + offset, -10),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 60 + offset, page.Bounds().Dy() - 33),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 40 + offset, 0),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 40 + offset, page.Bounds().Dy() - 25),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 20 + offset, 5),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 20 + offset, page.Bounds().Dy() - 12),
            },
        },
    }
}

func RightSideFlipLeftDistortions1(page *ebiten.Image) util.Distortion {
    offset := 50
    return util.Distortion{
        Top: image.Pt(page.Bounds().Dx()/2 - 110 + offset, -10),
        Bottom: image.Pt(page.Bounds().Dx()/2 - 110 + offset, page.Bounds().Dy() - 12),
        Segments: []util.Segment{
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 90 + offset, -0),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 90 + offset, page.Bounds().Dy() - 22),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 73 + offset, -20),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 73 + offset, page.Bounds().Dy() - 35),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 58 + offset, -13),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 58 + offset, page.Bounds().Dy() - 35),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 40 + offset, 0),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 40 + offset, page.Bounds().Dy() - 28),
            },
            util.Segment{
                Top: image.Pt(page.Bounds().Dx()/2 - 20 + offset, 5),
                Bottom: image.Pt(page.Bounds().Dx()/2 - 20 + offset, page.Bounds().Dy() - 15),
            },
        },
    }
}

func (engine *Engine) DrawPage1(screen *ebiten.Image, page *ebiten.Image, options ebiten.DrawImageOptions){
    screen.DrawImage(page, &options)
    distortions := spellbook.LeftSideDistortions1(page)
    util.DrawDistortion(screen, page, engine.PageImage, distortions, options.GeoM)
}

func (engine *Engine) DrawPage2(screen *ebiten.Image, page *ebiten.Image, options ebiten.DrawImageOptions){
    screen.DrawImage(page, &options)
    distortions := spellbook.LeftSideDistortions2(page)
    util.DrawDistortion(screen, page, engine.PageImage, distortions, options.GeoM)
}

func (engine *Engine) DrawPage3(screen *ebiten.Image, page *ebiten.Image, options ebiten.DrawImageOptions){
    screen.DrawImage(page, &options)
    distortions := spellbook.RightSideDistortions1(page)
    util.DrawDistortion(screen, page, engine.PageImage, distortions, options.GeoM)
}

func (engine *Engine) DrawPage4(screen *ebiten.Image, page *ebiten.Image, options ebiten.DrawImageOptions){
    screen.DrawImage(page, &options)
    distortions := spellbook.RightSideDistortions2(page)
    util.DrawDistortion(screen, page, engine.PageImage, distortions, options.GeoM)
}

func (engine *Engine) Draw(screen *ebiten.Image){
    screen.Fill(color.RGBA{R: 0, G: 150, B: 200, A: 255})
    pages, _ := engine.ImageCache.GetImages("book.lbx", 1)
    var options ebiten.DrawImageOptions
    options.GeoM.Scale(1.2, 1.2)
    options.GeoM.Translate(10, 20)

    engine.DrawPage3(screen, pages[2], options)

    options.GeoM.Translate(float64(pages[0].Bounds().Dx() + 10), 0)
    engine.DrawPage4(screen, pages[3], options)

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

    pageIndex := (engine.Counter / 50) % uint64(len(pages) + 1)
    if pageIndex == 0 {
        engine.DrawPage1(screen, pages[pageIndex], options)
    } else if pageIndex == 1 {
        engine.DrawPage2(screen, pages[pageIndex], options)
    } else if pageIndex == 2 {
        engine.DrawPage3(screen, pages[pageIndex], options)
    } else if pageIndex == 3 {
        engine.DrawPage4(screen, pages[pageIndex], options)
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

    textFont.PrintWrap(spellPage, 0, 5, 135, 1, ebiten.ColorScale{}, font.FontOptions{}, "This is a test of the emergency broadcast system. This is only a test. If this were a real emergency, you would be instructed to do something else.")
    // textFont.PrintWrap(spellPage, 0, 50, 135, 1, ebiten.ColorScale{}, "A sub-image returned by SubImage can be used as a rendering source and a rendering destination. If a sub-image is used as a rendering source, the image is used as if it is a small image. If a sub-image is used as a rendering destination, the region being rendered is clipped.")
    textFont.PrintWrap(spellPage, 0, 50, 135, 1, ebiten.ColorScale{}, font.FontOptions{}, "aaaaa aaaaaa bbbbb bbbbb bbbb ccccc ccccc ccccc ccc dddd ddddd dddd dddd eeee eeee eeee")

    vector.DrawFilledCircle(spellPage, 10, 90, 5, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, true)
    vector.DrawFilledCircle(spellPage, 45, 90, 5, color.RGBA{R: 0x0, G: 0xff, B: 0, A: 0xff}, true)
    
    vector.StrokeLine(spellPage, 10, 110, 120, 110, 3, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, true)
    vector.StrokeLine(spellPage, 10, 130, 120, 130, 3, color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff}, true)

    leftPage := ebiten.NewImage(150, 170)
    leftPage.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 0})
    textFont.PrintWrap(leftPage, 5, 5, 135, 1, ebiten.ColorScale{}, font.FontOptions{}, "This is some text on the left page. Note that an important logic should not rely on values returned by RGBA64At, since the returned values can include very slight differences between some machines.")
    rightPage := ebiten.NewImage(150, 170)
    textFont.PrintWrap(rightPage, 5, 5, 135, 1, ebiten.ColorScale{}, font.FontOptions{}, "If the shader unit is texels, one of the specified image is non-nil and its size is different from (width, height), DrawTrianglesShader panics. If one of the specified image is non-nil and is disposed, DrawTrianglesShader panics.")

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
