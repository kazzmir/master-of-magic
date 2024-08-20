package city

import (
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"

    "github.com/hajimehoshi/ebiten/v2"
)

type CityScreen struct {
    LbxCache *lbx.LbxCache
    ImageCache util.ImageCache
}

func MakeCityScreen(cache *lbx.LbxCache) *CityScreen {
    cityScreen := &CityScreen{
        LbxCache: cache,
        ImageCache: util.MakeImageCache(cache),
    }
    return cityScreen
}

func (cityScreen *CityScreen) Update() {
}

func (cityScreen *CityScreen) Draw(screen *ebiten.Image) {
    ui, err := cityScreen.ImageCache.GetImage("backgrnd.lbx", 6, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        screen.DrawImage(ui, &options)
    }
}
