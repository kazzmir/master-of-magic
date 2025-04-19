package mirror

import (
    "log"
    "fmt"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/draw"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
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

type MirrorFonts struct {
    Name *font.Font
    Small *font.Font
    Hero *font.Font
}

func MakeMirrorFonts(cache *lbx.LbxCache) MirrorFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Could not load fonts: %v", err)
        return MirrorFonts{}
    }

    return MirrorFonts{
        Name: loader(fontslib.BigOrangeGradient2),
        Small: loader(fontslib.DescriptionFont),
        Hero: loader(fontslib.LightFontSmall),
    }
}

func MakeMirrorUI(cache *lbx.LbxCache, player *playerlib.Player, ui *uilib.UI) *uilib.UIElement {
    cornerX := 50
    cornerY := 1

    fonts := MakeMirrorFonts(cache)

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

    wrappedAbilities := fonts.Small.CreateWrappedText(float64(160), 1, setup.JoinAbilities(player.Wizard.Retorts))

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
            scale.DrawScaled(screen, background, &options)

            if portrait != nil {
                options.GeoM.Translate(float64(11), float64(11))
                scale.DrawScaled(screen, portrait, &options)
            }

            centerOptions := font.FontOptions{Justify: font.FontJustifyCenter, Options: &options, Scale: scale.ScaleAmount, DropShadow: true}

            fonts.Name.PrintOptions(screen, float64(cornerX + 110), float64(cornerY + 10), centerOptions, player.Wizard.Name)

            fonts.Small.PrintOptions(screen, float64(cornerX + 30), float64(cornerY + 75), centerOptions, fmt.Sprintf("%v GP", player.Gold))
            fonts.Small.PrintOptions(screen, float64(cornerX + 170), float64(cornerY + 75), font.FontOptions{Justify: font.FontJustifyRight, Options: &options, Scale: scale.ScaleAmount, DropShadow: true}, fmt.Sprintf("%v MP", player.Mana))

            options.GeoM.Translate(float64(34), float64(55))
            newRand := rand.New(rand.NewPCG(player.BookOrderSeed1, player.BookOrderSeed2))
            draw.DrawBooks(screen, options, &imageCache, player.Wizard.Books, newRand)

            if player.GetFame() > 0 {
                fonts.Hero.PrintOptions(screen, float64(cornerX + 90), float64(cornerY + 95), centerOptions, fmt.Sprintf("%v Fame", player.GetFame()))
            }

            fonts.Small.RenderWrapped(screen, float64(cornerX + 13), float64(cornerY + 112), wrappedAbilities, font.FontOptions{Scale: scale.ScaleAmount, Options: &options, DropShadow: true})

            fonts.Hero.PrintOptions(screen, float64(cornerX + 90), float64(cornerY + 131), centerOptions, "Heroes")

            heroX := cornerX + 13
            heroY := cornerY + 142
            for _, hero := range player.AliveHeroes() {
                fonts.Small.PrintOptions(screen, float64(heroX), float64(heroY), font.FontOptions{Options: &options, Scale: scale.ScaleAmount, DropShadow: true}, hero.GetFullName())
                heroY += fonts.Small.Height()
            }
        },
    }

    return element
}
