package mouse

import (
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/scale"

    "github.com/hajimehoshi/ebiten/v2"
)

type GlobalMouse struct {
    CurrentMouse *ebiten.Image
    Enabled bool
}

var Mouse *GlobalMouse

func Initialize(){
    Mouse = &GlobalMouse{
        CurrentMouse: nil,
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
    mouse.CurrentMouse = image
}

func (mouse *GlobalMouse) Draw(screen *ebiten.Image) {
    if mouse != nil && mouse.Enabled && mouse.CurrentMouse != nil {
        x, y := inputmanager.MousePosition()
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(scale.Unscale(float64(x)), scale.Unscale(float64(y)))
        scale.DrawScaled(screen, mouse.CurrentMouse, &options)
    }
}
