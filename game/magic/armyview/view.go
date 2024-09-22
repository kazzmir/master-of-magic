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
    FirstRow int
}

func MakeArmyScreen(cache *lbx.LbxCache, player *playerlib.Player) *ArmyScreen {
    view := &ArmyScreen{
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        Player: player,
        State: ArmyScreenStateRunning,
        FirstRow: 0,
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

    // row := view.FirstRow
    rowY := 25
    rowCount := 0
    for i, stack := range view.Player.Stacks {
        if i < view.FirstRow {
            continue
        }

        options.GeoM.Reset()
        options.GeoM.Translate(78, float64(rowY))

        for _, unit := range stack.Units() {
            pic, _ := view.ImageCache.GetImage(unit.Unit.LbxFile, unit.Unit.Index, 0)
            if pic != nil {
                screen.DrawImage(pic, &options)
                options.GeoM.Translate(float64(pic.Bounds().Dx()) + 1, 0)
            }
        }

        // there are only 6 slots to show at a time
        rowCount += 1
        if rowCount > 6 {
            break
        }

        rowY += 22
    }

}
