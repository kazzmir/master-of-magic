package player

import (
    "github.com/kazzmir/master-of-magic/game/magic/units"
)

type DefaultAIEvents struct {}

func (ai *DefaultAIEvents) DidBanish(self *Player, player *Player) {
}

func (ai *DefaultAIEvents) DidDefeat(self *Player, player *Player) {
}

func (ai *DefaultAIEvents) DidSummonUnit(self *Player, unit *units.OverworldUnit) {
}

var _ AIEvents = (*DefaultAIEvents)(nil)
