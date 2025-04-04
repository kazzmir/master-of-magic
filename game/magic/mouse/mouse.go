package mouse

import (
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/scale"

    "github.com/hajimehoshi/ebiten/v2"
)

type GlobalMouse struct {
    DrawFunc func (*ebiten.Image, *ebiten.DrawImageOptions)
    Enabled bool
    Options ebiten.DrawImageOptions
}

var Mouse *GlobalMouse

func Initialize(){
    Mouse = &GlobalMouse{
        DrawFunc: func(screen *ebiten.Image, options *ebiten.DrawImageOptions) {
        },
        Enabled: true,
    }
}

func (mouse *GlobalMouse) Enable() {
    mouse.Enabled = true
}

func (mouse *GlobalMouse) Disable() {
    mouse.Enabled = false
}

func (mouse *GlobalMouse) SetImage(image *ebiten.Image) {
    mouse.DrawFunc = func(screen *ebiten.Image, options *ebiten.DrawImageOptions) {
        scale.DrawScaled(screen, image, options)
    }
}

func (mouse *GlobalMouse) SetImageFunc(imageFunc func (*ebiten.Image, *ebiten.DrawImageOptions)) {
    mouse.DrawFunc = imageFunc
}

func (mouse *GlobalMouse) Draw(screen *ebiten.Image) {
    if mouse != nil && mouse.Enabled {
        x, y := inputmanager.MousePosition()
        mouse.Options.GeoM.Reset()
        mouse.Options.GeoM.Translate(scale.Unscale(float64(x)), scale.Unscale(float64(y)))
        mouse.DrawFunc(screen, &mouse.Options)
    }
}
