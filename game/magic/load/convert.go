package load

import (
    "bytes"
    "image"
    "fmt"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/ai"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"

    "github.com/hajimehoshi/ebiten/v2"
)

func fromRaceValue(raceValue int) data.Race {
    switch raceValue {
        case 0: return data.RaceBarbarian
        case 1: return data.RaceBeastmen
        case 2: return data.RaceDarkElf
        case 3: return data.RaceDraconian
        case 4: return data.RaceDwarf
        case 5: return data.RaceGnoll
        case 6: return data.RaceHalfling
        case 7: return data.RaceHighElf
        case 8: return data.RaceHighMen
        case 9: return data.RaceKlackon
        case 10: return data.RaceLizard
        case 11: return data.RaceNomad
        case 12: return data.RaceOrc
        case 13: return data.RaceTroll
    }

    return data.RaceNone
}

func toRaceInt(race data.Race) int {
    switch race {
        case data.RaceBarbarian: return 0
        case data.RaceBeastmen: return 1
        case data.RaceDarkElf: return 2
        case data.RaceDraconian: return 3
        case data.RaceDwarf: return 4
        case data.RaceGnoll: return 5
        case data.RaceHalfling: return 6
        case data.RaceHighElf: return 7
        case data.RaceHighMen: return 8
        case data.RaceKlackon: return 9
        case data.RaceLizard: return 10
        case data.RaceNomad: return 11
        case data.RaceOrc: return 12
        case data.RaceTroll: return 13
    }

    return 0
}

func (saveGame *SaveGame) ConvertMap(terrainData *terrain.TerrainData, plane data.Plane, cityProvider maplib.CityProvider, players []*playerlib.Player) *maplib.Map {

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
    terrainFlags := saveGame.ArcanusMapSquareFlags
    if plane == data.PlaneMyrror {
        terrainSource = saveGame.MyrrorMap
        terrainOffset = terrain.MyrrorStart
        terrainSpecials = saveGame.MyrrorTerrainSpecials
        terrainFlags = saveGame.MyrrorMapSquareFlags
    }

    for y := range(WorldHeight) {
        for x := range(WorldWidth) {
            point := image.Pt(x, y)

            map_.Map.Terrain[x][y] = int(terrainSource.Data[x][y]) + terrainOffset
            map_.ExtraMap[point] = make(map[maplib.ExtraKind]maplib.ExtraTile)

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
                map_.ExtraMap[point][maplib.ExtraKindBonus] = &maplib.ExtraBonus{Bonus: bonus}
            }

            if terrainFlags[x][y] & 0x20 != 0 {
                map_.SetCorruption(x, y)
            }

            if terrainFlags[x][y] & 0x08 != 0 {
                map_.SetRoad(x, y, false)
            }

            if terrainFlags[x][y] & 0x10 != 0 {
                map_.SetRoad(x, y, true)
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

    for _, node := range saveGame.Nodes {

        if node.Plane == 0 && plane != data.PlaneArcanus || node.Plane != 0 && plane != data.PlaneMyrror {
            continue
        }

        var nodeType maplib.MagicNode
        switch NodeType(node.NodeType) {
            case NodeTypeSorcery:
                nodeType = maplib.MagicNodeSorcery
            case NodeTypeNature:
                nodeType = maplib.MagicNodeNature
            case NodeTypeChaos:
                nodeType = maplib.MagicNodeChaos
        }

        var meldingWizard maplib.Wizard
        if node.Owner > -1 && players != nil {
            meldingWizard = players[node.Owner]
        }

        zone := []image.Point{}
        for i := range node.Power {
            zone = append(zone, image.Pt(int(node.AuraX[i])-int(node.X), int(node.AuraY[i])-int(node.Y)))
        }

        point := image.Pt(int(node.X), int(node.Y))
        if map_.ExtraMap[point] == nil {
            map_.ExtraMap[point] = make(map[maplib.ExtraKind]maplib.ExtraTile)
        }
        map_.ExtraMap[point][maplib.ExtraKindMagicNode] = &maplib.ExtraMagicNode{
            Kind: nodeType,
            Zone: zone,
            MeldingWizard: meldingWizard,
            Warped: (node.Flags & 0x01) != 0,
            GuardianSpiritMeld: (node.Flags & 0x02) != 0,
            // FIXME: WarpedOwner
        }
    }

    return &map_
}

func (saveGame *SaveGame) convertSettings() setup.NewGameSettings {
    return setup.NewGameSettings{
        Difficulty:  data.DifficultySetting(saveGame.Difficulty),
        Opponents: int(saveGame.NumPlayers) - 1,
        LandSize: int(saveGame.LandSize),
        Magic: data.MagicSetting(saveGame.Magic),
    }
}

func (saveGame *SaveGame) convertWizard(playerIndex int) setup.WizardCustom {
    playerData := saveGame.PlayerData[playerIndex]

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

    race := fromRaceValue(int(playerData.CapitalRace))

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

func (saveGame *SaveGame) convertCities(player *playerlib.Player, playerIndex int, wizards []setup.WizardCustom, game *gamelib.Game) []*citylib.City {
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

    enchantmentMap := map[int]data.CityEnchantment{
        0x00: data.CityEnchantmentWallOfFire,
        0x01: data.CityEnchantmentChaosRift,
        0x02: data.CityEnchantmentDarkRituals,
        0x03: data.CityEnchantmentEvilPresence,
        0x04: data.CityEnchantmentCursedLands,
        0x05: data.CityEnchantmentPestilence,
        0x06: data.CityEnchantmentCloudOfShadow,
        0x07: data.CityEnchantmentFamine,
        0x08: data.CityEnchantmentFlyingFortress,
        0x09: data.CityEnchantmentNatureWard,
        0x0A: data.CityEnchantmentSorceryWard,
        0x0B: data.CityEnchantmentChaosWard,
        0x0C: data.CityEnchantmentLifeWard,
        0x0D: data.CityEnchantmentDeathWard,
        0x0E: data.CityEnchantmentNaturesEye,
        0x0F: data.CityEnchantmentEarthGate,
        0x10: data.CityEnchantmentStreamOfLife,
        0x11: data.CityEnchantmentGaiasBlessing,
        0x12: data.CityEnchantmentInspirations,
        0x13: data.CityEnchantmentProsperity,
        0x14: data.CityEnchantmentAstralGate,
        0x15: data.CityEnchantmentHeavenlyLight,
        0x16: data.CityEnchantmentConsecration,
        0x17: data.CityEnchantmentWallOfDarkness,
        0x18: data.CityEnchantmentAltarOfBattle,
        // 0x19: data.CityEnchantmentNightshade, // FIXME add nightshade
    }

    unitMap := map[int]units.Unit {
        0: units.HeroBrax,
        1: units.HeroGunther,
        2: units.HeroZaldron,
        3: units.HeroBShan,
        4: units.HeroRakir,
        5: units.HeroValana,
        6: units.HeroBahgtru,
        7: units.HeroSerena,
        8: units.HeroShuri,
        9: units.HeroTheria,
        10: units.HeroGreyfairer,
        11: units.HeroTaki,
        12: units.HeroReywind,
        13: units.HeroMalleus,
        14: units.HeroTumu,
        15: units.HeroJaer,
        16: units.HeroMarcus,
        17: units.HeroFang,
        18: units.HeroMorgana,
        19: units.HeroAureus,
        20: units.HeroShinBo,
        21: units.HeroSpyder,
        22: units.HeroShalla,
        23: units.HeroYramrag,
        24: units.HeroMysticX,
        25: units.HeroAerie,
        26: units.HeroDethStryke,
        27: units.HeroElana,
        28: units.HeroRoland,
        29: units.HeroMortu,
        30: units.HeroAlorra,
        31: units.HeroSirHarold,
        32: units.HeroRavashack,
        33: units.HeroWarrax,
        34: units.HeroTorin,
        35: units.Trireme,
        36: units.Galley,
        37: units.Catapult,
        38: units.Warship,
        39: units.BarbarianSpearmen,
        40: units.BarbarianSwordsmen,
        41: units.BarbarianBowmen,
        42: units.BarbarianCavalry,
        43: units.BarbarianShaman,
        44: units.BarbarianSettlers,
        45: units.Berserkers,
        46: units.BeastmenSpearmen,
        47: units.BeastmenSwordsmen,
        48: units.BeastmenHalberdiers,
        49: units.BeastmenBowmen,
        50: units.BeastmenPriest,
        51: units.BeastmenMagician,
        52: units.BeastmenEngineer,
        53: units.BeastmenSettlers,
        54: units.Centaur,
        55: units.Manticore,
        56: units.Minotaur,
        57: units.DarkElfSpearmen,
        58: units.DarkElfSwordsmen,
        59: units.DarkElfHalberdiers,
        60: units.DarkElfCavalry,
        61: units.DarkElfPriests,
        62: units.DarkElfSettlers,
        63: units.Nightblades,
        64: units.Warlocks,
        65: units.Nightmares,
        66: units.DraconianSpearmen,
        67: units.DraconianSwordsmen,
        68: units.DraconianHalberdiers,
        69: units.DraconianBowmen,
        70: units.DraconianShaman,
        71: units.DraconianMagician,
        // 72: units.DraconianEngineer
        73: units.DraconianSettlers,
        74: units.DoomDrake,
        75: units.AirShip,
        76: units.DwarfSwordsmen,
        77: units.DwarfHalberdiers,
        78: units.DwarfEngineer,
        79: units.Hammerhands,
        80: units.SteamCannon,
        81: units.Golem,
        82: units.DwarfSettlers,
        83: units.GnollSpearmen,
        84: units.GnollSwordsmen,
        85: units.GnollHalberdiers,
        86: units.GnollBowmen,
        87: units.GnollSettlers,
        88: units.WolfRiders,
        89: units.HalflingSpearmen,
        90: units.HalflingSwordsmen,
        91: units.HalflingBowmen,
        92: units.HalflingShamans,
        93: units.HalflingSettlers,
        94: units.Slingers,
        95: units.HighElfSpearmen,
        96: units.HighElfSwordsmen,
        97: units.HighElfHalberdiers,
        98: units.HighElfCavalry,
        99: units.HighElfMagician,
        100: units.HighElfSettlers,
        101: units.Longbowmen,
        102: units.ElvenLord,
        103: units.Pegasai,
        104: units.HighMenSpearmen,
        105: units.HighMenSwordsmen,
        106: units.HighMenBowmen,
        107: units.HighMenCavalry,
        108: units.HighMenPriest,
        109: units.HighMenMagician,
        110: units.HighMenEngineer,
        111: units.HighMenSettlers,
        112: units.HighMenPikemen,
        113: units.Paladin,
        114: units.KlackonSpearmen,
        115: units.KlackonSwordsmen,
        116: units.KlackonHalberdiers,
        117: units.KlackonEngineer,
        118: units.KlackonSettlers,
        119: units.StagBeetle,
        120: units.LizardSpearmen,
        121: units.LizardSwordsmen,
        122: units.LizardHalberdiers,
        123: units.LizardJavelineers,
        124: units.LizardShamans,
        125: units.LizardSettlers,
        126: units.DragonTurtle,
        127: units.NomadSpearmen,
        128: units.NomadSwordsmen,
        129: units.NomadBowmen,
        130: units.NomadPriest,
        // 131: units.NomadMagicians
        132: units.NomadSettlers,
        133: units.NomadHorsebowemen,
        134: units.NomadPikemen,
        135: units.NomadRangers,
        136: units.Griffin,
        137: units.OrcSpearmen,
        138: units.OrcSwordsmen,
        139: units.OrcHalberdiers,
        140: units.OrcBowmen,
        141: units.OrcCavalry,
        142: units.OrcShamans,
        143: units.OrcMagicians,
        144: units.OrcEngineers,
        145: units.OrcSettlers,
        146: units.WyvernRiders,
        147: units.TrollSpearmen,
        148: units.TrollSwordsmen,
        149: units.TrollHalberdiers,
        150: units.TrollShamans,
        151: units.TrollSettlers,
        152: units.WarTrolls,
        153: units.WarMammoths,
        154: units.MagicSpirit,
        155: units.HellHounds,
        156: units.Gargoyle,
        157: units.FireGiant,
        158: units.FireElemental,
        159: units.ChaosSpawn,
        160: units.Chimeras,
        161: units.DoomBat,
        162: units.Efreet,
        163: units.Hydra,
        164: units.GreatDrake,
        165: units.Skeleton,
        166: units.Ghoul,
        167: units.NightStalker,
        168: units.WereWolf,
        169: units.Demon,
        170: units.Wraith,
        171: units.ShadowDemons,
        172: units.DeathKnights,
        173: units.DemonLord,
        174: units.Zombie,
        175: units.Unicorn,
        176: units.GuardianSpirit,
        177: units.Angel,
        178: units.ArchAngel,
        179: units.WarBear,
        180: units.Sprites,
        181: units.Cockatrices,
        182: units.Basilisk,
        183: units.GiantSpiders,
        184: units.StoneGiant,
        185: units.Colossus,
        186: units.Gorgon,
        187: units.EarthElemental,
        188: units.Behemoth,
        189: units.GreatWyrm,
        190: units.FloatingIsland,
        191: units.PhantomBeast,
        192: units.PhantomWarrior,
        193: units.StormGiant,
        194: units.AirElemental,
        195: units.Djinn,
        196: units.SkyDrake,
        197: units.Nagas,
    }

    for index := range int(saveGame.NumCities) {
        cityData := saveGame.Cities[index]

        if int(cityData.Owner) != playerIndex {
            continue
        }

        plane := data.Plane(cityData.Plane)

        race := fromRaceValue(int(cityData.Race))

        buildings := set.MakeSet[buildinglib.Building]()
        for index, building := range buildingMap {
            if int8(cityData.Buildings[index]) == 0 || int8(cityData.Buildings[index]) == 1 {
                buildings.Insert(building)
            }
        }
        if cityData.X == saveGame.Fortresses[playerIndex].X && cityData.Y == saveGame.Fortresses[playerIndex].Y && cityData.Plane == saveGame.Fortresses[playerIndex].Plane && saveGame.Fortresses[playerIndex].Active == 1 {
            buildings.Insert(buildinglib.BuildingFortress)
        }
        if int16(cityData.X) == saveGame.PlayerData[playerIndex].SummonX && int16(cityData.Y) == saveGame.PlayerData[playerIndex].SummonY && int16(cityData.Plane) == saveGame.PlayerData[playerIndex].SummonPlane {
            buildings.Insert(buildinglib.BuildingSummoningCircle)
        }

        enchantments := set.MakeSet[citylib.Enchantment]()
        for index, enchantment := range enchantmentMap {
            value := int8(cityData.Enchantments[index])
            if value != 0 {
                enchantments.Insert(citylib.Enchantment{
                    Enchantment: enchantment,
                    Owner: wizards[value-1].Banner,
                })
            }
        }

        producingBuilding := buildinglib.BuildingNone
        producingUnit := units.UnitNone
        if cityData.Construction < 100 {
            producingBuilding = buildingMap[int(cityData.Construction)]
        } else {
            producingUnit = unitMap[int(cityData.Construction)-100]
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

func (saveGame *SaveGame) convertPlayer(playerIndex int, wizards []setup.WizardCustom, artifacts []*artifact.Artifact, game *gamelib.Game) *playerlib.Player {
    playerData := saveGame.PlayerData[playerIndex]
    human := playerIndex == 0

    var aiBehavior playerlib.AIBehavior
    if !human {
        aiBehavior = ai.MakeEnemyAI()
    }

    enchantmentMap := map[int]data.Enchantment{
        0x00: data.EnchantmentEternalNight,
        0x01: data.EnchantmentEvilOmens,
        0x02: data.EnchantmentZombieMastery,
        0x03: data.EnchantmentAuraOfMajesty,
        0x04: data.EnchantmentWindMastery,
        0x05: data.EnchantmentSuppressMagic,
        0x06: data.EnchantmentTimeStop,
        0x07: data.EnchantmentNatureAwareness,
        0x08: data.EnchantmentNaturesWrath,
        0x09: data.EnchantmentHerbMastery,
        0x0A: data.EnchantmentChaosSurge,
        0x0B: data.EnchantmentDoomMastery,
        0x0C: data.EnchantmentGreatWasting,
        0x0D: data.EnchantmentMeteorStorm,
        0x0E: data.EnchantmentArmageddon,
        0x0F: data.EnchantmentTranquility,
        0x10: data.EnchantmentLifeForce,
        0x11: data.EnchantmentCrusade,
        0x12: data.EnchantmentJustCause,
        0x13: data.EnchantmentHolyArms,
        0x14: data.EnchantmentPlanarSeal,
        0x15: data.EnchantmentCharmOfLife,
        0x16: data.EnchantmentDetectMagic,
        0x17: data.EnchantmentAwareness,
    }

    globalEnchantments := set.MakeSet[data.Enchantment]()
    for index, enchantment := range enchantmentMap {
        value := int8(playerData.GlobalEnchantments[index])
        if value != 0 {
            globalEnchantments.Insert(enchantment)
        }
    }

    // FIXME: parse player relations
    playerRelations := make(map[*playerlib.Player]*playerlib.Relationship)

    var arcanusFog, myrrorFog data.FogMap
    if human {
        arcanusFog = saveGame.convertFogMap(data.PlaneArcanus)
        myrrorFog = saveGame.convertFogMap(data.PlaneMyrror)
    } else {
        arcanusFog = make([][]data.FogType, WorldWidth)
        myrrorFog = make([][]data.FogType, WorldWidth)
        for x := range WorldWidth {
            arcanusFog[x] = make([]data.FogType, WorldHeight)
            myrrorFog[x] = make([]data.FogType, WorldHeight)
            for y := range WorldHeight{
                arcanusFog[x][y] = data.FogTypeExplored
                myrrorFog[x][y] = data.FogTypeExplored
            }
        }
    }

    spellMap := make(map[int]spellbook.Spell)
    for index, spell := range game.AllSpells().Spells {
        spellMap[index] = spell
    }

    researchingSpell := spellMap[int(playerData.ResearchingSpellIndex)]
    researchCandidateSpells := spellbook.Spells{}
    researchCandidateSpells.Spells = append(researchCandidateSpells.Spells, researchingSpell)

    knownSpells := spellbook.Spells{}
    researchPoolSpells := spellbook.Spells{}
    for index, spell := range playerData.SpellsList {
        if spell > 0 {
            researchPoolSpells.Spells = append(researchPoolSpells.Spells, spellMap[index+1])
        }
        if spell == 2 {
            knownSpells.Spells = append(knownSpells.Spells, spellMap[index+1])
        }
    }

    var vaultEquipment [4]*artifact.Artifact
    for spot, index := range playerData.VaultItems {
        if index != -1 {
            vaultEquipment[spot] = artifacts[index]
        }
    }

    // FIXME: Add remaining infos from playerData
    // Personality
    // Objective
    // VolcanoPower
    // AverageUnitCost
    // CombatSkillLeft
    // SkillLeft
    // NominalSkill
    // Diplomacy
    // SpellCastingSkill
    // DefeatedWizards
    // Astrology
    // Population
    // Historian
    // Hostility
    // ReevaluateHostilityCountdown
    // ReevaluateMagicStrategyCountdown
    // ReevaluateMagicPowerCountdown
    // PeaceDuration
    // TargetWizard
    // PrimaryRealm
    // SecondaryRealm

    // doesn't seem necessary to store this since we compute it anyway
    // PowerBase
    // Volcanoes

    player := playerlib.Player{
        Wizard: wizards[playerIndex],
        SpellOfMasteryCost: int(playerData.MasteryResearch),
        TaxRate: fraction.Make(int(playerData.TaxRate), 2),
        PowerDistribution: playerlib.PowerDistribution{
            Mana: float64(playerData.ManaRatio) / 100,
            Research: float64(playerData.ResearchRatio) / 100,
            Skill: float64(playerData.SkillRatio) / 100,
        },
        Gold: int(playerData.GoldReserve),
        Mana: int(playerData.ManaReserve),
        Human: human,
        AIBehavior: aiBehavior,
        // FIXME: Defeated
        // FIXME: Banished
        Fame: int(playerData.Fame),
        BookOrderSeed1: rand.Uint64(),
        BookOrderSeed2: rand.Uint64(),
        StrategicCombat: !human,
        Admin: false,
        KnownSpells: knownSpells,
        ResearchCandidateSpells: researchCandidateSpells,
        ResearchPoolSpells: researchPoolSpells,
        ResearchingSpell: researchingSpell,
        ResearchProgress: researchingSpell.ResearchCost - int(playerData.ResearchCostRemaining),
        CastingSpell: spellMap[int(playerData.CastingSpellIndex)],
        CastingSpellProgress: int(playerData.CastingCostOriginal - playerData.CastingCostRemaining),
        // FIXME: CastingSkillPower
        // FIXME: RemainingCastingSkill
        GlobalEnchantments: globalEnchantments,
        GlobalEnchantmentsProvider: game,
        PlayerRelations: playerRelations,
        // FIXME: HeroPool createHeroes(herolib.ReadNamesPerWizard(game.Cache))
        // FIXME: Heroes
        VaultEquipment: vaultEquipment,
        // FIXME: CreateArtifact
        // FIXME: Units
        // FIXME: Stacks
        // FIXME: UnitId
        // FIXME: SelectedStack
        ArcanusFog: arcanusFog,
        MyrrorFog: myrrorFog,
    }

    player.Cities = saveGame.convertCities(&player, playerIndex, wizards, game)
    player.UpdateResearchCandidates()
    player.UpdateFogVisibility()

    return &player
}

func (saveGame *SaveGame) convertArtifacts(spells spellbook.Spells) []*artifact.Artifact {
    _, typeMap, abilityMap := artifact.GetItemConversionMaps()

    artifacts := []*artifact.Artifact{}
    for _, item := range saveGame.Items {

        if item.Cost == 0 {
            continue
        }

        powers := []artifact.Power{}

        if item.Attack != 0 {
            powers = append(powers, artifact.Power{Type: artifact.PowerTypeAttack, Amount: int(item.Attack), Name: fmt.Sprintf("+%v Attack", item.Attack)})
        }

        if item.ToHit != 0 {
            powers = append(powers, artifact.Power{Type: artifact.PowerTypeToHit, Amount: int(item.ToHit), Name: fmt.Sprintf("+%v To Hit", item.ToHit)})
        }

        if item.Defense != 0 {
            powers = append(powers, artifact.Power{Type: artifact.PowerTypeDefense, Amount: int(item.Defense), Name: fmt.Sprintf("+%v Defense", item.Defense)})
        }

        if item.Movement != 0 {
            powers = append(powers, artifact.Power{Type: artifact.PowerTypeMovement, Amount: int(item.Movement), Name: fmt.Sprintf("+%v Movement", item.Movement)})
        }

        if item.Resistance != 0 {
            powers = append(powers, artifact.Power{Type: artifact.PowerTypeResistance, Amount: int(item.Resistance), Name: fmt.Sprintf("+%v Resistance", item.Resistance)})
        }

        if item.SpellSkill != 0 {
            powers = append(powers, artifact.Power{Type: artifact.PowerTypeSpellSkill, Amount: int(item.SpellSkill), Name: fmt.Sprintf("+%v Spell Skill", item.SpellSkill)})
        }

        if item.SpellSave != 0 {
            powers = append(powers, artifact.Power{Type: artifact.PowerTypeSpellSave, Amount: int(item.SpellSave), Name: fmt.Sprintf("+%v Spell Save", item.SpellSave)})
        }

        if item.Spell != 0 && item.Charges != 0 {
            useSpell := spells.FindById(int(item.Spell))
            powers = append(powers, artifact.Power{
                Type: artifact.PowerTypeSpellCharges,
                Amount: int(item.Charges),
                Spell: useSpell,
                SpellCharges: int(item.Charges),
                Name: fmt.Sprintf("%v Charges of %v", item.Charges, useSpell.Name),
            })
        }

        for mask, ability := range abilityMap {
            if item.Abilities & mask != 0 {
                powers = append(powers, artifact.Power{Type: artifact.PowerTypeAbility1, Amount: 0, Name: ability.Name(), Ability: ability})
            }
        }

        artifacts = append(artifacts, &artifact.Artifact{
            Name: string(bytes.Trim(item.Name, "\x00")),
            Image: int(item.IconIndex),
            Type: typeMap[item.Type],
            Cost: int(item.Cost),
            Powers: powers,
        })
    }

    return artifacts
}

func (saveGame *SaveGame) Convert(cache *lbx.LbxCache) *gamelib.Game {
    game := gamelib.MakeGame(cache, saveGame.convertSettings())
    game.TurnNumber = uint64(saveGame.Turn)

    artifacts := saveGame.convertArtifacts(game.AllSpells())
    for _, artifact := range artifacts {
        _, ok := game.ArtifactPool[artifact.Name]
        if ok {
            delete(game.ArtifactPool, artifact.Name)
        }
    }

    wizards := []setup.WizardCustom{}
    for playerIndex := range saveGame.NumPlayers {
        wizards = append(wizards, saveGame.convertWizard(int(playerIndex)))
    }

    for playerIndex := range saveGame.NumPlayers {
        player := saveGame.convertPlayer(int(playerIndex), wizards, artifacts, game)
        game.Players = append(game.Players, player)
    }
    // FIXME: add neutral player with brown banner and ai.MakeRaiderAI()

    // FIXME: add all remaining information from saveGame
    // saveGame.Unit
    // saveGame.HeroData
    // saveGame.GrandVizier
    // saveGame.Units / saveGame.NumUnits
    // saveGame.Events

    game.ArcanusMap = saveGame.ConvertMap(game.ArcanusMap.Data, data.PlaneArcanus, game, game.Players)
    game.MyrrorMap = saveGame.ConvertMap(game.MyrrorMap.Data, data.PlaneMyrror, game, game.Players)
    // FIXME: game.RandomEvents
    // FIXME: game.RoadWorkArcanus
    // FIXME: game.RoadWorkMyrror
    // FIXME: game.PurifyWorkArcanus
    // FIXME: game.PurifyWorkMyrror

    game.Camera.Center(20, 20)
    if len(game.Players[0].Cities) > 0 {
        game.Camera.Center(game.Players[0].Cities[0].X, game.Players[0].Cities[0].Y)
    }

    return game
}
