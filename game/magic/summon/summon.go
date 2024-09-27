package summon

import (
    "fmt"
    "log"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
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
    Font *font.Font
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
        Unit: unit,
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
    } else if unit.Equals(units.NightStalker) {
        monsterIndex = 13
    } else if unit.Equals(units.WereWolf) {
        monsterIndex = 14
    } else if unit.Equals(units.Wraith) {
        monsterIndex = 16
    } else if unit.Equals(units.NightStalker) {
        monsterIndex = 17
    } else if unit.Equals(units.DemonLord) {
        monsterIndex = 19
    } else if unit.Equals(units.Unicorn) {
        monsterIndex = 21
    } else if unit.Equals(units.GuardianSpirit) {
        monsterIndex = 22
    } else if unit.Equals(units.Angel) {
        monsterIndex = 23
    } else if unit.Equals(units.ArchAngel) {
        monsterIndex = 24
    } else if unit.Equals(units.WarBear) {
        monsterIndex = 25
    } else if unit.Equals(units.Sprite) {
        monsterIndex = 26
    } else if unit.Equals(units.Cockatrice) {
        monsterIndex = 27
    } else if unit.Equals(units.Basilisk) {
        monsterIndex = 28
    } else if unit.Equals(units.GiantSpider) {
        monsterIndex = 29
    } else if unit.Equals(units.StoneGiant) {
        monsterIndex = 30
    } else if unit.Equals(units.Colossus) {
        monsterIndex = 31
    } else if unit.Equals(units.Gorgon) {
        monsterIndex = 32
    } else if unit.Equals(units.EarthElemental) {
        monsterIndex = 33
    } else if unit.Equals(units.Behemoth) {
        monsterIndex = 34
    } else if unit.Equals(units.GreatWyrm) {
        monsterIndex = 35
    } else if unit.Equals(units.StormGiant) {
        monsterIndex = 39
    } else if unit.Equals(units.Djinn) {
        monsterIndex = 41
    } else if unit.Equals(units.SkyDrake) {
        monsterIndex = 42
    } else if unit.Equals(units.Nagas) {
        monsterIndex = 43
    } else {
        log.Printf("Invalid summoning for unit %v", unit)
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

    // FIXME: some of the pixels still have the wrong color, like the outer edges of the summoning circle
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

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        return nil
    }

    orange := color.RGBA{R: 0xc7, G: 0x82, B: 0x1b, A: 0xff}

    yellowPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        util.Lighten(orange, 0),
        util.Lighten(orange, 15),
        util.Lighten(orange, 30),
        util.Lighten(orange, 50),
        util.Lighten(orange, 70),
        util.Lighten(orange, 90),
    }

    infoFontYellow := font.MakeOptimizedFontWithPalette(fonts[4], yellowPalette)
    summon.Font = infoFontYellow

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
    options.GeoM.Translate(30, 40)
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

    x, y := options.GeoM.Apply(float64(summon.Background.Bounds().Dx())/2, float64(summon.Background.Bounds().Dy()) - 18)
    summon.Font.PrintCenter(screen, x, y, 1, options.ColorScale, fmt.Sprintf("%v Summoned", summon.Unit.Name))
}
