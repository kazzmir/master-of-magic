package intro

import (
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
)

type IntroState int
const (
    IntroStateRunning IntroState = iota
    IntroStateDone
)

const DefaultAnimationSpeed = 10

type Intro struct {
    Counter uint64
    CurrentScene int
    MaxScene int
    Scene *util.Animation
    CurrentIndex int
    ImageCache util.ImageCache
    LbxCache *lbx.LbxCache
    AnimationSpeed uint64
}

func MakeIntro(lbxCache *lbx.LbxCache, animationSpeed uint64) (*Intro, error) {

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
        AnimationSpeed: animationSpeed,
    }, nil
}

func (intro *Intro) Update() IntroState {
    if intro.CurrentScene >= intro.MaxScene {
        return IntroStateDone
    }

    intro.Counter += 1

    if intro.Counter % intro.AnimationSpeed == 0 {
        if !intro.Scene.Next() {
            intro.CurrentScene += 1

            if intro.CurrentScene == 3 {
                player, err := audio.LoadSound(intro.LbxCache, 3)
                if err == nil {
                    player.Play()
                }
            }

            intro.ImageCache.Clear()
            images, err := intro.ImageCache.GetImages("intro.lbx", intro.CurrentScene)
            if err == nil {
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

    if intro.Scene.Frame() != nil {
        var options ebiten.DrawImageOptions
        screen.DrawImage(intro.Scene.Frame(), &options)
    }
}
