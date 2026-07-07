package player

type DefaultAIEvents struct {}

func (ai *DefaultAIEvents) DidBanish(self *Player, player *Player) {
}

func (ai *DefaultAIEvents) DidDefeat(self *Player, player *Player) {
}

var _ AIEvents = (*DefaultAIEvents)(nil)
