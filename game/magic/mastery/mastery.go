package mastery

import (
    "fmt"

    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/fonts"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    fontlib "github.com/kazzmir/master-of-magic/lib/font"

    "github.com/hajimehoshi/ebiten/v2"
)

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

    font := fonts.MakeSpellOfMasteryFonts(cache)

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

func CastSpellOfMastery(cache *lbx.LbxCache, wizard setup.WizardCustom) (coroutine.AcceptYieldFunc, func (*ebiten.Image)) {
    // show wizlab with wizard standing there, vortex animation (splmastr.lbx 29-31) with wizard faces flying around (spelllose.lbx 0-13)

    // then show wizard from win.lbx 3-16, background win.lbx 0, hands 17-22, world animation 2, and some text about being the master of magic. then show score screen

    imageCache := util.MakeImageCache(cache)

    font := fonts.MakeSpellOfMasteryFonts(cache)

    talkingImages, _ := imageCache.GetImages("win.lbx", int(wizard.Base) + 3)
    talkingHead := util.MakeAnimation(talkingImages, true)

    // FIXME: hand image is dependent on the wizard
    hands, _ := imageCache.GetImage("win.lbx", 17, 0)

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
        return nil
    }

    draw := func (screen *ebiten.Image) {
        background, _ := imageCache.GetImage("win.lbx", 0, 0)
        var options ebiten.DrawImageOptions
        scale.DrawScaled(screen, background, &options)

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
