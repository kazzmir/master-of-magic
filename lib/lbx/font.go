package lbx

import (
    "bytes"
    "fmt"
    "log"
    "io"
)

/*
https://steamcommunity.com/app/1146370/discussions/0/5879990700799122827/

char = 1 byte
unsigned = 2 bytes
little endian

typedef struct
{
	FONT_HEADER	Hdr_Space;
	unsigned	Font_Heights[8]; // should start at offset 0x16a
	unsigned	Horz_Spacings[8];
	unsigned	Vert_Spacings[8];
	char		Glyph_Widths[8][96];
	unsigned	Glyph_Offsets[8][96]; // position of offset at 0x49a
	char		Glyph_Data[15066];
} FONT_ENTRY;

// all 0's when read out of the file
typedef struct
{
	char		Current_Colors[16];
	unsigned	Font_Height;
	char		Outline_Style;
	char		Color_Index;
	char		Colors_1[16];
	char		Colors_2[16];
	char		Colors_3[16];
	unsigned	Line_Height;
	unsigned	Vert_Spacing;
	unsigned	Horz_Spacing;
	char		Glyph_Widths[96];
	unsigned	Glyph_Offsets[96];
} FONT_HEADER;
 */

type internalFontHeader struct {
    CurrentColors [16]byte
    FontHeight uint16
    OutlineStyle byte
    ColorIndex byte
    Colors1 [16]byte
    Colors2 [16]byte
    Colors3 [16]byte
    LineHeight uint16
    VertSpacing uint16
    HorzSpacing uint16
    GlyphWidths [96]byte
    GlyphOffsets [96]uint16
}

func internalFontHeaderSize() int64 {
    // add all the sizes of the fields in internalFontHeader
    return 16 + 2 + 1 + 1 + 16 + 16 + 16 + 2 + 2 + 2 + 96 + 96 * 2
}

type Glyph struct {
    Data []int
}

type internalFontInfo struct {
    Height int
    HorizontalSpacing int
    VerticalSpacing int
    Widths []int
    GlyphOffsets []int64
    Glyphs []Glyph
}

type Font struct {
}

func readFonts(reader *bytes.Reader) ([]*Font, error) {
    _, err := reader.Seek(internalFontHeaderSize(), io.SeekStart)
    if err != nil {
        return nil, err
    }

    var fontInfo []internalFontInfo

    /* always seem to be 8 fonts */
    for i := 0; i < 8; i++ {
        fontInfo = append(fontInfo, internalFontInfo{})
    }

    for i := 0; i < 8; i++ {
        height, err := readUint16(reader)
        if err != nil {
            return nil, err
        }
        fontInfo[i].Height = int(height)
    }

    for i := 0; i < 8; i++ {
        width, err := readUint16(reader)
        if err != nil {
            return nil, err
        }
        fontInfo[i].HorizontalSpacing = int(width)
    }

    for i := 0; i < 8; i++ {
        height, err := readUint16(reader)
        if err != nil {
            return nil, err
        }
        fontInfo[i].VerticalSpacing = int(height)
    }

    for i := 0; i < 8; i++ {
        for g := 0; g < 96; g++ {
            width, err := reader.ReadByte()
            if err != nil {
                return nil, err
            }
            fontInfo[i].Widths = append(fontInfo[i].Widths, int(width))
        }
    }

    for i := 0; i < 8; i++ {
        for g := 0; g < 96; g++ {
            offset, err := readUint16(reader)
            if err != nil {
                return nil, err
            }
            fontInfo[i].GlyphOffsets = append(fontInfo[i].GlyphOffsets, int64(offset))
        }
    }

    log.Printf("Seek to font 0 glyph 0 offset 0x%x", fontInfo[0].GlyphOffsets[0])
    reader.Seek(fontInfo[0].GlyphOffsets[0], io.SeekStart)

    return nil, fmt.Errorf("fail")
}
