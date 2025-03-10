package mainview

import (
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

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

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
    }

    getAlpha = ui.MakeFadeIn(8)

    var elements []*uilib.UIElement

    makeButton := func(index int, x, y int, action func()) *uilib.UIElement {
        images, _ := main.ImageCache.GetImages("mainscrn.lbx", index)
        rect := util.ImageRect(x, y, images[0])
        imageIndex := 1
        return &uilib.UIElement{
            Rect: rect,
            LeftClick: func(element *uilib.UIElement){
                action()
            },
            Inside: func(element *uilib.UIElement, x, y int) {
                imageIndex = 0
            },
            NotInside: func(element *uilib.UIElement){
                imageIndex = 1
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
                options.ColorScale.ScaleAlpha(getAlpha())
                scale.DrawScaled(screen, images[imageIndex], &options)
            },
        }
    }

    // continue
    elements = append(elements, makeButton(1, 110, 130, func(){
        log.Printf("continue")
    }))

    // load game
    elements = append(elements, makeButton(2, 110, 130 + 16 * 1, func(){
        log.Printf("load")
    }))

    // new game
    elements = append(elements, makeButton(3, 110, 130 + 16 * 2, func(){
        main.State = MainScreenStateNewGame
    }))

    // exit
    elements = append(elements, makeButton(4, 110, 130 + 16 * 3, func(){
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
