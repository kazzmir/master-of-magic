package game

import (
    "testing"
    "strings"
)

func TestCityName(test *testing.T){
    names := []string{"A", "B", "C", "A 1", "B 1"}
    choices := []string{"A", "B", "C"}

    chosen := chooseCityName(names, choices)
    if chosen != "C 1" {
        test.Errorf("Expected C 1 but got %v", chosen)
    }

    names = []string{"A", "B", "C", "A 1", "B 1", "C 1"}
    choices = []string{"A", "B", "C"}
    chosen = chooseCityName(names, choices)
    if !strings.Contains(chosen, "2") {
        test.Errorf("Expected 2 but got %v", chosen)
    }
}
