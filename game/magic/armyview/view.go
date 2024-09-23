package armyview

import (
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

type ArmyScreenState int

const (
    ArmyScreenStateRunning ArmyScreenState = iota
    ArmyScreenStateDone
)

type ArmyScreen struct {
    Cache *lbx.LbxCache
    ImageCache util.ImageCache
    Player *playerlib.Player
    State ArmyScreenState
    FirstRow int
    UI *uilib.UI
}

func MakeArmyScreen(cache *lbx.LbxCache, player *playerlib.Player) *ArmyScreen {
    view := &ArmyScreen{
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        Player: player,
        State: ArmyScreenStateRunning,
        FirstRow: 0,
    }

    view.UI = view.MakeUI()

    return view
}

func (view *ArmyScreen) MakeUI() *uilib.UI {
    ui := &uilib.UI{
        Draw: func(this *uilib.UI, screen *ebiten.Image) {
            background, _ := view.ImageCache.GetImage("armylist.lbx", 0, 0)
            var options ebiten.DrawImageOptions
            screen.DrawImage(background, &options)

            this.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
    }

    var elements []*uilib.UIElement

    // row := view.FirstRow
    rowY := 25
    rowCount := 0

    highlightColor := util.PremultiplyAlpha(color.RGBA{R: 255, G: 255, B: 255, A: 90})

    for i, stack := range view.Player.Stacks {
        if i < view.FirstRow {
            continue
        }

        x := 78

        for _, unit := range stack.Units() {
            elementX := float64(x)
            elementY := float64(rowY)

            highlighted := false
            pic, _ := view.ImageCache.GetImage(unit.Unit.LbxFile, unit.Unit.Index, 0)
            if pic != nil {
                elements = append(elements, &uilib.UIElement{
                    Rect: util.ImageRect(int(elementX), int(elementY), pic),
                    Inside: func (this *uilib.UIElement, x, y int){
                        highlighted = true
                    },
                    NotInside: func (this *uilib.UIElement){
                        highlighted = false
                    },
                    Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                        var options ebiten.DrawImageOptions
                        options.GeoM.Translate(elementX, elementY)

                        if highlighted {
                            x, y := options.GeoM.Apply(0, 0)
                            vector.DrawFilledRect(screen, float32(x), float32(y+1), float32(pic.Bounds().Dx()), float32(pic.Bounds().Dy())-1, highlightColor, false)
                        }

                        screen.DrawImage(pic, &options)
                    },
                })
                x += pic.Bounds().Dx() + 1
            }
        }

        // there are only 6 slots to show at a time
        rowCount += 1
        if rowCount > 6 {
            break
        }

        rowY += 22
    }

    ui.SetElementsFromArray(elements)

    return ui
}

func (view *ArmyScreen) Update() ArmyScreenState {
    view.UI.StandardUpdate()
    return view.State
}

func (view *ArmyScreen) Draw(screen *ebiten.Image) {
    view.UI.Draw(view.UI, screen)

}
