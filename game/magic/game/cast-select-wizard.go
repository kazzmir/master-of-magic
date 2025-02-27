package game

import (
	"fmt"
	"image"

	fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
	playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
	uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/mirror"
	"github.com/kazzmir/master-of-magic/game/magic/util"
	"github.com/kazzmir/master-of-magic/lib/lbx"
	"github.com/hajimehoshi/ebiten/v2"
)

// onPlayerSelectedCallback CAN'T receive nil as argument
// TODO: transform this into more unversal reusable form.
func makeSelectSpellBlastTargetUI(ui *uilib.UI, cache *lbx.LbxCache, imageCache *util.ImageCache, castingPlayer *playerlib.Player, playersInGame int, onPlayerSelectedCallback func(selectedPlayer *playerlib.Player) bool) *uilib.UIElementGroup {
    const fadeSpeed = 7
    getAlpha := ui.MakeFadeIn(fadeSpeed)
    
    group := uilib.MakeGroup()

    var layer uilib.UILayer = 2

    x := 77
    y := 10

    fonts := fontslib.MakeSpellSpecialUIFonts(cache)
    header := "Choose target for a Spell Blast spell"

    // A func for creating a sparks element when a target is selected
    createSparksElement := func (faceRect image.Rectangle) *uilib.UIElement {
        sparksTick := 0 // Needed for sparks animation
        return &uilib.UIElement{
            Layer: layer+2,
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                const ticksPerFrame = 5
                frameToShow := (sparksTick / ticksPerFrame) % 6
                background, _ := imageCache.GetImage("specfx.lbx", 40, frameToShow)
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(getAlpha())
                options.GeoM.Translate(float64(faceRect.Min.X - 5 * data.ScreenScale), float64(faceRect.Min.Y - 10 * data.ScreenScale))
                screen.DrawImage(background, &options)
                sparksTick++
                if (sparksTick == ticksPerFrame * 6 * 6 - fadeSpeed) {
                    getAlpha = ui.MakeFadeOut(fadeSpeed)
                }
            },
        }
    }


    // The form itself
    group.AddElement(&uilib.UIElement{
        Layer: layer,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("spellscr.lbx", 72, 0)
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            options.GeoM.Translate(float64(x * data.ScreenScale), float64(y * data.ScreenScale))
            screen.DrawImage(background, &options)

            mx, my := options.GeoM.Apply(float64(84 * data.ScreenScale), float64(10 * data.ScreenScale))
            fonts.BigOrange.PrintWrapCenter(screen, mx, my, 120. * float64(data.ScreenScale), float64(data.ScreenScale), options.ColorScale, header)
        },
        Order: 0,
    })

    // Wizard faces/gems/broken gems
    crystalPicture, _ := imageCache.GetImage("magic.lbx", 6, 0)
    brokenCrystalPicture, _ := imageCache.GetImage("magic.lbx", 51, 0)
    wizardFacesOffsets := [][2]int {{24, 37}, {101, 37}, {24, 98}, {101, 98}}

    drawnWizardFaces := 0
    var currentMouseoverPlayer *playerlib.Player
    for index, target := range castingPlayer.GetKnownPlayers() {
        if target.Defeated || target.Banished {
            continue
        }
        portrait, _ := imageCache.GetImage("lilwiz.lbx", mirror.GetWizardPortraitIndex(target.Wizard.Base, target.Wizard.Banner), 0)
        faceRect := util.ImageRect((x + wizardFacesOffsets[index][0]) * data.ScreenScale, (y + wizardFacesOffsets[index][1]) * data.ScreenScale, portrait)
        group.AddElement(&uilib.UIElement{
            Layer: layer+1,
            Rect: faceRect,
            // Try casting the spell on click.
            LeftClickRelease: func(element *uilib.UIElement){
                if onPlayerSelectedCallback(target) {
                    // Spell cast successfully. Change header, create sound, add delay, draw the sparks and remove this uigroup.
                    group.AddElement(createSparksElement(faceRect))
                    sound, err := audio.LoadSound(cache, 29)
                    if err == nil {
                        sound.Play()
                    }
                    ui.AddDelay(80, func(){
                        header = fmt.Sprintf("%s has been spell blasted", target.Wizard.Name)
                        ui.AddDelay(100, func(){
                            ui.RemoveGroup(group)
                        })
                    })
                }
            },
            Inside: func(element *uilib.UIElement, x int, y int){
                currentMouseoverPlayer = target
            },
            NotInside: func(element *uilib.UIElement){
                if currentMouseoverPlayer == target {
                    currentMouseoverPlayer = nil
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(getAlpha())
                options.GeoM.Translate(float64(faceRect.Min.X), float64(faceRect.Min.Y))
                if target.Defeated {
                    screen.DrawImage(brokenCrystalPicture, &options)
                } else {
                    screen.DrawImage(portrait, &options)
                    // Draw current spell being cast
                    spellText := "None"
                    if target.CastingSpell.Valid() {
                        spellText = target.CastingSpell.Name   
                    }
                    fonts.InfoOrange.PrintWrapCenter(
                        screen, 
                        float64(faceRect.Min.X + faceRect.Dx()/2), float64(faceRect.Max.Y + 6 * data.ScreenScale),
                        120. * float64(data.ScreenScale), float64(data.ScreenScale), options.ColorScale, spellText,
                    )
                    // Draw current spell cost/progress
                    if currentMouseoverPlayer == target {
                        fonts.BigOrange.PrintWrapCenter(screen, float64((x + 47) * data.ScreenScale), float64((y + 159) * data.ScreenScale), 120. * float64(data.ScreenScale), float64(data.ScreenScale), options.ColorScale, fmt.Sprintf("%d MP", target.CastingSpellProgress))
                    }
                }
            },
        })
        drawnWizardFaces++
    }
    // Empty crystals
    for emptyPlaceIndex := drawnWizardFaces; emptyPlaceIndex < 4; emptyPlaceIndex++ {
        crystalRect := util.ImageRect((x + wizardFacesOffsets[emptyPlaceIndex][0]) * data.ScreenScale, (y + wizardFacesOffsets[emptyPlaceIndex][1]) * data.ScreenScale, crystalPicture)
        crystalToDraw := crystalPicture
        if emptyPlaceIndex > playersInGame - 1 {
            crystalToDraw = brokenCrystalPicture
            crystalRect = crystalRect.Add(image.Point{-1 * data.ScreenScale, -1 * data.ScreenScale})
        }
        group.AddElement(&uilib.UIElement{
            Layer: 3,
            Rect: crystalRect,
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(getAlpha())
                options.GeoM.Translate(float64(crystalRect.Min.X), float64(crystalRect.Min.Y))
                screen.DrawImage(crystalToDraw, &options)
            },
        })
    }

    // cancel button
    cancel, _ := imageCache.GetImages("spellscr.lbx", 71)
    cancelIndex := 0
    cancelRect := util.ImageRect((x + 83) * data.ScreenScale, (y + 155) * data.ScreenScale, cancel[0])
    group.AddElement(&uilib.UIElement{
        Layer: layer+1,
        Rect: cancelRect,
        LeftClick: func(element *uilib.UIElement){
            cancelIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            cancelIndex = 0
            ui.RemoveGroup(group)
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            screen.DrawImage(cancel[cancelIndex], &options)
        },
    })

    return group
}
