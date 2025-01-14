package main

import (
    "log"
    "image"
    "image/color"
    "math"

    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/shaders"
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 320
const ScreenHeight = 200

type Engine struct {
    Counter uint64
    Cache *lbx.LbxCache
    ImageCache *util.ImageCache
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()
    imageCache := util.MakeImageCache(cache)
    engine := &Engine{
        Counter: 0,
        Cache: cache,
        ImageCache: &imageCache,
    }

    return engine, nil
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

func toFloatArray(color color.Color) []float32 {
    r, g, b, a := color.RGBA()
    var max float32 = 65535.0
    return []float32{float32(r) / max, float32(g) / max, float32(b) / max, float32(a) / max}
}

// enlarge the image by 1 pixel on all sides
func add1PxBorder(src *image.Paletted) image.Image {
    out := image.NewPaletted(image.Rect(0, 0, src.Bounds().Dx()+2, src.Bounds().Dy()+2), src.Palette)

    for y := 0; y < src.Bounds().Dy(); y++ {
        for x := 0; x < src.Bounds().Dx(); x++ {
            out.SetColorIndex(x+1, y+1, src.ColorIndexAt(x, y))
        }
    }

    return out
}

func (engine *Engine) Draw(screen *ebiten.Image){
    screen.Fill(color.RGBA{R: 60, G: 60, B: 120, A: 255})

    lizardUnit, _ := engine.ImageCache.GetImage("units2.lbx", 0, 0)

    shader, err := engine.ImageCache.GetShader(shaders.ShaderEdgeGlow)
    if err != nil {
        log.Printf("Unable to get shader: %v", err)
        return
    }

    var options ebiten.DrawRectShaderOptions
    var regularOptions ebiten.DrawImageOptions
    options.GeoM.Translate(20, 20)
    options.Images[0] = lizardUnit
    options.Uniforms = make(map[string]interface{})
    options.Uniforms["Color1"] = toFloatArray(color.RGBA{R: 200, G: 0, B: 0, A: 255})
    options.Uniforms["Color2"] = toFloatArray(color.RGBA{R: 255, G: 0, B: 0, A: 255})
    options.Uniforms["Color3"] = toFloatArray(color.RGBA{R: 220, G: 40, B: 40, A: 255})
    options.Uniforms["Time"] = float32(math.Abs(float64(engine.Counter/10)))
    regularOptions.GeoM = options.GeoM
    screen.DrawImage(lizardUnit, &regularOptions)
    screen.DrawRectShader(lizardUnit.Bounds().Dx(), lizardUnit.Bounds().Dy(), shader, &options)

    lizardFigure, _ := engine.ImageCache.GetImage("figures9.lbx", 1, 0)
    options.GeoM.Translate(50, 0)
    options.Images[0] = lizardFigure
    regularOptions.GeoM = options.GeoM
    screen.DrawImage(lizardFigure, &regularOptions)
    screen.DrawRectShader(lizardFigure.Bounds().Dx(), lizardFigure.Bounds().Dy(), shader, &options)

    regularOptions.GeoM.Reset()
    regularOptions.GeoM.Translate(20, 60)
    screen.DrawImage(lizardUnit, &regularOptions)
    x, y := regularOptions.GeoM.Apply(0, 0)
    util.DrawOutline(screen, engine.ImageCache, lizardUnit, x, y, ebiten.ColorScale{}, engine.Counter/10, color.RGBA{R: 180, G: 0, B: 0, A: 255})

    axe, _ := engine.ImageCache.GetImageTransform("items.lbx", 23, 0, "1-px", util.ImageTransformFunc(add1PxBorder))
    regularOptions.GeoM.Reset()
    regularOptions.GeoM.Translate(20, 100)
    screen.DrawImage(axe, &regularOptions)

    options.GeoM.Reset()
    options.GeoM.Translate(50, 100)
    options.Images[0] = axe
    regularOptions.GeoM = options.GeoM
    screen.DrawImage(axe, &regularOptions)
    screen.DrawRectShader(axe.Bounds().Dx(), axe.Bounds().Dy(), shader, &options)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return ScreenWidth, ScreenHeight
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(ScreenWidth * 4, ScreenHeight * 4)
    ebiten.SetWindowTitle("glow test")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    engine, err := NewEngine()

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
