package audio

import (
    "github.com/kazzmir/master-of-magic/lib/lbx"
    audiolib "github.com/hajimehoshi/ebiten/v2/audio"
)

const CombatLbxPath = "cmbtsnd.lbx"
const NewSoundLbxPath = "newsound.lbx"
const SoundLbxPath = "soundfx.lbx"

type cacheKey struct {
    lbxPath string
    index int
}

type AudioCache struct {
    data map[cacheKey]MakePlayerFunc
    cache *lbx.LbxCache
}

func MakeAudioCache(cache *lbx.LbxCache) *AudioCache {
    return &AudioCache{
        data: make(map[cacheKey]MakePlayerFunc),
        cache: cache,
    }
}

func (cache *AudioCache) getSound(lbxPath string, index int) (*audiolib.Player, error) {
    key := cacheKey{lbxPath: lbxPath, index: index}
    if maker, ok := cache.data[key]; ok {
        return maker(), nil
    }

    soundLbx, err := cache.cache.GetLbxFile(lbxPath)
    if err != nil {
        return nil, err
    }

    maker, err := GetSoundMaker(soundLbx, index)
    if err != nil {
        return nil, err
    }

    cache.data[key] = maker

    return maker(), nil
}

func (cache *AudioCache) GetCombatSound(index int) (*audiolib.Player, error) {
    return cache.getSound(CombatLbxPath, index)
}

func (cache *AudioCache) GetNewSound(index int) (*audiolib.Player, error) {
    return cache.getSound(NewSoundLbxPath, index)
}

func (cache *AudioCache) GetSound(index int) (*audiolib.Player, error) {
    if index > 230 {
        return cache.GetNewSound(256 - index)
    }

    return cache.getSound(SoundLbxPath, index)
}
