package load

import (
    // "fmt"
    "maps"
    "slices"

    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/lib/fraction"
)

func CreateSaveGame(game *gamelib.Game) (*SaveGame, error) {

    var out SaveGame

    for _, player := range game.Model.Players {
        if player == nil {
            continue
        }
        if player.IsNeutral() {
            continue
        }
        out.NumPlayers += 1
    }

    out.LandSize = int16(game.Settings.LandSize)
    out.Magic = int16(game.Settings.Magic)
    out.Difficulty = int16(game.Settings.Difficulty)
    out.NumCities = int16(len(game.AllCities()))
    out.NumUnits = int16(len(game.AllUnits()))
    out.Turn = int16(game.TurnNumber)

    // FIXME
    // out.Unit = 0

    allSpells := game.AllSpells()

    out.HeroData = make([][]HeroData, out.NumPlayers)
    for i, player := range game.Model.Players {
        if player == nil {
            continue
        }
        if player.IsNeutral() {
            continue
        }
        out.HeroData[i] = makeHeroData(player, &allSpells)
    }

    for i, player := range game.Model.Players {
        out.PlayerData = append(out.PlayerData, makePlayerData(i, game, player))
    }


    /*
struct {
    // indexed by player number
    PlayerData []PlayerData

    ArcanusMap TerrainData
    MyrrorMap TerrainData

    GrandVizier uint16

    UU_table_1 []byte
    UU_table_2 []byte

    ArcanusLandMasses [][]uint8
    MyrrorLandMasses [][]uint8

    Nodes []NodeData
    Fortresses []FortressData
    Towers []TowerData
    Lairs []LairData
    Items []ItemData
    Cities []CityData
    Units []UnitData

    ArcanusTerrainSpecials [][]uint8
    MyrrorTerrainSpecials [][]uint8

    ArcanusExplored [][]int8
    MyrrorExplored [][]int8

    ArcanusMovementCost MovementCostData
    MyrrorMovementCost MovementCostData

    Events EventData

    ArcanusMapSquareFlags [][]uint8
    MyrrorMapSquareFlags [][]uint8

    PremadeItems []byte

    HeroNames []HeroNameData
}
*/

    return &out, nil
}

func makePlayerData(id int, game *gamelib.Game, player *playerlib.Player) PlayerData {

    summoningCity := player.FindSummoningCity()
    var summonX int16
    var summonY int16
    var summonPlane int16
    if summoningCity != nil {
        summonX = int16(summoningCity.X)
        summonY = int16(summoningCity.Y)
        if summoningCity.Plane == data.PlaneMyrror {
            summonPlane = 1 // Myrror
        } else {
            summonPlane = 0 // Arcanus
        }
    }

    createSpellRanks := func(player *playerlib.Player) []int16 {
        ranks := make([]int16, 5)
        for i, magic := range []data.MagicType{data.NatureMagic, data.SorceryMagic, data.ChaosMagic, data.LifeMagic, data.DeathMagic} {
            ranks[i] = int16(player.Wizard.MagicLevel(magic))
        }
        return ranks
    }

    boolToInt := func(b bool) int8 {
        if b {
            return 1
        }
        return 0
    }

    return PlayerData{
        WizardId: uint8(id),
        WizardName: []byte(player.Wizard.Name),
        CapitalRace: uint8(toRaceInt(player.Wizard.Race)),
        BannerId: uint8(player.Wizard.Banner),
        // FIXME
        // Personality: uint16(player.Wizard.Personality),
        // Objective: 0
        // VolcanoPower: 0,
        // AverageUnitCost: 0,
        MasteryResearch: uint16(player.SpellOfMasteryCost),
        Fame: uint16(player.Fame),
        PowerBase: uint16(game.ComputePower(player)),
        Volcanoes: uint16(len(game.Model.ArcanusMap.GetCastedVolcanoes(player)) + len(game.Model.MyrrorMap.GetCastedVolcanoes(player))),
        ResearchRatio: uint8(player.PowerDistribution.Research * 100),
        ManaRatio: uint8(player.PowerDistribution.Mana * 100),
        SkillRatio: uint8(player.PowerDistribution.Skill * 100),
        SummonX: summonX,
        SummonY: summonY,
        SummonPlane: summonPlane,
        ResearchSpells: mapSlice(func(spell spellbook.Spell) uint16 {
            if !spell.Valid() {
                return 0
            }
            return uint16(spell.Index)
        }, player.ResearchCandidateSpells.Spells...),
        CombatSkillLeft: uint16(player.RemainingCastingSkill),
        CastingCostRemaining: uint16(max(0, player.ComputeEffectiveSpellCost(player.CastingSpell, true) - player.CastingSpellProgress)),
        CastingCostOriginal: uint16(player.ComputeEffectiveSpellCost(player.CastingSpell, true)),
        CastingSpellIndex: uint16(player.CastingSpell.Index),
        NominalSkill: uint16(player.ComputeCastingSkill()),
        SkillLeft: uint16(player.RemainingCastingSkill),
        TaxRate: uint16(player.TaxRate.Multiply(fraction.FromInt(2)).ToInt()),
        SpellRanks: createSpellRanks(player),
        RetortAlchemy: boolToInt(player.Wizard.RetortEnabled(data.RetortAlchemy)),
        RetortWarlord: boolToInt(player.Wizard.RetortEnabled(data.RetortWarlord)),
        RetortChaosMastery: boolToInt(player.Wizard.RetortEnabled(data.RetortChaosMastery)),
        RetortNatureMastery: boolToInt(player.Wizard.RetortEnabled(data.RetortNatureMastery)),
        RetortSorceryMastery: boolToInt(player.Wizard.RetortEnabled(data.RetortSorceryMastery)),
        RetortInfernalPower: boolToInt(player.Wizard.RetortEnabled(data.RetortInfernalPower)),
        RetortDivinePower: boolToInt(player.Wizard.RetortEnabled(data.RetortDivinePower)),
        RetortSageMaster: boolToInt(player.Wizard.RetortEnabled(data.RetortSageMaster)),
        RetortChanneler: boolToInt(player.Wizard.RetortEnabled(data.RetortChanneler)),
        RetortMyrran: boolToInt(player.Wizard.RetortEnabled(data.RetortMyrran)),
        RetortArchmage: boolToInt(player.Wizard.RetortEnabled(data.RetortArchmage)),
        RetortNodeMastery: boolToInt(player.Wizard.RetortEnabled(data.RetortNodeMastery)),
        RetortManaFocusing: boolToInt(player.Wizard.RetortEnabled(data.RetortManaFocusing)),
        RetortFamous: boolToInt(player.Wizard.RetortEnabled(data.RetortFamous)),
        RetortRunemaster: boolToInt(player.Wizard.RetortEnabled(data.RetortRunemaster)),
        RetortConjurer: boolToInt(player.Wizard.RetortEnabled(data.RetortConjurer)),
        RetortCharismatic: boolToInt(player.Wizard.RetortEnabled(data.RetortCharismatic)),
        RetortArtificer: boolToInt(player.Wizard.RetortEnabled(data.RetortArtificer)),
        // HeroData: []PlayerHeroData
    }

    /*
type PlayerData struct {
    VaultItems []int16
    Diplomacy DiplomacyData

    ResearchCostRemaining uint16
    ManaReserve uint16
    SpellCastingSkill int32
    ResearchingSpellIndex int16
    SpellsList []uint8
    DefeatedWizards uint16
    GoldReserve uint16
    Astrology AstrologyData
    Population uint16
    Historian []uint8
    GlobalEnchantments []uint8
    MagicStrategy uint16
    Hostility []uint16
    ReevaluateHostilityCountdown uint16
    ReevaluateMagicStrategyCountdown uint16
    ReevaluateMagicPowerCountdown uint16
    PeaceDuration []uint8
    TargetWizard uint16
    PrimaryRealm uint16
    SecondaryRealm uint16

    Unknown1 uint8
    Unknown2 []uint8
    Unknown3 []uint8
    Unknown4 uint16
    Unknown5 int16
    Unknown6 uint16
    Unknown7 uint16
    Unknown8 uint8
    Unknown9 uint8
    Unknown10 uint16
    Unknown11 uint8
    Unknown12 uint8
    Unknown13 []uint8
}
*/
}

func makeAbilityMap() map[data.AbilityType]HeroAbility {
    return map[data.AbilityType]HeroAbility{
        data.AbilityLeadership: HeroAbility_LEADERSHIP,
        data.AbilitySuperLeadership: HeroAbility_LEADERSHIP2,
        data.AbilityLegendary: HeroAbility_LEGENDARY,
        data.AbilitySuperLegendary: HeroAbility_LEGENDARY2,
        data.AbilityBlademaster: HeroAbility_BLADEMASTER,
        data.AbilitySuperBlademaster: HeroAbility_BLADEMASTER2,
        data.AbilityArmsmaster: HeroAbility_ARMSMASTER,
        data.AbilitySuperArmsmaster: HeroAbility_ARMSMASTER2,
        data.AbilityConstitution: HeroAbility_CONSTITUTION,
        data.AbilitySuperConstitution: HeroAbility_CONSTITUTION2,
        data.AbilityMight: HeroAbility_MIGHT,
        data.AbilitySuperMight: HeroAbility_MIGHT2,
        data.AbilityArcanePower: HeroAbility_ARCANE_POWER,
        data.AbilitySuperArcanePower: HeroAbility_ARCANE_POWER2,
        data.AbilitySage: HeroAbility_SAGE,
        data.AbilitySuperSage: HeroAbility_SAGE2,
        data.AbilityPrayermaster: HeroAbility_PRAYERMASTER,
        data.AbilitySuperPrayermaster: HeroAbility_PRAYERMASTER2,
        data.AbilityAgility: HeroAbility_AGILITY,
        data.AbilitySuperAgility: HeroAbility_AGILITY2,
        data.AbilityLucky: HeroAbility_LUCKY,
        data.AbilityCharmed: HeroAbility_CHARMED,
        data.AbilityNoble: HeroAbility_NOBLE,
        // FIXME
        // data.AbilityFemale: HeroAbility_FEMALE,
    }
}

func convertAbility(ability data.Ability) HeroAbility {
    all := makeAbilityMap()
    value, ok := all[ability.Ability]
    if ok {
        return value
    }

    return HeroAbility_NONE
}

func makeAbilityBits(abilities map[data.AbilityType]data.Ability) uint32 {
    all := makeAbilityMap()

    var out uint32

    for ability := range abilities {
        bits, ok := all[ability]
        if ok {
            out |= uint32(bits)
        }
    }

    return out
}

func mapSlice[T any, U any](fn func(T) U, slice ...T) []U {
    out := make([]U, len(slice))
    for i, v := range slice {
        out[i] = fn(v)
    }
    return out
}

func makeHeroData(player *playerlib.Player, allSpells *spellbook.Spells) []HeroData {
    var out []HeroData

    for _, hero := range player.Heroes {
        if hero == nil {
            continue
        }

        level := hero.GetHeroExperienceLevel()
        casterValue := float32(hero.GetAbilityValue(data.AbilityCaster)) / float32(level.ToInt())
        var castingSkill int8
        // FIXME: the caster value will end up being something like 5, 7.5, which should be converted into 1, 2, 3, etc
        castingSkill = int8(casterValue)

        var spells [4]uint8
        for i, spellName := range hero.GetKnownSpells() {
            spell := allSpells.FindByName(spellName)
            if spell.Valid() {
                spells[i] = uint8(spell.Index)
            }
        }

        data := HeroData{
            Level: int16(hero.GetHeroExperienceLevel()),
            Abilities: makeAbilityBits(hero.Abilities),
            AbilitySet: set.NewSet(mapSlice(convertAbility, slices.Collect(maps.Values(hero.Abilities))...)...),
            CastingSkill: castingSkill,
            Spells: spells,
        }

        out = append(out, data)
    }

    return out
}
