package main

import (
    "log"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type Engine struct {
    Cache *lbx.LbxCache
    ImageCache util.ImageCache
    Images []*ebiten.Image
}

func (engine *Engine) Update() error {
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    return nil
}

func drawDistortedImage(destination *ebiten.Image, source *ebiten.Image, vertices []ebiten.Vertex){

    vertices[0].SrcX = 0
    vertices[0].SrcY = 0
    vertices[1].SrcX = float32(source.Bounds().Dx())
    vertices[1].SrcY = 0
    vertices[2].SrcX = float32(source.Bounds().Dx())
    vertices[2].SrcY = float32(source.Bounds().Dy())
    vertices[3].SrcX = 0
    vertices[3].SrcY = float32(source.Bounds().Dy())

    for i := 0; i < 4; i++ {
        vertices[i].ColorA = 1
        vertices[i].ColorR = 1
        vertices[i].ColorG = 1
        vertices[i].ColorB = 1
    }

    destination.DrawTriangles(vertices, []uint16{0, 1, 2, 2, 3, 0}, source, nil)
}

func (engine *Engine) DrawPage1(screen *ebiten.Image, page *ebiten.Image, options ebiten.DrawImageOptions){
    screen.DrawImage(page, &options)

        /*
        imageOptions := options
        imageOptions.GeoM.Translate(400, 40)
        imageOptions.GeoM.Skew(0.05, 0)
        */

    use := engine.Images[0]

    ax0, ay0 := options.GeoM.Apply(0, 0)
    ax1, ay1 := options.GeoM.Apply(float64(page.Bounds().Dx()), float64(page.Bounds().Dy()))
    subScreen := screen.SubImage(image.Rect(int(ax0), int(ay0), int(ax1), int(ay1))).(*ebiten.Image)

    x1, y1 := options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 20, 5)
    x2, y2 := options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 40, 0)
    x3, y3 := options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 40, float64(page.Bounds().Dy()) - 25)
    x4, y4 := options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 20, float64(page.Bounds().Dy()) - 12)

    drawDistortedImage(subScreen, use, []ebiten.Vertex{ebiten.Vertex{DstX: float32(x1), DstY: float32(y1)}, ebiten.Vertex{DstX: float32(x2), DstY: float32(y2)}, ebiten.Vertex{DstX: float32(x3), DstY: float32(y3)}, ebiten.Vertex{DstX: float32(x4), DstY: float32(y4)}})

    x1 = x2
    y1 = y2
    x4 = x3
    y4 = y3
    x2, y2 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 60, -10)
    x3, y3 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 60, float64(page.Bounds().Dy()) - 33)

    drawDistortedImage(subScreen, use, []ebiten.Vertex{ebiten.Vertex{DstX: float32(x1), DstY: float32(y1)}, ebiten.Vertex{DstX: float32(x2), DstY: float32(y2)}, ebiten.Vertex{DstX: float32(x3), DstY: float32(y3)}, ebiten.Vertex{DstX: float32(x4), DstY: float32(y4)}})

    x1 = x2
    y1 = y2
    x4 = x3
    y4 = y3
    x2, y2 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 80, -20)
    x3, y3 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 80, float64(page.Bounds().Dy()) - 30)

    drawDistortedImage(subScreen, engine.Images[1], []ebiten.Vertex{ebiten.Vertex{DstX: float32(x1), DstY: float32(y1)}, ebiten.Vertex{DstX: float32(x2), DstY: float32(y2)}, ebiten.Vertex{DstX: float32(x3), DstY: float32(y3)}, ebiten.Vertex{DstX: float32(x4), DstY: float32(y4)}})

    x1 = x2
    y1 = y2
    x4 = x3
    y4 = y3
    x2, y2 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 100, -30)
    x3, y3 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 100, float64(page.Bounds().Dy()) - 22)

    drawDistortedImage(subScreen, engine.Images[1], []ebiten.Vertex{ebiten.Vertex{DstX: float32(x1), DstY: float32(y1)}, ebiten.Vertex{DstX: float32(x2), DstY: float32(y2)}, ebiten.Vertex{DstX: float32(x3), DstY: float32(y3)}, ebiten.Vertex{DstX: float32(x4), DstY: float32(y4)}})

    x1 = x2
    y1 = y2
    x4 = x3
    y4 = y3
    x2, y2 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 120, -20)
    x3, y3 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 120, float64(page.Bounds().Dy()) - 12)

    drawDistortedImage(subScreen, engine.Images[1], []ebiten.Vertex{ebiten.Vertex{DstX: float32(x1), DstY: float32(y1)}, ebiten.Vertex{DstX: float32(x2), DstY: float32(y2)}, ebiten.Vertex{DstX: float32(x3), DstY: float32(y3)}, ebiten.Vertex{DstX: float32(x4), DstY: float32(y4)}})

    x1 = x2
    y1 = y2
    x4 = x3
    y4 = y3
    x2, y2 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 140, 10)
    x3, y3 = options.GeoM.Apply(float64(page.Bounds().Dx())/2 + 140, float64(page.Bounds().Dy()) - 3)

    drawDistortedImage(subScreen, engine.Images[1], []ebiten.Vertex{ebiten.Vertex{DstX: float32(x1), DstY: float32(y1)}, ebiten.Vertex{DstX: float32(x2), DstY: float32(y2)}, ebiten.Vertex{DstX: float32(x3), DstY: float32(y3)}, ebiten.Vertex{DstX: float32(x4), DstY: float32(y4)}})

}

func (engine *Engine) DrawPage2(screen *ebiten.Image, page *ebiten.Image, options ebiten.DrawImageOptions){
    screen.DrawImage(page, &options)
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

    // for _, page := range pages[:2] {

        /*
        imageOptions := ebiten.DrawTrianglesOptions{}

        vertex := []ebiten.Vertex{
            ebiten.Vertex{
                SrcX: 0,
                SrcY: 0,
                DstX: 7,
                DstY: 3,
                ColorA: 1,
                ColorR: 1,
                ColorG: 1,
                ColorB: 1,
            },
            ebiten.Vertex{
                SrcX: float32(use.Bounds().Dx()),
                SrcY: 0,
                DstX: 47,
                DstY: 8,
                ColorA: 1,
                ColorR: 1,
                ColorG: 1,
                ColorB: 1,
            },
            ebiten.Vertex{
                SrcX: float32(use.Bounds().Dx()-1),
                SrcY: float32(use.Bounds().Dy()-1),
                DstX: 80,
                DstY: 96,
                ColorA: 1,
                ColorR: 1,
                ColorG: 1,
                ColorB: 1,
            },
            ebiten.Vertex{
                SrcX: 0,
                SrcY: float32(use.Bounds().Dy()-1),
                DstX: 5,
                DstY: 93,
                ColorA: 1,
                ColorR: 1,
                ColorG: 1,
                ColorB: 1,
            },
        }

        screen.DrawTriangles(vertex, []uint16{0, 1, 2, 2, 3, 0}, engine.Images[0], &imageOptions)
        */

        /*
        options.GeoM.Translate(float64(page.Bounds().Dx() + 10), 0)
    }
    */

    options.GeoM.Reset()
    options.GeoM.Scale(1.2, 1.2)
    background, _ := engine.ImageCache.GetImage("scroll.lbx", 6, 0)
    options.GeoM.Translate(10, 300)
    screen.DrawImage(background, &options)

    options.GeoM.Translate(200, 10)
    screen.DrawImage(engine.Images[0], &options)
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

    return &Engine{
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        Images: images,
    }, nil
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("page turn")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    engine, err := NewEngine()

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
