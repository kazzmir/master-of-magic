package util

import (
    "log"
    "reflect"
    "fmt"
    "strings"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/xbr"
    "github.com/kazzmir/master-of-magic/game/magic/shaders"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/hajimehoshi/ebiten/v2"
)

type ImageTransformFunc func(*image.Paletted) image.Image
type ImageTransformGenericFunc func(image.Image) image.Image

type ScaleAlgorithm int

const (
    ScaleAlgorithmLinear ScaleAlgorithm = iota
    ScaleAlgorithmXbr
)

type ImageCache struct {
    LbxCache *lbx.LbxCache
    // FIXME: have some limit on the number of entries, and remove old ones LRU-style
    Cache map[string][]*ebiten.Image

    ShaderCache map[shaders.Shader]*ebiten.Shader

    Scaler ScaleAlgorithm
    ScaleAmount int
}

func MakeImageCache(lbxCache *lbx.LbxCache) ImageCache {
    return ImageCache{
        LbxCache: lbxCache,
        Cache:    make(map[string][]*ebiten.Image),
        ShaderCache: make(map[shaders.Shader]*ebiten.Shader),
        Scaler: ScaleAlgorithmLinear,
        ScaleAmount: data.ScreenScale,
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

    switch img.(type) {
        case *image.Paletted:
            return img.(*image.Paletted).SubImage(image.Rect(minX, minY, maxX, maxY))
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

    var err error
    switch shader {
        case shaders.ShaderEdgeGlow:
            out, err = shaders.LoadEdgeGlowShader()
            if err != nil {
                return nil, err
            }
    }

    if out == nil {
        return nil, fmt.Errorf("unknown shader: %v", shader)
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

    sprites, err := lbxFile.ReadImagesWithPalette(index, palette, false)
    if err != nil {
        return nil, err
    }

    var out []*ebiten.Image
    for i := 0; i < len(sprites); i++ {
        out = append(out, ebiten.NewImageFromImage(cache.ApplyScale(transform(sprites[i]))))
    }

    cache.Cache[key] = out

    return out, nil
}

func scale2x(input image.Image) image.Image {
    bounds := input.Bounds()
    scaledImage := image.NewRGBA(image.Rect(0, 0, 2 * bounds.Dx(), 2 * bounds.Dy()))
    smooth := true // use Scale2x by Andrea Mazzoleni

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

            scaledImage.Set(x*2, y*2, E0)
            scaledImage.Set(x*2+1, y*2, E1)
            scaledImage.Set(x*2, y*2+1, E2)
            scaledImage.Set(x*2+1, y*2+1, E3)
        }
    }

    return scaledImage
}

func scale3x(input image.Image) image.Image {
    bounds := input.Bounds()
    scaledImage := image.NewRGBA(image.Rect(0, 0, 3 * bounds.Dx(), 3 * bounds.Dy()))
    smooth := true // use Scale2x by Andrea Mazzoleni

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

            scaledImage.Set(x*3+0, y*3+0, E0)
            scaledImage.Set(x*3+1, y*3+0, E1)
            scaledImage.Set(x*3+2, y*3+0, E2)

            scaledImage.Set(x*3+0, y*3+1, E3)
            scaledImage.Set(x*3+1, y*3+1, E4)
            scaledImage.Set(x*3+2, y*3+1, E5)

            scaledImage.Set(x*3+0, y*3+2, E6)
            scaledImage.Set(x*3+1, y*3+2, E7)
            scaledImage.Set(x*3+2, y*3+2, E8)
        }
    }

    return scaledImage
}

func (cache *ImageCache) ApplyScale(input image.Image) image.Image {
    if cache.ScaleAmount == 1 {
        return input
    }

    switch cache.Scaler {
        case ScaleAlgorithmLinear:
            switch cache.ScaleAmount {
                case 2: return scale2x(input)
                case 3: return scale3x(input)
                case 4: return scale2x(scale2x(input))
                default: return input
            }
        case ScaleAlgorithmXbr: return xbr.ScaleImage(input, cache.ScaleAmount)
    }

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
