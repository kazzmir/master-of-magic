package xbr

import (
    "image"
)

func ScaleImage(input image.Image, zoom int) image.Image {

    util := NewPngScaler(true, true, false)
	scaler := NewXbrScaler(false)

	srcPixels := util.extractPixels(input)
	var dstPixels *[]uint32
	var scaledWidth int
	var scaledHeight int

    width := input.Bounds().Dx()
    height := input.Bounds().Dy()

    switch zoom {
        case 2: dstPixels, scaledWidth, scaledHeight = scaler.Xbr2x(&srcPixels, width, height, util.blendColors, util.scaleAlpha)
        case 3: dstPixels, scaledWidth, scaledHeight = scaler.Xbr3x(&srcPixels, width, height, util.blendColors, util.scaleAlpha)
        case 4: dstPixels, scaledWidth, scaledHeight = scaler.Xbr4x(&srcPixels, width, height, util.blendColors, util.scaleAlpha)
        default:
            return input
	}

	output := util.createImageFromPixels(*dstPixels, scaledWidth, scaledHeight)

    return output
}
