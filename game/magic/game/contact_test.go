package game

import (
    "testing"

    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
)

func makeContactPlayers() (*playerlib.Player, *playerlib.Player) {
    names := make(map[herolib.HeroType]string)
    human := playerlib.MakePlayer(setup.WizardCustom{Name: "Merlin", Banner: data.BannerRed}, true, 8, 8, names, nil)
    enemy := playerlib.MakePlayer(setup.WizardCustom{Name: "Jafar", Banner: data.BannerGreen}, false, 8, 8, names, nil)
    return human, enemy
}

func TestMakeWizardContact(test *testing.T) {
    model := &GameModel{}
    human, enemy := makeContactPlayers()

    event := model.MakeWizardContact(human, enemy)
    if event == nil {
        test.Fatalf("first meeting with a wizard should produce a diplomacy event")
    }
    if event.Player != human || event.Enemy != enemy {
        test.Errorf("diplomacy event should introduce the human to the enemy")
    }
    if !human.IsAwareOf(enemy) || !enemy.IsAwareOf(human) {
        test.Errorf("contact should make both wizards aware of each other")
    }

    if model.MakeWizardContact(human, enemy) != nil {
        test.Errorf("a second meeting should not produce another intro")
    }
}

func TestMakeWizardContactIgnoresRaiders(test *testing.T) {
    model := &GameModel{}
    human, _ := makeContactPlayers()
    raider := playerlib.MakePlayer(setup.WizardCustom{Banner: data.BannerBrown}, false, 8, 8, make(map[herolib.HeroType]string), nil)

    if model.MakeWizardContact(human, raider) != nil {
        test.Errorf("neutral raiders should not create wizard contact")
    }
    if human.IsAwareOf(raider) {
        test.Errorf("human should not become aware of raiders")
    }
}

func TestDiscoverVisibleWizards(test *testing.T) {
    human, enemy := makeContactPlayers()
    city := &citylib.City{Name: "Enemy", X: 3, Y: 3, Plane: data.PlaneArcanus}
    enemy.AddCity(city)

    model := &GameModel{
        Players: []*playerlib.Player{human, enemy},
    }

    if meetings := model.DiscoverVisibleWizards(); len(meetings) != 0 {
        test.Errorf("should not meet a wizard hidden in the fog, got %v", len(meetings))
    }

    human.LiftFogSquare(3, 3, 0, data.PlaneArcanus)
    meetings := model.DiscoverVisibleWizards()
    if len(meetings) != 1 {
        test.Fatalf("seeing an enemy city should create one intro, got %v", len(meetings))
    }
    if meetings[0].Player != human || meetings[0].Enemy != enemy {
        test.Errorf("intro should be the human meeting the city owner")
    }

    if meetings := model.DiscoverVisibleWizards(); len(meetings) != 0 {
        test.Errorf("already-known wizards should not be introduced again")
    }
}

func TestDiscoverVisibleWizardsFromEnemySight(test *testing.T) {
    human, enemy := makeContactPlayers()
    city := &citylib.City{Name: "Capital", X: 5, Y: 5, Plane: data.PlaneArcanus}
    human.AddCity(city)

    model := &GameModel{
        Players: []*playerlib.Player{human, enemy},
    }

    enemy.LiftFogSquare(5, 5, 0, data.PlaneArcanus)
    meetings := model.DiscoverVisibleWizards()
    if len(meetings) != 1 {
        test.Fatalf("an enemy seeing the human city should still introduce the human, got %v", len(meetings))
    }
    if meetings[0].Player != human || meetings[0].Enemy != enemy {
        test.Errorf("intro should still address the human player")
    }
}
