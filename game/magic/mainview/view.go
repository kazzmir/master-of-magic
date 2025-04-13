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

    creditsRect := image.Rect(60, 35, 270, 130)
    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {

            type creditsLine struct {
                lineRight, lineCenter, lineLeft string
            }
            credits := []creditsLine{
                { lineCenter: "MASTER OF MAGIC 2025"}, // TODO: update the name :D 
                {},
                { lineLeft: "Programming", lineRight: "Jon Rafkind (kazzmir)"},
                { lineRight: "Marc Sommerhalder (msom)"},
                { lineRight: "Vlad Kovun (sidav)"},
                {},
                {},
                { lineLeft: "Thanks to:", lineRight: "Master of Magic Wiki"},
                { lineRight: "https://masterofmagic.fandom.com"},
                {},
                {},
                {},
                {},
                { lineCenter: "MASTER OF MAGIC 1994"},
                {},
                { lineLeft: "Game Designer", lineRight: "Steve Barcia"},
                {}, 
                { lineLeft: "Programmers", lineRight: "Jim Cowlishaw"},
                { lineRight: "Ken Burd"},
                { lineRight: "Steve Barcia"},
                { lineRight: "Grissel Barcia"},
                {},
                { lineLeft: "Producer", lineRight: "Doug Caspian-Kaufman"},
                {},
                { lineLeft: "Art Director", lineRight: "Jeff Dee"},
                {},
                { lineLeft: "Artists", lineRight: "Shelly Hollen"},
                { lineRight: "Amanda Dee"},
                { lineRight: "Steve Austin"},
                { lineRight: "George Purdy"},
                { lineRight: "Patrick Owens"},
                { lineRight: "Grissel Barcia"},
                {},
                { lineLeft: "Music Producer", lineRight: "The Fat Man"},
                {},
                { lineLeft: "Composer", lineRight: "Dave Govett"},
                {},
                { lineLeft: "QA Lead", lineRight: "Destin Strader"},
                {},
                { lineLeft: "Play Test", lineRight: "Mike Balogh"},
                { lineRight: "Damon Harris"},
                { lineRight: "Geoff Gessner"},
                { lineRight: "Tammy Talbott"},
                { lineRight: "Mick Uhl"},
                { lineRight: "Jim Hendry"},
                { lineRight: "Frank Brown"},
                { lineRight: "Jim Tricario"},
                { lineRight: "Jen MacLean"},
                { lineRight: "Brian Wilson"},
                { lineRight: "Brian Helleson"},
                { lineRight: "Jeff Dinger"},
                { lineRight: "Chris Bowling"},
                { lineRight: "Charles Brubacker"},
                { lineRight: "Tom Hughes"},
                {},
                { lineLeft: "Sound Effects", lineRight: "Midian"},
                {},
                { lineLeft: "Speech", lineRight: "Mark Reis"},
                { lineRight: "Peter Woods"},
                { lineRight: "David Ellis"},
                {},
                { lineLeft: "Manual", lineRight: "Petra Schlunk"},
                {},
                { lineLeft: "Special thanks", lineRight: "Jenna Cowlishaw"},
            }

            sub := screen.SubImage(scale.ScaleRect(creditsRect)).(*ebiten.Image)

            var options ebiten.DrawImageOptions

            gap := 40

            where := (ui.Counter / 3) % uint64(creditsRect.Dy() + gap + (len(credits)) * mainFonts.Credits.Height())
            middle := creditsRect.Min.X + creditsRect.Dx() / 2
            for i, currentLine := range credits {
                if len(currentLine.lineLeft) + len(currentLine.lineCenter) + len(currentLine.lineRight) == 0 {
                    continue
                }
                y := creditsRect.Max.Y + i * mainFonts.Credits.Height() + gap - int(where)

                options.ColorScale.Reset()
                options.ColorScale.ScaleAlpha(getAlpha())
                distance := abs(y - (creditsRect.Min.Y + creditsRect.Dy() / 2))
                // log.Printf("i=%v distance=%v dy=%v", i, distance, creditsRect.Dy() - 20)

                alpha := float32(creditsRect.Dy()/2 + 10 - distance) / float32(creditsRect.Dy()/2)
                alpha = min(alpha, 1)
                alpha = max(alpha, 0)
                options.ColorScale.ScaleAlpha(alpha)

                if len(currentLine.lineLeft) > 0 {
                    mainFonts.Credits.PrintOptions(sub, float64(creditsRect.Min.X), float64(y), font.FontOptions{DropShadow: true, Scale: scale.ScaleAmount, Justify: font.FontJustifyLeft, Options: &options}, currentLine.lineLeft)
                }
                if len(currentLine.lineCenter) > 0 {
                    mainFonts.Credits.PrintOptions(sub, float64(middle), float64(y), font.FontOptions{DropShadow: true, Scale: scale.ScaleAmount, Justify: font.FontJustifyCenter, Options: &options}, currentLine.lineCenter)
                }
                if len(currentLine.lineRight) > 0 {
                    mainFonts.Credits.PrintOptions(sub, float64(creditsRect.Max.X), float64(y), font.FontOptions{DropShadow: true, Scale: scale.ScaleAmount, Justify: font.FontJustifyRight, Options: &options}, currentLine.lineRight)
                }
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
