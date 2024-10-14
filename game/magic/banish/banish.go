package banish

import (
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
)

func ShowBanishAnimation(cache *lbx.LbxCache, attackingWizard *playerlib.Player, defeatedWizard *playerlib.Player) (func (coroutine.YieldFunc) error, func (*ebiten.Image)) {

    imageCache := util.MakeImageCache(cache)

    background, _ := imageCache.GetImage("wizlab.lbx", 19, 0)

    draw := func (screen *ebiten.Image){
        var options ebiten.DrawImageOptions
        screen.DrawImage(background, &options)
    }

    logic := func (yield coroutine.YieldFunc) error {
        for i := 0; i < 200; i++ {
            yield()
        }

        return nil
    }

    return logic, draw
}
