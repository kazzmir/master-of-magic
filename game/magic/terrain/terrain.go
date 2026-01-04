package terrain

import (
    "bytes"
    "image"
    "fmt"
    "slices"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/set"
)

// terrain tiles are indicies 0-0x259 for arcanus, and 0x2fA - 0x5f4 for myrror

const MyrrorStart = 0x2FA

type TerrainIndex int

const (
    IndexOcean1 TerrainIndex = 0x000
    IndexBugGrass            = 0x001
    IndexShore1_1st          = 0x002
    IndexLake                = 0x012
    IndexShore1_end          = 0x0A1
    IndexGrass1              = 0x0A2
    IndexForest1             = 0x0A3
    IndexMountain1           = 0x0A4
    IndexDesert1             = 0x0A5
    IndexSwamp1              = 0x0A6
    IndexTundra1             = 0x0A7
    IndexSorcNode            = 0x0A8
    IndexNatNode             = 0x0A9
    IndexChaosNode           = 0x0AA
    IndexHills1              = 0x0AB
    IndexGrass2              = 0x0AC
    IndexGrass3              = 0x0AD
    IndexDesert2             = 0x0AE
    IndexDesert3             = 0x0AF
    IndexDesert4             = 0x0B0
    IndexSwamp2              = 0x0B1
    IndexSwamp3              = 0x0B2
    IndexVolcano             = 0x0B3
    IndexGrass4              = 0x0B4
    IndexTundra2             = 0x0B5
    IndexTundra3             = 0x0B6
    IndexForest2             = 0x0B7
    IndexForest3             = 0x0B8
    IndexRiverMStart         = 0x0B9
    IndexRiverMEnd           = 0x0C4
    IndexLake1               = 0x0C5
    IndexLake2               = 0x0C6
    IndexLake3               = 0x0C7
    IndexLake4               = 0x0C8
    IndexShore2FStart        = 0x0C9
    IndexShore2FEnd          = 0x0E8
    IndexRiversStart         = 0x0E9
    IndexRiversEnd           = 0x102
    IndexMountainsStart      = 0x103
    IndexMountainsEnd        = 0x112
    IndexHillsStart          = 0x113
    IndexHillsEnd            = 0x123
    IndexDesertStart         = 0x124
    IndexDesertEnd           = 0x1C3
    IndexShore2Start         = 0x1C4
    IndexShore2End           = 0x1D3
    Index4WRiver1            = 0x1D4
    Index4WRiver2            = 0x1D5
    Index4WRiver3            = 0x1D6
    Index4WRiver4            = 0x1D7
    Index4WRiver5            = 0x1D8
    IndexShore3Start         = 0x1D9
    IndexShore3End           = 0x258
    IndexOcean2              = 0x259
    IndexTundra_1st          = 0x25A
    IndexTundra_Last         = 0x2F9
    IndexTundra              = 0x26A
)

type TerrainType int

const (
    // 0 value is unknown
    Unknown TerrainType = iota
    Ocean
    River
    Shore
    Mountain
    Hill
    Grass
    Swamp
    Forest
    Desert
    Tundra
    Volcano
    Lake
    NatureNode
    SorceryNode
    ChaosNode
)

func (terrain TerrainType) String() string {
    switch terrain {
        case Ocean: return "Ocean"
        case River: return "River"
        case Shore: return "Shore"
        case Mountain: return "Mountain"
        case Hill: return "Hill"
        case Grass: return "Grass"
        case Forest: return "Forest"
        case Swamp: return "Swamp"
        case Desert: return "Desert"
        case Tundra: return "Tundra"
        case Volcano: return "Volcano"
        case Lake: return "Lake"
        case NatureNode: return "Nature Node"
        case SorceryNode: return "Sorcery Node"
        case ChaosNode: return "Chaos Node"
        case Unknown: return "Unknown"
        default: return "Error"
    }
}

type Direction int

const (
    NorthWest Direction = iota
    North
    NorthEast
    East
    SouthEast
    South
    SouthWest
    West
    Center
)

func (direction Direction) String() string {
    switch direction {
        case NorthWest: return "NorthWest"
        case North: return "North"
        case NorthEast: return "NorthEast"
        case East: return "East"
        case SouthEast: return "SouthEast"
        case South: return "South"
        case SouthWest: return "SouthWest"
        case West: return "West"
        case Center: return "Center"
        default: return "Unknown"
    }
}

// assume dx, dy are in {-1, 0, 1}
func ToDirection(dx int, dy int) Direction {
    switch {
        case dx == -1 && dy == -1: return NorthWest
        case dx == 0 && dy == -1: return North
        case dx == 1 && dy == -1: return NorthEast
        case dx == 1 && dy == 0: return East
        case dx == 1 && dy == 1: return SouthEast
        case dx == 0 && dy == 1: return South
        case dx == -1 && dy == 1: return SouthWest
        case dx == -1 && dy == 0: return West
        case dx == 0 && dy == 0: return Center
    }

    panic("invalid direction")
}

type CompatibilityType int

const (
    AnyOf CompatibilityType = iota
    NoneOf
)

type DirectedCompatibility struct {
    Direction Direction
    // for O(1) lookups
    Terrains *set.Set[TerrainType]
    Type CompatibilityType
}

func (compatibility DirectedCompatibility) String() string {
    if compatibility.Type == AnyOf {
        return fmt.Sprintf("(%s, %s)", compatibility.Direction, slices.Sorted(slices.Values(compatibility.Terrains.Values())))
    } else {
        return fmt.Sprintf("!(%s, %s)", compatibility.Direction, slices.Sorted(slices.Values(compatibility.Terrains.Values())))
    }
}

type Compatibility struct {
    Terrains *set.Set[TerrainType]
    Type CompatibilityType
}

func (compatibility Compatibility) String() string {
    if compatibility.Type == AnyOf {
        return fmt.Sprintf("%s", slices.Sorted(slices.Values(compatibility.Terrains.Values())))
    } else {
        return fmt.Sprintf("!%s", slices.Sorted(slices.Values(compatibility.Terrains.Values())))
    }
}


type Tile struct {
    // index into the TerrainTile array
    index int
    Compatibilities map[Direction]Compatibility
}

func InvalidTile() Tile {
    return Tile{
        index: -1,
        Compatibilities: make(map[Direction]Compatibility),
    }
}

func (tile Tile) Valid() bool {
    return tile.index != -1
}

func (tile Tile) Index(plane data.Plane) int {
    switch plane {
        case data.PlaneArcanus: return tile.index
        case data.PlaneMyrror: return tile.index + MyrrorStart
    }

    return tile.index
}

func (tile Tile) IsMagic() bool {
    switch TerrainIndex(tile.index) {
        case IndexSorcNode, IndexNatNode, IndexChaosNode: return true
        default: return false
    }
}

func (tile Tile) TerrainType() TerrainType {
    switch TerrainIndex(tile.index) {
        case IndexOcean1, IndexOcean2: return Ocean
        case IndexBugGrass, IndexGrass1, IndexGrass2, IndexGrass3, IndexGrass4: return Grass
        case IndexForest1, IndexForest2, IndexForest3: return Forest
        case IndexMountain1: return Mountain
        case IndexDesert1, IndexDesert2, IndexDesert3, IndexDesert4: return Desert
        case IndexSwamp1, IndexSwamp2, IndexSwamp3: return Swamp
        case IndexTundra, IndexTundra1, IndexTundra2, IndexTundra3: return Tundra
        case IndexSorcNode: return SorceryNode
        case IndexNatNode: return NatureNode
        case IndexChaosNode: return ChaosNode
        case IndexHills1: return Hill
        case IndexVolcano: return Volcano
        case IndexLake, IndexLake1, IndexLake2, IndexLake3, IndexLake4: return Lake
    }

    if tile.index >= IndexRiverMStart && tile.index <= IndexRiverMEnd {
        return River
    }

    if tile.index >= IndexRiversStart && tile.index <= IndexRiversEnd {
        return River
    }

    if tile.index >= Index4WRiver1 && tile.index <= Index4WRiver5 {
        return River
    }

    if tile.index >= IndexShore1_1st && tile.index <= IndexShore1_end {
        return Shore
    }

    if tile.index >= IndexShore2FStart && tile.index <= IndexShore2FEnd {
        return Shore
    }

    if tile.index >= IndexShore2Start && tile.index <= IndexShore2End {
        return Shore
    }

    if tile.index >= IndexShore3Start && tile.index <= IndexShore3End {
        return Shore
    }

    if tile.index >= IndexMountainsStart && tile.index <= IndexMountainsEnd {
        return Mountain
    }

    if tile.index >= IndexHillsStart && tile.index <= IndexHillsEnd {
        return Hill
    }

    if tile.index >= IndexDesertStart && tile.index <= IndexDesertEnd {
        return Desert
    }

    if tile.index >= IndexTundra_1st && tile.index <= IndexTundra_Last {
        return Tundra
    }

    return Unknown
}

func (tile Tile) Name() string {
    switch tile.TerrainType() {
        case Ocean: return "Ocean"
        case Grass: return "Grasslands"
        case Forest: return "Forest"
        case Mountain: return "Mountain"
        case Desert: return "Desert"
        case Swamp: return "Swamp"
        case Tundra: return "Tundra"
        case SorceryNode: return "Sorcery Node"
        case NatureNode: return "Nature Node"
        case ChaosNode: return "Chaos Node"
        case Hill: return "Hills"
        case Volcano: return "Volcano"
        case Lake: return "Lake"
        case River: return "River"
        case Shore: return "Shore"
        default: return "Unknown"
    }
}

func (tile Tile) IsShore() bool {
    return tile.TerrainType() == Shore
}

func (tile Tile) IsLand() bool {
    switch tile.TerrainType() {
        case Ocean, Shore, Lake: return false
        case River, Mountain,
             Hill, Grass, Swamp, Forest,
             Desert, Tundra, Volcano,
             NatureNode, SorceryNode, ChaosNode: return true
    }

    return false
}

func (tile Tile) IsWater() bool {
    switch tile.TerrainType() {
        case Ocean, Shore, Lake: return true
        default: return false
    }
}

func (tile Tile) IsLakeWithFlow() bool {
    switch tile.index {
        case IndexLake1, IndexLake2, IndexLake3, IndexLake4: return true
    }

    return false
}

func (tile Tile) IsRiver() bool {
    switch tile.TerrainType() {
        case River: return true
        default: return false
    }
}

// match the given match to a terrain
// terrain can contain more compatibilites than what the match has
// example: match: {east: anyof(ocean)}, tile terrain: ocean -> matches
// match: {east: anyof(ocean), south: anyof(shore)}, tile terrain: shore with land on south -> matches
func (tile *Tile) Matches(match map[Direction]TerrainType) bool {
    for direction, compatibility := range tile.Compatibilities {
        // only consider directions that are specified in the match
        _, ok := match[direction]
        if ok {
            if compatibility.Type == AnyOf {
                if !compatibility.Terrains.Contains(match[direction]) {
                    return false
                }

                /*
                isAny := false
                for _, terrain := range compatibility.Terrains {
                    if match[direction] == terrain {
                        isAny = true
                        break
                    }
                }
                if !isAny {
                    return false
                }
                */
            } else {
                if compatibility.Terrains.Contains(match[direction]) {
                    return false
                }

                /*
                none := true
                for _, terrain := range compatibility.Terrains {
                    if match[direction] == terrain {
                        none = false
                        break
                    }
                }
                if !none {
                    return false
                }
                */
            }
        }
    }
    return true
}

func (tile *Tile) GetDirection(direction Direction) Compatibility {
    compatibility, ok := tile.Compatibilities[direction]
    if ok {
        return compatibility
    }

    return Compatibility{Terrains: set.NewSet([]TerrainType{Unknown}...)}
}

func (tile Tile) String() string {
    return fmt.Sprintf("Tile{Index: %d, Compatibilities: %v}", tile.index, tile.Compatibilities)
}

var allTiles []Tile

/* create directions from 8-bit pattern */
func makeDirections(bitPattern uint8) []Direction {
    var directions []Direction

    choices := []Direction{West, SouthWest, South, SouthEast, East, NorthEast, North, NorthWest}
    for i, choice := range choices {
        if bitPattern & (1 << i) != 0 {
            directions = append(directions, choice)
        }
    }

    return directions
}

func makeCompatibilities(directions []Direction, terrains []TerrainType, type_ CompatibilityType) []DirectedCompatibility {
    var out []DirectedCompatibility

    for _, direction := range directions {
        out = append(out, DirectedCompatibility{
            Direction: direction,
            Terrains: set.NewSet(terrains...),
            Type: type_,
        })
    }

    return out
}

func makeSimpleTile(index int, terrain TerrainType) Tile {
    return makeTile(terrain, index, nil)
}

func makeTile(terrain TerrainType, index int, compatibilities []DirectedCompatibility) Tile {
    compatibilities = append(makeCompatibilities([]Direction{Center}, []TerrainType{terrain}, AnyOf), compatibilities...)

    all := make(map[Direction]Compatibility)

    for _, compatibility := range compatibilities {
        all[compatibility.Direction] = Compatibility{Terrains: compatibility.Terrains, Type: compatibility.Type}
    }

    tile := Tile{
        index: index,
        Compatibilities: all,
    }

    allTiles = append(allTiles, tile)

    return tile
}

// expand from 4 cardinal directions to 8 directions
// 0001 -> 0000 0001
// 0010 -> 0000 0100
// 0100 -> 0001 0000
// 1000 -> 0100 0000
func expand4(value uint8) uint8 {
    v1 := value & 1
    v2 := (value & 2) << 1
    v3 := (value & 4) << 2
    v4 := (value & 8) << 3

    return v1 | v2 | v3 | v4
}

func GetTile(index int) Tile {
    if index >= MyrrorStart {
        index -= MyrrorStart
    }

    if index >= len(allTiles) {
        return Tile{}
    }

    return allTiles[index]
}

// pattern is type in 4-bit cardindal directions, 0's are not type
func makeCardinalTile(terrain TerrainType, index int, bitPattern uint8) Tile {
    terrains := []TerrainType{terrain}
    comp := makeCompatibilities(makeDirections(expand4(bitPattern)), terrains, AnyOf)
    incomp := makeCompatibilities(makeDirections(expand4(^bitPattern)), terrains, NoneOf)
    return makeTile(terrain, index, append(comp, incomp...))
}

// pattern is type not type, 0's are not type; ignores the corners if the adjacent are set
func makeCornerIgnoringTile(terrain TerrainType, index int, bitPattern uint8) Tile {
    var mask uint8
    mask = 0b11111111
    switch {
        case bitPattern & 0b01010000 == 0b01010000: mask &= ^(bitPattern & 0b00100000); fallthrough
        case bitPattern & 0b01000001 == 0b01000001: mask &= ^(bitPattern & 0b10000000); fallthrough
        case bitPattern & 0b00010100 == 0b00010100: mask &= ^(bitPattern & 0b00001000); fallthrough
        case bitPattern & 0b00000101 == 0b00000101: mask &= ^(bitPattern & 0b00000010)
    }

    terrains := []TerrainType{terrain}
    comp := makeCompatibilities(makeDirections(^bitPattern), terrains, AnyOf)
    incomp := makeCompatibilities(makeDirections(bitPattern & mask), terrains, NoneOf)
    return makeTile(terrain, index, append(comp, incomp...))
}

// pattern is not ocean/shore, 0's are ocean/shore; ignores the corners if the adjacent are set; also excludes rivers in cardinal directions
func makeShoreTile(index int, bitPattern uint8) Tile {
    var mask uint8
    mask = 0b11111111
    switch {
        case bitPattern & 0b01010000 == 0b01010000: mask &= ^(bitPattern & 0b00100000); fallthrough
        case bitPattern & 0b01000001 == 0b01000001: mask &= ^(bitPattern & 0b10000000); fallthrough
        case bitPattern & 0b00010100 == 0b00010100: mask &= ^(bitPattern & 0b00001000); fallthrough
        case bitPattern & 0b00000101 == 0b00000101: mask &= ^(bitPattern & 0b00000010)
    }
    comp := makeCompatibilities(makeDirections(^bitPattern), []TerrainType{Ocean, Shore}, AnyOf)
    incomp1 := makeCompatibilities(makeDirections(bitPattern & mask & 0b10101010), []TerrainType{Ocean, Shore}, NoneOf)
    incomp2 := makeCompatibilities(makeDirections(bitPattern & mask & 0b01010101), []TerrainType{Ocean, Shore, River}, NoneOf)
    return makeTile(Shore, index, append(comp, append(incomp1, incomp2...)...))
}

// pattern is not desert, 0's are desert
func makeDesertTile(index int, bitPattern uint8) Tile {
    return makeCornerIgnoringTile(Desert, index, bitPattern)
}

// pattern is not tundra, 0's are tundra
func makeTundraTile(index int, bitPattern uint8) Tile {
    return makeCornerIgnoringTile(Tundra, index, bitPattern)
}

// center is lake, cardinal directions are not river
func makeLakeTile(index int) Tile {
    var pattern uint8 = 0b1111
    return makeTile(Lake, index, makeCompatibilities(makeDirections(expand4(pattern)), []TerrainType{River}, NoneOf))
}

// pattern is 4-bit cardinal directions with river/lake/ocean, 0s are not river/lake/ocean
func makeRiverTile(index int, bitPattern uint8) Tile {
    terrains := []TerrainType{River, Lake, Ocean, Shore}
    comp := makeCompatibilities(makeDirections(expand4(bitPattern)), terrains, AnyOf)
    incomp := makeCompatibilities(makeDirections(expand4(^bitPattern)), terrains, NoneOf)
    return makeTile(River, index, append(comp, incomp...))
}

// center is lake, pattern is 4-bit cardinal directions with river/lake/ocean, 0s are not river/lake/ocean
func makeLakeRiverTile(index int, bitPattern uint8) Tile {
    comp := makeCompatibilities(makeDirections(expand4(bitPattern)), []TerrainType{River}, AnyOf)
    incomp := makeCompatibilities(makeDirections(expand4(^bitPattern)), []TerrainType{River}, NoneOf)
    return makeTile(Lake, index, append(comp, incomp...))
}

// pattern is 4-bit cardinal directions with hills, 0s are not mountain
func makeHillTile(index int, bitPattern uint8) Tile {
    return makeCardinalTile(Hill, index, bitPattern)
}

// pattern is 4-bit cardinal directions with mountain, 0s are not mountain
func makeMountainTile(index int, bitPattern uint8) Tile {
    return makeCardinalTile(Mountain, index, bitPattern)
}

func makeShoreRiverTile(index int, landPattern uint8, riverPattern uint8) Tile {
    nonOceanPattern := landPattern | riverPattern
    var mask uint8 = 0b11111111
    switch {
        case nonOceanPattern & 0b01010000 == 0b01010000: mask &= ^(nonOceanPattern & 0b00100000); fallthrough
        case nonOceanPattern & 0b01000001 == 0b01000001: mask &= ^(nonOceanPattern & 0b10000000); fallthrough
        case nonOceanPattern & 0b00010100 == 0b00010100: mask &= ^(nonOceanPattern & 0b00001000); fallthrough
        case nonOceanPattern & 0b00000101 == 0b00000101: mask &= ^(nonOceanPattern & 0b00000010)
    }
    landCompatibilities := makeCompatibilities(makeDirections(landPattern & mask), []TerrainType{Ocean, Shore}, NoneOf)
    riverCompatibilities := makeCompatibilities(makeDirections(riverPattern), []TerrainType{River}, AnyOf)
    oceanCompatibilities := makeCompatibilities(makeDirections(^nonOceanPattern), []TerrainType{Ocean, Shore}, AnyOf)

    return makeTile(Shore, index, append(append(landCompatibilities, riverCompatibilities...), oceanCompatibilities...))
}

const AllDirections uint8 = 0b1111_1111

// a bit pattern on a tile indicates the positions where the tile can match up with another tile
// bit index: 0123 4567
//            0000 1000
//            ↖↑↗→ ↘↓↙←
//
// bit 0: north west
// bit 1: north
// bit 2: north east
// bit 3: east
// bit 4: south east
// bit 5: south
// bit 6: south west
// bit 7: west

var (
    TileOcean = makeTile(Ocean, 0x0, makeCompatibilities(makeDirections(AllDirections), []TerrainType{Ocean, Shore}, AnyOf))
    TileLand = makeSimpleTile(0x1, Grass)

    TileShore1_00001000 = makeShoreTile(0x02, 0b00001000)
    TileShore1_00001100 = makeShoreTile(0x03, 0b00001100)
    TileShore1_00001110 = makeShoreTile(0x04, 0b00001110)
    TileShore1_00000110 = makeShoreTile(0x05, 0b00000110)
    TileShore1_00000010 = makeShoreTile(0x06, 0b00000010)
    TileShore1_00001010 = makeShoreTile(0x07, 0b00001010)
    TileShore1_00100010 = makeShoreTile(0x08, 0b00100010)
    TileShore1_10000010 = makeShoreTile(0x09, 0b10000010)
    TileShore1_00011000 = makeShoreTile(0x0A, 0b00011000)
    TileShore1_00000100 = makeShoreTile(0x0B, 0b00000100)
    TileShore1_00000011 = makeShoreTile(0x0C, 0b00000011)
    TileShore1_10100000 = makeShoreTile(0x0D, 0b10100000)
    TileShore1_10001000 = makeShoreTile(0x0E, 0b10001000)
    TileShore1_00101000 = makeShoreTile(0x0F, 0b00101000)
    TileShore1_00111000 = makeShoreTile(0x10, 0b00111000)
    TileShore1_00010000 = makeShoreTile(0x11, 0b00010000)
    TileLake            = makeLakeTile(0x12)
    TileShore1_00000001 = makeShoreTile(0x13, 0b00000001)
    TileShore1_10000011 = makeShoreTile(0x14, 0b10000011)
    TileShore1_00110000 = makeShoreTile(0x15, 0b00110000)
    TileShore1_01000000 = makeShoreTile(0x16, 0b01000000)
    TileShore1_10000001 = makeShoreTile(0x17, 0b10000001)
    TileShore1_10101000 = makeShoreTile(0x18, 0b10101000)
    TileShore1_00101010 = makeShoreTile(0x19, 0b00101010)
    TileShore1_10001010 = makeShoreTile(0x1A, 0b10001010)
    TileShore1_00100000 = makeShoreTile(0x1B, 0b00100000)
    TileShore1_01100000 = makeShoreTile(0x1C, 0b01100000)
    TileShore1_11100000 = makeShoreTile(0x1D, 0b11100000)
    TileShore1_11000000 = makeShoreTile(0x1E, 0b11000000)
    TileShore1_10000000 = makeShoreTile(0x1F, 0b10000000)
    TileShore1_10100010 = makeShoreTile(0x20, 0b10100010)
    TileShore1_10101010 = makeShoreTile(0x21, 0b10101010)
    TileShore1_11000001 = makeShoreTile(0x22, 0b11000001)
    TileShore1_11100001 = makeShoreTile(0x23, 0b11100001)
    TileShore1_11000011 = makeShoreTile(0x24, 0b11000011)
    TileShore1_11100011 = makeShoreTile(0x25, 0b11100011)
    TileShore1_00011100 = makeShoreTile(0x26, 0b00011100)
    TileShore1_00111100 = makeShoreTile(0x27, 0b00111100)
    TileShore1_00011110 = makeShoreTile(0x28, 0b00011110)
    TileShore1_00111110 = makeShoreTile(0x29, 0b00111110)
    TileShore1_01110000 = makeShoreTile(0x2A, 0b01110000)
    TileShore1_01111000 = makeShoreTile(0x2B, 0b01111000)
    TileShore1_11110000 = makeShoreTile(0x2C, 0b11110000)
    TileShore1_11111000 = makeShoreTile(0x2D, 0b11111000)
    TileShore1_00000111 = makeShoreTile(0x2E, 0b00000111)
    TileShore1_00001111 = makeShoreTile(0x2F, 0b00001111)
    TileShore1_10000111 = makeShoreTile(0x30, 0b10000111)
    TileShore1_10001111 = makeShoreTile(0x31, 0b10001111)
    TileShore1_11101110 = makeShoreTile(0x32, 0b11101110)
    TileShore1_11100110 = makeShoreTile(0x33, 0b11100110)
    TileShore1_11101100 = makeShoreTile(0x34, 0b11101100)
    TileShore1_11100100 = makeShoreTile(0x35, 0b11100100)
    TileShore1_11001110 = makeShoreTile(0x36, 0b11001110)
    TileShore1_11000110 = makeShoreTile(0x37, 0b11000110)
    TileShore1_11001100 = makeShoreTile(0x38, 0b11001100)
    TileShore1_11000100 = makeShoreTile(0x39, 0b11000100)
    TileShore1_01101110 = makeShoreTile(0x3A, 0b01101110)
    TileShore1_01100110 = makeShoreTile(0x3B, 0b01100110)
    TileShore1_01101100 = makeShoreTile(0x3C, 0b01101100)
    TileShore1_01100100 = makeShoreTile(0x3D, 0b01100100)
    TileShore1_01001110 = makeShoreTile(0x3E, 0b01001110)
    TileShore1_01000110 = makeShoreTile(0x3F, 0b01000110)
    TileShore1_01001100 = makeShoreTile(0x40, 0b01001100)
    TileShore1_01000100 = makeShoreTile(0x41, 0b01000100)
    TileShore1_10010011 = makeShoreTile(0x42, 0b10010011)
    TileShore1_10011011 = makeShoreTile(0x43, 0b10011011)
    TileShore1_10110011 = makeShoreTile(0x44, 0b10110011)
    TileShore1_10111011 = makeShoreTile(0x45, 0b10111011)
    TileShore1_10010001 = makeShoreTile(0x46, 0b10010001)
    TileShore1_10011001 = makeShoreTile(0x47, 0b10011001)
    TileShore1_10110001 = makeShoreTile(0x48, 0b10110001)
    TileShore1_10111001 = makeShoreTile(0x49, 0b10111001)
    TileShore1_00010011 = makeShoreTile(0x4A, 0b00010011)
    TileShore1_00011011 = makeShoreTile(0x4B, 0b00011011)
    TileShore1_00110011 = makeShoreTile(0x4C, 0b00110011)
    TileShore1_00111011 = makeShoreTile(0x4D, 0b00111011)
    TileShore1_00010001 = makeShoreTile(0x4E, 0b00010001)
    TileShore1_00011001 = makeShoreTile(0x4F, 0b00011001)
    TileShore1_00110001 = makeShoreTile(0x50, 0b00110001)
    TileShore1_00111001 = makeShoreTile(0x51, 0b00111001)
    TileShore1_00011111 = makeShoreTile(0x52, 0b00011111)
    TileShore1_11000111 = makeShoreTile(0x53, 0b11000111)
    TileShore1_11110001 = makeShoreTile(0x54, 0b11110001)
    TileShore1_01111100 = makeShoreTile(0x55, 0b01111100)
    TileShore1_10011111 = makeShoreTile(0x56, 0b10011111)
    TileShore1_11100111 = makeShoreTile(0x57, 0b11100111)
    TileShore1_11111001 = makeShoreTile(0x58, 0b11111001)
    TileShore1_01111110 = makeShoreTile(0x59, 0b01111110)
    TileShore1_00111111 = makeShoreTile(0x5A, 0b00111111)
    TileShore1_11001111 = makeShoreTile(0x5B, 0b11001111)
    TileShore1_11110011 = makeShoreTile(0x5C, 0b11110011)
    TileShore1_11111100 = makeShoreTile(0x5D, 0b11111100)
    TileShore1_10111111 = makeShoreTile(0x5E, 0b10111111)
    TileShore1_11101111 = makeShoreTile(0x5F, 0b11101111)
    TileShore1_11111011 = makeShoreTile(0x60, 0b11111011)
    TileShore1_11111110 = makeShoreTile(0x61, 0b11111110)
    TileShore1_10111000 = makeShoreTile(0x62, 0b10111000)
    TileShore1_10110000 = makeShoreTile(0x63, 0b10110000)
    TileShore1_10011000 = makeShoreTile(0x64, 0b10011000)
    TileShore1_10010000 = makeShoreTile(0x65, 0b10010000)
    TileShore1_10111010 = makeShoreTile(0x66, 0b10111010)
    TileShore1_10110010 = makeShoreTile(0x67, 0b10110010)
    TileShore1_10011010 = makeShoreTile(0x68, 0b10011010)
    TileShore1_10010010 = makeShoreTile(0x69, 0b10010010)
    TileShore1_00111010 = makeShoreTile(0x6A, 0b00111010)
    TileShore1_00110010 = makeShoreTile(0x6B, 0b00110010)
    TileShore1_00011010 = makeShoreTile(0x6C, 0b00011010)
    TileShore1_00010010 = makeShoreTile(0x6D, 0b00010010)
    TileShore1_10001110 = makeShoreTile(0x6E, 0b10001110)
    TileShore1_10101110 = makeShoreTile(0x6F, 0b10101110)
    TileShore1_00101110 = makeShoreTile(0x70, 0b00101110)
    TileShore1_10001100 = makeShoreTile(0x71, 0b10001100)
    TileShore1_10101100 = makeShoreTile(0x72, 0b10101100)
    TileShore1_00101100 = makeShoreTile(0x73, 0b00101100)
    TileShore1_10000110 = makeShoreTile(0x74, 0b10000110)
    TileShore1_10100110 = makeShoreTile(0x75, 0b10100110)
    TileShore1_00100110 = makeShoreTile(0x76, 0b00100110)
    TileShore1_10000100 = makeShoreTile(0x77, 0b10000100)
    TileShore1_10100100 = makeShoreTile(0x78, 0b10100100)
    TileShore1_00100100 = makeShoreTile(0x79, 0b00100100)
    TileShore1_00100001 = makeShoreTile(0x7A, 0b00100001)
    TileShore1_10100001 = makeShoreTile(0x7B, 0b10100001)
    TileShore1_00100011 = makeShoreTile(0x7C, 0b00100011)
    TileShore1_10100011 = makeShoreTile(0x7D, 0b10100011)
    TileShore1_00101001 = makeShoreTile(0x7E, 0b00101001)
    TileShore1_10101001 = makeShoreTile(0x7F, 0b10101001)
    TileShore1_00101011 = makeShoreTile(0x80, 0b00101011)
    TileShore1_10101011 = makeShoreTile(0x81, 0b10101011)
    TileShore1_00001001 = makeShoreTile(0x82, 0b00001001)
    TileShore1_10001001 = makeShoreTile(0x83, 0b10001001)
    TileShore1_00001011 = makeShoreTile(0x84, 0b00001011)
    TileShore1_10001011 = makeShoreTile(0x85, 0b10001011)
    TileShore1_01000010 = makeShoreTile(0x86, 0b01000010)
    TileShore1_01001010 = makeShoreTile(0x87, 0b01001010)
    TileShore1_01001000 = makeShoreTile(0x88, 0b01001000)
    TileShore1_11000010 = makeShoreTile(0x89, 0b11000010)
    TileShore1_11001010 = makeShoreTile(0x8A, 0b11001010)
    TileShore1_11001000 = makeShoreTile(0x8B, 0b11001000)
    TileShore1_01100010 = makeShoreTile(0x8C, 0b01100010)
    TileShore1_01101010 = makeShoreTile(0x8D, 0b01101010)
    TileShore1_01101000 = makeShoreTile(0x8E, 0b01101000)
    TileShore1_11100010 = makeShoreTile(0x8F, 0b11100010)
    TileShore1_11101010 = makeShoreTile(0x90, 0b11101010)
    TileShore1_11101000 = makeShoreTile(0x91, 0b11101000)
    TileShore1_11001001 = makeShoreTile(0x92, 0b11001001)
    TileShore1_11101001 = makeShoreTile(0x93, 0b11101001)
    TileShore1_11001011 = makeShoreTile(0x94, 0b11001011)
    TileShore1_11101011 = makeShoreTile(0x95, 0b11101011)
    TileShore1_10011100 = makeShoreTile(0x96, 0b10011100)
    TileShore1_10111100 = makeShoreTile(0x97, 0b10111100)
    TileShore1_10011110 = makeShoreTile(0x98, 0b10011110)
    TileShore1_10111110 = makeShoreTile(0x99, 0b10111110)
    TileShore1_01110010 = makeShoreTile(0x9A, 0b01110010)
    TileShore1_01111010 = makeShoreTile(0x9B, 0b01111010)
    TileShore1_11110010 = makeShoreTile(0x9C, 0b11110010)
    TileShore1_11111010 = makeShoreTile(0x9D, 0b11111010)
    TileShore1_00100111 = makeShoreTile(0x9E, 0b00100111)
    TileShore1_00101111 = makeShoreTile(0x9F, 0b00101111)
    TileShore1_10100111 = makeShoreTile(0xA0, 0b10100111)
    TileShore1_10101111 = makeShoreTile(0xA1, 0b10101111)

    TileGrasslands1     = makeSimpleTile(0xA2, Grass)
    TileForest1         = makeSimpleTile(0xA3, Forest)
    TileMountain1       = makeMountainTile(0xA4, 0b0000)
    TileAllDesert1      = makeDesertTile(0xA5, 0b00000000)
    TileSwamp1          = makeSimpleTile(0xA6, Swamp)
    TileAllTundra1      = makeTundraTile(0xA7, 0b0000)
    TileSorceryLake     = makeSimpleTile(0xA8, SorceryNode)
    TileNatureForest    = makeSimpleTile(0xA9, NatureNode)
    TileChaosVolcano    = makeSimpleTile(0xAA, ChaosNode)
    TileHills1          = makeHillTile(0xAB,0b0000)
    TileGrasslands2     = makeSimpleTile(0xAC, Grass)
    TileGrasslands3     = makeSimpleTile(0xAD, Grass)
    TileAllDesert2      = makeDesertTile(0xAE, 0b00000000)
    TileAllDesert3      = makeDesertTile(0xAF, 0b00000000)
    TileAllDesert4      = makeDesertTile(0xB0, 0b00000000)
    TileSwamp2          = makeSimpleTile(0xB1, Swamp)
    TileSwamp3          = makeSimpleTile(0xB2, Swamp)
    TileVolcano         = makeSimpleTile(0xB3, Volcano)
    TileGrasslands4     = makeSimpleTile(0xB4, Grass)
    TileAllTundra2      = makeTundraTile(0xB5, 0b0000)
    TileAllTundra3      = makeTundraTile(0xB6, 0b0000)
    TileForest2         = makeSimpleTile(0xB7, Forest)
    TileForest3         = makeSimpleTile(0xB8, Forest)

    TileRiver0010       = makeRiverTile(0xB9, 0b0010)
    TileRiver0001       = makeRiverTile(0xBA, 0b0001)
    TileRiver1000       = makeRiverTile(0xBB, 0b1000)
    TileRiver0100       = makeRiverTile(0xBC, 0b0100)
    TileRiver1100       = makeRiverTile(0xBD, 0b1100)
    TileRiver0011       = makeRiverTile(0xBE, 0b0011)
    TileRiver0110       = makeRiverTile(0xBF, 0b0110)
    TileRiver1001       = makeRiverTile(0xC0, 0b1001)
    TileRiver1100_1     = makeRiverTile(0xC1, 0b1100)
    TileRiver0011_1     = makeRiverTile(0xC2, 0b0011)
    TileRiver0110_1     = makeRiverTile(0xC3, 0b0110)
    TileRiver1001_1     = makeRiverTile(0xC4, 0b1001)

    TileLakeRiverWest   = makeLakeRiverTile(0xC5, 0b0001)
    TileLakeRiverNorth  = makeLakeRiverTile(0xC6, 0b1000)
    TileLakeRiverEast   = makeLakeRiverTile(0xC7, 0b0100)
    TileLakeRiverSouth  = makeLakeRiverTile(0xC8, 0b0010)

    // land at north-west, river at west and north
    // first direction is land, second is river
    TileShore_1R00000R   = makeShoreRiverTile(0xC9, 0b10000000, 0b01000001)
    TileShore_1R10000R   = makeShoreRiverTile(0xCA, 0b10100000, 0b01000001)
    TileShore_1R00001R   = makeShoreRiverTile(0xCB, 0b10000010, 0b01000001)
    TileShore_1R10001R   = makeShoreRiverTile(0xCC, 0b10100010, 0b01000001)
    TileShore_000R1R00   = makeShoreRiverTile(0xCD, 0b00001000, 0b00010100)
    TileShore_000R1R10   = makeShoreRiverTile(0xCE, 0b00001010, 0b00010100)
    TileShore_001R1R00   = makeShoreRiverTile(0xCF, 0b00101000, 0b00010100)
    TileShore_001R1R10   = makeShoreRiverTile(0xD0, 0b00101010, 0b00010100)
    TileShore_0R1R0000   = makeShoreRiverTile(0xD1, 0b00100000, 0b01010000)
    TileShore_0R1R1000   = makeShoreRiverTile(0xD2, 0b00101000, 0b01010000)
    TileShore_1R1R0000   = makeShoreRiverTile(0xD3, 0b10100000, 0b01010000)
    TileShore_1R1R1000   = makeShoreRiverTile(0xD4, 0b10101000, 0b01010000)
    TileShore_00000R1R   = makeShoreRiverTile(0xD5, 0b00000010, 0b00000101)
    TileShore_00001R1R   = makeShoreRiverTile(0xD6, 0b00001010, 0b00000101)
    TileShore_10000R1R   = makeShoreRiverTile(0xD7, 0b10000010, 0b00000101)
    TileShore_10001R1R   = makeShoreRiverTile(0xD8, 0b10001010, 0b00000101)
    TileShore_00001R10   = makeShoreRiverTile(0xD9, 0b00001010, 0b00000100)
    TileShore_00001R00   = makeShoreRiverTile(0xDA, 0b00001000, 0b00000100)
    TileShore_00000R10   = makeShoreRiverTile(0xDB, 0b00000010, 0b00000100)
    TileShore_00000R00   = makeShoreRiverTile(0xDC, 0b00000000, 0b00000100)
    TileShore_1000001R   = makeShoreRiverTile(0xDD, 0b10000010, 0b00000001)
    TileShore_0000001R   = makeShoreRiverTile(0xDE, 0b00000010, 0b00000001)
    TileShore_1000000R   = makeShoreRiverTile(0xDF, 0b10000000, 0b00000001)
    TileShore_0000000R   = makeShoreRiverTile(0xE0, 0b00000000, 0b00000001)
    TileShore_1R100000   = makeShoreRiverTile(0xE1, 0b10100000, 0b01000000)
    TileShore_1R000000   = makeShoreRiverTile(0xE2, 0b10000000, 0b01000000)
    TileShore_0R100000   = makeShoreRiverTile(0xE3, 0b00100000, 0b01000000)
    TileShore_0R000000   = makeShoreRiverTile(0xE4, 0b00000000, 0b01000000)
    TileShore_001R1000   = makeShoreRiverTile(0xE5, 0b00101000, 0b00010000)
    TileShore_001R0000   = makeShoreRiverTile(0xE6, 0b00100000, 0b00010000)
    TileShore_000R1000   = makeShoreRiverTile(0xE7, 0b00001000, 0b00010000)
    TileShore_000R0000   = makeShoreRiverTile(0xE8, 0b00000000, 0b00010000)

    TileRiver1100_3      = makeRiverTile(0xE9, 0b1100)
    TileRiver0011_3      = makeRiverTile(0xEA, 0b0011)
    TileRiver0110_3      = makeRiverTile(0xEB, 0b0110)
    TileRiver1001_3      = makeRiverTile(0xEC, 0b1001)
    TileRiver1010_1      = makeRiverTile(0xED, 0b1010)
    TileRiver1010_2      = makeRiverTile(0xEE, 0b1010)
    TileRiver1010_3      = makeRiverTile(0xEF, 0b1010)
    TileRiver0101_1      = makeRiverTile(0xF0, 0b0101)
    TileRiver0101_2      = makeRiverTile(0xF1, 0b0101)
    TileRiver0101_3      = makeRiverTile(0xF2, 0b0101)
    TileRiver1101_1      = makeRiverTile(0xF3, 0b1101)
    TileRiver1101_2      = makeRiverTile(0xF4, 0b1101)
    TileRiver1101_3      = makeRiverTile(0xF5, 0b1101)
    TileRiver1101_4      = makeRiverTile(0xF6, 0b1101)
    TileRiver0111_1      = makeRiverTile(0xF7, 0b0111)
    TileRiver0111_2      = makeRiverTile(0xF8, 0b0111)
    TileRiver0111_3      = makeRiverTile(0xF9, 0b0111)
    TileRiver0111_4      = makeRiverTile(0xFA, 0b0111)
    TileRiver1110_1      = makeRiverTile(0xFB, 0b1110)
    TileRiver1110_2      = makeRiverTile(0xFC, 0b1110)
    TileRiver1110_3      = makeRiverTile(0xFD, 0b1110)
    TileRiver1110_4      = makeRiverTile(0xFE, 0b1110)
    TileRiver1011_1      = makeRiverTile(0xFF, 0b1011)
    TileRiver1011_2      = makeRiverTile(0x100, 0b1011)
    TileRiver1011_3      = makeRiverTile(0x101, 0b1011)
    TileRiver1011_4      = makeRiverTile(0x102, 0b1011)

    TileMountain_0010    = makeMountainTile(0x103, 0b0010)
    TileMountain_0100    = makeMountainTile(0x104, 0b0100)
    TileMountain_1111    = makeMountainTile(0x105, 0b1111)
    TileMountain_0101    = makeMountainTile(0x106, 0b0101)
    TileMountain_0001    = makeMountainTile(0x107, 0b0001)
    TileMountain_1010    = makeMountainTile(0x108, 0b1010)
    TileMountain_1000    = makeMountainTile(0x109, 0b1000)
    TileMountain_0110    = makeMountainTile(0x10A, 0b0110)
    TileMountain_0111    = makeMountainTile(0x10B, 0b0111)
    TileMountain_0011    = makeMountainTile(0x10C, 0b0011)
    TileMountain_1110    = makeMountainTile(0x10D, 0b1110)
    TileMountain2_1111   = makeMountainTile(0x10E, 0b1111)
    TileMountain_1011    = makeMountainTile(0x10F, 0b1011)
    TileMountain_1100    = makeMountainTile(0x110, 0b1100)
    TileMountain_1101    = makeMountainTile(0x111, 0b1101)
    TileMountain_1001    = makeMountainTile(0x112, 0b1001)

    TileHills_0010       = makeHillTile(0x113, 0b0010)
    TileHills_0100       = makeHillTile(0x114, 0b0100)
    TileHills_1111       = makeHillTile(0x115, 0b1111)
    TileHills_0101       = makeHillTile(0x116, 0b0101)
    TileHills_0001       = makeHillTile(0x117, 0b0001)
    TileHills_1010       = makeHillTile(0x118, 0b1010)
    TileHills_1000       = makeHillTile(0x119, 0b1000)
    TileHills_0110       = makeHillTile(0x11A, 0b0110)
    TileHills_0111       = makeHillTile(0x11B, 0b0111)
    TileHills_0011       = makeHillTile(0x11C, 0b0011)
    TileHills_1110       = makeHillTile(0x11D, 0b1110)
    TileHill2_1111       = makeHillTile(0x11E, 0b1111)
    TileHills_1011       = makeHillTile(0x11F, 0b1011)
    TileHills_1100       = makeHillTile(0x120, 0b1100)
    TileHills_1101       = makeHillTile(0x121, 0b1101)
    TileHills_1001       = makeHillTile(0x122, 0b1001)
    // not sure on this one
    TileHills2_1111      = makeHillTile(0x123, 0b1111)

    TileDesert_00001000  = makeDesertTile(0x124, 0b00001000)
    TileDesert_00001100  = makeDesertTile(0x125, 0b00001100)
    TileDesert_00001110  = makeDesertTile(0x126, 0b00001110)
    TileDesert_00000110  = makeDesertTile(0x127, 0b00000110)
    TileDesert_00000010  = makeDesertTile(0x128, 0b00000010)
    TileDesert_00001010  = makeDesertTile(0x129, 0b00001010)
    TileDesert_00100010  = makeDesertTile(0x12A, 0b00100010)
    TileDesert_10000010  = makeDesertTile(0x12B, 0b10000010)
    TileDesert_00011000  = makeDesertTile(0x12C, 0b00011000)
    TileDesert_00000100  = makeDesertTile(0x12D, 0b00000100)
    TileDesert_00000011  = makeDesertTile(0x12E, 0b00000011)
    TileDesert_10100000  = makeDesertTile(0x12F, 0b10100000)
    TileDesert_10001000  = makeDesertTile(0x130, 0b10001000)
    TileDesert_00101000  = makeDesertTile(0x131, 0b00101000)
    TileDesert_00111000  = makeDesertTile(0x132, 0b00111000)
    TileDesert_00010000  = makeDesertTile(0x133, 0b00010000)
    TileDesert_00000000  = makeCardinalTile(Desert, 0x134, 0b00000000)
    TileDesert_00000001  = makeDesertTile(0x135, 0b00000001)
    TileDesert_10000011  = makeDesertTile(0x136, 0b10000011)
    TileDesert_00110000  = makeDesertTile(0x137, 0b00110000)
    TileDesert_01000000  = makeDesertTile(0x138, 0b01000000)
    TileDesert_10000001  = makeDesertTile(0x139, 0b10000001)
    TileDesert_10101000  = makeDesertTile(0x13A, 0b10101000)
    TileDesert_00101010  = makeDesertTile(0x13B, 0b00101010)
    TileDesert_10001010  = makeDesertTile(0x13C, 0b10001010)
    TileDesert_00100000  = makeDesertTile(0x13D, 0b00100000)
    TileDesert_01100000  = makeDesertTile(0x13E, 0b01100000)
    TileDesert_11100000  = makeDesertTile(0x13F, 0b11100000)
    TileDesert_11000000  = makeDesertTile(0x140, 0b11000000)
    TileDesert_10000000  = makeDesertTile(0x141, 0b10000000)
    TileDesert_10100010  = makeDesertTile(0x142, 0b10100010)
    TileDesert_10101010  = makeDesertTile(0x143, 0b10101010)
    TileDesert_11000001  = makeDesertTile(0x144, 0b11000001)
    TileDesert_11100001  = makeDesertTile(0x145, 0b11100001)
    TileDesert_11000011  = makeDesertTile(0x146, 0b11000011)
    TileDesert_11100011  = makeDesertTile(0x147, 0b11100011)
    TileDesert_00011100  = makeDesertTile(0x148, 0b00011100)
    TileDesert_00111100  = makeDesertTile(0x149, 0b00111100)
    TileDesert_00011110  = makeDesertTile(0x14A, 0b00011110)
    TileDesert_00111110  = makeDesertTile(0x14B, 0b00111110)
    TileDesert_01110000  = makeDesertTile(0x14C, 0b01110000)
    TileDesert_01111000  = makeDesertTile(0x14D, 0b01111000)
    TileDesert_11110000  = makeDesertTile(0x14E, 0b11110000)
    TileDesert_11111000  = makeDesertTile(0x14F, 0b11111000)
    TileDesert_00000111  = makeDesertTile(0x150, 0b00000111)
    TileDesert_00001111  = makeDesertTile(0x151, 0b00001111)
    TileDesert_10000111  = makeDesertTile(0x152, 0b10000111)
    TileDesert_10001111  = makeDesertTile(0x153, 0b10001111)
    TileDesert_11101110  = makeDesertTile(0x154, 0b11101110)
    TileDesert_11100110  = makeDesertTile(0x155, 0b11100110)
    TileDesert_11101100  = makeDesertTile(0x156, 0b11101100)
    TileDesert_11100100  = makeDesertTile(0x157, 0b11100100)
    TileDesert_11001110  = makeDesertTile(0x158, 0b11001110)
    TileDesert_11000110  = makeDesertTile(0x159, 0b11000110)
    TileDesert_11001100  = makeDesertTile(0x15A, 0b11001100)
    TileDesert_11000100  = makeDesertTile(0x15B, 0b11000100)
    TileDesert_01101110  = makeDesertTile(0x15C, 0b01101110)
    TileDesert_01100110  = makeDesertTile(0x15D, 0b01100110)
    TileDesert_01101100  = makeDesertTile(0x15E, 0b01101100)
    TileDesert_01100100  = makeDesertTile(0x15F, 0b01100100)
    TileDesert_01001110  = makeDesertTile(0x160, 0b01001110)
    TileDesert_01000110  = makeDesertTile(0x161, 0b01000110)
    TileDesert_01001100  = makeDesertTile(0x162, 0b01001100)
    TileDesert_01000100  = makeDesertTile(0x163, 0b01000100)
    TileDesert_10010011  = makeDesertTile(0x164, 0b10010011)
    TileDesert_10011011  = makeDesertTile(0x165, 0b10011011)
    TileDesert_10110011  = makeDesertTile(0x166, 0b10110011)
    TileDesert_10111011  = makeDesertTile(0x167, 0b10111011)
    TileDesert_10010001  = makeDesertTile(0x168, 0b10010001)
    TileDesert_10011001  = makeDesertTile(0x169, 0b10011001)
    TileDesert_10110001  = makeDesertTile(0x16A, 0b10110001)
    TileDesert_10111001  = makeDesertTile(0x16B, 0b10111001)
    TileDesert_00010011  = makeDesertTile(0x16C, 0b00010011)
    TileDesert_00011011  = makeDesertTile(0x16D, 0b00011011)
    TileDesert_00110011  = makeDesertTile(0x16E, 0b00110011)
    TileDesert_00111011  = makeDesertTile(0x16F, 0b00111011)
    TileDesert_00010001  = makeDesertTile(0x170, 0b00010001)
    TileDesert_00011001  = makeDesertTile(0x171, 0b00011001)
    TileDesert_00110001  = makeDesertTile(0x172, 0b00110001)
    TileDesert_00111001  = makeDesertTile(0x173, 0b00111001)
    TileDesert_00011111  = makeDesertTile(0x174, 0b00011111)
    TileDesert_11000111  = makeDesertTile(0x175, 0b11000111)
    TileDesert_11110001  = makeDesertTile(0x176, 0b11110001)
    TileDesert_01111100  = makeDesertTile(0x177, 0b01111100)
    TileDesert_10011111  = makeDesertTile(0x178, 0b10011111)
    TileDesert_11100111  = makeDesertTile(0x179, 0b11100111)
    TileDesert_11111001  = makeDesertTile(0x17A, 0b11111001)
    TileDesert_01111110  = makeDesertTile(0x17B, 0b01111110)
    TileDesert_00111111  = makeDesertTile(0x17C, 0b00111111)
    TileDesert_11001111  = makeDesertTile(0x17D, 0b11001111)
    TileDesert_11110011  = makeDesertTile(0x17E, 0b11110011)
    TileDesert_11111100  = makeDesertTile(0x17F, 0b11111100)
    TileDesert_10111111  = makeDesertTile(0x180, 0b10111111)
    TileDesert_11101111  = makeDesertTile(0x181, 0b11101111)
    TileDesert_11111011  = makeDesertTile(0x182, 0b11111011)
    TileDesert_11111110  = makeDesertTile(0x183, 0b11111110)
    TileDesert_10111000  = makeDesertTile(0x184, 0b10111000)
    TileDesert_10110000  = makeDesertTile(0x185, 0b10110000)
    TileDesert_10011000  = makeDesertTile(0x186, 0b10011000)
    TileDesert_10010000  = makeDesertTile(0x187, 0b10010000)
    TileDesert_10111010  = makeDesertTile(0x188, 0b10111010)
    TileDesert_10110010  = makeDesertTile(0x189, 0b10110010)
    TileDesert_10011010  = makeDesertTile(0x18A, 0b10011010)
    TileDesert_10010010  = makeDesertTile(0x18B, 0b10010010)
    TileDesert_00111010  = makeDesertTile(0x18C, 0b00111010)
    TileDesert_00110010  = makeDesertTile(0x18D, 0b00110010)
    TileDesert_00011010  = makeDesertTile(0x18E, 0b00011010)
    TileDesert_00010010  = makeDesertTile(0x18F, 0b00010010)
    TileDesert_10001110  = makeDesertTile(0x190, 0b10001110)
    TileDesert_10101110  = makeDesertTile(0x191, 0b10101110)
    TileDesert_00101110  = makeDesertTile(0x192, 0b00101110)
    TileDesert_10001100  = makeDesertTile(0x193, 0b10001100)
    TileDesert_10101100  = makeDesertTile(0x194, 0b10101100)
    TileDesert_00101100  = makeDesertTile(0x195, 0b00101100)
    TileDesert_10000110  = makeDesertTile(0x196, 0b10000110)
    TileDesert_10100110  = makeDesertTile(0x197, 0b10100110)
    TileDesert_00100110  = makeDesertTile(0x198, 0b00100110)
    TileDesert_10000100  = makeDesertTile(0x199, 0b10000100)
    TileDesert_10100100  = makeDesertTile(0x19A, 0b10100100)
    TileDesert_00100100  = makeDesertTile(0x19B, 0b00100100)
    TileDesert_00100001  = makeDesertTile(0x19C, 0b00100001)
    TileDesert_10100001  = makeDesertTile(0x19D, 0b10100001)
    TileDesert_00100011  = makeDesertTile(0x19E, 0b00100011)
    TileDesert_10100011  = makeDesertTile(0x19F, 0b10100011)
    TileDesert_00101001  = makeDesertTile(0x1A0, 0b00101001)
    TileDesert_10101001  = makeDesertTile(0x1A1, 0b10101001)
    TileDesert_00101011  = makeDesertTile(0x1A2, 0b00101011)
    TileDesert_10101011  = makeDesertTile(0x1A3, 0b10101011)
    TileDesert_00001001  = makeDesertTile(0x1A4, 0b00001001)
    TileDesert_10001001  = makeDesertTile(0x1A5, 0b10001001)
    TileDesert_00001011  = makeDesertTile(0x1A6, 0b00001011)
    TileDesert_10001011  = makeDesertTile(0x1A7, 0b10001011)
    TileDesert_01000010  = makeDesertTile(0x1A8, 0b01000010)
    TileDesert_01001010  = makeDesertTile(0x1A9, 0b01001010)
    TileDesert_01001000  = makeDesertTile(0x1AA, 0b01001000)
    TileDesert_11000010  = makeDesertTile(0x1AB, 0b11000010)
    TileDesert_11001010  = makeDesertTile(0x1AC, 0b11001010)
    TileDesert_11001000  = makeDesertTile(0x1AD, 0b11001000)
    TileDesert_01100010  = makeDesertTile(0x1AE, 0b01100010)
    TileDesert_01101010  = makeDesertTile(0x1AF, 0b01101010)
    TileDesert_01101000  = makeDesertTile(0x1B0, 0b01101000)
    TileDesert_11100010  = makeDesertTile(0x1B1, 0b11100010)
    TileDesert_11101010  = makeDesertTile(0x1B2, 0b11101010)
    TileDesert_11101000  = makeDesertTile(0x1B3, 0b11101000)
    TileDesert_11001001  = makeDesertTile(0x1B4, 0b11001001)
    TileDesert_11101001  = makeDesertTile(0x1B5, 0b11101001)
    TileDesert_11001011  = makeDesertTile(0x1B6, 0b11001011)
    TileDesert_11101011  = makeDesertTile(0x1B7, 0b11101011)
    TileDesert_10011100  = makeDesertTile(0x1B8, 0b10011100)
    TileDesert_10111100  = makeDesertTile(0x1B9, 0b10111100)
    TileDesert_10011110  = makeDesertTile(0x1BA, 0b10011110)
    TileDesert_10111110  = makeDesertTile(0x1BB, 0b10111110)
    TileDesert_01110010  = makeDesertTile(0x1BC, 0b01110010)
    TileDesert_01111010  = makeDesertTile(0x1BD, 0b01111010)
    TileDesert_11110010  = makeDesertTile(0x1BE, 0b11110010)
    TileDesert_11111010  = makeDesertTile(0x1BF, 0b11111010)
    TileDesert_00100111  = makeDesertTile(0x1C0, 0b00100111)
    TileDesert_00101111  = makeDesertTile(0x1C1, 0b00101111)
    TileDesert_10100111  = makeDesertTile(0x1C2, 0b10100111)
    TileDesert_10101111  = makeDesertTile(0x1C3, 0b10101111)

    // pattern1 = land, pattern2 = river
    TileShore2_00011R11  = makeShoreRiverTile(0x1C4, 0b00011011, 0b00000100)
    TileShore2_1100011R  = makeShoreRiverTile(0x1C5, 0b11000110, 0b00000001)
    TileShore2_1R110001  = makeShoreRiverTile(0x1C6, 0b10110001, 0b01000000)
    TileShore2_011R1100  = makeShoreRiverTile(0x1C7, 0b01101100, 0b00010000)
    TileShore2_10011R11  = makeShoreRiverTile(0x1C8, 0b10011011, 0b00000100)
    TileShore2_1110011R  = makeShoreRiverTile(0x1C9, 0b11100110, 0b00000001)
    TileShore2_1R111001  = makeShoreRiverTile(0x1CA, 0b10111001, 0b01000000)
    TileShore2_011R1110  = makeShoreRiverTile(0x1CB, 0b01101110, 0b00010000)
    TileShore2_00111R11  = makeShoreRiverTile(0x1CC, 0b00111011, 0b00000100)
    TileShore2_1100111R  = makeShoreRiverTile(0x1CD, 0b11001110, 0b00000001)
    TileShore2_1R110011  = makeShoreRiverTile(0x1CE, 0b10110011, 0b01000000)
    TileShore2_111R1100  = makeShoreRiverTile(0x1CF, 0b11101100, 0b00010000)
    TileShore2_10111R11  = makeShoreRiverTile(0x1D0, 0b10111011, 0b00000100)
    TileShore2_1110111R  = makeShoreRiverTile(0x1D1, 0b11101110, 0b00000001)
    TileShore2_1R111011  = makeShoreRiverTile(0x1D2, 0b10111011, 0b01000000)
    TileShore2_111R1110  = makeShoreRiverTile(0x1D3, 0b11101110, 0b00010000)

    TileRiver1111_1     = makeRiverTile(0x1D4, 0b1111)
    TileRiver1111_2     = makeRiverTile(0x1D5, 0b1111)
    TileRiver1111_3     = makeRiverTile(0x1D6, 0b1111)
    TileRiver1111_4     = makeRiverTile(0x1D7, 0b1111)
    TileRiver1111_5     = makeRiverTile(0x1D8, 0b1111)

    // land, river patterns
    TileShore2_1100000R  = makeShoreRiverTile(0x1D9, 0b11000000, 0b00000001)
    TileShore2_1110000R  = makeShoreRiverTile(0x1DA, 0b11100000, 0b00000001)
    TileShore2_1100001R  = makeShoreRiverTile(0x1DB, 0b11000010, 0b00000001)
    TileShore2_1110001R  = makeShoreRiverTile(0x1DC, 0b11100010, 0b00000001)
    TileShore2_00011R00  = makeShoreRiverTile(0x1DD, 0b00011000, 0b00000100)
    TileShore2_00111R00  = makeShoreRiverTile(0x1DE, 0b00111000, 0b00000100)
    TileShore2_00011R10  = makeShoreRiverTile(0x1DF, 0b00011010, 0b00000100)
    TileShore2_00111R10  = makeShoreRiverTile(0x1E0, 0b00111010, 0b00000100)
    TileShore2_0R110000  = makeShoreRiverTile(0x1E1, 0b00110000, 0b01000000)
    TileShore2_0R111000  = makeShoreRiverTile(0x1E2, 0b00111000, 0b01000000)
    TileShore2_1R110000  = makeShoreRiverTile(0x1E3, 0b10110000, 0b01000000)
    TileShore2_1R111000  = makeShoreRiverTile(0x1E4, 0b10111000, 0b01000000)
    TileShore2_00000R11  = makeShoreRiverTile(0x1E5, 0b00000011, 0b00000100)
    TileShore2_00001R11  = makeShoreRiverTile(0x1E6, 0b00001011, 0b00000100)
    TileShore2_10000R11  = makeShoreRiverTile(0x1E7, 0b10000011, 0b00000100)
    TileShore2_10001R11  = makeShoreRiverTile(0x1E8, 0b10001011, 0b00000100)
    TileShore2_1R000001  = makeShoreRiverTile(0x1E9, 0b10000001, 0b01000000)
    TileShore2_1R100001  = makeShoreRiverTile(0x1EA, 0b10100001, 0b01000000)
    TileShore2_1R000011  = makeShoreRiverTile(0x1EB, 0b10000011, 0b01000000)
    TileShore2_1R100011  = makeShoreRiverTile(0x1EC, 0b10100011, 0b01000000)
    TileShore2_000R1100  = makeShoreRiverTile(0x1ED, 0b00001100, 0b00010000)
    TileShore2_001R1100  = makeShoreRiverTile(0x1EE, 0b00101100, 0b00010000)
    TileShore2_000R1110  = makeShoreRiverTile(0x1EF, 0b00001110, 0b00010000)
    TileShore2_001R1110  = makeShoreRiverTile(0x1F0, 0b00101110, 0b00010000)
    TileShore2_011R0000  = makeShoreRiverTile(0x1F1, 0b01100000, 0b00010000)
    TileShore2_011R1000  = makeShoreRiverTile(0x1F2, 0b01101000, 0b00010000)
    TileShore2_111R0000  = makeShoreRiverTile(0x1F3, 0b11100000, 0b00010000)
    TileShore2_111R1000  = makeShoreRiverTile(0x1F4, 0b11101000, 0b00010000)
    TileShore2_0000011R  = makeShoreRiverTile(0x1F5, 0b00000110, 0b00000001)
    TileShore2_0000111R  = makeShoreRiverTile(0x1F6, 0b00001110, 0b00000001)
    TileShore2_1000011R  = makeShoreRiverTile(0x1F7, 0b10000110, 0b00000001)
    TileShore2_1000111R  = makeShoreRiverTile(0x1F8, 0b10001110, 0b00000001)
    TileShore2_0001111R  = makeShoreRiverTile(0x1F9, 0b00011110, 0b00000001)
    TileShore2_1R000111  = makeShoreRiverTile(0x1FA, 0b10000111, 0b01000000)
    TileShore2_111R0001  = makeShoreRiverTile(0x1FB, 0b11100001, 0b00010000)
    TileShore2_0R111100  = makeShoreRiverTile(0x1FC, 0b00111100, 0b01000000)
    TileShore2_1001111R  = makeShoreRiverTile(0x1FD, 0b10011110, 0b00000001)
    TileShore2_1R100111  = makeShoreRiverTile(0x1FE, 0b10100111, 0b01000000)
    TileShore2_111R1001  = makeShoreRiverTile(0x1FF, 0b11101001, 0b00010000)
    TileShore2_0R111110  = makeShoreRiverTile(0x200, 0b00111110, 0b01000000)
    TileShore2_0011111R  = makeShoreRiverTile(0x201, 0b00111110, 0b00000001)
    TileShore2_1R001111  = makeShoreRiverTile(0x202, 0b10001111, 0b01000000)
    TileShore2_111R0011  = makeShoreRiverTile(0x203, 0b11100011, 0b00010000)
    TileShore2_1R111100  = makeShoreRiverTile(0x204, 0b10111100, 0b01000000)
    TileShore2_1011111R  = makeShoreRiverTile(0x205, 0b10111110, 0b00000001)
    TileShore2_1R101111  = makeShoreRiverTile(0x206, 0b10101111, 0b01000000)
    TileShore2_111R1011  = makeShoreRiverTile(0x207, 0b11101011, 0b00010000)
    TileShore2_1R111110  = makeShoreRiverTile(0x208, 0b10111110, 0b01000000)
    TileShore2_000R1111  = makeShoreRiverTile(0x209, 0b00001111, 0b00010000)
    TileShore2_11000R11  = makeShoreRiverTile(0x20A, 0b11000011, 0b00000100)
    TileShore2_1111000R  = makeShoreRiverTile(0x20B, 0b11110000, 0b00000001)
    TileShore2_01111R00  = makeShoreRiverTile(0x20C, 0b01111000, 0b00000100)
    TileShore2_100R1111  = makeShoreRiverTile(0x20D, 0b10001111, 0b00010000)
    TileShore2_11100R11  = makeShoreRiverTile(0x20E, 0b11100011, 0b00000100)
    TileShore2_1111100R  = makeShoreRiverTile(0x20F, 0b11111000, 0b00000001)
    TileShore2_01111R10  = makeShoreRiverTile(0x210, 0b01111010, 0b00000100)
    TileShore2_001R1111  = makeShoreRiverTile(0x211, 0b00101111, 0b00010000)
    TileShore2_11001R11  = makeShoreRiverTile(0x212, 0b11001011, 0b00000100)
    TileShore2_1111001R  = makeShoreRiverTile(0x213, 0b11110010, 0b00000001)
    TileShore2_11111R00  = makeShoreRiverTile(0x214, 0b11111000, 0b00000100)
    TileShore2_101R1111  = makeShoreRiverTile(0x215, 0b10101111, 0b00010000)
    TileShore2_11101R11  = makeShoreRiverTile(0x216, 0b11101011, 0b00000100)
    TileShore2_1111101R  = makeShoreRiverTile(0x217, 0b11111010, 0b00000001)
    TileShore2_11111R10  = makeShoreRiverTile(0x218, 0b11111010, 0b00000100)
    TileShore2_1R101110  = makeShoreRiverTile(0x219, 0b10101110, 0b01000000)
    TileShore2_1R100110  = makeShoreRiverTile(0x21A, 0b10100110, 0b01000000)
    TileShore2_1R101100  = makeShoreRiverTile(0x21B, 0b10101100, 0b01000000)
    TileShore2_1R100100  = makeShoreRiverTile(0x21C, 0b10100100, 0b01000000)
    TileShore2_1R001110  = makeShoreRiverTile(0x21D, 0b10001110, 0b01000000)
    TileShore2_1R000110  = makeShoreRiverTile(0x21E, 0b10000110, 0b01000000)
    TileShore2_1R001100  = makeShoreRiverTile(0x21F, 0b10001100, 0b01000000)
    TileShore2_1R000100  = makeShoreRiverTile(0x220, 0b10000100, 0b01000000)
    TileShore2_0R101110  = makeShoreRiverTile(0x221, 0b00101110, 0b01000000)
    TileShore2_0R100110  = makeShoreRiverTile(0x222, 0b00100110, 0b01000000)
    TileShore2_0R101100  = makeShoreRiverTile(0x223, 0b00101100, 0b01000000)
    TileShore2_0R100100  = makeShoreRiverTile(0x224, 0b00100100, 0b01000000)
    TileShore2_0R001110  = makeShoreRiverTile(0x225, 0b00001110, 0b01000000)
    TileShore2_0R000110  = makeShoreRiverTile(0x226, 0b00000110, 0b01000000)
    TileShore2_0R001100  = makeShoreRiverTile(0x227, 0b00001100, 0b01000000)
    TileShore2_0R000100  = makeShoreRiverTile(0x228, 0b00000100, 0b01000000)
    TileShore2_11101R10  = makeShoreRiverTile(0x229, 0b11101010, 0b00000100)
    TileShore2_11100R10  = makeShoreRiverTile(0x22A, 0b11100010, 0b00000100)
    TileShore2_11101R00  = makeShoreRiverTile(0x22B, 0b11101000, 0b00000100)
    TileShore2_11100R00  = makeShoreRiverTile(0x22C, 0b11100000, 0b00000100)
    TileShore2_11001R10  = makeShoreRiverTile(0x22D, 0b11001010, 0b00000100)
    TileShore2_11000R10  = makeShoreRiverTile(0x22E, 0b11000010, 0b00000100)
    TileShore2_11001R00  = makeShoreRiverTile(0x22F, 0b11001000, 0b00000100)
    TileShore2_11000R00  = makeShoreRiverTile(0x230, 0b11000000, 0b00000100)
    TileShore2_01101R10  = makeShoreRiverTile(0x231, 0b01101010, 0b00000100)
    TileShore2_01100R10  = makeShoreRiverTile(0x232, 0b01100010, 0b00000100)
    TileShore2_01101R00  = makeShoreRiverTile(0x233, 0b01101000, 0b00000100)
    TileShore2_01100R00  = makeShoreRiverTile(0x234, 0b01100000, 0b00000100)
    TileShore2_01001R10  = makeShoreRiverTile(0x235, 0b01001010, 0b00000100)
    TileShore2_01000R10  = makeShoreRiverTile(0x236, 0b01000010, 0b00000100)
    TileShore2_01001R00  = makeShoreRiverTile(0x237, 0b01001000, 0b00000100)
    TileShore2_01000R00  = makeShoreRiverTile(0x238, 0b01000000, 0b00000100)
    TileShore2_1001001R  = makeShoreRiverTile(0x239, 0b10010010, 0b00000001)
    TileShore2_1001101R  = makeShoreRiverTile(0x23A, 0b10011010, 0b00000001)
    TileShore2_1011001R  = makeShoreRiverTile(0x23B, 0b10110010, 0b00000001)
    TileShore2_1011101R  = makeShoreRiverTile(0x23C, 0b10111010, 0b00000001)
    TileShore2_1001000R  = makeShoreRiverTile(0x23D, 0b10010000, 0b00000001)
    TileShore2_1001100R  = makeShoreRiverTile(0x23E, 0b10011000, 0b00000001)
    TileShore2_1011000R  = makeShoreRiverTile(0x23F, 0b10110000, 0b00000001)
    TileShore2_1011100R  = makeShoreRiverTile(0x240, 0b10111000, 0b00000001)
    TileShore2_0001001R  = makeShoreRiverTile(0x241, 0b00010010, 0b00000001)
    TileShore2_0001101R  = makeShoreRiverTile(0x242, 0b00011010, 0b00000001)
    TileShore2_0011001R  = makeShoreRiverTile(0x243, 0b00110010, 0b00000001)
    TileShore2_0011101R  = makeShoreRiverTile(0x244, 0b00111010, 0b00000001)
    TileShore2_0001000R  = makeShoreRiverTile(0x245, 0b00010000, 0b00000001)
    TileShore2_0001100R  = makeShoreRiverTile(0x246, 0b00011000, 0b00000001)
    TileShore2_0011000R  = makeShoreRiverTile(0x247, 0b00110000, 0b00000001)
    TileShore2_0011100R  = makeShoreRiverTile(0x248, 0b00111000, 0b00000001)
    TileShore2_100R0011  = makeShoreRiverTile(0x249, 0b10000011, 0b00010000)
    TileShore2_100R1011  = makeShoreRiverTile(0x24A, 0b10001011, 0b00010000)
    TileShore2_101R0011  = makeShoreRiverTile(0x24B, 0b10100011, 0b00010000)
    TileShore2_101R1011  = makeShoreRiverTile(0x24C, 0b10101011, 0b00010000)
    TileShore2_100R0001  = makeShoreRiverTile(0x24D, 0b10000001, 0b00010000)
    TileShore2_100R1001  = makeShoreRiverTile(0x24E, 0b10001001, 0b00010000)
    TileShore2_101R0001  = makeShoreRiverTile(0x24F, 0b10100001, 0b00010000)
    TileShore2_101R1001  = makeShoreRiverTile(0x250, 0b10101001, 0b00010000)
    TileShore2_000R0011  = makeShoreRiverTile(0x251, 0b00000011, 0b00010000)
    TileShore2_000R1011  = makeShoreRiverTile(0x252, 0b00001011, 0b00010000)
    TileShore2_001R0011  = makeShoreRiverTile(0x253, 0b00100011, 0b00010000)
    TileShore2_001R1011  = makeShoreRiverTile(0x254, 0b00101011, 0b00010000)
    TileShore2_000R0001  = makeShoreRiverTile(0x255, 0b00000001, 0b00010000)
    TileShore2_000R1001  = makeShoreRiverTile(0x256, 0b00001001, 0b00010000)
    TileShore2_001R0001  = makeShoreRiverTile(0x257, 0b00100001, 0b00010000)
    TileShore2_001R1001  = makeShoreRiverTile(0x258, 0b00101001, 0b00010000)

    TileAnimOcean        = makeTile(Ocean, 0x0, makeCompatibilities(makeDirections(AllDirections), []TerrainType{Ocean}, AnyOf))

    TileTundra_00001000  = makeTundraTile(0x25A, 0b00001000)
    TileTundra_00001100  = makeTundraTile(0x25B, 0b00001100)
    TileTundra_00001110  = makeTundraTile(0x25C, 0b00001110)
    TileTundra_00000110  = makeTundraTile(0x25D, 0b00000110)
    TileTundra_00000010  = makeTundraTile(0x25E, 0b00000010)
    TileTundra_00001010  = makeTundraTile(0x25F, 0b00001010)
    TileTundra_00100010  = makeTundraTile(0x260, 0b00100010)
    TileTundra_10000010  = makeTundraTile(0x261, 0b10000010)
    TileTundra_00011000  = makeTundraTile(0x262, 0b00011000)
    TileTundra_00000100  = makeTundraTile(0x263, 0b00000100)
    TileTundra_00000011  = makeTundraTile(0x264, 0b00000011)
    TileTundra_10100000  = makeTundraTile(0x265, 0b10100000)
    TileTundra_10001000  = makeTundraTile(0x266, 0b10001000)
    TileTundra_00101000  = makeTundraTile(0x267, 0b00101000)
    TileTundra_00111000  = makeTundraTile(0x268, 0b00111000)
    TileTundra_00010000  = makeTundraTile(0x269, 0b00010000)
    TileTundra           = makeTundraTile(0x26A, 0b11111111)
    TileTundra_00000001  = makeTundraTile(0x26B, 0b00000001)
    TileTundra_10000011  = makeTundraTile(0x26C, 0b10000011)
    TileTundra_00110000  = makeTundraTile(0x26D, 0b00110000)
    TileTundra_01000000  = makeTundraTile(0x26E, 0b01000000)
    TileTundra_10000001  = makeTundraTile(0x26F, 0b10000001)
    TileTundra_10101000  = makeTundraTile(0x270, 0b10101000)
    TileTundra_00101010  = makeTundraTile(0x271, 0b00101010)
    TileTundra_10001010  = makeTundraTile(0x272, 0b10001010)
    TileTundra_00100000  = makeTundraTile(0x273, 0b00100000)
    TileTundra_01100000  = makeTundraTile(0x274, 0b01100000)
    TileTundra_11100000  = makeTundraTile(0x275, 0b11100000)
    TileTundra_11000000  = makeTundraTile(0x276, 0b11000000)
    TileTundra_10000000  = makeTundraTile(0x277, 0b10000000)
    TileTundra_10100010  = makeTundraTile(0x278, 0b10100010)
    TileTundra_10101010  = makeTundraTile(0x279, 0b10101010)
    TileTundra_11000001  = makeTundraTile(0x27A, 0b11000001)
    TileTundra_11100001  = makeTundraTile(0x27B, 0b11100001)
    TileTundra_11000011  = makeTundraTile(0x27C, 0b11000011)
    TileTundra_11100011  = makeTundraTile(0x27D, 0b11100011)
    TileTundra_00011100  = makeTundraTile(0x27E, 0b00011100)
    TileTundra_00111100  = makeTundraTile(0x27F, 0b00111100)
    TileTundra_00011110  = makeTundraTile(0x280, 0b00011110)
    TileTundra_00111110  = makeTundraTile(0x281, 0b00111110)
    TileTundra_01110000  = makeTundraTile(0x282, 0b01110000)
    TileTundra_01111000  = makeTundraTile(0x283, 0b01111000)
    TileTundra_11110000  = makeTundraTile(0x284, 0b11110000)
    TileTundra_11111000  = makeTundraTile(0x285, 0b11111000)
    TileTundra_00000111  = makeTundraTile(0x286, 0b00000111)
    TileTundra_00001111  = makeTundraTile(0x287, 0b00001111)
    TileTundra_10000111  = makeTundraTile(0x288, 0b10000111)
    TileTundra_10001111  = makeTundraTile(0x289, 0b10001111)
    TileTundra_11101110  = makeTundraTile(0x28A, 0b11101110)
    TileTundra_11100110  = makeTundraTile(0x28B, 0b11100110)
    TileTundra_11101100  = makeTundraTile(0x28C, 0b11101100)
    TileTundra_11100100  = makeTundraTile(0x28D, 0b11100100)
    TileTundra_11001110  = makeTundraTile(0x28E, 0b11001110)
    TileTundra_11000110  = makeTundraTile(0x28F, 0b11000110)
    TileTundra_11001100  = makeTundraTile(0x290, 0b11001100)
    TileTundra_11000100  = makeTundraTile(0x291, 0b11000100)
    TileTundra_01101110  = makeTundraTile(0x292, 0b01101110)
    TileTundra_01100110  = makeTundraTile(0x293, 0b01100110)
    TileTundra_01101100  = makeTundraTile(0x294, 0b01101100)
    TileTundra_01100100  = makeTundraTile(0x295, 0b01100100)
    TileTundra_01001110  = makeTundraTile(0x296, 0b01001110)
    TileTundra_01000110  = makeTundraTile(0x297, 0b01000110)
    TileTundra_01001100  = makeTundraTile(0x298, 0b01001100)
    TileTundra_01000100  = makeTundraTile(0x299, 0b01000100)
    TileTundra_10010011  = makeTundraTile(0x29A, 0b10010011)
    TileTundra_10011011  = makeTundraTile(0x29B, 0b10011011)
    TileTundra_10110011  = makeTundraTile(0x29C, 0b10110011)
    TileTundra_10111011  = makeTundraTile(0x29D, 0b10111011)
    TileTundra_10010001  = makeTundraTile(0x29E, 0b10010001)
    TileTundra_10011001  = makeTundraTile(0x29F, 0b10011001)
    TileTundra_10110001  = makeTundraTile(0x2A0, 0b10110001)
    TileTundra_10111001  = makeTundraTile(0x2A1, 0b10111001)
    TileTundra_00010011  = makeTundraTile(0x2A2, 0b00010011)
    TileTundra_00011011  = makeTundraTile(0x2A3, 0b00011011)
    TileTundra_00110011  = makeTundraTile(0x2A4, 0b00110011)
    TileTundra_00111011  = makeTundraTile(0x2A5, 0b00111011)
    TileTundra_00010001  = makeTundraTile(0x2A6, 0b00010001)
    TileTundra_00011001  = makeTundraTile(0x2A7, 0b00011001)
    TileTundra_00110001  = makeTundraTile(0x2A8, 0b00110001)
    TileTundra_00111001  = makeTundraTile(0x2A9, 0b00111001)
    TileTundra_00011111  = makeTundraTile(0x2AA, 0b00011111)
    TileTundra_11000111  = makeTundraTile(0x2AB, 0b11000111)
    TileTundra_11110001  = makeTundraTile(0x2AC, 0b11110001)
    TileTundra_01111100  = makeTundraTile(0x2AD, 0b01111100)
    TileTundra_10011111  = makeTundraTile(0x2AE, 0b10011111)
    TileTundra_11100111  = makeTundraTile(0x2AF, 0b11100111)
    TileTundra_11111001  = makeTundraTile(0x2B0, 0b11111001)
    TileTundra_01111110  = makeTundraTile(0x2B1, 0b01111110)
    TileTundra_00111111  = makeTundraTile(0x2B2, 0b00111111)
    TileTundra_11001111  = makeTundraTile(0x2B3, 0b11001111)
    TileTundra_11110011  = makeTundraTile(0x2B4, 0b11110011)
    TileTundra_11111100  = makeTundraTile(0x2B5, 0b11111100)
    TileTundra_10111111  = makeTundraTile(0x2B6, 0b10111111)
    TileTundra_11101111  = makeTundraTile(0x2B7, 0b11101111)
    TileTundra_11111011  = makeTundraTile(0x2B8, 0b11111011)
    TileTundra_11111110  = makeTundraTile(0x2B9, 0b11111110)
    TileTundra_10111000  = makeTundraTile(0x2BA, 0b10111000)
    TileTundra_10110000  = makeTundraTile(0x2BB, 0b10110000)
    TileTundra_10011000  = makeTundraTile(0x2BC, 0b10011000)
    TileTundra_10010000  = makeTundraTile(0x2BD, 0b10010000)
    TileTundra_10111010  = makeTundraTile(0x2BE, 0b10111010)
    TileTundra_10110010  = makeTundraTile(0x2BF, 0b10110010)
    TileTundra_10011010  = makeTundraTile(0x2C0, 0b10011010)
    TileTundra_10010010  = makeTundraTile(0x2C1, 0b10010010)
    TileTundra_00111010  = makeTundraTile(0x2C2, 0b00111010)
    TileTundra_00110010  = makeTundraTile(0x2C3, 0b00110010)
    TileTundra_00011010  = makeTundraTile(0x2C4, 0b00011010)
    TileTundra_00010010  = makeTundraTile(0x2C5, 0b00010010)
    TileTundra_10001110  = makeTundraTile(0x2C6, 0b10001110)
    TileTundra_10101110  = makeTundraTile(0x2C7, 0b10101110)
    TileTundra_00101110  = makeTundraTile(0x2C8, 0b00101110)
    TileTundra_10001100  = makeTundraTile(0x2C9, 0b10001100)
    TileTundra_10101100  = makeTundraTile(0x2CA, 0b10101100)
    TileTundra_00101100  = makeTundraTile(0x2CB, 0b00101100)
    TileTundra_10000110  = makeTundraTile(0x2CC, 0b10000110)
    TileTundra_10100110  = makeTundraTile(0x2CD, 0b10100110)
    TileTundra_00100110  = makeTundraTile(0x2CE, 0b00100110)
    TileTundra_10000100  = makeTundraTile(0x2CF, 0b10000100)
    TileTundra_10100100  = makeTundraTile(0x2D0, 0b10100100)
    TileTundra_00100100  = makeTundraTile(0x2D1, 0b00100100)
    TileTundra_00100001  = makeTundraTile(0x2D2, 0b00100001)
    TileTundra_10100001  = makeTundraTile(0x2D3, 0b10100001)
    TileTundra_00100011  = makeTundraTile(0x2D4, 0b00100011)
    TileTundra_10100011  = makeTundraTile(0x2D5, 0b10100011)
    TileTundra_00101001  = makeTundraTile(0x2D6, 0b00101001)
    TileTundra_10101001  = makeTundraTile(0x2D7, 0b10101001)
    TileTundra_00101011  = makeTundraTile(0x2D8, 0b00101011)
    TileTundra_10101011  = makeTundraTile(0x2D9, 0b10101011)
    TileTundra_00001001  = makeTundraTile(0x2DA, 0b00001001)
    TileTundra_10001001  = makeTundraTile(0x2DB, 0b10001001)
    TileTundra_00001011  = makeTundraTile(0x2DC, 0b00001011)
    TileTundra_10001011  = makeTundraTile(0x2DD, 0b10001011)
    TileTundra_01000010  = makeTundraTile(0x2DE, 0b01000010)
    TileTundra_01001010  = makeTundraTile(0x2DF, 0b01001010)
    TileTundra_01001000  = makeTundraTile(0x2E0, 0b01001000)
    TileTundra_11000010  = makeTundraTile(0x2E1, 0b11000010)
    TileTundra_11001010  = makeTundraTile(0x2E2, 0b11001010)
    TileTundra_11001000  = makeTundraTile(0x2E3, 0b11001000)
    TileTundra_01100010  = makeTundraTile(0x2E4, 0b01100010)
    TileTundra_01101010  = makeTundraTile(0x2E5, 0b01101010)
    TileTundra_01101000  = makeTundraTile(0x2E6, 0b01101000)
    TileTundra_11100010  = makeTundraTile(0x2E7, 0b11100010)
    TileTundra_11101010  = makeTundraTile(0x2E8, 0b11101010)
    TileTundra_11101000  = makeTundraTile(0x2E9, 0b11101000)
    TileTundra_11001001  = makeTundraTile(0x2EA, 0b11001001)
    TileTundra_11101001  = makeTundraTile(0x2EB, 0b11101001)
    TileTundra_11001011  = makeTundraTile(0x2EC, 0b11001011)
    TileTundra_11101011  = makeTundraTile(0x2ED, 0b11101011)
    TileTundra_10011100  = makeTundraTile(0x2EE, 0b10011100)
    TileTundra_10111100  = makeTundraTile(0x2EF, 0b10111100)
    TileTundra_10011110  = makeTundraTile(0x2F0, 0b10011110)
    TileTundra_10111110  = makeTundraTile(0x2F1, 0b10111110)
    TileTundra_01110010  = makeTundraTile(0x2F2, 0b01110010)
    TileTundra_01111010  = makeTundraTile(0x2F3, 0b01111010)
    TileTundra_11110010  = makeTundraTile(0x2F4, 0b11110010)
    TileTundra_11111010  = makeTundraTile(0x2F5, 0b11111010)
    TileTundra_00100111  = makeTundraTile(0x2F6, 0b00100111)
    TileTundra_00101111  = makeTundraTile(0x2F7, 0b00101111)
    TileTundra_10100111  = makeTundraTile(0x2F8, 0b10100111)
    TileTundra_10101111  = makeTundraTile(0x2F9, 0b10101111)
)

type TerrainData struct {
    // the full array of all tile images
    Images []image.Image
    Tiles []TerrainTile
    // store a map from the terrain type to the tiles that match that type
    OnlyTiles map[TerrainType][]TerrainTile
}

func MakeTerrainData(images []image.Image, tiles []TerrainTile) *TerrainData {
    out := &TerrainData{
        Images: images,
        Tiles: tiles,
    }

    out.optimize()

    return out
}

func (data *TerrainData) optimize() {
    data.OnlyTiles = make(map[TerrainType][]TerrainTile)

    for _, tile := range data.Tiles {
        tiles := data.OnlyTiles[tile.Tile.TerrainType()]
        data.OnlyTiles[tile.Tile.TerrainType()] = append(tiles, tile)
    }
}

func (data *TerrainData) TileWidth() int {
    if len(data.Images) > 0 {
        return data.Images[0].Bounds().Dx()
    }

    return 0
}

func (data *TerrainData) TileHeight() int {
    if len(data.Images) > 0 {
        return data.Images[0].Bounds().Dy()
    }

    return 0
}

// returns an array of tile indicies that match the given map
func (data *TerrainData) FindMatchingAllTiles(match map[Direction]TerrainType, plane data.Plane) []int {
    var out []int
    for i, tile := range data.Tiles {
        if tile.IsPlane(plane) && tile.Tile.Matches(match) {
            out = append(out, i)
        }
    }

    return out
}

func (terrain *TerrainData) FindMatchingTile(match map[Direction]TerrainType, plane data.Plane) int {
    var tiles []TerrainTile

    center, ok := match[Center]
    if ok {
        for _, tile := range terrain.OnlyTiles[center] {
            if tile.IsPlane(plane) && tile.Tile.Matches(match) {
                return tile.TileIndex
            }
        }
    }

    switch plane {
        case data.PlaneMyrror:
            tiles = terrain.Tiles[MyrrorStart:]
        case data.PlaneArcanus:
            tiles = terrain.Tiles[:MyrrorStart]
    }

    for _, tile := range tiles {
        if tile.Tile.Matches(match) {
            return tile.TileIndex
        }
    }

    return -1
}

type TerrainTile struct {
    // the index into the original array of images, if needed
    ImageIndex int
    // the index of the tile, useful for indexing into the terrain metadata
    TileIndex int
    // the images associated with this tile. for non-animated tiles this will be of length 1
    // for animated tiles this will be length 4
    Images []image.Image
    Tile Tile
}

func (tile *TerrainTile) IsPlane(plane data.Plane) bool {
    if tile.IsMyrror() && plane == data.PlaneMyrror {
        return true
    }

    if tile.IsArcanus() && plane == data.PlaneArcanus {
        return true
    }

    return false
}

func (tile *TerrainTile) IsMyrror() bool {
    return tile.TileIndex >= MyrrorStart
}

func (tile *TerrainTile) IsArcanus() bool {
    return tile.TileIndex < MyrrorStart
}

func (tile *TerrainTile) ContainsImageIndex(index int) bool {
    return index >= tile.ImageIndex && index < tile.ImageIndex + len(tile.Images)
}

// pass in terrain.lbx
func ReadTerrainData(lbxFile *lbx.LbxFile) (*TerrainData, error) {
    data, err := lbxFile.RawData(1)
    if err != nil {
        return nil, err
    }

    images, err := lbxFile.ReadTerrainImages(0)
    if err != nil {
        return nil, err
    }

    // TODO: lbxFile entry 2 is the terrain palette for the minimap

    reader := bytes.NewReader(data)

    // larger than we need, but should be fine
    tiles := make([]TerrainTile, 0, len(images))

    tileIndex := 0

    for reader.Len() > 0 {
        var animation = false
        value1, err := reader.ReadByte()
        if err != nil {
            return nil, err
        }

        value2, err := reader.ReadByte()
        if err != nil {
            return nil, err
        }

        if value1 & 0x80 != 0 {
            value1 &= 0x7f
            animation = true
            // value2 should be incremented based on an animation counter
        }

        // the index into the terrain image
        var index int

        if value1 == 0 {
            index = int(value2) - 2
        } else {
            // this formula comes from
            // terrain_lbx_000_offset = (terrain_001_0 * 16384) + (terrain_001_1 * 384) - 0xC0 - 384;
            // this only works if value1 is a multiple of 3
            // 3 -> 126, 6 -> 254, 9 -> 382
            index = int(value1) * 16384 / 384 + int(value2) - 2
        }

        var tileImages []image.Image
        if animation {
            // animation tiles are always 4 images
            for i := 0; i < 4; i++ {
                tileImages = append(tileImages, images[index + i])
            }
        } else {
            tileImages = append(tileImages, images[index])
        }

        tiles = append(tiles, TerrainTile{
            ImageIndex: index,
            TileIndex: tileIndex,
            Tile: GetTile(tileIndex),
            Images: tileImages,
        })

        tileIndex += 1
    }

    return MakeTerrainData(images, tiles), nil
}
