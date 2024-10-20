package combat

import (
    "image"
    "image/color"
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/hajimehoshi/ebiten/v2"
)

type CombatEndScreenState int
const (
    CombatEndScreenRunning CombatEndScreenState = iota
    CombatEndScreenDone
)

type CombatEndScreen struct {
    CombatScreen *CombatScreen
    Win bool
    Cache *lbx.LbxCache
    ImageCache util.ImageCache
    UI *uilib.UI
    State CombatEndScreenState
}

func MakeCombatEndScreen(cache *lbx.LbxCache, combat *CombatScreen, win bool) *CombatEndScreen {
    end := &CombatEndScreen{
        CombatScreen: combat,
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        Win: win,
        State: CombatEndScreenRunning,
    }

    end.UI = end.MakeUI()
    return end
}

func (end *CombatEndScreen) MakeUI() *uilib.UI {
    const fadeSpeed = 7

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
    }

    getAlpha := ui.MakeFadeIn(fadeSpeed)

    fontLbx, err := end.Cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

    titleRed := color.RGBA{R: 0x50, G: 0x00, B: 0x0e, A: 0xff}
    titlePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        titleRed,
        titleRed,
        titleRed,
        titleRed,
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    titleFont := font.MakeOptimizedFontWithPalette(fonts[4], titlePalette)

    extraText := "You have gained 1 fame"

    black := color.RGBA{R: 0, G: 0, B: 0, A: 0xff}
    extraPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        black, black, black,
        black, black, black,
    }

    extraFont := font.MakeOptimizedFontWithPalette(fonts[1], extraPalette)

    element := &uilib.UIElement{
        Rect: image.Rect(0, 0, data.ScreenWidth, data.ScreenHeight),
        LeftClick: func(element *uilib.UIElement){
            getAlpha = ui.MakeFadeOut(fadeSpeed)
            ui.AddDelay(fadeSpeed, func(){
                end.State = CombatEndScreenDone
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var pic *ebiten.Image
            if end.Win {
                pic, _ = end.ImageCache.GetImage("scroll.lbx", 10, 0)
            } else {
                pic, _ = end.ImageCache.GetImage("scroll.lbx", 11, 0)
            }

            bottom, _ := end.ImageCache.GetImage("help.lbx", 1, 0)

            picLength := 90

            fontY := picLength

            picLength += extraFont.Height()

            subPic := pic.SubImage(image.Rect(0, 0, pic.Bounds().Dx(), picLength)).(*ebiten.Image)

            var options ebiten.DrawImageOptions
            options.GeoM.Translate(50, 30)
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(subPic, &options)

            titleX, titleY := options.GeoM.Apply(110, 25)
            if end.Win {
                titleFont.PrintCenter(screen, titleX, titleY, 1, options.ColorScale, "You are triumphant")
            } else {
                titleFont.PrintCenter(screen, titleX, titleY, 1, options.ColorScale, "You have been defeated")
            }

            extraX, extraY := options.GeoM.Apply(110, float64(fontY))

            options.GeoM.Translate(0, float64(picLength))
            screen.DrawImage(bottom, &options)

            extraFont.PrintCenter(screen, extraX, extraY, 1, options.ColorScale, extraText)
        },
    }

    ui.SetElementsFromArray([]*uilib.UIElement{element})

    return ui
}

func (end *CombatEndScreen) Update() CombatEndScreenState {
    end.CombatScreen.MouseState = CombatClickHud
    end.UI.StandardUpdate()
    return end.State
}

func (end *CombatEndScreen) Draw(screen *ebiten.Image) {
    end.CombatScreen.Draw(screen)
    end.UI.Draw(end.UI, screen)
}
