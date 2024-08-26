package magicview

import (
    "image"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/hajimehoshi/ebiten/v2"
)

type MagicScreenState int

const (
    MagicScreenStateRunning MagicScreenState = iota
    MagicScreenStateDone
)

type MagicScreen struct {
    Cache *lbx.LbxCache
    ImageCache util.ImageCache
}

func MakeMagicScreen(cache *lbx.LbxCache) *MagicScreen {
    magic := &MagicScreen{
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
    }
    return magic
}

func (magic *MagicScreen) Update() MagicScreenState {

    return MagicScreenStateRunning
}

func (magic *MagicScreen) Draw(screen *ebiten.Image){
    background, err := magic.ImageCache.GetImage("magic.lbx", 0, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        screen.DrawImage(background, &options)
    }

    gemPositions := []image.Point{
        image.Pt(24, 4),
        image.Pt(101, 4),
        image.Pt(178, 4),
        image.Pt(255, 4),
    }

    for _, position := range gemPositions {
        // FIXME: the gem color is based on what the banner color of the known wizard is
        gemUnknown, err := magic.ImageCache.GetImage("magic.lbx", 6, 0)
        if err == nil {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(position.X), float64(position.Y))
            screen.DrawImage(gemUnknown, &options)
        }
    }

    manaStaff, err := magic.ImageCache.GetImage("magic.lbx", 7, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(29, 83)
        screen.DrawImage(manaStaff, &options)
    }

    researchStaff, err := magic.ImageCache.GetImage("magic.lbx", 9, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(75, 85)
        screen.DrawImage(researchStaff, &options)
    }

    skillStaff, err := magic.ImageCache.GetImage("magic.lbx", 11, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(122, 83)
        screen.DrawImage(skillStaff, &options)
    }

}
