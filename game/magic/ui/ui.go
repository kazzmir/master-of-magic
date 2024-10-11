package ui

import (
    // "log"
    "image"
    "slices"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type UIInsideElementFunc func(element *UIElement, x int, y int)
type UINotInsideElementFunc func(element *UIElement)
type UIClickElementFunc func(element *UIElement)
type UIDrawFunc func(element *UIElement, window *ebiten.Image)
type UIKeyFunc func(key []ebiten.Key)
type UIGainFocusFunc func(*UIElement)
type UILoseFocusFunc func(*UIElement)
type UITextEntry func(*UIElement, []rune)

type UILayer int

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

    // fires when this element is left clicked
    GainFocus UIGainFocusFunc
    // fires when some other element is left clicked
    LoseFocus UILoseFocusFunc
    // fires when the user types some keys and this element is focused
    TextEntry UITextEntry
    // fires when a key is pressed and this element is focused
    HandleKeys UIKeyFunc

    Draw UIDrawFunc
    Layer UILayer
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
    // keep track of the minimum and maximum keys so we don't have to sort
    minLayer UILayer
    maxLayer UILayer
    Draw func(*UI, *ebiten.Image)
    HandleKeys UIKeyFunc
    Counter uint64

    focusedElement *UIElement

    doubleClickCandidates []doubleClick

    LeftClickedElements []*UIElement

    Delays []UIDelay

    // disabled so that the zero value is enabled
    Disabled bool
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

func (ui *UI) AddElement(element *UIElement){
    if element.Layer < ui.minLayer {
        ui.minLayer = element.Layer
    }
    if element.Layer > ui.maxLayer {
        ui.maxLayer = element.Layer
    }

    ui.Elements[element.Layer] = append(ui.Elements[element.Layer], element)
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
    for i := ui.minLayer; i <= ui.maxLayer; i++ {
        for _, element := range ui.Elements[i] {
            f(element)
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

func (ui *UI) StandardUpdate() {
    ui.Counter += 1

    var keepDelays []UIDelay

    for _, delay := range ui.Delays {
        if ui.Counter <= delay.Time {
            keepDelays = append(keepDelays, delay)
        } else {
            delay.Func()
        }
    }
    ui.Delays = keepDelays

    if !ui.Disabled {
        keys := inpututil.AppendJustPressedKeys(nil)
        if len(keys) > 0 {
            if ui.HandleKeys != nil {
                ui.HandleKeys(keys)
            }

            if ui.focusedElement != nil && ui.focusedElement.HandleKeys != nil {
                ui.focusedElement.HandleKeys(keys)
            }
        }
    }

    leftClick := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
    leftClickReleased := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
    rightClick := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)

    mouseX, mouseY := ebiten.CursorPosition()

    if leftClickReleased {
        for _, element := range ui.LeftClickedElements {
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

    for _, element := range ui.GetHighestLayer() {
        if image.Pt(mouseX, mouseY).In(element.Rect) {
            if element.Inside != nil {
                element.Inside(element, mouseX - element.Rect.Min.X, mouseY - element.Rect.Min.Y)
            }
            if !ui.Disabled && leftClick {
                elementLeftClicked = true
                if element.LeftClick != nil {
                    element.LeftClick(element)
                }
                ui.LeftClickedElements = append(ui.LeftClickedElements, element)

                if ui.focusedElement != element {
                    if ui.focusedElement != nil && ui.focusedElement.LoseFocus != nil {
                        ui.focusedElement.LoseFocus(ui.focusedElement)
                    }

                    ui.focusedElement = element
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

    if ui.focusedElement != nil && ui.focusedElement.TextEntry != nil {
        chars := ebiten.AppendInputChars(nil)
        ui.focusedElement.TextEntry(ui.focusedElement, chars)
    }

    if !ui.Disabled && leftClick && !elementLeftClicked {
        for _, element := range ui.GetHighestLayer() {
            if element.NotLeftClicked != nil {
                element.NotLeftClicked(element)
            }
        }
    }
}
