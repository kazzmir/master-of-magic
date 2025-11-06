package ai

import (
    "testing"

    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
)

func TestBuyItem(test *testing.T) {
    enemy := MakeEnemyAI()

    self := playerlib.MakePlayer(setup.WizardCustom{}, false, 2, 2, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{})
    self.Gold = 1000

    weapon := artifact.Artifact{
    }

    if !enemy.HandleMerchantItem(self, &weapon, 300) {
        test.Errorf("expected to buy item, but did not")
    }

    artifacts := 0
    for _, item := range self.VaultEquipment {
        if item != nil {
            artifacts += 1
        }
    }

    if artifacts != 1 {
        test.Errorf("expected 1 artifact in vault, got %d", artifacts)
    }
}
