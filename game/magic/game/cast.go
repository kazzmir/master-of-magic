package game

import (
    _ "log"
    "image"
    "math/rand"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/util"
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

    pics, _ := game.ImageCache.GetImages("specfx.lbx", 0)

    type SpellAnimation struct {
        Delay uint64
        Animation *util.Animation
        Done bool
        Point image.Point
    }

    var animations []SpellAnimation

    for i := 0; i < 5; i++ {
        animation := util.MakeAnimation(pics, false)
        animations = append(animations, SpellAnimation{
            Delay: uint64(rand.Intn(20)),
            Animation: animation,
            Done: false,
            Point: image.Point{X: x + rand.Intn(100) - 50, Y: y + rand.Intn(100) - 50},
        })
    }

    game.Drawer = func(screen *ebiten.Image, game *Game) {
        oldDrawer(screen, game)

        for i := 0; i < len(animations); i++ {
            if animations[i].Delay == 0 && animations[i].Done == false {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(animations[i].Point.X), float64(animations[i].Point.Y))
                screen.DrawImage(animations[i].Animation.Frame(), &options)
            }
        }
    }

    // FIXME: play earth lore sound

    quit := false
    for !quit {
        game.Counter += 1

        quit = true
        for j := 0; j < len(animations); j++ {
            if animations[j].Delay > 0 {
                animations[j].Delay -= 1
            } else if game.Counter % 5 == 0 {
                animations[j].Done = !animations[j].Animation.Next()
            }

            if animations[j].Delay > 0 || !animations[j].Done {
                quit = false
            }
        }

        yield()
    }
}
