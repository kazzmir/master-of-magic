package armyview

import (
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/hajimehoshi/ebiten/v2"
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
    for i, stack := range view.Player.Stacks {
        if i < view.FirstRow {
            continue
        }

        x := 78

        for _, unit := range stack.Units() {
            elementX := float64(x)
            elementY := float64(rowY)

            pic, _ := view.ImageCache.GetImage(unit.Unit.LbxFile, unit.Unit.Index, 0)
            if pic != nil {
                elements = append(elements, &uilib.UIElement{
                    Draw: func(this *uilib.UIElement, screen *ebiten.Image) {
                        var options ebiten.DrawImageOptions
                        options.GeoM.Translate(elementX, elementY)
                        if pic != nil {
                            screen.DrawImage(pic, &options)
                        }
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
