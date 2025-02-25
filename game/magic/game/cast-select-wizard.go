package game

import (
	"github.com/kazzmir/master-of-magic/game/magic/data"
	fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
	"github.com/kazzmir/master-of-magic/game/magic/mirror"
	playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
	uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
	"github.com/kazzmir/master-of-magic/game/magic/util"
	"github.com/kazzmir/master-of-magic/lib/lbx"

	"github.com/hajimehoshi/ebiten/v2"
)

// selectedCallback MAY return nil, that means the spell was canceled after cast.
// TODO: transform this into more unversal reusable form.
func makeSelectSpellBlastTargetUI(cache *lbx.LbxCache, imageCache *util.ImageCache, castingPlayer *playerlib.Player, selectedCallback func(selectedPlayer *playerlib.Player)) *uilib.UIElementGroup {
    group := uilib.MakeGroup()

    var layer uilib.UILayer = 2

    x := 77
    y := 10

    fonts := fontslib.MakeSpellbookFonts(cache)

    group.AddElement(&uilib.UIElement{
        Layer: layer,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("spellscr.lbx", 72, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(x * data.ScreenScale), float64(y * data.ScreenScale))
            screen.DrawImage(background, &options)

            mx, my := options.GeoM.Apply(float64(84 * data.ScreenScale), float64(10 * data.ScreenScale))
            fonts.BigOrange.PrintWrapCenter(screen, mx, my, 120. * float64(data.ScreenScale), float64(data.ScreenScale), options.ColorScale, "Choose target for a Spell Blast spell")
            // mx, my = options.GeoM.Apply(float64((x + 34) * data.ScreenScale), float64((y + 20) * data.ScreenScale))
            // fonts.BigOrange.Print(screen, mx, my, float64(data.ScreenScale), options.ColorScale, "Spell Blast spell")
        },
        Order: 0,
    })

    // Wizard faces
    wizardFacesOffsets := [][2]int {{24, 37}, {101, 37}, {24, 98}, {101, 98}}
    drawnWizardFaces := 0
    for index, target := range castingPlayer.GetKnownPlayers() {
        portrait, _ := imageCache.GetImage("lilwiz.lbx", mirror.GetWizardPortraitIndex(target.Wizard.Base, target.Wizard.Banner), 0)
        faceRect := util.ImageRect((x + wizardFacesOffsets[index][0]) * data.ScreenScale, (y + wizardFacesOffsets[index][1]) * data.ScreenScale, portrait)
        group.AddElement(&uilib.UIElement{
            Layer: 5,
            Rect: faceRect,
            LeftClickRelease: func(element *uilib.UIElement){
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(faceRect.Min.X), float64(faceRect.Min.Y))
                screen.DrawImage(portrait, &options)
            },
        })
        drawnWizardFaces++
    }
    // Empty crystals
    for emptyPlaceIndex := drawnWizardFaces; emptyPlaceIndex < 4; emptyPlaceIndex++ {
        crystalPicture, _ := imageCache.GetImage("magic.lbx", 6, 0)
        faceRect := util.ImageRect((x + wizardFacesOffsets[emptyPlaceIndex][0]) * data.ScreenScale, (y + wizardFacesOffsets[emptyPlaceIndex][1]) * data.ScreenScale, crystalPicture)
        group.AddElement(&uilib.UIElement{
            Layer: 5,
            Rect: faceRect,
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(faceRect.Min.X), float64(faceRect.Min.Y))
                screen.DrawImage(crystalPicture, &options)
            },
        })
    }

    // cancel button
    cancel, _ := imageCache.GetImages("spellscr.lbx", 71)
    cancelIndex := 0
    cancelRect := util.ImageRect((x + 83) * data.ScreenScale, (y + 155) * data.ScreenScale, cancel[0])
    group.AddElement(&uilib.UIElement{
        Layer: 5,
        Rect: cancelRect,
        LeftClick: func(element *uilib.UIElement){
            cancelIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            cancelIndex = 0
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            screen.DrawImage(cancel[cancelIndex], &options)
        },
    })

    return group
}
