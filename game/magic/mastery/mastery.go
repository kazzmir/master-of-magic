package mastery

import (
    "fmt"
    "log"
    "slices"

    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/fonts"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    fontlib "github.com/kazzmir/master-of-magic/lib/font"

    "github.com/hajimehoshi/ebiten/v2"
)

type SpellOfMasteryFonts struct {
    Font *fontlib.Font
    RedFont *fontlib.Font
}

func MakeSpellOfMasteryFonts(cache *lbx.LbxCache) *SpellOfMasteryFonts {
    loader, err := fonts.Loader(cache)
    if err != nil {
        log.Printf("Unable to load fonts: %v", err)
        return nil
    }

    return &SpellOfMasteryFonts{
        Font: loader(fonts.HugeOrange),
        RedFont: loader(fonts.HugeRed),
    }
}

func ShowSpellOfMasteryScreen(cache *lbx.LbxCache, wizard string) (coroutine.AcceptYieldFunc, func (*ebiten.Image)) {
    // play animations spellscr.lbx 67, 68, 69, 70 in order
    // overlay text "$wizard has started casting the Spell of Mastery"

    imageCache := util.MakeImageCache(cache)

    order := []int{67, 68, 69, 70}
    index := 0

    images1, _ := imageCache.GetImages("spellscr.lbx", order[0])
    animation := util.MakeAnimation(images1, false)

    var logic coroutine.AcceptYieldFunc

    var counter uint64 = 0

    font := MakeSpellOfMasteryFonts(cache)

    wrapped := font.Font.CreateWrappedText(200, 1, fmt.Sprintf("%v has started casting the Spell of Mastery", wizard))

    logic = func (yield coroutine.YieldFunc) error {
        yield()
        for !inputmanager.LeftClick() && counter < 60 * 10 {
            counter += 1
            if yield() != nil {
                return nil
            }

            if counter % 9 == 0 {
                if !animation.Next() {
                    index += 1
                    if index < len(order) {
                        images, _ := imageCache.GetImages("spellscr.lbx", order[index])
                        animation = util.MakeAnimation(images, index == len(order) - 1)
                    }
                }
            }
        }

        yield()

        return nil
    }

    draw := func (screen *ebiten.Image) {
        var options ebiten.DrawImageOptions
        scale.DrawScaled(screen, animation.Frame(), &options)

        font.Font.RenderWrapped(screen, 160, 160, wrapped, fontlib.FontOptions{Justify: fontlib.FontJustifyCenter, DropShadow: true, Scale: scale.ScaleAmount})
    }

    return logic, draw
}

// shows a vortex in the wizard lab with the main caster raising and lowering their arms, while orbs of the other wizards fly around the vortex,
// break up, and get absorbed into the vortex
func LabVortexScreen(cache *lbx.LbxCache, caster data.WizardBase, losers []data.WizardBase) (coroutine.AcceptYieldFunc, func (*ebiten.Image)) {

    imageCache := util.MakeImageCache(cache)

    casterImages, _ := imageCache.GetImages("splmastr.lbx", 14 + int(caster))
    reversedImages := slices.Clone(casterImages)
    slices.Reverse(reversedImages)
    casterAnimation := util.MakeAnimation(append(casterImages, reversedImages...), true)

    vortexOpenImages, _ := imageCache.GetImages("splmastr.lbx", 29)
    vortexOpenAnimation := util.MakeAnimation(vortexOpenImages, false)

    vortexImages, _ := imageCache.GetImages("splmastr.lbx", 30)
    vortexAnimation := util.MakeAnimation(vortexImages, true)

    vortexCloseImages, _ := imageCache.GetImages("splmastr.lbx", 31)
    vortexCloseAnimation := util.MakeAnimation(vortexCloseImages, false)

    loserIndex := -1
    loserAnimation := util.MakeAnimation(nil, false)

    sparkleImages, _ := imageCache.GetImages("splmastr.lbx", 28)
    sparkleAnimation := util.MakeAnimation(sparkleImages, false)

    var counter uint64 = 0

    vortexMode := 0

    logic := func (yield coroutine.YieldFunc) error {
        // open animation
        for !vortexOpenAnimation.Done() {
            if inputmanager.LeftClick() {
                return nil
            }

            counter += 1
            if yield() != nil {
                return nil
            }

            if counter % 9 == 0 {
                casterAnimation.Next()
                vortexOpenAnimation.Next()
            }
        }

        vortexMode = 1

        // vortex animation + wizard orbs
        for {
            if inputmanager.LeftClick() {
                return nil
            }

            if sparkleAnimation.Done() {
                if loserAnimation.Done() {
                    if loserIndex < len(losers) - 1 {
                        loserIndex += 1
                        images, _ := imageCache.GetImages("splmastr.lbx", int(losers[loserIndex]))
                        loserAnimation = util.MakeAnimation(images, false)
                        sparkleAnimation = util.MakeAnimation(sparkleImages, false)
                    } else {
                        break
                    }
                }
            }

            counter += 1
            if yield() != nil {
                return nil
            }

            if counter % 9 == 0 {
                casterAnimation.Next()
                vortexAnimation.Next()
                loserAnimation.Next()
                sparkleAnimation.Next()
            }
        }

        vortexMode = 2

        // close animation
        for !vortexCloseAnimation.Done() {
            if inputmanager.LeftClick() {
                return nil
            }

            counter += 1
            if yield() != nil {
                return nil
            }

            if counter % 9 == 0 {
                casterAnimation.Next()
                vortexCloseAnimation.Next()
            }
        }

        return nil
    }

    draw := func (screen *ebiten.Image) {
        wizlab, _ := imageCache.GetImage("wizlab.lbx", 19, 0)
        var options ebiten.DrawImageOptions
        scale.DrawScaled(screen, wizlab, &options)

        options.GeoM.Translate(68, 82)
        scale.DrawScaled(screen, casterAnimation.Frame(), &options)

        switch vortexMode {
            case 0:
                options.GeoM.Reset()
                options.GeoM.Translate(84, 5)
                scale.DrawScaled(screen, vortexOpenAnimation.Frame(), &options)
            case 1:
                options.GeoM.Reset()
                options.GeoM.Translate(90, 8)
                scale.DrawScaled(screen, vortexAnimation.Frame(), &options)
                if !sparkleAnimation.Done() {
                    options.GeoM.Translate(4, 55)
                    scale.DrawScaled(screen, sparkleAnimation.Frame(), &options)
                } else if loserAnimation.Frame() != nil {
                    options.GeoM.Translate(-90, 0)
                    scale.DrawScaled(screen, loserAnimation.Frame(), &options)
                }
            case 2:
                options.GeoM.Reset()
                options.GeoM.Translate(84, 5)
                scale.DrawScaled(screen, vortexCloseAnimation.Frame(), &options)
        }
    }

    return logic, draw
}

func SpellOfMasteryEndScreen(cache *lbx.LbxCache, wizard data.WizardBase) (coroutine.AcceptYieldFunc, func (*ebiten.Image)) {
    // show wizlab with wizard standing there, vortex animation (splmastr.lbx 29-31) with wizard faces flying around (spelllose.lbx 0-13)

    // then show wizard from win.lbx 3-16, background win.lbx 0, hands 17-22, world animation 2, and some text about being the master of magic. then show score screen

    imageCache := util.MakeImageCache(cache)

    font := MakeSpellOfMasteryFonts(cache)

    talkingImages, _ := imageCache.GetImages("win.lbx", int(wizard) + 3)
    talkingHead := util.MakeAnimation(talkingImages, true)

    handIndex := 17

    // these are mostly just guesses
    switch wizard {
        case data.WizardMerlin: handIndex = 17
        case data.WizardRaven: handIndex = 20
        case data.WizardSharee: handIndex = 19
        case data.WizardLoPan: handIndex = 17
        case data.WizardJafar: handIndex = 21
        case data.WizardOberic: handIndex = 21
        case data.WizardRjak: handIndex = 22
        case data.WizardSssra: handIndex = 22
        case data.WizardTauron: handIndex = 17
        case data.WizardFreya: handIndex = 18
        case data.WizardHorus: handIndex = 21
        // ariel is verified
        case data.WizardAriel: handIndex = 18
        case data.WizardTlaloc: handIndex = 20
        case data.WizardKali: handIndex = 19
    }

    hands, _ := imageCache.GetImage("win.lbx", handIndex, 0)

    worldImages, _ := imageCache.GetImages("win.lbx", 2)
    worldAnimation := util.MakeAnimation(worldImages, false)

    var counter uint64 = 0

    text := []string{
        "Having conquered both the",
        "world of Arcanus and Myrror",
        "I and only I remain the one",
        "and true Master of Magic.",
    }

    textIndex := 0
    textCounter := 0
    maxTextCounter := int(60 * 2.1)
    fadeScale := float32(1)

    logic := func (yield coroutine.YieldFunc) error {
        yield()
        for !inputmanager.LeftClick() && !worldAnimation.Done() {
            counter += 1
            if yield() != nil {
                return nil
            }

            textCounter += 1
            if textCounter >= maxTextCounter {
                if textIndex < len(text) - 1 {
                    textIndex += 1
                    textCounter = 0
                }
            }

            if counter % 9 == 0 {
                talkingHead.Next()
                worldAnimation.Next()
            }
        }

        for range 10 {
            fadeScale -= 0.1
            if yield() != nil {
                break
            }
        }

        return nil
    }

    draw := func (screen *ebiten.Image) {
        backgrounds, _ := imageCache.GetImages("win.lbx", 0)
        var options ebiten.DrawImageOptions
        options.ColorScale.ScaleAlpha(fadeScale)
        scale.DrawScaled(screen, backgrounds[(counter / 8) % uint64(len(backgrounds))], &options)

        options.GeoM.Translate(95, 5)
        scale.DrawScaled(screen, talkingHead.Frame(), &options)

        options.GeoM.Reset()
        options.GeoM.Translate(20, 140)
        scale.DrawScaled(screen, hands, &options)

        options.GeoM.Reset()
        options.GeoM.Translate(5, 5)
        scale.DrawScaled(screen, worldAnimation.Frame(), &options)

        if textIndex < len(text) {
            if textCounter < 10 {
                options.ColorScale.ScaleAlpha(float32(textCounter) / 10)
            } else if textCounter > int(maxTextCounter - 10) && textIndex < len(text) - 1 {
                options.ColorScale.ScaleAlpha(float32(maxTextCounter - textCounter) / 10)
            }
            font.RedFont.PrintOptions(screen, 160, 170, fontlib.FontOptions{Justify: fontlib.FontJustifyCenter, DropShadow: true, Scale: scale.ScaleAmount, Options: &options}, text[textIndex])
        }
    }

    return logic, draw
}
