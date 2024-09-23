package citylistview

import (
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/hajimehoshi/ebiten/v2"
)

type CityListScreenState int

const (
    CityListScreenStateRunning CityListScreenState = iota
    CityListScreenStateDone
)

type CityListScreen struct {
    Cache *lbx.LbxCache
    Player *playerlib.Player
    ImageCache util.ImageCache
    UI *uilib.UI
    State CityListScreenState
}

func MakeCityListScreen(cache *lbx.LbxCache, player *playerlib.Player) *CityListScreen {
    view := &CityListScreen{
        Cache: cache,
        Player: player,
        ImageCache: util.MakeImageCache(cache),
        State: CityListScreenStateRunning,
    }

    view.UI = view.MakeUI()

    return view
}

func (view *CityListScreen) MakeUI() *uilib.UI {
    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image) {
            background, _ := view.ImageCache.GetImage("reload.lbx", 21, 0)
            var options ebiten.DrawImageOptions
            screen.DrawImage(background, &options)

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
    }

    var elements []*uilib.UIElement
    ui.SetElementsFromArray(elements)

    return ui
}

func (view *CityListScreen) Update() CityListScreenState {
    view.UI.StandardUpdate()
    return view.State
}

func (view *CityListScreen) Draw(screen *ebiten.Image) {
    view.UI.Draw(view.UI, screen)
}
