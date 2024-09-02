package intro

import (
    "log"

    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
    audiolib "github.com/hajimehoshi/ebiten/v2/audio"
)

type IntroState int
const (
    IntroStateRunning IntroState = iota
    IntroStateDone
)

const DefaultAnimationSpeed = 9

type Scene int

const (
    SceneTitleGraphics Scene = 2
    SceneMarching Scene = 3
    SceneEvilWizardIntro = 4
    SceneGoodWizardWalk = 5
    SceneGoodWizardIntro = 6
    SceneWorkMustStop = 7

    SceneEvilScream = 11
)

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

            log.Printf("Switching to scene %d", intro.CurrentScene)

            var player *audiolib.Player
            var err error
            switch Scene(intro.CurrentScene) {
                case SceneWorkMustStop:
                    player, err = audio.LoadSound(intro.LbxCache, 1)
                case SceneEvilScream:
                    player, err = audio.LoadSound(intro.LbxCache, 3)
                case SceneMarching:
                    player, err = audio.LoadSound(intro.LbxCache, 5)
                case SceneGoodWizardIntro:
                    player, err = audio.LoadSound(intro.LbxCache, 4)
            }

            if err == nil && player != nil {
                player.Play()
            } else if err != nil {
                log.Printf("Unable to load sound for scene %d: %v", intro.CurrentScene, err)
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
