package main

import (
    "fmt"
    "sync"
    "math"
    "math/rand"
    "strings"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    _ "github.com/hajimehoshi/ebiten/v2/vector"

    "github.com/hajimehoshi/ebiten/v2"
)

type MagicType int

const (
    LifeMagic MagicType = iota
    SorceryMagic
    NatureMagic
    DeathMagic
    ChaosMagic
)

const MaxPicks = 11

/* the number of books a wizard has of a specific magic type */
type wizardBook struct {
    Magic MagicType
    Count int
}

type wizardSlot struct {
    Name string
    // the block that the wizard's name is printed on in the ui
    Background *ebiten.Image
    // the portrait of the wizard shown when the user's cursor is on top of their name
    Portrait *ebiten.Image
    Books []wizardBook
    ExtraAbility WizardAbility
    X int
    Y int
}

type WizardAbility int
const (
        AbilityAlchemy WizardAbility = iota
        AbilityWarlord
        AbilityChanneler
        AbilityArchmage
        AbilityArtificer
        AbilityConjurer
        AbilitySageMaster
        AbilityMyrran
        AbilityDivinePower
        AbilityFamous
        AbilityRunemaster
        AbilityCharismatic
        AbilityChaosMastery
        AbilityNatureMastery
        AbilitySorceryMastery
        AbilityInfernalPower
        AbilityManaFocusing
        AbilityNodeMastery
        AbilityNone
)

// some abilities can only be selected if other properties of the wizard are set
func (ability WizardAbility) SatisifiedDependencies(wizard *wizardCustom) bool {
    switch ability {
        case AbilityAlchemy: return true
        case AbilityWarlord: return true
        case AbilityChanneler: return true
        case AbilityArchmage:
            // need at least 4 books of some magic type
            for _, book := range wizard.Books {
                if book.Count >= 4 {
                    return true
                }
            }
            return false
        case AbilityArtificer: return true
        case AbilityConjurer: return true
        case AbilitySageMaster:
            // need at least 2 books of different magic types
            count := 0
            for _, book := range wizard.Books {
                if book.Count > 0 {
                    count += 1
                }
            }
            return count >= 2
        case AbilityMyrran: return true
        case AbilityDivinePower: return wizard.MagicLevel(LifeMagic) >= 4
        case AbilityFamous: return true
        case AbilityRunemaster:
            // need at least 3 books of different magic types
            count := 0
            for _, book := range wizard.Books {
                if book.Count > 0 {
                    count += 1
                }
            }
            return count >= 3
        case AbilityCharismatic: return true
        case AbilityChaosMastery: return wizard.MagicLevel(ChaosMagic) >= 4
        case AbilityNatureMastery: return wizard.MagicLevel(NatureMagic) >= 4
        case AbilitySorceryMastery: return wizard.MagicLevel(SorceryMagic) >= 4
        case AbilityInfernalPower: return wizard.MagicLevel(DeathMagic) >= 4
        case AbilityManaFocusing:
            // need at least 4 books of some magic type
            for _, book := range wizard.Books {
                if book.Count >= 4 {
                    return true
                }
            }
            return false
        case AbilityNodeMastery:
            // one pick in chaos, nature, and sorcery
            return wizard.MagicLevel(ChaosMagic) >= 1 && wizard.MagicLevel(NatureMagic) >= 1 && wizard.MagicLevel(SorceryMagic) >= 1
        case AbilityNone: return true
    }

    return true
}

func (ability WizardAbility) String() string {
    switch ability {
        case AbilityAlchemy: return "Alchemy"
        case AbilityWarlord: return "Warlord"
        case AbilityChanneler: return "Channeler"
        case AbilityArchmage: return "Archmage"
        case AbilityArtificer: return "Artificer"
        case AbilityConjurer: return "Conjurer"
        case AbilitySageMaster: return "Sage Master"
        case AbilityMyrran: return "Myrran"
        case AbilityDivinePower: return "Divine Power"
        case AbilityFamous: return "Famous"
        case AbilityRunemaster: return "Runemaster"
        case AbilityCharismatic: return "Charismatic"
        case AbilityChaosMastery: return "Chaos Mastery"
        case AbilityNatureMastery: return "Nature Mastery"
        case AbilitySorceryMastery: return "Sorcery Mastery"
        case AbilityInfernalPower: return "Infernal Power"
        case AbilityManaFocusing: return "Mana Focusing"
        case AbilityNodeMastery: return "Node Mastery"
        case AbilityNone: return "invalid"
    }

    return "?"
}

// number of picks this ability costs when choosing a custom wizard
func (ability WizardAbility) PickCost() int {
    switch ability {
        case AbilityAlchemy: return 1
        case AbilityWarlord: return 2
        case AbilityChanneler: return 2
        case AbilityArchmage: return 1
        case AbilityArtificer: return 1
        case AbilityConjurer: return 1
        case AbilitySageMaster: return 1
        case AbilityMyrran: return 3
        case AbilityDivinePower: return 2
        case AbilityFamous: return 2
        case AbilityRunemaster: return 1
        case AbilityCharismatic: return 1
        case AbilityChaosMastery: return 1
        case AbilityNatureMastery: return 1
        case AbilitySorceryMastery: return 1
        case AbilityInfernalPower: return 2
        case AbilityManaFocusing: return 1
        case AbilityNodeMastery: return 1
        case AbilityNone: return 0
    }

    return 1
}

type NewWizardScreenState int

const (
    NewWizardScreenStateSelectWizard NewWizardScreenState = iota
    NewWizardScreenStateCustomPicture
    NewWizardScreenStateCustomName
    NewWizardScreenStateCustomAbility
    NewWizardScreenStateCustomBooks
)

type wizardCustom struct {
    Name string
    Portrait *ebiten.Image
    Abilities []WizardAbility
    Books []wizardBook
}

func (wizard *wizardCustom) AbilityEnabled(ability WizardAbility) bool {
    for _, check := range wizard.Abilities {
        if check == ability {
            return true
        }
    }

    return false
}

func (wizard *wizardCustom) ToggleAbility(ability WizardAbility, picksLeft int){
    var out []WizardAbility

    found := false

    for _, check := range wizard.Abilities {
        if check == ability {
            found = true
        } else {
            out = append(out, check)
        }
    }

    if !found && ability.PickCost() <= picksLeft {
        out = append(out, ability)
    }

    wizard.Abilities = out
}

func (wizard *wizardCustom) SetMagicLevel(kind MagicType, count int){
    var out []wizardBook

    found := false

    for _, book := range wizard.Books {
        if book.Magic == kind {
            found = true
            if count != 0 {
                book.Count = count
                out = append(out, book)
            }
        } else {
            out = append(out, book)
        }
    }

    if !found {
        out = append(out, wizardBook{
            Magic: kind,
            Count: count,
        })
    }

    // fmt.Printf("Books: %+v\n", out)

    wizard.Books = out
}

func (wizard *wizardCustom) MagicLevel(kind MagicType) int {
    for _, book := range wizard.Books {
        if book.Magic == kind {
            return book.Count
        }
    }

    return 0
}

type UIInsideElementFunc func(element *UIElement)
type UINotInsideElementFunc func(element *UIElement)
type UIClickElementFunc func(element *UIElement)
type UIDrawFunc func(element *UIElement, window *ebiten.Image)
type UIKeyFunc func(key ebiten.Key)

type UILayer int

type UIElement struct {
    Rect image.Rectangle
    NotInside UINotInsideElementFunc
    Inside UIInsideElementFunc
    LeftClick UIClickElementFunc
    RightClick UIClickElementFunc
    Draw UIDrawFunc
    Layer UILayer
}

type UI struct {
    // track the layer number of the elements
    Elements map[UILayer][]*UIElement
    // keep track of the minimum and maximum keys so we don't have to sort
    minLayer UILayer
    maxLayer UILayer
    Draw func(*UI, *ebiten.Image)
    HandleKey UIKeyFunc
}

func (ui *UI) AddElement(element *UIElement){
    if element.Layer < ui.minLayer {
        ui.minLayer = element.Layer
    }
    if element.Layer > ui.maxLayer {
        ui.maxLayer = element.Layer
    }

    ui.Elements[element.Layer] = append(ui.Elements[element.Layer], element)
}

func (ui *UI) RemoveElement(toRemove *UIElement){
    elements := ui.Elements[toRemove.Layer]
    var out []*UIElement
    for _, element := range elements {
        if element != toRemove {
            out = append(out, element)
        }
    }

    ui.Elements[toRemove.Layer] = out

    /*
    // recompute min/max layers
    // this is a minor optimization really, so implement it later
    if len(out) == 0 {
        min := 0
        max := 0

        for layer, elements := range ui.Elements {
            if layer < min {
                min = layer
            }
            if layer > max {
                max = layer
            }
        }
    }
    */
}

func (ui *UI) IterateElementsByLayer(f func(*UIElement)){
    for i := ui.minLayer; i <= ui.maxLayer; i++ {
        for _, element := range ui.Elements[i] {
            f(element)
        }
    }
}

func (ui *UI) GetHighestLayer() []*UIElement {
    for i := ui.maxLayer; i >= ui.minLayer; i-- {
        elements := ui.Elements[i]
        if len(elements) > 0 {
            return elements
        }
    }

    return nil
}

func (ui *UI) SetElementsFromArray(elements []*UIElement){
    out := make(map[UILayer][]*UIElement)

    for _, element := range elements {
        if element.Layer < ui.minLayer {
            ui.minLayer = element.Layer
        }

        if element.Layer > ui.maxLayer {
            ui.maxLayer = element.Layer
        }

        out[element.Layer] = append(out[element.Layer], element)
    }

    ui.Elements = out
}

type NewWizardScreen struct {
    Background *ebiten.Image
    CustomPictureBackground *ebiten.Image
    CustomWizardBooks *ebiten.Image
    Slots *ebiten.Image
    Font *font.Font
    AbilityFont *font.Font
    AbilityFontSelected *font.Font
    AbilityFontAvailable *font.Font
    HelpFont *font.Font
    HelpTitleFont *font.Font
    CheckMark *ebiten.Image
    NameFont *font.Font
    NameFontBright *font.Font
    SelectFont *font.Font
    loaded sync.Once
    WizardSlots []wizardSlot

    OkReady *ebiten.Image
    OkNotReady *ebiten.Image

    Help lbx.Help
    HelpImageLoader func(string, int) (*ebiten.Image, error)
    HelpTop *ebiten.Image
    HelpBottom *ebiten.Image

    UI *UI

    NameBox *ebiten.Image

    LifeBooks [3]*ebiten.Image
    SorceryBooks [3]*ebiten.Image
    NatureBooks [3]*ebiten.Image
    DeathBooks [3]*ebiten.Image
    ChaosBooks [3]*ebiten.Image

    BooksOrder []int

    State NewWizardScreenState

    CustomWizard wizardCustom

    CurrentWizard int
    Active bool

    counter uint64
}

func (screen *NewWizardScreen) MakeCustomNameUI() *UI {
    const portraitX = 24
    const portraitY = 10

    const nameX = 75
    const nameY = 120

    ui := &UI{
        Elements: make(map[UILayer][]*UIElement),
        HandleKey: func(key ebiten.Key){
            switch key {
                case ebiten.KeyBackspace:
                    length := len(screen.CustomWizard.Name)
                    if length > 0 {
                        length -= 1
                    }
                    screen.CustomWizard.Name = screen.CustomWizard.Name[0:length]
                case ebiten.KeyEnter:
                    screen.State = NewWizardScreenStateCustomBooks
                    screen.UI = screen.MakeCustomWizardBooksUI()
                case ebiten.KeySpace:
                    screen.CustomWizard.Name += " "
                default:
                    str := strings.ToLower(key.String())
                    if str != "" && validNameString(str) {
                        screen.CustomWizard.Name += str
                    }
            }

            if len(screen.CustomWizard.Name) > MaxNameLength {
                screen.CustomWizard.Name = screen.CustomWizard.Name[0:MaxNameLength]
            }
        },
        Draw: func(this *UI, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            window.DrawImage(screen.Background, &options)

            options.GeoM.Translate(portraitX, portraitY)
            window.DrawImage(screen.CustomWizard.Portrait, &options)
            screen.Font.PrintCenter(window, nameX, nameY, 1, screen.CustomWizard.Name)
            screen.SelectFont.PrintCenter(window, 245, 2, 1, "Wizard's Name")

            options.GeoM.Reset()
            options.GeoM.Translate(184, 20)
            window.DrawImage(screen.NameBox, &options)

            name := screen.CustomWizard.Name

            // add blinking _ to show cursor position
            if (screen.counter / 30) % 2 == 0 {
                name += "_"
            }

            screen.NameFont.Print(window, 195, 39, 1, name)

            return
        },
    }

    return ui
}

func (screen *NewWizardScreen) MakeCustomPictureUI() *UI {

    clickFunc := func(wizard int){
        screen.State = NewWizardScreenStateCustomName

        screen.UI = screen.MakeCustomNameUI()
    }

    insideFunc := func(wizard int){
        screen.CustomWizard.Portrait = screen.WizardSlots[wizard].Portrait
        screen.CustomWizard.Name = screen.WizardSlots[wizard].Name
    }

    elements := screen.MakeWizardUIElements(clickFunc, insideFunc)

    ui := &UI{
        Draw: func(this *UI, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            window.DrawImage(screen.Background, &options)

            screen.SelectFont.PrintCenter(window, 245, 2, 1, "Select Wizard")

            const portraitX = 24
            const portraitY = 10

            options.GeoM.Reset()
            options.GeoM.Translate(166, 18)
            window.DrawImage(screen.CustomPictureBackground, &options)

            this.IterateElementsByLayer(func (element *UIElement){
                element.Draw(element, window)
            })

            if screen.CustomWizard.Portrait != nil {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(portraitX, portraitY)
                window.DrawImage(screen.CustomWizard.Portrait, &options)
            }
        },
    }

    ui.SetElementsFromArray(elements)

    return ui
}

func (screen *NewWizardScreen) MakeWizardUIElements(clickFunc func(wizard int), insideFunc func(wizard int)) []*UIElement {
    var elements []*UIElement

    top := 28
    space := 22
    columnSpace := 76

    left := 170

    counter := 0
    for column := 0; column < 2; column += 1 {
        for row := 0; row < 7; row++ {
            wizard := counter
            background := screen.WizardSlots[counter].Background
            name := screen.WizardSlots[counter].Name
            counter += 1

            x1 := left + column * columnSpace
            y1 := top + row * space
            x2 := x1 + background.Bounds().Dx()
            y2 := y1 + background.Bounds().Dy()

            elements = append(elements, &UIElement{
                Rect: image.Rect(x1, y1, x2, y2),
                LeftClick: func(this *UIElement){
                    clickFunc(wizard)
                },
                Inside: func(this *UIElement){
                    insideFunc(wizard)
                    // screen.CurrentWizard = wizard
                },
                Draw: func(this *UIElement, window *ebiten.Image){
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(x1), float64(y1))
                    window.DrawImage(background, &options)
                    screen.Font.PrintCenter(window, float64(x1) + float64(background.Bounds().Dx()) / 2, float64(y1) + 3, 1, name)
                },
            })
        }
    }

    return elements
}

func (screen *NewWizardScreen) MakeSelectWizardUI() *UI {
    top := 28
    space := 22
    columnSpace := 76

    left := 170

    clickFunc := func(wizard int){
        // TODO: make this the selected wizard
    }

    insideFunc := func(wizard int){
        screen.CurrentWizard = wizard
    }

    elements := screen.MakeWizardUIElements(clickFunc, insideFunc)

    // custom element
    elements = append(elements, (func () *UIElement {
        background := screen.WizardSlots[len(elements)].Background
        x1 := left + columnSpace
        y1 := top + 7 * space
        x2 := x1 + background.Bounds().Dx()
        y2 := y1 + background.Bounds().Dy()

        return &UIElement{
            Rect: image.Rect(x1, y1, x2, y2),
            LeftClick: func(this *UIElement){
                screen.State = NewWizardScreenStateCustomPicture

                screen.UI = screen.MakeCustomPictureUI()

            },
            Inside: func(this *UIElement){
                screen.CurrentWizard = -1
            },
            Draw: func(this *UIElement, window *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(x1), float64(y1))
                window.DrawImage(background, &options)
                screen.Font.PrintCenter(window, float64(x1) + float64(background.Bounds().Dx()) / 2, float64(y1) + 3, 1, "Custom")
            },
        }
    })())

    ui := &UI{
        Draw: func(this *UI, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            window.DrawImage(screen.Background, &options)
            screen.SelectFont.PrintCenter(window, 245, 2, 1, "Select Wizard")

            this.IterateElementsByLayer(func (element *UIElement){
                element.Draw(element, window)
            })

            if screen.CurrentWizard >= 0 && screen.CurrentWizard < len(screen.WizardSlots) {
                const portraitX = 24
                const portraitY = 10

                const nameX = 75
                const nameY = 120

                portrait := screen.WizardSlots[screen.CurrentWizard].Portrait
                if portrait != nil {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(portraitX, portraitY)
                    window.DrawImage(portrait, &options)
                    screen.Font.PrintCenter(window, nameX, nameY, 1, screen.WizardSlots[screen.CurrentWizard].Name)

                    screen.DrawBooks(window, 36, 135, screen.WizardSlots[screen.CurrentWizard].Books)
                    if screen.WizardSlots[screen.CurrentWizard].ExtraAbility != AbilityNone {
                        screen.AbilityFont.Print(window, 12, 180, 1, screen.WizardSlots[screen.CurrentWizard].ExtraAbility.String())
                    }
                }
            }
        },
    }

    ui.SetElementsFromArray(elements)

    return ui
}

func (screen *NewWizardScreen) IsActive() bool {
    return screen.Active
}

func (screen *NewWizardScreen) Activate() {
    screen.Active = true
}

func (screen *NewWizardScreen) Deactivate() {
    screen.Active = false
}

func validNameString(text string) bool {
    // FIXME: only allow a-zA-Z0-9 and maybe a few extra chars lke ' and ,
    return true
}

const MaxNameLength = 18

func (screen *NewWizardScreen) Update() {
    screen.counter += 1

    if screen.UI.HandleKey != nil {
        keys := make([]ebiten.Key, 0)
        keys = inpututil.AppendJustPressedKeys(keys)

        for _, key := range keys {
            screen.UI.HandleKey(key)
        }
    }

    leftClick := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
    rightClick := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)

    mouseX, mouseY := ebiten.CursorPosition()

    if screen.UI != nil {
        for _, element := range screen.UI.GetHighestLayer() {
            if mouseX >= element.Rect.Min.X && mouseY >= element.Rect.Min.Y && mouseX < element.Rect.Max.X && mouseY <= element.Rect.Max.Y {
                if element.Inside != nil {
                    element.Inside(element)
                }
                if leftClick && element.LeftClick != nil {
                    element.LeftClick(element)
                }
                if rightClick && element.RightClick != nil {
                    element.RightClick(element)
                }
            } else {
                if element.NotInside != nil {
                    element.NotInside(element)
                }
            }
        }
    }
}

func (screen *NewWizardScreen) LoadHelp(cache *lbx.LbxCache) error {
    helpLbx, err := cache.GetLbxFile("HELP.LBX")
    if err != nil {
        return err
    }

    screen.Help, err = helpLbx.ReadHelp(2)
    if err != nil {
        return err
    }

    screen.HelpImageLoader = func(lbxName string, index int) (*ebiten.Image, error) {
        lbxFile, err := cache.GetLbxFile(lbxName)
        if err != nil {
            return nil, err
        }

        images, err := lbxFile.ReadImages(index)
        if err != nil {
            return nil, err
        }

        if len(images) == 0 {
            return nil, fmt.Errorf("no images found in %s entry %d", lbxName, index)
        }

        return ebiten.NewImageFromImage(images[0]), nil
    }

    scrollTopImages, err := helpLbx.ReadImages(0)
    if err != nil {
        return err
    }

    if len(scrollTopImages) == 0 {
        return fmt.Errorf("no images found in HELP.LBX entry 0")
    }

    screen.HelpTop = ebiten.NewImageFromImage(scrollTopImages[0])

    scrollBottomImages, err := helpLbx.ReadImages(1)
    if err != nil {
        return err
    }

    if len(scrollBottomImages) == 0 {
        return fmt.Errorf("no images found in HELP.LBX entry 1")
    }

    screen.HelpBottom = ebiten.NewImageFromImage(scrollBottomImages[0])

    return nil
}

func (screen *NewWizardScreen) Load(cache *lbx.LbxCache) error {
    // NEWGAME.LBX entry 8 contains boxes for wizard names
    // 9-23 are backgrounds for names
    // 24-26 are life books
    // 27-29 are sorcery/blue books
    // 30-32 are nature/green books
    // 33-35 are death books
    // 36-38 are chaos/red books
    // 41 is custom screen

    var outError error

    screen.loaded.Do(func() {
        fontLbx, err := cache.GetLbxFile("FONTS.LBX")
        if err != nil {
            outError = fmt.Errorf("Unable to read FONTS.LBX: %v", err)
            return
        }

        fonts, err := fontLbx.ReadFonts(0)
        if err != nil {
            outError = fmt.Errorf("Unable to read fonts from FONTS.LBX: %v", err)
            return
        }

        err = screen.LoadHelp(cache)
        if err != nil {
            outError = fmt.Errorf("Error reading help.lbx: %v", err)
            return
        }

        // FIXME: this is a fudged palette to look like the original, but its probably slightly wrong
        brightYellowPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            // orangish
            color.RGBA{R: 0xff, G: 0xaa, B: 0x00, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        }

        // FIXME: also a fudged palette
        whitishPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0xc6, G: 0x9d, B: 0x65, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        }

        pickPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0xc6, G: 0x9d, B: 0x65, A: 0xff},
            color.RGBA{R: 0xc6, G: 0x9d, B: 0x65, A: 0xff},
            color.RGBA{R: 0xc6, G: 0x9d, B: 0x65, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        }

        transparentPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: uint8(math.Round(0xc6 * 0.3)), G: uint8(math.Round(0x9d * 0.3)), B: uint8(math.Round(0x65 * 0.3)), A: uint8(math.Round(0xff * 0.3))},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        }

        screen.Font = font.MakeOptimizedFont(fonts[4])

        // FIXME: load with a yellowish palette
        screen.SelectFont = font.MakeOptimizedFont(fonts[5])

        screen.AbilityFont = font.MakeOptimizedFontWithPalette(fonts[0], transparentPalette)
        screen.AbilityFontSelected = font.MakeOptimizedFontWithPalette(fonts[0], brightYellowPalette)
        screen.AbilityFontAvailable = font.MakeOptimizedFontWithPalette(fonts[0], whitishPalette)

        helpPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0x5e, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        }

        screen.HelpFont = font.MakeOptimizedFontWithPalette(fonts[1], helpPalette)

        titleRed := color.RGBA{R: 0x50, G: 0x00, B: 0x0e, A: 0xff}
        titlePalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            titleRed,
            titleRed,
            titleRed,
            titleRed,
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        }

        screen.HelpTitleFont = font.MakeOptimizedFontWithPalette(fonts[4], titlePalette)

        // FIXME: use a monochrome color scheme, light-brownish
        screen.NameFont = font.MakeOptimizedFont(fonts[3])
        screen.NameFontBright = font.MakeOptimizedFontWithPalette(fonts[3], pickPalette)

        newGameLbx, err := cache.GetLbxFile("NEWGAME.LBX")
        if err != nil {
            outError = fmt.Errorf("Unable to load NEWGAME.LBX: %v", err)
            return
        }

        loadImage := func(index int, subIndex int) *ebiten.Image {
            if outError != nil {
                return nil
            }

            sprites, err := newGameLbx.ReadImages(index)
            if err != nil {
                outError = fmt.Errorf("Unable to read background image from NEWGAME.LBX: %v", err)
                return nil
            }

            if len(sprites) <= subIndex {
                outError = fmt.Errorf("Unable to read background image from NEWGAME.LBX: index %d out of range", subIndex)
                return nil
            }

            return ebiten.NewImageFromImage(sprites[subIndex])
        }

        wizardsLbx, err := cache.GetLbxFile("WIZARDS.LBX")
        if err != nil {
            outError = err
            return
        }

        loadWizardPortrait := func(index int) *ebiten.Image {
            if outError != nil {
                return nil
            }

            sprites, err := wizardsLbx.ReadImages(index)
            if err != nil {
                outError = fmt.Errorf("Unable to read wizard portrait from WIZARDS.LBX: %v", err)
                return nil
            }

            if len(sprites) == 0 {
                outError = fmt.Errorf("Unable to read wizard portrait from WIZARDS.LBX: no images found")
                return nil
            }

            return ebiten.NewImageFromImage(sprites[0])
        }

        screen.Background = loadImage(0, 0)
        screen.Slots = loadImage(8, 0)
        screen.NameBox = loadImage(40, 0)

        screen.OkReady = loadImage(42, 0)
        screen.OkNotReady = loadImage(43, 0)

        screen.CustomPictureBackground = loadImage(39, 0)
        screen.CustomWizardBooks = loadImage(41, 0)

        for i := 0; i < 3; i++ {
            screen.LifeBooks[i] = loadImage(24 + i, 0)
        }

        for i := 0; i < 3; i++ {
            screen.SorceryBooks[i] = loadImage(27 + i, 0)
        }

        for i := 0; i < 3; i++ {
            screen.NatureBooks[i] = loadImage(30 + i, 0)
        }

        for i := 0; i < 3; i++ {
            screen.DeathBooks[i] = loadImage(33 + i, 0)
        }

        for i := 0; i < 3; i++ {
            screen.ChaosBooks[i] = loadImage(36 + i, 0)
        }

        screen.CheckMark = loadImage(52, 0)

        top := 28
        space := 22

        screen.WizardSlots = []wizardSlot{
            wizardSlot{
                Name: "Merlin",
                Background: loadImage(9, 0),
                Portrait: loadWizardPortrait(0),
                Books: []wizardBook{
                    wizardBook{Magic: LifeMagic, Count: 5},
                    wizardBook{Magic: NatureMagic, Count: 5},
                },
                ExtraAbility: AbilitySageMaster,
                X: 170,
                Y: top,
            },
            wizardSlot{
                Name: "Raven",
                Background: loadImage(10, 0),
                Portrait: loadWizardPortrait(1),
                Books: []wizardBook{
                    wizardBook{Magic: SorceryMagic, Count: 6},
                    wizardBook{Magic: NatureMagic, Count: 5},
                },
                ExtraAbility: AbilityNone,
                X: 170,
                Y: top + 1 * space,
            },
            wizardSlot{
                Name: "Sharee",
                Background: loadImage(11, 0),
                Portrait: loadWizardPortrait(2),
                Books: []wizardBook{
                    wizardBook{Magic: DeathMagic, Count: 5},
                    wizardBook{Magic: ChaosMagic, Count: 5},
                },
                ExtraAbility: AbilityConjurer,
                X: 170,
                Y: top + 2 * space,
            },
            wizardSlot{
                Name: "Lo Pan",
                Background: loadImage(12, 0),
                Portrait: loadWizardPortrait(3),
                Books: []wizardBook{
                    wizardBook{Magic: SorceryMagic, Count: 5},
                    wizardBook{Magic: ChaosMagic, Count: 5},
                },
                ExtraAbility: AbilityChanneler,
                X: 170,
                Y: top + 3 * space,
            },
            wizardSlot{
                Name: "Jafar",
                Background: loadImage(13, 0),
                Portrait: loadWizardPortrait(4),
                Books: []wizardBook{
                    wizardBook{Magic: SorceryMagic, Count: 10},
                },
                ExtraAbility: AbilityAlchemy,
                X: 170,
                Y: top + 4 * space,
            },
            wizardSlot{
                Name: "Oberic",
                Background: loadImage(14, 0),
                Portrait: loadWizardPortrait(5),
                Books: []wizardBook{
                    wizardBook{Magic: NatureMagic, Count: 5},
                    wizardBook{Magic: ChaosMagic, Count: 5},
                },
                ExtraAbility: AbilityManaFocusing,
                X: 170,
                Y: top + 5 * space,
            },
            wizardSlot{
                Name: "Rjak",
                Background: loadImage(15, 0),
                Portrait: loadWizardPortrait(6),
                Books: []wizardBook{
                    wizardBook{Magic: DeathMagic, Count: 9},
                },
                ExtraAbility: AbilityInfernalPower,
                X: 170,
                Y: top + 6 * space,
            },
            wizardSlot{
                Name: "Ssr'ra",
                Background: loadImage(16, 0),
                Portrait: loadWizardPortrait(7),
                Books: []wizardBook{
                    wizardBook{Magic: LifeMagic, Count: 4},
                    wizardBook{Magic: ChaosMagic, Count: 4},
                },
                ExtraAbility: AbilityMyrran,
                X: 246,
                Y: top + 0 * space,
            },
            wizardSlot{
                Name: "Tauron",
                Background: loadImage(17, 0),
                Portrait: loadWizardPortrait(8),
                Books: []wizardBook{
                    wizardBook{Magic: ChaosMagic, Count: 10},
                },
                ExtraAbility: AbilityChaosMastery,
                X: 246,
                Y: top + 1 * space,
            },
            wizardSlot{
                Name: "Freya",
                Background: loadImage(18, 0),
                Portrait: loadWizardPortrait(9),
                Books: []wizardBook{
                    wizardBook{Magic: NatureMagic, Count: 10},
                },
                ExtraAbility: AbilityNatureMastery,
                X: 246,
                Y: top + 2 * space,
            },
            wizardSlot{
                Name: "Horus",
                Background: loadImage(19, 0),
                Portrait: loadWizardPortrait(10),
                Books: []wizardBook{
                    wizardBook{Magic: LifeMagic, Count: 5},
                    wizardBook{Magic: SorceryMagic, Count: 5},
                },
                ExtraAbility: AbilityArchmage,
                X: 246,
                Y: top + 3 * space,
            },
            wizardSlot{
                Name: "Ariel",
                Background: loadImage(20, 0),
                Portrait: loadWizardPortrait(11),
                Books: []wizardBook{
                    wizardBook{Magic: LifeMagic, Count: 10},
                },
                ExtraAbility: AbilityCharismatic,
                X: 246,
                Y: top + 4 * space,
            },
            wizardSlot{
                Name: "Tlaloc",
                Background: loadImage(21, 0),
                Portrait: loadWizardPortrait(12),
                Books: []wizardBook{
                    wizardBook{Magic: NatureMagic, Count: 4},
                    wizardBook{Magic: DeathMagic, Count: 5},
                },
                ExtraAbility: AbilityWarlord,
                X: 246,
                Y: top + 5 * space,
            },
            wizardSlot{
                Name: "Kali",
                Background: loadImage(22, 0),
                Portrait: loadWizardPortrait(13),
                Books: []wizardBook{
                    wizardBook{Magic: SorceryMagic, Count: 5},
                    wizardBook{Magic: DeathMagic, Count: 5},
                },
                ExtraAbility: AbilityArtificer,
                X: 246,
                Y: top + 6 * space,
            },
            wizardSlot{
                Name: "Custom",
                Background: loadImage(23, 0),
                Books: nil,
                Portrait: nil,
                X: 246,
                Y: top + 7 * space,
            },
        }

        // set custom wizard to merlin for now
        screen.CustomWizard.Portrait = screen.WizardSlots[0].Portrait
        screen.CustomWizard.Name = screen.WizardSlots[0].Name
        // screen.CustomWizard.Abilities = []WizardAbility{AbilityAlchemy, AbilityConjurer, AbilityFamous}
        screen.CustomWizard.Books = []wizardBook{
            wizardBook{
                Magic: NatureMagic,
                Count: 2,
            },
            wizardBook{
                Magic: DeathMagic,
                Count: 3,
            },
            wizardBook{
                Magic: SorceryMagic,
                Count: 2,
            },
        }

        if screen.State == NewWizardScreenStateSelectWizard {
            screen.UI = screen.MakeSelectWizardUI()
        } else if screen.State == NewWizardScreenStateCustomBooks {
            screen.UI = screen.MakeCustomWizardBooksUI()
        }
    })

    return outError
}

func joinAbilities(abilities []WizardAbility) string {
    // this could be simplified by iterating backwards through the array and
    // preprending 'and' or ', ' for each element, depending on its index

    if len(abilities) == 0 {
        return ""
    }

    // x
    if len(abilities) == 1 {
        return abilities[0].String()
    }

    // x and y
    if len(abilities) == 2 {
        return fmt.Sprintf("%v and %v", abilities[0], abilities[1])
    }

    // x, y, and z
    out := ""
    for i := 0; i < len(abilities) - 2; i++ {
        out += fmt.Sprintf("%v, ", abilities[i])
    }
    out += fmt.Sprintf("%v and %v", abilities[len(abilities) - 2], abilities[len(abilities) - 1])
    return out
}

func (screen *NewWizardScreen) DrawBooks(window *ebiten.Image, x float64, y float64, books []wizardBook){
    offsetX := 0
    index := 0

    for _, book := range books {

        for i := 0; i < book.Count; i++ {
            // can't draw more books than we have
            if index >= len(screen.BooksOrder) {
                return
            }

            var img *ebiten.Image
            switch book.Magic {
                case LifeMagic: img = screen.LifeBooks[screen.BooksOrder[index]]
                case SorceryMagic: img = screen.SorceryBooks[screen.BooksOrder[index]]
                case NatureMagic: img = screen.NatureBooks[screen.BooksOrder[index]]
                case DeathMagic: img = screen.DeathBooks[screen.BooksOrder[index]]
                case ChaosMagic: img = screen.ChaosBooks[screen.BooksOrder[index]]
            }

            var options ebiten.DrawImageOptions
            options.GeoM.Translate(x + float64(offsetX), y)
            window.DrawImage(img, &options)
            offsetX += img.Bounds().Dx() - 1
            index += 1
        }
    }
}

func (screen *NewWizardScreen) makeHelpElement(help lbx.HelpEntry) *UIElement {
    infoX := 55
    infoY := 30
    infoWidth := screen.HelpTop.Bounds().Dx()
    // infoHeight := screen.HelpTop.Bounds().Dy()
    infoLeftMargin := 18
    infoTopMargin := 26
    infoBodyMargin := 3
    maxInfoWidth := infoWidth - infoLeftMargin - infoBodyMargin - 15

    wrapped := screen.HelpFont.CreateWrappedText(float64(maxInfoWidth), 1, help.Text)

    helpTextY := infoY + infoTopMargin
    titleYAdjust := 0

    var extraImage *ebiten.Image
    if help.Lbx != "" {
        use, err := screen.HelpImageLoader(help.Lbx, help.LbxIndex)
        if err == nil && use != nil {
            extraImage = use
        }
    }

    if extraImage != nil {
        titleYAdjust = extraImage.Bounds().Dy() / 2 - screen.HelpTitleFont.Height() / 2

        if extraImage.Bounds().Dy() > screen.HelpTitleFont.Height() {
            helpTextY += extraImage.Bounds().Dy() + 1
        } else {
            helpTextY += screen.HelpTitleFont.Height() + 1
        }
    } else {
        helpTextY += screen.HelpTitleFont.Height() + 1
    }

    bottom := float64(helpTextY) + wrapped.TotalHeight

    // only draw as much of the top scroll as there are lines of text
    topImage := screen.HelpTop.SubImage(image.Rect(0, 0, screen.HelpTop.Bounds().Dx(), int(bottom) - infoY)).(*ebiten.Image)

    infoElement := &UIElement{
        // Rect: image.Rect(infoX, infoY, infoX + infoWidth, infoY + infoHeight),
        Rect: image.Rect(0, 0, ScreenWidth, ScreenHeight),
        Draw: func (infoThis *UIElement, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(infoX), float64(infoY))
            window.DrawImage(topImage, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(infoX), float64(bottom))
            window.DrawImage(screen.HelpBottom, &options)

            // for debugging
            // vector.StrokeRect(window, float32(infoX), float32(infoY), float32(infoWidth), float32(infoHeight), 1, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, true)
            // vector.StrokeRect(window, float32(infoX + infoLeftMargin), float32(infoY + infoTopMargin), float32(maxInfoWidth), float32(screen.HelpTitleFont.Height() + 20 + 1), 1, color.RGBA{R: 0, G: 0xff, B: 0, A: 0xff}, false)

            titleX := infoX + infoLeftMargin

            if extraImage != nil {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(titleX), float64(infoY + infoTopMargin))
                window.DrawImage(extraImage, &options)
                titleX += extraImage.Bounds().Dx() + 5
            }

            screen.HelpTitleFont.Print(window, float64(titleX), float64(infoY + infoTopMargin + titleYAdjust), 1, help.Headline)
            screen.HelpFont.RenderWrapped(window, float64(infoX + infoLeftMargin + infoBodyMargin), float64(helpTextY), wrapped)
        },
        LeftClick: func(infoThis *UIElement){
            screen.UI.RemoveElement(infoThis)
        },
        Layer: 1,
    }

    return infoElement
}

func (screen *NewWizardScreen) MakeCustomWizardBooksUI() *UI {

    picksLeft := func() int {
        picks := MaxPicks

        for _, ability := range screen.CustomWizard.Abilities {
            picks -= ability.PickCost()
        }

        for _, book := range screen.CustomWizard.Books {
            picks -= book.Count
        }

        return picks
    }

    var elements []*UIElement

    const bookWidth = 8
    const bookHeight = 20
    const bookX = 197

    type bookData struct {
        Kind MagicType
        Help string
        Image *ebiten.Image
        Y int
    }

    books := []bookData{
        bookData{
            Kind: LifeMagic,
            Help: "Life Spells",
            Image: screen.LifeBooks[0],
            Y: 49,
        },
        bookData{
            Kind: DeathMagic,
            Help: "Death Spells",
            Image: screen.DeathBooks[0],
            Y: 75,
        },
        bookData{
            Kind: ChaosMagic,
            Help: "Chaos Spells",
            Image: screen.ChaosBooks[0],
            Y: 101,
        },
        bookData{
            Kind: NatureMagic,
            Help: "Nature Spells",
            Image: screen.NatureBooks[0],
            Y: 127,
        },
        bookData{
            Kind: SorceryMagic,
            Help: "Sorcery Spells",
            Image: screen.SorceryBooks[0],
            Y: 153,
        },
    }

    // for each magic book, create a UI element that contains the book dimensions and can draw the book

    for _, book := range books {
        bookY := book.Y
        bookMagic := book.Kind
        bookImage := book.Image

        x1 := bookX - bookWidth
        y1 := bookY
        x2 := x1 + bookWidth
        y2 := y1 + bookHeight

        // element to remove all books
        elements = append(elements, &UIElement{
            Rect: image.Rect(x1, y1, x2, y2),
            LeftClick: func(this *UIElement){
                screen.CustomWizard.SetMagicLevel(bookMagic, 0)
            },
            Draw: func(this *UIElement, window *ebiten.Image){
            },
        })

        ghostBooks := -1

        minX := bookX
        maxX := bookX

        for i := 0; i < 11; i++ {
            // Rect image.Rectangle
            // Inside UIInsideElementFunc
            // Click UIClickElementFunc
            // Draw UIDrawFunc

            x1 := bookX + bookWidth * i
            y1 := bookY
            x2 := x1 + bookWidth
            y2 := y1 + bookHeight

            if x1 < minX {
                minX = x1
            }

            if x2 > maxX {
                maxX = x2
            }

            level := i

            element := &UIElement{
                Rect: image.Rect(x1, y1, x2, y2),
                LeftClick: func(this *UIElement){

                    // user cannot hold both life and death magic
                    if bookMagic == LifeMagic && screen.CustomWizard.MagicLevel(DeathMagic) > 0 {
                        return
                    }

                    if bookMagic == DeathMagic && screen.CustomWizard.MagicLevel(LifeMagic) > 0 {
                        return
                    }

                    // current := screen.CustomWizard.MagicLevel(bookMagic)
                    screen.CustomWizard.SetMagicLevel(bookMagic, level+1)
                    if picksLeft() < 0 {
                        screen.CustomWizard.SetMagicLevel(bookMagic, screen.CustomWizard.MagicLevel(bookMagic) + picksLeft())
                    }
                },
                RightClick: func(this *UIElement){
                    helpEntries := screen.Help.GetEntriesByName(book.Help)
                    if helpEntries == nil {
                        return
                    }

                    screen.UI.AddElement(screen.makeHelpElement(helpEntries[0]))
                },
                Inside: func(this *UIElement){
                    // if the user hovers over this element, then draw partially transparent books
                    ghostBooks = level
                },
                Draw: func(this *UIElement, window *ebiten.Image){
                    if screen.CustomWizard.MagicLevel(bookMagic) > level {
                        var options ebiten.DrawImageOptions
                        options.GeoM.Translate(float64(x1), float64(y1))
                        window.DrawImage(bookImage, &options)
                    } else if ghostBooks >= level {
                        // draw a transparent book that shows what the user would have if they selected this
                        // TODO: use a fragment shader to draw the book in a different color
                        var options ebiten.DrawImageOptions
                        options.ColorScale.Scale(1.4 * 0.5, 1 * 0.5, 1 * 0.5, 0.5)
                        options.GeoM.Translate(float64(x1), float64(y1))
                        window.DrawImage(bookImage, &options)
                    }
                },
            }

            elements = append(elements, element)
        }

        // add a non-drawing UI element that is used to detect if the user is pointing at any of the books
        elements = append(elements, &UIElement{
            Rect: image.Rect(minX, bookY, maxX, bookY + bookHeight),
            NotInside: func(this *UIElement){
                ghostBooks = -1
            },
        })
    }

    abilities := []WizardAbility{
        AbilityAlchemy,
        AbilityWarlord,
        AbilityChanneler,
        AbilityArchmage,
        AbilityArtificer,
        AbilityConjurer,
        AbilitySageMaster,
        AbilityMyrran,
        AbilityDivinePower,
        AbilityFamous,
        AbilityRunemaster,
        AbilityCharismatic,
        AbilityChaosMastery,
        AbilityNatureMastery,
        AbilitySorceryMastery,
        AbilityInfernalPower,
        AbilityManaFocusing,
        AbilityNodeMastery,
    }

    // FIXME: compute this based on the largest string in a single column
    tabs := []float64{172, 210, 260, 320}

    type abilityUI struct {
        Ability WizardAbility
        Length int
        X float64
        Y float64
    }

    // iterator to produce all ability ui element positions
    produceAbilityPositions := func() chan abilityUI {
        out := make(chan abilityUI)

        go func(){
            defer close(out)
            const topY = 5
            const veriticalGap = 7
            const maxY = 45

            tab := 0
            y := topY

            for _, ability := range abilities {
                out <- abilityUI{
                    Ability: ability,
                    Length: int(tabs[tab+1] - tabs[tab]),
                    X: tabs[tab],
                    Y: float64(y),
                }

                y += veriticalGap
                if y >= maxY {
                    tab += 1
                    y = topY
                }
            }
        }()

        return out
    }

    isAbilityAvailable := func(ability WizardAbility) bool {
        if picksLeft() < ability.PickCost() {
            return false
        }

        return ability.SatisifiedDependencies(&screen.CustomWizard)
    }

    for ability := range produceAbilityPositions() {
        elements = append(elements, &UIElement{
            Rect: image.Rect(int(ability.X), int(ability.Y), int(ability.X) + ability.Length, int(ability.Y) + screen.AbilityFont.Height()),
            LeftClick: func(this *UIElement){
                screen.CustomWizard.ToggleAbility(ability.Ability, picksLeft())
            },
            RightClick: func(this *UIElement){

                helpEntries := screen.Help.GetEntriesByName(ability.Ability.String())
                if helpEntries == nil {
                    return
                }

                // Hack! There are two FAMOUS entries in help.lbx, one for the ability and one for the spell
                if ability.Ability == AbilityFamous {
                    helpEntries = []lbx.HelpEntry{screen.Help.GetRawEntry(702)}
                }

                screen.UI.AddElement(screen.makeHelpElement(helpEntries[0]))
            },
            Draw: func(this *UIElement, window *ebiten.Image){
                font := screen.AbilityFont

                if screen.CustomWizard.AbilityEnabled(ability.Ability) {
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(ability.X - float64(screen.CheckMark.Bounds().Dx()) - 1, ability.Y + 1)
                    window.DrawImage(screen.CheckMark, &options)
                    font = screen.AbilityFontSelected
                } else if isAbilityAvailable(ability.Ability) {
                    font = screen.AbilityFontAvailable
                }

                font.Print(window, ability.X, ability.Y, 1, ability.Ability.String())
            },
        })
    }

    // ok button
    elements = append(elements, &UIElement{
        Rect: image.Rect(252, 182, 252 + screen.OkReady.Bounds().Dx(), 182 + screen.OkReady.Bounds().Dy()),
        RightClick: func(this *UIElement){
            helpEntries := screen.Help.GetEntriesByName("ok button")
            if helpEntries == nil {
                return
            }

            screen.UI.AddElement(screen.makeHelpElement(helpEntries[0]))
        },
        Draw: func(this *UIElement, window *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(252, 182)
            if picksLeft() == 0 {
                window.DrawImage(screen.OkReady, &options)
            } else {
                window.DrawImage(screen.OkNotReady, &options)
            }
        },
    })

    ui := &UI{
        Draw: func(ui *UI, window *ebiten.Image){
            const portraitX = 24
            const portraitY = 10

            const nameX = 75
            const nameY = 120

            var options ebiten.DrawImageOptions
            window.DrawImage(screen.CustomWizardBooks, &options)

            options.GeoM.Translate(portraitX, portraitY)
            window.DrawImage(screen.CustomWizard.Portrait, &options)
            screen.Font.PrintCenter(window, nameX, nameY, 1, screen.CustomWizard.Name)

            screen.DrawBooks(window, 37, 135, screen.CustomWizard.Books)

            ui.IterateElementsByLayer(func (element *UIElement){
                if element.Draw != nil {
                    element.Draw(element, window)
                }
            })

            screen.AbilityFontSelected.Print(window, 12, 180, 1, joinAbilities(screen.CustomWizard.Abilities))

            screen.NameFontBright.PrintCenter(window, 223, 185, 1, fmt.Sprintf("%v picks", picksLeft()))
        },
    }

    ui.SetElementsFromArray(elements)

    return ui

}

func (screen *NewWizardScreen) Draw(window *ebiten.Image) {
    if screen.UI != nil {
        screen.UI.Draw(screen.UI, window)
    }
}

// create an array of N integers where each integer is some value between 0 and 2
// these values correlate to the index of the book image to draw under the wizard portrait
func randomizeBookOrder(books int) []int {
    order := make([]int, books)
    for i := 0; i < books; i++ {
        order[i] = rand.Intn(3)
    }
    return order
}

func MakeNewWizardScreen() *NewWizardScreen {
    return &NewWizardScreen{
        CurrentWizard: 0,
        BooksOrder: randomizeBookOrder(12),
        State: NewWizardScreenStateCustomBooks,
    }
}
