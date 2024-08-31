package mainview

import (
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"

    "github.com/hajimehoshi/ebiten/v2"
)

type MainScreenState int

const (
    MainScreenStateRunning MainScreenState = iota
)

type MainScreen struct {
    Counter uint64
    Cache *lbx.LbxCache
    ImageCache util.ImageCache
}

func MakeMainScreen(cache *lbx.LbxCache) *MainScreen {
    return &MainScreen{
        Counter: 0,
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
    }
}

func (main *MainScreen) Update() MainScreenState {
    main.Counter += 1

    return MainScreenStateRunning
}

func (main *MainScreen) Draw(screen *ebiten.Image) {
    var options ebiten.DrawImageOptions
    top, err := main.ImageCache.GetImages("mainscrn.lbx", 0)
    if err == nil {
        use := top[(main.Counter / 4) % uint64(len(top))]
        screen.DrawImage(use, &options)
        options.GeoM.Translate(0, float64(use.Bounds().Dy()))
    }

    background, err := main.ImageCache.GetImage("mainscrn.lbx", 5, 0)
    if err == nil {
        screen.DrawImage(background, &options)
    }

    options.GeoM.Reset()
    options.GeoM.Translate(110, 130)

    continueImage, err := main.ImageCache.GetImage("mainscrn.lbx", 1, 0)
    if err == nil {
        screen.DrawImage(continueImage, &options)
        options.GeoM.Translate(0, float64(continueImage.Bounds().Dy()))
    }

    loadGameImage, err := main.ImageCache.GetImage("mainscrn.lbx", 2, 0)
    if err == nil {
        screen.DrawImage(loadGameImage, &options)
        options.GeoM.Translate(0, float64(loadGameImage.Bounds().Dy()))
    }

    newGameImage, err := main.ImageCache.GetImage("mainscrn.lbx", 3, 0)
    if err == nil {
        screen.DrawImage(newGameImage, &options)
        options.GeoM.Translate(0, float64(newGameImage.Bounds().Dy()))
    }

    exitImage, err := main.ImageCache.GetImage("mainscrn.lbx", 4, 0)
    if err == nil {
        screen.DrawImage(exitImage, &options)
    }
}
