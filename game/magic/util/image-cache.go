package util

import (
    "fmt"
    "strings"
    "image"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/hajimehoshi/ebiten/v2"
)

type ImageCache struct {
    LbxCache *lbx.LbxCache
    // FIXME: have some limit on the number of entries, and remove old ones LRU-style
    Cache map[string][]*ebiten.Image
}

func MakeImageCache(lbxCache *lbx.LbxCache) ImageCache {
    return ImageCache{
        LbxCache: lbxCache,
        Cache:    make(map[string][]*ebiten.Image),
    }
}

// remove all alpha-0 pixels from the border of the image
func AutoCrop(img image.Image) image.Image {
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

    paletted, ok := img.(*image.Paletted)
    if ok {
        return paletted.SubImage(image.Rect(minX, minY, maxX, maxY))
    }

    return img
}


/* remove all entries from the cache */
func (cache *ImageCache) Clear(){
    cache.Cache = make(map[string][]*ebiten.Image)
}

func (cache *ImageCache) GetImagesTransform(lbxPath string, index int, transform func(image.Image) image.Image) ([]*ebiten.Image, error) {
    lbxPath = strings.ToLower(lbxPath)
    key := fmt.Sprintf("%s:%d", lbxPath, index)

    if images, ok := cache.Cache[key]; ok {
        return images, nil
    }

    lbxFile, err := cache.LbxCache.GetLbxFile(lbxPath)
    if err != nil {
        return nil, err
    }

    customPalette, err := lbx.GetPaletteOverrideMap(lbxFile, lbxPath)
    if err != nil {
        return nil, err
    }

    sprites, err := lbxFile.ReadImagesWithPalette(index, customPalette[index])
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
    return cache.GetImagesTransform(lbxPath, index, func (img image.Image) image.Image {
        return img
    })
}

func (cache *ImageCache) GetImageTransform(lbxFile string, spriteIndex int, animationIndex int, transform func(image.Image) image.Image) (*ebiten.Image, error) {
    images, err := cache.GetImagesTransform(lbxFile, spriteIndex, transform)
    if err != nil {
        return nil, err
    }

    if animationIndex < len(images) {
        return images[animationIndex], nil
    }

    return nil, fmt.Errorf("invalid animation index: %d for %v:%v", animationIndex, lbxFile, spriteIndex)
}

func (cache *ImageCache) GetImage(lbxFile string, spriteIndex int, animationIndex int) (*ebiten.Image, error) {
    return cache.GetImageTransform(lbxFile, spriteIndex, animationIndex, func (img image.Image) image.Image {
        return img
    })
}
