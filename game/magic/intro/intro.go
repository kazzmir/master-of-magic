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
    Scene *util.Animation
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

    imageCache := util.MakeImageCache(lbxCache)
    images, err := imageCache.GetImages("intro.lbx", 2)
    if err != nil {
        return nil, err
    }

    return &Intro{
        // skip corporate graphics
        CurrentScene: 2,
        MaxScene: sceneCount,
        Scene: util.MakeAnimation(images, false),
        ImageCache: imageCache,
        LbxCache: lbxCache,
    }, nil
}

func (intro *Intro) Update() IntroState {
    if intro.CurrentScene >= intro.MaxScene {
        return IntroStateDone
    }

    intro.Counter += 1

    if intro.Counter % animationSpeed == 0 {
        if !intro.Scene.Next() {
            intro.CurrentScene += 1
            images, err := intro.ImageCache.GetImages("intro.lbx", intro.CurrentScene)
            if err == nil {
                intro.ImageCache.Clear()
                intro.Scene = util.MakeAnimation(images, false)
            }
        }
    }
    
    return IntroStateRunning
}

func (intro *Intro) Draw(screen *ebiten.Image){
    if intro.CurrentScene >= intro.MaxScene {
        return
    }

    var options ebiten.DrawImageOptions
    if intro.Scene.Frame() != nil {
        screen.DrawImage(intro.Scene.Frame(), &options)
    }
    /*
    img, err := intro.ImageCache.GetImage("intro.lbx", intro.CurrentScene, intro.CurrentIndex)
    if err == nil {
        var options ebiten.DrawImageOptions
        screen.DrawImage(img, &options)
    }
    */
}
