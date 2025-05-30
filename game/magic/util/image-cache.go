package util

import (
    "log"
    "reflect"
    "fmt"
    "strings"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    // "github.com/kazzmir/master-of-magic/lib/xbr"
    "github.com/kazzmir/master-of-magic/game/magic/shaders"
    "github.com/hajimehoshi/ebiten/v2"
)

type ImageTransformFunc func(*image.Paletted) image.Image
type ImageTransformGenericFunc func(image.Image) image.Image

type Scaler interface {
    ApplyScale(image.Image) image.Image
}

type ImageCache struct {
    LbxCache *lbx.LbxCache
    // FIXME: have some limit on the number of entries, and remove old ones LRU-style
    Cache map[string][]*ebiten.Image

    ShaderCache map[shaders.Shader]*ebiten.Shader

    /*
    Scaler data.ScaleAlgorithm
    ScaleAmount int
    */
}

func MakeImageCache(lbxCache *lbx.LbxCache) ImageCache {
    return ImageCache{
        LbxCache: lbxCache,
        Cache:    make(map[string][]*ebiten.Image),
        ShaderCache: make(map[shaders.Shader]*ebiten.Shader),
        /*
        Scaler: data.ScreenScaleAlgorithm,
        ScaleAmount: 1,
        */
        // ScaleAmount: data.ScreenScale,
    }
}

func ComposeImageTransform(transform1 ImageTransformFunc, transform2 ImageTransformGenericFunc) ImageTransformFunc {
    return func (img *image.Paletted) image.Image {
        return transform2(transform1(img))
    }
}

func AutoCrop(img *image.Paletted) image.Image {
    return AutoCropGeneric(img)
}

// create a new image from a portion of the original image
// the new image has the same width/height as the rect. the source pixels come from img at the bounds of the rect
func copyImagePortion(img *image.Paletted, rect image.Rectangle) *image.Paletted {
    out := image.NewPaletted(image.Rect(0, 0, rect.Bounds().Dx(), rect.Bounds().Dy()), img.Palette)

    for y := rect.Min.Y; y < rect.Max.Y; y++ {
        for x := rect.Min.X; x < rect.Max.X; x++ {
            out.SetColorIndex(x - rect.Min.X, y - rect.Min.Y, img.ColorIndexAt(x, y))
        }
    }

    return out
}

// remove all alpha-0 pixels from the border of the image
func AutoCropGeneric(img image.Image) image.Image {
    bounds := img.Bounds()
    minX := bounds.Max.X
    minY := bounds.Max.Y
    maxX := bounds.Min.X
    maxY := bounds.Min.Y

    for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
        for x := bounds.Min.X; x < bounds.Max.X; x++ {
            _, _, _, a := img.At(x, y).RGBA()
            if a > 0 {
                if x < minX {
                    minX = x
                }
                if y < minY {
                    minY = y
                }
                if x > maxX {
                    maxX = x
                }
                if y > maxY {
                    maxY = y
                }
            }
        }
    }

    // log.Printf("Auto crop on %v", reflect.TypeOf(img))
    // log.Printf("auto crop: %d %d %d %d", minX, minY, maxX, maxY)

    switch img.(type) {
        case *image.Paletted:
            // note that the resulting image might not have its upper left hand pixel at 0,0. Instead it will be at
            // bounds.Min.X, bounds.Min.Y

            return img.(*image.Paletted).SubImage(image.Rect(minX, minY, maxX+1, maxY+1))

            // return copyImagePortion(img.(*image.Paletted), image.Rect(minX, minY, maxX+1, maxY+1))
        default:
            log.Printf("Auto crop not implemented for %v", reflect.TypeOf(img))
    }

    return img
}

func (cache *ImageCache) GetShader(shader shaders.Shader) (*ebiten.Shader, error) {
    out, ok := cache.ShaderCache[shader]
    if ok {
        return out, nil
    }

    type ShaderLoader func() (*ebiten.Shader, error)
    loaderMap := map[shaders.Shader]ShaderLoader{
        shaders.ShaderEdgeGlow: shaders.LoadEdgeGlowShader,
        shaders.ShaderWarp: shaders.LoadWarpShader,
        shaders.ShaderDropShadow: shaders.LoadDropShadowShader,
        shaders.ShaderOutline: shaders.LoadOutlineShader,
    }

    var err error
    loader, ok := loaderMap[shader]
    if !ok {
        return nil, fmt.Errorf("unknown shader: %v", shader)
    }

    out, err = loader()
    if err != nil {
        return nil, err
    }

    if out == nil {
        return nil, fmt.Errorf("unknown error loading shader: %v", shader)
    }

    cache.ShaderCache[shader] = out
    return out, nil
}

/* remove all entries from the cache */
func (cache *ImageCache) Clear(){
    cache.Cache = make(map[string][]*ebiten.Image)
    cache.ShaderCache = make(map[shaders.Shader]*ebiten.Shader)
}

func (cache *ImageCache) GetImagesTransform(lbxPath string, index int, extra string, transform ImageTransformFunc) ([]*ebiten.Image, error) {
    lbxPath = strings.ToLower(lbxPath)
    key := fmt.Sprintf("%s:%s:%d", lbxPath, extra, index)

    if images, ok := cache.Cache[key]; ok {
        return images, nil
    }

    lbxFile, err := cache.LbxCache.GetLbxFile(lbxPath)
    if err != nil {
        return nil, err
    }

    // FIXME: cache this for the given lbxFile and lbxPath
    customPaletteMap, err := lbx.GetPaletteOverrideMap(cache.LbxCache, lbxFile, lbxPath)
    if err != nil {
        return nil, err
    }

    palette := customPaletteMap[index]
    if palette == nil {
        // -1 is a default palette for all images
        palette = customPaletteMap[-1]
    }

    sprites, err := lbxFile.ReadImagesWithPalette(index, palette, palette != nil)
    if err != nil {
        return nil, err
    }

    var out []*ebiten.Image
    for i := 0; i < len(sprites); i++ {
        out = append(out, ebiten.NewImageFromImage(transform(sprites[i])))
    }

    cache.Cache[key] = out

    return out, nil
}

func Scale2x(input image.Image, smooth bool) image.Image {
    bounds := input.Bounds()
    scaledImage := image.NewRGBA(image.Rect(0, 0, 2 * bounds.Dx(), 2 * bounds.Dy()))

    getColor := func(x, y int) color.Color {
        if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
            return color.RGBA{}
        }

        return input.At(x, y)
    }

    for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
        for x := bounds.Min.X; x < bounds.Max.X; x++ {
            B := getColor(x, y-1)
            D := getColor(x-1, y)
            E := getColor(x, y)
            F := getColor(x+1, y)
            H := getColor(x, y+1)

            E0, E1, E2, E3 := E, E, E, E
            if smooth && B != H && D != F {
                if D == B {
                    E0 = D
                }
                if B == F {
                    E1 = F
                }
                if D == H {
                    E2 = D
                }
                if H == F {
                    E3 = F
                }
            }

            x1 := x - bounds.Min.X
            y1 := y - bounds.Min.Y

            scaledImage.Set(x1*2+0, y1*2+0, E0)
            scaledImage.Set(x1*2+1, y1*2+0, E1)
            scaledImage.Set(x1*2+0, y1*2+1, E2)
            scaledImage.Set(x1*2+1, y1*2+1, E3)
        }
    }

    return scaledImage
}

func Scale3x(input image.Image, smooth bool) image.Image {
    bounds := input.Bounds()
    scaledImage := image.NewRGBA(image.Rect(0, 0, 3 * bounds.Dx(), 3 * bounds.Dy()))

    getColor := func(x, y int) color.Color {
        if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
            return color.RGBA{}
        }

        return input.At(x, y)
    }

    for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
        for x := bounds.Min.X; x < bounds.Max.X; x++ {
            B := getColor(x, y-1)
            C := getColor(x+1, y-1)
            D := getColor(x-1, y)
            E := getColor(x, y)
            F := getColor(x+1, y)
            G := getColor(x-1, y+1)
            H := getColor(x, y+1)
            I := getColor(x+1, y+1)

            E0, E1, E2, E3, E4, E5, E6, E7, E8 := E, E, E, E, E, E, E, E, E
            if smooth && B != H && D != F {
                if D == B {
                    E0 = D
                }
                if B == F {
                    E1 = B
                }
                if D == H {
                    E2 = F
                }
                if H == F {
                    E3 = D
                }
                E4 = E
                if (B == F && E != I) || (H == F && E != C) {
                    E5 = F
                }

                if D == H {
                    E6 = D
                }

                if (D == H && E != I) || (H == F && E != G) {
                    E7 = H
                }

                if H == F {
                    E8 = F
                }
            }

            x1 := x - bounds.Min.X
            y1 := y - bounds.Min.Y

            scaledImage.Set(x1*3+0, y1*3+0, E0)
            scaledImage.Set(x1*3+1, y1*3+0, E1)
            scaledImage.Set(x1*3+2, y1*3+0, E2)

            scaledImage.Set(x1*3+0, y1*3+1, E3)
            scaledImage.Set(x1*3+1, y1*3+1, E4)
            scaledImage.Set(x1*3+2, y1*3+1, E5)

            scaledImage.Set(x1*3+0, y1*3+2, E6)
            scaledImage.Set(x1*3+1, y1*3+2, E7)
            scaledImage.Set(x1*3+2, y1*3+2, E8)
        }
    }

    return scaledImage
}

func (cache *ImageCache) ApplyScale(input image.Image) image.Image {
    /*
    if cache.ScaleAmount == 1 {
        return input
    }

    switch cache.Scaler {
        case data.ScaleAlgorithmNormal:
            switch cache.ScaleAmount {
                case 2: return Scale2x(input, false)
                case 3: return Scale3x(input, false)
                case 4: return Scale2x(Scale2x(input, false), false)
                default: return input
            }
        case data.ScaleAlgorithmScale:
            switch cache.ScaleAmount {
                case 2: return Scale2x(input, true)
                case 3: return Scale3x(input, true)
                case 4: return Scale2x(Scale2x(input, true), true)
                default: return input
            }
        case data.ScaleAlgorithmXbr: return xbr.ScaleImage(input, cache.ScaleAmount)
    }
    */

    return input
}

func (cache *ImageCache) GetImages(lbxPath string, index int) ([]*ebiten.Image, error) {
    return cache.GetImagesTransform(lbxPath, index, "_", func (img *image.Paletted) image.Image {
        return img
    })
}

func (cache *ImageCache) GetImageTransform(lbxFile string, spriteIndex int, animationIndex int, extra string, transform ImageTransformFunc) (*ebiten.Image, error) {
    images, err := cache.GetImagesTransform(lbxFile, spriteIndex, extra, transform)
    if err != nil {
        return nil, err
    }

    if animationIndex < len(images) {
        return images[animationIndex], nil
    }

    return nil, fmt.Errorf("invalid animation index: %d for %v:%v", animationIndex, lbxFile, spriteIndex)
}

func (cache *ImageCache) GetImage(lbxFile string, spriteIndex int, animationIndex int) (*ebiten.Image, error) {
    return cache.GetImageTransform(lbxFile, spriteIndex, animationIndex, "_", func (img *image.Paletted) image.Image {
        return img
    })
}
