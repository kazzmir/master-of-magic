package terrain

import (
    "bytes"
    "image"

    "github.com/kazzmir/master-of-magic/lib/lbx"
)

// terrain tiles are indicies 0-0x259 for arcanus, and 0x25A - 0x5f4 for myrror

type TerrainIndex int

const (
    IndexOcean1 TerrainIndex      = 0x0
    IndexBugGrass    = 0x1
    IndexShore1_1st  = 0x2
    IndexLake        = 0x12
    IndexShore1_end  = 0x0A1
    IndexGrass1      = 0x0A2
    IndexForest1     = 0x0A3
    IndexMountain1   = 0x0A4
    IndexDesert1     = 0x0A5
    IndexSwamp1      = 0x0A6
    IndexTundra1     = 0x0A7
    IndexSorcNode    = 0x0A8
    IndexNatNode     = 0x0A9
    IndexChaosNode   = 0x0AA
    IndexHills1      = 0x0AB
    IndexGrass2      = 0x0AC
    IndexGrass3      = 0x0AD
    IndexDesert2     = 0x0AE
    IndexDesert3     = 0x0AF
    IndexDesert4     = 0x0B0
    IndexSwamp2      = 0x0B1
    IndexSwamp3      = 0x0B2
    IndexVolcano     = 0x0B3
    IndexGrass4      = 0x0B4
    IndexTundra2     = 0x0B5
    IndexTundra3     = 0x0B6
    IndexForest2     = 0x0B7
    IndexForest3     = 0x0B8
    IndexRiverMStart  = 0x0B9
    IndexRiverMEnd  = 0x0C4
    IndexLake1       = 0x0C5
    IndexLake2       = 0x0C6
    IndexLake3       = 0x0C7
    IndexLake4       = 0x0C8
    IndexShore2FStart = 0x0C9
    IndexShore2FEnd = 0x0E8
    IndexRiversStart  = 0x0E9
    IndexRiversEnd  = 0x102
    IndexMountainsStart   = 0x103
    IndexMountainsEnd = 0x112
    IndexHillsStart   = 0x113
    IndexHillsEnd   = 0x123
    IndexDesertStart  = 0x124
    IndexDesertEnd  = 0x1C3
    IndexShore2Start  = 0x1C4
    IndexShore2End  = 0x1D3
    Index4WRiver1    = 0x1D4
    Index4WRiver2    = 0x1D5
    Index4WRiver3    = 0x1D6
    Index4WRiver4    = 0x1D7
    Index4WRiver5    = 0x1D8
    IndexShore3Start  = 0x1D9
    IndexShore3End  = 0x258
    IndexOcean2      = 0x259
    IndexTundra_1st  = 0x25A
    IndexTundra_Last = 0x2F9
)

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
)

type Tile {
    // index into the TerrainTile array
    Index int
    Directions []Direction
}

func (tile Tile) String() string {
    return fmt.Sprintf("Tile{Index: %d, Directions: %v}", tile.Index, tile.Directions)
}

func makeTile(index int, bitPattern uint8) Tile {
    var directions []Direction

    // bit 0: north west
    // bit 1: north
    // bit 2: north east
    // bit 3: east
    // bit 4: south east
    // bit 5: south
    // bit 6: south west
    // bit 7: west

    choices := []Direction{West, SouthWest, South, SouthEast, East, NorthEast, North, NorthWest}
    for i, choice := range choices {
        if bitPattern & (1 << i) != 0 {
            directions = append(directions, choice)
        }
    }

    return Tile{
        Index: index,
        Directions: directions,
    }
}

var (
    TileOcean = makeTile(0x0, 0)
    TileLand = makeTile(0x1, 0b1111_1111)
    TileShore1_00001000 = makeTile(0x2, 0b00001000)
    TileShore1_00001100 = makeTile(0x3, 0b00001100)

    // _Shore00001110   = 0x4,
    // _Shore00000110   = 0x5,
    // _Shore00000010   = 0x6,
    // _Shore00001010   = 0x7,
    // _Shore00100010   = 0x8,
    // _Shore10000010   = 0x9,
    // _Shore00011000   = 0x0A,
    // _Shore00000100   = 0x0B,
    // _Shore00000011   = 0x0C,
    // _Shore10100000   = 0x0D,
    // _Shore10001000   = 0x0E,
    // _Shore00101000   = 0x0F,
    // _Shore00111000   = 0x10,
    // _Shore00010000   = 0x11,

)

// a bit pattern on a tile indicates the positions where the tile can match up with another tile
// bit index: 0123 4567
//            0000 1000
//
// bit 0: north west
// bit 1: north
// bit 2: north east
// bit 3: east
// bit 4: south east
// bit 5: south
// bit 6: south west
// bit 7: west
/*
enum OVL_Tiles_Extended
{
    _Ocean           = 0x0,
    _Land            = 0x1,
    _Shore00001000   = 0x2,
    _Shore00001100   = 0x3,
    _Shore00001110   = 0x4,
    _Shore00000110   = 0x5,
    _Shore00000010   = 0x6,
    _Shore00001010   = 0x7,
    _Shore00100010   = 0x8,
    _Shore10000010   = 0x9,
    _Shore00011000   = 0x0A,
    _Shore00000100   = 0x0B,
    _Shore00000011   = 0x0C,
    _Shore10100000   = 0x0D,
    _Shore10001000   = 0x0E,
    _Shore00101000   = 0x0F,
    _Shore00111000   = 0x10,
    _Shore00010000   = 0x11,
    _1Lake           = 0x12,
    _Shore00000001   = 0x13,
    _Shore10000011   = 0x14,
    _Shore00110000   = 0x15,
    _Shore01000000   = 0x16,
    _Shore10000001   = 0x17,
    _Shore10101000   = 0x18,
    _Shore00101010   = 0x19,
    _Shore10001010   = 0x1A,
    _Shore00100000   = 0x1B,
    _Shore01100000   = 0x1C,
    _Shore11100000   = 0x1D,
    _Shore11000000   = 0x1E,
    _Shore10000000   = 0x1F,
    _Shore10100010   = 0x20,
    _Shore10101010   = 0x21,
    _Shore11000001   = 0x22,
    _Shore11100001   = 0x23,
    _Shore11000011   = 0x24,
    _Shore11100011   = 0x25,
    _Shore00011100   = 0x26,
    _Shore00111100   = 0x27,
    _Shore00011110   = 0x28,
    _Shore00111110   = 0x29,
    _Shore01110000   = 0x2A,
    _Shore01111000   = 0x2B,
    _Shore11110000   = 0x2C,
    _Shore11111000   = 0x2D,
    _Shore00000111   = 0x2E,
    _Shore00001111   = 0x2F,
    _Shore10000111   = 0x30,
    _Shore10001111   = 0x31,
    _Shore11101110   = 0x32,
    _Shore11100110   = 0x33,
    _Shore11101100   = 0x34,
    _Shore11100100   = 0x35,
    _Shore11001110   = 0x36,
    _Shore11000110   = 0x37,
    _Shore11001100   = 0x38,
    _Shore11000100   = 0x39,
    _Shore01101110   = 0x3A,
    _Shore01100110   = 0x3B,
    _Shore01101100   = 0x3C,
    _Shore01100100   = 0x3D,
    _Shore01001110   = 0x3E,
    _Shore01000110   = 0x3F,
    _Shore01001100   = 0x40,
    _Shore01000100   = 0x41,
    _Shore10010011   = 0x42,
    _Shore10011011   = 0x43,
    _Shore10110011   = 0x44,
    _Shore10111011   = 0x45,
    _Shore10010001   = 0x46,
    _Shore10011001   = 0x47,
    _Shore10110001   = 0x48,
    _Shore10111001   = 0x49,
    _Shore00010011   = 0x4A,
    _Shore00011011   = 0x4B,
    _Shore00110011   = 0x4C,
    _Shore00111011   = 0x4D,
    _Shore00010001   = 0x4E,
    _Shore00011001   = 0x4F,
    _Shore00110001   = 0x50,
    _Shore00111001   = 0x51,
    _Shore00011111   = 0x52,
    _Shore11000111   = 0x53,
    _Shore11110001   = 0x54,
    _Shore01111100   = 0x55,
    _Shore10011111   = 0x56,
    _Shore11100111   = 0x57,
    _Shore11111001   = 0x58,
    _Shore01111110   = 0x59,
    _Shore00111111   = 0x5A,
    _Shore11001111   = 0x5B,
    _Shore11110011   = 0x5C,
    _Shore11111100   = 0x5D,
    _Shore10111111   = 0x5E,
    _Shore11101111   = 0x5F,
    _Shore11111011   = 0x60,
    _Shore11111110   = 0x61,
    _Shore10111000   = 0x62,
    _Shore10110000   = 0x63,
    _Shore10011000   = 0x64,
    _Shore10010000   = 0x65,
    _Shore10111010   = 0x66,
    _Shore10110010   = 0x67,
    _Shore10011010   = 0x68,
    _Shore10010010   = 0x69,
    _Shore00111010   = 0x6A,
    _Shore00110010   = 0x6B,
    _Shore00011010   = 0x6C,
    _Shore00010010   = 0x6D,
    _Shore10001110   = 0x6E,
    _Shore10101110   = 0x6F,
    _Shore00101110   = 0x70,
    _Shore10001100   = 0x71,
    _Shore10101100   = 0x72,
    _Shore00101100   = 0x73,
    _Shore10000110   = 0x74,
    _Shore10100110   = 0x75,
    _Shore00100110   = 0x76,
    _Shore10000100   = 0x77,
    _Shore10100100   = 0x78,
    _Shore00100100   = 0x79,
    _Shore00100001   = 0x7A,
    _Shore10100001   = 0x7B,
    _Shore00100011   = 0x7C,
    _Shore10100011   = 0x7D,
    _Shore00101001   = 0x7E,
    _Shore10101001   = 0x7F,
    _Shore00101011   = 0x80,
    _Shore10101011   = 0x81,
    _Shore00001001   = 0x82,
    _Shore10001001   = 0x83,
    _Shore00001011   = 0x84,
    _Shore10001011   = 0x85,
    _Shore01000010   = 0x86,
    _Shore01001010   = 0x87,
    _Shore01001000   = 0x88,
    _Shore11000010   = 0x89,
    _Shore11001010   = 0x8A,
    _Shore11001000   = 0x8B,
    _Shore01100010   = 0x8C,
    _Shore01101010   = 0x8D,
    _Shore01101000   = 0x8E,
    _Shore11100010   = 0x8F,
    _Shore11101010   = 0x90,
    _Shore11101000   = 0x91,
    _Shore11001001   = 0x92,
    _Shore11101001   = 0x93,
    _Shore11001011   = 0x94,
    _Shore11101011   = 0x95,
    _Shore10011100   = 0x96,
    _Shore10111100   = 0x97,
    _Shore10011110   = 0x98,
    _Shore10111110   = 0x99,
    _Shore01110010   = 0x9A,
    _Shore01111010   = 0x9B,
    _Shore11110010   = 0x9C,
    _Shore11111010   = 0x9D,
    _Shore00100111   = 0x9E,
    _Shore00101111   = 0x9F,
    _Shore10100111   = 0xA0,
    _Shore10101111   = 0xA1,
    _Grasslands1     = 0xA2,
    _Forest1         = 0xA3,
    _1Mountain1      = 0xA4,
    _AllDesert1      = 0xA5,
    _Swamp1          = 0xA6,
    _AllTundra1      = 0xA7,
    _SorceryLake     = 0xA8,
    _NatureForest    = 0xA9,
    _ChaosVolcano    = 0xAA,
    _1Hills1         = 0xAB,
    _Grasslands2     = 0xAC,
    _Grasslands3     = 0xAD,
    _AllDesert2      = 0xAE,
    _AllDesert3      = 0xAF,
    _AllDesert4      = 0xB0,
    _Swamp2          = 0xB1,
    _Swamp3          = 0xB2,
    _Volcano         = 0xB3,
    _Grasslands4     = 0xB4,
    _AllTundra2      = 0xB5,
    _AllTundra3      = 0xB6,
    _Forest2         = 0xB7,
    _Forest3         = 0xB8,
    _River0010       = 0xB9,
    _River0001       = 0xBA,
    _River1000       = 0xBB,
    _River0100       = 0xBC,
    _River1100_1     = 0xBD,
    _River0011_1     = 0xBE,
    _River0110_1     = 0xBF,
    _River1001_1     = 0xC0,
    _River1100_2     = 0xC1,
    _River0011_2     = 0xC2,
    _River0110_2     = 0xC3,
    _River1001_2     = 0xC4,
    _1LakeRiv_W      = 0xC5,
    _1LakeRiv_N      = 0xC6,
    _1LakeRiv_E      = 0xC7,
    _1LakeRiv_S      = 0xC8,
    _Shore1R00000R   = 0xC9,
    _Shore1R10000R   = 0xCA,
    _Shore1R00001R   = 0xCB,
    _Shore1R10001R   = 0xCC,
    _Shore000R1R00   = 0xCD,
    _Shore000R1R10   = 0xCE,
    _Shore001R1R00   = 0xCF,
    _Shore001R1R10   = 0xD0,
    _Shore0R1R0000   = 0xD1,
    _Shore0R1R1000   = 0xD2,
    _Shore1R1R0000   = 0xD3,
    _Shore1R1R1000   = 0xD4,
    _Shore00000R1R   = 0xD5,
    _Shore00001R1R   = 0xD6,
    _Shore10000R1R   = 0xD7,
    _Shore10001R1R   = 0xD8,
    _Shore00001R10   = 0xD9,
    _Shore00001R00   = 0xDA,
    _Shore00000R10   = 0xDB,
    _Shore00000R00   = 0xDC,
    _Shore1000001R   = 0xDD,
    _Shore0000001R   = 0xDE,
    _Shore1000000R   = 0xDF,
    _Shore0000000R   = 0xE0,
    _Shore1R100000   = 0xE1,
    _Shore1R000000   = 0xE2,
    _Shore0R100000   = 0xE3,
    _Shore0R000000   = 0xE4,
    _Shore001R1000   = 0xE5,
    _Shore001R0000   = 0xE6,
    _Shore000R1000   = 0xE7,
    _Shore000R0000   = 0xE8,
    _River1100_3     = 0xE9,
    _River0011_3     = 0xEA,
    _River0110_3     = 0xEB,
    _River1001_3     = 0xEC,
    _River1010_1     = 0xED,
    _River1010_2     = 0xEE,
    _River1010_3     = 0xEF,
    _River0101_1     = 0xF0,
    _River0101_2     = 0xF1,
    _River0101_3     = 0xF2,
    _River1101_1     = 0xF3,
    _River1101_2     = 0xF4,
    _River1101_3     = 0xF5,
    _River1101_4     = 0xF6,
    _River0111_1     = 0xF7,
    _River0111_2     = 0xF8,
    _River0111_3     = 0xF9,
    _River0111_4     = 0xFA,
    _River1110_1     = 0xFB,
    _River1110_2     = 0xFC,
    _River1110_3     = 0xFD,
    _River1110_4     = 0xFE,
    _River1011_1     = 0xFF,
    _River1011_2     = 0x100,
    _River1011_3     = 0x101,
    _River1011_4     = 0x102,
    _Mount0010       = 0x103,
    _Mount0100       = 0x104,
    _Mnt1_1111       = 0x105,
    _Mount0101       = 0x106,
    _Mount0001       = 0x107,
    _Mount1010       = 0x108,
    _Mount1000       = 0x109,
    _Mount0110       = 0x10A,
    _Mount0111       = 0x10B,
    _Mount0011       = 0x10C,
    _Mount1110       = 0x10D,
    _Mnt2_1111       = 0x10E,
    _Mount1011       = 0x10F,
    _Mount1100       = 0x110,
    _Mount1101       = 0x111,
    _Mount1001       = 0x112,
    _Hills_0010      = 0x113,
    _Hills_0100      = 0x114,
    _Hill1_1111      = 0x115,
    _Hills_0101      = 0x116,
    _Hills_0001      = 0x117,
    _Hills_1010      = 0x118,
    _Hills_1000      = 0x119,
    _Hills_0110      = 0x11A,
    _Hills_0111      = 0x11B,
    _Hills_0011      = 0x11C,
    _Hills_1110      = 0x11D,
    _Hill2_1111      = 0x11E,
    _Hills_1011      = 0x11F,
    _Hills_1100      = 0x120,
    _Hills_1101      = 0x121,
    _Hills_1001      = 0x122,
    _1Hills2         = 0x123,
    _Desert00001000  = 0x124,
    _Desert00001100  = 0x125,
    _Desert00001110  = 0x126,
    _Desert00000110  = 0x127,
    _Desert00000010  = 0x128,
    _Desert00001010  = 0x129,
    _Desert00100010  = 0x12A,
    _Desert10000010  = 0x12B,
    _Desert00011000  = 0x12C,
    _Desert00000100  = 0x12D,
    _Desert00000011  = 0x12E,
    _Desert10100000  = 0x12F,
    _Desert10001000  = 0x130,
    _Desert00101000  = 0x131,
    _Desert00111000  = 0x132,
    _Desert00010000  = 0x133,
    _1Desert         = 0x134,
    _Desert00000001  = 0x135,
    _Desert10000011  = 0x136,
    _Desert00110000  = 0x137,
    _Desert01000000  = 0x138,
    _Desert10000001  = 0x139,
    _Desert10101000  = 0x13A,
    _Desert00101010  = 0x13B,
    _Desert10001010  = 0x13C,
    _Desert00100000  = 0x13D,
    _Desert01100000  = 0x13E,
    _Desert11100000  = 0x13F,
    _Desert11000000  = 0x140,
    _Desert10000000  = 0x141,
    _Desert10100010  = 0x142,
    _Desert10101010  = 0x143,
    _Desert11000001  = 0x144,
    _Desert11100001  = 0x145,
    _Desert11000011  = 0x146,
    _Desert11100011  = 0x147,
    _Desert00011100  = 0x148,
    _Desert00111100  = 0x149,
    _Desert00011110  = 0x14A,
    _Desert00111110  = 0x14B,
    _Desert01110000  = 0x14C,
    _Desert01111000  = 0x14D,
    _Desert11110000  = 0x14E,
    _Desert11111000  = 0x14F,
    _Desert00000111  = 0x150,
    _Desert00001111  = 0x151,
    _Desert10000111  = 0x152,
    _Desert10001111  = 0x153,
    _Desert11101110  = 0x154,
    _Desert11100110  = 0x155,
    _Desert11101100  = 0x156,
    _Desert11100100  = 0x157,
    _Desert11001110  = 0x158,
    _Desert11000110  = 0x159,
    _Desert11001100  = 0x15A,
    _Desert11000100  = 0x15B,
    _Desert01101110  = 0x15C,
    _Desert01100110  = 0x15D,
    _Desert01101100  = 0x15E,
    _Desert01100100  = 0x15F,
    _Desert01001110  = 0x160,
    _Desert01000110  = 0x161,
    _Desert01001100  = 0x162,
    _Desert01000100  = 0x163,
    _Desert10010011  = 0x164,
    _Desert10011011  = 0x165,
    _Desert10110011  = 0x166,
    _Desert10111011  = 0x167,
    _Desert10010001  = 0x168,
    _Desert10011001  = 0x169,
    _Desert10110001  = 0x16A,
    _Desert10111001  = 0x16B,
    _Desert00010011  = 0x16C,
    _Desert00011011  = 0x16D,
    _Desert00110011  = 0x16E,
    _Desert00111011  = 0x16F,
    _Desert00010001  = 0x170,
    _Desert00011001  = 0x171,
    _Desert00110001  = 0x172,
    _Desert00111001  = 0x173,
    _Desert00011111  = 0x174,
    _Desert11000111  = 0x175,
    _Desert11110001  = 0x176,
    _Desert01111100  = 0x177,
    _Desert10011111  = 0x178,
    _Desert11100111  = 0x179,
    _Desert11111001  = 0x17A,
    _Desert01111110  = 0x17B,
    _Desert00111111  = 0x17C,
    _Desert11001111  = 0x17D,
    _Desert11110011  = 0x17E,
    _Desert11111100  = 0x17F,
    _Desert10111111  = 0x180,
    _Desert11101111  = 0x181,
    _Desert11111011  = 0x182,
    _Desert11111110  = 0x183,
    _Desert10111000  = 0x184,
    _Desert10110000  = 0x185,
    _Desert10011000  = 0x186,
    _Desert10010000  = 0x187,
    _Desert10111010  = 0x188,
    _Desert10110010  = 0x189,
    _Desert10011010  = 0x18A,
    _Desert10010010  = 0x18B,
    _Desert00111010  = 0x18C,
    _Desert00110010  = 0x18D,
    _Desert00011010  = 0x18E,
    _Desert00010010  = 0x18F,
    _Desert10001110  = 0x190,
    _Desert10101110  = 0x191,
    _Desert00101110  = 0x192,
    _Desert10001100  = 0x193,
    _Desert10101100  = 0x194,
    _Desert00101100  = 0x195,
    _Desert10000110  = 0x196,
    _Desert10100110  = 0x197,
    _Desert00100110  = 0x198,
    _Desert10000100  = 0x199,
    _Desert10100100  = 0x19A,
    _Desert00100100  = 0x19B,
    _Desert00100001  = 0x19C,
    _Desert10100001  = 0x19D,
    _Desert00100011  = 0x19E,
    _Desert10100011  = 0x19F,
    _Desert00101001  = 0x1A0,
    _Desert10101001  = 0x1A1,
    _Desert00101011  = 0x1A2,
    _Desert10101011  = 0x1A3,
    _Desert00001001  = 0x1A4,
    _Desert10001001  = 0x1A5,
    _Desert00001011  = 0x1A6,
    _Desert10001011  = 0x1A7,
    _Desert01000010  = 0x1A8,
    _Desert01001010  = 0x1A9,
    _Desert01001000  = 0x1AA,
    _Desert11000010  = 0x1AB,
    _Desert11001010  = 0x1AC,
    _Desert11001000  = 0x1AD,
    _Desert01100010  = 0x1AE,
    _Desert01101010  = 0x1AF,
    _Desert01101000  = 0x1B0,
    _Desert11100010  = 0x1B1,
    _Desert11101010  = 0x1B2,
    _Desert11101000  = 0x1B3,
    _Desert11001001  = 0x1B4,
    _Desert11101001  = 0x1B5,
    _Desert11001011  = 0x1B6,
    _Desert11101011  = 0x1B7,
    _Desert10011100  = 0x1B8,
    _Desert10111100  = 0x1B9,
    _Desert10011110  = 0x1BA,
    _Desert10111110  = 0x1BB,
    _Desert01110010  = 0x1BC,
    _Desert01111010  = 0x1BD,
    _Desert11110010  = 0x1BE,
    _Desert11111010  = 0x1BF,
    _Desert00100111  = 0x1C0,
    _Desert00101111  = 0x1C1,
    _Desert10100111  = 0x1C2,
    _Desert10101111  = 0x1C3,
    _Shore00011R11  = 0x1C4,
    _Shore1100011R  = 0x1C5,
    _Shore1R110001  = 0x1C6,
    _Shore011R1100  = 0x1C7,
    _Shore10011R11  = 0x1C8,
    _Shore1110011R  = 0x1C9,
    _Shore1R111001  = 0x1CA,
    _Shore011R1110  = 0x1CB,
    _Shore00111R11  = 0x1CC,
    _Shore1100111R  = 0x1CD,
    _Shore1R110011  = 0x1CE,
    _Shore111R1100  = 0x1CF,
    _Shore10111R11  = 0x1D0,
    _Shore1110111R  = 0x1D1,
    _Shore1R111011  = 0x1D2,
    _Shore111R1110  = 0x1D3,
    _River1111_1     = 0x1D4,
    _River1111_2     = 0x1D5,
    _River1111_3     = 0x1D6,
    _River1111_4     = 0x1D7,
    _River1111_5     = 0x1D8,
    _Shore1100000R  = 0x1D9,
    _Shore1110000R  = 0x1DA,
    _Shore1100001R  = 0x1DB,
    _Shore1110001R  = 0x1DC,
    _Shore00011R00  = 0x1DD,
    _Shore00111R00  = 0x1DE,
    _Shore00011R10  = 0x1DF,
    _Shore00111R10  = 0x1E0,
    _Shore0R110000  = 0x1E1,
    _Shore0R111000  = 0x1E2,
    _Shore1R110000  = 0x1E3,
    _Shore1R111000  = 0x1E4,
    _Shore00000R11  = 0x1E5,
    _Shore00001R11  = 0x1E6,
    _Shore10000R11  = 0x1E7,
    _Shore10001R11  = 0x1E8,
    _Shore1R000001  = 0x1E9,
    _Shore1R100001  = 0x1EA,
    _Shore1R000011  = 0x1EB,
    _Shore1R100011  = 0x1EC,
    _Shore000R1100  = 0x1ED,
    _Shore001R1100  = 0x1EE,
    _Shore000R1110  = 0x1EF,
    _Shore001R1110  = 0x1F0,
    _Shore011R0000  = 0x1F1,
    _Shore011R1000  = 0x1F2,
    _Shore111R0000  = 0x1F3,
    _Shore111R1000  = 0x1F4,
    _Shore0000011R  = 0x1F5,
    _Shore0000111R  = 0x1F6,
    _Shore1000011R  = 0x1F7,
    _Shore1000111R  = 0x1F8,
    _Shore0001111R  = 0x1F9,
    _Shore1R000111  = 0x1FA,
    _Shore111R0001  = 0x1FB,
    _Shore0R111100  = 0x1FC,
    _Shore1001111R  = 0x1FD,
    _Shore1R100111  = 0x1FE,
    _Shore111R1001  = 0x1FF,
    _Shore0R111110  = 0x200,
    _Shore0011111R  = 0x201,
    _Shore1R001111  = 0x202,
    _Shore111R0011  = 0x203,
    _Shore1R111100  = 0x204,
    _Shore1011111R  = 0x205,
    _Shore1R101111  = 0x206,
    _Shore111R1011  = 0x207,
    _Shore1R111110  = 0x208,
    _Shore000R1111  = 0x209,
    _Shore11000R11  = 0x20A,
    _Shore1111000R  = 0x20B,
    _Shore01111R00  = 0x20C,
    _Shore100R1111  = 0x20D,
    _Shore11100R11  = 0x20E,
    _Shore1111100R  = 0x20F,
    _Shore01111R10  = 0x210,
    _Shore001R1111  = 0x211,
    _Shore11001R11  = 0x212,
    _Shore1111001R  = 0x213,
    _Shore11111R00  = 0x214,
    _Shore101R1111  = 0x215,
    _Shore11101R11  = 0x216,
    _Shore1111101R  = 0x217,
    _Shore11111R10  = 0x218,
    _Shore1R101110  = 0x219,
    _Shore1R100110  = 0x21A,
    _Shore1R101100  = 0x21B,
    _Shore1R100100  = 0x21C,
    _Shore1R001110  = 0x21D,
    _Shore1R000110  = 0x21E,
    _Shore1R001100  = 0x21F,
    _Shore1R000100  = 0x220,
    _Shore0R101110  = 0x221,
    _Shore0R100110  = 0x222,
    _Shore0R101100  = 0x223,
    _Shore0R100100  = 0x224,
    _Shore0R001110  = 0x225,
    _Shore0R000110  = 0x226,
    _Shore0R001100  = 0x227,
    _Shore0R000100  = 0x228,
    _Shore11101R10  = 0x229,
    _Shore11100R10  = 0x22A,
    _Shore11101R00  = 0x22B,
    _Shore11100R00  = 0x22C,
    _Shore11001R10  = 0x22D,
    _Shore11000R10  = 0x22E,
    _Shore11001R00  = 0x22F,
    _Shore11000R00  = 0x230,
    _Shore01101R10  = 0x231,
    _Shore01100R10  = 0x232,
    _Shore01101R00  = 0x233,
    _Shore01100R00  = 0x234,
    _Shore01001R10  = 0x235,
    _Shore01000R10  = 0x236,
    _Shore01001R00  = 0x237,
    _Shore01000R00  = 0x238,
    _Shore1001001R  = 0x239,
    _Shore1001101R  = 0x23A,
    _Shore1011001R  = 0x23B,
    _Shore1011101R  = 0x23C,
    _Shore1001000R  = 0x23D,
    _Shore1001100R  = 0x23E,
    _Shore1011000R  = 0x23F,
    _Shore1011100R  = 0x240,
    _Shore0001001R  = 0x241,
    _Shore0001101R  = 0x242,
    _Shore0011001R  = 0x243,
    _Shore0011101R  = 0x244,
    _Shore0001000R  = 0x245,
    _Shore0001100R  = 0x246,
    _Shore0011000R  = 0x247,
    _Shore0011100R  = 0x248,
    _Shore100R0011  = 0x249,
    _Shore100R1011  = 0x24A,
    _Shore101R0011  = 0x24B,
    _Shore101R1011  = 0x24C,
    _Shore100R0001  = 0x24D,
    _Shore100R1001  = 0x24E,
    _Shore101R0001  = 0x24F,
    _Shore101R1001  = 0x250,
    _Shore000R0011  = 0x251,
    _Shore000R1011  = 0x252,
    _Shore001R0011  = 0x253,
    _Shore001R1011  = 0x254,
    _Shore000R0001  = 0x255,
    _Shore000R1001  = 0x256,
    _Shore001R0001  = 0x257,
    _Shore001R1001  = 0x258,
    _AnimOcean       = 0x259,
    _Tundra00001000  = 0x25A,
    _Tundra00001100  = 0x25B,
    _Tundra00001110  = 0x25C,
    _Tundra00000110  = 0x25D,
    _Tundra00000010  = 0x25E,
    _Tundra00001010  = 0x25F,
    _Tundra00100010  = 0x260,
    _Tundra10000010  = 0x261,
    _Tundra00011000  = 0x262,
    _Tundra00000100  = 0x263,
    _Tundra00000011  = 0x264,
    _Tundra10100000  = 0x265,
    _Tundra10001000  = 0x266,
    _Tundra00101000  = 0x267,
    _Tundra00111000  = 0x268,
    _Tundra00010000  = 0x269,
    _1Tundra         = 0x26A,
    _Tundra00000001  = 0x26B,
    _Tundra10000011  = 0x26C,
    _Tundra00110000  = 0x26D,
    _Tundra01000000  = 0x26E,
    _Tundra10000001  = 0x26F,
    _Tundra10101000  = 0x270,
    _Tundra00101010  = 0x271,
    _Tundra10001010  = 0x272,
    _Tundra00100000  = 0x273,
    _Tundra01100000  = 0x274,
    _Tundra11100000  = 0x275,
    _Tundra11000000  = 0x276,
    _Tundra10000000  = 0x277,
    _Tundra10100010  = 0x278,
    _Tundra10101010  = 0x279,
    _Tundra11000001  = 0x27A,
    _Tundra11100001  = 0x27B,
    _Tundra11000011  = 0x27C,
    _Tundra11100011  = 0x27D,
    _Tundra00011100  = 0x27E,
    _Tundra00111100  = 0x27F,
    _Tundra00011110  = 0x280,
    _Tundra00111110  = 0x281,
    _Tundra01110000  = 0x282,
    _Tundra01111000  = 0x283,
    _Tundra11110000  = 0x284,
    _Tundra11111000  = 0x285,
    _Tundra00000111  = 0x286,
    _Tundra00001111  = 0x287,
    _Tundra10000111  = 0x288,
    _Tundra10001111  = 0x289,
    _Tundra11101110  = 0x28A,
    _Tundra11100110  = 0x28B,
    _Tundra11101100  = 0x28C,
    _Tundra11100100  = 0x28D,
    _Tundra11001110  = 0x28E,
    _Tundra11000110  = 0x28F,
    _Tundra11001100  = 0x290,
    _Tundra11000100  = 0x291,
    _Tundra01101110  = 0x292,
    _Tundra01100110  = 0x293,
    _Tundra01101100  = 0x294,
    _Tundra01100100  = 0x295,
    _Tundra01001110  = 0x296,
    _Tundra01000110  = 0x297,
    _Tundra01001100  = 0x298,
    _Tundra01000100  = 0x299,
    _Tundra10010011  = 0x29A,
    _Tundra10011011  = 0x29B,
    _Tundra10110011  = 0x29C,
    _Tundra10111011  = 0x29D,
    _Tundra10010001  = 0x29E,
    _Tundra10011001  = 0x29F,
    _Tundra10110001  = 0x2A0,
    _Tundra10111001  = 0x2A1,
    _Tundra00010011  = 0x2A2,
    _Tundra00011011  = 0x2A3,
    _Tundra00110011  = 0x2A4,
    _Tundra00111011  = 0x2A5,
    _Tundra00010001  = 0x2A6,
    _Tundra00011001  = 0x2A7,
    _Tundra00110001  = 0x2A8,
    _Tundra00111001  = 0x2A9,
    _Tundra00011111  = 0x2AA,
    _Tundra11000111  = 0x2AB,
    _Tundra11110001  = 0x2AC,
    _Tundra01111100  = 0x2AD,
    _Tundra10011111  = 0x2AE,
    _Tundra11100111  = 0x2AF,
    _Tundra11111001  = 0x2B0,
    _Tundra01111110  = 0x2B1,
    _Tundra00111111  = 0x2B2,
    _Tundra11001111  = 0x2B3,
    _Tundra11110011  = 0x2B4,
    _Tundra11111100  = 0x2B5,
    _Tundra10111111  = 0x2B6,
    _Tundra11101111  = 0x2B7,
    _Tundra11111011  = 0x2B8,
    _Tundra11111110  = 0x2B9,
    _Tundra10111000  = 0x2BA,
    _Tundra10110000  = 0x2BB,
    _Tundra10011000  = 0x2BC,
    _Tundra10010000  = 0x2BD,
    _Tundra10111010  = 0x2BE,
    _Tundra10110010  = 0x2BF,
    _Tundra10011010  = 0x2C0,
    _Tundra10010010  = 0x2C1,
    _Tundra00111010  = 0x2C2,
    _Tundra00110010  = 0x2C3,
    _Tundra00011010  = 0x2C4,
    _Tundra00010010  = 0x2C5,
    _Tundra10001110  = 0x2C6,
    _Tundra10101110  = 0x2C7,
    _Tundra00101110  = 0x2C8,
    _Tundra10001100  = 0x2C9,
    _Tundra10101100  = 0x2CA,
    _Tundra00101100  = 0x2CB,
    _Tundra10000110  = 0x2CC,
    _Tundra10100110  = 0x2CD,
    _Tundra00100110  = 0x2CE,
    _Tundra10000100  = 0x2CF,
    _Tundra10100100  = 0x2D0,
    _Tundra00100100  = 0x2D1,
    _Tundra00100001  = 0x2D2,
    _Tundra10100001  = 0x2D3,
    _Tundra00100011  = 0x2D4,
    _Tundra10100011  = 0x2D5,
    _Tundra00101001  = 0x2D6,
    _Tundra10101001  = 0x2D7,
    _Tundra00101011  = 0x2D8,
    _Tundra10101011  = 0x2D9,
    _Tundra00001001  = 0x2DA,
    _Tundra10001001  = 0x2DB,
    _Tundra00001011  = 0x2DC,
    _Tundra10001011  = 0x2DD,
    _Tundra01000010  = 0x2DE,
    _Tundra01001010  = 0x2DF,
    _Tundra01001000  = 0x2E0,
    _Tundra11000010  = 0x2E1,
    _Tundra11001010  = 0x2E2,
    _Tundra11001000  = 0x2E3,
    _Tundra01100010  = 0x2E4,
    _Tundra01101010  = 0x2E5,
    _Tundra01101000  = 0x2E6,
    _Tundra11100010  = 0x2E7,
    _Tundra11101010  = 0x2E8,
    _Tundra11101000  = 0x2E9,
    _Tundra11001001  = 0x2EA,
    _Tundra11101001  = 0x2EB,
    _Tundra11001011  = 0x2EC,
    _Tundra11101011  = 0x2ED,
    _Tundra10011100  = 0x2EE,
    _Tundra10111100  = 0x2EF,
    _Tundra10011110  = 0x2F0,
    _Tundra10111110  = 0x2F1,
    _Tundra01110010  = 0x2F2,
    _Tundra01111010  = 0x2F3,
    _Tundra11110010  = 0x2F4,
    _Tundra11111010  = 0x2F5,
    _Tundra00100111  = 0x2F6,
    _Tundra00101111  = 0x2F7,
    _Tundra10100111  = 0x2F8,
    _Tundra10101111  = 0x2F9,
 */

type TerrainData struct {
    Images []image.Image
    Tiles []TerrainTile
}

type TerrainTile struct {
    Images []image.Image
}

// pass in terrain.lbx
func ReadTerrainData(lbxFile *LbxFile) (*TerrainData, error) {
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

    var tiles []TerrainTile

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
            index = value2 - 2
        } else {
            // this formula comes from
            // terrain_lbx_000_offset = (terrain_001_0 * 16384) + (terrain_001_1 * 384) - 0xC0 - 384;
            // this only works if value1 is a multiple of 3
            // 3 -> 126, 6 -> 254, 9 -> 382
            index = value1 * 16384 / 384 + value2 - 2
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
            Image: tileImages,
        })
    }

    return &TerrainData{
        Images: images,
        Tiles: tiles,
    }
}
