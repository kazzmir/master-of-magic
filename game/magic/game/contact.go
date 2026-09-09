package game

import (
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
)

// make two wizards mutually aware. returns a diplomacy event when the human
// is meeting the other wizard for the first time. raiders and defeated
// wizards are ignored.
func (model *GameModel) MakeWizardContact(a *playerlib.Player, b *playerlib.Player) *GameEventDiplomacy {
    if a == nil || b == nil || a == b {
        return nil
    }
    if a.IsNeutral() || b.IsNeutral() || a.Defeated || b.Defeated {
        return nil
    }

    knewA := a.IsAwareOf(b)
    knewB := b.IsAwareOf(a)
    if knewA && knewB {
        return nil
    }

    a.AwarePlayer(b)
    b.AwarePlayer(a)

    if a.IsHuman() && !knewA {
        return &GameEventDiplomacy{Player: a, Enemy: b}
    }
    if b.IsHuman() && !knewB {
        return &GameEventDiplomacy{Player: b, Enemy: a}
    }

    return nil
}

func (model *GameModel) DiscoverVisibleWizards() []*GameEventDiplomacy {
    var meetings []*GameEventDiplomacy

    if model == nil {
        return meetings
    }

    for _, observer := range model.Players {
        if observer.IsNeutral() || observer.Defeated {
            continue
        }

        for _, other := range model.Players {
            if other == observer || other.IsNeutral() || other.Defeated {
                continue
            }
            if observer.IsAwareOf(other) {
                continue
            }
            if observer.CanSeePlayer(other) {
                if event := model.MakeWizardContact(observer, other); event != nil {
                    meetings = append(meetings, event)
                }
            }
        }
    }

    return meetings
}

func (game *Game) playFirstMeeting(yield coroutine.YieldFunc, event *GameEventDiplomacy) {
    if event == nil {
        return
    }

    if yield != nil {
        game.doDiplomacy(yield, event.Player, event.Enemy)
        return
    }

    if game.Events != nil {
        select {
            case game.Events <- event:
            default:
        }
    }
}

func (game *Game) meetWizards(yield coroutine.YieldFunc, a *playerlib.Player, b *playerlib.Player) {
    if game.Model == nil {
        return
    }
    game.playFirstMeeting(yield, game.Model.MakeWizardContact(a, b))
}

func (game *Game) discoverWizards(yield coroutine.YieldFunc) bool {
    if game.Model == nil {
        return false
    }

    meetings := game.Model.DiscoverVisibleWizards()
    for _, meeting := range meetings {
        game.playFirstMeeting(yield, meeting)
    }
    return len(meetings) > 0
}
