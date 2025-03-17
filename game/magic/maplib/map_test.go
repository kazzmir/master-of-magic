package maplib

import (
    "testing"
    "image"

    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/data"
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
        distance := xmap.XDistance(test.x1, test.x2)
        if distance != test.distance {
            t.Errorf("Distance from %d to %d is %d, expected %d", test.x1, test.x2, distance, test.distance)
        }
    }
}

type FakeWizard struct {
}

func (fake *FakeWizard) GetBanner() data.BannerType {
    return data.BannerGreen
}

func TestInfluence(test *testing.T) {
    xmap := Map{
        ExtraMap: make(map[image.Point]map[ExtraKind]ExtraTile),
    }

    xmap.ExtraMap[image.Pt(2, 2)] = map[ExtraKind]ExtraTile{
        ExtraKindMagicNode: &ExtraMagicNode{
            Kind: MagicNodeNature,
            MeldingWizard: &FakeWizard{},
            Zone: []image.Point{image.Pt(0, 0), image.Pt(1, 0), image.Pt(3, 3)},
        },
    }

    node := xmap.GetMagicInfluence(5, 5)
    if node == nil || node.Kind.MagicType() != data.NatureMagic {
        test.Errorf("Expected NatureMagic at 5,5")
    }
}
