package gamemenu

import (
    "io"
    "io/fs"
    "log"
    "context"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/hajimehoshi/ebiten/v2"
)

type GameLoader interface {
    Load(reader io.Reader) error
}

func MakeGameMenuUI(cache *lbx.LbxCache, gameLoader GameLoader, doQuit func()) (*uilib.UIElementGroup, context.Context) {
    quit, cancel := context.WithCancel(context.Background())

    imageCache := util.MakeImageCache(cache)

    background, _ := imageCache.GetImage("load.lbx", 0, 0)

    group := uilib.MakeGroup()

    group.Update = func(){
        if gameLoader != nil {
            dropped := ebiten.DroppedFiles()

            if dropped != nil {
                files, err := fs.ReadDir(dropped, ".")
                if err == nil {
                    for _, file := range files {
                        log.Printf("Dropped file: %v", file.Name())
                        func (){
                            opened, err := dropped.Open(file.Name())
                            if err != nil {
                                return
                            }
                            defer opened.Close()

                            // load has a side effect of storing the new game. the main loop will pick this up and switch to the new game
                            err = gameLoader.Load(opened)
                            if err != nil {
                                log.Printf("Error loading dropped save file: %v", err)
                            } else {
                                cancel()
                            }
                        }()
                    }
                }
            }
        }
    }

    makeButton := func (index int, x int, y int, action func()) *uilib.UIElement {
        useImage, _ := imageCache.GetImage("load.lbx", index, 0)
        return &uilib.UIElement{
            Layer: 3,
            Rect: util.ImageRect(x, y, useImage),
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement){
                action()
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
                scale.DrawScaled(screen, useImage, &options)
            },
        }
    }

    var backgroundOptions ebiten.DrawImageOptions
    group.AddElement(&uilib.UIElement{
        Layer: 2,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            scale.DrawScaled(screen, background, &backgroundOptions)
        },
    })

    // quit
    group.AddElement(makeButton(2, 43, 171, func(){
        doQuit()
        cancel()
    }))

    // load
    group.AddElement(makeButton(1, 83, 171, func(){
        // FIXME
    }))

    // save
    group.AddElement(makeButton(3, 122, 171, func(){
        // FIXME
    }))

    // settings
    group.AddElement(makeButton(12, 172, 171, func(){
        // disable for now
        /*
        ui.RemoveElements(elements)

        game.MakeSettingsUI(&imageCache, ui, &background, func(){
            quit = true
            // re-enter the game menu
            game.Events <- &GameEventGameMenu{}
        })
        */
    }))

    // ok
    group.AddElement(makeButton(4, 231, 171, func(){
        cancel()
    }))

    return group, quit
}
