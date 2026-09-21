package game

import (
    "testing"

    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/lib/set"
)

func TestEarthquakeBuildingProtected(test *testing.T) {
    expect := func(name string, building buildinglib.Building, intact *set.Set[buildinglib.Building], want bool) {
        got := earthquakeBuildingProtected(building, intact)
        if got != want {
            test.Errorf("%s: protected=%v, want %v", name, got, want)
        }
    }

    shrineTempleParthenon := set.NewSet(
        buildinglib.BuildingShrine,
        buildinglib.BuildingTemple,
        buildinglib.BuildingParthenon,
    )
    expect("shrine under parthenon", buildinglib.BuildingShrine, shrineTempleParthenon, true)
    expect("temple under parthenon", buildinglib.BuildingTemple, shrineTempleParthenon, true)
    expect("parthenon is top of line", buildinglib.BuildingParthenon, shrineTempleParthenon, false)

    shrineTemple := set.NewSet(buildinglib.BuildingShrine, buildinglib.BuildingTemple)
    expect("shrine under temple", buildinglib.BuildingShrine, shrineTemple, true)
    expect("temple is top of line", buildinglib.BuildingTemple, shrineTemple, false)

    shrineOnly := set.NewSet(buildinglib.BuildingShrine)
    expect("lone shrine", buildinglib.BuildingShrine, shrineOnly, false)

    granaryOnly := set.NewSet(buildinglib.BuildingGranary)
    expect("granary without farmers market", buildinglib.BuildingGranary, granaryOnly, false)

    marketplaceBank := set.NewSet(buildinglib.BuildingMarketplace, buildinglib.BuildingBank)
    expect("marketplace under bank", buildinglib.BuildingMarketplace, marketplaceBank, true)
    expect("bank is top of line", buildinglib.BuildingBank, marketplaceBank, false)

    // fortress is always protected no matter what the set of intact buildings is
    expect("fortress is protected", buildinglib.BuildingFortress, shrineTempleParthenon, true)
}

func TestEarthquakeSparesLowerTierBuildings(test *testing.T) {
    model := &GameModel{}
    player := playerlib.MakePlayer(
        setup.WizardCustom{Banner: data.BannerRed, Race: data.RaceHighMen},
        true, 1, 1, make(map[herolib.HeroType]string), nil,
    )

    parthenonDestroyed := 0
    for i := 0; i < 200; i++ {
        city := &citylib.City{
            X: 1,
            Y: 1,
            Plane: data.PlaneArcanus,
            Buildings: set.NewSet(
                buildinglib.BuildingShrine,
                buildinglib.BuildingTemple,
                buildinglib.BuildingParthenon,
            ),
        }

        people, _, destroyed := model.DoEarthquake(city, player)
        if people != 0 {
            test.Fatalf("trial %d: earthquake killed %d citizens, want 0", i, people)
        }
        if !city.Buildings.Contains(buildinglib.BuildingShrine) {
            test.Fatalf("trial %d: shrine was destroyed while temple/parthenon still existed at quake start", i)
        }
        if !city.Buildings.Contains(buildinglib.BuildingTemple) {
            test.Fatalf("trial %d: temple was destroyed while parthenon still existed at quake start", i)
        }
        if !city.Buildings.Contains(buildinglib.BuildingParthenon) {
            parthenonDestroyed++
        }
        for _, building := range destroyed {
            if building == buildinglib.BuildingShrine || building == buildinglib.BuildingTemple {
                test.Fatalf("trial %d: destroyed list included protected building %v", i, building)
            }
        }
    }

    if parthenonDestroyed == 0 {
        test.Errorf("parthenon was never destroyed in 200 quakes; it should be a valid target")
    }
}
