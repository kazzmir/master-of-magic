package lbx

import (
    "bytes"
    "fmt"
    _ "log"
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
    Data []byte
    Width int
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
    Height int
    HorizontalSpacing int
    VerticalSpacing int
    Glyph []Glyph
}

func (font *Font) GlyphCount() int {
    return len(font.Glyph)
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

    /*
    log.Printf("Seek to font 0 glyph 0 offset 0x%x", fontInfo[0].GlyphOffsets[0])
    reader.Seek(fontInfo[0].GlyphOffsets[0], io.SeekStart)

    glyphData := make([]byte, fontInfo[0].Widths[0] * fontInfo[0].Height)
    n, err := reader.Read(glyphData)
    if err != nil {
        return nil, err
    }
    if n != len(glyphData) {
        return nil, fmt.Errorf("unable to read entire glyph size %v", len(glyphData))
    }

    log.Printf("Read glyph")
    for _, b := range glyphData {
        fmt.Printf("0x%x ", b)
    }
    fmt.Println()
    */

    var fonts []*Font

    for fontIndex := 0; fontIndex < 8; fontIndex++ {
        font := Font{
            Height: fontInfo[fontIndex].Height,
            HorizontalSpacing: fontInfo[fontIndex].HorizontalSpacing,
            VerticalSpacing: fontInfo[fontIndex].VerticalSpacing,
        }

        for glyphIndex, glyphOffset := range fontInfo[fontIndex].GlyphOffsets {
            reader.Seek(glyphOffset, io.SeekStart)

            if fontInfo[fontIndex].Widths[glyphIndex] == 0 {
                // log.Printf("Empty glyph at font=%v glyph=%v", fontIndex, glyphIndex)
                font.Glyph = append(font.Glyph, Glyph{Width: 0})
            } else {
                glyphData := make([]byte, fontInfo[fontIndex].Widths[glyphIndex] * fontInfo[fontIndex].Height)
                n, err := reader.Read(glyphData)
                if err != nil {
                    return nil, err
                }

                if n == 0 {
                    return nil, fmt.Errorf("unable to read glyph at font %v glyph %v offset 0x%x", fontIndex, glyphIndex, glyphOffset)
                }

                /*
                if n != len(glyphData) {
                    return nil, fmt.Errorf("unable to read entire glyph size %v font=%v glyph=%v offset=0x%x, read %v", len(glyphData), fontIndex, glyphIndex, glyphOffset, n)
                }
                */

                glyph := Glyph{
                    Data: glyphData[0:n],
                    Width: fontInfo[fontIndex].Widths[glyphIndex],
                }

                font.Glyph = append(font.Glyph, glyph)
            }
        }

        fonts = append(fonts, &font)
    }

    return fonts, nil
}
