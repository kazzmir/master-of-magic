package game

import (
    "github.com/hajimehoshi/ebiten/v2"
)

type Map struct {
}

func MakeMap() *Map {
    return &Map{
    }
}

func (mapObject *Map) Draw(screen *ebiten.Image){
}
