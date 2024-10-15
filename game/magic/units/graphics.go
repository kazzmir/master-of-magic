package units

import (
    "log"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/hajimehoshi/ebiten/v2"
)

func GetUnitBackgroundImage(banner data.BannerType, imageCache *util.ImageCache) (*ebiten.Image, error) {
    index := -1
    switch banner {
        case data.BannerBlue: index = 14
        case data.BannerGreen: index = 15
        case data.BannerPurple: index = 16
        case data.BannerRed: index = 17
        case data.BannerYellow: index = 18
        case data.BannerBrown: index = 19
    }

    image, err := imageCache.GetImage("mapback.lbx", index, 0)
    if err != nil {
        log.Printf("Error: image in mapback.lbx is missing: %v", err)
    }

    return image, err
}

func MakeUpdateUnitColorsFunc(banner data.BannerType) util.ImageTransformFunc {
    return func (original *image.Paletted) image.Image {
        var baseColor color.RGBA

        switch banner {
            case data.BannerBlue: baseColor = color.RGBA{R: 0x00, G: 0x00, B: 0xff, A: 0xff}
            // don't really need to do anything for green because the base color is green
            case data.BannerGreen: baseColor = color.RGBA{R: 0x00, G: 0xf0, B: 0x00, A: 0xff}
            case data.BannerPurple: baseColor = color.RGBA{R: 0x8f, G: 0x30, B: 0xff, A: 0xff}
            case data.BannerRed: baseColor = color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}
            case data.BannerYellow: baseColor = color.RGBA{R: 0xff, G: 0xff, B: 0x00, A: 0xff}
            case data.BannerBrown: baseColor = color.RGBA{R: 0xce, G: 0x65, B: 0x00, A: 0xff}
        }

        light := float64(10)
        original.Palette = util.ClonePalette(original.Palette)
        for i := 0; i < 4; i++ {
            original.Palette[215 + i] = util.Lighten(baseColor, light)
            light -= 10
        }

        return original
    }
}
