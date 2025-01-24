package xbr

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

type PngUtil struct {
	blendColors                 bool
	scaleAlpha                  bool
	use3XOriginalImplementation bool
}

// NewPngScaler creates and returns a new instance of PngUtil, a utility for scaling PNG images.
//
// This function initializes a PngUtil struct with the specified configuration for image scaling.
// The PngUtil struct provides methods for scaling PNG images using the xBR algorithm, allowing
// customization through various options.
//
// Parameters:
//
//	blendColors bool: Determines whether color blending should be applied during scaling.
//	                  When set to true, the algorithm blends colors to create smoother transitions.
//	scaleAlpha bool: Indicates whether alpha values should be scaled, affecting transparency handling.
//	use3XOriginalImplementation bool: Specifies which version of the xBR algorithm to use.
//	                                  If set to true, the original 3X implementation of the algorithm is used.
//	                                  If false, an alternative or updated version is used.
//
// Returns:
//
//	*PngUtil: A pointer to a new PngUtil instance, configured with the provided scaling options.
//
// The returned PngUtil instance can be used to perform PNG image scaling operations. The provided
// options allow for control over color blending, alpha scaling, and the choice of xBR algorithm
// implementation, affecting the visual outcome and performance of the scaling process.
func NewPngScaler(blendColors, scaleAlpha, use3XOriginalImplementation bool) *PngUtil {
	return &PngUtil{blendColors: blendColors, scaleAlpha: scaleAlpha, use3XOriginalImplementation: use3XOriginalImplementation}
}

// ScalePng scales a PNG image using the xBR algorithm and saves the result to a file.
//
// This method applies the xBR scaling algorithm to a PNG image specified by `inputFilename`.
// The scaled image is saved to the file specified by `outputFilename`. The scaling factor
// is determined by the `zoom` parameter, which can be set to 2, 3, or 4, corresponding to
// 2x, 3x, or 4x scaling respectively.
//
// Parameters:
//
//	inputFilename string: The file path of the input PNG image to be scaled.
//	outputFilename string: The file path where the scaled PNG image will be saved.
//	zoom byte: The scaling factor (2, 3, or 4) to apply to the image.
//
// Returns:
//
//	error: An error if the scaling process fails or if invalid arguments are provided.
//	       It returns an error if `zoom` is not in the range of 2 to 4, if the input file
//	       cannot be loaded, or if there's an issue saving the scaled image.
//
// The method utilizes an `Xbr` scaler instance (created via `NewXbrScaler`) to perform the
// image scaling. It also uses internal methods `loadPngImage`, `extractPixels`,
// `createImageFromPixels`, and `savePngImage` for handling image data. The choice of xBR
// scaling (2x, 3x, or 4x) depends on the `zoom` parameter. The method ensures that only
// valid zoom levels are used and returns an error for invalid inputs.
func (scl *PngUtil) ScalePng(inputFilename string, outputFilename string, zoom byte) error {
	if zoom < 2 || zoom > 4 {
		return fmt.Errorf("zoom must be 2,3 or 4")
	}

	srcImg, width, height, err := scl.loadPngImage(inputFilename)
	if err != nil {
		return err
	}

	srcPixels := scl.extractPixels(srcImg)
	var dstPixels *[]uint32
	var scaledWidth int
	var scaledHeight int

	scaler := NewXbrScaler(false)

	if zoom == 2 {
		dstPixels, scaledWidth, scaledHeight = scaler.Xbr2x(&srcPixels, width, height, scl.blendColors, scl.scaleAlpha)
	} else if zoom == 3 {
		dstPixels, scaledWidth, scaledHeight = scaler.Xbr3x(&srcPixels, width, height, scl.blendColors, scl.scaleAlpha)
	} else {
		dstPixels, scaledWidth, scaledHeight = scaler.Xbr4x(&srcPixels, width, height, scl.blendColors, scl.scaleAlpha)
	}

	dstImg := scl.createImageFromPixels(*dstPixels, scaledWidth, scaledHeight) // Convert dstPixels back to an image.Image
	return scl.savePngImage(outputFilename, dstImg)
}

func (scl *PngUtil) loadPngImage(filename string) (image.Image, int, int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, -1, -1, err
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, -1, -1, err
	}

	return img, img.Bounds().Dx(), img.Bounds().Dy(), nil
}

func (scl *PngUtil) extractPixels(img image.Image) []uint32 {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	pixels := make([]uint32, width*height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels[y*width+x] = uint32(a>>8)<<24 | uint32(r>>8)<<16 | uint32(g>>8)<<8 | uint32(b>>8)
		}
	}

	return pixels
}

func (scl *PngUtil) savePngImage(filename string, img image.Image) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = png.Encode(file, img)
	return err
}

func (scl *PngUtil) createImageFromPixels(pixels []uint32, width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			pixel := pixels[idx]
			a := uint8((pixel >> 24) & 0xFF)
			r := uint8((pixel >> 16) & 0xFF)
			g := uint8((pixel >> 8) & 0xFF)
			b := uint8(pixel & 0xFF)
			c := color.RGBA{R: r, G: g, B: b, A: a}
			img.SetRGBA(x, y, c)
		}
	}
	return img
}
