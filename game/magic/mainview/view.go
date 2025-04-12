package mainview

import (
    "log"
    "image"
    // "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/lib/font"

    "github.com/hajimehoshi/ebiten/v2"
)

type MainScreenState int

const (
    MainScreenStateRunning MainScreenState = iota
    MainScreenStateQuit
    MainScreenStateNewGame
)

type MainScreen struct {
    Counter uint64
    Cache *lbx.LbxCache
    State MainScreenState
    ImageCache util.ImageCache
    UI *uilib.UI
}

func MakeMainScreen(cache *lbx.LbxCache) *MainScreen {
    main := &MainScreen{
        Counter: 0,
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        State: MainScreenStateRunning,
    }

    main.UI = main.MakeUI()
    return main
}

func (main *MainScreen) MakeUI() *uilib.UI {

    var getAlpha util.AlphaFadeFunc

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image) {
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(getAlpha())

            top, err := main.ImageCache.GetImages("mainscrn.lbx", 0)
            if err == nil {
                use := top[(main.Counter / 4) % uint64(len(top))]
                scale.DrawScaled(screen, use, &options)
                options.GeoM.Translate(0, float64(use.Bounds().Dy()))
            }

            background, err := main.ImageCache.GetImage("mainscrn.lbx", 5, 0)
            if err == nil {
                scale.DrawScaled(screen, background, &options)
            }

            ui.StandardDraw(screen)
        },
    }

    getAlpha = ui.MakeFadeIn(8)

    var elements []*uilib.UIElement

    makeButton := func(index int, x, y int, isActive bool, action func()) *uilib.UIElement {
        images, _ := main.ImageCache.GetImages("mainscrn.lbx", index)
        rect := util.ImageRect(x, y, images[0])
        imageIndex := 1
        return &uilib.UIElement{
            Rect: rect,
            LeftClick: func(element *uilib.UIElement){
                if isActive {
                    action()
                }
            },
            Inside: func(element *uilib.UIElement, x, y int) {
                if isActive {
                    imageIndex = 0
                }
            },
            NotInside: func(element *uilib.UIElement){
                imageIndex = 1
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
                options.ColorScale.ScaleAlpha(getAlpha())
                if !isActive {
                    options.ColorScale.Scale(0.4, 0.4, 0.4, 1.0)
                }
                scale.DrawScaled(screen, images[imageIndex], &options)
            },
        }
    }

    mainFonts := fontslib.MakeMainFonts(main.Cache)

    abs := func(x int) int {
        if x < 0 {
            return -x
        }
        return x
    }

    creditsRect := image.Rect(30, 35, 300, 130)
    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {

            credits := []string{
                "Programming",
                "Jon Rafkind (kazzmir)",
                "msom",
                "sidav",
            }

            sub := screen.SubImage(scale.ScaleRect(creditsRect)).(*ebiten.Image)

            var options ebiten.DrawImageOptions

            gap := 40

            where := (ui.Counter / 3) % uint64(creditsRect.Dy() + gap + (len(credits)) * mainFonts.Credits.Height())
            middle := creditsRect.Min.X + creditsRect.Dx() / 2
            for i, line := range credits {
                y := creditsRect.Max.Y + i * mainFonts.Credits.Height() + gap - int(where)

                options.ColorScale.Reset()
                options.ColorScale.ScaleAlpha(getAlpha())
                distance := abs(y - (creditsRect.Min.Y + creditsRect.Dy() / 2))
                // log.Printf("i=%v distance=%v dy=%v", i, distance, creditsRect.Dy() - 20)

                alpha := float32(creditsRect.Dy()/2 + 10 - distance) / float32(creditsRect.Dy()/2)
                if alpha > 1 {
                    alpha = 1
                }
                if alpha < 0 {
                    alpha = 0
                }
                options.ColorScale.ScaleAlpha(alpha)

                mainFonts.Credits.PrintOptions(sub, float64(middle), float64(y), font.FontOptions{DropShadow: true, Scale: scale.ScaleAmount, Justify: font.FontJustifyCenter, Options: &options}, line)
            }

            // for debugging
            // util.DrawRect(screen, scale.ScaleRect(creditsRect), color.RGBA{R: 255, G: 255, B: 255, A: 255})
        },
    })

    // TODO: when load game functionality is there, take save files presence into account for these vars
    isContinueBtnActive := false
    isLoadGameBtnActive := false

    // continue
    elements = append(elements, makeButton(1, 110, 130, isContinueBtnActive, func(){
        log.Printf("continue")
    }))

    // load game
    elements = append(elements, makeButton(2, 110, 130 + 16 * 1, isLoadGameBtnActive, func(){
        log.Printf("load")
    }))

    // new game
    elements = append(elements, makeButton(3, 110, 130 + 16 * 2, true, func(){
        main.State = MainScreenStateNewGame
    }))

    // exit
    elements = append(elements, makeButton(4, 110, 130 + 16 * 3, true, func(){
        main.State = MainScreenStateQuit
    }))

    ui.SetElementsFromArray(elements)

    return ui
}

func (main *MainScreen) Update() MainScreenState {
    main.Counter += 1

    main.UI.StandardUpdate()

    return main.State
}

func (main *MainScreen) Draw(screen *ebiten.Image) {
    main.UI.Draw(main.UI, screen)
}
