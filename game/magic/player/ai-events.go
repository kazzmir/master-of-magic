package player

import (
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
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

func (ai *DefaultAIEvents) DidLoseUnit(unit units.StackUnit) {
}

func (ai *DefaultAIEvents) DidCreateUnit(unit units.StackUnit) {
}

func (*DefaultAIEvents) DidLearnSpell(spell spellbook.Spell) {
}

func (*DefaultAIEvents) DidGainHero(hero *herolib.Hero) {
}

func (*DefaultAIEvents) DidLoseHero(hero *herolib.Hero) {
}

var _ AIEvents = (*DefaultAIEvents)(nil)
