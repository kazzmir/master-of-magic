package maplib

import (
    "testing"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
)

func TestMap(t *testing.T) {
    rawMap := terrain.MakeMap(10, 10)
    xmap := Map{
        Map: rawMap,
    }

    type Test struct {
        x1, x2, distance int
    }

    tests := []Test{
        {3,2,-1},
        {3,0,-3},
        {2,3,1},
        {3,2,-1},
        {9,0,1},
        {0,9,-1},
    }

    for _, test := range tests {
        distance := xmap.Distance(test.x1, test.x2)
        if distance != test.distance {
            t.Errorf("Distance from %d to %d is %d, expected %d", test.x1, test.x2, distance, test.distance)
        }
    }
}
