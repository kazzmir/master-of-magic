package spellbook

import (
    "bytes"
    "fmt"
    "image"
    "image/color"
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/hajimehoshi/ebiten/v2"
)

// pass in desc.lbx
func ReadSpellDescriptions(file *lbx.LbxFile) ([]string, error) {
    entries, err := file.RawData(0)
    if err != nil {
        return nil, err
    }

    reader := bytes.NewReader(entries)

    count, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    if count > 10000 {
        return nil, fmt.Errorf("Spell count was too high: %v", count)
    }

    size, err := lbx.ReadUint16(reader)
    if err != nil {
        return nil, err
    }

    if size > 10000 {
        return nil, fmt.Errorf("Size of each spell entry was too high: %v", size)
    }

    var descriptions []string

    for i := 0; i < int(count); i++ {
        data := make([]byte, size)
        _, err := reader.Read(data)

        if err != nil {
            break
        }

        nullByte := bytes.IndexByte(data, 0)
        if nullByte != -1 {
            descriptions = append(descriptions, string(data[0:nullByte]))
        } else {
            descriptions = append(descriptions, string(data))
        }
    }

    return descriptions, nil
}

func MakeSpellBookUI(ui *uilib.UI, cache *lbx.LbxCache) []*uilib.UIElement {
    var elements []*uilib.UIElement

    imageCache := util.MakeImageCache(cache)

    fadeSpeed := uint64(7)

    getAlpha := ui.MakeFadeIn(fadeSpeed)

    bookFlip, _ := imageCache.GetImages("book.lbx", 1)
    bookFlipIndex := uint64(0)
    bookFlipReverse := false

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts: %v", err)
        return nil
    }

    red := color.RGBA{R: 0x5a, G: 0, B: 0, A: 0xff}
    redPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        red, red, red,
        red, red, red,
        red, red, red,
        red, red, red,
        red, red, red,
        red, red, red,
    }

    titleFont := font.MakeOptimizedFontWithPalette(fonts[5], redPalette)

    // title font: fonts[5]
    // mystery font title: fonts[7]
    // mystery font normal: fonts[6]

    elements = append(elements, &uilib.UIElement{
        Layer: 1,
        NotLeftClicked: func(this *uilib.UIElement){
            getAlpha = ui.MakeFadeOut(fadeSpeed)
            ui.AddDelay(fadeSpeed, func(){
                ui.RemoveElements(elements)
            })
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := imageCache.GetImage("scroll.lbx", 6, 0)

            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(background, &options)

            titleFont.PrintCenter(screen, 90, 11, 1, options.ColorScale, "Unit Spells")

            animationIndex := ui.Counter
            if bookFlipIndex > 0 && (animationIndex - bookFlipIndex) / 6 < uint64(len(bookFlip)) {
                index := (animationIndex - bookFlipIndex) / 6
                if bookFlipReverse {
                    index = uint64(len(bookFlip)) - 1 - index
                }
                options.GeoM.Translate(0, 0)
                screen.DrawImage(bookFlip[index], &options)
            }

        },
    })

    // left page turn
    leftTurn, _ := imageCache.GetImage("scroll.lbx", 7, 0)
    leftRect := image.Rect(15, 9, 15 + leftTurn.Bounds().Dx(), 9 + leftTurn.Bounds().Dy())
    elements = append(elements, &uilib.UIElement{
        Rect: leftRect,
        Layer: 1,
        LeftClick: func(this *uilib.UIElement){
            bookFlipIndex = ui.Counter
            bookFlipReverse = true
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(leftRect.Min.X), float64(leftRect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(leftTurn, &options)
        },
    })

    // right page turn
    rightTurn, _ := imageCache.GetImage("scroll.lbx", 8, 0)
    rightRect := image.Rect(289, 9, 295 + rightTurn.Bounds().Dx(), 5 + rightTurn.Bounds().Dy())
    elements = append(elements, &uilib.UIElement{
        Rect: rightRect,
        Layer: 1,
        LeftClick: func(this *uilib.UIElement){
            bookFlipIndex = ui.Counter
            bookFlipReverse = false
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(rightRect.Min.X), float64(rightRect.Min.Y))
            options.ColorScale.ScaleAlpha(getAlpha())
            screen.DrawImage(rightTurn, &options)
        },
    })

    return elements
}
