package game

import (
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/lib/coroutine"

    "github.com/hajimehoshi/ebiten/v2"
)

func (game *Game) showVaultScreen() (func(coroutine.YieldFunc), func (*ebiten.Image)) {

    imageCache := util.MakeImageCache(game.Cache)

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            background, _ := imageCache.GetImage("armylist.lbx", 5, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(data.ScreenWidth / 2 - background.Bounds().Dx() / 2), 2)
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

    return logic, drawer
}
