package load

import (
    "image"
    "fmt"
    "maps"
    "math/rand/v2"
    "log"

    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
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

func firstNonZero(arr []byte) []byte {
    var out []byte
    for _, b := range arr {
        if b == 0 {
            break
        }
        out = append(out, b)
    }
    return out
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
        if (node.Plane == 0 && plane != data.PlaneArcanus) || (node.Plane != 0 && plane != data.PlaneMyrror) {
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

func (saveGame *SaveGame) convertCities(player *playerlib.Player, playerIndex int, wizards []setup.WizardCustom, game *gamelib.Game, arcanusMap *maplib.Map, myrrorMap *maplib.Map) []*citylib.City {
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

    for index := range int(saveGame.NumCities) {
        cityData := saveGame.Cities[index]

        if int(cityData.Owner) != playerIndex {
            continue
        }

        plane := data.Plane(cityData.Plane)

        race := fromRaceValue(int(cityData.Race))

        // log.Printf("City %v buildings %v", string(cityData.Name), cityData.Buildings)

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
            producingUnit = getUnitType(int(cityData.Construction)-100)
        }

        // log.Printf("City data: %+v", cityData)
        catchmentProvider := game.ArcanusMap
        if plane == data.PlaneMyrror {
            catchmentProvider = game.MyrrorMap
        }

        city := citylib.City{
            Population: 1000 * int(cityData.Population) + 10 * int(cityData.Population10),
            Farmers: int(cityData.Farmers),
            Workers: int(cityData.Population - cityData.Farmers),
            Rebels: 0,
            Name: string(firstNonZero(cityData.Name)),
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
            CityServices: game,
            CatchmentProvider: catchmentProvider,
            ReignProvider: player,
            BuildingInfo: game.BuildingInfo,
        }

        city.UpdateUnrest()
        cities = append(cities, &city)
    }

    return cities
}

func convertHeroAbility(ability HeroAbility) data.Ability {

    switch ability {
        case HeroAbility_CHARMED: return data.MakeAbility(data.AbilityCharmed)
        case HeroAbility_PRAYERMASTER: return data.MakeAbility(data.AbilityPrayermaster)
        case HeroAbility_PRAYERMASTER2: return data.MakeAbility(data.AbilitySuperPrayermaster)
        case HeroAbility_LEADERSHIP: return data.MakeAbility(data.AbilityLeadership)
        case HeroAbility_LEADERSHIP2: return data.MakeAbility(data.AbilitySuperLeadership)
        case HeroAbility_LEGENDARY: return data.MakeAbility(data.AbilityLegendary)
        case HeroAbility_LEGENDARY2: return data.MakeAbility(data.AbilitySuperLegendary)
        case HeroAbility_BLADEMASTER: return data.MakeAbility(data.AbilityBlademaster)
        case HeroAbility_BLADEMASTER2: return data.MakeAbility(data.AbilitySuperBlademaster)
        case HeroAbility_ARMSMASTER: return data.MakeAbility(data.AbilityArmsmaster)
        case HeroAbility_ARMSMASTER2: return data.MakeAbility(data.AbilitySuperArmsmaster)
        case HeroAbility_CONSTITUTION: return data.MakeAbility(data.AbilityConstitution)
        case HeroAbility_CONSTITUTION2: return data.MakeAbility(data.AbilitySuperConstitution)
        case HeroAbility_MIGHT: return data.MakeAbility(data.AbilityMight)
        case HeroAbility_MIGHT2: return data.MakeAbility(data.AbilitySuperMight)
        case HeroAbility_ARCANE_POWER: return data.MakeAbility(data.AbilityArcanePower)
        case HeroAbility_ARCANE_POWER2: return data.MakeAbility(data.AbilitySuperArcanePower)
        case HeroAbility_SAGE: return data.MakeAbility(data.AbilitySage)
        case HeroAbility_SAGE2: return data.MakeAbility(data.AbilitySuperSage)
        case HeroAbility_AGILITY: return data.MakeAbility(data.AbilityAgility)
        case HeroAbility_AGILITY2: return data.MakeAbility(data.AbilitySuperAgility)
        case HeroAbility_LUCKY: return data.MakeAbility(data.AbilityLucky)
        case HeroAbility_NOBLE: return data.MakeAbility(data.AbilityNoble)
        // case HeroAbility_FEMALE: return data.MakeAbility(data.AbilityFemale)

        /*
    */

    }

    return data.MakeAbility(data.AbilityNone)
}

func isHeroAbility(ability data.Ability) bool {
    return ability.IsHeroAbility()
}

// initial casting skill pool
// https://masterofmagic.fandom.com/wiki/Caster#Improvement_Table
func convertCastingSkill(castingSkill int8) float32 {
    switch castingSkill {
        case 0: return 0
        case 1: return 5
        case 2: return 7.5
        case 3: return 10
        case 4: return 12.5
        case 5: return 15
        case 6: return 17.5
        case 7: return 20
    }

    return 0
}

func setHeroData(hero *herolib.Hero, heroData *HeroData) {
    // the hero object might have some abilities already initialized on it
    maps.DeleteFunc(hero.Abilities, func (ability data.AbilityType, value data.Ability) bool {
        return value.IsHeroAbility()
    })

    // log.Printf("  set hero data for %v to %v", hero.Name, heroData.AbilitySet)

    for _, ability := range heroData.AbilitySet.Values() {
        newAbility := convertHeroAbility(ability)
        hero.Abilities[newAbility.Ability] = newAbility
    }

    if heroData.CastingSkill != 0 {
        hero.Abilities[data.AbilityCaster] = data.MakeAbilityValue(data.AbilityCaster, convertCastingSkill(heroData.CastingSkill))
    }
}

func (saveGame *SaveGame) convertPlayer(playerIndex int, wizards []setup.WizardCustom, artifacts []*artifact.Artifact, game *gamelib.Game) (*playerlib.Player, map[*playerlib.UnitStack]image.Point, func()) {
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
    for _, spell := range game.AllSpells().Spells {
        spellMap[spell.Index] = spell
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

    convertNominalSkill := func(nominalSkill uint16) int {
        var z int = int(nominalSkill)
        return z * z - z + 1
    }

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
        RemainingCastingSkill: int(playerData.SkillLeft),
        CastingSkillPower: convertNominalSkill(playerData.NominalSkill),
        // FIXME: CastingSkillPower
        // FIXME: RemainingCastingSkill
        GlobalEnchantments: globalEnchantments,
        GlobalEnchantmentsProvider: game,
        PlayerRelations: playerRelations,
        // FIXME: HeroPool createHeroes(herolib.ReadNamesPerWizard(game.Cache))
        // FIXME: Heroes
        VaultEquipment: vaultEquipment,
        // FIXME: CreateArtifact
        ArcanusFog: arcanusFog,
        MyrrorFog: myrrorFog,
    }

    if player.CastingSpell.Valid() && (player.CastingSpell.Name == "Create Artifact" || player.CastingSpell.Name == "Enchant Item") {
        // the artifact being created is the last one. not sure about enemy wizards
        if len(artifacts) > 0 {
            last := artifacts[len(artifacts)-1]
            player.CreateArtifact = last
            player.CastingSpellProgress = last.Cost - int(playerData.CastingCostRemaining)
            player.CastingSpell.OverrideCost = last.Cost
        }
    }

    const (
        StatusReady = 0
        StatusPatrol = 1
        StatusBuildRoad = 2
        StatusGoto = 3
        StatusReachedDest = 4
        StatusWait = 5
        StatusCasting = 6
        StatusPurify = 8
        StatusMeld = 9
        StatusSettle = 10
        StatusSeekTransport = 11
        StatusMove = 16
        StatusPurifyDone = 111
    )

    const (
        MutationNone = 0
        MutationMagicWeapon = 1
        MutationMythrilWeapon = 2
        MutationAdamantiumWeapon = 3
        MutationChaosChannelsDemonSkin = 4
        MutationChaosChannelsDemonWings = 8
        MutationChaosChannelsDemonBreath = 16
        MutationUndead = 32
        MutationStasisInit = 64
        MutationStasisLinger = 128
    )

    const (
        EnchantmentImmolation      = 0b00000000000000000000000000000001
        EnchantmentGuardianWind    = 0b00000000000000000000000000000010
        EnchantmentBerserk         = 0b00000000000000000000000000000100
        EnchantmentCloakOfFear     = 0b00000000000000000000000000001000
        EnchantmentBlackChannels   = 0b00000000000000000000000000010000
        EnchantmentWraithForm      = 0b00000000000000000000000000100000
        EnchantmentRegeneration    = 0b00000000000000000000000001000000
        EnchantmentPathfinding     = 0b00000000000000000000000010000000
        EnchantmentWaterWalking    = 0b00000000000000000000000100000000
        EnchantmentResistElements  = 0b00000000000000000000001000000000
        EnchantmentElementalArmor  = 0b00000000000000000000010000000000
        EnchantmentStoneSkin       = 0b00000000000000000000100000000000
        EnchantmentIronSkin        = 0b00000000000000000001000000000000
        EnchantmentEndurance       = 0b00000000000000000010000000000000
        EnchantmentSpellLock       = 0b00000000000000000100000000000000
        EnchantmentInvisibility    = 0b00000000000000001000000000000000
        EnchantmentWindWalking     = 0b00000000000000010000000000000000
        EnchantmentFlight          = 0b00000000000000100000000000000000
        EnchantmentResistMagic     = 0b00000000000001000000000000000000
        EnchantmentMagicImmunity   = 0b00000000000010000000000000000000
        EnchantmentFlameBlade      = 0b00000000000100000000000000000000
        EnchantmentEldritchWeapon  = 0b00000000001000000000000000000000
        EnchantmentTrueSight       = 0b00000000010000000000000000000000
        EnchantmentHolyWeapon      = 0b00000000100000000000000000000000
        EnchantmentHeroism         = 0b00000001000000000000000000000000
        EnchantmentBless           = 0b00000010000000000000000000000000
        EnchantmentLionHeart       = 0b00000100000000000000000000000000
        EnchantmentGiantStrength   = 0b00001000000000000000000000000000
        EnchantmentPlanarTravel    = 0b00010000000000000000000000000000
        EnchantmentHolyArmor       = 0b00100000000000000000000000000000
        EnchantmentRighteousness   = 0b01000000000000000000000000000000
        EnchantmentInvulnerability = 0b10000000000000000000000000000000
    )

    unitEnchantmentMap := map[int]data.UnitEnchantment{
        EnchantmentImmolation: data.UnitEnchantmentImmolation,
        EnchantmentGuardianWind: data.UnitEnchantmentGuardianWind,
        EnchantmentBerserk: data.UnitEnchantmentBerserk,
        EnchantmentCloakOfFear: data.UnitEnchantmentCloakOfFear,
        EnchantmentBlackChannels: data.UnitEnchantmentBlackChannels,
        EnchantmentWraithForm: data.UnitEnchantmentWraithForm,
        EnchantmentRegeneration: data.UnitEnchantmentRegeneration,
        EnchantmentPathfinding: data.UnitEnchantmentPathFinding,
        EnchantmentWaterWalking: data.UnitEnchantmentWaterWalking,
        EnchantmentResistElements: data.UnitEnchantmentResistElements,
        EnchantmentElementalArmor: data.UnitEnchantmentElementalArmor,
        EnchantmentStoneSkin: data.UnitEnchantmentStoneSkin,
        EnchantmentIronSkin: data.UnitEnchantmentIronSkin,
        EnchantmentEndurance: data.UnitEnchantmentEndurance,
        EnchantmentSpellLock: data.UnitEnchantmentSpellLock,
        EnchantmentInvisibility: data.UnitEnchantmentInvisibility,
        EnchantmentWindWalking: data.UnitEnchantmentWindWalking,
        EnchantmentFlight: data.UnitEnchantmentFlight,
        EnchantmentResistMagic: data.UnitEnchantmentResistMagic,
        EnchantmentMagicImmunity: data.UnitEnchantmentMagicImmunity,
        EnchantmentFlameBlade: data.UnitEnchantmentFlameBlade,
        EnchantmentEldritchWeapon: data.UnitEnchantmentEldritchWeapon,
        EnchantmentTrueSight: data.UnitEnchantmentTrueSight,
        EnchantmentHolyWeapon: data.UnitEnchantmentHolyWeapon,
        EnchantmentHeroism: data.UnitEnchantmentHeroism,
        EnchantmentBless: data.UnitEnchantmentBless,
        EnchantmentLionHeart: data.UnitEnchantmentLionHeart,
        EnchantmentGiantStrength: data.UnitEnchantmentGiantStrength,
        EnchantmentPlanarTravel: data.UnitEnchantmentPlanarTravel,
        EnchantmentHolyArmor: data.UnitEnchantmentHolyArmor,
        EnchantmentRighteousness: data.UnitEnchantmentRighteousness,
        EnchantmentInvulnerability: data.UnitEnchantmentInvulnerability,
    }

    // keep track of stacks that want to move to some destination point
    stackMoves := make(map[*playerlib.UnitStack]image.Point)

    heroIndex := 0
    for _, playerHeroData := range playerData.HeroData {
        if playerHeroData.Unit > 0 {
            if playerHeroData.Unit < saveGame.NumUnits {
                heroUnitData := saveGame.Units[playerHeroData.Unit]

                // log.Printf("Player %v has hero %v %v: %+v", playerIndex, playerHeroData.Unit, playerHeroData.Name, heroUnitData)

                hero := makeHero(&player, playerHeroData, &heroUnitData, game)
                if hero.HeroType != herolib.HeroNone {
                    heroData := &saveGame.HeroData[playerIndex][heroUnitData.TypeIndex]

                    // log.Printf("  hero data: %+v", heroData)

                    setHeroData(hero, heroData)

                    for bit, enchantment := range unitEnchantmentMap {
                        if int(heroUnitData.Enchantments) & bit != 0 {
                            hero.AddEnchantment(enchantment)
                        }
                    }

                    for slot, item := range playerHeroData.Items {
                        if item > -1 && int(item) < len(artifacts) {
                            // we could in theory check the ItemSlot, but the slots are hard coded anyway
                            hero.Equipment[slot] = artifacts[item]
                            // log.Printf("  hero itemslot %v: %+v", slot, artifacts[item])
                        }
                    }

                    player.Heroes[heroIndex] = hero
                    player.AddUnit(hero)
                    heroIndex += 1
                    if heroIndex >= len(player.Heroes) {
                        break
                    }
                }
            }

        }
    }

    for unitIndex := range saveGame.NumUnits {
        unit := &saveGame.Units[unitIndex]
        if unit.Owner == int8(playerIndex) && getHeroType(unit.TypeIndex) == herolib.HeroNone {
            plane := data.PlaneArcanus
            if unit.Plane == 1 {
                plane = data.PlaneMyrror
            }

            newUnit := player.AddUnit(units.MakeOverworldUnitFromUnit(getUnitType(int(unit.TypeIndex)), int(unit.X), int(unit.Y), plane, player.GetBanner(), player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider()))

            newUnit.AddExperience(int(unit.Experience))

            switch unit.Mutations & 3 {
                case MutationMagicWeapon:
                    newUnit.SetWeaponBonus(data.WeaponMagic)
                case MutationMythrilWeapon:
                    newUnit.SetWeaponBonus(data.WeaponMythril)
                case MutationAdamantiumWeapon:
                    newUnit.SetWeaponBonus(data.WeaponAdamantium)
            }

            if unit.Mutations & MutationChaosChannelsDemonSkin != 0 {
                newUnit.AddEnchantment(data.UnitEnchantmentChaosChannelsDemonSkin)
            }

            if unit.Mutations & MutationChaosChannelsDemonWings != 0 {
                newUnit.AddEnchantment(data.UnitEnchantmentChaosChannelsDemonWings)
            }

            if unit.Mutations & MutationChaosChannelsDemonBreath != 0 {
                newUnit.AddEnchantment(data.UnitEnchantmentChaosChannelsFireBreath)
            }

            // FIXME: handle undead and stasis

            for bit, enchantment := range unitEnchantmentMap {
                if int(unit.Enchantments) & bit != 0 {
                    newUnit.AddEnchantment(enchantment)
                }
            }

            newUnit.SetMovesLeft(fraction.FromInt(int(unit.Moves) * 2))
            if unit.Finished == 1 {
                newUnit.SetMovesLeft(fraction.Zero())

                switch unit.Status {
                    case StatusPatrol: newUnit.SetBusy(units.BusyStatusPatrol)
                    case StatusBuildRoad: newUnit.SetBusy(units.BusyStatusBuildRoad)
                    case StatusPurify: newUnit.SetBusy(units.BusyStatusPurify)
                    case StatusGoto:
                        log.Printf("Unit %v going to %v,%v status %v", newUnit.GetName(), unit.DestinationX, unit.DestinationY, unit.Status)
                        stackMoves[player.FindStackByUnit(newUnit)] = image.Pt(int(unit.DestinationX), int(unit.DestinationY))
                    // FIXME: stasis
                }
            }
        }
    }

    for _, stack := range player.Stacks {
        stack.ResetActive()
    }

    player.UpdateResearchCandidates()
    player.UpdateFogVisibility()

    updateCities := func() {
        player.Cities = saveGame.convertCities(&player, playerIndex, wizards, game, game.ArcanusMap, game.MyrrorMap)
    }

    return &player, stackMoves, updateCities
}

func getUnitType(index int) units.Unit {
    switch index {
        case 0: return units.HeroBrax
        case 1: return units.HeroGunther
        case 2: return units.HeroZaldron
        case 3: return units.HeroBShan
        case 4: return units.HeroRakir
        case 5: return units.HeroValana
        case 6: return units.HeroBahgtru
        case 7: return units.HeroSerena
        case 8: return units.HeroShuri
        case 9: return units.HeroTheria
        case 10: return units.HeroGreyfairer
        case 11: return units.HeroTaki
        case 12: return units.HeroReywind
        case 13: return units.HeroMalleus
        case 14: return units.HeroTumu
        case 15: return units.HeroJaer
        case 16: return units.HeroMarcus
        case 17: return units.HeroFang
        case 18: return units.HeroMorgana
        case 19: return units.HeroAureus
        case 20: return units.HeroShinBo
        case 21: return units.HeroSpyder
        case 22: return units.HeroShalla
        case 23: return units.HeroYramrag
        case 24: return units.HeroMysticX
        case 25: return units.HeroAerie
        case 26: return units.HeroDethStryke
        case 27: return units.HeroElana
        case 28: return units.HeroRoland
        case 29: return units.HeroMortu
        case 30: return units.HeroAlorra
        case 31: return units.HeroSirHarold
        case 32: return units.HeroRavashack
        case 33: return units.HeroWarrax
        case 34: return units.HeroTorin
        case 35: return units.Trireme
        case 36: return units.Galley
        case 37: return units.Catapult
        case 38: return units.Warship
        case 39: return units.BarbarianSpearmen
        case 40: return units.BarbarianSwordsmen
        case 41: return units.BarbarianBowmen
        case 42: return units.BarbarianCavalry
        case 43: return units.BarbarianShaman
        case 44: return units.BarbarianSettlers
        case 45: return units.Berserkers
        case 46: return units.BeastmenSpearmen
        case 47: return units.BeastmenSwordsmen
        case 48: return units.BeastmenHalberdiers
        case 49: return units.BeastmenBowmen
        case 50: return units.BeastmenPriest
        case 51: return units.BeastmenMagician
        case 52: return units.BeastmenEngineer
        case 53: return units.BeastmenSettlers
        case 54: return units.Centaur
        case 55: return units.Manticore
        case 56: return units.Minotaur
        case 57: return units.DarkElfSpearmen
        case 58: return units.DarkElfSwordsmen
        case 59: return units.DarkElfHalberdiers
        case 60: return units.DarkElfCavalry
        case 61: return units.DarkElfPriests
        case 62: return units.DarkElfSettlers
        case 63: return units.Nightblades
        case 64: return units.Warlocks
        case 65: return units.Nightmares
        case 66: return units.DraconianSpearmen
        case 67: return units.DraconianSwordsmen
        case 68: return units.DraconianHalberdiers
        case 69: return units.DraconianBowmen
        case 70: return units.DraconianShaman
        case 71: return units.DraconianMagician
        // case 72: return units.DraconianEngineer
        case 73: return units.DraconianSettlers
        case 74: return units.DoomDrake
        case 75: return units.AirShip
        case 76: return units.DwarfSwordsmen
        case 77: return units.DwarfHalberdiers
        case 78: return units.DwarfEngineer
        case 79: return units.Hammerhands
        case 80: return units.SteamCannon
        case 81: return units.Golem
        case 82: return units.DwarfSettlers
        case 83: return units.GnollSpearmen
        case 84: return units.GnollSwordsmen
        case 85: return units.GnollHalberdiers
        case 86: return units.GnollBowmen
        case 87: return units.GnollSettlers
        case 88: return units.WolfRiders
        case 89: return units.HalflingSpearmen
        case 90: return units.HalflingSwordsmen
        case 91: return units.HalflingBowmen
        case 92: return units.HalflingShamans
        case 93: return units.HalflingSettlers
        case 94: return units.Slingers
        case 95: return units.HighElfSpearmen
        case 96: return units.HighElfSwordsmen
        case 97: return units.HighElfHalberdiers
        case 98: return units.HighElfCavalry
        case 99: return units.HighElfMagician
        case 100: return units.HighElfSettlers
        case 101: return units.Longbowmen
        case 102: return units.ElvenLord
        case 103: return units.Pegasai
        case 104: return units.HighMenSpearmen
        case 105: return units.HighMenSwordsmen
        case 106: return units.HighMenBowmen
        case 107: return units.HighMenCavalry
        case 108: return units.HighMenPriest
        case 109: return units.HighMenMagician
        case 110: return units.HighMenEngineer
        case 111: return units.HighMenSettlers
        case 112: return units.HighMenPikemen
        case 113: return units.Paladin
        case 114: return units.KlackonSpearmen
        case 115: return units.KlackonSwordsmen
        case 116: return units.KlackonHalberdiers
        case 117: return units.KlackonEngineer
        case 118: return units.KlackonSettlers
        case 119: return units.StagBeetle
        case 120: return units.LizardSpearmen
        case 121: return units.LizardSwordsmen
        case 122: return units.LizardHalberdiers
        case 123: return units.LizardJavelineers
        case 124: return units.LizardShamans
        case 125: return units.LizardSettlers
        case 126: return units.DragonTurtle
        case 127: return units.NomadSpearmen
        case 128: return units.NomadSwordsmen
        case 129: return units.NomadBowmen
        case 130: return units.NomadPriest
        // case 131: return units.NomadMagicians
        case 132: return units.NomadSettlers
        case 133: return units.NomadHorsebowemen
        case 134: return units.NomadPikemen
        case 135: return units.NomadRangers
        case 136: return units.Griffin
        case 137: return units.OrcSpearmen
        case 138: return units.OrcSwordsmen
        case 139: return units.OrcHalberdiers
        case 140: return units.OrcBowmen
        case 141: return units.OrcCavalry
        case 142: return units.OrcShamans
        case 143: return units.OrcMagicians
        case 144: return units.OrcEngineers
        case 145: return units.OrcSettlers
        case 146: return units.WyvernRiders
        case 147: return units.TrollSpearmen
        case 148: return units.TrollSwordsmen
        case 149: return units.TrollHalberdiers
        case 150: return units.TrollShamans
        case 151: return units.TrollSettlers
        case 152: return units.WarTrolls
        case 153: return units.WarMammoths
        case 154: return units.MagicSpirit
        case 155: return units.HellHounds
        case 156: return units.Gargoyle
        case 157: return units.FireGiant
        case 158: return units.FireElemental
        case 159: return units.ChaosSpawn
        case 160: return units.Chimeras
        case 161: return units.DoomBat
        case 162: return units.Efreet
        case 163: return units.Hydra
        case 164: return units.GreatDrake
        case 165: return units.Skeleton
        case 166: return units.Ghoul
        case 167: return units.NightStalker
        case 168: return units.WereWolf
        case 169: return units.Demon
        case 170: return units.Wraith
        case 171: return units.ShadowDemons
        case 172: return units.DeathKnights
        case 173: return units.DemonLord
        case 174: return units.Zombie
        case 175: return units.Unicorn
        case 176: return units.GuardianSpirit
        case 177: return units.Angel
        case 178: return units.ArchAngel
        case 179: return units.WarBear
        case 180: return units.Sprites
        case 181: return units.Cockatrices
        case 182: return units.Basilisk
        case 183: return units.GiantSpiders
        case 184: return units.StoneGiant
        case 185: return units.Colossus
        case 186: return units.Gorgon
        case 187: return units.EarthElemental
        case 188: return units.Behemoth
        case 189: return units.GreatWyrm
        case 190: return units.FloatingIsland
        case 191: return units.PhantomBeast
        case 192: return units.PhantomWarrior
        case 193: return units.StormGiant
        case 194: return units.AirElemental
        case 195: return units.Djinn
        case 196: return units.SkyDrake
        case 197: return units.Nagas
    }

    return units.UnitNone
}

func getHeroType(index uint8) herolib.HeroType {
    switch index {
        case 0: return herolib.HeroBrax
        case 1: return herolib.HeroGunther
        case 2: return herolib.HeroZaldron
        case 3: return herolib.HeroBShan
        case 4: return herolib.HeroRakir
        case 5: return herolib.HeroValana
        case 6: return herolib.HeroBahgtru
        case 7: return herolib.HeroSerena
        case 8: return herolib.HeroShuri
        case 9: return herolib.HeroTheria
        case 10: return herolib.HeroGreyfairer
        case 11: return herolib.HeroTaki
        case 12: return herolib.HeroReywind
        case 13: return herolib.HeroMalleus
        case 14: return herolib.HeroTumu
        case 15: return herolib.HeroJaer
        case 16: return herolib.HeroMarcus
        case 17: return herolib.HeroFang
        case 18: return herolib.HeroMorgana
        case 19: return herolib.HeroAureus
        case 20: return herolib.HeroShinBo
        case 21: return herolib.HeroSpyder
        case 22: return herolib.HeroShalla
        case 23: return herolib.HeroYramrag
        case 24: return herolib.HeroMysticX
        case 25: return herolib.HeroAerie
        case 26: return herolib.HeroDethStryke
        case 27: return herolib.HeroElana
        case 28: return herolib.HeroRoland
        case 29: return herolib.HeroMortu
        case 30: return herolib.HeroAlorra
        case 31: return herolib.HeroSirHarold
        case 32: return herolib.HeroRavashack
        case 33: return herolib.HeroWarrax
        case 34: return herolib.HeroTorin
    }

    return herolib.HeroNone
}

func makeUnit(unitData *UnitData, player *playerlib.Player) *units.OverworldUnit {
    plane := data.PlaneArcanus
    if unitData.Plane == 1 {
        plane = data.PlaneMyrror
    }
    return units.MakeOverworldUnitFromUnit(getUnitType(int(unitData.TypeIndex)), int(unitData.X), int(unitData.Y), plane, player.GetBanner(), player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider())
}

func makeHero(player *playerlib.Player, heroData PlayerHeroData, unitData *UnitData, game *gamelib.Game) *herolib.Hero {
    hero := herolib.MakeHero(makeUnit(unitData, player), getHeroType(unitData.TypeIndex), heroData.Name)
    hero.AddExperience(int(unitData.Experience))
    return hero
}

func (saveGame *SaveGame) convertArtifacts(spells spellbook.Spells) []*artifact.Artifact {
    _, typeMap, abilityMap := artifact.GetItemConversionMaps()

    artifacts := []*artifact.Artifact{}
    for _, item := range saveGame.Items {
        if item.Cost == 0 {
            // we need nil here to keep the indexes correct
            artifacts = append(artifacts, nil)
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
            Name: string(firstNonZero(item.Name)),
            Image: int(item.IconIndex),
            Type: typeMap[item.Type],
            Cost: int(item.Cost),
            Powers: powers,
        })
    }

    return artifacts
}

func setupRelations(player *playerlib.Player, index int, playerData *PlayerData, allPlayers []*playerlib.Player) {
    const (
        NoTreaty int8 = 0
        Pact int8 = 1
        Alliance int8 = 2
        War int8 = 3
        FinalWar int8 = 4
    )

    convertTreatyType := func(treaty int8) data.TreatyType {
        switch treaty {
            case NoTreaty: return data.TreatyNone
            case Pact: return data.TreatyPact
            case Alliance: return data.TreatyAlliance
            case War, FinalWar: return data.TreatyWar
        }

        return data.TreatyNone
    }

    for contactIndex := range playerData.Diplomacy.Contacted {
        if contactIndex != index && contactIndex < len(allPlayers) {
            otherPlayer := allPlayers[contactIndex]
            player.AwarePlayer(otherPlayer)
            relation, ok := player.GetDiplomaticRelation(otherPlayer)
            if ok {

                relation.TreatyInterest = int(playerData.Diplomacy.TreatyInterest[contactIndex])
                relation.TradeInterest = int(playerData.Diplomacy.TradeInterest[contactIndex])
                relation.PeaceInterest = int(playerData.Diplomacy.PeaceInterest[contactIndex])
                relation.StartingRelation = int(playerData.Diplomacy.DefaultRelations[contactIndex])
                relation.VisibleRelation = int(playerData.Diplomacy.VisibleRelations[contactIndex])
                relation.Treaty = convertTreatyType(playerData.Diplomacy.DiplomacyStatus[contactIndex])

                /*
                Treaty data.TreatyType
                // from -100 to +100, where -100 means this player hates the other, and +100 means this player loves the other
                StartingRelation int
                VisibleRelation int

                // these values indicate how likely the player will be to accept a treaty or trade
                TreatyInterest int
                TradeInterest int
                PeaceInterest int

                // how many turns since a peace treaty has been made with this player. While this value is above 0
                // the owner will not perform any hostile actions
                PeaceCounter int

                // how hostile this player is to the other
                Hostility Hostility

                // how many times this player has been warned about their actions
                WarningCounter int
                */

            }
        }
    }
}

func (saveGame *SaveGame) Convert(cache *lbx.LbxCache) *gamelib.Game {
    game := gamelib.MakeGame(cache, saveGame.convertSettings())
    game.TurnNumber = uint64(saveGame.Turn)

    artifacts := saveGame.convertArtifacts(game.AllSpells())
    // artifacts that are in the game are removed from the pool
    // because the ones in the pool are the ones not found yet.
    // an artifact that is not in the pool must have been created
    // by a player via the 'enchant item' or 'create artifact' spells.
    for _, artifact := range artifacts {
        if artifact == nil {
            continue
        }

        _, ok := game.ArtifactPool[artifact.Name]
        if ok {
            delete(game.ArtifactPool, artifact.Name)
        }
    }

    /*
    log.Printf("Units:")
    for i, unit := range saveGame.Units {
        if unit.HeroSlot > 0 {
            log.Printf("%v: hero %+v", i, unit)
        }
    }
    */

    wizards := []setup.WizardCustom{}
    for playerIndex := range saveGame.NumPlayers {
        if int(playerIndex) < len(saveGame.PlayerData) {
            wizards = append(wizards, saveGame.convertWizard(int(playerIndex)))
        }
    }

    /*
    for player, heros := range saveGame.HeroData {
        for i, heroData := range heros {
            if getHeroType(uint8(i)) == herolib.HeroDethStryke {
                log.Printf("Player %v hero %d: %+v", player, i, heroData)
            }
        }
    }
    */

    var playerDefers []func()

    for playerIndex := range saveGame.NumPlayers {
        player, stackMoves, deferred := saveGame.convertPlayer(int(playerIndex), wizards, artifacts, game)
        game.Players = append(game.Players, player)

        playerDefers = append(playerDefers, deferred)

        defer func(){
            for stack, destination := range stackMoves {
                // FIXME: associate the player with the stack
                path := game.FindPath(stack.X(), stack.Y(), destination.X, destination.Y, player, stack, player.GetFog(stack.Plane()))
                if path != nil {
                    stack.CurrentPath = path
                }
            }
        }()
    }

    // now set up player relations
    for playerIndex := range saveGame.NumPlayers {
        data := &saveGame.PlayerData[playerIndex]
        player := game.Players[playerIndex]
        setupRelations(player, int(playerIndex), data, game.Players)
    }

    // the players must exist before we can convert the maps
    game.ArcanusMap = saveGame.ConvertMap(game.ArcanusMap.Data, data.PlaneArcanus, game, game.Players)
    game.MyrrorMap = saveGame.ConvertMap(game.MyrrorMap.Data, data.PlaneMyrror, game, game.Players)

    // any initialization that needs the maps to occur can now run
    for _, f := range playerDefers {
        f()
    }


    // FIXME: add neutral player with brown banner and ai.MakeRaiderAI()

    // FIXME: add all remaining information from saveGame
    
    // saveGame.GrandVizier
    // saveGame.Events

    // FIXME: game.RandomEvents
    // FIXME: game.RoadWorkArcanus
    // FIXME: game.RoadWorkMyrror
    // FIXME: game.PurifyWorkArcanus
    // FIXME: game.PurifyWorkMyrror

    game.Camera.Center(20, 20)
    if len(game.Players[0].Cities) > 0 {
        city := game.Players[0].Cities[0]
        game.Events <- &gamelib.GameEventMoveCamera{
            Instant: true,
            Plane: city.Plane,
            X: city.X,
            Y: city.Y,
        }
    }

    return game
}
