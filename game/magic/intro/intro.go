package intro

import (
    "log"
    "time"

    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
    audiolib "github.com/hajimehoshi/ebiten/v2/audio"
)

type IntroState int
const (
    IntroStateRunning IntroState = iota
    IntroStateDone
)

const DefaultAnimationSpeed = 10

type Scene int

const (
    SceneTitleGraphics Scene = 2
    SceneMarching Scene = 3
    SceneEvilWizardIntro = 4
    SceneGoodWizardWalk = 5
    SceneGoodWizardIntro = 6
    SceneWorkMustStop = 7
    SceneLightningHitsTower = 8
    SceneGoodWizardCast = 9
    SceneLightningHitsShield = 10

    SceneEvilScream = 11
)

type Intro struct {
    Counter uint64
    CurrentScene Scene
    MaxScene Scene
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

    startScene := SceneTitleGraphics
    // startScene = SceneLightningHitsTower

    imageCache := util.MakeImageCache(lbxCache)

    return &Intro{
        // skip corporate graphics
        CurrentScene: startScene,
        MaxScene: Scene(sceneCount),
        Scene: util.MakeAnimation(nil, false),
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

    if intro.Counter % intro.AnimationSpeed == 0 || len(intro.Scene.Frames) == 0 {
        if !intro.Scene.Next() {
            if len(intro.Scene.Frames) > 0 {
                intro.CurrentScene += 1
            }

            log.Printf("Switching to scene %d", intro.CurrentScene)

            var player *audiolib.Player
            var err error
            position := time.Duration(0)
            plays := 1
            delay := time.Duration(0)
            var maxDuration []time.Duration
            switch Scene(intro.CurrentScene) {
                case SceneWorkMustStop:
                    player, err = audio.LoadSound(intro.LbxCache, 1)
                case SceneEvilWizardIntro:
                    player, err = audio.LoadSound(intro.LbxCache, 115)
                case SceneEvilScream:
                    player, err = audio.LoadSound(intro.LbxCache, 3)

                case SceneLightningHitsTower:
                    player, err = audio.LoadSound(intro.LbxCache, 1)
                    if err == nil && player != nil {
                        position = time.Millisecond * 3650
                        plays = 2
                        // first lightning strike is cut short
                        maxDuration = append(maxDuration, time.Millisecond * 4800)
                    }
                case SceneLightningHitsShield:
                    player, err = audio.LoadSound(intro.LbxCache, 1)
                    if err == nil && player != nil {
                        position = time.Millisecond * 3650
                    }
                case SceneGoodWizardCast:
                    player, err = audio.LoadSound(intro.LbxCache, 118)
                case SceneGoodWizardWalk:
                    delay = time.Millisecond * 1200
                    player, err = audio.LoadSound(intro.LbxCache, 4)
                case SceneGoodWizardIntro:
                    player, err = audio.LoadSound(intro.LbxCache, 116)
            }

            if err == nil && player != nil {
                go func(player *audiolib.Player, position time.Duration, plays int, delay time.Duration, maxDuration []time.Duration){
                    if delay > 0 {
                        time.Sleep(delay)
                    }
                    for i := 0; i < plays; i++ {
                        for player.IsPlaying() {
                            // log.Printf("i %v position %v maxDuration %v", i, player.Position(), maxDuration)
                            if i <= len(maxDuration) && player.Position() > maxDuration[i-1] {
                                break
                            }

                            time.Sleep(time.Millisecond * 10)
                        }
                        // player.Pause()
                        player.SetPosition(position)
                        player.Play()
                    }
                }(player, position, plays, delay, maxDuration)
            } else if err != nil {
                log.Printf("Unable to load sound for scene %d: %v", intro.CurrentScene, err)
            }

            intro.ImageCache.Clear()
            images, err := intro.ImageCache.GetImages("intro.lbx", int(intro.CurrentScene))
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
        scale.DrawScaled(screen, intro.Scene.Frame(), &options)
    }
}
