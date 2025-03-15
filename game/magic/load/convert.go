package load

import (
    "image"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"

    "github.com/hajimehoshi/ebiten/v2"
)

func (saveGame *SaveGame) ConvertMap(terrainData *terrain.TerrainData, plane data.Plane, cityProvider maplib.CityProvider) *maplib.Map {

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

func (saveGame *SaveGame) ConvertSettings() setup.NewGameSettings {
    return setup.NewGameSettings{
        Difficulty:  data.DifficultySetting(saveGame.Difficulty),
        Opponents: int(saveGame.NumPlayers) - 1,
        LandSize: int(saveGame.LandSize),
        Magic: data.MagicSetting(saveGame.Magic),
    }
}

func (saveGame *SaveGame) convertWizard(index int) setup.WizardCustom {
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

func (saveGame *SaveGame) convertFogMap(plane data.Plane) data.FogMap {
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

func (saveGame *SaveGame) convertCities(player *playerlib.Player, playerIndex int8, game *gamelib.Game) []*citylib.City {
    cities := []*citylib.City{}
    buildingMap := map[int]buildinglib.Building{
        0x01: buildinglib.BuildingTradeGoods,
        0x02: buildinglib.BuildingHousing,
        0x03: buildinglib.BuildingBarracks,
        0x04: buildinglib.BuildingArmory,
        0x05: buildinglib.BuildingFightersGuild,
        0x06: buildinglib.BuildingArmorersGuild,
        0x07: buildinglib.BuildingWarCollege,
        0x08: buildinglib.BuildingSmithy,
        0x09: buildinglib.BuildingStables,
        0x0A: buildinglib.BuildingAnimistsGuild,
        0x0B: buildinglib.BuildingFantasticStable,
        0x0C: buildinglib.BuildingShipwrightsGuild,
        0x0D: buildinglib.BuildingShipYard,
        0x0E: buildinglib.BuildingMaritimeGuild,
        0x0F: buildinglib.BuildingSawmill,
        0x10: buildinglib.BuildingLibrary,
        0x11: buildinglib.BuildingSagesGuild,
        0x12: buildinglib.BuildingOracle,
        0x13: buildinglib.BuildingAlchemistsGuild,
        0x14: buildinglib.BuildingUniversity,
        0x15: buildinglib.BuildingWizardsGuild,
        0x16: buildinglib.BuildingShrine,
        0x17: buildinglib.BuildingTemple,
        0x18: buildinglib.BuildingParthenon,
        0x19: buildinglib.BuildingCathedral,
        0x1A: buildinglib.BuildingMarketplace,
        0x1B: buildinglib.BuildingBank,
        0x1C: buildinglib.BuildingMerchantsGuild,
        0x1D: buildinglib.BuildingGranary,
        0x1E: buildinglib.BuildingFarmersMarket,
        0x1F: buildinglib.BuildingForestersGuild,
        0x20: buildinglib.BuildingBuildersHall,
        0x21: buildinglib.BuildingMechaniciansGuild,
        0x22: buildinglib.BuildingMinersGuild,
        0x23: buildinglib.BuildingCityWalls,
    }

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

        buildings := set.MakeSet[buildinglib.Building]()
        for index, building := range buildingMap {
            if int8(cityData.Buildings[index]) == 0 || int8(cityData.Buildings[index]) == 1 {
                buildings.Insert(building)
            }
        }

        // FIXME: parse cityData.Enchantments
        enchantments := set.MakeSet[citylib.Enchantment]()

        producingBuilding := buildinglib.BuildingNone
        producingUnit := units.UnitNone
        if cityData.Construction < 100 {  // FIXME: verify
            producingBuilding = buildingMap[int(cityData.Construction)]
        } else {
            producingUnit = units.UnitNone // FIXME: parse unit
        }

        catchmentProvider := game.ArcanusMap
        if plane == data.PlaneMyrror {
            catchmentProvider = game.MyrrorMap
        }

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

func (saveGame *SaveGame) Convert(cache *lbx.LbxCache) *gamelib.Game {
    game := gamelib.MakeGame(cache, saveGame.ConvertSettings())

    // load data
    game.ArcanusMap = saveGame.ConvertMap(game.ArcanusMap.Data, data.PlaneArcanus, nil)
    game.MyrrorMap = saveGame.ConvertMap(game.MyrrorMap.Data, data.PlaneMyrror, nil)
    game.TurnNumber = uint64(saveGame.Turn)
    // FIXME: game.ArtifactPool
    // FIXME: game.RandomEvents
    // FIXME: game.RoadWorkArcanus
    // FIXME: game.RoadWorkMyrror
    // FIXME: game.PurifyWorkArcanus
    // FIXME: game.PurifyWorkMyrror
    // FIXME: game.Players

    wizard := saveGame.convertWizard(0)

    player := game.AddPlayer(wizard, true)
    player.ArcanusFog = saveGame.convertFogMap(data.PlaneArcanus)
    player.MyrrorFog = saveGame.convertFogMap(data.PlaneMyrror)
    player.Cities = saveGame.convertCities(player, 0, game)
    player.UpdateFogVisibility()

    for i := 1; i < int(saveGame.NumPlayers); i++ {
        wizard := saveGame.convertWizard(i)
        enemy := game.AddPlayer(wizard, false)
        enemy.Cities = saveGame.convertCities(enemy, int8(i), game)
    }

    // FIXME: neutral player

    game.Camera.Center(20, 20)
    if len(player.Cities) > 0 {
        game.Camera.Center(player.Cities[0].X, player.Cities[0].Y)
    }

    return game
}