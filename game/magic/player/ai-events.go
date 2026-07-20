package player

import (
    "github.com/kazzmir/master-of-magic/game/magic/units"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
)

type DefaultAIEvents struct {}

func (ai *DefaultAIEvents) DidBanish(self *Player, player *Player) {
}

func (ai *DefaultAIEvents) DidDefeat(self *Player, player *Player) {
}

func (ai *DefaultAIEvents) DidSummonUnit(self *Player, unit *units.OverworldUnit) {
}

func (ai *DefaultAIEvents) DidConquerCity(city *citylib.City, raze bool) {
}

func (ai *DefaultAIEvents) DidLoseCity(city *citylib.City) {
}

var _ AIEvents = (*DefaultAIEvents)(nil)
