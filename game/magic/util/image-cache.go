package game

import (
    "fmt"
    "strings"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/hajimehoshi/ebiten/v2"
)

type ImageCache struct {
    LbxCache *lbx.LbxCache
    // FIXME: have some limit on the number of entries, and remove old ones LRU-style
    Cache map[string][]*ebiten.Image
}

func (cache *ImageCache) GetImages(lbxPath string, index int) ([]*ebiten.Image, error) {
    lbxPath = strings.ToLower(lbxPath)
    key := fmt.Sprintf("%s:%d", lbxPath, index)

    if images, ok := cache.Cache[key]; ok {
        return images, nil
    }

    lbxFile, err := cache.LbxCache.GetLbxFile(lbxPath)
    if err != nil {
        return nil, err
    }

    sprites, err := lbxFile.ReadImages(index)
    if err != nil {
        return nil, err
    }

    var out []*ebiten.Image
    for i := 0; i < len(sprites); i++ {
        out = append(out, ebiten.NewImageFromImage(sprites[i]))
    }

    cache.Cache[key] = out

    return out, nil
}

func (cache *ImageCache) GetImage(lbxFile string, spriteIndex int, animationIndex int) (*ebiten.Image, error) {
    images, err := cache.GetImages(lbxFile, spriteIndex)
    if err != nil {
        return nil, err
    }

    if animationIndex < len(images) {
        return images[animationIndex], nil
    }

    return nil, fmt.Errorf("invalid animation index: %d for %v:%v", animationIndex, lbxFile, spriteIndex)
}
