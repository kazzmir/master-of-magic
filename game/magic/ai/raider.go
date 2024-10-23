package ai

import (
    "github.com/kazzmir/master-of-magic/game/magic/player"
)

type RaiderAI struct {
}

func MakeRaiderAI() *RaiderAI {
    return &RaiderAI{}
}

func (raider *RaiderAI) Update(player *player.Player) {
}
