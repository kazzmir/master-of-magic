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

func RenderArtifactImage(screen *ebiten.Image, imageCache *util.ImageCache, artifact Artifact, counter uint64, options ebiten.DrawImageOptions) *ebiten.Image {
    itemImage, _ := imageCache.GetImageTransform("items.lbx", artifact.Image, 0, "1px-border", add1PxBorder)
    options.GeoM.Translate(-1, -1)
    screen.DrawImage(itemImage, &options)

    enchanted := artifact.HasAbilities()
    if enchanted {
        util.DrawOutline(screen, imageCache, itemImage, options.GeoM, options.ColorScale, counter, data.GetMagicColor(artifact.FirstAbility().MagicType()))
    }

    return itemImage
}

func RenderArtifactBox(screen *ebiten.Image, imageCache *util.ImageCache, artifact Artifact, counter uint64, titleFont *font.Font, attributeFont *font.Font, options ebiten.DrawImageOptions) {
    itemBackground, _ := imageCache.GetImage("itemisc.lbx", 25, 0)
    screen.DrawImage(itemBackground, &options)

    options.GeoM.Translate(float64(10 * data.ScreenScale), float64(8 * data.ScreenScale))

    itemImage := RenderArtifactImage(screen, imageCache, artifact, counter, options)

    x, y := options.GeoM.Apply(float64(itemImage.Bounds().Max.X + 3 * data.ScreenScale), float64(4 * data.ScreenScale))
    titleFont.Print(screen, x, y, float64(data.ScreenScale), options.ColorScale, artifact.Name)

    dot, _ := imageCache.GetImage("itemisc.lbx", 26, 0)
    savedGeom := options.GeoM
    for i, power := range artifact.Powers {
        options.GeoM = savedGeom
        options.GeoM.Translate(float64(3 * data.ScreenScale), float64(26 * data.ScreenScale))
        // integer division is important here
        options.GeoM.Translate(float64((i/2) * data.ScreenScale * 80), float64((i % 2) * 13 * data.ScreenScale))

        screen.DrawImage(dot, &options)

        x, y := options.GeoM.Apply(float64(dot.Bounds().Dx() + 1 * data.ScreenScale), 0)
        attributeFont.Print(screen, x, y, float64(data.ScreenScale), options.ColorScale, power.Name)
    }
}
