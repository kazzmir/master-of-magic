package game

import (
	"fmt"
	"image"
    "context"

	playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
	uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/kazzmir/master-of-magic/game/magic/audio"
	"github.com/kazzmir/master-of-magic/game/magic/scale"
	"github.com/kazzmir/master-of-magic/game/magic/mirror"
	"github.com/kazzmir/master-of-magic/game/magic/util"
	"github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/hajimehoshi/ebiten/v2"
)

// onPlayerSelectedCallback CAN'T receive nil as argument
func makeSelectSpellBlastTargetUI(finish context.CancelFunc, cache *lbx.LbxCache, imageCache *util.ImageCache, castingPlayer *playerlib.Player, playersInGame int, onPlayerSelectedCallback func(selectedPlayer *playerlib.Player) bool) *uilib.UIElementGroup {
    group := uilib.MakeGroup()

    const fadeSpeed = 7
    getAlpha := group.MakeFadeIn(fadeSpeed)

    var layer uilib.UILayer = 2

    x := 77
    y := 10

    fonts := MakeSpellSpecialUIFonts(cache)
    header := "Choose target for a Spell Blast spell"

    // A func for creating a sparks element when a target is selected
    createSparksElement := func (faceRect image.Rectangle) *uilib.UIElement {
        sparksCreationTick := group.Counter // Needed for sparks animation
        return &uilib.UIElement{
            Layer: layer+2,
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                const ticksPerFrame = 5
                frameToShow := int((group.Counter - sparksCreationTick) / ticksPerFrame) % 6
                background, _ := imageCache.GetImage("specfx.lbx", 40, frameToShow)
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(getAlpha())
                options.GeoM.Translate(float64(faceRect.Min.X - 5), float64(faceRect.Min.Y - 10))
                scale.DrawScaled(screen, background, &options)
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
            options.GeoM.Translate(float64(x), float64(y))
            scale.DrawScaled(screen, background, &options)

            mx, my := options.GeoM.Apply(float64(84), float64(10))
            fonts.BigOrange.PrintWrapCenter(screen, mx, my, 120, scale.ScaleAmount, options.ColorScale, header)
        },
        Order: 0,
    })

    // Wizard faces/gems/broken gems
    crystalPicture, _ := imageCache.GetImage("magic.lbx", 6, 0)
    brokenCrystalPicture, _ := imageCache.GetImage("magic.lbx", 51, 0)
    wizardFacesOffsets := []image.Point{ image.Pt(24, 37), image.Pt(101, 37), image.Pt(24, 98), image.Pt(101, 98) }

    drawnWizardFaces := 0
    var currentMouseoverPlayer *playerlib.Player
    for index, target := range castingPlayer.GetKnownPlayers() {
        if target.Defeated || target.Banished {
            continue
        }
        portrait, _ := imageCache.GetImage("lilwiz.lbx", mirror.GetWizardPortraitIndex(target.Wizard.Base, target.Wizard.Banner), 0)
        faceRect := util.ImageRect((x + wizardFacesOffsets[index].X), (y + wizardFacesOffsets[index].Y), portrait)
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
                    group.AddDelay(60, func(){
                        header = fmt.Sprintf("%s has been spell blasted", target.Wizard.Name)
                        group.AddDelay(113, func(){
                            getAlpha = group.MakeFadeOut(fadeSpeed)
                            group.AddDelay(7, func(){
                                finish()
                            })
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
                    scale.DrawScaled(screen, brokenCrystalPicture, &options)
                } else {
                    scale.DrawScaled(screen, portrait, &options)
                    // Draw current spell being cast
                    spellText := "None"
                    if target.CastingSpell.Valid() {
                        spellText = target.CastingSpell.Name   
                    }
                    fonts.InfoOrange.PrintWrapCenter(
                        screen, 
                        float64(faceRect.Min.X + faceRect.Dx()/2), float64(faceRect.Max.Y + 6),
                        120, scale.ScaleAmount, options.ColorScale, spellText,
                    )
                    // Draw current spell cost/progress
                    if currentMouseoverPlayer == target {
                        fonts.BigOrange.PrintWrapCenter(screen, float64(x + 47), float64(y + 159), 120, scale.ScaleAmount, options.ColorScale, fmt.Sprintf("%d MP", target.CastingSpellProgress))
                    }
                }
            },
        })
        drawnWizardFaces++
    }
    // Empty crystals
    for emptyPlaceIndex := drawnWizardFaces; emptyPlaceIndex < 4; emptyPlaceIndex++ {
        crystalRect := util.ImageRect((x + wizardFacesOffsets[emptyPlaceIndex].X), (y + wizardFacesOffsets[emptyPlaceIndex].Y), crystalPicture)
        crystalToDraw := crystalPicture
        if emptyPlaceIndex > playersInGame - 1 {
            crystalToDraw = brokenCrystalPicture
            crystalRect = crystalRect.Add(image.Point{-1, -1})
        }
        group.AddElement(&uilib.UIElement{
            Layer: layer - 1, // That's correct, gems are behind the UI form itself (should fit into transparent windows)
            Rect: crystalRect,
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(getAlpha())
                options.GeoM.Translate(float64(crystalRect.Min.X), float64(crystalRect.Min.Y))
                scale.DrawScaled(screen, crystalToDraw, &options)
            },
        })
    }

    // cancel button
    cancel, _ := imageCache.GetImages("spellscr.lbx", 71)
    cancelIndex := 0
    cancelRect := util.ImageRect((x + 83), (y + 155), cancel[0])
    group.AddElement(&uilib.UIElement{
        Layer: layer+1,
        Rect: cancelRect,
        LeftClick: func(element *uilib.UIElement){
            cancelIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            cancelIndex = 0
            finish()
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            scale.DrawScaled(screen, cancel[cancelIndex], &options)
        },
    })

    return group
}

// FIXME: if there are no selectable wizards (because they are all defeated, or the player just doesn't have relations with any others)
// then return an error instead of showing the UI
func makeSelectTargetWizardUI(finish context.CancelFunc, cache *lbx.LbxCache, imageCache *util.ImageCache, 
    initialHeader string, sparksGraphicIndex int, sparksSoundIndex int, castingPlayer *playerlib.Player, playersInGame int,
    // This callback receives a spell target and returns if the selection was valid (bool) and the new menu header change (string)
    onPlayerSelectedCallback func(selectedPlayer *playerlib.Player) (bool, string)) *uilib.UIElementGroup {

    group := uilib.MakeGroup()

    const fadeSpeed = 7
    getAlpha := group.MakeFadeIn(fadeSpeed)

    var layer uilib.UILayer = 2

    x := 77
    y := 10

    fonts := MakeSpellSpecialUIFonts(cache)
    header := initialHeader

    // A func for creating a sparks element when a target is selected
    createSparksElement := func (faceRect image.Rectangle) *uilib.UIElement {
        sparksCreationTick := group.Counter // Needed for sparks animation
        return &uilib.UIElement{
            Layer: layer+2,
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                const ticksPerFrame = 5
                frameToShow := int((group.Counter - sparksCreationTick) / ticksPerFrame) % 6
                background, _ := imageCache.GetImage("specfx.lbx", sparksGraphicIndex, frameToShow)
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(getAlpha())
                options.GeoM.Translate(float64(faceRect.Min.X - 8), float64(faceRect.Min.Y - 5))
                scale.DrawScaled(screen, background, &options)
            },
        }
    }


    // The form itself
    group.AddElement(&uilib.UIElement{
        Layer: layer,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("spellscr.lbx", 0, 0)
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            options.GeoM.Translate(float64(x), float64(y))
            scale.DrawScaled(screen, background, &options)

            mx, my := options.GeoM.Apply(float64(84), float64(10))
            fonts.BigOrange.PrintWrapCenter(screen, mx, my, 150, scale.ScaleAmount, options.ColorScale, header)
        },
        Order: 0,
    })

    // Wizard faces/gems/broken gems
    crystalPicture, _ := imageCache.GetImage("magic.lbx", 6, 0)
    brokenCrystalPicture, _ := imageCache.GetImage("magic.lbx", 51, 0)
    wizardFacesOffsets := []image.Point{ image.Pt(24, 37), image.Pt(101, 37), image.Pt(24, 98), image.Pt(101, 98) }

    drawnWizardFaces := 0
    var currentMouseoverPlayer *playerlib.Player
    for index, target := range castingPlayer.GetKnownPlayers() {
        if target.Banished {
            continue
        }
        portrait, _ := imageCache.GetImage("lilwiz.lbx", mirror.GetWizardPortraitIndex(target.Wizard.Base, target.Wizard.Banner), 0)
        faceRect := util.ImageRect((x + wizardFacesOffsets[index].X), (y + wizardFacesOffsets[index].Y), portrait)
        // FIXME: right click should open the wizard's mirror info screen
        group.AddElement(&uilib.UIElement{
            Layer: layer+1,
            Rect: faceRect,
            // Try casting the spell on click.
            LeftClickRelease: func(element *uilib.UIElement){
                if target.Defeated {
                    return
                }

                castSuccessful, newHeader := onPlayerSelectedCallback(target)
                if castSuccessful {
                    // Spell cast successfully. Change header, create sound, add delay, draw the sparks and remove this uigroup.
                    group.AddElement(createSparksElement(faceRect))
                    sound, err := audio.LoadSound(cache, sparksSoundIndex)
                    if err == nil {
                        sound.Play()
                    }
                    group.AddDelay(60, func(){
                        header = newHeader
                        group.AddDelay(113, func(){
                            getAlpha = group.MakeFadeOut(fadeSpeed)
                            group.AddDelay(7, func(){
                                finish()
                            })
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
                    scale.DrawScaled(screen, brokenCrystalPicture, &options)
                } else {
                    scale.DrawScaled(screen, portrait, &options)
                    diplomaticRelation, has := target.GetDiplomaticRelation(castingPlayer)
                    if has {
                        fonts.InfoOrange.PrintWrapCenter(
                            screen,
                            float64(faceRect.Min.X + faceRect.Dx()/2), float64(faceRect.Max.Y + 6),
                            120, scale.ScaleAmount, options.ColorScale, diplomaticRelation.Description(),
                        )
                    }
                }
            },
        })
        drawnWizardFaces += 1
    }
    // Empty crystals
    for emptyPlaceIndex := drawnWizardFaces; emptyPlaceIndex < 4; emptyPlaceIndex++ {
        crystalRect := util.ImageRect((x + wizardFacesOffsets[emptyPlaceIndex].X), (y + wizardFacesOffsets[emptyPlaceIndex].Y), crystalPicture)
        crystalToDraw := crystalPicture
        if emptyPlaceIndex > playersInGame - 1 {
            crystalToDraw = brokenCrystalPicture
            crystalRect = crystalRect.Add(image.Point{-1, -1})
        }
        group.AddElement(&uilib.UIElement{
            Layer: layer - 1, // That's correct, gems are behind the UI form itself (should fit into transparent windows)
            Rect: crystalRect,
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(getAlpha())
                options.GeoM.Translate(float64(crystalRect.Min.X), float64(crystalRect.Min.Y))
                scale.DrawScaled(screen, crystalToDraw, &options)
            },
        })
    }

    // cancel button
    cancel, _ := imageCache.GetImages("spellscr.lbx", 71)
    cancelIndex := 0
    cancelRect := util.ImageRect((x + 45), (y + 155), cancel[0])
    group.AddElement(&uilib.UIElement{
        Layer: layer+1,
        Rect: cancelRect,
        LeftClick: func(element *uilib.UIElement){
            cancelIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            cancelIndex = 0
            finish()
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            scale.DrawScaled(screen, cancel[cancelIndex], &options)
        },
    })

    return group
}
