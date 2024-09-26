package summon

import (
    "image"

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
    CircleBack *util.Animation
    CircleFront *util.Animation
    SummonHeight int
}

func MakeSummonUnit(cache *lbx.LbxCache, unit units.Unit, wizard data.WizardBase) *SummonUnit {
    summon := &SummonUnit{
        Unit: unit,
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        Wizard: wizard,
        State: SummonStateRunning,
    }

    summonBack, _ := summon.ImageCache.GetImages("spellscr.lbx", 10)
    summon.CircleBack = util.MakeAnimation(summonBack, true)

    summonFront, _ := summon.ImageCache.GetImages("spellscr.lbx", 11)
    summon.CircleFront = util.MakeAnimation(summonFront, true)

    return summon
}

func (summon *SummonUnit) Update() SummonState {
    summon.Counter += 1

    if summon.Counter % 8 == 0 {
        summon.CircleBack.Next()
        summon.CircleFront.Next()
    }

    if summon.Counter % 2 == 0 {
        // kind of a hack, but the summon images are all 80px in height
        if summon.SummonHeight < 80 {
            summon.SummonHeight += 1
        }
    }

    return summon.State
}

func (summon *SummonUnit) Draw(screen *ebiten.Image){
    // magic spirit is monster.lbx, 0

    monsterIndex := 0

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

    circleOptions := options
    circleOptions.GeoM.Translate(50, 53)
    screen.DrawImage(summon.CircleBack.Frame(), &circleOptions)

    wizard, _ := summon.ImageCache.GetImage("spellscr.lbx", wizardIndex, 0)
    wizardOptions := options
    wizardOptions.GeoM.Translate(7, 3)
    screen.DrawImage(wizard, &wizardOptions)

    monster, _ := summon.ImageCache.GetImage("monster.lbx", monsterIndex, 0)
    monsterOptions := options
    monsterOptions.GeoM.Translate(75, 30 + 70 - float64(summon.SummonHeight))
    partialMonster := monster.SubImage(image.Rect(0, 0, monster.Bounds().Dx(), summon.SummonHeight)).(*ebiten.Image)
    screen.DrawImage(partialMonster, &monsterOptions)

    circleOptions.GeoM.Translate(10, 30)
    screen.DrawImage(summon.CircleFront.Frame(), &circleOptions)
}
