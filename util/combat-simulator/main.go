package main

import (
    "embed"
    "log"
    "fmt"
    "image"
    "image/color"
    "slices"
    "cmp"
    "errors"
    "strings"
    "math/rand/v2"
    "encoding/json"
    "net/http"
    "time"
    "os"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/util/common"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/colorm"
    "github.com/hajimehoshi/ebiten/v2/text/v2"

    "github.com/ebitenui/ebitenui"
    // "github.com/ebitenui/ebitenui/input"
    "github.com/ebitenui/ebitenui/widget"
    "github.com/ebitenui/ebitenui/event"
    ui_image "github.com/ebitenui/ebitenui/image"
)

//go:embed assets/*
var assets embed.FS

//go:embed key/*
var key embed.FS

func loadKey() (string, error) {
    keyFile, err := key.ReadFile("key/key.txt")
    if err != nil {
        return "", err
    }

    return strings.TrimSpace(string(keyFile)), nil
}

const ReportServer = "https://magic.jonrafkind.com/report"
// const ReportServer = "http://localhost:5000/report"

type EngineMode int
const (
    EngineModeMenu EngineMode = iota
    EngineModeCombat
    EngineModeBugReport
)

type CombatDescription struct {
    DefenderUnits []units.Unit
    AttackerUnits []units.Unit
}

func (description *CombatDescription) Save(filename string) error {
    return os.WriteFile(filename, []byte(description.String()), 0644)
}

func UnitFromName(name string) (units.Unit, error) {
    allRaces := append(append(data.ArcanianRaces(), data.MyrranRaces()...), []data.Race{data.RaceFantastic, data.RaceHero, data.RaceAll}...)

    for _, race := range allRaces {
        if strings.HasPrefix(name, race.String()) {
            name = strings.TrimSpace(name[len(race.String()):])
            choices := units.UnitsByRace(race)
            for _, unit := range choices {
                if unit.Name == name {
                    return unit, nil
                }
            }
        }
    }

    return units.Unit{}, fmt.Errorf("unit not found: %v", name)
}

func LoadCombatDescription(filename string) (*CombatDescription, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    output := make(map[string]any)
    err = json.Unmarshal([]byte(data), &output)
    if err != nil {
        return nil, err
    }

    defenders := make([]units.Unit, 0)
    for _, name := range output["defenders"].([]any) {
        unit, err := UnitFromName(name.(string))
        if err != nil {
            return nil, err
        }
        defenders = append(defenders, unit)
    }

    attackers := make([]units.Unit, 0)
    for _, name := range output["attackers"].([]any) {
        unit, err := UnitFromName(name.(string))
        if err != nil {
            return nil, err
        }
        attackers = append(attackers, unit)
    }

    return &CombatDescription{
        DefenderUnits: defenders,
        AttackerUnits: attackers,
    }, nil
}

func (description *CombatDescription) String() string {
    out := make(map[string]any)

    defenders := make([]string, 0)

    for _, unit := range description.DefenderUnits {
        defenders = append(defenders, fmt.Sprintf("%v %v", unit.Race, unit.Name))
    }

    out["defenders"] = defenders

    attackers := make([]string, 0)
    for _, unit := range description.AttackerUnits {
        attackers = append(attackers, fmt.Sprintf("%v %v", unit.Race, unit.Name))
    }

    out["attackers"] = attackers

    jsonString, err := json.MarshalIndent(out, "", "  ")
    if err != nil {
        log.Printf("Error with json: %v", err)
        return fmt.Sprintf("Error: %v", err)
    }

    return string(jsonString)
}

func sendReport(report string) error {
    client := &http.Client{
        Timeout: time.Second * 10,
    }

    reportKey, err := loadKey()
    if err != nil {
        return err
    }

    request, err := http.NewRequest("POST", ReportServer, strings.NewReader(report))
    if err != nil {
        return err
    }

    request.Header.Set("Content-Type", "text/plain")
    request.Header.Set("X-Report-Key", reportKey)

    response, err := client.Do(request)
    if err != nil {
        return err
    }

    defer response.Body.Close()

    if response.StatusCode != 200 {
        return fmt.Errorf("server returned status %v", response.StatusCode)
    } else {
        log.Printf("Bug report successfully sent")
        return nil
    }
}

type Engine struct {
    Cache *lbx.LbxCache
    Mode EngineMode
    UI *ebitenui.UI
    BugUI *ebitenui.UI
    Combat *combat.CombatScreen
    LastCombatScreen *ebiten.Image
    Coroutine *coroutine.Coroutine
    CombatDescription CombatDescription
    Counter uint64
    UIUpdates chan func()
}

func MakeEngine(cache *lbx.LbxCache) *Engine {
    engine := Engine{
        Cache: cache,
        LastCombatScreen: ebiten.NewImage(data.ScreenWidth, data.ScreenHeight),
        UIUpdates: make(chan func(), 10000),
    }

    engine.UI = engine.MakeUI()

    /*
    engine.BugUI = engine.MakeBugUI()
    engine.Mode = EngineModeBugReport
    */

    return &engine
}

var CombatDoneErr = errors.New("combat done")

func makeRoundedButtonImage(width int, height int, border int, col color.Color) *ebiten.Image {
    img := ebiten.NewImage(width, height)

    vector.DrawFilledRect(img, float32(border), 0, float32(width - border * 2), float32(height), col, true)
    vector.DrawFilledRect(img, 0, float32(border), float32(width), float32(height - border * 2), col, true)
    vector.DrawFilledCircle(img, float32(border), float32(border), float32(border), col, true)
    vector.DrawFilledCircle(img, float32(width-border), float32(border), float32(border), col, true)
    vector.DrawFilledCircle(img, float32(border), float32(height-border), float32(border), col, true)
    vector.DrawFilledCircle(img, float32(width-border), float32(height-border), float32(border), col, true)

    return img
}

func padding(n int) widget.Insets {
    return widget.Insets{Top: n, Bottom: n, Left: n, Right: n}
}

func space(size int) *widget.Container {
    return widget.NewContainer(
        widget.ContainerOpts.WidgetOpts(
            widget.WidgetOpts.MinSize(size, size),
        ),
    )
}

func lighten(c color.Color, amount float64) color.Color {
    var change colorm.ColorM
    change.ChangeHSV(0, 1 - amount/100, 1 + amount/100)
    return change.Apply(c)
}

func makeNineImage(img *ebiten.Image, border int) *ui_image.NineSlice {
    width := img.Bounds().Dx()
    return ui_image.NewNineSliceSimple(img, border, width - border * 2)
}

func makeNineRoundedButtonImage(width int, height int, border int, col color.Color) *widget.ButtonImage {
    return &widget.ButtonImage{
        Idle: makeNineImage(makeRoundedButtonImage(width, height, border, col), border),
        Hover: makeNineImage(makeRoundedButtonImage(width, height, border, lighten(col, 50)), border),
        Pressed: makeNineImage(makeRoundedButtonImage(width, height, border, lighten(col, 20)), border),
    }
}

func makeBorderOutline(col color.Color) *ui_image.NineSlice {
    img := ebiten.NewImage(20, 20)
    vector.StrokeRect(img, 0, 0, 18, 18, 1, col, true)
    vector.StrokeLine(img, 19, 0, 19, 19, 1, lighten(col, -80), true)
    return makeNineImage(img, 3)
}

func makeRow(spacing int, children ...widget.PreferredSizeLocateableWidget) *widget.Container {
    container := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
            widget.RowLayoutOpts.Spacing(spacing),
        )),
    )

    for _, child := range children {
        container.AddChild(child)
    }

    return container
}

func (engine *Engine) Update() error {
    engine.Counter += 1

    keys := inpututil.AppendJustPressedKeys(nil)

    for _, key := range keys {
        switch key {
            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                switch engine.Mode {
                    case EngineModeMenu:
                        return ebiten.Termination
                    case EngineModeCombat:
                        engine.Mode = EngineModeBugReport
                        engine.BugUI = engine.MakeBugUI()
                        engine.Combat.Draw(engine.LastCombatScreen)
                    case EngineModeBugReport:
                        engine.Mode = EngineModeCombat
                }
        }
    }

    done := false
    for !done {
        select {
            case update := <-engine.UIUpdates:
                update()
            default:
                done = true
        }
    }

    switch engine.Mode {
        case EngineModeMenu:
            engine.UI.Update()
        case EngineModeBugReport:
            engine.BugUI.Update()
        case EngineModeCombat:
            inputmanager.Update()
            err := engine.Coroutine.Run()
            if errors.Is(err, CombatDoneErr) {
                engine.Combat = nil
                engine.Coroutine = nil
                engine.Mode = EngineModeMenu
            }
    }

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    switch engine.Mode {
        case EngineModeMenu:
            engine.UI.Draw(screen)
        case EngineModeBugReport:
            var options ebiten.DrawImageOptions
            // fixme: do aspect ratio scaling
            xScale := float64(screen.Bounds().Dx()) / float64(engine.LastCombatScreen.Bounds().Dx())
            yScale := float64(screen.Bounds().Dy()) / float64(engine.LastCombatScreen.Bounds().Dy())
            scale := xScale
            if yScale < xScale {
                scale = yScale
            }
            options.GeoM.Scale(scale, scale)
            if xScale < yScale {
                options.GeoM.Translate(0, (float64(screen.Bounds().Dy()) - float64(engine.LastCombatScreen.Bounds().Dy()) * scale) / 2)
            } else {
                options.GeoM.Translate((float64(screen.Bounds().Dx()) - float64(engine.LastCombatScreen.Bounds().Dx()) * scale) / 2, 0)
            }
            screen.DrawImage(engine.LastCombatScreen, &options)
            engine.BugUI.Draw(screen)
        case EngineModeCombat:
            engine.Combat.Draw(screen)
            mouse.Mouse.Draw(screen)
    }
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    switch engine.Mode {
        case EngineModeMenu, EngineModeBugReport:
            return outsideWidth, outsideHeight
        case EngineModeCombat:
            return scale.Scale2(data.ScreenWidth, data.ScreenHeight)
    }

    return outsideWidth, outsideHeight
}

func (engine *Engine) EnterCombat(combatDescription CombatDescription) {
    // defenderUnits []units.Unit, attackerUnits []units.Unit
    engine.Mode = EngineModeCombat
    engine.CombatDescription = combatDescription

    allSpells, err := spellbook.ReadSpellsFromCache(engine.Cache)
    if err != nil {
        log.Printf("Cannot load spells: %v", err)
    }

    cpuPlayer := playerlib.MakePlayer(setup.WizardCustom{
        Name: "CPU",
        Banner: data.BannerRed,
    }, false, 0, 0, nil, &playerlib.NoGlobalEnchantments{})

    humanPlayer := playerlib.MakePlayer(setup.WizardCustom{
        Name: "Human",
        Banner: data.BannerGreen,
    }, true, 0, 0, nil, &playerlib.NoGlobalEnchantments{})

    humanPlayer.CastingSkillPower = 10000
    humanPlayer.Mana = 1000

    for _, spell := range allSpells.Spells {
        humanPlayer.KnownSpells.AddSpell(spell)
    }

    defendingArmy := combat.Army{
        // Player: cpuPlayer,
        Player: humanPlayer,
    }

    makeHero := func (unit *units.OverworldUnit) *herolib.Hero {
        for _, hero := range herolib.AllHeroTypes() {
            heroUnit := hero.GetUnit()
            if heroUnit.Equals(unit.Unit) {
                newHero := herolib.MakeHero(unit, hero, hero.DefaultName())
                newHero.SetExtraAbilities()
                return newHero
            }
        }

        log.Printf("Unknown hero unit: %v", unit.Unit)

        return nil
    }

    for _, unit := range combatDescription.DefenderUnits {
        made := units.MakeOverworldUnitFromUnit(unit, 1, 1, data.PlaneArcanus, cpuPlayer.Wizard.Banner, cpuPlayer.MakeExperienceInfo(), cpuPlayer.MakeUnitEnchantmentProvider())

        if made.GetRace() == data.RaceHero {
            defendingArmy.AddUnit(makeHero(made))
        } else {
            defendingArmy.AddUnit(made)
        }
    }

    defendingArmy.LayoutUnits(combat.TeamDefender)

    attackingArmy := combat.Army{
        // Player: humanPlayer,
        Player: cpuPlayer,
    }

    for _, unit := range combatDescription.AttackerUnits {
        made := units.MakeOverworldUnitFromUnit(unit, 1, 1, data.PlaneArcanus, humanPlayer.Wizard.Banner, humanPlayer.MakeExperienceInfo(), humanPlayer.MakeUnitEnchantmentProvider())
        // made.AddExperience(200)

        if made.GetRace() == data.RaceHero {
            attackingArmy.AddUnit(makeHero(made))
        } else {
            attackingArmy.AddUnit(made)
        }
    }

    attackingArmy.LayoutUnits(combat.TeamAttacker)

    combatScreen := combat.MakeCombatScreen(engine.Cache, &defendingArmy, &attackingArmy, humanPlayer, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, data.MagicNone, 0, 0)
    engine.Combat = combatScreen

    run := func(yield coroutine.YieldFunc) error {
        for combatScreen.Update(yield) == combat.CombatStateRunning {
            yield()
        }

        for _, entry := range combatScreen.Model.Log {
            log.Printf("%+v", entry)
        }

        return CombatDoneErr
    }

    // make sure inputmanager is updated at least once first
    inputmanager.Update()
    engine.Coroutine = coroutine.MakeCoroutine(run)
}

func enlargeTransform(factor int) util.ImageTransformFunc {
    var f util.ImageTransformFunc

    f = func (original *image.Paletted) image.Image {
        newImage := image.NewPaletted(image.Rect(0, 0, original.Bounds().Dx() * factor, original.Bounds().Dy() * factor), original.Palette)

        for y := 0; y < original.Bounds().Dy(); y++ {
            for x := 0; x < original.Bounds().Dx(); x++ {
                colorIndex := original.ColorIndexAt(x, y)

                for dy := 0; dy < factor; dy++ {
                    for dx := 0; dx < factor; dx++ {
                        newImage.SetColorIndex(x * factor + dx, y * factor + dy, colorIndex)
                    }
                }
            }
        }

        return newImage
    }

    return f
}

func scaleImage(img *ebiten.Image, newHeight int) *ebiten.Image {
    scale := float64(newHeight) / float64(img.Bounds().Dy())
    newWidth := int(float64(img.Bounds().Dx()) * scale)

    newImage := ebiten.NewImage(newWidth, newHeight)
    var options ebiten.DrawImageOptions
    options.GeoM.Scale(scale, scale)
    newImage.DrawImage(img, &options)
    return newImage
}

func (engine *Engine) MakeBugUI() *ebitenui.UI {
    face, _ := loadFont(19)

    backgroundImage := ui_image.NewNineSliceColor(color.NRGBA{R: 32, G: 32, B: 32, A: 128})

    rootContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(12),
            widget.RowLayoutOpts.Padding(padding(5)),
        )),
        widget.ContainerOpts.BackgroundImage(backgroundImage),
        // widget.ContainerOpts.BackgroundImage(backgroundImageNine),
    )

    rootContainer.AddChild(widget.NewText(
        widget.TextOpts.Text("Press ESC to return to combat", face, color.White),
    ))

    rootContainer.AddChild(makeRow(5,
        widget.NewButton(
            widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x2d, G: 0xbf, B: 0x5a, A: 0xff})),
            widget.ButtonOpts.Text("Back to combat", face, &widget.ButtonTextColor{
                Idle: color.White,
                Hover: color.White,
                Pressed: color.White,
            }),
            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                engine.Mode = EngineModeCombat
            }),
        ),

        widget.NewButton(
            widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xab, G: 0x3e, B: 0x2e, A: 0xff})),
            widget.ButtonOpts.Text("Exit to Main Menu", face, &widget.ButtonTextColor{
                Idle: color.White,
                Hover: color.White,
                Pressed: color.White,
            }),
            /*
            widget.ButtonOpts.TextAndImage("Exit Combat", face, &widget.ButtonImageImage{Idle: rescaled, Disabled: raceImage}, &widget.ButtonTextColor{
                Idle: color.White,
                Hover: color.White,
                Pressed: color.White,
            }),
            */
            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                engine.Mode = EngineModeMenu
            }),
        ),
    ))

    rootContainer.AddChild(space(30))

    rootContainer.AddChild(widget.NewText(
        widget.TextOpts.Text("Report a bug", face, color.White),
    ))

    inputContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(12),
            widget.RowLayoutOpts.Padding(padding(5)),
        )),
        widget.ContainerOpts.BackgroundImage(makeBorderOutline(color.White)),
    )

    bugText := widget.NewTextInput(
        widget.TextInputOpts.Color(&widget.TextInputColor{
            Idle: color.White,
            Disabled: color.White,
            Caret: color.White,
            DisabledCaret: color.White,
        }),
        widget.TextInputOpts.CaretOpts(
            widget.CaretOpts.Color(color.White),
            widget.CaretOpts.Size(face, 3),
        ),
        widget.TextInputOpts.Face(face),
        widget.TextInputOpts.Placeholder("Type here"),
        widget.TextInputOpts.WidgetOpts(
            widget.WidgetOpts.MinSize(800, 0),
        ),
    )

    doSendBugReport := func(callback func(error)) {
        info := strings.TrimSpace(bugText.GetText())
        extraInfo := engine.CombatDescription.String()

        if len(info) > 1000 {
            info = info[:1000]
        }

        if len(extraInfo) > 5000 {
            extraInfo = extraInfo[:5000]
        }

        go func(){
            err := sendReport("Master of magic combat simulator bug report:\n" + info + "\n\n" + extraInfo)
            if err != nil {
                log.Printf("Error sending report: %v", err)
            }

            callback(err)
        }()
    }

    inputContainer.AddChild(bugText)

    rootContainer.AddChild(inputContainer)

    rootContainer.AddChild(widget.NewButton(
        widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
        widget.ButtonOpts.Text("Send bug report", face, &widget.ButtonTextColor{
            Idle: color.White,
            Hover: color.White,
            Pressed: color.White,
        }),
        widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
            if strings.TrimSpace(bugText.GetText()) == "" {
                return
            }

            log.Printf("Sending bug report")
            // reset event state to remove previous click handlers (the one running right now)
            args.Button.ClickedEvent = &event.Event{}
            args.Button.Text().Label = "Sending..."
            width := 40
            height := 40
            border := 3
            col := lighten(color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff}, -30)
            nine := makeNineImage(makeRoundedButtonImage(width, height, border, col), border)
            args.Button.Image = &widget.ButtonImage{
                Idle: nine,
                Hover: nine,
                Pressed: nine,
            }

            doSendBugReport(func (err error){
                if err == nil {
                    engine.UIUpdates <- func(){
                        args.Button.Text().Label = "Sent successfully!"
                    }
                } else {
                    engine.UIUpdates <- func(){
                        args.Button.Text().Label = "Failed to send!"
                    }
                }
            })
        }),
    ))

    rootContainer.AddChild(space(30))

    rootContainer.AddChild(widget.NewText(
        widget.TextOpts.Text("Combat Description", face, color.White),
    ))
    rootContainer.AddChild(widget.NewTextArea(
        widget.TextAreaOpts.Text(engine.CombatDescription.String()),
        widget.TextAreaOpts.FontFace(face),
        widget.TextAreaOpts.FontColor(color.White),
        widget.TextAreaOpts.ShowVerticalScrollbar(),
        widget.TextAreaOpts.ContainerOpts(
            widget.ContainerOpts.WidgetOpts(
                widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                    MaxHeight: 400,
                }),
                widget.WidgetOpts.MinSize(400, 400),
            ),
        ),
        widget.TextAreaOpts.ScrollContainerOpts(
            widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
                Idle: backgroundImage,
                Disabled: backgroundImage,
                Mask: backgroundImage,
            }),
        ),
        widget.TextAreaOpts.SliderOpts(
            widget.SliderOpts.Images(&widget.SliderTrackImage{
                    Idle: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                    Hover: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                },
                makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xad, G: 0x8d, B: 0x55, A: 0xff}),
            ),
        ),
    ))

    ui := ebitenui.UI{
        Container: rootContainer,
    }

    return &ui

}

func (engine *Engine) MakeUI() *ebitenui.UI {
    imageCache := util.MakeImageCache(engine.Cache)

    face, _ := loadFont(19)

    backgroundImageRaw, _, err := ebitenutil.NewImageFromFileSystem(assets, "assets/box.png")
    if err != nil {
        log.Printf("Could not load box.png: %v", err)
        return nil
    }

    backgroundImageNine := ui_image.NewNineSliceSimple(backgroundImageRaw, 1, 39)

    _ = backgroundImageNine

    backgroundImage := ui_image.NewNineSliceColor(color.NRGBA{R: 32, G: 32, B: 32, A: 255})

    rootContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(12),
            widget.RowLayoutOpts.Padding(widget.Insets{Top: 10, Left: 10, Right: 10}),
        )),
        widget.ContainerOpts.BackgroundImage(backgroundImage),
        // widget.ContainerOpts.BackgroundImage(backgroundImageNine),
    )

    var label1 *widget.Text

    label1 = widget.NewText(
        widget.TextOpts.Text("Master of Magic Combat Simulator", face, color.White),
        widget.TextOpts.WidgetOpts(
            widget.WidgetOpts.CursorEnterHandler(func(args *widget.WidgetCursorEnterEventArgs) {
                label1.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
            }),
            widget.WidgetOpts.CursorExitHandler(func(args *widget.WidgetCursorExitEventArgs) {
                label1.Color = color.White
            }),
        ),
    )

    rootContainer.AddChild(label1)

    fakeImage := ui_image.NewNineSliceColor(color.NRGBA{R: 32, G: 32, B: 32, A: 255})

    /*
    unitList1 := widget.NewListComboButton(
        widget.ListComboButtonOpts.SelectComboButtonOpts(
            widget.SelectComboButtonOpts.ComboButtonOpts(
                widget.ComboButtonOpts.MaxContentHeight(200),
                widget.ComboButtonOpts.ButtonOpts(
                    widget.ButtonOpts.Image(&widget.ButtonImage{
                        Idle: buttonImage,
                        Hover: buttonImage,
                        Pressed: buttonImage,
                    }),
                    widget.ButtonOpts.Text("Select Unit", face, &widget.ButtonTextColor{
                        Idle: color.White,
                        Disabled: color.White,
                    }),
                ),
            ),
        ),
        widget.ListComboButtonOpts.EntryLabelFunc(
            func (e any) string {
                return "Button " + e.(string)
            },
            func (e any) string {
                return "List " + e.(string)
            },
        ),
        widget.ListComboButtonOpts.ListOpts(
            widget.ListOpts.EntryFontFace(face),
            widget.ListOpts.Entries([]any{"x", "y", "z", "a", "b"}),
            widget.ListOpts.EntryColor(&widget.ListEntryColor{
                Selected: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
                Unselected: color.NRGBA{R: 0, G: 255, B: 0, A: 255},
            }),
            widget.ListOpts.SliderOpts(
                widget.SliderOpts.Images(&widget.SliderTrackImage{
                    Idle: fakeImage,
                    Hover: fakeImage,
                }, &widget.ButtonImage{
                    Idle: fakeImage,
                    Hover: fakeImage,
                    Pressed: fakeImage,
                }),
            ),
            widget.ListOpts.ScrollContainerOpts(
                widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
                    Idle: fakeImage,
                    Disabled: fakeImage,
                    Mask: fakeImage,
                }),
            ),
        ),
    )
    rootContainer.AddChild(unitList1)
    */

    type UnitItem struct {
        Race data.Race
        Unit units.Unit
    }

    var armyList *widget.List
    var armyCount *widget.Text

    allRaces := append(append(data.ArcanianRaces(), data.MyrranRaces()...), []data.Race{data.RaceFantastic, data.RaceHero, data.RaceAll}...)

    unitListContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(5),
        )),
    )

    var raceButtons []*widget.Button

    makeRaceButton := func (race data.Race, update func()) *widget.Button {
        raceLbx := "backgrnd.lbx"
        raceIndex := 44

        switch race {
            case data.RaceHero:
                raceLbx = "figures1.lbx"
                raceIndex = 2
            case data.RaceAll:
                raceLbx = "figures3.lbx"
                raceIndex = 42
            case data.RaceFantastic:
                raceLbx = "figure11.lbx"
                raceIndex = 115
            case data.RaceLizard:
                raceIndex = 55
            case data.RaceNomad:
                raceIndex = 56
            case data.RaceOrc:
                raceIndex = 57
            case data.RaceTroll:
                raceIndex = 58
            case data.RaceBarbarian:
                raceIndex = 45
            case data.RaceBeastmen:
                raceIndex = 46
            case data.RaceDarkElf:
                raceIndex = 47
            case data.RaceDraconian:
                raceIndex = 48
            case data.RaceDwarf:
                raceIndex = 49
            case data.RaceGnoll:
                raceIndex = 50
            case data.RaceHalfling:
                raceIndex = 51
            case data.RaceHighElf:
                raceIndex = 52
            case data.RaceHighMen:
                raceIndex = 53
            case data.RaceKlackon:
                raceIndex = 54
        }

        raceImage, _ := imageCache.GetImageTransform(raceLbx, raceIndex, 0, "race-button-enlarge", enlargeTransform(2))
        rescaled := scaleImage(raceImage, 30)

        return widget.NewButton(
            widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
            widget.ButtonOpts.TextAndImage(race.String(), face, &widget.ButtonImageImage{Idle: rescaled, Disabled: raceImage}, &widget.ButtonTextColor{
                Idle: color.White,
                Hover: color.White,
                Pressed: color.White,
            }),
            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                update()
            }),
        )
    }

    for _, race := range allRaces {
        clickTimer := make(map[string]uint64)

        unitGraphic := widget.NewGraphic()

        updateGraphic := func (unit units.Unit) {
            unitImage, err := imageCache.GetImageTransform(unit.CombatLbxFile, unit.CombatIndex + 2, 0, "race-enlarge", enlargeTransform(4))
            if err == nil {
                unitGraphic.Image = unitImage
            }
        }

        unitList := widget.NewList(
            widget.ListOpts.EntryFontFace(face),
            widget.ListOpts.SliderOpts(
                widget.SliderOpts.Images(
                    &widget.SliderTrackImage{
                        Idle: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                        Hover: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                    },
                    makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xad, G: 0x8d, B: 0x55, A: 0xff}),
                ),
            ),

            widget.ListOpts.HideHorizontalSlider(),

            widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(
                widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                    MaxHeight: 200,
                }),
                widget.WidgetOpts.MinSize(0, 200),
            )),

            /*
            widget.ListOpts.ContainerOpts(widget.ContainerOpts.BackgroundImage(
                ui_image.NewNineSliceColor(color.NRGBA{R: 128, G: 64, B: 64, A: 255}),
            )),
            */

            widget.ListOpts.EntryLabelFunc(
                func (e any) string {
                    item := e.(*UnitItem)
                    if item.Race == data.RaceFantastic || item.Race == data.RaceHero || item.Race == data.RaceAll {
                        return item.Unit.Name
                    }
                    return fmt.Sprintf("%v %v", item.Race, item.Unit.Name)
                },
            ),

            widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
                entry := args.Entry.(*UnitItem)

                updateGraphic(entry.Unit)

                lastTime, ok := clickTimer[entry.Unit.Name]
                // log.Printf("Entry %v lastTime %v counter %v ok %v", entry, lastTime, engine.Counter, ok)
                if ok && engine.Counter - lastTime < 30 {
                    // log.Printf("  adding %v to defending army", entry)
                    newItem := *entry
                    armyList.AddEntry(&newItem)
                    armyCount.Label = fmt.Sprintf("%v", len(armyList.Entries()))
                    clickTimer[entry.Unit.Name] = engine.Counter + 30
                } else {
                    clickTimer[entry.Unit.Name] = engine.Counter
                }
            }),

            widget.ListOpts.EntryColor(&widget.ListEntryColor{
                Selected: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                Unselected: color.NRGBA{R: 128, G: 128, B: 128, A: 255},
            }),

            widget.ListOpts.ScrollContainerOpts(
                widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
                    Idle: ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 64, B: 64, A: 255}),
                    Disabled: fakeImage,
                    Mask: fakeImage,
                }),
            ),

            widget.ListOpts.AllowReselect(),
        )

        space := widget.NewGraphic(widget.GraphicOpts.Image(ebiten.NewImage(60, 30)))

        raceRow := makeRow(5, space, unitGraphic, unitList)

        // tab.AddChild(makeRow(5, space, unitGraphic, unitList))

        for _, unit := range slices.SortedFunc(slices.Values(units.UnitsByRace(race)), func (a units.Unit, b units.Unit) int {
            return cmp.Compare(a.Name, b.Name)
        }) {
            unitList.AddEntry(&UnitItem{
                Race: race,
                Unit: unit,
            })
        }

        moreButtonsRow := makeRow(5,
            space,
            widget.NewButton(
                widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
                widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
                widget.ButtonOpts.Text("Add", face, &widget.ButtonTextColor{
                    Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                    Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                    Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
                }),
                widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                    entry := unitList.SelectedEntry()
                    if entry != nil {
                        entry := entry.(*UnitItem)
                        newItem := *entry
                        armyList.AddEntry(&newItem)
                        armyCount.Label = fmt.Sprintf("%v", len(armyList.Entries()))
                    }
                }),
            ),
            widget.NewButton(
                widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
                widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
                widget.ButtonOpts.Text("Add Random", face, &widget.ButtonTextColor{
                    Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                    Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                    Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
                }),
                widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                    choices := units.UnitsByRace(race)
                    use := choices[rand.N(len(choices))]
                    newItem := UnitItem{
                        Race: race,
                        Unit: use,
                    }
                    armyList.AddEntry(&newItem)
                    armyCount.Label = fmt.Sprintf("%v", len(armyList.Entries()))
                }),
            ),
        )

        update := func(){
            unitListContainer.RemoveChildren()
            unitListContainer.AddChild(raceRow)
            unitListContainer.AddChild(moreButtonsRow)
        }

        raceButtons = append(raceButtons, makeRaceButton(race, update))
    }

    defendingArmyCount := widget.NewText(widget.TextOpts.Text("0", face, color.NRGBA{R: 255, G: 255, B: 255, A: 255}))

    makeArmyList := func() *widget.List {
        return widget.NewList(
            widget.ListOpts.EntryFontFace(face),

            widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(
                widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                    MaxHeight: 300,
                }),
                widget.WidgetOpts.MinSize(0, 300),
            )),

            widget.ListOpts.SliderOpts(
                widget.SliderOpts.Images(&widget.SliderTrackImage{
                        Idle: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                        Hover: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                    },
                    makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xad, G: 0x8d, B: 0x55, A: 0xff}),
                ),
            ),

            widget.ListOpts.HideHorizontalSlider(),

            widget.ListOpts.EntryLabelFunc(
                func (e any) string {
                    item := e.(*UnitItem)

                    if item.Race == data.RaceFantastic || item.Race == data.RaceAll {
                        return item.Unit.Name
                    }

                    return fmt.Sprintf("%v %v", item.Race, item.Unit.Name)
                },
            ),

            widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
                /*
                entry := args.Entry.(*ListItem)
                fmt.Println("Entry Selected: ", entry.Name)
                */
            }),

            widget.ListOpts.EntryColor(&widget.ListEntryColor{
                Selected: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
                Unselected: color.NRGBA{R: 0, G: 255, B: 0, A: 255},
            }),

            widget.ListOpts.ScrollContainerOpts(
                widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
                    Idle: ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 64, B: 64, A: 255}),
                    Disabled: fakeImage,
                    Mask: fakeImage,
                }),
            ),
        )
    }

    defendingArmyList := makeArmyList()
    defendingArmyName := widget.NewText(
        widget.TextOpts.Text("Defending Army", face, color.NRGBA{R: 255, G: 0, B: 0, A: 255}),
    )

    attackingArmyCount := widget.NewText(widget.TextOpts.Text("0", face, color.NRGBA{R: 255, G: 255, B: 255, A: 255}))
    attackingArmyList := makeArmyList()
    attackingArmyName := widget.NewText(
        widget.TextOpts.Text("Attacking Army", face, color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
    )

    armyList = defendingArmyList
    armyCount = defendingArmyCount

    raceRows := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(5),
        )),
    )

    buttonsPerRow := 9
    for i := 0; i < len(raceButtons); i += buttonsPerRow {
        max := i + buttonsPerRow
        if max >= len(raceButtons) {
            max = len(raceButtons)
        }

        var widgets []widget.PreferredSizeLocateableWidget
        for j := i; j < max; j++ {
            widgets = append(widgets, raceButtons[j])
        }

        raceRows.AddChild(makeRow(5, widgets...))
    }

    raceButtons[0].Click()

    allContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(5),
            widget.RowLayoutOpts.Padding(padding(5)),
        )),
        widget.ContainerOpts.BackgroundImage(
            // makeNineImage(makeRoundedButtonImage(40, 40, 5, color.NRGBA{R: 64, G: 64, B: 64, A: 0xff}), 5),
            makeBorderOutline(color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
        ),
    )

    allContainer.AddChild(raceRows)
    allContainer.AddChild(unitListContainer)

    rootContainer.AddChild(allContainer)

    rootContainer.AddChild(widget.NewButton(
        widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xbc, G: 0x84, B: 0x2f, A: 0xff})),
        widget.ButtonOpts.Text("Add Random Unit", face, &widget.ButtonTextColor{
            Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
            Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
        }),
        widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
            unit := units.AllUnits[rand.N(len(units.AllUnits))]
            newItem := UnitItem{
                Race: unit.Race,
                Unit: unit,
            }

            armyList.AddEntry(&newItem)
            armyCount.Label = fmt.Sprintf("%v", len(armyList.Entries()))
        }),
    ))

    var armyButtons []*widget.Button
    armyButtons = append(armyButtons, widget.NewButton(
        widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
        widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Text("Select Defending Army", face, &widget.ButtonTextColor{
            Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
            Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
        widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
            armyList = defendingArmyList
            armyCount = defendingArmyCount
            defendingArmyName.Color = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
            attackingArmyName.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
        }),
    ))

    armyButtons = append(armyButtons, widget.NewButton(
        widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
        widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Text("Select Attacking Army", face, &widget.ButtonTextColor{
            Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
            Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
        widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
            armyList = attackingArmyList
            armyCount = attackingArmyCount
            attackingArmyName.Color = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
            defendingArmyName.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
        }),
    ))

    var armyRadioElements []widget.RadioGroupElement
    for _, button := range armyButtons {
        armyRadioElements = append(armyRadioElements, button)
    }

    widget.NewRadioGroup(widget.RadioGroupOpts.Elements(armyRadioElements...))

    swapArmies := widget.NewButton(
        widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x5a, G: 0xbc, B: 0x3c, A: 0xff})),
        widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Text("Swap armies", face, &widget.ButtonTextColor{
            Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
            Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
        widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
            defenders := defendingArmyList.Entries()
            attackers := attackingArmyList.Entries()

            defendingArmyList.SetEntries(attackers)
            defendingArmyCount.Label = fmt.Sprintf("%v", len(defendingArmyList.Entries()))
            attackingArmyList.SetEntries(defenders)
            attackingArmyCount.Label = fmt.Sprintf("%v", len(attackingArmyList.Entries()))
        }),
    )

    rootContainer.AddChild(makeRow(10, armyButtons[0], armyButtons[1], swapArmies))

    armyContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
            widget.RowLayoutOpts.Spacing(30),
        )),
    )

    defendingArmyContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(5),
            widget.RowLayoutOpts.Padding(padding(5)),
        )),
        widget.ContainerOpts.BackgroundImage(
            makeBorderOutline(color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
        ),
    )

    defendingArmyContainer.AddChild(makeRow(4, defendingArmyName, defendingArmyCount))

    defendingArmyContainer.AddChild(defendingArmyList)
    defendingArmyContainer.AddChild(makeRow(4,
        widget.NewButton(
            widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
            widget.ButtonOpts.Text("Clear Defending Army", face, &widget.ButtonTextColor{
                Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            }),
            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                defendingArmyList.SetEntries(nil)
                defendingArmyCount.Label = "0"
            }),
        ),
        widget.NewButton(
            widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
            widget.ButtonOpts.Text("Remove Selected Unit", face, &widget.ButtonTextColor{
                Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            }),
            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                selected := defendingArmyList.SelectedEntry()
                if selected != nil {
                    defendingArmyList.RemoveEntry(selected)
                    defendingArmyCount.Label = fmt.Sprintf("%v", len(defendingArmyList.Entries()))
                }
            }),
        ),
    ))

    armyContainer.AddChild(defendingArmyContainer)

    attackingArmyContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(5),
            widget.RowLayoutOpts.Padding(padding(5)),
        )),
        widget.ContainerOpts.BackgroundImage(
            makeBorderOutline(color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
        ),
    )

    attackingArmyContainer.AddChild(makeRow(4, attackingArmyName, attackingArmyCount))

    attackingArmyContainer.AddChild(attackingArmyList)
    attackingArmyContainer.AddChild(makeRow(4,
        widget.NewButton(
            widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
            widget.ButtonOpts.Text("Clear Attacking Army", face, &widget.ButtonTextColor{
                Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            }),
            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                attackingArmyList.SetEntries(nil)
                attackingArmyCount.Label = "0"
            }),
        ),
        widget.NewButton(
            widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
            widget.ButtonOpts.Text("Remove Selected Unit", face, &widget.ButtonTextColor{
                Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            }),
            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                selected := attackingArmyList.SelectedEntry()
                if selected != nil {
                    attackingArmyList.RemoveEntry(selected)
                    attackingArmyCount.Label = fmt.Sprintf("%v", len(attackingArmyList.Entries()))
                }
            }),
        ),
    ))

    armyContainer.AddChild(attackingArmyContainer)

    rootContainer.AddChild(armyContainer)

    combatPicture, _ := imageCache.GetImageTransform("special.lbx", 29, 0, "combat-enlarge", enlargeTransform(2))
    randomCombatPicture, _ := imageCache.GetImageTransform("special.lbx", 32, 0, "combat-enlarge", enlargeTransform(2))
    savePicture, _ := imageCache.GetImageTransform("compix.lbx", 11, 0, "save-enlarge", enlargeTransform(2))
    loadPicture, _ := imageCache.GetImageTransform("compix.lbx", 13, 0, "save-enlarge", enlargeTransform(2))
    rootContainer.AddChild(makeRow(5,
        widget.NewButton(
            widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x2d, G: 0xbf, B: 0x5a, A: 0xff})),
            widget.ButtonOpts.TextAndImage("Enter Combat!", face, &widget.ButtonImageImage{Idle: combatPicture, Disabled: combatPicture}, &widget.ButtonTextColor{
                Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
            }),

            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                var defenders []units.Unit

                for _, entry := range defendingArmyList.Entries() {
                    defenders = append(defenders, entry.(*UnitItem).Unit)
                }

                var attackers []units.Unit

                for _, entry := range attackingArmyList.Entries() {
                    attackers = append(attackers, entry.(*UnitItem).Unit)
                }

                if len(defenders) > 0 && len(attackers) > 0 {
                    engine.EnterCombat(CombatDescription{
                        DefenderUnits: defenders,
                        AttackerUnits: attackers,
                    })
                }
            }),
        ),
        widget.NewButton(
            widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xc9, G: 0x25, B: 0xcd, A: 0xff})),
            widget.ButtonOpts.TextAndImage("Random Combat!", face, &widget.ButtonImageImage{Idle: randomCombatPicture, Disabled: randomCombatPicture}, &widget.ButtonTextColor{
                Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
            }),

            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                var defenders []units.Unit
                var attackers []units.Unit

                for range 3 {
                    defenders = append(defenders, units.AllUnits[rand.N(len(units.AllUnits))])
                    attackers = append(attackers, units.AllUnits[rand.N(len(units.AllUnits))])
                }

                engine.EnterCombat(CombatDescription{
                    DefenderUnits: defenders,
                    AttackerUnits: attackers,
                })
            }),
        ),
        space(30),
        widget.NewButton(
            widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x29, G: 0x9d, B: 0x39, A: 0xff})),
            widget.ButtonOpts.TextAndImage("Save Configuration", face, &widget.ButtonImageImage{Idle: savePicture, Disabled: savePicture}, &widget.ButtonTextColor{
                Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
            }),

            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                var defenders []units.Unit

                for _, entry := range defendingArmyList.Entries() {
                    defenders = append(defenders, entry.(*UnitItem).Unit)
                }

                var attackers []units.Unit

                for _, entry := range attackingArmyList.Entries() {
                    attackers = append(attackers, entry.(*UnitItem).Unit)
                }

                // FIXME: use a file picker widget to select the filename
                filename := "combat-config.json"

                if len(defenders) > 0 && len(attackers) > 0 {
                    description := CombatDescription{
                        DefenderUnits: defenders,
                        AttackerUnits: attackers,
                    }

                    err := description.Save(filename)

                    if err != nil {
                        log.Printf("Error saving configuration: %v", err)
                    } else {
                        log.Printf("Saved configuration to %v", filename)
                    }
                }

            }),
        ),
        widget.NewButton(
            widget.ButtonOpts.TextPadding(widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x29, G: 0x9d, B: 0x7a, A: 0xff})),
            widget.ButtonOpts.TextAndImage("Load Configuration", face, &widget.ButtonImageImage{Idle: loadPicture, Disabled: loadPicture}, &widget.ButtonTextColor{
                Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
            }),

            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                // FIXME: use a file picker widget to select the filename
                description, err := LoadCombatDescription("combat-config.json")
                if err == nil {
                    defendingArmyList.SetEntries(nil)
                    attackingArmyList.SetEntries(nil)

                    for _, unit := range description.DefenderUnits {
                        newItem := UnitItem{
                            Race: unit.Race,
                            Unit: unit,
                        }

                        defendingArmyList.AddEntry(&newItem)
                    }
                    defendingArmyCount.Label = fmt.Sprintf("%v", len(description.DefenderUnits))

                    for _, unit := range description.AttackerUnits {
                        newItem := UnitItem{
                            Race: unit.Race,
                            Unit: unit,
                        }

                        attackingArmyList.AddEntry(&newItem)
                    }
                    attackingArmyCount.Label = fmt.Sprintf("%v", len(description.AttackerUnits))
                } else {
                    log.Printf("Unable to load configuration: %v", err)
                }
            }),
        ),
    ))

    ui := ebitenui.UI{
        Container: rootContainer,
    }

    return &ui
}

func loadFont(size float64) (text.Face, error) {
    source, err := common.LoadFont()

    if err != nil {
        log.Fatal(err)
        return nil, err
    }

    return &text.GoTextFace{
        Source: source,
        Size:   size,
    }, nil
}

func main(){
    cache := lbx.AutoCache()

    audio.Initialize()
    mouse.Initialize()

    engine := MakeEngine(cache)
    ebiten.SetWindowSize(1200, 900)
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    err := ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
