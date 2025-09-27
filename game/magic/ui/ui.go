package ui

import (
    "log"
    "image"
    "cmp"
    "slices"
    "strings"
    "sync"

    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/exp/textinput"
)

type UIInsideElementFunc func(element *UIElement, x int, y int)
type UINotInsideElementFunc func(element *UIElement)
type UIClickElementFunc func(element *UIElement)
type UIDrawFunc func(element *UIElement, window *ebiten.Image)
type UIToolTipFunc func(element *UIElement) (string, *font.Font)
type UIKeyFunc func(key []ebiten.Key)
type UIGainFocusFunc func(*UIElement)
type UILoseFocusFunc func(*UIElement)
type UITextEntry func(*UIElement, string) string
type UIScrollFunc func(*UIElement, float64, float64)
type UIHackFunc func(*UIElement)

type UILayer int

type UIContainer interface {
    MakeFadeIn(time uint64) util.AlphaFadeFunc
    MakeFadeOut(time uint64) util.AlphaFadeFunc
    AddDelay(time uint64, f func())
    AddElement(*UIElement)
    AddElements([]*UIElement)
    RemoveElement(*UIElement)
    RemoveElements([]*UIElement)
}

type UIElement struct {
    Rect image.Rectangle
    // fires if the mouse is not inside this element
    NotInside UINotInsideElementFunc
    // fires if the mouse is inside this element
    Inside UIInsideElementFunc
    // fires if a left click occurred but no other element was clicked on
    NotLeftClicked UIClickElementFunc
    // fires when a left click occurs on this element
    LeftClick UIClickElementFunc
    // fires when the left click button is released on this element
    LeftClickRelease UIClickElementFunc
    // fires when this element is double clicked
    DoubleLeftClick UIClickElementFunc
    // fires when this element is right clicked
    RightClick UIClickElementFunc

    // if not nil and doesn't return the empty string, the tooltip will be rendered on top of the ui when
    // this element is hovered
    Tooltip UIToolTipFunc

    // fires when this element is left clicked
    GainFocus UIGainFocusFunc
    // fires when some other element is left clicked
    LoseFocus UILoseFocusFunc
    // fires when the user types some keys and this element is focused
    TextEntry UITextEntry
    // fires when a key is pressed and this element is focused
    HandleKeys UIKeyFunc

    // fires when the mouse wheel/pad is scrolled
    Scroll UIScrollFunc

    // implement this when trying to avoid go vet issues, such as invoking a cancel context function
    // this function is never called
    Hack UIHackFunc

    Draw UIDrawFunc
    Layer UILayer

    // for specifying which element is drawn first
    Order int

    // if true, the standard ui sound will play when this element is left clicked
    PlaySoundLeftClick bool

    IsOffsetWhenPressed bool
    isPressed bool
}

// a collection of ui elements that can be removed all at once
type UIElementGroup struct {
    Elements map[UILayer][]*UIElement
    minLayer UILayer
    maxLayer UILayer
    Counter uint64
    Delays []UIDelay
    // a function that is called on each tick. this can be set to any function that is useful
    // for the ui (such as updating animations)
    Update func()
}

func MakeGroup() *UIElementGroup {
    return &UIElementGroup{
        Elements: make(map[UILayer][]*UIElement),
        Update: func(){}, // no-op by default
    }
}

func (group *UIElementGroup) AddElement(element *UIElement){
    if element.Layer < group.minLayer {
        group.minLayer = element.Layer
    }
    if element.Layer > group.maxLayer {
        group.maxLayer = element.Layer
    }

    group.Elements[element.Layer] = slices.SortedFunc(slices.Values(append(group.Elements[element.Layer], element)), func (a, b *UIElement) int {
        return cmp.Compare(a.Order, b.Order)
    })
}

func (group *UIElementGroup) AddElements(elements []*UIElement){
    for _, element := range elements {
        group.AddElement(element)
    }
}

func (group *UIElementGroup) RemoveElement(element *UIElement){
    elements := group.Elements[element.Layer]
    elements = slices.DeleteFunc(elements, func (e *UIElement) bool {
        return e == element
    })
    group.Elements[element.Layer] = elements
}

func (group *UIElementGroup) RemoveElements(elements []*UIElement){
    for _, element := range elements {
        group.RemoveElement(element)
    }
}

func (group *UIElementGroup) MakeFadeIn(time uint64) util.AlphaFadeFunc {
    return util.MakeFadeIn(time, &group.Counter)
}

func (group *UIElementGroup) MakeFadeOut(time uint64) util.AlphaFadeFunc {
    return util.MakeFadeOut(time, &group.Counter)
}

func (group *UIElementGroup) AddDelay(time uint64, f func()){
    group.Delays = append(group.Delays, UIDelay{Time: group.Counter + time, Func: f})
}

const DoubleClickThreshold = 20

type doubleClick struct {
    Element *UIElement
    Time uint64
}

type UIDelay struct {
    Time uint64
    Func func()
}

type UI struct {
    // track the layer number of the elements
    Elements map[UILayer][]*UIElement
    Groups []*UIElementGroup
    // keep track of the minimum and maximum keys so we don't have to sort
    minLayer UILayer
    maxLayer UILayer
    Draw func(*UI, *ebiten.Image)
    HandleKeys UIKeyFunc
    Counter uint64

    focusedElement *UIElement
    lastTouchX int
    lastTouchY int
    touchRightClick bool
    touchStartTime uint64

    textField textinput.Field

    // the element that is currently being hovered over
    TooltipElement *UIElement
    TooltipPosition image.Point
    TooltipTime uint64

    doubleClickCandidates []doubleClick

    // fires when the left mouse button is clicked, even if no element is clicked
    LeftClick func()

    LeftClickedElements []*UIElement

    Delays []UIDelay

    // disabled so that the zero value is enabled
    Disabled bool

    // lazily initialized in the sync.Once
    StandardSoundMaker audio.MakePlayerFunc
    StandardSoundSetup sync.Once
    Cache *lbx.LbxCache
}

func (ui *UI) Enable() {
    ui.Disabled = false
}

func (ui *UI) Disable() {
    ui.Disabled = true
}

func (ui *UI) IsDisabled() bool {
    return ui.Disabled
}

func (ui *UI) MakeFadeIn(time uint64) util.AlphaFadeFunc {
    return util.MakeFadeIn(time, &ui.Counter)
}

func (ui *UI) MakeFadeOut(time uint64) util.AlphaFadeFunc {
    return util.MakeFadeOut(time, &ui.Counter)
}

func (ui *UI) AddElements(elements []*UIElement){
    for _, element := range elements {
        ui.AddElement(element)
    }
}

func (ui *UI) AddDelay(time uint64, f func()){
    ui.Delays = append(ui.Delays, UIDelay{Time: ui.Counter + time, Func: f})
}

func (ui *UI) AddGroup(group *UIElementGroup){
    ui.Groups = append(ui.Groups, group)

    if group.minLayer < ui.minLayer {
        ui.minLayer = group.minLayer
    }

    if group.maxLayer > ui.maxLayer {
        ui.maxLayer = group.maxLayer
    }
}

func (ui *UI) RemoveGroup(group *UIElementGroup){
    ui.Groups = slices.DeleteFunc(ui.Groups, func (g *UIElementGroup) bool {
        return g == group
    })
}

func (ui *UI) AddElement(element *UIElement){
    if element.Layer < ui.minLayer {
        ui.minLayer = element.Layer
    }
    if element.Layer > ui.maxLayer {
        ui.maxLayer = element.Layer
    }

    ui.Elements[element.Layer] = slices.SortedFunc(slices.Values(append(ui.Elements[element.Layer], element)), func (a, b *UIElement) int {
        return cmp.Compare(a.Order, b.Order)
    })
}

func (ui *UI) RemoveElements(toRemove []*UIElement){
    for _, element := range toRemove {
        ui.RemoveElement(element)
    }
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
    lowest := ui.minLayer
    highest := ui.maxLayer

    for _, group := range ui.Groups {
        if group.minLayer < lowest {
            lowest = group.minLayer
        }
        if group.maxLayer > highest {
            highest = group.maxLayer
        }
    }

    for i := lowest; i <= highest; i++ {
        for _, element := range ui.Elements[i] {
            f(element)
        }

        for _, group := range ui.Groups {
            for _, element := range group.Elements[i] {
                f(element)
            }
        }
    }
}

func (ui *UI) GetHighestLayerValue() UILayer {
    elements := ui.GetHighestLayer()

    if len(elements) > 0 {
        return elements[0].Layer
    }

    return 0
}

func (ui *UI) GetHighestLayer() []*UIElement {
    highest := ui.maxLayer
    lowest := ui.minLayer
    for _, group := range ui.Groups {
        if group.maxLayer > highest {
            highest = group.maxLayer
        }
        if group.minLayer < lowest {
            lowest = group.minLayer
        }
    }

    for i := highest; i >= lowest; i-- {
        out := ui.Elements[i]
        needClone := true

        for _, group := range ui.Groups {
            if len(group.Elements[i]) > 0 {
                if needClone {
                    out = slices.Clone(out)
                    needClone = false
                }

                out = append(out, group.Elements[i]...)
            }
        }

        if len(out) > 0 {
            return out
        }
    }

    // log.Printf("Warning: no elements in UI")

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

func (ui *UI) UnfocusElement(){
    if ui.focusedElement != nil {
        if ui.focusedElement.LoseFocus != nil {
            ui.focusedElement.LoseFocus(ui.focusedElement)
        }

        ui.focusedElement = nil
        ui.textField.Blur()
    }
}

func (ui *UI) FocusElement(element *UIElement, text string){
    if ui.focusedElement != nil && ui.focusedElement != element && ui.focusedElement.LoseFocus != nil {
        ui.focusedElement.LoseFocus(ui.focusedElement)
    }

    ui.focusedElement = element

    if ui.focusedElement.TextEntry != nil {
        ui.textField.Focus()
        ui.textField.SetTextAndSelection(text, len(text), len(text))
    }

    /*
    if element.GainFocus != nil {
        element.GainFocus(element)
    }
    */
}

func (ui *UI) PlayStandardSound() {
    if ui.StandardSoundMaker == nil && ui.Cache != nil {
        ui.StandardSoundSetup.Do(func(){
            maker, err := audio.LoadSoundMaker(ui.Cache, audio.SoundClick)
            if err != nil {
                log.Printf("Unable to load ui sound: %v", err)
                return
            }
            ui.StandardSoundMaker = maker
        })
    }

    if ui.StandardSoundMaker != nil {
        player := ui.StandardSoundMaker()
        player.Play()
    }
}

func (ui *UI) RenderTooltip(screen *ebiten.Image) {
    if ui.TooltipElement != nil && ui.TooltipElement.Tooltip != nil {
        tip, renderFont := ui.TooltipElement.Tooltip(ui.TooltipElement)
        if tip != "" {
            var options ebiten.DrawImageOptions

            diff := ui.Counter - ui.TooltipTime
            var delay uint64 = 0
            if diff > delay {
                alpha := diff - delay

                var alphaRange uint64 = 60 * 1 / 3

                if alpha < alphaRange {
                    options.ColorScale.ScaleAlpha(float32(alpha) / float32(alphaRange))
                }

                renderFont.PrintOptions(screen, float64(ui.TooltipPosition.X), float64(ui.TooltipPosition.Y - renderFont.Height() - 1), font.FontOptions{Options: &options, Scale: scale.ScaleAmount, DropShadow: true}, tip)
            }
        }
    }
}

// iterates through the elements in the highest layer and draws them
// then draws the tooltip on top
func (ui *UI) StandardDraw(screen *ebiten.Image) {
    ui.IterateElementsByLayer(func (element *UIElement){
        if element.Draw != nil {
            var offsetPoint image.Point
            if element.IsOffsetWhenPressed && element.isPressed {
                offsetPoint = image.Pt(1, 1)
            }

            element.Rect = element.Rect.Add(offsetPoint)
            element.Draw(element, screen)
            element.Rect = element.Rect.Sub(offsetPoint)
        }
    })

    ui.RenderTooltip(screen)
}

func (ui *UI) StandardUpdate() {
    ui.Counter += 1

    if len(ui.Delays) > 0 {
        var keepDelays []UIDelay
        // invoking a delay might cause another delay to be added
        oldDelays := slices.Clone(ui.Delays)
        ui.Delays = nil
        for _, delay := range oldDelays {
            if ui.Counter <= delay.Time {
                keepDelays = append(keepDelays, delay)
            } else {
                delay.Func()
            }
        }
        ui.Delays = append(ui.Delays, keepDelays...)
    }

    for _, group := range ui.Groups {
        group.Update()
        group.Counter += 1

        var keepDelays []UIDelay
        // invoking a delay might cause another delay to be added
        oldDelays := slices.Clone(group.Delays)
        group.Delays = nil
        for _, delay := range oldDelays {
            if group.Counter <= delay.Time {
                keepDelays = append(keepDelays, delay)
            } else {
                delay.Func()
            }
        }
        group.Delays = append(group.Delays, keepDelays...)
    }

    if !ui.Disabled {
        keys := inpututil.AppendJustPressedKeys(nil)
        if len(keys) > 0 {
            if ui.HandleKeys != nil {
                ui.HandleKeys(keys)
            }

            if ui.focusedElement != nil && ui.focusedElement.TextEntry == nil && ui.focusedElement.HandleKeys != nil {
                ui.focusedElement.HandleKeys(keys)
            }
        }
    }

    leftClick := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
    leftClickReleased := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
    rightClick := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)

    mouseX, mouseY := ebiten.CursorPosition()

    /*
    touchIds := inpututil.AppendJustPressedTouchIDs(nil)
    if len(touchIds) > 0 {
        touchId := touchIds[0]
        mouseX, mouseY = ebiten.TouchPosition(touchId)
        leftClick = true
    }
    */

    pressedTouchIds := inpututil.AppendJustPressedTouchIDs(nil)
    if len(pressedTouchIds) > 0 {
        touchId := pressedTouchIds[0]
        ui.lastTouchX, ui.lastTouchY = ebiten.TouchPosition(touchId)
        ui.touchStartTime = ui.Counter
    }

    touchIds := inpututil.AppendJustReleasedTouchIDs(nil)
    if len(touchIds) > 0 {
        // touchId := touchIds[0]

        duration := ui.Counter - ui.touchStartTime

        // log.Printf("Touch %v duration %v", touchId, duration)

        if duration < 40 {
            leftClick = true
            leftClickReleased = true
        } else {
            rightClick = true
        }

        mouseX, mouseY = ui.lastTouchX, ui.lastTouchY
        // log.Printf("Touch %v %v %v", touchId, mouseX, mouseY)
        // log.Printf("Touches %v", touchIds)
    }

    /*
    if inpututil.IsTouchJustReleased(0) {
        leftClickReleased = true
        mouseX, mouseY = ebiten.TouchPosition(0)
    }
    */

    if leftClickReleased {
        for _, element := range ui.LeftClickedElements {
            element.isPressed = false
            if element.LeftClickRelease != nil {
                element.LeftClickRelease(element)
            }
        }

        ui.LeftClickedElements = nil
    }

    var keepDoubleClick []doubleClick
    for _, candidate := range ui.doubleClickCandidates {
        if ui.Counter - candidate.Time < DoubleClickThreshold {
            keepDoubleClick = append(keepDoubleClick, candidate)
        }
    }
    ui.doubleClickCandidates = keepDoubleClick

    elementLeftClicked := false

    wheelX, wheelY := inputmanager.Wheel()

    mouseX, mouseY = scale.Unscale2(mouseX, mouseY)

    lastTooltip := ui.TooltipElement

    ui.TooltipElement = nil

    for _, element := range slices.Backward(ui.GetHighestLayer()) {
        if image.Pt(mouseX, mouseY).In(element.Rect) {
            if lastTooltip != element {
                ui.TooltipTime = ui.Counter
            }

            ui.TooltipElement = element
            ui.TooltipPosition = image.Pt(mouseX, mouseY)

            if element.Inside != nil {
                element.Inside(element, mouseX - element.Rect.Min.X, mouseY - element.Rect.Min.Y)
            }

            if element.Scroll != nil {
                element.Scroll(element, wheelX, wheelY)
            }

            if !ui.Disabled && leftClick && !elementLeftClicked {
                // if the element is interested in left click at all
                if element.LeftClick != nil || element.LeftClickRelease != nil || element.DoubleLeftClick != nil {
                    elementLeftClicked = true
                    element.isPressed = true
                }
                if element.LeftClick != nil {
                    if element.PlaySoundLeftClick {
                        ui.PlayStandardSound()
                    }
                    element.LeftClick(element)
                }

                // might release the click on the same tick, due to touch input
                if leftClickReleased {
                    if element.LeftClickRelease != nil {
                        element.LeftClickRelease(element)
                    }
                } else {
                    ui.LeftClickedElements = append(ui.LeftClickedElements, element)
                }

                if ui.focusedElement != element {
                    if ui.focusedElement != nil && ui.focusedElement.LoseFocus != nil {
                        ui.focusedElement.LoseFocus(ui.focusedElement)
                        ui.textField.Blur()
                    }

                    ui.focusedElement = element
                    if ui.focusedElement.TextEntry != nil {
                        ui.textField.Focus()
                    }

                    if element.GainFocus != nil {
                        element.GainFocus(element)
                    }
                }

                addDoubleClick := true

                for i, candidate := range ui.doubleClickCandidates {
                    if candidate.Element == element {
                        diff := ui.Counter - candidate.Time
                        if diff < DoubleClickThreshold && element.DoubleLeftClick != nil {
                            element.DoubleLeftClick(element)
                            // an alternative here is just to set ui.doubleClickCandidates[i].Element = nil
                            // and let the list be cleaned up later
                            ui.doubleClickCandidates = slices.Delete(ui.doubleClickCandidates, i, i + 1)
                        }

                        addDoubleClick = false
                        break
                    }
                }

                if addDoubleClick {
                    ui.doubleClickCandidates = append(ui.doubleClickCandidates, doubleClick{Element: element, Time: ui.Counter})
                }

            }
            if !ui.Disabled && rightClick && element.RightClick != nil {
                element.RightClick(element)
            }
        } else {
            if element.NotInside != nil {
                element.NotInside(element)
            }
        }
    }

    if ui.focusedElement != nil {
        var err error
        handled := false

        if ui.focusedElement.TextEntry != nil {
            if !ui.textField.IsFocused() {
                ui.textField.Focus()
            }
            // log.Printf("text field is focused %v", ui.textField.IsFocused())
            if ui.textField.IsFocused() {
                /*
                start, end := ui.textField.Selection()
                if start != 0 || end != 0 {
                    // log.Printf("selection %v %v", start, end)
                    // ui.textField.SetTextAndSelection(ui.textField.Text(), end, end)
                }
                */

                bounds := ui.focusedElement.Rect.Bounds()
                handled, err = ui.textField.HandleInput(bounds.Min.X, bounds.Max.Y + 1)
                // log.Printf("Handle input %v", err)
                if err != nil {
                    log.Printf("input error %v", err)
                }

                doEnter := false

                if !handled {
                    if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
                        // ui.focusedElement.HandleKeys([]ebiten.Key{ebiten.KeyBackspace})
                        start, _ := ui.textField.Selection()
                        if start > 0 {
                            text := ui.textField.TextForRendering()
                            ui.textField.SetTextAndSelection(text[0:len(text)-1], start - 1, start - 1)
                        }
                    } else if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
                        doEnter = true
                    }
                }

                for strings.Contains(ui.textField.Text(), "\n") {
                    start, _ := ui.textField.Selection()
                    text := ui.textField.TextForRendering()
                    ui.textField.SetTextAndSelection(text[0:len(text)-1], start - 1, start - 1)
                    doEnter = true
                }

                out := ui.focusedElement.TextEntry(ui.focusedElement, ui.textField.TextForRendering())
                if out != ui.textField.TextForRendering() {
                    ui.textField.SetTextAndSelection(out, len(out), len(out))
                }

                if doEnter {
                    ui.focusedElement.HandleKeys([]ebiten.Key{ebiten.KeyEnter})
                }
            }
        }

        /*
        if !handled && ui.focusedElement != nil && ui.focusedElement.TextEntry != nil {
            chars := ebiten.AppendInputChars(nil)
            ui.focusedElement.TextEntry(ui.focusedElement, chars)
        }
        */
    }

    if !ui.Disabled && leftClick && !elementLeftClicked {
        for _, element := range ui.GetHighestLayer() {
            if element.NotLeftClicked != nil {
                element.NotLeftClicked(element)
            }
        }
    }

    if leftClick && !elementLeftClicked && ui.LeftClick != nil {
        ui.LeftClick()
    }
}
