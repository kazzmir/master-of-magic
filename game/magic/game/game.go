package game

import (
    "fmt"

    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
    active bool

    MainHud *ebiten.Image

}

func (game *Game) Load(cache *lbx.LbxCache) error {
    mainLbx, err := cache.GetLbxFile("MAIN.LBX")
    if err != nil {
        return fmt.Errorf("Unable to load MAIN.LBX: %v", err)
    }

    var outError error

    loadImage := func(index int, subIndex int) *ebiten.Image {
        if outError != nil {
            return nil
        }

        sprites, err := mainLbx.ReadImages(index)
        if err != nil {
            outError = fmt.Errorf("Unable to read background image from NEWGAME.LBX: %v", err)
            return nil
        }

        if len(sprites) <= subIndex {
            outError = fmt.Errorf("Unable to read background image from NEWGAME.LBX: index %d out of range", subIndex)
            return nil
        }

        return ebiten.NewImageFromImage(sprites[subIndex])
    }

    game.MainHud = loadImage(0, 0)

    return outError
}

func MakeGame(wizard setup.WizardCustom) *Game {
    game := &Game{}
    return game
}

func (game *Game) IsActive() bool {
    return game.active
}

func (game *Game) Activate() {
    game.active = true
}

func (game *Game) Update(){
}

func (game *Game) Draw(screen *ebiten.Image){
    var options ebiten.DrawImageOptions
    screen.DrawImage(game.MainHud, &options)
}
