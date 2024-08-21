package intro

import (
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
)

type IntroState int
const (
    IntroStateRunning IntroState = iota
    IntroStateDone
)

type Intro struct {
    Counter uint64
    CurrentScene int
    MaxScene int
    CurrentIndex int
    ImageCache util.ImageCache
    LbxCache *lbx.LbxCache
}

const animationSpeed = 10

func MakeIntro(lbxCache *lbx.LbxCache) (*Intro, error) {

    introLbx, err := lbxCache.GetLbxFile("intro.lbx")
    if err != nil {
        return nil, err
    }

    sceneCount := introLbx.TotalEntries()

    return &Intro{
        // skip corporate graphics
        CurrentScene: 2,
        MaxScene: sceneCount,
        ImageCache: util.MakeImageCache(lbxCache),
        LbxCache: lbxCache,
    }, nil
}

func (intro *Intro) Update() IntroState {
    if intro.CurrentScene >= intro.MaxScene {
        return IntroStateDone
    }

    intro.Counter += 1

    if intro.Counter % animationSpeed == 0 {
        intro.CurrentIndex += 1

        images, err := intro.ImageCache.GetImages("intro.lbx", intro.CurrentScene)
        if err == nil {
            if intro.CurrentIndex >= len(images) {
                intro.CurrentScene += 1
                intro.CurrentIndex = 0
                intro.ImageCache.Clear()
            }
        }
    }
    
    return IntroStateRunning
}

func (intro *Intro) Draw(screen *ebiten.Image){
    if intro.CurrentScene >= intro.MaxScene {
        return
    }

    img, err := intro.ImageCache.GetImage("intro.lbx", intro.CurrentScene, intro.CurrentIndex)
    if err == nil {
        var options ebiten.DrawImageOptions
        screen.DrawImage(img, &options)
    }
}
