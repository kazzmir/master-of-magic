package game

import (
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
    active bool
}

func MakeGame(wizard setup.WizardCustom) *Game {
    return &Game{}
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
}
