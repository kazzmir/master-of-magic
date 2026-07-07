package data

import "testing"

func TestAbilitySupportsValue(t *testing.T) {
    testCases := []struct {
        ability  AbilityType
        supports bool
    }{
        {ability: AbilityFireBreath, supports: true},
        {ability: AbilityLeadership, supports: true},
        {ability: AbilityAgility, supports: true},
        {ability: AbilityTransport, supports: true},
        {ability: AbilityArmorPiercing, supports: false},
        {ability: AbilityCauseFear, supports: false},
        {ability: AbilityLucky, supports: false},
    }

    for _, testCase := range testCases {
        if got := testCase.ability.SupportsValue(); got != testCase.supports {
            t.Fatalf("%v SupportsValue() = %v, want %v", testCase.ability, got, testCase.supports)
        }
    }
}
