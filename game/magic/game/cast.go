package game

import (
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/util"

    "github.com/hajimehoshi/ebiten/v2"
)

func (game *Game) doCastSpell(yield coroutine.YieldFunc, player *playerlib.Player, spell spellbook.Spell) {
    switch spell.Name {
        case "Earth Lore":
            game.doCastEarthLore(yield, player)
    }
}

func (game *Game) doCastEarthLore(yield coroutine.YieldFunc, player *playerlib.Player) {
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    pics, _ := game.ImageCache.GetImages("specfx.lbx", 0)
    animation := util.MakeAnimation(pics, false)
    animationAlive := true

    game.Drawer = func(screen *ebiten.Image, game *Game) {
        oldDrawer(screen, game)

        if animationAlive {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(20, 20)
            screen.DrawImage(animation.Frame(), &options)
        }
    }

    for i := 0; i < 100; i++ {
        game.Counter += 1

        if game.Counter % 5 == 0 && animationAlive {
            animationAlive = animation.Next()
        }

        yield()
    }
}
