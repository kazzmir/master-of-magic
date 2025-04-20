package cartographer

import (
    _ "log"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/data"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

func MakeCartographer(cache *lbx.LbxCache, cities []*citylib.City, arcanusMap *maplib.Map, arcanusFog data.FogMap, myrrorMap *maplib.Map, myrrorFog data.FogMap) (func(coroutine.YieldFunc), func (*ebiten.Image)) {

    imageCache := util.MakeImageCache(cache)

    counter := uint64(0)

    getAlpha := util.MakeFadeIn(7, &counter)

    logic := func (yield coroutine.YieldFunc) {
        quit := false
        for !quit {
            counter += 1
            if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
                quit = true
            }

            if yield() != nil {
                return
            }
        }

        getAlpha = util.MakeFadeOut(7, &counter)
        for range 7 {
            counter += 1
            yield()
        }
    }

    draw := func (screen *ebiten.Image) {
        background, _ := imageCache.GetImage("reload.lbx", 2, 0)
        var options ebiten.DrawImageOptions
        options.ColorScale.ScaleAlpha(getAlpha())
        scale.DrawScaled(screen, background, &options)
    }

    return logic, draw
}
