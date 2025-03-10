package main

import (
    "fmt"
    "image"
    "image/color"
    "log"
    "math"
    "os"
    "strconv"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/shaders"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type DrawFunction func (engine *Engine, screen *ebiten.Image)

type Engine struct {
    Counter uint64
    Cache *lbx.LbxCache
    ImageCache *util.ImageCache
    Drawer DrawFunction
}

func NewEngine(scenario int) (*Engine, error) {
    var drawer DrawFunction
    switch scenario {
        case 2: drawer = DrawScenario2
        case 3: drawer = DrawScenario3
        default: drawer = DrawScenario1
    }

    cache := lbx.AutoCache()
    imageCache := util.MakeImageCache(cache)
    engine := &Engine{
        Counter: 0,
        Cache: cache,
        ImageCache: &imageCache,
        Drawer: drawer,
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

func (engine *Engine) Draw(screen *ebiten.Image) {
    engine.Drawer(engine, screen)
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

func DrawScenario1(engine *Engine, screen *ebiten.Image){
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
    // x, y := regularOptions.GeoM.Apply(0, 0)
    util.DrawOutline(screen, engine.ImageCache, lizardUnit, regularOptions.GeoM, ebiten.ColorScale{}, engine.Counter/10, color.RGBA{R: 180, G: 0, B: 0, A: 255})

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

func DrawScenario2(engine *Engine, screen *ebiten.Image){
    screen.Fill(color.RGBA{R: 60, G: 60, B: 120, A: 255})

    lbxFile, err := engine.ImageCache.LbxCache.GetLbxFile("terrain.lbx")
    if err != nil {
        fmt.Printf("could not read terrain.lbx")
        return
    }

    data, err := terrain.ReadTerrainData(lbxFile)
    if err != nil {
        fmt.Printf("could not read terrain data")
        return
    }

    mask, _ := engine.ImageCache.GetImage("mapback.lbx", 93, 0)

    // FIXME: Test with animation
    // FIXME: Test with sparkles
    // FIXME: Test with unit
    // FIXME: Test with other node type
    natureNode := ebiten.NewImageFromImage(data.Tiles[terrain.IndexNatNode].Images[0])

    shader, err := engine.ImageCache.GetShader(shaders.ShaderWarp)
    if err != nil {
        log.Printf("Unable to get shader: %v", err)
        return
    }

    var options ebiten.DrawRectShaderOptions
    var regularOptions ebiten.DrawImageOptions
    options.GeoM.Translate(20, 20)
    options.Images[0] = natureNode
    options.Images[1] = mask
    options.Uniforms = make(map[string]interface{})
    options.Uniforms["Time"] = float32(math.Abs(float64(engine.Counter/10)))
    regularOptions.GeoM = options.GeoM
    screen.DrawImage(natureNode, &regularOptions)
    screen.DrawRectShader(natureNode.Bounds().Dx(), natureNode.Bounds().Dy(), shader, &options)
}

func DrawScenario3(engine *Engine, screen *ebiten.Image){
    screen.Fill(color.RGBA{R: 60, G: 60, B: 120, A: 255})

    // data.ScreenScale = 3

    newCache := util.MakeImageCache(engine.Cache)
    engine.ImageCache = &newCache

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

    // scale := (float64((engine.Counter / 5) % 10) - 5) / 5.0
    scale := math.Sin(float64(engine.Counter % 10000) / 20)

    regularOptions.GeoM.Scale(1.5 + scale, 1.5 + scale)
    regularOptions.GeoM.Translate(20, 100)
    screen.DrawImage(lizardUnit, &regularOptions)
    // x, y := regularOptions.GeoM.Apply(0, 0)
    util.DrawOutline(screen, engine.ImageCache, lizardUnit, regularOptions.GeoM, ebiten.ColorScale{}, engine.Counter/10, color.RGBA{R: 180, G: 0, B: 0, A: 255})
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    scale.UpdateScale(1)
    ebiten.SetWindowSize(data.ScreenWidth * 4, data.ScreenHeight * 4)
    ebiten.SetWindowTitle("shader test")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    scenario := 1

    if len(os.Args) >= 2 {
        x, err := strconv.Atoi(os.Args[1])
        if err != nil {
            log.Fatalf("Error with scenario: %v", err)
        }

        scenario = x
    }

    engine, err := NewEngine(scenario)

    if err != nil {
        log.Printf("Error: unable to load engine: %v", err)
        return
    }

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
