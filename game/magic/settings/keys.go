package settings

import (
    "context"
    "fmt"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/keybinds"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

const keysLayer = uilib.UILayer(6)
const keyPressLayer = uilib.UILayer(7)
const confirmLayer = uilib.UILayer(8)

// runNestedUI adds group to parentUI, pumps it each frame via yield until quit is
// cancelled, then removes it. Mirrors game.go's doRunUI, which needs *Game and so
// can't be reused directly from this package.
func runNestedUI(yield coroutine.YieldFunc, parentUI *uilib.UI, group *uilib.UIElementGroup, quit context.Context) {
    parentUI.AddGroup(group)
    defer parentUI.RemoveGroup(group)

    // let the click that opened this screen finish out its frame before this
    // group starts processing input, otherwise an element here that happens to
    // overlap the triggering click's screen position fires immediately.
    yield()

    for quit.Err() == nil {
        parentUI.StandardUpdate()
        if yield() != nil {
            break
        }
    }
}

func keyDisplayName(key ebiten.Key) string {
    if key == keybinds.Unbound {
        return "---"
    }

    return key.String()
}

// waitForKeyPress shows a modal asking the player to press a key to bind to the
// given action, blocking until a key is pressed. Returns the key and true, or
// an unspecified key and false if the player pressed Escape to cancel.
func waitForKeyPress(yield coroutine.YieldFunc, parentUI *uilib.UI, cache *lbx.LbxCache, actionName string) (ebiten.Key, bool) {
    fonts := fontslib.MakeSettingsFonts(cache)

    quit, cancel := context.WithCancel(context.Background())
    cancelled := false
    var pressed ebiten.Key

    boxRect := image.Rect(50, 85, 270, 118)

    element := &uilib.UIElement{
        Layer: keyPressLayer,
        HandleKeys: func(keys []ebiten.Key){
            if len(keys) == 0 {
                return
            }

            key := keys[0]
            if key == ebiten.KeyEscape {
                cancelled = true
            } else {
                pressed = key
            }
            cancel()
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            vector.FillRect(screen, float32(scale.Scale(boxRect.Min.X)), float32(scale.Scale(boxRect.Min.Y)), float32(scale.Scale(boxRect.Dx())), float32(scale.Scale(boxRect.Dy())), color.NRGBA{A: 230}, false)
            util.DrawRect(screen, scale.ScaleRect(boxRect), color.NRGBA{R: 255, G: 200, B: 100, A: 255})

            var options ebiten.DrawImageOptions
            fonts.OptionFont.PrintOptions(screen, float64(boxRect.Min.X + 6), float64(boxRect.Min.Y + 6), font.FontOptions{Scale: scale.ScaleAmount, DropShadow: true, Options: &options}, fmt.Sprintf("Press the key to bind to %v,", actionName))
            fonts.OptionFont.PrintOptions(screen, float64(boxRect.Min.X + 6), float64(boxRect.Min.Y + 18), font.FontOptions{Scale: scale.ScaleAmount, DropShadow: true, Options: &options}, "or Escape to cancel.")
        },
    }

    group := uilib.MakeGroup()
    group.AddElement(element)

    parentUI.AddGroup(group)
    parentUI.FocusElement(element, "")
    defer parentUI.RemoveGroup(group)
    defer parentUI.UnfocusElement()

    // see runNestedUI - let the triggering click's frame finish before this
    // group starts processing input.
    yield()

    for quit.Err() == nil {
        parentUI.StandardUpdate()
        if yield() != nil {
            break
        }
    }

    return pressed, !cancelled
}

// confirmRebind pops the standard game confirm dialog before applying a rebind
// that would kick a key out from under another action. It matches the original
// game's behaviour: the previous owner is unbound and the new action takes the
// key. Returns true when the player confirmed (the caller performs the swap), or
// false when they cancelled (leave the bindings unchanged). Mirrors the nested-UI
// pump of waitForKeyPress: add a group and pump it via yield until it is gone.
func confirmRebind(yield coroutine.YieldFunc, parentUI *uilib.UI, cache *lbx.LbxCache, imageCache *util.ImageCache, actionName, otherName string, key ebiten.Key) bool {
    group := uilib.MakeGroup()
    parentUI.AddGroup(group)
    defer parentUI.RemoveGroup(group)

    quit, cancel := context.WithCancel(context.Background())
    accepted := false

     // the dialog's container and its elements must be the same group: the
     // confirm buttons' fade/removal are driven by this group's delays, which
     // only fire when its elements live in this group. It is rendered on
     // confirmLayer so it sits above the keys screen instead of being hidden
     // behind it.
    group.AddElements(uilib.MakeConfirmDialogWithLayer(group, cache, imageCache, confirmLayer,
        fmt.Sprintf("Key %v is already bound to %q. Bind %q to it and unbind it from %q?", key.String(), otherName, actionName, otherName),
        true,
        func() {
            accepted = true
            cancel()
        },
        cancel,
    ))

     // let the click that opened this screen finish out its frame before the
     // confirm group starts processing input, otherwise a same-frame click on the
     // confirm group would fire immediately.
    yield()

    for quit.Err() == nil {
        parentUI.StandardUpdate()
        if yield() != nil {
            break
          }
      }

    return accepted
}

// MakeKeysUI builds the list of rebindable actions and their current key.
// Clicking a row opens waitForKeyPress to rebind it; if the chosen key is
// already held by another action, confirmRebind asks whether to unbind the
// previous owner. Back/Ok just close the screen - there is no separate
// save/cancel state for the screen as a whole.
func MakeKeysUI(yield coroutine.YieldFunc, parentUI *uilib.UI, cache *lbx.LbxCache, imageCache *util.ImageCache, keybindings *keybinds.Keybindings) (*uilib.UIElementGroup, context.Context) {
    fonts := fontslib.MakeSettingsFonts(cache)
    background, _ := imageCache.GetImage("load.lbx", 11, 0)

    group := uilib.MakeGroup()
    quit, cancel := context.WithCancel(context.Background())

    group.AddElement(&uilib.UIElement{
        Layer: keysLayer,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            scale.DrawScaled(screen, background, &options)
        },
    })

    const rowWidth = 150
    const rowHeight = 9
    const leftX = 10
    const rightX = 165
    const startY = 42

    makeRow := func(action keybinds.Action, x int, y int) *uilib.UIElement {
        return &uilib.UIElement{
            Layer: keysLayer,
            Rect: image.Rect(x, y, x + rowWidth, y + rowHeight),
            LeftClick: func(element *uilib.UIElement){
                key, ok := waitForKeyPress(yield, parentUI, cache, action.Name())
                if !ok {
                    return
                }

                    // If another action currently owns this key, ask the player how to
                    // proceed before applying the rebind so a keypress can never fire
                    // two actions at once.
                if other, hasConflict := keybindings.ConflictingActionForKey(key); hasConflict && other != action {
                    if !confirmRebind(yield, parentUI, cache, imageCache, action.Name(), other.Name(), key) {
                        return
                    }

                    // confirmed: unbind the previous owner first, matching the original
                    // game, then bind the key to this action.
                    keybindings.Set(other, keybinds.Unbound)
                }

                keybindings.Set(action, key)
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                fonts.OptionFont.PrintOptions(screen, float64(x), float64(y), font.FontOptions{Scale: scale.ScaleAmount, DropShadow: true, Options: &options}, action.Name())
                fonts.OptionFont.PrintOptions(screen, float64(x + 112), float64(y), font.FontOptions{Scale: scale.ScaleAmount, DropShadow: true, Options: &options}, keyDisplayName(keybindings.Get(action)))
            },
        }
    }

    half := (len(keybinds.AllActions) + 1) / 2
    for i, action := range keybinds.AllActions {
        x := leftX
        row := i
        if i >= half {
            x = rightX
            row = i - half
        }
        group.AddElement(makeRow(action, x, startY + row * rowHeight))
    }

    makeCloseButton := func(x int, y int, label string) *uilib.UIElement {
        rect := image.Rect(x, y, x + 40, y + 13)
        return &uilib.UIElement{
            Layer: keysLayer,
            Rect: rect,
            LeftClick: func(element *uilib.UIElement){
                cancel()
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                vector.FillRect(screen, float32(scale.Scale(rect.Min.X)), float32(scale.Scale(rect.Min.Y)), float32(scale.Scale(rect.Dx())), float32(scale.Scale(rect.Dy())), color.NRGBA{R: 96, G: 60, B: 20, A: 255}, false)
                util.DrawRect(screen, scale.ScaleRect(rect), color.NRGBA{R: 255, G: 200, B: 100, A: 255})

                var options ebiten.DrawImageOptions
                fonts.OptionFont.PrintOptions(screen, float64(rect.Min.X + 10), float64(rect.Min.Y + 3), font.FontOptions{Scale: scale.ScaleAmount, DropShadow: true, Options: &options}, label)
            },
        }
    }

    makeActionButton := func(x int, y int, label string, rectW int, onClick func()) *uilib.UIElement {
        rect := image.Rect(x, y, x + rectW, y + 13)
        return &uilib.UIElement{
            Layer: keysLayer,
            Rect: rect,
            LeftClick: func(element *uilib.UIElement){
                onClick()
             },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                vector.FillRect(screen, float32(scale.Scale(rect.Min.X)), float32(scale.Scale(rect.Min.Y)), float32(scale.Scale(rect.Dx())), float32(scale.Scale(rect.Dy())), color.NRGBA{R: 96, G: 60, B: 20, A: 255}, false)
                util.DrawRect(screen, scale.ScaleRect(rect), color.NRGBA{R: 255, G: 200, B: 100, A: 255})

                var options ebiten.DrawImageOptions
                fonts.OptionFont.PrintOptions(screen, float64(rect.Min.X + 6), float64(rect.Min.Y + 3), font.FontOptions{Scale: scale.ScaleAmount, DropShadow: true, Options: &options}, label)
             },
        }
    }

     // sits in the empty band between the action rows and the Back/Ok buttons, so
     // a player can restore the original game's default bindings for every action
     // in one click instead of re-pressing each row by hand.
    group.AddElement(makeActionButton(10, 150, "Reset To Defaults", 180, func() {
        keybindings.ResetToDefaults()
         }))

    // matches the two button slots already baked into the load.lbx background art
    // (the same ones the settings screen's own Keys/Ok buttons sit on)
    group.AddElement(makeCloseButton(214, 176, "Back"))
    group.AddElement(makeCloseButton(266, 176, "Ok"))

    return group, quit
}
