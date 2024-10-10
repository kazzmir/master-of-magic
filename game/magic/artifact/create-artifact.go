package artifact

import (
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
    // "github.com/hajimehoshi/ebiten/v2/inpututil"
)

func ShowCreateArtifactScreen(yield coroutine.YieldFunc, cache *lbx.LbxCache, draw *func(*ebiten.Image)) {
    quit := false

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
        },
    }

    *draw = func(screen *ebiten.Image) {
        ui.Draw(ui, screen)
    }

    for !quit {
        yield()
    }

}
