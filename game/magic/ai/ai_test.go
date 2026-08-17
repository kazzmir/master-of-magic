package ai

import (
    "testing"

    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

func TestUnitAttackPowerUsesToHit(test *testing.T) {
    swordsmen := units.MakeOverworldUnit(units.HighMenSwordsmen, 0, 0, data.PlaneArcanus)
    if unitAttackPower(swordsmen) <= 0 {
        test.Errorf("swordsmen should have positive attack power after to-hit scaling")
    }

    settler := units.MakeOverworldUnit(units.HighMenSettlers, 0, 0, data.PlaneArcanus)
    if unitAttackPower(settler) != 0 {
        test.Errorf("settlers should still count as 0 attack power, got %v", unitAttackPower(settler))
    }
}

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

    if self.Gold != 700 {
        test.Errorf("expected 700 gold after purchase, got %d", self.Gold)
    }
}
