package city

import (
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
)

type CityScreen struct {
    LbxCache *lbx.LbxCache
}

func MakeCityScreen(cache *lbx.LbxCache) *CityScreen {
    cityScreen := &CityScreen{
        LbxCache: cache,
    }
    return cityScreen
}

func (cityScreen *CityScreen) Update() {
}

func (cityScreen *CityScreen) Draw(screen *ebiten.Image) {
}
