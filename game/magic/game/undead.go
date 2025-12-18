package game

import (
    "context"
    "image"

    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"

    "github.com/hajimehoshi/ebiten/v2"
)

// shows the animation of undead creatures rising from the ground
// and prints text saying how many undead have risen

func MakeUndeadUI(imageCache *util.ImageCache) (*uilib.UIElementGroup, context.Context) {
    // X units rises from the dead to serve its creator

    group := uilib.MakeGroup()

    quit, cancel := context.WithCancel(context.Background())

    fadeDelay := uint64(7)
    getAlpha := group.MakeFadeIn(fadeDelay)

    background, _ := imageCache.GetImage("cmbtfx.lbx", 27, 0)

    rect := util.ImageRect(0, 0, background)
    rect = rect.Add(image.Pt(data.ScreenWidth / 2, data.ScreenHeight / 2))
    rect = rect.Sub(image.Pt(background.Bounds().Dx()/2, background.Bounds().Dy()/2))

    group.AddElement(&uilib.UIElement{
        Rect: rect,
        Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, background, &options)
        },
        LeftClick: func(this *uilib.UIElement) {
            getAlpha = group.MakeFadeOut(fadeDelay)
            group.AddDelay(fadeDelay, func() {
                cancel()
            })
        },
    })

    return group, quit
}
