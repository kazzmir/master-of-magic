package main

import (
    "log"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/fonts"
    "github.com/kazzmir/master-of-magic/game/magic/shaders"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Engine struct {
    LbxCache *lbx.LbxCache
    ImageCache util.ImageCache
    VaultFonts *fonts.VaultFonts
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()

    vaultFonts := fonts.MakeVaultFonts(cache)

    return &Engine{
        LbxCache: cache,
        ImageCache: util.MakeImageCache(cache),
        VaultFonts: vaultFonts,
    }, nil
}

func (engine *Engine) Update() error {

    keys := inpututil.AppendJustPressedKeys(nil)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{R: 80, G: 80, B: 80, A: 255})

    shader, err := engine.ImageCache.GetShader(shaders.ShaderEdgeGlow)
    if err != nil {
        log.Printf("Error getting shader: %v", err)
        return
    }

    // engine.VaultFonts.ItemName.PrintOutline(screen, shader, 10, 10, 4, ebiten.ColorScale{}, "This is a test of font outlines")
    y := float64(10)
    engine.VaultFonts.ItemName.PrintDropShadow(screen, 10, y, 4, ebiten.ColorScale{}, "This is a test of font outlines")
    y += 60
    engine.VaultFonts.ItemName.Print(screen, 10, y, 4, ebiten.ColorScale{}, "This is a test of font outlines")
    y += 60
    engine.VaultFonts.ItemName.PrintOptions2(screen, 10, y, font.FontOptions{DropShadow: true, ShadowColor: color.RGBA{R: 255, G: 0, B: 0, A: 255}, Scale: 4}, "This is a test of font outlines")
    y += 60
    engine.VaultFonts.ItemName.PrintOptions2(screen, 10, y, font.FontOptions{DropShadow: true, Scale: 4}, "This is a test of font outlines")

    y += 60

    engine.VaultFonts.PowerFont.PrintOutline(screen, shader, 10, y, 4, ebiten.ColorScale{}, "This is a test of font outlines")
    y += 50
    engine.VaultFonts.PowerFont.Print(screen, 10, y, 4, ebiten.ColorScale{}, "This is a test of font outlines")

    y += 50

    engine.VaultFonts.ResourceFont.PrintOutline(screen, shader, 10, y, 4, ebiten.ColorScale{}, "This is a test of font outlines")
    y += 50
    engine.VaultFonts.ResourceFont.Print(screen, 10, y, 4, ebiten.ColorScale{}, "This is a test of font outlines")
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(data.ScreenWidth * 4, data.ScreenHeight * 4)
    ebiten.SetWindowTitle("intro")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    engine, err := NewEngine()

    if err != nil {
        log.Printf("Error: unable to load engine: %v", err)
        return
    }

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
