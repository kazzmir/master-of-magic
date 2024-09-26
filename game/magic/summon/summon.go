package summon

import (
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/util"

    "github.com/hajimehoshi/ebiten/v2"
)

type SummonState int
const (
    SummonStateRunning SummonState = iota
    SummonStateDone
)

type SummonUnit struct {
    Counter uint64
    Cache *lbx.LbxCache
    ImageCache util.ImageCache
    Unit units.Unit
    Wizard data.WizardBase
    State SummonState
}

func MakeSummonUnit(cache *lbx.LbxCache, unit units.Unit, wizard data.WizardBase) *SummonUnit {
    return &SummonUnit{
        Unit: unit,
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        Wizard: wizard,
        State: SummonStateRunning,
    }
}

func (summon *SummonUnit) Update() SummonState {
    summon.Counter += 1
    return summon.State
}

func (summon *SummonUnit) Draw(screen *ebiten.Image){
}
