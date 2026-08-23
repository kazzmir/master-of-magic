package artifact

import (
    "image"
    "slices"
    "cmp"
    "fmt"
    "strings"

    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"

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
    scale.DrawScaled(screen, itemImage, &options)

    enchanted := artifact.HasAbilities()
    if enchanted {
        util.DrawOutline(screen, imageCache, itemImage, scale.ScaleGeom(options.GeoM), options.ColorScale, counter, data.GetMagicColor(artifact.FirstAbility().MagicType()))
    }

    return itemImage
}

func RenderArtifactBox(screen *ebiten.Image, imageCache *util.ImageCache, artifact Artifact, counter uint64, titleFont *font.Font, attributeFont *font.Font, options ebiten.DrawImageOptions) {
    itemBackground, _ := imageCache.GetImage("itemisc.lbx", 25, 0)
    scale.DrawScaled(screen, itemBackground, &options)

    options.GeoM.Translate(float64(10), float64(8))

    itemImage := RenderArtifactImage(screen, imageCache, artifact, counter, options)

    x, y := options.GeoM.Apply(float64(itemImage.Bounds().Max.X + 3), float64(4))
    titleFont.PrintOptions(screen, x, y, font.FontOptions{DropShadow: true, Scale: scale.ScaleAmount, Options: &options}, artifact.Name)

    dot, _ := imageCache.GetImage("itemisc.lbx", 26, 0)
    savedGeom := options.GeoM

    maxRowLength := itemBackground.Bounds().Dx() - 20

    type PowerColumn struct {
        name string
        length float64
    }

    // a row can contain X powers (usually at most 2). If a row contains a power that is too long
    // then that row will only contain 1 power, and the next power will be on the next row
    type Row struct {
        Columns []PowerColumn
    }

    var rows []Row

    columns := make([]PowerColumn, len(artifact.Powers))
    for i := range artifact.Powers {
        power := artifact.Powers[i]

        name := power.Name
        if enchantment := power.Ability.Enchantment(); enchantment != data.UnitEnchantmentNone {
            name = fmt.Sprintf("%v, as %v spell", power.Ability.Name(), strings.ToLower(enchantment.Name()))
        }

        width := float64(attributeFont.MeasureTextWidth(name, 1))
        powerLength := width + 3 + float64(dot.Bounds().Dx()) + 1

        // force column alignment
        if powerLength < 80 {
            powerLength = 80
        }
        columns[i] = PowerColumn{name: name, length: powerLength}
    }

    columns = slices.SortedFunc(slices.Values(columns), func(a, b PowerColumn) int {
        return cmp.Compare(a.length, b.length)
    })

    // combine powers into rows greedily
    for i := range columns {
        column := columns[i]

        added := false
        for rowI := range rows {
            size := float64(0)
            for _, col := range rows[rowI].Columns {
                size += col.length
            }
            if size + column.length <= float64(maxRowLength) {
                rows[rowI].Columns = append(rows[rowI].Columns, column)
                added = true
                break
            }
        }

        if !added {
            rows = append(rows, Row{Columns: []PowerColumn{column}})
        }
    }

    for rowI, row := range rows {
        options.GeoM = savedGeom

        options.GeoM.Translate(float64(3), float64(26))
        options.GeoM.Translate(0, float64(rowI * 13))

        for _, column := range row.Columns {
            scale.DrawScaled(screen, dot, &options)

            x, y := options.GeoM.Apply(float64(dot.Bounds().Dx() + 1), 0)
            attributeFont.PrintOptions(screen, x, y, font.FontOptions{DropShadow: true, Options: &options, Scale: scale.ScaleAmount}, column.name)

            options.GeoM.Translate(column.length + 2, 0)
        }
    }
}
