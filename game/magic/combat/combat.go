package combat

import (
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/hajimehoshi/ebiten/v2"
)

// hard coding the points is what the real master of magic does
// see Unit_Figure_Position() in UnitView.C
// https://github.com/jbalcomb/ReMoM/blob/8642bb8c46433cc31c058759b28f297947b3b501/src/UnitView.C#L2685
func combatPoints(count int) []image.Point {
    switch count {
        case 0: return nil
        case 1: return []image.Point{image.Pt(0, 0)}
        case 2:
            // FIXME: this was just copied from case 8
            return []image.Point{
                image.Pt(2, -4),
                image.Pt(6, -2),
            }
        case 3:
            // FIXME: this was just copied from case 8
            return []image.Point{
                image.Pt(2, -4),
                image.Pt(6, -2),
                image.Pt(-1, 0),
            }
        case 4:
            // FIXME: this was just copied from case 8
            return []image.Point{
                image.Pt(2, -4),
                image.Pt(6, -2),
                image.Pt(-1, 0),
                image.Pt(-8, 0),
            }
        case 5:
            // FIXME: this was just copied from case 8
            return []image.Point{
                image.Pt(2, -4),
                image.Pt(6, -2),
                image.Pt(-1, 0),
                image.Pt(-8, 0),
                image.Pt(10, 0),
            }
        case 6:
            // FIXME: this was just copied from case 8
            return []image.Point{
                image.Pt(2, -4),
                image.Pt(6, -2),
                image.Pt(-1, 0),
                image.Pt(-8, 0),
                image.Pt(10, 0),
                image.Pt(3, 1),
            }
        case 7:
            // FIXME: this was just copied from case 8
            return []image.Point{
                image.Pt(2, -4),
                image.Pt(6, -2),
                image.Pt(-1, 0),
                image.Pt(-8, 0),
                image.Pt(10, 0),
                image.Pt(3, 1),
                image.Pt(-4, 3),
            }
        case 8:
            // fairly accurate
            return []image.Point{
                image.Pt(2, -4),
                image.Pt(6, -2),
                image.Pt(-1, 0),
                image.Pt(-8, 0),
                image.Pt(10, 0),
                image.Pt(3, 1),
                image.Pt(-4, 3),
                image.Pt(1, 5),
            }
    }

    return nil
}

type Enchanted interface {
    Color() color.Color
}

func RenderCombatUnit(screen *ebiten.Image, use *ebiten.Image, options ebiten.DrawImageOptions, count int, enchantment Enchanted, timeCounter uint64, imageCache *util.ImageCache){
    // the ground is always 6 pixels above the bottom of the unit image
    groundHeight := float64(6 * data.ScreenScale)

    geoM := options.GeoM
    for _, point := range combatPoints(count) {
        options.GeoM.Reset()
        options.GeoM.Translate(float64(point.X * data.ScreenScale), float64(point.Y * data.ScreenScale))
        options.GeoM.Translate(-float64(use.Bounds().Dx() / 2), -float64(use.Bounds().Dy()) + groundHeight)

        options.GeoM.Concat(geoM)

        /*
        x, y := options.GeoM.Apply(0, 0)
        vector.DrawFilledCircle(screen, float32(x), float32(y), 1, color.RGBA{255, 0, 0, 255}, true)
        */

        _ = point
        _ = groundHeight
        // options.GeoM.Translate(-float64(use.Bounds().Dx() / 2), -float64(use.Bounds().Dy()) + groundHeight)
        // options.GeoM.Translate(-13, -22)
        screen.DrawImage(use, &options)

        if enchantment != nil {
            util.DrawOutline(screen, imageCache, use, options.GeoM, options.ColorScale, timeCounter/10, enchantment.Color())
        }
    }
}

func RenderCombatTile(screen *ebiten.Image, imageCache *util.ImageCache, options ebiten.DrawImageOptions){
    // FIXME: make the tile image a parameter
    grass, err := imageCache.GetImage("cmbgrass.lbx", 0, 0)
    if err == nil {
        options.GeoM.Translate(-float64(grass.Bounds().Dx() / 2), -float64(grass.Bounds().Dy() / 2))
        screen.DrawImage(grass, &options)
    }
}
