package summon

import (
    "log"
    "image"
    "image/color"

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
    Wizard data.WizardBase
    State SummonState
    CircleBack *util.Animation
    CircleFront *util.Animation
    Background *ebiten.Image
    Monster *ebiten.Image
    SummonHeight int
}

func MakeSummonUnit(cache *lbx.LbxCache, unit units.Unit, wizard data.WizardBase) *SummonUnit {
    summon := &SummonUnit{
        Cache: cache,
        ImageCache: util.MakeImageCache(cache),
        Wizard: wizard,
        State: SummonStateRunning,
    }

    monsterIndex := 0
    // magic spirit is monster.lbx, 0
    if unit.Equals(units.MagicSpirit) {
        monsterIndex = 0
    } else if unit.Equals(units.HellHounds) {
        monsterIndex = 1
    } else if unit.Equals(units.Gargoyle) {
        monsterIndex = 2
    } else if unit.Equals(units.FireGiant) {
        monsterIndex = 3
    } else if unit.Equals(units.ChaosSpawn) {
        monsterIndex = 5
    } else if unit.Equals(units.Chimeras) {
        monsterIndex = 6
    } else if unit.Equals(units.DoomBat) {
        monsterIndex = 7
    } else if unit.Equals(units.Efreet) {
        monsterIndex = 8
    } else if unit.Equals(units.Hydra) {
        monsterIndex = 9
    } else if unit.Equals(units.GreatDrake) {
        monsterIndex = 10
    } else if unit.Equals(units.Skeleton) {
        monsterIndex = 11
    } else if unit.Equals(units.Ghoul) {
        monsterIndex = 12
    }

    monsterPicture, err := summon.ImageCache.GetImage("monster.lbx", monsterIndex, 0)
    if err != nil {
        log.Printf("Error: could not load monster image at index %v: %v", monsterIndex, err)
    }

    baseColor := color.RGBA{R: 0, B: 0, G: 0xff, A: 0xff}
    switch unit.Realm {
        case data.LifeMagic: baseColor = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
        case data.SorceryMagic: baseColor = color.RGBA{R: 0, G: 0, B: 0xff, A: 0xff}
        case data.NatureMagic: baseColor = color.RGBA{R: 0, B: 0, G: 0xff, A: 0xff}
        case data.DeathMagic: baseColor = color.RGBA{R: 0xd6, G: 0x63, B: 0xff, A: 0xff}
        case data.ChaosMagic: baseColor = color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}
        case data.ArcaneMagic: baseColor = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
    }

    updateColors := func (img *image.Paletted) image.Image {
        // 228-245 remap colors
        // colorRange := 245 - 226

        newPalette := make(color.Palette, len(img.Palette))
        copy(newPalette, img.Palette)
        img.Palette = newPalette

        light := 0
        for i := 225; i <= 247; i++ {
            img.Palette[i] = util.Lighten(baseColor, float64(light))
            light -= 4
        }

        /*
        img.Palette[227] = color.RGBA{R: 0, G: 0, B: 0, A: 0}
        img.Palette[228] = color.RGBA{R: 0, G: 0, B: 0, A: 0}
        img.Palette[237] = color.RGBA{R: 0, G: 0, B: 0, A: 0}
        img.Palette[238] = color.RGBA{R: 0, G: 0, B: 0, A: 0}
        img.Palette[239] = color.RGBA{R: 0, G: 0, B: 0, A: 0}
        */

        return img
    }

    summonBack, _ := summon.ImageCache.GetImagesTransform("spellscr.lbx", 10, updateColors)
    summon.CircleBack = util.MakeAnimation(summonBack, true)

    summonFront, _ := summon.ImageCache.GetImagesTransform("spellscr.lbx", 11, updateColors)
    summon.CircleFront = util.MakeAnimation(summonFront, true)

    background, _ := summon.ImageCache.GetImageTransform("spellscr.lbx", 9, 0, updateColors)
    summon.Background = background

    summon.Monster = monsterPicture

    return summon
}

func (summon *SummonUnit) Update() SummonState {
    summon.Counter += 1

    if summon.Counter % 7 == 0 {
        summon.CircleBack.Next()
        summon.CircleFront.Next()
    }

    if summon.Counter % 2 == 0 {
        if summon.SummonHeight < summon.Monster.Bounds().Dy() {
            summon.SummonHeight += 1
        }
    }

    return summon.State
}

func (summon *SummonUnit) Draw(screen *ebiten.Image){

    // background, _ := summon.ImageCache.GetImage("spellscr.lbx", 9, 0)
    var options ebiten.DrawImageOptions
    options.GeoM.Translate(70, 20)
    screen.DrawImage(summon.Background, &options)

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
    circleOptions.GeoM.Translate(53, 54)
    screen.DrawImage(summon.CircleBack.Frame(), &circleOptions)

    wizard, _ := summon.ImageCache.GetImage("spellscr.lbx", wizardIndex, 0)
    wizardOptions := options
    wizardOptions.GeoM.Translate(7, 3)
    screen.DrawImage(wizard, &wizardOptions)

    monster := summon.Monster
    monsterOptions := options
    monsterOptions.GeoM.Translate(75, 30 + 70 - float64(summon.SummonHeight))
    partialMonster := monster.SubImage(image.Rect(0, 0, monster.Bounds().Dx(), summon.SummonHeight)).(*ebiten.Image)
    screen.DrawImage(partialMonster, &monsterOptions)

    circleOptions.GeoM.Translate(11, 26)
    circleOptions.ColorScale.ScaleAlpha(1.0)
    screen.DrawImage(summon.CircleFront.Frame(), &circleOptions)
}
