package ui

import (
    "image"
    "github.com/hajimehoshi/ebiten/v2"
)

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

