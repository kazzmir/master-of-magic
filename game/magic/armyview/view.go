package armyview

import (
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"

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
}

func MakeArmyScreen(cache *lbx.LbxCache, player *playerlib.Player) *ArmyScreen {
    view := &ArmyScreen{
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        Player: player,
        State: ArmyScreenStateRunning,
    }

    return view
}

func (view *ArmyScreen) Update() ArmyScreenState {
    return view.State
}

func (view *ArmyScreen) Draw(screen *ebiten.Image) {
    background, _ := view.ImageCache.GetImage("armylist.lbx", 0, 0)
    var options ebiten.DrawImageOptions
    screen.DrawImage(background, &options)
}
