package combat

import (
    "image"
    "image/color"
    "log"
    "fmt"

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

type CombatEndScreenResult int
const (
    CombatEndScreenResultWin CombatEndScreenResult = iota
    CombatEndScreenResultLoose
    CombatEndScreenResultRetreat
)

type CombatEndScreen struct {
    CombatScreen *CombatScreen
    Result CombatEndScreenResult
    UnitsLost int
    Fame int
    Cache *lbx.LbxCache
    ImageCache util.ImageCache
    UI *uilib.UI
    State CombatEndScreenState
}

func MakeCombatEndScreen(cache *lbx.LbxCache, combat *CombatScreen, result CombatEndScreenResult, unitsLost int, fame int) *CombatEndScreen {
    end := &CombatEndScreen{
        CombatScreen: combat,
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        Result: result,
        UnitsLost: unitsLost,
        Fame: fame,
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

    extraText := ""
    switch {
        case end.Result == CombatEndScreenResultWin && end.Fame > 0:
            extraText = fmt.Sprintf("You gained %v fame", end.Fame)
        case end.Result == CombatEndScreenResultLoose && end.Fame > 0:
            extraText = fmt.Sprintf("You lost %v fame", end.Fame)
        case end.Result == CombatEndScreenResultRetreat && end.UnitsLost == 1 && end.Fame == 0:
            extraText = "You lost 1 unit while fleeing"
        case end.Result == CombatEndScreenResultRetreat && end.UnitsLost > 1 && end.Fame == 0:
            extraText = fmt.Sprintf("You lost %v units while fleeing.", end.UnitsLost)
        case end.Result == CombatEndScreenResultRetreat && end.UnitsLost == 1 && end.Fame > 0:
            extraText = fmt.Sprintf("You lost %v fame and 1 unit while fleeing.", end.Fame)
        case end.Result == CombatEndScreenResultRetreat && end.UnitsLost > 1 && end.Fame > 0:
            extraText = fmt.Sprintf("You lost %v fame and %v units while fleeing", end.Fame, end.UnitsLost)
    }

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
            switch end.Result {
                case CombatEndScreenResultWin:
                    pic, _ = end.ImageCache.GetImage("scroll.lbx", 10, 0)
                case CombatEndScreenResultLoose, CombatEndScreenResultRetreat:
                    pic, _ = end.ImageCache.GetImage("scroll.lbx", 11, 0)
            }

            bottom, _ := end.ImageCache.GetImage("help.lbx", 1, 0)

            picLength := 90

            fontY := picLength

            picLength += extraFont.Height()

            subPic := pic.SubImage(image.Rect(0, 0, pic.Bounds().Dx(), picLength * data.ScreenScale)).(*ebiten.Image)

            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(50 * data.ScreenScale), float64(30 * data.ScreenScale))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(subPic, &options)

            titleX, titleY := options.GeoM.Apply(float64(110 * data.ScreenScale), float64(25 * data.ScreenScale))
            switch end.Result {
                case CombatEndScreenResultWin:
                    titleFont.PrintCenter(screen, titleX, titleY, float64(data.ScreenScale), options.ColorScale, "You are triumphant")
                case CombatEndScreenResultLoose:
                    titleFont.PrintCenter(screen, titleX, titleY, float64(data.ScreenScale), options.ColorScale, "You have been defeated")
                case CombatEndScreenResultRetreat:
                    titleFont.PrintCenter(screen, titleX, titleY, float64(data.ScreenScale), options.ColorScale, "Your forces have retreated")
            }

            extraX, extraY := options.GeoM.Apply(float64(110 * data.ScreenScale), float64(fontY * data.ScreenScale))

            options.GeoM.Translate(0, float64(picLength * data.ScreenScale))
            screen.DrawImage(bottom, &options)

            extraFont.PrintCenter(screen, extraX, extraY, float64(data.ScreenScale), options.ColorScale, extraText)
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
