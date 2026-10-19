package engine

import (
	"fmt"
	"testing"

	"github.com/kazzmir/master-of-magic/game/arena/player"
	"github.com/kazzmir/master-of-magic/game/magic/data"
)

func TestAIUnitPurchasesStayWithinBudget(t *testing.T) {
	for _, budget := range []uint64{0, 1, 50, 80, 150, 1000, 10218347, ^uint64(0)} {
		t.Run(fmt.Sprint(budget), func(t *testing.T) {
			for range 20 {
				enemy := player.MakeAIPlayer(data.BannerRed)
				enemy.Level = 15
				for _, magic := range []data.MagicType{data.LifeMagic, data.NatureMagic, data.SorceryMagic, data.ChaosMagic, data.DeathMagic} {
					enemy.GetWizard().AddMagicLevel(magic, 5)
				}
				spent := buyAIUnit(enemy, budget)
				if spent > budget {
					t.Fatalf("spent %d exceeds budget %d", spent, budget)
				}
				if spent == 0 {
					if len(enemy.Units) != 0 {
						t.Fatal("zero spending added a unit")
					}
					if len(getValidChoices(budget)) > 0 {
						t.Fatal("failed to purchase an affordable unit")
					}
					continue
				}
				if len(enemy.Units) != 1 || spent != getCloneCost(enemy.Units[0]) {
					t.Fatal("spending does not match the unit and its upgrades")
				}
				if budget == 10218347 && budget-spent <= 10000000 {
					t.Fatal("regression fixture must retain a balance above the old overflow threshold")
				}
			}
		})
	}
}
