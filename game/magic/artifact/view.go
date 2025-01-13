package artifact

import (
    "image"

    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"

    "github.com/hajimehoshi/ebiten/v2"
)

// enlarge the image by 1 pixel on all sides
func add1PxBorder(src *image.Paletted) image.Image {
    out := image.NewPaletted(image.Rect(0, 0, src.Bounds().Dx()+2, src.Bounds().Dy()+2), src.Palette)

    for y := 0; y < src.Bounds().Dy(); y++ {
        for x := 0; x < src.Bounds().Dx(); x++ {
            out.SetColorIndex(x+1, y+1, src.ColorIndexAt(x, y))
        }
    }

    return out
}

func RenderArtifactBox(screen *ebiten.Image, imageCache *util.ImageCache, artifact Artifact, font *font.Font, options ebiten.DrawImageOptions) {
    itemBackground, _ := imageCache.GetImage("itemisc.lbx", 25, 0)
    screen.DrawImage(itemBackground, &options)

    var itemImage *ebiten.Image

    enchanted := artifact.HasAbilities()

    if enchanted {
        itemImage, _ = imageCache.GetImageTransform("items.lbx", artifact.Image, 0, "1px-border", add1PxBorder)
        options.GeoM.Translate(10, 8)
        screen.DrawImage(itemImage, &options)

        x, y := options.GeoM.Apply(0, 0)
        util.DrawOutline(screen, imageCache, itemImage, x, y, 0, data.GetMagicColor(artifact.FirstAbility().MagicType()))
    } else {
        itemImage, _ = imageCache.GetImage("items.lbx", artifact.Image, 0)
        options.GeoM.Translate(10, 8)
        screen.DrawImage(itemImage, &options)
    }

    x, y := options.GeoM.Apply(float64(itemImage.Bounds().Max.X) + 3, 4)
    font.Print(screen, x, y, 1, options.ColorScale, artifact.Name)

    dot, _ := imageCache.GetImage("itemisc.lbx", 26, 0)
    savedGeom := options.GeoM
    for i, power := range artifact.Powers {
        options.GeoM = savedGeom
        options.GeoM.Translate(3, 26)
        options.GeoM.Translate(float64(i / 2 * 80), float64(i % 2 * 13))

        screen.DrawImage(dot, &options)

        x, y := options.GeoM.Apply(float64(dot.Bounds().Dx() + 1), 0)
        font.Print(screen, x, y, 1, options.ColorScale, power.Name)
    }
}
