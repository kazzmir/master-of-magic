package xbr

import (
	"math"
)

const (
	// Masks for ARGB
	RedMask   uint32 = 0x000000FF
	GreenMask uint32 = 0x0000FF00
	BlueMask  uint32 = 0x00FF0000
	AlphaMask uint32 = 0xFF000000

	// Thresholds
	ThresholdY float64 = 48
	ThresholdU float64 = 7
	ThresholdV float64 = 6
)

type Xbr struct {
	use3XOriginalImplementation bool
}

// NewXbrScaler creates and returns a new instance of the Xbr scaler.
//
// This function initializes an Xbr struct, which implements the xBR scaling algorithm for pixel art images.
// The xBR algorithm is used for upscaling low-resolution pixel art while preserving its original artistic style.
//
// Parameter:
//
//	use3XOriginalImplementation bool: A flag that determines which version of the xBR algorithm to use.
//	                                  If set to true, the original 3X implementation of the algorithm is used.
//	                                  If false, an alternative or updated version of the algorithm is used.
//
// Returns:
//
//	*Xbr: A pointer to a new instance of the Xbr struct configured with the specified algorithm version.
//
// The returned Xbr instance can be used to perform pixel art scaling operations. The choice of algorithm
// version (original 3X implementation or an alternative) can affect the visual results and performance
// of the scaling process.
func NewXbrScaler(use3XOriginalImplementation bool) *Xbr {
	return &Xbr{use3XOriginalImplementation: use3XOriginalImplementation}
}

// Xbr2x applies the 2x xBR scaling algorithm to a given pixel array.
//
// This method takes a slice of uint32 representing the original pixel data in ARGB format,
// along with the width and height of the original image. It then scales the image by a factor
// of 2 using the xBR algorithm, which is designed for pixel art upscaling.
//
// Parameters:
//
//	pixelArray *[]uint32: Pointer to a slice of uint32 representing the original pixel data in ARGB format.
//	width, height int: The width and height of the original image.
//	blendColors: A boolean flag indicating whether color blending should be used during the scaling process.
//	             When set to true, the algorithm blends colors to create smoother transitions.
//	scaleAlpha: A boolean flag indicating whether alpha values should be scaled, affecting the handling of transparency.
//
// Returns:
//
//	*[]uint32: Pointer to a slice of uint32 containing the scaled pixel data.
//	int: The width of the scaled image, which is twice the width of the original image.
//	int: The height of the scaled image, which is twice the height of the original image.
//
// The method uses the computeXbr2x function internally to process each pixel and write the results
// to a new slice that represents the scaled image. The returned slice contains the pixel data of the
// scaled image, and its dimensions are double those of the input image.
func (xbr *Xbr) Xbr2x(pixelArray *[]uint32, width, height int, blendColors, scaleAlpha bool) (*[]uint32, int, int) {
	scaledWidth := width * 2
	scaledHeight := height * 2
	scaledPixelArray := make([]uint32, scaledWidth*scaledHeight)

	for c := 0; c < width; c++ {
		for d := 0; d < height; d++ {
			xbr.computeXbr2x(pixelArray, c, d, width, height, &scaledPixelArray, c*2, d*2, scaledWidth, blendColors, scaleAlpha)
		}
	}

	return &scaledPixelArray, scaledWidth, scaledHeight
}

// Xbr3x applies the 3x xBR scaling algorithm to a given pixel array.
//
// This method takes a slice of uint32 representing the original pixel data in ARGB format,
// along with the width and height of the original image. It then scales the image by a factor
// of 2 using the xBR algorithm, which is designed for pixel art upscaling.
//
// Parameters:
//
//	pixelArray *[]uint32: Pointer to a slice of uint32 representing the original pixel data in ARGB format.
//	width, height int: The width and height of the original image.
//	blendColors: A boolean flag indicating whether color blending should be used during the scaling process.
//	             When set to true, the algorithm blends colors to create smoother transitions.
//	scaleAlpha: A boolean flag indicating whether alpha values should be scaled, affecting the handling of transparency.
//
// Returns:
//
//	*[]uint32: Pointer to a slice of uint32 containing the scaled pixel data.
//	int: The width of the scaled image, which is twice the width of the original image.
//	int: The height of the scaled image, which is twice the height of the original image.
//
// The method uses the computeXbr2x function internally to process each pixel and write the results
// to a new slice that represents the scaled image. The returned slice contains the pixel data of the
// scaled image, and its dimensions are double those of the input image.
func (xbr *Xbr) Xbr3x(pixelArray *[]uint32, width, height int, blendColors, scaleAlpha bool) (*[]uint32, int, int) {
	scaledWidth := width * 3
	scaledHeight := height * 3
	scaledPixelArray := make([]uint32, scaledWidth*scaledHeight)

	for c := 0; c < width; c++ {
		for d := 0; d < height; d++ {
			xbr.computeXbr3x(pixelArray, c, d, width, height, &scaledPixelArray, c*3, d*3, scaledWidth, blendColors, scaleAlpha)
		}
	}

	return &scaledPixelArray, scaledWidth, scaledHeight
}

// Xbr4x applies the 4x xBR scaling algorithm to a given pixel array.
//
// This method takes a slice of uint32 representing the original pixel data in ARGB format,
// along with the width and height of the original image. It then scales the image by a factor
// of 2 using the xBR algorithm, which is designed for pixel art upscaling.
//
// Parameters:
//
//	pixelArray *[]uint32: Pointer to a slice of uint32 representing the original pixel data in ARGB format.
//	width, height int: The width and height of the original image.
//	blendColors: A boolean flag indicating whether color blending should be used during the scaling process.
//	             When set to true, the algorithm blends colors to create smoother transitions.
//	scaleAlpha: A boolean flag indicating whether alpha values should be scaled, affecting the handling of transparency.
//
// Returns:
//
//	*[]uint32: Pointer to a slice of uint32 containing the scaled pixel data.
//	int: The width of the scaled image, which is twice the width of the original image.
//	int: The height of the scaled image, which is twice the height of the original image.
//
// The method uses the computeXbr2x function internally to process each pixel and write the results
// to a new slice that represents the scaled image. The returned slice contains the pixel data of the
// scaled image, and its dimensions are double those of the input image.
func (xbr *Xbr) Xbr4x(pixelArray *[]uint32, width, height int, blendColors, scaleAlpha bool) (*[]uint32, int, int) {
	scaledWidth := width * 4
	scaledHeight := height * 4
	scaledPixelArray := make([]uint32, scaledWidth*scaledHeight)

	for c := 0; c < width; c++ {
		for d := 0; d < height; d++ {
			xbr.computeXbr4x(pixelArray, c, d, width, height, &scaledPixelArray, c*4, d*4, scaledWidth, blendColors, scaleAlpha)
		}
	}

	return &scaledPixelArray, scaledWidth, scaledHeight
}

// getYuv converts an ARGB pixel to YUV color space
func (xbr *Xbr) getYuv(p uint32) [3]float64 {
	r := float64(p & RedMask)
	g := float64((p & GreenMask) >> 8)
	b := float64((p & BlueMask) >> 16)

	y := r*0.299000 + g*0.587000 + b*0.114000
	u := r*-0.168736 + g*-0.331264 + b*0.500000
	v := r*0.500000 + g*-0.418688 + b*-0.081312

	return [3]float64{y, u, v}
}

// yuvDifference calculates the difference between two YUV colors
func (xbr *Xbr) yuvDifference(A, B uint32, scaleAlpha bool) int {
	alphaA := (A & AlphaMask) >> 24
	alphaB := (B & AlphaMask) >> 24

	if alphaA == 0 && alphaB == 0 {
		return 0
	}

	if !scaleAlpha && (alphaA < 255 || alphaB < 255) {
		return 1000000 // Arbitrary large value
	}

	if alphaA == 0 || alphaB == 0 {
		return 1000000 // Arbitrary large value
	}

	yuvA := xbr.getYuv(A)
	yuvB := xbr.getYuv(B)

	// Calculating YUV difference with threshold scaling
	diff := math.Abs(yuvA[0]-yuvB[0])*ThresholdY +
		math.Abs(yuvA[1]-yuvB[1])*ThresholdU +
		math.Abs(yuvA[2]-yuvB[2])*ThresholdV
	return int(diff)
}

func (xbr *Xbr) isEqual(A, B uint32, scaleAlpha bool) bool {
	alphaA := (A & AlphaMask) >> 24
	alphaB := (B & AlphaMask) >> 24

	if alphaA == 0 && alphaB == 0 {
		return true
	}

	if !scaleAlpha && (alphaA < 255 || alphaB < 255) {
		return false
	}

	if alphaA == 0 || alphaB == 0 {
		return false
	}

	yuvA := xbr.getYuv(A)
	yuvB := xbr.getYuv(B)

	if math.Abs(yuvA[0]-yuvB[0]) > ThresholdY {
		return false
	}
	if math.Abs(yuvA[1]-yuvB[1]) > ThresholdU {
		return false
	}
	if math.Abs(yuvA[2]-yuvB[2]) > ThresholdV {
		return false
	}

	return true
}

func (xbr *Xbr) pixelInterpolate(A, B uint32, q1, q2 float64) uint32 {
	alphaA := (A & AlphaMask) >> 24
	alphaB := (B & AlphaMask) >> 24

	var r, g, b, a uint32

	if alphaA == 0 {
		r = B & RedMask
		g = (B & GreenMask) >> 8
		b = (B & BlueMask) >> 16
	} else if alphaB == 0 {
		r = A & RedMask
		g = (A & GreenMask) >> 8
		b = (A & BlueMask) >> 16
	} else {
		r = uint32((q2*float64(B&RedMask) + q1*float64(A&RedMask)) / (q1 + q2))
		g = uint32((q2*float64((B&GreenMask)>>8) + q1*float64((A&GreenMask)>>8)) / (q1 + q2))
		b = uint32((q2*float64((B&BlueMask)>>16) + q1*float64((A&BlueMask)>>16)) / (q1 + q2))
	}
	a = uint32((q2*float64(alphaB) + q1*float64(alphaA)) / (q1 + q2))

	return r | (g << 8) | (b << 16) | (a << 24)
}

func (xbr *Xbr) getRelatedPoints(oriPixelView *[]uint32, oriX, oriY, oriW, oriH int) []uint32 {
	xm1 := max(oriX-1, 0)
	xm2 := max(oriX-2, 0)
	xp1 := min(oriX+1, oriW-1)
	xp2 := min(oriX+2, oriW-1)
	ym1 := max(oriY-1, 0)
	ym2 := max(oriY-2, 0)
	yp1 := min(oriY+1, oriH-1)
	yp2 := min(oriY+2, oriH-1)

	return []uint32{
		(*oriPixelView)[xm1+ym2*oriW],  /* a1 */
		(*oriPixelView)[oriX+ym2*oriW], /* b1 */
		(*oriPixelView)[xp1+ym2*oriW],  /* c1 */

		(*oriPixelView)[xm2+ym1*oriW],  /* a0 */
		(*oriPixelView)[xm1+ym1*oriW],  /* pa */
		(*oriPixelView)[oriX+ym1*oriW], /* pb */
		(*oriPixelView)[xp1+ym1*oriW],  /* pc */
		(*oriPixelView)[xp2+ym1*oriW],  /* c4 */

		(*oriPixelView)[xm2+oriY*oriW],  /* d0 */
		(*oriPixelView)[xm1+oriY*oriW],  /* pd */
		(*oriPixelView)[oriX+oriY*oriW], /* pe */
		(*oriPixelView)[xp1+oriY*oriW],  /* pf */
		(*oriPixelView)[xp2+oriY*oriW],  /* f4 */

		(*oriPixelView)[xm2+yp1*oriW],  /* g0 */
		(*oriPixelView)[xm1+yp1*oriW],  /* pg */
		(*oriPixelView)[oriX+yp1*oriW], /* ph */
		(*oriPixelView)[xp1+yp1*oriW],  /* pi */
		(*oriPixelView)[xp2+yp1*oriW],  /* i4 */

		(*oriPixelView)[xm1+yp2*oriW],  /* g5 */
		(*oriPixelView)[oriX+yp2*oriW], /* h5 */
		(*oriPixelView)[xp1+yp2*oriW],  /* i5 */
	}
}

func (xbr *Xbr) min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (xbr *Xbr) max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (xbr *Xbr) alphaBlend32W(dst, src uint32, blendColors bool) uint32 {
	if blendColors {
		return xbr.pixelInterpolate(dst, src, 7, 1)
	}
	return dst
}

func (xbr *Xbr) alphaBlend64W(dst, src uint32, blendColors bool) uint32 {
	if blendColors {
		return xbr.pixelInterpolate(dst, src, 3, 1)
	}
	return dst
}

func (xbr *Xbr) alphaBlend128W(dst, src uint32, blendColors bool) uint32 {
	if blendColors {
		return xbr.pixelInterpolate(dst, src, 1, 1)
	}
	return dst
}

func (xbr *Xbr) alphaBlend192W(dst, src uint32, blendColors bool) uint32 {
	if blendColors {
		return xbr.pixelInterpolate(dst, src, 1, 3)
	}
	return src
}

func (xbr *Xbr) alphaBlend224W(dst, src uint32, blendColors bool) uint32 {
	if blendColors {
		return xbr.pixelInterpolate(dst, src, 1, 7)
	}
	return src
}

func (xbr *Xbr) left2_2X(n3, n2, pixel uint32, blendColors bool) [2]uint32 {
	return [2]uint32{
		xbr.alphaBlend192W(n3, pixel, blendColors),
		xbr.alphaBlend64W(n2, pixel, blendColors),
	}
}

func (xbr *Xbr) up2_2X(n3, n1, pixel uint32, blendColors bool) [2]uint32 {
	return [2]uint32{
		xbr.alphaBlend192W(n3, pixel, blendColors),
		xbr.alphaBlend64W(n1, pixel, blendColors),
	}
}

func (xbr *Xbr) dia_2X(n3, pixel uint32, blendColors bool) uint32 {
	return xbr.alphaBlend128W(n3, pixel, blendColors)
}

func (xbr *Xbr) kernel2Xv5(pe, pi, ph, pf, pg, pc, pd, pb, f4, i4, h5, i5, n1, n2, n3 uint32, blendColors, scaleAlpha bool) (uint32, uint32, uint32) {
	ex := pe != ph && pe != pf
	if !ex {
		return n1, n2, n3
	}

	e := xbr.yuvDifference(pe, pc, scaleAlpha) + xbr.yuvDifference(pe, pg, scaleAlpha) + xbr.yuvDifference(pi, h5, scaleAlpha) + xbr.yuvDifference(pi, f4, scaleAlpha) + (xbr.yuvDifference(ph, pf, scaleAlpha) << 2)
	i := xbr.yuvDifference(ph, pd, scaleAlpha) + xbr.yuvDifference(ph, i5, scaleAlpha) + xbr.yuvDifference(pf, i4, scaleAlpha) + xbr.yuvDifference(pf, pb, scaleAlpha) + (xbr.yuvDifference(pe, pi, scaleAlpha) << 2)
	var px uint32
	if xbr.yuvDifference(pe, pf, scaleAlpha) <= xbr.yuvDifference(pe, ph, scaleAlpha) {
		px = pf
	} else {
		px = ph
	}

	if (e < i) && (!xbr.isEqual(pf, pb, scaleAlpha) && !xbr.isEqual(ph, pd, scaleAlpha) || xbr.isEqual(pe, pi, scaleAlpha) && (!xbr.isEqual(pf, i4, scaleAlpha) && !xbr.isEqual(ph, i5, scaleAlpha)) || xbr.isEqual(pe, pg, scaleAlpha) || xbr.isEqual(pe, pc, scaleAlpha)) {
		ke := xbr.yuvDifference(pf, pg, scaleAlpha)
		ki := xbr.yuvDifference(ph, pc, scaleAlpha)
		ex2 := pe != pc && pb != pc
		ex3 := pe != pg && pd != pg
		if ((ke<<1) <= ki && ex3) || (ke >= (ki<<1) && ex2) {
			if ((ke << 1) <= ki) && ex3 {
				leftOut := xbr.left2_2X(n3, n2, px, blendColors)
				n3 = leftOut[0]
				n2 = leftOut[1]
			}
			if (ke >= (ki << 1)) && ex2 {
				upOut := xbr.up2_2X(n3, n1, px, blendColors)
				n3 = upOut[0]
				n1 = upOut[1]
			}
		} else {
			n3 = xbr.dia_2X(n3, px, blendColors)
		}
	} else if e <= i {
		n3 = xbr.alphaBlend64W(n3, px, blendColors)
	}
	return n1, n2, n3
}

func (xbr *Xbr) computeXbr2x(oriPixelView *[]uint32, oriX, oriY, oriW, oriH int, dstPixelView *[]uint32, dstX, dstY, dstW int, blendColors, scaleAlpha bool) {
	relatedPoints := xbr.getRelatedPoints(oriPixelView, oriX, oriY, oriW, oriH)

	a1, b1, c1, a0, pa, pb, pc, c4, d0, pd, pe, pf, f4, g0, pg, ph, pi, i4, g5, h5, i5 := relatedPoints[0], relatedPoints[1], relatedPoints[2], relatedPoints[3], relatedPoints[4], relatedPoints[5], relatedPoints[6], relatedPoints[7], relatedPoints[8], relatedPoints[9], relatedPoints[10], relatedPoints[11], relatedPoints[12], relatedPoints[13], relatedPoints[14], relatedPoints[15], relatedPoints[16], relatedPoints[17], relatedPoints[18], relatedPoints[19], relatedPoints[20]

	e0, e1, e2, e3 := pe, pe, pe, pe

	e1, e2, e3 = xbr.kernel2Xv5(pe, pi, ph, pf, pg, pc, pd, pb, f4, i4, h5, i5, e1, e2, e3, blendColors, scaleAlpha)
	e0, e3, e1 = xbr.kernel2Xv5(pe, pc, pf, pb, pi, pa, ph, pd, b1, c1, f4, c4, e0, e3, e1, blendColors, scaleAlpha)
	e2, e1, e0 = xbr.kernel2Xv5(pe, pa, pb, pd, pc, pg, pf, ph, d0, a0, b1, a1, e2, e1, e0, blendColors, scaleAlpha)
	e3, e0, e2 = xbr.kernel2Xv5(pe, pg, pd, ph, pa, pi, pb, pf, h5, g5, d0, g0, e3, e0, e2, blendColors, scaleAlpha)

	(*dstPixelView)[dstX+dstY*dstW] = e0
	(*dstPixelView)[dstX+1+dstY*dstW] = e1
	(*dstPixelView)[dstX+(dstY+1)*dstW] = e2
	(*dstPixelView)[dstX+1+(dstY+1)*dstW] = e3
}

/*
 3x
*/

func (xbr *Xbr) leftUp2_3X(n7, n5, n6, n2, n8, pixel uint32, blendColors bool) [5]uint32 {
	blendedN7 := xbr.alphaBlend192W(n7, pixel, blendColors)
	blendedN6 := xbr.alphaBlend64W(n6, pixel, blendColors)
	return [5]uint32{blendedN7, blendedN7, blendedN6, blendedN6, pixel}
}

func (xbr *Xbr) left2_3X(n7, n5, n6, n8, pixel uint32, blendColors bool) [4]uint32 {
	return [4]uint32{
		xbr.alphaBlend192W(n7, pixel, blendColors),
		xbr.alphaBlend64W(n5, pixel, blendColors),
		xbr.alphaBlend64W(n6, pixel, blendColors),
		pixel,
	}
}

func (xbr *Xbr) up2_3X(n5, n7, n2, n8, pixel uint32, blendColors bool) [4]uint32 {
	return [4]uint32{
		xbr.alphaBlend192W(n5, pixel, blendColors),
		xbr.alphaBlend64W(n7, pixel, blendColors),
		xbr.alphaBlend64W(n2, pixel, blendColors),
		pixel,
	}
}

func (xbr *Xbr) dia_3X(n8, n5, n7, pixel uint32, blendColors bool) [3]uint32 {
	return [3]uint32{
		xbr.alphaBlend224W(n8, pixel, blendColors),
		xbr.alphaBlend32W(n5, pixel, blendColors),
		xbr.alphaBlend32W(n7, pixel, blendColors),
	}
}

func (xbr *Xbr) kernel3X(pe, pi, ph, pf, pg, pc, pd, pb, f4, i4, h5, i5, n2, n5, n6, n7, n8 uint32, blendColors, scaleAlpha bool) (uint32, uint32, uint32, uint32, uint32) {
	ex := pe != ph && pe != pf
	if !ex {
		return n2, n5, n6, n7, n8
	}

	e := xbr.yuvDifference(pe, pc, scaleAlpha) + xbr.yuvDifference(pe, pg, scaleAlpha) + xbr.yuvDifference(pi, h5, scaleAlpha) + xbr.yuvDifference(pi, f4, scaleAlpha) + (xbr.yuvDifference(ph, pf, scaleAlpha) << 2)
	i := xbr.yuvDifference(ph, pd, scaleAlpha) + xbr.yuvDifference(ph, i5, scaleAlpha) + xbr.yuvDifference(pf, i4, scaleAlpha) + xbr.yuvDifference(pf, pb, scaleAlpha) + (xbr.yuvDifference(pe, pi, scaleAlpha) << 2)

	var state bool

	if xbr.use3XOriginalImplementation {
		state = (e < i) && (!xbr.isEqual(pf, pb, scaleAlpha) && !xbr.isEqual(ph, pd, scaleAlpha) || xbr.isEqual(pe, pi, scaleAlpha) && (!xbr.isEqual(pf, i4, scaleAlpha) && !xbr.isEqual(ph, i5, scaleAlpha)) || xbr.isEqual(pe, pg, scaleAlpha) || xbr.isEqual(pe, pc, scaleAlpha))
	} else {
		state = (e < i) && (!xbr.isEqual(pf, pb, scaleAlpha) && !xbr.isEqual(pf, pc, scaleAlpha) || !xbr.isEqual(ph, pd, scaleAlpha) && !xbr.isEqual(ph, pg, scaleAlpha) || xbr.isEqual(pe, pi, scaleAlpha) && (!xbr.isEqual(pf, f4, scaleAlpha) && !xbr.isEqual(pf, i4, scaleAlpha) || !xbr.isEqual(ph, h5, scaleAlpha) && !xbr.isEqual(ph, i5, scaleAlpha)) || xbr.isEqual(pe, pg, scaleAlpha) || xbr.isEqual(pe, pc, scaleAlpha))
	}

	if state {
		ke := xbr.yuvDifference(pf, pg, scaleAlpha)
		ki := xbr.yuvDifference(ph, pc, scaleAlpha)
		ex2 := pe != pc && pb != pc
		ex3 := pe != pg && pd != pg
		var px uint32
		if xbr.yuvDifference(pe, pf, scaleAlpha) <= xbr.yuvDifference(pe, ph, scaleAlpha) {
			px = pf
		} else {
			px = ph
		}

		if ((ke<<1) <= ki && ex3) || (ke >= (ki<<1) && ex2) {
			output := xbr.leftUp2_3X(n7, n5, n6, n2, n8, px, blendColors)
			n7, n5, n6, n2, n8 = output[0], output[1], output[2], output[3], output[4]
		} else if ((ke << 1) <= ki) && ex3 {
			output := xbr.left2_3X(n7, n5, n6, n8, px, blendColors)
			n7, n5, n6, n8 = output[0], output[1], output[2], output[3]
		} else if (ke >= (ki << 1)) && ex2 {
			output := xbr.up2_3X(n5, n7, n2, n8, px, blendColors)
			n5, n7, n2, n8 = output[0], output[1], output[2], output[3]
		} else {
			output := xbr.dia_3X(n8, n5, n7, px, blendColors)
			n8, n5, n7 = output[0], output[1], output[2]
		}
	} else if e <= i {
		n8 = xbr.alphaBlend128W(n8, func() uint32 {
			if xbr.yuvDifference(pe, pf, scaleAlpha) <= xbr.yuvDifference(pe, ph, scaleAlpha) {
				return pf
			}
			return ph
		}(), blendColors)
	}
	return n2, n5, n6, n7, n8
}

func (xbr *Xbr) computeXbr3x(oriPixelView *[]uint32, oriX, oriY, oriW, oriH int, dstPixelView *[]uint32, dstX, dstY, dstW int, blendColors, scaleAlpha bool) {
	relatedPoints := xbr.getRelatedPoints(oriPixelView, oriX, oriY, oriW, oriH)
	a1, b1, c1, a0, pa, pb, pc, c4, d0, pd, pe, pf, f4, g0, pg, ph, pi, i4, g5, h5, i5 := relatedPoints[0], relatedPoints[1], relatedPoints[2], relatedPoints[3], relatedPoints[4], relatedPoints[5], relatedPoints[6], relatedPoints[7], relatedPoints[8], relatedPoints[9], relatedPoints[10], relatedPoints[11], relatedPoints[12], relatedPoints[13], relatedPoints[14], relatedPoints[15], relatedPoints[16], relatedPoints[17], relatedPoints[18], relatedPoints[19], relatedPoints[20]

	e0, e1, e2, e3, e4, e5, e6, e7, e8 := pe, pe, pe, pe, pe, pe, pe, pe, pe

	e2, e5, e6, e7, e8 = xbr.kernel3X(pe, pi, ph, pf, pg, pc, pd, pb, f4, i4, h5, i5, e2, e5, e6, e7, e8, blendColors, scaleAlpha)
	e0, e1, e8, e5, e2 = xbr.kernel3X(pe, pc, pf, pb, pi, pa, ph, pd, b1, c1, f4, c4, e0, e1, e8, e5, e2, blendColors, scaleAlpha)
	e6, e3, e2, e1, e0 = xbr.kernel3X(pe, pa, pb, pd, pc, pg, pf, ph, d0, a0, b1, a1, e6, e3, e2, e1, e0, blendColors, scaleAlpha)
	e8, e7, e0, e3, e6 = xbr.kernel3X(pe, pg, pd, ph, pa, pi, pb, pf, h5, g5, d0, g0, e8, e7, e0, e3, e6, blendColors, scaleAlpha)

	(*dstPixelView)[dstX+dstY*dstW] = e0
	(*dstPixelView)[dstX+1+dstY*dstW] = e1
	(*dstPixelView)[dstX+2+dstY*dstW] = e2
	(*dstPixelView)[dstX+(dstY+1)*dstW] = e3
	(*dstPixelView)[dstX+1+(dstY+1)*dstW] = e4
	(*dstPixelView)[dstX+2+(dstY+1)*dstW] = e5
	(*dstPixelView)[dstX+(dstY+2)*dstW] = e6
	(*dstPixelView)[dstX+1+(dstY+2)*dstW] = e7
	(*dstPixelView)[dstX+2+(dstY+2)*dstW] = e8
}

/*
	4x
*/

func (xbr *Xbr) leftUp2(n15, n14, n11, n13, n12, n10, n7, n3, pixel uint32, blendColors bool) [8]uint32 {
	blendedN13 := xbr.alphaBlend192W(n13, pixel, blendColors)
	blendedN12 := xbr.alphaBlend64W(n12, pixel, blendColors)

	return [8]uint32{pixel, pixel, pixel, blendedN12, blendedN12, blendedN12, blendedN13, n3}
}

func (xbr *Xbr) left2(n15, n14, n11, n13, n12, n10, pixel uint32, blendColors bool) [6]uint32 {
	return [6]uint32{
		pixel,
		pixel,
		xbr.alphaBlend192W(n11, pixel, blendColors),
		xbr.alphaBlend192W(n13, pixel, blendColors),
		xbr.alphaBlend64W(n12, pixel, blendColors),
		xbr.alphaBlend64W(n10, pixel, blendColors),
	}
}

func (xbr *Xbr) up2(n15, n14, n11, n3, n7, n10, pixel uint32, blendColors bool) [6]uint32 {
	return [6]uint32{
		pixel,
		xbr.alphaBlend192W(n14, pixel, blendColors),
		pixel,
		xbr.alphaBlend64W(n3, pixel, blendColors),
		xbr.alphaBlend192W(n7, pixel, blendColors),
		xbr.alphaBlend64W(n10, pixel, blendColors),
	}
}

func (xbr *Xbr) dia(n15, n14, n11, pixel uint32, blendColors bool) [3]uint32 {
	return [3]uint32{
		pixel,
		xbr.alphaBlend128W(n14, pixel, blendColors),
		xbr.alphaBlend128W(n11, pixel, blendColors),
	}
}

func (xbr *Xbr) kernel4Xv2(pe, pi, ph, pf, pg, pc, pd, pb, f4, i4, h5, i5, n15, n14, n11, n3, n7, n10, n13, n12 uint32, blendColors, scaleAlpha bool) (uint32, uint32, uint32, uint32, uint32, uint32, uint32, uint32) {
	ex := pe != ph && pe != pf
	if !ex {
		return n15, n14, n11, n3, n7, n10, n13, n12
	}

	e := xbr.yuvDifference(pe, pc, scaleAlpha) + xbr.yuvDifference(pe, pg, scaleAlpha) + xbr.yuvDifference(pi, h5, scaleAlpha) + xbr.yuvDifference(pi, f4, scaleAlpha) + (xbr.yuvDifference(ph, pf, scaleAlpha) << 2)
	i := xbr.yuvDifference(ph, pd, scaleAlpha) + xbr.yuvDifference(ph, i5, scaleAlpha) + xbr.yuvDifference(pf, i4, scaleAlpha) + xbr.yuvDifference(pf, pb, scaleAlpha) + (xbr.yuvDifference(pe, pi, scaleAlpha) << 2)
	var px uint32
	if xbr.yuvDifference(pe, pf, scaleAlpha) <= xbr.yuvDifference(pe, ph, scaleAlpha) {
		px = pf
	} else {
		px = ph
	}

	if (e < i) && (!xbr.isEqual(pf, pb, scaleAlpha) && !xbr.isEqual(ph, pd, scaleAlpha) || xbr.isEqual(pe, pi, scaleAlpha) && (!xbr.isEqual(pf, i4, scaleAlpha) && !xbr.isEqual(ph, i5, scaleAlpha)) || xbr.isEqual(pe, pg, scaleAlpha) || xbr.isEqual(pe, pc, scaleAlpha)) {
		ke := xbr.yuvDifference(pf, pg, scaleAlpha)
		ki := xbr.yuvDifference(ph, pc, scaleAlpha)
		ex2 := pe != pc && pb != pc
		ex3 := pe != pg && pd != pg
		if ((ke<<1) <= ki && ex3) || (ke >= (ki<<1) && ex2) {
			if ((ke << 1) <= ki) && ex3 {
				output := xbr.left2(n15, n14, n11, n13, n12, n10, px, blendColors)
				n15, n14, n11, n13, n12, n10 = output[0], output[1], output[2], output[3], output[4], output[5]
			}
			if (ke >= (ki << 1)) && ex2 {
				output := xbr.up2(n15, n14, n11, n3, n7, n10, px, blendColors)
				n15, n14, n11, n3, n7, n10 = output[0], output[1], output[2], output[3], output[4], output[5]
			}
		} else {
			output := xbr.dia(n15, n14, n11, px, blendColors)
			n15, n14, n11 = output[0], output[1], output[2]
		}
	} else if e <= i {
		n15 = xbr.alphaBlend128W(n15, px, blendColors)
	}

	return n15, n14, n11, n3, n7, n10, n13, n12
}

func (xbr *Xbr) computeXbr4x(oriPixelView *[]uint32, oriX, oriY, oriW, oriH int, dstPixelView *[]uint32, dstX, dstY, dstW int, blendColors, scaleAlpha bool) {
	relatedPoints := xbr.getRelatedPoints(oriPixelView, oriX, oriY, oriW, oriH)
	a1, b1, c1, a0, pa, pb, pc, c4, d0, pd, pe, pf, f4, g0, pg, ph, pi, i4, g5, h5, i5 := relatedPoints[0], relatedPoints[1], relatedPoints[2], relatedPoints[3], relatedPoints[4], relatedPoints[5], relatedPoints[6], relatedPoints[7], relatedPoints[8], relatedPoints[9], relatedPoints[10], relatedPoints[11], relatedPoints[12], relatedPoints[13], relatedPoints[14], relatedPoints[15], relatedPoints[16], relatedPoints[17], relatedPoints[18], relatedPoints[19], relatedPoints[20]

	e0, e1, e2, e3, e4, e5, e6, e7, e8, e9, ea, eb, ec, ed, ee, ef := pe, pe, pe, pe, pe, pe, pe, pe, pe, pe, pe, pe, pe, pe, pe, pe

	ef, ee, eb, e3, e7, ea, ed, ec = xbr.kernel4Xv2(pe, pi, ph, pf, pg, pc, pd, pb, f4, i4, h5, i5, ef, ee, eb, e3, e7, ea, ed, ec, blendColors, scaleAlpha)
	e3, e7, e2, e0, e1, e6, eb, ef = xbr.kernel4Xv2(pe, pc, pf, pb, pi, pa, ph, pd, b1, c1, f4, c4, e3, e7, e2, e0, e1, e6, eb, ef, blendColors, scaleAlpha)
	e0, e1, e4, ec, e8, e5, e2, e3 = xbr.kernel4Xv2(pe, pa, pb, pd, pc, pg, pf, ph, d0, a0, b1, a1, e0, e1, e4, ec, e8, e5, e2, e3, blendColors, scaleAlpha)
	ec, e8, ed, ef, ee, e9, e4, e0 = xbr.kernel4Xv2(pe, pg, pd, ph, pa, pi, pb, pf, h5, g5, d0, g0, ec, e8, ed, ef, ee, e9, e4, e0, blendColors, scaleAlpha)

	(*dstPixelView)[dstX+dstY*dstW] = e0
	(*dstPixelView)[dstX+1+dstY*dstW] = e1
	(*dstPixelView)[dstX+2+dstY*dstW] = e2
	(*dstPixelView)[dstX+3+dstY*dstW] = e3
	(*dstPixelView)[dstX+(dstY+1)*dstW] = e4
	(*dstPixelView)[dstX+1+(dstY+1)*dstW] = e5
	(*dstPixelView)[dstX+2+(dstY+1)*dstW] = e6
	(*dstPixelView)[dstX+3+(dstY+1)*dstW] = e7
	(*dstPixelView)[dstX+(dstY+2)*dstW] = e8
	(*dstPixelView)[dstX+1+(dstY+2)*dstW] = e9
	(*dstPixelView)[dstX+2+(dstY+2)*dstW] = ea
	(*dstPixelView)[dstX+3+(dstY+2)*dstW] = eb
	(*dstPixelView)[dstX+(dstY+3)*dstW] = ec
	(*dstPixelView)[dstX+1+(dstY+3)*dstW] = ed
	(*dstPixelView)[dstX+2+(dstY+3)*dstW] = ee
	(*dstPixelView)[dstX+3+(dstY+3)*dstW] = ef
}
