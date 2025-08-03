package load

import (
    // "fmt"

    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/lib/set"
)

func CreateSaveGame(game *gamelib.Game) (*SaveGame, error) {

    var out SaveGame

    for _, player := range game.Players {
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
    for i, player := range game.Players {
        if player == nil {
            continue
        }
        if player.IsNeutral() {
            continue
        }
        out.HeroData[i] = makeHeroData(player, &allSpells)
    }

    /*
struct {
    // first index player, second index hero
    HeroData [][]HeroData

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

func makeAbilityBits(abilities []data.Ability) uint32 {
    all := makeAbilityMap()

    var out uint32

    for _, ability := range abilities {
        bits, ok := all[ability.Ability]
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

        caster := hero.GetAbilityReference(data.AbilityCaster)
        var castingSkill int8
        if caster != nil {
            castingSkill = int8(caster.Value)
        }

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
            AbilitySet: set.NewSet(mapSlice(convertAbility, hero.Abilities...)...),
            CastingSkill: castingSkill,
            Spells: spells,
        }

        out = append(out, data)
    }

    return out
}
