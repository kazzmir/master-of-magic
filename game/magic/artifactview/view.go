package artifactview

import (
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"

    "github.com/hajimehoshi/ebiten/v2"
)

func RenderArtifactBox(screen *ebiten.Image, imageCache *util.ImageCache, artifact artifact.Artifact, font *font.Font, options ebiten.DrawImageOptions) {
    itemBackground, _ := imageCache.GetImage("itemisc.lbx", 25, 0)
    screen.DrawImage(itemBackground, &options)

    itemImage, _ := imageCache.GetImage("items.lbx", artifact.Image, 0)
    options.GeoM.Translate(10, 8)
    screen.DrawImage(itemImage, &options)

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