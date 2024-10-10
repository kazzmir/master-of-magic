package artifact

import (
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
    // "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type ArtifactType int
const (
    ArtifactTypeNone ArtifactType = iota
    ArtifactTypeSword
    ArtifactTypeMace
    ArtifactTypeAxe
    ArtifactTypeBow
    ArtifactTypeStaff
    ArtifactTypeWand
    ArtifactTypeMisc
    ArtifactTypeShield
    ArtifactTypeChain
    ArtifactTypePlate
)

type Power interface {
}

type PowerAttack struct {
    Amount int
}

type PowerDefense struct {
    Amount int
}

type PowerToHit struct {
    Amount int
}

type PowerSpellSkill struct {
    Amount int
}

type PowerSpellSave struct {
    Amount int
}

type PowerResistance struct {
    Amount int
}

type PowerMovement struct {
    Amount int
}

type PowerSpellCharges struct {
    Spell spellbook.Spell
    Charges int
}

type Artifact struct {
    Type ArtifactType
    Image int
    Name string
    Powers []Power
}

/* returns the artifact that was created and true,
 * otherwise false for cancelled
 */
func ShowCreateArtifactScreen(yield coroutine.YieldFunc, cache *lbx.LbxCache, draw *func(*ebiten.Image)) (*Artifact, bool) {
    quit := false

    imageCache := util.MakeImageCache(cache)

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            background, _ := imageCache.GetImage("spellscr.lbx", 13, 0)
            screen.DrawImage(background, &options)

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
    }

    *draw = func(screen *ebiten.Image) {
        ui.Draw(ui, screen)
    }

    for !quit {
        yield()
    }

    return nil, false
}
