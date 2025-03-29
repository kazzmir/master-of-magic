package mastery

import (
    "log"
    "image/color"
    "fmt"

    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
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

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        return nil, nil
    }

    fonts, err := fontlib.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil, nil
    }

    orangePalette := color.Palette{
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x0},
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xc6, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xc6, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xc1, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xc1, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xc1, B: 0x0, A: 0xff},
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x0},
        color.RGBA{R: 0xff, G: 0xc1, B: 0x0, A: 0xff},
        color.RGBA{R: 0xe3, G: 0xb0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xe3, G: 0xb0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xc1, B: 0x0, A: 0xff},
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x0},
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x0},
    }

    font := fontlib.MakeOptimizedFontWithPalette(fonts[5], orangePalette)

    wrapped := font.CreateWrappedText(200, 1, fmt.Sprintf("%v has started casting the Spell of Mastery", wizard))

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

        font.RenderWrapped(screen, 160, 160, wrapped, fontlib.FontOptions{Justify: fontlib.FontJustifyCenter, DropShadow: true, Scale: scale.ScaleAmount})
    }

    return logic, draw
}

func CastSpellOfMastery() {
    // show wizlab with wizard standing there, vortex animation (splmastr.lbx 29-31) with wizard faces flying around (spelllose.lbx 0-13)

    // then show wizard from win.lbz 3-16, background win.lbz 0, hands 17-22, world animation 2, and some text about being the master of magic. then show score screen
}
