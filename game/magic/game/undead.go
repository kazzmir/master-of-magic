package game

import (
    "context"
    "image"
    "fmt"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"

    "github.com/hajimehoshi/ebiten/v2"
)

// shows the animation of undead creatures rising from the ground
// and prints text saying how many undead have risen

// zombie=true, show the zombie, otherwise show the ghoul
// units is how many units have risen from the dead (shows up in the text)
func MakeUndeadUI(cache *lbx.LbxCache, imageCache *util.ImageCache, zombie bool, units int) (*uilib.UIElementGroup, context.Context) {
    // X units rises from the dead to serve its creator
    loader, err := fontslib.Loader(cache)
    if err != nil {
        quit, cancel := context.WithCancel(context.Background())
        cancel()
        return uilib.MakeGroup(), quit
    }

    showFont := loader(fontslib.LightFont)

    group := uilib.MakeGroup()

    quit, cancel := context.WithCancel(context.Background())

    fadeDelay := uint64(7)
    getAlpha := group.MakeFadeIn(fadeDelay)

    background, _ := imageCache.GetImage("cmbtfx.lbx", 27, 0)

    rect := util.ImageRect(0, 0, background)
    rect = rect.Add(image.Pt(data.ScreenWidth / 2, data.ScreenHeight / 2))
    rect = rect.Sub(image.Pt(background.Bounds().Dx()/2, background.Bounds().Dy()/2))

    // the tall zombie guy
    zombieWarrior, _ := imageCache.GetImage("cmbtfx.lbx", 32, 0)
    ghoulGuy, _ := imageCache.GetImage("monster.lbx", 12, 0)

    useImage := zombieWarrior
    if !zombie {
        // the ghoul image is smaller than zombie, so resize it
        ghoulImage := ebiten.NewImage(zombieWarrior.Bounds().Dx(), zombieWarrior.Bounds().Dy())
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(80, 30)
        ghoulImage.DrawImage(ghoulGuy, &options)
        useImage = ghoulImage
    }

    backgroundGuys1, _ := imageCache.GetImage("cmbtfx.lbx", 30, 0)
    backgroundGuys2, _ := imageCache.GetImage("cmbtfx.lbx", 31, 0)

    frontRocks, _ := imageCache.GetImage("cmbtfx.lbx", 29, 0)

    // draw onto buffer first so that alpha blending doesn't affect overlapping images
    buffer := ebiten.NewImage(background.Bounds().Dx(), background.Bounds().Dy())

    var text string
    if units == 1 {
        text = "1 unit rises from the dead to serve its creator."
    } else {
        text = fmt.Sprintf("%d units rise from the dead to serve their creator.", units)
    }
    wrapped := showFont.CreateWrappedText(float64(buffer.Bounds().Dx() - 10), 1, text)

    group.AddElement(&uilib.UIElement{
        Rect: rect,
        Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            buffer.DrawImage(background, &options)

            areaRect := background.Bounds()
            areaRect.Max.Y -= 36
            drawArea := buffer.SubImage(areaRect).(*ebiten.Image)

            counter := group.Counter / 2

            backgroundY := int(min(counter, 19))
            options.GeoM.Translate(0, 19)
            options.GeoM.Translate(0, float64(-backgroundY))

            drawArea.DrawImage(backgroundGuys1, &options)

            options.GeoM.Reset()
            backgroundY = int(min(counter, 42))
            options.GeoM.Translate(0, 42)
            options.GeoM.Translate(0, float64(-backgroundY))
            drawArea.DrawImage(backgroundGuys2, &options)

            options.GeoM.Reset()
            drawArea.DrawImage(frontRocks, &options)

            moveY := int(min(counter, uint64(useImage.Bounds().Dy() / 2)))

            options.GeoM.Reset()
            options.GeoM.Translate(0, float64(useImage.Bounds().Dy() / 2))
            options.GeoM.Translate(0, float64(-moveY))

            drawArea.DrawImage(useImage, &options)

            showFont.RenderWrapped(buffer, float64(buffer.Bounds().Dx()/2), float64(buffer.Bounds().Dy()-30), wrapped, font.FontOptions{Justify: font.FontJustifyCenter, DropShadow: true})

            options.GeoM.Reset()
            options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, buffer, &options)
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
