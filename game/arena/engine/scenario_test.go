package engine

import (
	"reflect"
	"testing"

	"github.com/kazzmir/master-of-magic/game/arena/player"
	"github.com/kazzmir/master-of-magic/game/magic/data"
)

func TestTestScenarios(t *testing.T) {
	counts := []int{2, 5, 10, 18, 30}
	difficulties := []Difficulty{DifficultyEasy, DifficultyNormal, DifficultyHard, DifficultyImpossible, DifficultyImpossible}
	var previous *player.Player
	for scenario := 1; scenario <= TestScenarioCount; scenario++ {
		p, difficulty, err := makeTestPlayer(scenario)
		if err != nil {
			t.Fatal(err)
		}
		if len(p.Units) != counts[scenario-1] || difficulty != difficulties[scenario-1] {
			t.Fatalf("scenario %d: got %d units, %v difficulty", scenario, len(p.Units), difficulty)
		}
		if p.Mana != p.OriginalMana || p.Mana == 0 || p.Wizard.Banner != data.BannerGreen {
			t.Fatalf("scenario %d: invalid starting player", scenario)
		}
		if previous != nil && (p.Money <= previous.Money || p.Mana <= previous.Mana || p.Level <= previous.Level) {
			t.Fatalf("scenario %d did not increase resources", scenario)
		}
		for _, unit := range p.Units {
			if unit.GetHealth() <= 0 || len(unit.GetEnchantments()) != scenario {
				t.Fatalf("scenario %d: unit must be alive with %d enchantments", scenario, scenario)
			}
		}
		if p.Units[0].GetDamage() <= 0 {
			t.Fatalf("scenario %d: expected a damaged unit for testing healing", scenario)
		}
		again, _, err := makeTestPlayer(scenario)
		if err != nil || !reflect.DeepEqual(p, again) {
			t.Fatalf("scenario %d is not deterministic: %v", scenario, err)
		}
		// Fixtures must not share mutable player or unit state.
		again.Units[0].AdjustHealth(-1)
		if p.Units[0].GetHealth() == again.Units[0].GetHealth() {
			t.Fatalf("scenario %d shares unit state", scenario)
		}
		previous = p
	}
}

func TestInvalidTestScenario(t *testing.T) {
	for _, scenario := range []int{-1, 0, TestScenarioCount + 1, 1000} {
		if _, _, err := makeTestPlayer(scenario); err == nil {
			t.Errorf("scenario %d should be rejected", scenario)
		}
	}
}
