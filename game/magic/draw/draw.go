package draw

import (
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/hajimehoshi/ebiten/v2"
)

func DrawBooks(screen *ebiten.Image, options ebiten.DrawImageOptions, imageCache *util.ImageCache, books []data.WizardBook, random *rand.Rand){
    index := 0

    var lifeBooks [3]*ebiten.Image
    var sorceryBooks [3]*ebiten.Image
    var natureBooks [3]*ebiten.Image
    var deathBooks [3]*ebiten.Image
    var chaosBooks [3]*ebiten.Image

    loadImage := func (index int) *ebiten.Image {
        img, _ := imageCache.GetImage("newgame.lbx", index, 0)
        return img
    }

    for i := 0; i < 3; i++ {
        lifeBooks[i] = loadImage(24 + i)
        sorceryBooks[i] = loadImage(27 + i)
        natureBooks[i] = loadImage(30 + i)
        deathBooks[i] = loadImage(33 + i)
        chaosBooks[i] = loadImage(36 + i)
    }

    for _, book := range books {

        for i := 0; i < book.Count; i++ {

            element := random.IntN(3)

            var img *ebiten.Image
            switch book.Magic {
                case data.LifeMagic: img = lifeBooks[element]
                case data.SorceryMagic: img = sorceryBooks[element]
                case data.NatureMagic: img = natureBooks[element]
                case data.DeathMagic: img = deathBooks[element]
                case data.ChaosMagic: img = chaosBooks[element]
            }

            // var options ebiten.DrawImageOptions
            // options.GeoM.Translate(x + float64(offsetX), y)
            screen.DrawImage(img, &options)
            options.GeoM.Translate(float64(img.Bounds().Dx() - 1), 0)
            // offsetX += img.Bounds().Dx() - 1
            index += 1
        }
    }
}


