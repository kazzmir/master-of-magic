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
    "github.com/kazzmir/master-of-magic/game/magic/scale"
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
    CombatEndScreenResultLose
    CombatEndScreenResultRetreat
)

type CombatEndScreen struct {
    Result CombatEndScreenResult
    UnitsLost int
    Fame int
    // when fighting in a city, citizens and buildings may be lost
    PopulationLost int
    BuildingsLost int
    Cache *lbx.LbxCache
    ImageCache util.ImageCache
    UI *uilib.UI
    State CombatEndScreenState
}

func MakeCombatEndScreen(cache *lbx.LbxCache, result CombatEndScreenResult, unitsLost int, fame int, populationLost int, buildingsLost int) *CombatEndScreen {
    end := &CombatEndScreen{
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        Result: result,
        UnitsLost: unitsLost,
        Fame: fame,
        PopulationLost: populationLost,
        BuildingsLost: buildingsLost,
        State: CombatEndScreenRunning,
    }

    end.UI = end.MakeUI()
    return end
}

func (end *CombatEndScreen) MakeUI() *uilib.UI {
    const fadeSpeed = 7

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            ui.StandardDraw(screen)
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

    // for population/buildings lost
    extraText2 := ""
    if end.PopulationLost > 0 || end.BuildingsLost > 0 {
        switch {
            case end.PopulationLost > 0 && end.BuildingsLost > 0:
                extraText2 = fmt.Sprintf("%v citizens were lost and %v buildings were destroyed.", end.PopulationLost, end.BuildingsLost)
            case end.PopulationLost > 0:
                extraText2 = fmt.Sprintf("%v citizens were lost.", end.PopulationLost)
            case end.BuildingsLost > 0:
                extraText2 = fmt.Sprintf("%v buildings were destroyed.", end.BuildingsLost)
        }
    }

    extraText := ""
    switch {
        case end.Result == CombatEndScreenResultWin:
            if end.Fame > 0 {
                extraText = fmt.Sprintf("You gained %v fame", end.Fame)
            } else if end.Fame < 0 {
                extraText = fmt.Sprintf("You lost %v fame", -end.Fame)
            }
        case end.Result == CombatEndScreenResultLose && end.Fame > 0:
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

    var pic *ebiten.Image
    switch end.Result {
        case CombatEndScreenResultWin:
            pic, _ = end.ImageCache.GetImage("scroll.lbx", 10, 0)
        case CombatEndScreenResultLose, CombatEndScreenResultRetreat:
            pic, _ = end.ImageCache.GetImage("scroll.lbx", 11, 0)
    }

    extraText2Render := extraFont.CreateWrappedText(float64(pic.Bounds().Dx() - 2), 1, extraText2)

    element := &uilib.UIElement{
        Rect: image.Rect(0, 0, data.ScreenWidth, data.ScreenHeight),
        LeftClick: func(element *uilib.UIElement){
            getAlpha = ui.MakeFadeOut(fadeSpeed)
            ui.AddDelay(fadeSpeed, func(){
                end.State = CombatEndScreenDone
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            bottom, _ := end.ImageCache.GetImage("help.lbx", 1, 0)

            picLength := 90
            fontY := picLength

            if extraText2 != "" {
                picLength += extraFont.Height()
            }

            picLength += extraFont.Height()

            subPic := pic.SubImage(image.Rect(0, 0, pic.Bounds().Dx(), picLength)).(*ebiten.Image)

            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(50), float64(30))
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, subPic, &options)

            fontOptions := font.FontOptions{Justify: font.FontJustifyCenter, Options: &options, Scale: scale.ScaleAmount}

            titleX, titleY := options.GeoM.Apply(float64(110), float64(25))
            switch end.Result {
                case CombatEndScreenResultWin:
                    titleFont.PrintOptions(screen, titleX, titleY, fontOptions, "You are triumphant")
                case CombatEndScreenResultLose:
                    titleFont.PrintOptions(screen, titleX, titleY, fontOptions, "You have been defeated")
                case CombatEndScreenResultRetreat:
                    titleFont.PrintOptions(screen, titleX, titleY, fontOptions, "Your forces have retreated")
            }

            extraX, extraY := options.GeoM.Apply(float64(110), float64(fontY))

            options.GeoM.Translate(0, float64(picLength))
            scale.DrawScaled(screen, bottom, &options)

            extraFont.PrintOptions(screen, extraX, extraY, fontOptions, extraText)

            if extraText2 != "" {
                extraY += float64((extraFont.Height() + 1))
                extraFont.RenderWrapped(screen, extraX, extraY, extraText2Render, font.FontOptions{Justify: font.FontJustifyCenter, Scale: scale.ScaleAmount, Options: &options})
            }
        },
    }

    ui.SetElementsFromArray([]*uilib.UIElement{element})

    return ui
}

func (end *CombatEndScreen) Update() CombatEndScreenState {
    end.UI.StandardUpdate()
    return end.State
}

func (end *CombatEndScreen) Draw(screen *ebiten.Image) {
    end.UI.Draw(end.UI, screen)
}
