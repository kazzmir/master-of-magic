package game

import (
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/lib/coroutine"

    "github.com/hajimehoshi/ebiten/v2"
)

func (game *Game) showVaultScreen() (func (*ebiten.Image), func(coroutine.YieldFunc)) {

    imageCache := util.MakeImageCache(game.Cache)

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            background, _ := imageCache.GetImage("xyz", 0, 0)
            var options ebiten.DrawImageOptions
            screen.DrawImage(background, &options)
        },
    }

    quit := false

    drawer := func (screen *ebiten.Image){
        ui.Draw(ui, screen)
    }

    logic := func (yield coroutine.YieldFunc) {
        for !quit {
            yield()
        }
    }

    return drawer, logic
}
