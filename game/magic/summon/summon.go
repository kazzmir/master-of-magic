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
    // magic spirit is monster.lbx, 0

    background, _ := summon.ImageCache.GetImage("spellscr.lbx", 9, 0)
    var options ebiten.DrawImageOptions
    options.GeoM.Translate(70, 20)
    screen.DrawImage(background, &options)

    wizardIndex := 46
    switch summon.Wizard {
        case data.WizardMerlin: wizardIndex = 46
        case data.WizardRaven: wizardIndex = 47
        case data.WizardSharee: wizardIndex = 48
        case data.WizardLoPan: wizardIndex = 49
        case data.WizardJafar: wizardIndex = 50
        case data.WizardOberic: wizardIndex = 51
        case data.WizardRjak: wizardIndex = 52
        case data.WizardSssra: wizardIndex = 53
        case data.WizardTauron: wizardIndex = 54
        case data.WizardFreya: wizardIndex = 55
        case data.WizardHorus: wizardIndex = 56
        case data.WizardAriel: wizardIndex = 57
        case data.WizardTlaloc: wizardIndex = 58
        case data.WizardKali: wizardIndex = 59
    }

    wizard, _ := summon.ImageCache.GetImage("spellscr.lbx", wizardIndex, 0)
    wizardOptions := options
    wizardOptions.GeoM.Translate(7, 3)
    screen.DrawImage(wizard, &wizardOptions)
}
