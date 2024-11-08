package util

import (
    "fmt"
    "strings"
    "image"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/shaders"
    "github.com/hajimehoshi/ebiten/v2"
)

type ImageTransformFunc func(*image.Paletted) image.Image

type ImageCache struct {
    LbxCache *lbx.LbxCache
    // FIXME: have some limit on the number of entries, and remove old ones LRU-style
    Cache map[string][]*ebiten.Image

    ShaderCache map[shaders.Shader]*ebiten.Shader
}

func MakeImageCache(lbxCache *lbx.LbxCache) ImageCache {
    return ImageCache{
        LbxCache: lbxCache,
        Cache:    make(map[string][]*ebiten.Image),
        ShaderCache: make(map[shaders.Shader]*ebiten.Shader),
    }
}

// remove all alpha-0 pixels from the border of the image
func AutoCrop(img *image.Paletted) image.Image {
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

    return img.SubImage(image.Rect(minX, minY, maxX, maxY))
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
        out = append(out, ebiten.NewImageFromImage(transform(sprites[i])))
    }

    cache.Cache[key] = out

    return out, nil
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
