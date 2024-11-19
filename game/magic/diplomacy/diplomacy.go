package diplomacy

import (
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/hajimehoshi/ebiten/v2"
)

/* player is talking to enemy
 */
func ShowDiplomacyScreen(cache *lbx.LbxCache, player *playerlib.Player, enemy *playerlib.Player) (func (coroutine.YieldFunc), func (*ebiten.Image)) {

    imageCache := util.MakeImageCache(cache)

    quit := false

    logic := func (yield coroutine.YieldFunc) {
        for !quit {
            yield()
        }
    }

    draw := func (screen *ebiten.Image) {
        background, _ := imageCache.GetImage("diplomac.lbx", 0, 0)
        var options ebiten.DrawImageOptions
        screen.DrawImage(background, &options)

        // red left eye
        leftEye, _ := imageCache.GetImage("diplomac.lbx", 2, 0)
        // FIXME: what do the other eye colors mean? is it related to the diplomatic relationship level between the wizards?
        // red right eye
        rightEye, _ := imageCache.GetImage("diplomac.lbx", 13, 0)

        options.GeoM.Translate(63, 58)
        screen.DrawImage(leftEye, &options)

        options.GeoM.Translate(170, 0)
        screen.DrawImage(rightEye, &options)
    }

    return logic, draw
}
