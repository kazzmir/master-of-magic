package game

import (
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (game *Game) wizlabAnimation(yield coroutine.YieldFunc, wizard setup.WizardCustom){
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    var fade util.AlphaFadeFunc

    wizardIndex := -1
    switch wizard.Base {
        case data.WizardMerlin: wizardIndex = 0
        case data.WizardRaven: wizardIndex = 1
        case data.WizardSharee: wizardIndex = 2
        case data.WizardLoPan: wizardIndex = 3
        case data.WizardJafar: wizardIndex = 4
        case data.WizardOberic: wizardIndex = 5
        case data.WizardRjak: wizardIndex = 6
        case data.WizardSssra: wizardIndex = 7
        case data.WizardTauron: wizardIndex = 8
        case data.WizardFreya: wizardIndex = 9
        case data.WizardHorus: wizardIndex = 10
        case data.WizardAriel: wizardIndex = 11
        case data.WizardTlaloc: wizardIndex = 12
        case data.WizardKali: wizardIndex = 13
    }

    animalIndex := 14

    // FIXME: base animal index on magic books?
    switch wizard.MostBooks() {
        case data.NatureMagic: animalIndex = 14
        case data.SorceryMagic: animalIndex = 15
        case data.ChaosMagic: animalIndex = 16
        case data.LifeMagic: animalIndex = 17
        case data.DeathMagic: animalIndex = 18
    }

    sparkleImages, _ := game.ImageCache.GetImages("wizlab.lbx", 21)

    sparkles := util.MakeAnimation(sparkleImages, true)

    game.Drawer = func(screen *ebiten.Image, game *Game){
        background, _ := game.ImageCache.GetImage("wizlab.lbx", 19, 0)
        var options ebiten.DrawImageOptions
        options.ColorScale.ScaleAlpha(fade())

        screen.DrawImage(background, &options)

        wizardPic, _ := game.ImageCache.GetImage("wizlab.lbx", wizardIndex, 0)
        options.GeoM.Translate(70, 74)
        screen.DrawImage(wizardPic, &options)

        options.GeoM.Reset()
        options.GeoM.Translate(132, -5)
        screen.DrawImage(sparkles.Frame(), &options)

        pulpit, _ := game.ImageCache.GetImage("wizlab.lbx", 20, 0)
        options.GeoM.Reset()
        options.GeoM.Translate(150, 130)
        screen.DrawImage(pulpit, &options)

        options.GeoM.Reset()
        options.GeoM.Translate(190, 157)
        animalPic, _ := game.ImageCache.GetImage("wizlab.lbx", animalIndex, 0)
        screen.DrawImage(animalPic, &options)
    }

    counter := uint64(0)

    fade = util.MakeFadeIn(7, &counter)

    yield()

    for counter = 0; counter < 10; counter++ {
        if counter % 5 == 0 {
            sparkles.Next()
        }

        if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
            break
        }

        yield()
    }

    counter = 0
    fade = util.MakeFadeOut(7, &counter)

    for counter = 0; counter < 7; counter++ {
        if counter % 5 == 0 {
            sparkles.Next()
        }
        yield()
    }
}

func (game *Game) doLearnSpell(yield coroutine.YieldFunc, player *playerlib.Player, spell spellbook.Spell){
    game.wizlabAnimation(yield, player.Wizard)
}