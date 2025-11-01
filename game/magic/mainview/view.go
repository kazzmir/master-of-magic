package mainview

import (
    "log"
    "image"
    // "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/gamemenu"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    helplib "github.com/kazzmir/master-of-magic/game/magic/help"
    "github.com/kazzmir/master-of-magic/lib/font"

    "github.com/hajimehoshi/ebiten/v2"
)

type MainScreenState int

const (
    MainScreenStateRunning MainScreenState = iota
    MainScreenStateQuit
    MainScreenStateNewGame
    MainScreenStateLoadGame
)

type MainScreenEvent interface {
}

type MainScreenEventLoad struct {
}

type MainScreen struct {
    Counter uint64
    Cache *lbx.LbxCache
    State MainScreenState
    ImageCache util.ImageCache
    UI *uilib.UI
    GameLoader gamemenu.GameLoader
    Events chan MainScreenEvent
    Drawer func(screen *ebiten.Image)
}

func MakeMainScreen(cache *lbx.LbxCache, gameLoader gamemenu.GameLoader) *MainScreen {
    main := &MainScreen{
        Counter: 0,
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        State: MainScreenStateRunning,
        GameLoader: gameLoader,
        Events: make(chan MainScreenEvent, 10),
    }

    main.Drawer = func(screen *ebiten.Image) {
        main.UI.Draw(main.UI, screen)
    }

    main.UI = main.MakeUI()
    return main
}

func (main *MainScreen) MakeUI() *uilib.UI {

    help, err := helplib.ReadHelpFromCache(main.Cache)
    if err != nil {
        log.Printf("error reading help: %v", err)
        return nil
    }

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

    fontLoader, err := fontslib.Loader(main.Cache)
    if err != nil {
        log.Printf("error loading fonts: %v", err)
        return nil
    }
    creditsFont := fontLoader(fontslib.NormalYellow)
    titleFont := fontLoader(fontslib.TitleFontOrange)
    titleFontHighlight := fontLoader(fontslib.BigFont)

    getAlpha = ui.MakeFadeIn(8)

    var elements []*uilib.UIElement

    makeButton := func(cx, cy int, helpName string, isActive bool, action func()) *uilib.UIElement {
        length := int(titleFont.MeasureTextWidth(helpName, 1))

        x := cx - length / 2
        y := cy

        rect := image.Rect(x, y, x + length, y + titleFont.Height())
        inside := false
        return &uilib.UIElement{
            Rect: rect,
            LeftClick: func(element *uilib.UIElement){
                if isActive {
                    action()
                }
            },
            RightClick: func(element *uilib.UIElement){
                helpEntries := help.GetEntriesByName(helpName)
                if helpEntries != nil {
                    ui.AddElement(uilib.MakeHelpElementWithLayer(ui, main.Cache, &main.ImageCache, 1, helpEntries[0], helpEntries[1:]...))
                }
            },
            Inside: func(element *uilib.UIElement, x, y int) {
                if isActive {
                    inside = true
                }
            },
            NotInside: func(element *uilib.UIElement){
                inside = false
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {

                use := titleFont
                if inside {
                    use = titleFontHighlight
                }

                use.PrintOptions(screen, float64(cx), float64(y), font.FontOptions{DropShadow: true, Scale: scale.ScaleAmount, Justify: font.FontJustifyCenter}, helpName)
            },
        }
    }

    abs := func(x int) int {
        if x < 0 {
            return -x
        }
        return x
    }

    type creditsLine struct {
        lineRight, lineCenter, lineLeft string
    }

    makeCreditsSection := func (title string, things ...string) []creditsLine {
        lines := make([]creditsLine, len(things))
        for i, thing := range things {
            if i == 0 {
                lines[i].lineLeft = title
            }

            lines[i].lineRight = thing
        }

        return lines
    }

    makeCreditsTitle := func (title string) []creditsLine {
        return []creditsLine{
            {lineCenter: title},
        }
    }

    makeCreditsBlank := func (count int) []creditsLine {
        lines := make([]creditsLine, count)
        for i := range lines {
            lines[i] = creditsLine{}
        }
        return lines
    }

    appendAll := func (lines ...[]creditsLine) []creditsLine {
        result := make([]creditsLine, 0)
        for _, line := range lines {
            result = append(result, line...)
        }
        return result
    }

    credits := appendAll(
        makeCreditsTitle("MASTER OF MAGIC 2025"), // TODO: update the name :D
        makeCreditsBlank(1),
        makeCreditsSection("Programming", "Jon Rafkind (kazzmir)", "Marc Sommerhalder (msom)", "Vlad Kovun (sidav)"),
        makeCreditsBlank(2),
        makeCreditsSection("Thanks to:", "Master of Magic Wiki", "https://masterofmagic.fandom.com"),
        makeCreditsBlank(4),
        makeCreditsTitle("MASTER OF MAGIC 1994"),
        makeCreditsBlank(1),
        makeCreditsSection("Game Designer", "Steve Barcia"),
        makeCreditsBlank(1),
        makeCreditsSection("Programmers", "Jim Cowlishaw", "Ken Burd", "Steve Barcia", "Grissel Barcia"),
        makeCreditsBlank(1),
        makeCreditsSection("Producer", "Doug Caspian-Kaufman"),
        makeCreditsBlank(1),
        makeCreditsSection("Art Director", "Jeff Dee"),
        makeCreditsBlank(1),
        makeCreditsSection("Artists", "Shelly Hollen", "Amanda Dee", "Steve Austin", "George Purdy", "Patrick Owens", "Grissel Barcia"),
        makeCreditsBlank(1),
        makeCreditsSection("Music Producer", "The Fat Man"),
        makeCreditsBlank(1),
        makeCreditsSection("Composer", "Dave Govett"),
        makeCreditsBlank(1),
        makeCreditsSection("QA Lead", "Destin Strader"),
        makeCreditsBlank(1),
        makeCreditsSection("Play Test", "Mike Balogh", "Damon Harris", "Geoff Gessner", "Tammy Talbott", "Mick Uhl", "Jim Hendry", "Frank Brown", "Jim Tricario", "Jen MacLean", "Brian Wilson", "Brian Helleson", "Jeff Dinger", "Chris Bowling", "Charles Brubacker", "Tom Hughes"),
        makeCreditsBlank(1),
        makeCreditsSection("Sound Effects", "Midian"),
        makeCreditsBlank(1),
        makeCreditsSection("Speech", "Mark Reis", "Peter Woods", "David Ellis"),
        makeCreditsBlank(1),
        makeCreditsSection("Manual", "Petra Schlunk"),
        makeCreditsBlank(1),
        makeCreditsSection("Special thanks", "Jenna Cowlishaw"),
    )

    creditsRect := image.Rect(60, 35, 270, 130)
    elements = append(elements, &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
            sub := screen.SubImage(scale.ScaleRect(creditsRect)).(*ebiten.Image)

            var options ebiten.DrawImageOptions

            gap := 40

            where := (ui.Counter / 3) % uint64(creditsRect.Dy() + gap + (len(credits)) * creditsFont.Height())
            middle := creditsRect.Min.X + creditsRect.Dx() / 2
            for i, currentLine := range credits {
                if len(currentLine.lineLeft) + len(currentLine.lineCenter) + len(currentLine.lineRight) == 0 {
                    continue
                }
                y := creditsRect.Max.Y + i * creditsFont.Height() + gap - int(where)

                options.ColorScale.Reset()
                options.ColorScale.ScaleAlpha(getAlpha())
                distance := abs(y - (creditsRect.Min.Y + creditsRect.Dy() / 2))
                // log.Printf("i=%v distance=%v dy=%v", i, distance, creditsRect.Dy() - 20)

                alpha := float32(creditsRect.Dy()/2 + 10 - distance) / float32(creditsRect.Dy()/2)
                alpha = min(alpha, 1)
                alpha = max(alpha, 0)
                options.ColorScale.ScaleAlpha(alpha)

                if len(currentLine.lineLeft) > 0 {
                    creditsFont.PrintOptions(sub, float64(creditsRect.Min.X), float64(y), font.FontOptions{DropShadow: true, Scale: scale.ScaleAmount, Justify: font.FontJustifyLeft, Options: &options}, currentLine.lineLeft)
                }
                if len(currentLine.lineCenter) > 0 {
                    creditsFont.PrintOptions(sub, float64(middle), float64(y), font.FontOptions{DropShadow: true, Scale: scale.ScaleAmount, Justify: font.FontJustifyCenter, Options: &options}, currentLine.lineCenter)
                }
                if len(currentLine.lineRight) > 0 {
                    creditsFont.PrintOptions(sub, float64(creditsRect.Max.X), float64(y), font.FontOptions{DropShadow: true, Scale: scale.ScaleAmount, Justify: font.FontJustifyRight, Options: &options}, currentLine.lineRight)
                }
            }

            // for debugging
            // util.DrawRect(screen, scale.ScaleRect(creditsRect), color.RGBA{R: 255, G: 255, B: 255, A: 255})
        },
    })

    // TODO: when load game functionality is there, take save files presence into account for these vars
    isContinueBtnActive := false
    isLoadGameBtnActive := true

    centerX := data.ScreenWidth / 2
    yBase := 154
    yGap := titleFont.Height() + 1

    // continue
    elements = append(elements, makeButton(centerX, yBase, "Continue", isContinueBtnActive, func(){
        log.Printf("continue")
    }))

    // load game
    elements = append(elements, makeButton(centerX, yBase + yGap * 1, "Load", isLoadGameBtnActive, func(){
        select {
            case main.Events <- &MainScreenEventLoad{}:
            default:
        }
    }))

    // new game
    elements = append(elements, makeButton(centerX, yBase + yGap * 2, "New Game", true, func(){
        main.State = MainScreenStateNewGame
    }))

    // FIXME: add "Hall of Fame" button

    // exit
    elements = append(elements, makeButton(centerX, yBase + yGap * 3, "Quit to Dos", true, func(){
        main.State = MainScreenStateQuit
    }))

    ui.SetElementsFromArray(elements)

    return ui
}

func (main *MainScreen) RunGameScreen(yield coroutine.YieldFunc) MainScreenState {
    oldDrawer := main.Drawer
    defer func() {
        main.Drawer = oldDrawer
    }()

    gameScreen, quit := gamemenu.MakeGameMenuUI(main.Cache, main.GameLoader, func(){})

    ui := &uilib.UI{
    }

    main.Drawer = func(screen *ebiten.Image) {
        ui.StandardDraw(screen)
    }

    ui.SetElementsFromArray(nil)
    ui.AddGroup(gameScreen)

    for quit.Err() == nil {
        ui.StandardUpdate()
        if yield() != nil {
            break
        }
    }

    yield()

    return MainScreenStateRunning
}

func (main *MainScreen) Update(yield coroutine.YieldFunc) MainScreenState {
    main.Counter += 1

    main.UI.StandardUpdate()

    select {
    case event := <-main.Events:
            switch event.(type) {
                case *MainScreenEventLoad:
                    main.State = main.RunGameScreen(yield)
            }
        default:
    }

    return main.State
}

func (main *MainScreen) Draw(screen *ebiten.Image) {
    main.Drawer(screen)
}
