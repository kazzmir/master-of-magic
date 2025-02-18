package player

import (
    "testing"

    "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
)

func TestHeroNames(test *testing.T) {
    names := make(map[hero.HeroType]string)
    names[hero.HeroFang] = "goofy"

    player := MakePlayer(setup.WizardCustom{}, true, 5, 5, names)
    fangName := player.HeroPool[hero.HeroFang].GetName()

    if fangName != "goofy" {
        test.Errorf("Expected hero name to be 'goofy', got '%s'", fangName)
    }
}
