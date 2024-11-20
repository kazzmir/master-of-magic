package diplomacy

import (
    "image"
    "log"
    "image/color"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/hajimehoshi/ebiten/v2"
)

/* player is talking to enemy
 */
func ShowDiplomacyScreen(cache *lbx.LbxCache, player *playerlib.Player, enemy *playerlib.Player) (func (coroutine.YieldFunc), func (*ebiten.Image)) {

    imageCache := util.MakeImageCache(cache)

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil, nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil, nil
    }

    solid := util.Lighten(color.RGBA{R: 0xca, G: 0x8a, B: 0x4a, A: 0xff}, -10)

    yellowGradient := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        solid,
        util.Lighten(solid, 10),
        util.Lighten(solid, 20),
        util.Lighten(solid, 40),
        util.Lighten(solid, 30),
        util.Lighten(solid, 30),
        solid,
        solid,
        solid,
    }

    solidOrange := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        solid, solid, solid,
        solid, solid, solid,
        solid, solid, solid,
    }

    bigFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowGradient)
    normalFont := font.MakeOptimizedFontWithPalette(fonts[1], solidOrange)

    quit := false

    animationIndex := 0
    switch enemy.Wizard.Base {
        case data.WizardMerlin: animationIndex = 0
        case data.WizardRaven: animationIndex = 1
        case data.WizardSharee: animationIndex = 2
        case data.WizardLoPan: animationIndex = 3
        case data.WizardJafar: animationIndex = 4
        case data.WizardOberic: animationIndex = 5
        case data.WizardRjak: animationIndex = 6
        case data.WizardSssra: animationIndex = 7
        case data.WizardTauron: animationIndex = 8
        case data.WizardFreya: animationIndex = 9
        case data.WizardHorus: animationIndex = 10
        case data.WizardAriel: animationIndex = 11
        case data.WizardTlaloc: animationIndex = 12
        case data.WizardKali: animationIndex = 13
    }

    // the fade in animation
    images, _ := imageCache.GetImages("diplomac.lbx", 38 + animationIndex)
    wizardAnimation := util.MakeAnimation(images, false)

    diplomacLbx, _ := cache.GetLbxFile("diplomac.lbx")
    // the tauron fade in, but any mask will work
    maskSprites, _ := diplomacLbx.ReadImages(46)
    mask := maskSprites[0]

    var makeCutoutMask util.ImageTransformFunc = func (img *image.Paletted) image.Image {
        properImage := img.SubImage(mask.Bounds()).(*image.Paletted)
        imageOut := image.NewPaletted(properImage.Bounds(), properImage.Palette)

        for x := properImage.Bounds().Min.X; x < properImage.Bounds().Max.X; x++ {
            for y := properImage.Bounds().Min.Y; y < properImage.Bounds().Max.Y; y++ {
                maskColor := mask.At(x, y)
                _, _, _, a := maskColor.RGBA()
                if a > 0 {
                    imageOut.Set(x, y, properImage.At(x, y))
                } else {
                    imageOut.SetColorIndex(x, y, 0)
                }
            }
        }

        return imageOut
    }

    var counter uint64
    logic := func (yield coroutine.YieldFunc) {
        animating := true

        for !quit {
            counter += 1
            if counter % 7 == 0 {
                if animating && wizardAnimation.Done() {
                    // 0 = happy, 1 = angry, 2 = neutral
                    moodIndex := 0
                    mood, _ := imageCache.GetImageTransform("moodwiz.lbx", animationIndex, moodIndex, "cutout", makeCutoutMask)
                    wizardAnimation = util.MakeAnimation([]*ebiten.Image{mood}, false)
                    animating = false
                }

                wizardAnimation.Next()
            }

            // if rand.N(100) == 0 {
            if counter == 80 && rand.N(1) == 0 {
                // talking
                images, _ := imageCache.GetImagesTransform("diplomac.lbx", 24 + animationIndex, "cutout", makeCutoutMask)
                wizardAnimation = util.MakeAnimation(images, false)
                animating = true
            }

            yield()
        }
    }

    draw := func (screen *ebiten.Image) {
        background, _ := imageCache.GetImage("diplomac.lbx", 0, 0)
        var options ebiten.DrawImageOptions
        screen.DrawImage(background, &options)

        // red left eye
        leftEye, _ := imageCache.GetImage("diplomac.lbx", 2, 0)
        // FIXME: what do the other eye colors mean? is it related to the diplomatic relationship level between the wizards?
        // red right eye
        rightEye, _ := imageCache.GetImage("diplomac.lbx", 13, 0)

        options.GeoM.Translate(63, 58)
        screen.DrawImage(leftEye, &options)

        options.GeoM.Translate(170, 0)
        screen.DrawImage(rightEye, &options)

        options.GeoM.Reset()
        options.GeoM.Translate(106, 11)
        screen.DrawImage(wizardAnimation.Frame(), &options)

        bigFont.Print(screen, 60, 140, 1, ebiten.ColorScale{}, "How may I serve you:")
        normalFont.Print(screen, 70, 160, 1, ebiten.ColorScale{}, "Good bye")
    }

    return logic, draw
}
