package mirror

import (
    "log"
    "fmt"
    "image/color"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/draw"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/hajimehoshi/ebiten/v2"
)

// returns the index into lilwiz.lbx
func GetWizardPortraitIndex(base data.WizardBase, banner data.BannerType) int {
    bannerIndex := 0
    switch banner {
        case data.BannerBlue: bannerIndex = 0
        case data.BannerGreen: bannerIndex = 1
        case data.BannerPurple: bannerIndex = 2
        case data.BannerRed: bannerIndex = 3
        case data.BannerYellow: bannerIndex = 4
    }

    wizardIndex := 0

    switch base {
        case data.WizardMerlin: wizardIndex = 0
        case data.WizardRaven: wizardIndex = 5
        case data.WizardSharee: wizardIndex = 10
        case data.WizardLoPan: wizardIndex = 15
        case data.WizardJafar: wizardIndex = 20
        case data.WizardOberic: wizardIndex = 25
        case data.WizardRjak: wizardIndex = 30
        case data.WizardSssra: wizardIndex = 35
        case data.WizardTauron: wizardIndex = 40
        case data.WizardFreya: wizardIndex = 45
        case data.WizardHorus: wizardIndex = 50
        case data.WizardAriel: wizardIndex = 55
        case data.WizardTlaloc: wizardIndex = 60
        case data.WizardKali: wizardIndex = 65
    }

    return wizardIndex + bannerIndex
}

func MakeMirrorUI(cache *lbx.LbxCache, player *playerlib.Player, ui *uilib.UI) *uilib.UIElement {
    cornerX := 50 * data.ScreenScale
    cornerY := 1 * data.ScreenScale

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Could not read fonts: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Could not read fonts: %v", err)
        return nil
    }

    yellow := color.RGBA{R: 0xea, G: 0xb6, B: 0x00, A: 0xff}
    yellowPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        yellow, yellow, yellow,
        yellow, yellow, yellow,
    }

    nameFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowPalette)

    smallFont := font.MakeOptimizedFontWithPalette(fonts[0], yellowPalette)

    heroFont := font.MakeOptimizedFontWithPalette(fonts[2], yellowPalette)

    var element *uilib.UIElement

    getAlpha := ui.MakeFadeIn(7)

    var portrait *ebiten.Image

    imageCache := util.MakeImageCache(cache)

    portrait, _ = imageCache.GetImage("lilwiz.lbx", GetWizardPortraitIndex(player.Wizard.Base, player.Wizard.Banner), 0)

    doClose := func(){
        getAlpha = ui.MakeFadeOut(7)
        ui.AddDelay(7, func(){
            ui.RemoveElement(element)
        })
    }

    wrappedAbilities := smallFont.CreateWrappedText(float64(160 * data.ScreenScale), float64(data.ScreenScale), setup.JoinAbilities(player.Wizard.Retorts))

    element = &uilib.UIElement{
        Layer: 1,
        LeftClick: func(this *uilib.UIElement){
            doClose()
        },
        NotLeftClicked: func(this *uilib.UIElement){
            doClose()
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("backgrnd.lbx", 4, 0)

            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(cornerX), float64(cornerY))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(background, &options)

            if portrait != nil {
                options.GeoM.Translate(float64(11 * data.ScreenScale), float64(11 * data.ScreenScale))
                screen.DrawImage(portrait, &options)
            }

            nameFont.PrintCenter(screen, float64(cornerX + 110 * data.ScreenScale), float64(cornerY + 10 * data.ScreenScale), float64(data.ScreenScale), options.ColorScale, player.Wizard.Name)

            smallFont.PrintCenter(screen, float64(cornerX + 30 * data.ScreenScale), float64(cornerY + 75 * data.ScreenScale), float64(data.ScreenScale), options.ColorScale, fmt.Sprintf("%v GP", player.Gold))
            smallFont.PrintRight(screen, float64(cornerX + 170 * data.ScreenScale), float64(cornerY + 75 * data.ScreenScale), float64(data.ScreenScale), options.ColorScale, fmt.Sprintf("%v MP", player.Mana))

            options.GeoM.Translate(float64(34 * data.ScreenScale), float64(55 * data.ScreenScale))
            newRand := rand.New(rand.NewPCG(player.BookOrderSeed1, player.BookOrderSeed2))
            draw.DrawBooks(screen, options, &imageCache, player.Wizard.Books, newRand)

            if player.GetFame() > 0 {
                heroFont.PrintCenter(screen, float64(cornerX + 90 * data.ScreenScale), float64(cornerY + 95 * data.ScreenScale), float64(data.ScreenScale), options.ColorScale, fmt.Sprintf("%v Fame", player.GetFame()))
            }

            smallFont.RenderWrapped(screen, float64(cornerX + 13 * data.ScreenScale), float64(cornerY + 112 * data.ScreenScale), wrappedAbilities, options.ColorScale, font.FontOptions{})

            heroFont.PrintCenter(screen, float64(cornerX + 90 * data.ScreenScale), float64(cornerY + 131 * data.ScreenScale), float64(data.ScreenScale), options.ColorScale, "Heroes")

            heroX := cornerX + 13 * data.ScreenScale
            heroY := cornerY + 142 * data.ScreenScale
            for _, hero := range player.AliveHeroes() {
                smallFont.Print(screen, float64(heroX), float64(heroY), float64(data.ScreenScale), options.ColorScale, hero.FullName())
                heroY += smallFont.Height() * data.ScreenScale
            }
        },
    }

    return element
}
