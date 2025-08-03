package load

import (
    // "fmt"
    gamelib "github.com/kazzmir/master-of-magic/game/magic/game"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
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

    out.HeroData = make([][]HeroData, out.NumPlayers)
    for i, player := range game.Players {
        if player == nil {
            continue
        }
        if player.IsNeutral() {
            continue
        }
        out.HeroData[i] = makeHeroData(player)
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


func makeHeroData(player *playerlib.Player) []HeroData {
    var out []HeroData

    for _, hero := range player.Heroes {
        if hero == nil {
            continue
        }

        data := HeroData{
            // Level: int16(hero.GetHeroExperienceLevel().),
        }

        /*
        Abilities uint32
        AbilitySet *set.Set[HeroAbility]
        CastingSkill int8
        Spells [4]uint8
        ExtraByte byte // unknown value
        */

        out = append(out, data)
    }

    return out
}
