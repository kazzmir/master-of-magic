package city

import (
    "testing"
    "image"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/lib/fraction"
)

type Catchment struct {
}

func (catchment *Catchment) GetCatchmentArea(x int, y int) map[image.Point]maplib.FullTile {
    out := make(map[image.Point]maplib.FullTile)
    for x := -2; x <= 2; x++ {
        for y := -2; y <= 2; y++ {
            out[image.Point{x, y}] = maplib.FullTile{
                Tile: terrain.TileGrasslands1,
            }
        }
    }

    return out
}

func TestBasicCity(test *testing.T){
    city := MakeCity("Test City", 10, 10, data.RaceHighMen, data.BannerBlue, fraction.Make(3, 2), nil, &Catchment{})
    city.Population = 6000
    city.Farmers = 6
    city.Workers = 0
    city.ResetCitizens(nil)
    if city.Name != "Test City" {
        test.Error("City name is not correct")
    }
    if city.X != 10 {
        test.Error("City X is not correct")
    }
    if city.Y != 10 {
        test.Error("City Y is not correct")
    }

    if city.Citizens() != 6 {
        test.Errorf("Citizens should have been 6 but was %v", city.Citizens())
    }

    if city.ComputeSubsistenceFarmers() != 3 {
        test.Errorf("Subsistence farmers should have been 3 but was %v", city.ComputeSubsistenceFarmers())
    }

    if city.ComputeUnrest(nil) != 1 {
        test.Errorf("Unrest should have been 1 but was %v", city.ComputeUnrest(nil))
    }

    if city.Rebels != 1 {
        test.Errorf("Rebels should have been 1 but was %v", city.Rebels)
    }

    city.UpdateTaxRate(fraction.Make(3, 1), nil)

    if city.ComputeUnrest(nil) != 4 {
        test.Errorf("Unrest should have been 4 but was %v", city.ComputeUnrest(nil))
    }

    if city.Rebels != 3 {
        test.Errorf("Rebels should have been 3 but was %v", city.Rebels)
    }

    if city.Workers != 0 {
        test.Errorf("Workers should have been 0 but was %v", city.Workers)
    }

    if city.Farmers != 3 {
        test.Errorf("Farmers should have been 3 but was %v", city.Farmers)
    }
}
