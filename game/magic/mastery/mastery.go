package mastery

import (
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
)

func ShowSpellOfMasteryScreen(cache *lbx.LbxCache, wizard string) (coroutine.AcceptYieldFunc, func (*ebiten.Image)) {
    // play animations spellscr.lbx 67, 68, 69, 70 in order
    // overlay text "$wizard has started casting the Spell of Mastery"

    imageCache := util.MakeImageCache(cache)

    order := []int{67, 68, 69, 70}
    index := 0

    images1, _ := imageCache.GetImages("spellscr.lbx", 67)
    animation := util.MakeAnimation(images1, false)

    var logic coroutine.AcceptYieldFunc

    var counter uint64 = 0

    logic = func (yield coroutine.YieldFunc) error {
        for !animation.Done() {
            counter += 1
            if yield() != nil {
                return nil
            }

            if counter % 10 == 0 {
                if !animation.Next() {
                    index += 1
                    if index < len(order) {
                        images, _ := imageCache.GetImages("spellscr.lbx", order[index])
                        animation = util.MakeAnimation(images, false)
                    }
                }
            }
        }

        return nil
    }

    draw := func (screen *ebiten.Image) {
        var options ebiten.DrawImageOptions
        scale.DrawScaled(screen, animation.Frame(), &options)
    }

    return logic, draw
}

func CastSpellOfMastery() {
    // show wizlab with wizard standing there, vortex animation (splmastr.lbx 29-31) with wizard faces flying around (spelllose.lbx 0-13)

    // then show wizard from win.lbz 3-16, background win.lbz 0, hands 17-22, world animation 2, and some text about being the master of magic. then show score screen
}
