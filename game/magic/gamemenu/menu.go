package gamemenu

import (
    "os"
    "io"
    "io/fs"
    "fmt"
    "bufio"
    "log"
    "context"
    "compress/gzip"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/hajimehoshi/ebiten/v2"
)

type GameLoader interface {
    Load(reader io.Reader) error
}

type SettingsUI interface {
    RunSettingsUI()
}

type GameSaver interface {
    Save(writer io.Writer, saveName string) error
}

func MakeGameMenuUI(cache *lbx.LbxCache, gameLoader GameLoader, saver GameSaver, settingsUI SettingsUI, doQuit func()) (*uilib.UIElementGroup, context.Context) {
    quit, cancel := context.WithCancel(context.Background())

    imageCache := util.MakeImageCache(cache)

    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Error loading fonts: %v", err)
        cancel()
        return nil, quit
    }

    background, _ := imageCache.GetImage("load.lbx", 0, 0)

    group := uilib.MakeGroup()

    getAlpha := group.MakeFadeIn(7)

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

    var selectedSlot *uilib.UIElement
    selectedIndex := -1
    var slotName *string

    source := ebiten.NewImage(1, 1)
    source.Fill(color.RGBA{R: 0xcf, G: 0xef, B: 0xf9, A: 0xff})

    useFont := loader(fontslib.NameFont)

    makeSaveSlot := func(index int) *uilib.UIElement {
        x := 43
        y := 44 + (index - 1) * 15

        // inside := false
        name := ""
        return &uilib.UIElement{
            Layer: 3,
            Rect: image.Rect(0, 0, 229, 12).Add(image.Pt(x, y)),
            /*
            Inside: func(element *uilib.UIElement, x int, y int){
                inside = true
            },
            NotInside: func(element *uilib.UIElement){
                inside = false
            },
            */
            GetText: func() string {
                return name
            },
            TextEntry: func(element *uilib.UIElement, text string) string {
                name = text

                for len(name) > 0 && useFont.MeasureTextWidth(name, 1) > float64(element.Rect.Bounds().Dx()) {
                    name = name[:len(name)-1]
                }

                return name
            },
            HandleKeys: func(keys []ebiten.Key) {
                for _, key := range keys {
                    switch key {
                    case ebiten.KeyEnter:
                        /*
                        if len(name) > 0 {
                            group.RemoveElement(self)
                        }
                        */
                    case ebiten.KeyBackspace:
                        if len(name) > 0 {
                            name = name[:len(name) - 1]
                        }
                    }
                }
            },
            LeftClick: func(element *uilib.UIElement){
                if selectedSlot != element {
                    selectedSlot = element
                    selectedIndex = index

                    slotName = &name
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                /*
                c := color.RGBA{R: 255, A: 255}
                if inside {
                    c = color.RGBA{R: 255, G: 255, A: 255}
                }
                util.DrawRect(screen, scale.ScaleRect(element.Rect), c)
                */

                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(getAlpha())
                useFont.PrintOptions(screen, float64(x + 2), float64(y + 3), font.FontOptions{Scale: scale.ScaleAmount, DropShadow: true, Options: &options}, name)

                if selectedSlot == element {
                    // draw cursor
                    cursorX := float64(x + 2) + useFont.MeasureTextWidth(name, 1)

                    // maybe pass in alpha here
                    util.DrawTextCursor(screen, source, cursorX, float64(y + 3), group.Counter)
                }
            },
        }
    }

    // save slot
    for i := range 8 {
        group.AddElement(makeSaveSlot(i + 1))
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
                options.ColorScale.ScaleAlpha(getAlpha())
                scale.DrawScaled(screen, useImage, &options)
            },
        }
    }

    group.AddElement(&uilib.UIElement{
        Layer: 2,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var backgroundOptions ebiten.DrawImageOptions
            backgroundOptions.ColorScale.ScaleAlpha(getAlpha())
            scale.DrawScaled(screen, background, &backgroundOptions)
        },
    })

    // quit
    group.AddElement(makeButton(2, 43, 171, func(){
        getAlpha = group.MakeFadeOut(7)
        group.AddDelay(7, func(){
            doQuit()
            cancel()
        })
    }))

    // load
    group.AddElement(makeButton(1, 83, 171, func(){
        // FIXME
    }))

    // save
    group.AddElement(makeButton(3, 122, 171, func(){
        if selectedIndex != -1 {
            path := fmt.Sprintf("file%d.magic-save", selectedIndex)
            saveFile, err := os.Create(path)
            if err != nil {
                log.Printf("Error creating save file: %v", err)
            } else {
                defer saveFile.Close()

                bufferedOut := bufio.NewWriter(saveFile)
                defer bufferedOut.Flush()

                gzipWriter := gzip.NewWriter(bufferedOut)
                defer gzipWriter.Close()

                err = saver.Save(gzipWriter, *slotName)
                if err != nil {
                    log.Printf("Error saving game: %v", err)
                } else {
                    if err != nil {
                        log.Printf("Error flushing save file: %v", err)
                    }

                    log.Printf("Game saved to '%s'", path)
                }
            }
        }
    }))

    // settings
    group.AddElement(makeButton(12, 172, 171, func(){
        // disable for now
        // ui.RemoveElements(elements)

        settingsUI.RunSettingsUI()
    }))

    // ok
    group.AddElement(makeButton(4, 231, 171, func(){
        getAlpha = group.MakeFadeOut(7)
        group.AddDelay(7, func(){
            cancel()
        })
    }))

    return group, quit
}
