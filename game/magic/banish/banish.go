package banish

import (
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
)

// the animation where the wizard dissappears into the air
func getWizardDissappearImageIndex(wizard data.WizardBase) int {
    switch wizard {
        case data.WizardMerlin: return 18
        case data.WizardRaven: return 19
        case data.WizardSharee: return 20
        case data.WizardLoPan: return 21
        case data.WizardJafar: return 22
        case data.WizardOberic: return 23
        case data.WizardRjak: return 24
        case data.WizardSssra: return 25
        case data.WizardTauron: return 26
        case data.WizardFreya: return 27
        case data.WizardHorus: return 28
        case data.WizardAriel: return 29
        case data.WizardTlaloc: return 30
        case data.WizardKali: return 31
    }

    return -1
}

// just standin' around
func getWizardStandingImageIndex(wizard data.WizardBase) int {
    switch wizard {
        case data.WizardMerlin: return 0
        case data.WizardRaven: return 1
        case data.WizardSharee: return 2
        case data.WizardLoPan: return 3
        case data.WizardJafar: return 4
        case data.WizardOberic: return 5
        case data.WizardRjak: return 6
        case data.WizardSssra: return 7
        case data.WizardTauron: return 8
        case data.WizardFreya: return 9
        case data.WizardHorus: return 10
        case data.WizardAriel: return 11
        case data.WizardTlaloc: return 12
        case data.WizardKali: return 13
    }

    return -1
}

func ShowBanishAnimation(cache *lbx.LbxCache, attackingWizard *playerlib.Player, defeatedWizard *playerlib.Player) (func (coroutine.YieldFunc) error, func (*ebiten.Image)) {

    imageCache := util.MakeImageCache(cache)

    background, _ := imageCache.GetImage("wizlab.lbx", 19, 0)

    defeatedWizardImage, _ := imageCache.GetImage("wizlab.lbx", getWizardStandingImageIndex(defeatedWizard.Wizard.Base), 0)

    draw := func (screen *ebiten.Image){
        var options ebiten.DrawImageOptions
        screen.DrawImage(background, &options)

        options.GeoM.Translate(70, 75)
        screen.DrawImage(defeatedWizardImage, &options)
    }

    logic := func (yield coroutine.YieldFunc) error {
        for i := 0; i < 200; i++ {
            yield()
        }

        return nil
    }

    return logic, draw
}
