package main

import (
    "embed"
    "log"
    "fmt"
    "image"
    "image/color"
    "slices"
    "strconv"
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
    "github.com/kazzmir/master-of-magic/lib/optional"
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
    HumanDefender bool
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

    defender, ok := output["human_defender"].(bool)
    if !ok {
        defender = true
    }

    return &CombatDescription{
        DefenderUnits: defenders,
        AttackerUnits: attackers,
        HumanDefender: defender,
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

    out["human_defender"] = description.HumanDefender

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
        CombatDescription: CombatDescription{
            HumanDefender: true,
        },
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

    vector.FillRect(img, float32(border), 0, float32(width - border * 2), float32(height), col, false)
    vector.FillRect(img, 0, float32(border), float32(width), float32(height - border * 2), col, false)
    vector.FillCircle(img, float32(border), float32(border), float32(border), col, false)
    vector.FillCircle(img, float32(width-border), float32(border), float32(border), col, false)
    vector.FillCircle(img, float32(border), float32(height-border), float32(border), col, false)
    vector.FillCircle(img, float32(width-border), float32(height-border), float32(border), col, false)

    // img.Fill(col)

    return img
}

func padding(n int) *widget.Insets {
    return &widget.Insets{Top: n, Bottom: n, Left: n, Right: n}
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
    vector.StrokeRect(img, 0, 0, 18, 18, 1, col, false)
    vector.StrokeLine(img, 19, 0, 19, 19, 1, lighten(col, -80), false)
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

func makeColumn(spacing int, children ...widget.PreferredSizeLocateableWidget) *widget.Container {
    container := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
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
                engine.Combat.Cancel()
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

    defendingPlayer := cpuPlayer
    attackingPlayer := humanPlayer
    if combatDescription.HumanDefender {
        defendingPlayer = humanPlayer
        attackingPlayer = cpuPlayer
    }

    defendingArmy := combat.Army{
        Player: defendingPlayer,
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

    attackingArmy := combat.Army{
        Player: attackingPlayer,
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

    model := combat.MakeCombatModel(allSpells, &defendingArmy, &attackingArmy, combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, data.MagicNone, 0, 0, make(chan combat.CombatEvent, 100))
    combatScreen := combat.MakeCombatScreen(engine.Cache, &defendingArmy, &attackingArmy, optional.Of[combat.ArmyPlayer](humanPlayer), combat.CombatLandscapeGrass, data.PlaneArcanus, combat.ZoneType{}, model)
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
        widget.TextOpts.Text("Press ESC to return to combat", &face, color.White),
    ))

    rootContainer.AddChild(makeRow(5,
        widget.NewButton(
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x2d, G: 0xbf, B: 0x5a, A: 0xff})),
            widget.ButtonOpts.Text("Back to combat", &face, &widget.ButtonTextColor{
                Idle: color.White,
                Hover: color.White,
                Pressed: color.White,
            }),
            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                engine.Mode = EngineModeCombat
            }),
        ),

        widget.NewButton(
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xab, G: 0x3e, B: 0x2e, A: 0xff})),
            widget.ButtonOpts.Text("Exit to Main Menu", &face, &widget.ButtonTextColor{
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
                engine.Combat.Cancel()
                engine.Mode = EngineModeMenu
            }),
        ),
    ))

    rootContainer.AddChild(space(30))

    rootContainer.AddChild(widget.NewText(
        widget.TextOpts.Text("Report a bug", &face, color.White),
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
        widget.TextInputOpts.CaretWidth(3),
        widget.TextInputOpts.Face(&face),
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
        widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
        widget.ButtonOpts.Text("Send bug report", &face, &widget.ButtonTextColor{
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
            /*
            width := 40
            height := 40
            border := 3
            col := lighten(color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff}, -30)
            // FIXME: needs Button.SetImage() in ebitenui
            nine := makeNineImage(makeRoundedButtonImage(width, height, border, col), border)
            args.Button.Image = &widget.ButtonImage{
                Idle: nine,
                Hover: nine,
                Pressed: nine,
            }
            */

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
        widget.TextOpts.Text("Combat Description", &face, color.White),
    ))
    rootContainer.AddChild(widget.NewTextArea(
        widget.TextAreaOpts.Text(engine.CombatDescription.String()),
        widget.TextAreaOpts.FontFace(&face),
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
        widget.TextAreaOpts.ScrollContainerImage(&widget.ScrollContainerImage{
            Idle: backgroundImage,
            Disabled: backgroundImage,
            Mask: backgroundImage,
        }),
        widget.TextAreaOpts.SliderParams(&widget.SliderParams{
            TrackImage: &widget.SliderTrackImage{
                    Idle: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                    Hover: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                },
            HandleImage: makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xad, G: 0x8d, B: 0x55, A: 0xff}),
        }),
    ))

    ui := ebitenui.UI{
        Container: rootContainer,
    }

    return &ui

}

func (engine *Engine) MakeUI() *ebitenui.UI {
    imageCache := util.MakeImageCache(engine.Cache)

    face, _ := loadFont(19)

    allSpells, err := spellbook.ReadSpellsFromCache(engine.Cache)
    if err != nil {
        log.Printf("Cannot load spells: %v", err)
        panic(err)
    }

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
            widget.RowLayoutOpts.Padding(&widget.Insets{Top: 10, Left: 10, Right: 10}),
        )),
        widget.ContainerOpts.BackgroundImage(backgroundImage),
        // widget.ContainerOpts.BackgroundImage(backgroundImageNine),
    )

    var label1 *widget.Text

    label1 = widget.NewText(
        widget.TextOpts.Text("Master of Magic Combat Simulator", &face, color.White),
        widget.TextOpts.WidgetOpts(
            widget.WidgetOpts.CursorEnterHandler(func(args *widget.WidgetCursorEnterEventArgs) {
                label1.SetColor(color.RGBA{R: 255, G: 0, B: 0, A: 255})
            }),
            widget.WidgetOpts.CursorExitHandler(func(args *widget.WidgetCursorExitEventArgs) {
                label1.SetColor(color.White)
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
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
            widget.ButtonOpts.TextAndImage(race.String(), &face, &widget.GraphicImage{Idle: rescaled, Disabled: raceImage}, &widget.ButtonTextColor{
                Idle: color.White,
                Hover: color.White,
                Pressed: color.White,
            }),
            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                update()
            }),
        )
    }

    var addWindow func(*widget.Window)

    makeEditUnitWindow := func(unit *units.Unit, onClose func()) *widget.Window {

        transparent := uint8(240)

        contents := widget.NewContainer(
            widget.ContainerOpts.BackgroundImage(ui_image.NewBorderedNineSliceColor(color.NRGBA{R: 64, G: 64, B: 64, A: transparent}, color.NRGBA{R: 200, G: 200, B: 200, A: transparent}, 2)),
            widget.ContainerOpts.Layout(widget.NewRowLayout(
                widget.RowLayoutOpts.Direction(widget.DirectionVertical),
                widget.RowLayoutOpts.Spacing(3),
                widget.RowLayoutOpts.Padding(padding(5)),
            )),
        )

        makeWhiteText := func(label string) *widget.Text {
            return widget.NewText(
                widget.TextOpts.Text(label, &face, color.White),
            )
        }

        five := 5

        makeComboBox := func(entries []any, current any, onSelect func(any)) *widget.ListComboButton {
            return widget.NewListComboButton(
                widget.ListComboButtonOpts.Entries(entries),
                widget.ListComboButtonOpts.MaxContentHeight(150),
                widget.ListComboButtonOpts.WidgetOpts(
                    widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
                        HorizontalPosition: widget.AnchorLayoutPositionCenter,
                        VerticalPosition: widget.AnchorLayoutPositionCenter,
                    }),
                ),
                widget.ListComboButtonOpts.InitialEntry(current),
                widget.ListComboButtonOpts.ButtonParams(&widget.ButtonParams{
                    Image: makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff}),
                    TextPadding: widget.NewInsetsSimple(5),
                    TextFace: &face,
                    TextColor: &widget.ButtonTextColor{
                        Idle: color.White,
                        Disabled: color.White,
                    },
                }),
                widget.ListComboButtonOpts.ListParams(&widget.ListParams{
                    ScrollContainerImage: &widget.ScrollContainerImage{
                        Idle:     ui_image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
                        Disabled: ui_image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
                        Mask:     ui_image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
                    },
                    Slider: &widget.SliderParams{
                        TrackImage: &widget.SliderTrackImage{
                            Idle:  ui_image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
                            Hover: ui_image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
                        },
                        HandleImage: makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xad, G: 0x8d, B: 0x55, A: 0xff}),
                        MinHandleSize: &five,
                        TrackPadding:  widget.NewInsetsSimple(2),
                    },
                    EntryFace: &face,
                    EntryColor: &widget.ListEntryColor{
                        Selected:                   color.NRGBA{254, 255, 255, 255},             //Foreground color for the unfocused selected entry
                        Unselected:                 color.NRGBA{254, 255, 255, 255},             //Foreground color for the unfocused unselected entry
                        SelectedBackground:         color.NRGBA{R: 130, G: 130, B: 200, A: 255}, //Background color for the unfocused selected entry
                        SelectedFocusedBackground:  color.NRGBA{R: 130, G: 130, B: 170, A: 255}, //Background color for the focused selected entry
                        FocusedBackground:          color.NRGBA{R: 170, G: 170, B: 180, A: 255}, //Background color for the focused unselected entry
                        DisabledUnselected:         color.NRGBA{100, 100, 100, 255},             //Foreground color for the disabled unselected entry
                        DisabledSelected:           color.NRGBA{100, 100, 100, 255},             //Foreground color for the disabled selected entry
                        DisabledSelectedBackground: color.NRGBA{100, 100, 100, 255},             //Background color for the disabled selected entry
                    },
                    EntryTextPadding: widget.NewInsetsSimple(5),
                    MinSize:          &image.Point{200, 0},
                }),

                widget.ListComboButtonOpts.EntryLabelFunc(
                    func (e any) string {
                        return fmt.Sprintf("%v", e)
                    },
                    func (e any) string {
                        return fmt.Sprintf("%v", e)
                    },
                ),

                widget.ListComboButtonOpts.EntrySelectedHandler(func(args *widget.ListComboButtonEntrySelectedEventArgs) {
                    onSelect(args.Entry)
                }),
            )
        }

        makeInput := func(placeHolder string, width int, accept func(string) bool, onSet func(string)) *widget.TextInput {
            lastText := placeHolder
            input := widget.NewTextInput(
                widget.TextInputOpts.WidgetOpts(
                    // Set the layout information to center the textbox in the parent
                    widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                        Position: widget.RowLayoutPositionCenter,
                        Stretch:  true,
                    }),
                    widget.WidgetOpts.MinSize(width * 10, 0),
                ),

                // Set the keyboard type when opened on mobile devices.
                // widget.TextInputOpts.MobileInputMode(mobile.TEXT),

                // Set the Idle and Disabled background image for the text input.
                // If the NineSlice image has a minimum size, the widget will use that or
                // widget.WidgetOpts.MinSize; whichever is greater.
                widget.TextInputOpts.Image(&widget.TextInputImage{
                    Idle:     ui_image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 100, A: 255}),
                    Disabled: ui_image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 100, A: 255}),
                }),

                // Set the font face and size for the widget
                widget.TextInputOpts.Face(&face),

                // Set the colors for the text and caret
                widget.TextInputOpts.Color(&widget.TextInputColor{
                    Idle:          color.NRGBA{254, 255, 255, 255},
                    Disabled:      color.NRGBA{R: 200, G: 200, B: 200, A: 255},
                    Caret:         color.NRGBA{254, 255, 255, 255},
                    DisabledCaret: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
                }),

                // Set how much padding there is between the edge of the input and the text
                widget.TextInputOpts.Padding(widget.NewInsetsSimple(5)),

                // This text is displayed if the input is empty
                // widget.TextInputOpts.Placeholder(placeHolder),

                // This is called when the user hits the "Enter" key.
                // There are other options that can configure this behavior.
                /*
                widget.TextInputOpts.SubmitHandler(func(args *widget.TextInputChangedEventArgs) {
                    fmt.Println("Text Submitted: ", args.InputText)
                }),
                */

                // This is called whenver there is a change to the text
                widget.TextInputOpts.ChangedHandler(func(args *widget.TextInputChangedEventArgs) {
                    if !accept(args.InputText) {
                        args.TextInput.SetText(lastText)
                    } else {
                        lastText = args.InputText
                        onSet(lastText)
                    }
                }),
            )

            input.SetText(placeHolder)

            return input
        }

        makeNumberInput := func (value *int) *widget.TextInput {
            return makeInput(fmt.Sprintf("%d", *value), 4, func(text string) bool {
                if text == "" {
                    return true
                }
                // every character must be a number
                for _, char := range text {
                    if char < '0' || char > '9' {
                        return false
                    }
                }
                return true
            }, func(text string) {
                num, err := strconv.Atoi(text)
                if err == nil {
                    *value = num
                }
            })
        }

        makeTextInput := func(text *string) *widget.TextInput {
            return makeInput(*text, 20, func(text string) bool {
                return true
            }, func(data string) {
                *text = data
            })
        }

        if unit.Name == "" {
            unit.Name = "name"
        }

        contents.AddChild(makeRow(5, makeWhiteText("Name"), makeTextInput(&unit.Name)))

        // add drop down of all races
        var raceEntries []any
        for _, race := range allRaces {
            raceEntries = append(raceEntries, race)
        }

        if unit.Race == data.RaceNone {
            unit.Race = data.RaceBarbarian
        }

        contents.AddChild(makeRow(5, makeWhiteText("Race"), makeComboBox(raceEntries, unit.Race, func (selected any){
            race, ok := selected.(data.Race)
            if ok {
                unit.Race = race
            }
        })))

        makeSquare := func(size int, col color.NRGBA) *ebiten.Image {
            img := ebiten.NewImage(size, size)
            img.Fill(col)
            return img
        }

        makeCheckbox := func(checked bool, changedHandler func(bool)) *widget.Checkbox {
            return widget.NewCheckbox(
                widget.CheckboxOpts.WidgetOpts(
                    widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
                        HorizontalPosition: widget.AnchorLayoutPositionCenter,
                        VerticalPosition: widget.AnchorLayoutPositionCenter,
                    }),
                ),
                widget.CheckboxOpts.Image(&widget.CheckboxImage{
                    Unchecked: ui_image.NewFixedNineSlice(makeSquare(20, color.NRGBA{R: 96, G: 96, B: 96, A: 255})),
                    Checked: ui_image.NewFixedNineSlice(makeSquare(20, color.NRGBA{R: 100, G: 255, B: 100, A: 255})),
                }),
                widget.CheckboxOpts.StateChangedHandler(func (args *widget.CheckboxChangedEventArgs) {
                    switch args.State {
                    case widget.WidgetChecked:
                        changedHandler(true)
                    case widget.WidgetUnchecked:
                        changedHandler(false)
                    }
                }),
                widget.CheckboxOpts.InitialState(func() widget.WidgetState {
                    if checked {
                        return widget.WidgetChecked
                    }
                    return widget.WidgetUnchecked
                }()),
            )
        }

        makeIconText := func(icon *ebiten.Image, label string) *widget.Container {
            return makeRow(3, widget.NewGraphic(widget.GraphicOpts.Image(icon)), makeWhiteText(label))
        }

        makeUnitViewIconText := func(index int, label string) *widget.Container {
            icon, _ := imageCache.GetImageTransform("unitview.lbx", index, 0, "icon-enlarge", enlargeTransform(3))
            return makeIconText(icon, label)
        }

        // Flying: checkbox


        // Magic realm: combo box with all magic realms
        // ranged attack combo box

        inputs := widget.NewContainer(
            widget.ContainerOpts.Layout(widget.NewGridLayout(
                widget.GridLayoutOpts.Columns(2),
                widget.GridLayoutOpts.Spacing(5, 2),
            )),
        )

        var flyingCallbacks []func(bool)
        inputs.AddChild(makeWhiteText("Flying"), makeCheckbox(unit.Flying, func(flying bool){
            unit.Flying = flying
            for _, callback := range flyingCallbacks {
                callback(flying)
            }
        }))

        makeRangedAttackElements := func() []widget.PreferredSizeLocateableWidget {
            damageTypes := []any{
                units.DamageRangedPhysical,
                units.DamageRangedMagical,
                units.DamageRangedBoulder,
            }

            initial := unit.RangedAttackDamageType
            ok := false
            for _, damageType := range damageTypes {
                if damageType == initial {
                    ok = true
                    break
                }
            }

            if !ok {
                initial = damageTypes[0].(units.Damage)
                unit.RangedAttackDamageType = initial
            }

            rangedAttackIndex := func () int {
                switch unit.RangedAttackDamageType {
                    case units.DamageRangedPhysical: return 18
                    case units.DamageRangedMagical: return 14
                    case units.DamageRangedBoulder: return 19
                }
                return 0
            }

            rangedAttackPowerLabel := makeUnitViewIconText(rangedAttackIndex(), "Ranged Attack Power")
            rangedAttacksLabel := makeUnitViewIconText(rangedAttackIndex(), "Ranged Attacks")

            // ranged attack power: text input as number
            rangedAttackPowerRow := makeRow(5, rangedAttackPowerLabel, makeNumberInput(&unit.RangedAttackPower))

            // ranged attacks: text input as number
            rangedAttacksRow := makeRow(5, rangedAttacksLabel, makeNumberInput(&unit.RangedAttacks))

            return []widget.PreferredSizeLocateableWidget{
                makeRow(5, makeWhiteText("Ranged Attack Type"), makeComboBox(damageTypes, initial, func(selected any){
                    damage, ok := selected.(units.Damage)
                    if ok {
                        unit.RangedAttackDamageType = damage
                    }

                    newRangedAttackPowerLabel := makeUnitViewIconText(rangedAttackIndex(), "Ranged Attack Power")
                    newRangedAttacksLabel := makeUnitViewIconText(rangedAttackIndex(), "Ranged Attacks")

                    rangedAttackPowerRow.ReplaceChild(rangedAttackPowerLabel, newRangedAttackPowerLabel)
                    rangedAttacksRow.ReplaceChild(rangedAttacksLabel, newRangedAttacksLabel)

                    rangedAttackPowerLabel = newRangedAttackPowerLabel
                    rangedAttacksLabel = newRangedAttacksLabel
                })),

                rangedAttackPowerRow,

                rangedAttacksRow,
            }
        }

        rangedAttackArea := widget.NewContainer(
            widget.ContainerOpts.Layout(widget.NewRowLayout(
                widget.RowLayoutOpts.Direction(widget.DirectionVertical),
                widget.RowLayoutOpts.Spacing(3),
            )),
        )

        hasRanged := unit.RangedAttackDamageType != units.DamageNone
        inputs.AddChild(makeWhiteText("Range Attack"), makeCheckbox(hasRanged, func(ranged bool){
            hasRanged = ranged
            rangedAttackArea.RemoveChildren()

            if hasRanged {
                rangedAttackArea.AddChild(makeRangedAttackElements()...)
            } else {
                unit.RangedAttackDamageType = units.DamageNone
            }
        }))

        if hasRanged {
            rangedAttackArea.AddChild(makeRangedAttackElements()...)
        }

        inputs.AddChild(rangedAttackArea, space(1))

        // count: combo box, 1-8
        inputs.AddChild(makeWhiteText("Count"), makeNumberInput(&unit.Count))

        // movement speed: text input as number
        flyingIndex := 25
        walkingIndex := 24
        movementIcon := makeUnitViewIconText(func () int {
            if unit.Flying {
                return flyingIndex
            }
            return walkingIndex
        }(), "Movement Speed")
        inputs.AddChild(movementIcon, makeNumberInput(&unit.MovementSpeed))

        flyingCallbacks = append(flyingCallbacks, func(flying bool){
            var newWidget *widget.Container
            if flying {
                newWidget = makeUnitViewIconText(flyingIndex, "Movement Speed")
            } else {
                newWidget = makeUnitViewIconText(walkingIndex, "Movement Speed")
            }
            inputs.ReplaceChild(movementIcon, newWidget)
            movementIcon = newWidget
        })

        // melee attack power: text input as number
        inputs.AddChild(makeUnitViewIconText(13, "Melee Attack Power"), makeNumberInput(&unit.MeleeAttackPower))

        // defense: text input as number
        inputs.AddChild(makeUnitViewIconText(22, "Defense"), makeNumberInput(&unit.Defense))

        // resistance: text input as number
        inputs.AddChild(makeUnitViewIconText(27, "Resistance"), makeNumberInput(&unit.Resistance))

        // hit points: text input as number
        inputs.AddChild(makeUnitViewIconText(23, "Hit Points"), makeNumberInput(&unit.HitPoints))

        contents.AddChild(inputs)

        // spells: set of spells this unit can cast
        // abilities: set of abilities this unit has

        titleContainer := widget.NewContainer(
            widget.ContainerOpts.BackgroundImage(ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 64, B: 64, A: transparent})),
            widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
        )
        titleContainer.AddChild(widget.NewText(
            widget.TextOpts.Text("Edit Unit", &face, color.White),
            widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
                HorizontalPosition: widget.AnchorLayoutPositionCenter,
                VerticalPosition: widget.AnchorLayoutPositionCenter,
            })),
        ))

        var window *widget.Window

        /*
        contents.AddChild(widget.NewButton(
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
                widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x00, G: 0xf0, B: 0x00, A: 0xff})),
                widget.ButtonOpts.Text("Add Unit", &face, &widget.ButtonTextColor{
                    Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                    Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                    Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
                }),
                widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                    window.Close()
                }),
            ),
        )

        contents.AddChild(space(10))
        */

        var abilityList *widget.List
        var availableAbilityList *widget.List

        makeEditAbility := func(ability *data.Ability) *widget.Container {

            var row *widget.Container
            text := makeWhiteText(ability.Name())

            update := func() {
                newText := makeWhiteText(ability.Name())
                row.ReplaceChild(text, newText)
                text = newText

                for i := range unit.Abilities {
                    if unit.Abilities[i].Ability == ability.Ability {
                        unit.Abilities[i] = *ability
                    }
                }

            }

            up := widget.NewButton(
                widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
                widget.ButtonOpts.Image(makeNineRoundedButtonImage(20, 20, 5, color.NRGBA{R: 0x00, G: 0xf0, B: 0x00, A: 0xff})),
                widget.ButtonOpts.Text("+", &face, &widget.ButtonTextColor{
                    Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                    Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                    Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
                }),
                widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs){
                    ability.Value += 1
                    update()
                }),
            )

            down := widget.NewButton(
                widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
                widget.ButtonOpts.Image(makeNineRoundedButtonImage(20, 20, 5, color.NRGBA{R: 0x00, G: 0xf0, B: 0x00, A: 0xff})),
                widget.ButtonOpts.Text("-", &face, &widget.ButtonTextColor{
                    Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                    Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                    Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
                }),
                widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs){
                    ability.Value -= 1
                    update()
                }),
            )

            row = makeRow(5, text, up, down)
            return row
        }

        var abilitiesContainer *widget.Container
        editAbilityContainer := space(1)

        abilityListTimer := make(map[data.AbilityType]uint64)
        abilityList = widget.NewList(
            widget.ListOpts.EntryFontFace(&face),
            widget.ListOpts.SliderParams(&widget.SliderParams{
                TrackImage: &widget.SliderTrackImage{
                    Idle: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                    Hover: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                },
                HandleImage: makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xad, G: 0x8d, B: 0x55, A: 0xff}),
            }),

            widget.ListOpts.HideHorizontalSlider(),

            widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(
                widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                    MaxHeight: 120,
                }),
                widget.WidgetOpts.MinSize(0, 50),
            )),

            widget.ListOpts.EntryLabelFunc(
                func (e any) string {
                    return fmt.Sprintf("%v", e)
                },
            ),

            widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
                entry := args.Entry.(*data.Ability)

                newEdit := makeEditAbility(entry)
                abilitiesContainer.ReplaceChild(editAbilityContainer, newEdit)
                editAbilityContainer = newEdit

                lastTime, ok := abilityListTimer[entry.Ability]
                // log.Printf("Entry %v lastTime %v counter %v ok %v", entry, lastTime, engine.Counter, ok)
                if ok && engine.Counter - lastTime < 30 {
                    // log.Printf("  adding %v to defending army", entry)
                    abilityListTimer[entry.Ability] = engine.Counter + 30

                    list := abilityList
                    list.RemoveEntry(args.Entry)
                    availableAbilityList.AddEntry(args.Entry)

                    unit.Abilities = slices.DeleteFunc(unit.Abilities, func(a data.Ability) bool {
                        return a.Ability == entry.Ability
                    })
                } else {
                    abilityListTimer[entry.Ability] = engine.Counter
                }

            }),

            widget.ListOpts.EntryColor(&widget.ListEntryColor{
                Selected: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
                Unselected: color.NRGBA{R: 0, G: 255, B: 0, A: 255},
            }),

            widget.ListOpts.ScrollContainerImage(&widget.ScrollContainerImage{
                Idle: ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 64, B: 64, A: 255}),
                Disabled: fakeImage,
                Mask: fakeImage,
            }),

            widget.ListOpts.AllowReselect(),
        )

        availableAbilityListTimer := make(map[data.AbilityType]uint64)
        availableAbilityList = widget.NewList(
            widget.ListOpts.EntryFontFace(&face),
            widget.ListOpts.SliderParams(&widget.SliderParams{
                TrackImage: &widget.SliderTrackImage{
                    Idle: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                    Hover: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                },
                HandleImage: makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xad, G: 0x8d, B: 0x55, A: 0xff}),
            }),

            widget.ListOpts.HideHorizontalSlider(),

            widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(
                widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                    MaxHeight: 120,
                }),
                widget.WidgetOpts.MinSize(0, 50),
            )),

            widget.ListOpts.EntryLabelFunc(
                func (e any) string {
                    return fmt.Sprintf("%v", e)
                },
            ),

            widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
                entry := args.Entry.(*data.Ability)

                newEdit := makeEditAbility(entry)
                abilitiesContainer.ReplaceChild(editAbilityContainer, newEdit)
                editAbilityContainer = newEdit

                lastTime, ok := availableAbilityListTimer[entry.Ability]
                // log.Printf("Entry %v lastTime %v counter %v ok %v", entry, lastTime, engine.Counter, ok)
                if ok && engine.Counter - lastTime < 30 {
                    // log.Printf("  adding %v to defending army", entry)
                    availableAbilityListTimer[entry.Ability] = engine.Counter + 30

                    list := availableAbilityList
                    list.RemoveEntry(args.Entry)
                    abilityList.AddEntry(args.Entry)

                    unit.Abilities = append(unit.Abilities, *entry)
                } else {
                    availableAbilityListTimer[entry.Ability] = engine.Counter
                }
            }),

            widget.ListOpts.EntryColor(&widget.ListEntryColor{
                Selected: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
                Unselected: color.NRGBA{R: 0, G: 255, B: 0, A: 255},
            }),

            widget.ListOpts.ScrollContainerImage(&widget.ScrollContainerImage{
                Idle: ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 64, B: 64, A: 255}),
                Disabled: fakeImage,
                Mask: fakeImage,
            }),

            widget.ListOpts.AllowReselect(),
        )

        transferButtons := widget.NewContainer(
            widget.ContainerOpts.Layout(widget.NewRowLayout(
                widget.RowLayoutOpts.Direction(widget.DirectionVertical),
                widget.RowLayoutOpts.Spacing(5),
                widget.RowLayoutOpts.Padding(&widget.Insets{Top: 20, Bottom: 20}),
            )),
        )

        makeArrowImage := func(size int, fill color.NRGBA, edge color.NRGBA) *ebiten.Image {
            img := ebiten.NewImage(size, size)

            // img.Fill(color.NRGBA{R: 255, A: 255})

            var path vector.Path

            fsize := float32(size)

            x1 := float32(1)
            y1 := fsize * 0.30

            x2 := fsize * 0.6
            y2 := y1

            x3 := x2
            y3 := fsize * 0.1

            x4 := fsize
            y4 := fsize * 0.5

            x5 := x3
            y5 := fsize * 0.9

            x6 := x2
            y6 := fsize * 0.70

            x7 := x1
            y7 := y6

            path.MoveTo(x1, y1)
            path.LineTo(x2, y2)
            path.LineTo(x3, y3)
            path.LineTo(x4, y4)
            path.LineTo(x5, y5)
            path.LineTo(x6, y6)
            path.LineTo(x7, y7)
            path.Close()

            var edgeScale ebiten.ColorScale
            edgeScale.ScaleWithColor(edge)

            var fillScale ebiten.ColorScale
            fillScale.ScaleWithColor(fill)

            vector.FillPath(img, &path, nil, &vector.DrawPathOptions{
                ColorScale: fillScale,
            })

            vector.StrokePath(img, &path, &vector.StrokeOptions{
                Width: 1,
            }, &vector.DrawPathOptions{
                ColorScale: edgeScale,
            })

            return img
        }

        makeLeftArrow := func(size int) *widget.ButtonImage {
            mk := func (fill color.NRGBA, edge color.NRGBA) *ui_image.NineSlice {
                img1 := makeArrowImage(size, fill, edge)
                img := ebiten.NewImage(size, size)
                op := &ebiten.DrawImageOptions{}
                op.GeoM.Scale(-1, 1)
                op.GeoM.Translate(float64(size), 0)
                img.DrawImage(img1, op)
                return ui_image.NewNineSliceSimple(img, 3, size - 3)
            }

            return &widget.ButtonImage{
                Idle:  mk(color.NRGBA{R: 128, G: 128, B: 128, A: 255}, color.NRGBA{R: 164, G: 164, B: 164, A: 255}),
                Hover: mk(color.NRGBA{R: 192, G: 192, B: 192, A: 255}, color.NRGBA{R: 200, G: 200, B: 200, A: 255}),
                Pressed: mk(color.NRGBA{R: 255, G: 255, B: 255, A: 255}, color.NRGBA{R: 200, G: 200, B: 200, A: 255}),
            }
        }

        makeRightArrow := func(size int) *widget.ButtonImage {
            return &widget.ButtonImage{
                Idle:  ui_image.NewNineSliceSimple(makeArrowImage(size, color.NRGBA{R: 128, G: 128, B: 128, A: 255}, color.NRGBA{R: 164, G: 164, B: 164, A: 255}), 3, size - 3),
                Hover: ui_image.NewNineSliceSimple(makeArrowImage(size, color.NRGBA{R: 192, G: 192, B: 192, A: 255}, color.NRGBA{R: 200, G: 200, B: 200, A: 255}), 3, size - 3),
                Pressed: ui_image.NewNineSliceSimple(makeArrowImage(size, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, color.NRGBA{R: 200, G: 200, B: 200, A: 255}), 3, size - 3),
            }
        }

        // available => abilities
        transferButtons.AddChild(widget.NewButton(
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeLeftArrow(30)),
            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {

                availableEntry := availableAbilityList.SelectedEntry()
                if availableEntry != nil {
                    availableAbilityList.RemoveEntry(availableEntry)
                    abilityList.AddEntry(availableEntry)

                    entry := availableEntry.(*data.Ability)
                    unit.Abilities = append(unit.Abilities, *entry)
                }
            }),
        ))

        // abilities => available
        transferButtons.AddChild(widget.NewButton(
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeRightArrow(30)),
            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                abilityEntry := abilityList.SelectedEntry()
                if abilityEntry != nil {
                    abilityList.RemoveEntry(abilityEntry)
                    availableAbilityList.AddEntry(abilityEntry)

                    entry := abilityEntry.(*data.Ability)

                    unit.Abilities = slices.DeleteFunc(unit.Abilities, func(a data.Ability) bool {
                        return a.Ability == entry.Ability
                    })
                }
            }),
        ))


        hasAbilities := make(map[data.AbilityType]struct{})

        for _, ability := range unit.Abilities {
            abilityList.AddEntry(&ability)
            hasAbilities[ability.Ability] = struct{}{}
        }

        for _, abilityType := range data.AllAbilities() {
            _, has := hasAbilities[abilityType]
            if !has {
                ability := data.MakeAbility(abilityType)
                availableAbilityList.AddEntry(&ability)
            }
        }

        abilitiesContainer = widget.NewContainer(
            widget.ContainerOpts.Layout(widget.NewRowLayout(
                widget.RowLayoutOpts.Direction(widget.DirectionVertical),
                widget.RowLayoutOpts.Spacing(5),
                widget.RowLayoutOpts.Padding(padding(5)),
            )),
            widget.ContainerOpts.BackgroundImage(ui_image.NewBorderedNineSliceColor(color.NRGBA{A: 0}, color.NRGBA{R: 128, G: 128, B: 128, A: 255}, 2)),
        )

        abilitiesContainer.AddChild(
            makeRow(20, 
                makeColumn(5, makeWhiteText("Abilities"), abilityList),
                makeColumn(5, transferButtons),
                makeColumn(5, makeWhiteText("Available Abilities"), availableAbilityList),
            ),
        )

        abilitiesContainer.AddChild(editAbilityContainer)

        spellsContainer := widget.NewContainer(
            widget.ContainerOpts.Layout(widget.NewRowLayout(
                widget.RowLayoutOpts.Direction(widget.DirectionVertical),
                widget.RowLayoutOpts.Spacing(5),
                widget.RowLayoutOpts.Padding(padding(5)),
            )),
            widget.ContainerOpts.BackgroundImage(ui_image.NewBorderedNineSliceColor(color.NRGBA{A: 0}, color.NRGBA{R: 128, G: 128, B: 128, A: 255}, 2)),
        )

        var spellsList *widget.List
        var spellsAvailableList *widget.List

        transferSpellToAvailable := func() {
            spellEntry := spellsList.SelectedEntry()
            if spellEntry != nil {
                spellsList.RemoveEntry(spellEntry)
                spellsAvailableList.AddEntry(spellEntry)

                entry := spellEntry.(string)

                unit.Spells = slices.DeleteFunc(unit.Spells, func(a string) bool {
                    return a == entry
                })
            }
        }

        transferAvailableToSpell := func() {
            spellEntry := spellsAvailableList.SelectedEntry()
            if spellEntry != nil {
                spellsAvailableList.RemoveEntry(spellEntry)
                spellsList.AddEntry(spellEntry)

                entry := spellEntry.(string)
                unit.Spells = append(unit.Spells, entry)
            }
        }

        spellsList = (func() *widget.List {
            doubleClick := make(map[string]uint64)
            out := widget.NewList(
                widget.ListOpts.EntryFontFace(&face),
                widget.ListOpts.SliderParams(&widget.SliderParams{
                    TrackImage: &widget.SliderTrackImage{
                        Idle: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                        Hover: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                    },
                    HandleImage: makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xad, G: 0x8d, B: 0x55, A: 0xff}),
                }),

                widget.ListOpts.HideHorizontalSlider(),

                widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(
                    widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                        MaxHeight: 120,
                    }),
                    widget.WidgetOpts.MinSize(0, 50),
                )),

                widget.ListOpts.EntryLabelFunc(
                    func (e any) string {
                        spell, _ := e.(string)
                        return fmt.Sprintf("%v", spell)
                    },
                ),

                widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
                    entry := args.Entry.(string)

                    lastTime, ok := doubleClick[entry]
                    // log.Printf("Entry %v lastTime %v counter %v ok %v", entry, lastTime, engine.Counter, ok)
                    if ok && engine.Counter - lastTime < 30 {
                        // log.Printf("  adding %v to defending army", entry)
                        doubleClick[entry] = engine.Counter + 30
                        transferSpellToAvailable()
                    } else {
                        doubleClick[entry] = engine.Counter
                    }
                }),

                widget.ListOpts.EntryColor(&widget.ListEntryColor{
                    Selected: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
                    Unselected: color.NRGBA{R: 0, G: 255, B: 0, A: 255},
                }),

                widget.ListOpts.ScrollContainerImage(&widget.ScrollContainerImage{
                    Idle: ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 64, B: 64, A: 255}),
                    Disabled: fakeImage,
                    Mask: fakeImage,
                }),

                widget.ListOpts.AllowReselect(),
            )

            return out
        })()

        spellsAvailableList = (func () *widget.List {
            doubleClick := make(map[string]uint64)
            out := widget.NewList(
                widget.ListOpts.EntryFontFace(&face),
                widget.ListOpts.SliderParams(&widget.SliderParams{
                    TrackImage: &widget.SliderTrackImage{
                        Idle: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                        Hover: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                    },
                    HandleImage: makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xad, G: 0x8d, B: 0x55, A: 0xff}),
                }),

                widget.ListOpts.HideHorizontalSlider(),

                widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(
                    widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                        MaxHeight: 120,
                    }),
                    widget.WidgetOpts.MinSize(0, 50),
                )),

                widget.ListOpts.EntryLabelFunc(
                    func (e any) string {
                        spell, _ := e.(string)
                        return fmt.Sprintf("%v", spell)
                    },
                ),

                widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
                    entry := args.Entry.(string)

                    lastTime, ok := doubleClick[entry]
                    // log.Printf("Entry %v lastTime %v counter %v ok %v", entry, lastTime, engine.Counter, ok)
                    if ok && engine.Counter - lastTime < 30 {
                        // log.Printf("  adding %v to defending army", entry)
                        doubleClick[entry] = engine.Counter + 30
                        transferAvailableToSpell()
                    } else {
                        doubleClick[entry] = engine.Counter
                    }
                }),

                widget.ListOpts.EntryColor(&widget.ListEntryColor{
                    Selected: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
                    Unselected: color.NRGBA{R: 0, G: 255, B: 0, A: 255},
                }),

                widget.ListOpts.ScrollContainerImage(&widget.ScrollContainerImage{
                    Idle: ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 64, B: 64, A: 255}),
                    Disabled: fakeImage,
                    Mask: fakeImage,
                }),

                widget.ListOpts.AllowReselect(),
            )

            return out
        })()

        spellTransferButtons := widget.NewContainer(
            widget.ContainerOpts.Layout(widget.NewRowLayout(
                widget.RowLayoutOpts.Direction(widget.DirectionVertical),
                widget.RowLayoutOpts.Spacing(5),
                widget.RowLayoutOpts.Padding(&widget.Insets{Top: 20, Bottom: 20}),
            )),
        )

        // available spells => spells
        spellTransferButtons.AddChild(widget.NewButton(
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeLeftArrow(30)),
            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                transferAvailableToSpell()
            }),
        ))

        // spells => available spells
        spellTransferButtons.AddChild(widget.NewButton(
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeRightArrow(30)),
            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                transferSpellToAvailable()
            }),
        ))

        hasSpells := make(map[string]struct{})
        for _, spell := range unit.Spells {
            spellsList.AddEntry(spell)
            hasSpells[spell] = struct{}{}
        }

        for _, spell := range allSpells.Spells {
            _, has := hasSpells[spell.Name]
            if !has {
                spellsAvailableList.AddEntry(spell.Name)
            }
        }

        spellsContainer.AddChild(
            makeRow(20,
                makeColumn(5, makeWhiteText("Spells"), spellsList),
                spellTransferButtons,
                makeColumn(5, makeWhiteText("Available Spells"), spellsAvailableList),
            ),
        )

        contents.AddChild(makeRow(5, abilitiesContainer, spellsContainer))

        // close button
        contents.AddChild(widget.NewButton(
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
                widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xf2, G: 0x00, B: 0x00, A: 0xff})),
                widget.ButtonOpts.Text("Close", &face, &widget.ButtonTextColor{
                    Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                    Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                    Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
                }),
                widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                    window.Close()
                    onClose()
                }),
            ),
        )

        window = widget.NewWindow(
            widget.WindowOpts.Contents(contents),
            widget.WindowOpts.TitleBar(titleContainer, 25),
            widget.WindowOpts.Draggable(),
            widget.WindowOpts.Resizeable(),
            widget.WindowOpts.MinSize(500, 800),
        )

        return window
    }

    for _, race := range allRaces {
        clickTimer := make(map[string]uint64)

        var graphicOptions widget.GraphicOptions
        unitGraphic := widget.NewGraphic(graphicOptions.Image(ebiten.NewImage(60, 30)))

        updateGraphic := func (unit units.Unit) {
            unitImage, err := imageCache.GetImageTransform(unit.CombatLbxFile, unit.CombatIndex + 2, 0, "race-enlarge", enlargeTransform(4))
            if err == nil {
                unitGraphic.Image = unitImage
            }
        }

        unitList := widget.NewList(
            widget.ListOpts.EntryFontFace(&face),
            widget.ListOpts.SliderParams(&widget.SliderParams{
                TrackImage: &widget.SliderTrackImage{
                    Idle: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                    Hover: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                },
                HandleImage: makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xad, G: 0x8d, B: 0x55, A: 0xff}),
            }),

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

            widget.ListOpts.ScrollContainerImage(&widget.ScrollContainerImage{
                Idle: ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 64, B: 64, A: 255}),
                Disabled: fakeImage,
                Mask: fakeImage,
            }),

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
                widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
                widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
                widget.ButtonOpts.Text("Add", &face, &widget.ButtonTextColor{
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
                widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
                widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
                widget.ButtonOpts.Text("Add Random", &face, &widget.ButtonTextColor{
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
            widget.NewButton(
                widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
                widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xbf, G: 0xbf, B: 0x00, A: 0xff})),
                widget.ButtonOpts.Text("Edit", &face, &widget.ButtonTextColor{
                    Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                    Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                    Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
                }),
                widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                    entry := unitList.SelectedEntry()
                    if entry != nil {
                        entry := entry.(*UnitItem)
                        addWindow(makeEditUnitWindow(&entry.Unit, func(){}))
                    }
                }),
            ),
            widget.NewButton(
                widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
                widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xbf, G: 0xbf, B: 0x00, A: 0xff})),
                widget.ButtonOpts.Text("Copy", &face, &widget.ButtonTextColor{
                    Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                    Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
                    Pressed: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
                }),
                widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                    entry := unitList.SelectedEntry()
                    if entry != nil {
                        entry := entry.(*UnitItem)
                        clone := entry.Unit.Clone()

                        newItem := UnitItem{
                            Race: entry.Race,
                            Unit: clone,
                        }

                        addWindow(makeEditUnitWindow(&newItem.Unit, func() {
                            unitList.AddEntry(&newItem)
                        }))
                    }
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

    /*
    makeNewUnitButton := func() *widget.Button {
        return widget.NewButton(
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xbf, G: 0xbf, B: 0x00, A: 0xff})),
            widget.ButtonOpts.Text("New Unit", &face, &widget.ButtonTextColor{
                Idle: color.White,
                Hover: color.White,
                Pressed: color.White,
            }),
            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                newUnit := units.UnitNone
                addWindow(makeEditUnitWindow(&newUnit))
            }),
        )
    }

    raceButtons = append(raceButtons, makeNewUnitButton())
    */

    defendingArmyCount := widget.NewText(widget.TextOpts.Text("0", &face, color.NRGBA{R: 255, G: 255, B: 255, A: 255}))

    makeArmyList := func() *widget.List {
        return widget.NewList(
            widget.ListOpts.EntryFontFace(&face),

            widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(
                widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                    MaxHeight: 300,
                }),
                widget.WidgetOpts.MinSize(0, 300),
            )),

            widget.ListOpts.SliderParams(&widget.SliderParams{
                TrackImage: &widget.SliderTrackImage{
                    Idle: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                    Hover: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                },
                HandleImage: makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xad, G: 0x8d, B: 0x55, A: 0xff}),
            }),

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

            widget.ListOpts.ScrollContainerImage(&widget.ScrollContainerImage{
                Idle: ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 64, B: 64, A: 255}),
                Disabled: fakeImage,
                Mask: fakeImage,
            }),
        )
    }

    defendingArmyList := makeArmyList()
    defendingArmyName := widget.NewText(
        widget.TextOpts.Text("Defending Army", &face, color.NRGBA{R: 255, G: 0, B: 0, A: 255}),
    )

    attackingArmyCount := widget.NewText(widget.TextOpts.Text("0", &face, color.NRGBA{R: 255, G: 255, B: 255, A: 255}))
    attackingArmyList := makeArmyList()
    attackingArmyName := widget.NewText(
        widget.TextOpts.Text("Attacking Army", &face, color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
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
        widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xbc, G: 0x84, B: 0x2f, A: 0xff})),
        widget.ButtonOpts.Text("Add Random Unit", &face, &widget.ButtonTextColor{
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
        widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Text("Select Defending Army", &face, &widget.ButtonTextColor{
            Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
            Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
        widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
            armyList = defendingArmyList
            armyCount = defendingArmyCount
            defendingArmyName.SetColor(color.NRGBA{R: 255, G: 0, B: 0, A: 255})
            attackingArmyName.SetColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
        }),
    ))

    armyButtons = append(armyButtons, widget.NewButton(
        widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
        widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Text("Select Attacking Army", &face, &widget.ButtonTextColor{
            Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
            Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
        }),
        widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
            armyList = attackingArmyList
            armyCount = attackingArmyCount
            attackingArmyName.SetColor(color.NRGBA{R: 255, G: 0, B: 0, A: 255})
            defendingArmyName.SetColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
        }),
    ))

    var armyRadioElements []widget.RadioGroupElement
    for _, button := range armyButtons {
        armyRadioElements = append(armyRadioElements, button)
    }

    widget.NewRadioGroup(widget.RadioGroupOpts.Elements(armyRadioElements...))

    swapArmies := widget.NewButton(
        widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x5a, G: 0xbc, B: 0x3c, A: 0xff})),
        widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
        widget.ButtonOpts.Text("Swap armies", &face, &widget.ButtonTextColor{
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
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
            widget.ButtonOpts.Text("Clear Defending Army", &face, &widget.ButtonTextColor{
                Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            }),
            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                defendingArmyList.SetEntries(nil)
                defendingArmyCount.Label = "0"
            }),
        ),
        widget.NewButton(
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
            widget.ButtonOpts.Text("Remove Selected Unit", &face, &widget.ButtonTextColor{
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
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
            widget.ButtonOpts.Text("Clear Attacking Army", &face, &widget.ButtonTextColor{
                Idle: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
                Hover: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
            }),
            widget.ButtonOpts.ClickedHandler(func (args *widget.ButtonClickedEventArgs) {
                attackingArmyList.SetEntries(nil)
                attackingArmyCount.Label = "0"
            }),
        ),
        widget.NewButton(
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x52, G: 0x78, B: 0xc3, A: 0xff})),
            widget.ButtonOpts.Text("Remove Selected Unit", &face, &widget.ButtonTextColor{
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
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x2d, G: 0xbf, B: 0x5a, A: 0xff})),
            widget.ButtonOpts.TextAndImage("Enter Combat!", &face, &widget.GraphicImage{Idle: combatPicture, Disabled: combatPicture}, &widget.ButtonTextColor{
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
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xc9, G: 0x25, B: 0xcd, A: 0xff})),
            widget.ButtonOpts.TextAndImage("Random Combat!", &face, &widget.GraphicImage{Idle: randomCombatPicture, Disabled: randomCombatPicture}, &widget.ButtonTextColor{
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
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x29, G: 0x9d, B: 0x39, A: 0xff})),
            widget.ButtonOpts.TextAndImage("Save Configuration", &face, &widget.GraphicImage{Idle: savePicture, Disabled: savePicture}, &widget.ButtonTextColor{
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
            widget.ButtonOpts.TextPadding(&widget.Insets{Top: 2, Bottom: 2, Left: 5, Right: 5}),
            widget.ButtonOpts.Image(makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0x29, G: 0x9d, B: 0x7a, A: 0xff})),
            widget.ButtonOpts.TextAndImage("Load Configuration", &face, &widget.GraphicImage{Idle: loadPicture, Disabled: loadPicture}, &widget.ButtonTextColor{
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

    addWindow = func(window *widget.Window) {
        x, y := window.Contents.PreferredSize()
        rect := image.Rect(0, 0, x, y)
        rect = rect.Add(image.Pt(100 + rand.N(50), 100 + rand.N(50)))
        window.SetLocation(rect)
        ui.AddWindow(window)
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
