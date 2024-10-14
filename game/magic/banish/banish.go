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

// the wizard with their hand raised, casting the spell
func getWizardAttackImageIndex(wizard data.WizardBase) int {
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

    badGuy1, _ := imageCache.GetImage("conquest.lbx", 15, 0)
    badGuy2, _ := imageCache.GetImage("conquest.lbx", 14, 0)

    attackImage, _ := imageCache.GetImage("conquest.lbx", getWizardAttackImageIndex(attackingWizard.Wizard.Base), 0)

    type Sprite struct {
        StartX, StartY float64
        DestX, DestY float64
        Steps int
        NumSteps int
        Image *ebiten.Image
    }

    animationSteps := 30

    badGuy1Sprite := Sprite{
        StartX: 320,
        StartY: 80,
        DestX: 250,
        DestY: 80,
        Steps: animationSteps,
        Image: badGuy1,
    }

    badGuy2Sprite := Sprite{
        StartX: 180,
        StartY: 200,
        DestX: 100,
        DestY: 130,
        Steps: animationSteps,
        Image: badGuy2,
    }

    wizardSprite := Sprite{
        StartX: 320,
        StartY: 200,
        DestX: 200,
        DestY: 60,
        Steps: animationSteps,
        Image: attackImage,
    }

    sprites := []*Sprite{&badGuy1Sprite, &wizardSprite, &badGuy2Sprite}

    spell1Images, _ := imageCache.GetImages("conquest.lbx", 16)
    spell2Images, _ := imageCache.GetImages("conquest.lbx", 17)
    var spellAnimation *util.Animation
    spellX := float64(0)
    spellY := float64(0)

    dissappear := false
    dissappearImages, _ := imageCache.GetImages("conquest.lbx", getWizardDissappearImageIndex(defeatedWizard.Wizard.Base))
    dissappearAnimation := util.MakeAnimation(dissappearImages, false)

    wizardGone := false

    draw := func (screen *ebiten.Image){
        var options ebiten.DrawImageOptions
        screen.DrawImage(background, &options)

        if !wizardGone {
            if dissappear {
                options.GeoM.Translate(67, 16)
                screen.DrawImage(dissappearAnimation.Frame(), &options)
            } else {
                options.GeoM.Translate(70, 75)
                screen.DrawImage(defeatedWizardImage, &options)
            }
        }

        if spellAnimation != nil {
            options.GeoM.Reset()
            options.GeoM.Translate(spellX, spellY)
            screen.DrawImage(spellAnimation.Frame(), &options)
        }

        for _, sprite := range sprites {
            options.GeoM.Reset()

            var x float64
            var y float64
            if sprite.NumSteps > sprite.Steps {
                x = sprite.DestX
                y = sprite.DestY
            } else {
                x = sprite.StartX + (sprite.DestX - sprite.StartX) * float64(sprite.NumSteps) / float64(sprite.Steps)
                y = sprite.StartY + (sprite.DestY - sprite.StartY) * float64(sprite.NumSteps) / float64(sprite.Steps)
            }

            options.GeoM.Translate(x, y)
            screen.DrawImage(sprite.Image, &options)
        }
    }

    logic := func (yield coroutine.YieldFunc) error {
        animationSpeed := 6

        for i := 0; i < animationSteps; i++ {
            for _, sprite := range sprites {
                sprite.NumSteps += 1
            }

            yield()
        }

        spellAnimation = util.MakeAnimation(spell1Images, false)
        spellX = 175
        spellY = 64

        for i := 0; i < 2000; i++ {
            if i % animationSpeed == 0 {
                spellAnimation.Next()
            }
            yield()

            if spellAnimation.Done() {
                break
            }
        }

        spellAnimation = util.MakeAnimation(spell2Images, true)
        spellX = 0
        spellY = 0

        for i := 0; i < 45; i++ {
            if i % animationSpeed == 0 {
                spellAnimation.Next()
            }
            yield()
        }

        spellAnimation = util.MakeAnimation(spell1Images[len(spell1Images)-2:len(spell1Images)], true)
        spellX = 175
        spellY = 64

        dissappear = true
        for i := 0; i < 2000; i++ {
            if i % animationSpeed == 0 {
                dissappearAnimation.Next()
                spellAnimation.Next()
            }
            yield()

            if dissappearAnimation.Done() {
                break
            }
        }

        wizardGone = true

        spellAnimation = util.MakeReverseAnimation(spell1Images, false)
        spellX = 175
        spellY = 64

        for i := 0; i < 200; i++ {
            if i % animationSpeed == 0 {
                spellAnimation.Next()
            }
            yield()

            if spellAnimation.Done() {
                break
            }
        }

        spellAnimation = nil

        for i := 0; i < 60; i++ {
            yield()
        }

        return nil
    }

    return logic, draw
}
