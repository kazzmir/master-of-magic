package game

import (
    _ "log"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    // "github.com/kazzmir/master-of-magic/game/magic/data"

    "github.com/hajimehoshi/ebiten/v2"
)

func (game *Game) doCastSpell(yield coroutine.YieldFunc, player *playerlib.Player, spell spellbook.Spell) {
    switch spell.Name {
        case "Earth Lore":
            screenX := 100
            screenY := 100
            game.doCastEarthLore(yield, player, screenX, screenY)

            tileX := game.cameraX + screenX / game.Map.TileWidth()
            tileY := game.cameraY + screenY / game.Map.TileHeight()

            player.LiftFog(tileX, tileY, 5)
    }
}

func (game *Game) doCastEarthLore(yield coroutine.YieldFunc, player *playerlib.Player, x int, y int) {
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    pics, _ := game.ImageCache.GetImages("specfx.lbx", 45)

    animation := util.MakeAnimation(pics, false)

    game.Drawer = func(screen *ebiten.Image, game *Game) {
        oldDrawer(screen, game)

        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(x - animation.Frame().Bounds().Dx() / 2), float64(y - animation.Frame().Bounds().Dy() / 2))
        screen.DrawImage(animation.Frame(), &options)
    }

    // FIXME: play earth lore sound
    // probably soundfx.lbx sound 28
    // newsound.fx 18

    sound, err := audio.LoadNewSound(game.Cache, 18)
    if err == nil {
        sound.Play()
    }

    quit := false
    for !quit {
        game.Counter += 1

        quit = false
        if game.Counter % 6 == 0 {
            quit = !animation.Next()
        }

        yield()
    }
}
