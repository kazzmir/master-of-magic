package audio

import (
    "github.com/kazzmir/master-of-magic/lib/lbx"
)

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
