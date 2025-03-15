package load

import (
	"fmt"
	"image"

	buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
	citylib "github.com/kazzmir/master-of-magic/game/magic/city"
	"github.com/kazzmir/master-of-magic/game/magic/data"
	gamelib "github.com/kazzmir/master-of-magic/game/magic/game"
	"github.com/kazzmir/master-of-magic/game/magic/maplib"
	playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
	"github.com/kazzmir/master-of-magic/game/magic/setup"
	"github.com/kazzmir/master-of-magic/game/magic/terrain"
	"github.com/kazzmir/master-of-magic/game/magic/units"
	"github.com/kazzmir/master-of-magic/lib/set"

	"github.com/hajimehoshi/ebiten/v2"
)

func (saveGame *SaveGame) ToMap(terrainData *terrain.TerrainData, plane data.Plane, cityProvider maplib.CityProvider) *maplib.Map {

    map_ := maplib.Map{
        Data: terrainData,
        Map: terrain.MakeMap(WorldHeight, WorldWidth),
        Plane: plane,
        TileCache: make(map[int]*ebiten.Image),
        ExtraMap: make(map[image.Point]map[maplib.ExtraKind]maplib.ExtraTile),
        CityProvider: cityProvider,
    }

    terrainSource := saveGame.ArcanusMap
    terrainOffset := 0
    terrainSpecials := saveGame.ArcanusTerrainSpecials
    if plane == data.PlaneMyrror {
        terrainSource = saveGame.MyrrorMap
        terrainOffset = terrain.MyrrorStart
        terrainSpecials = saveGame.MyrrorTerrainSpecials
    }

    for y := range(WorldHeight) {
        for x := range(WorldWidth) {
            point := image.Pt(x, y)

            map_.Map.Terrain[x][y] = int(terrainSource.Data[x][y]) + terrainOffset

            var bonus data.BonusType
            switch terrainSpecials[x][y] {
                case 1: bonus = data.BonusIronOre
                case 2: bonus = data.BonusCoal
                case 3: bonus = data.BonusSilverOre
                case 4: bonus = data.BonusGoldOre
                case 5: bonus = data.BonusGem
                case 6: bonus = data.BonusMithrilOre
                case 7: bonus = data.BonusAdamantiumOre
                case 8: bonus = data.BonusQuorkCrystal
                case 9: bonus = data.BonusCrysxCrystal
                case 64: bonus = data.BonusWildGame
                case 128: bonus = data.BonusNightshade
            }
            if bonus != data.BonusNone {
                map_.ExtraMap[point] = make(map[maplib.ExtraKind]maplib.ExtraTile)
                map_.ExtraMap[point][maplib.ExtraKindBonus] = &maplib.ExtraBonus{Bonus: bonus}
            }
        }
    }

    for _, tower := range saveGame.Towers {
        point := image.Pt(int(tower.X), int(tower.Y))
        if map_.ExtraMap[point] == nil {
            map_.ExtraMap[point] = make(map[maplib.ExtraKind]maplib.ExtraTile)
        }
        if tower.Owner == -1 {
            continue
        }
        map_.ExtraMap[point][maplib.ExtraKindOpenTower] = &maplib.ExtraOpenTower{}
    }

    for _, lair := range saveGame.Lairs {
        if lair.Intact == 0 {
            continue
        }

        if lair.Plane == 0 && plane != data.PlaneArcanus || lair.Plane != 0 && plane != data.PlaneMyrror {
            continue
        }

        var encounterType maplib.EncounterType
        switch lair.Kind {
            case 0: encounterType = maplib.EncounterTypePlaneTower
            case 1: encounterType = maplib.EncounterTypeChaosNode
            case 2: encounterType = maplib.EncounterTypeNatureNode
            case 3: encounterType = maplib.EncounterTypeSorceryNode
            case 4: encounterType = maplib.EncounterTypeCave
            case 5: encounterType = maplib.EncounterTypeDungeon
            case 6: encounterType = maplib.EncounterTypeAncientTemple
            case 7: encounterType = maplib.EncounterTypeAbandonedKeep
            case 8: encounterType = maplib.EncounterTypeLair
            case 9: encounterType = maplib.EncounterTypeRuins
            case 10: encounterType = maplib.EncounterTypeFallenTemple
        }

        point := image.Pt(int(lair.X), int(lair.Y))
        if map_.ExtraMap[point] == nil {
            map_.ExtraMap[point] = make(map[maplib.ExtraKind]maplib.ExtraTile)
        }
        map_.ExtraMap[point][maplib.ExtraKindEncounter] = &maplib.ExtraEncounter{
            Type: encounterType,
            // FIXME: set budget, units and explored by
            Budget: 0,
            Units: nil,
            ExploredBy: nil,
        }
    }

    return &map_
}

func (saveGame *SaveGame) ToSettings() setup.NewGameSettings {
    return setup.NewGameSettings{
        Difficulty:  data.DifficultySetting(saveGame.Difficulty),
        Opponents: int(saveGame.NumPlayers) - 1,
        LandSize: int(saveGame.LandSize),
        Magic: data.MagicSetting(saveGame.Magic),
    }
}

func (saveGame *SaveGame) ToWizard(index int) setup.WizardCustom {
    playerData := saveGame.PlayerData[index]

    retorts := []data.Retort{}
    if playerData.RetortAlchemy == 1 {
        retorts = append(retorts, data.RetortAlchemy)
    }
    if playerData.RetortWarlord == 1 {
        retorts = append(retorts, data.RetortWarlord)
    }
    if playerData.RetortChaosMastery == 1 {
        retorts = append(retorts, data.RetortChaosMastery)
    }
    if playerData.RetortNatureMastery == 1 {
        retorts = append(retorts, data.RetortNatureMastery)
    }
    if playerData.RetortSorceryMastery == 1 {
        retorts = append(retorts, data.RetortSorceryMastery)
    }
    if playerData.RetortInfernalPower == 1 {
        retorts = append(retorts, data.RetortInfernalPower)
    }
    if playerData.RetortDivinePower == 1 {
        retorts = append(retorts, data.RetortDivinePower)
    }
    if playerData.RetortSageMaster == 1 {
        retorts = append(retorts, data.RetortSageMaster)
    }
    if playerData.RetortChanneler == 1 {
        retorts = append(retorts, data.RetortChanneler)
    }
    if playerData.RetortMyrran == 1 {
        retorts = append(retorts, data.RetortMyrran)
    }
    if playerData.RetortArchmage == 1 {
        retorts = append(retorts, data.RetortArchmage)
    }
    if playerData.RetortNodeMastery == 1 {
        retorts = append(retorts, data.RetortNodeMastery)
    }
    if playerData.RetortManaFocusing == 1 {
        retorts = append(retorts, data.RetortManaFocusing)
    }
    if playerData.RetortFamous == 1 {
        retorts = append(retorts, data.RetortFamous)
    }
    if playerData.RetortRunemaster == 1 {
        retorts = append(retorts, data.RetortRunemaster)
    }
    if playerData.RetortConjurer == 1 {
        retorts = append(retorts, data.RetortConjurer)
    }
    if playerData.RetortCharismatic == 1 {
        retorts = append(retorts, data.RetortCharismatic)
    }
    if playerData.RetortArtificer == 1 {
        retorts = append(retorts, data.RetortArtificer)
    }

    books := []data.WizardBook{}
    if playerData.SpellRanks[0] != 0 {
        books = append(books, data.WizardBook{Magic: data.NatureMagic, Count: int(playerData.SpellRanks[0])})
    }
    if playerData.SpellRanks[1] != 0 {
        books = append(books, data.WizardBook{Magic: data.SorceryMagic, Count: int(playerData.SpellRanks[1])})
    }
    if playerData.SpellRanks[2] != 0 {
        books = append(books, data.WizardBook{Magic: data.ChaosMagic, Count: int(playerData.SpellRanks[2])})
    }
    if playerData.SpellRanks[3] != 0 {
        books = append(books, data.WizardBook{Magic: data.LifeMagic, Count: int(playerData.SpellRanks[3])})
    }
    if playerData.SpellRanks[4] != 0 {
        books = append(books, data.WizardBook{Magic: data.DeathMagic, Count: int(playerData.SpellRanks[4])})
    }

    var race data.Race
    switch playerData.CapitalRace {
        case 0: race = data.RaceBarbarian
        case 1: race = data.RaceBeastmen
        case 2: race = data.RaceDarkElf
        case 3: race = data.RaceDraconian
        case 4: race = data.RaceDwarf
        case 5: race = data.RaceGnoll
        case 6: race = data.RaceHalfling
        case 7: race = data.RaceHighElf
        case 8: race = data.RaceHighMen
        case 9: race = data.RaceKlackon
        case 10: race = data.RaceLizard
        case 11: race = data.RaceNomad
        case 12: race = data.RaceOrc
        case 13: race = data.RaceTroll
    }

    var banner data.BannerType
    switch playerData.BannerId {
        case 0: banner = data.BannerBlue
        case 1: banner = data.BannerGreen
        case 2: banner = data.BannerPurple
        case 3: banner = data.BannerRed
        case 4: banner = data.BannerYellow
    }

    var base data.WizardBase
    switch playerData.WizardId {
        case 0: base = data.WizardMerlin
        case 1: base = data.WizardRaven
        case 2: base = data.WizardSharee
        case 3: base = data.WizardLoPan
        case 4: base = data.WizardJafar
        case 5: base = data.WizardOberic
        case 6: base = data.WizardRjak
        case 7: base = data.WizardSssra
        case 8: base = data.WizardTauron
        case 9: base = data.WizardFreya
        case 10: base = data.WizardHorus
        case 11: base = data.WizardAriel
        case 12: base = data.WizardTlaloc
        case 13: base = data.WizardKali
    }

    return setup.WizardCustom{
        Name: string(playerData.WizardName),
        Portrait: int(playerData.WizardId),
        Base: base,
        Retorts: retorts,
        Books: books,
        Race: race,
        Banner: banner,
    }
}

func (saveGame *SaveGame) ToFogMap(plane data.Plane) data.FogMap {
    out := make([][]data.FogType, WorldWidth)
    for i := range WorldWidth {
        out[i] = make([]data.FogType, WorldHeight)
    }

    for x := range WorldWidth {
        for y := range WorldHeight {
            var value int8
            if plane == data.PlaneArcanus {
                value = saveGame.ArcanusExplored[x][y]
            } else {
                value = saveGame.MyrrorExplored[x][y]
            }

            if value == 0 {
                out[x][y] = data.FogTypeUnexplored
            } else {
                out[x][y] = data.FogTypeExplored
            }
        }
    }

    return out
}

func (saveGame *SaveGame) ToCities(player *playerlib.Player, playerIndex int8, game *gamelib.Game) []*citylib.City {
    cities := []*citylib.City{}

    for index := 0; index < int(saveGame.NumCities); index++ {
        cityData := saveGame.Cities[index]

        if cityData.Owner != playerIndex {
            continue
        }

        plane := data.Plane(cityData.Plane)

        var race data.Race
        switch cityData.Race {
            case 0: race = data.RaceBarbarian
            case 1: race = data.RaceBeastmen
            case 2: race = data.RaceDarkElf
            case 3: race = data.RaceDraconian
            case 4: race = data.RaceDwarf
            case 5: race = data.RaceGnoll
            case 6: race = data.RaceHalfling
            case 7: race = data.RaceHighElf
            case 8: race = data.RaceHighMen
            case 9: race = data.RaceKlackon
            case 10: race = data.RaceLizard
            case 11: race = data.RaceNomad
            case 12: race = data.RaceOrc
            case 13: race = data.RaceTroll
        }

        // FIXME: parse cityData.NumBuildings and cityData.Buildings
        buildings := set.MakeSet[buildinglib.Building]()

        // FIXME: parse cityData.Enchantments
        enchantments := set.MakeSet[citylib.Enchantment]()

        // FIXME: parse cityData.Construction
        producingBuilding := buildinglib.BuildingTradeGoods
        producingUnit := units.UnitNone

        catchmentProvider := game.ArcanusMap
        if plane == data.PlaneMyrror {
            catchmentProvider = game.MyrrorMap
        }

        fmt.Printf("%v\n", cityData.Size)

        city := citylib.City{
            Population: 1000 * int(cityData.Population) + 10 * int(cityData.Population10),
            Farmers: int(cityData.Farmers),
            Workers: int(cityData.Population - cityData.Farmers),
            Rebels: 0,
            Name: string(cityData.Name),
            Plane: plane,
            Race: race,
            X: int(cityData.X),
            Y: int(cityData.Y),
            Outpost: cityData.Size == 0,
            Buildings: buildings,
            SoldBuilding: cityData.SoldBuilding == 1,
            Enchantments: enchantments,
            Production: float32(cityData.Production),
            ProducingBuilding: producingBuilding,
            ProducingUnit: producingUnit,
            CatchmentProvider: catchmentProvider,
            CityServices: game,
            ReignProvider: player,
            BuildingInfo: game.BuildingInfo,
        }
        city.UpdateUnrest()
        cities = append(cities, &city)
    }

    return cities
}