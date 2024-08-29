package font

import (
    "bytes"
    "fmt"
    "log"
    "image"
    "io"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
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
    Height int
}

var fontPalette = color.Palette {
    color.RGBA{R: 0x0,  G: 0x0,  B: 0x0, A: 0xff},
    color.RGBA{R: 0x8,  G: 0x4,  B: 0x4, A: 0xff},
    color.RGBA{R: 0x24, G: 0x1c, B: 0x18, A: 0xff},
    color.RGBA{R: 0x38, G: 0x30, B: 0x2c, A: 0xff},
    color.RGBA{R: 0x48, G: 0x40, B: 0x3c, A: 0xff},
    color.RGBA{R: 0x58, G: 0x50, B: 0x4c, A: 0xff},
    color.RGBA{R: 0x68, G: 0x60, B: 0x5c, A: 0xff},
    color.RGBA{R: 0x7c, G: 0x74, B: 0x70, A: 0xff},
    color.RGBA{R: 0x8c, G: 0x84, B: 0x80, A: 0xff},
    color.RGBA{R: 0x9c, G: 0x94, B: 0x90, A: 0xff},
    color.RGBA{R: 0xac, G: 0xa4, B: 0xa0, A: 0xff},
    color.RGBA{R: 0xc0, G: 0xb8, B: 0xb4, A: 0xff},
    color.RGBA{R: 0xd0, G: 0xc8, B: 0xc4, A: 0xff},
    color.RGBA{R: 0xe0, G: 0xd8, B: 0xd4, A: 0xff},
    color.RGBA{R: 0xf0, G: 0xe8, B: 0xe4, A: 0xff},
    color.RGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff},
    color.RGBA{R: 0x38, G: 0x20, B: 0x1c, A: 0xff},
    color.RGBA{R: 0x40, G: 0x2c, B: 0x24, A: 0xff},
    color.RGBA{R: 0x48, G: 0x34, B: 0x2c, A: 0xff},
    color.RGBA{R: 0x50, G: 0x3c, B: 0x30, A: 0xff},
}

func (glyph *Glyph) MakeImage() image.Image {
    return glyph.MakeImageWithPalette(lbx.GetDefaultPalette())
}

func (glyph *Glyph) MakeImageWithPalette(palette color.Palette) image.Image {
    if glyph.Width == 0 {
        return nil
    }

    out := image.NewPaletted(image.Rect(0, 0, glyph.Width, glyph.Height), palette)

    dataIndex := 0
    for column := 0; column < glyph.Width; column++ {
        row := 0

        for row <= glyph.Height && dataIndex < len(glyph.Data) {
            value := glyph.Data[dataIndex]
            dataIndex += 1

            if value >> 7 == 1 {
                remaining := value & 0x7f

                // done with this column
                if remaining == 0 {
                    break
                }

                // skip down remaining rows
                row += int(remaining)
            } else {
                if row == glyph.Height {
                    log.Printf("Error: glyph had data after last row")
                    return out
                }

                length := value >> 4
                color := value & 0x0f

                if length == 0 {
                    log.Printf("Error: glyph had 0-streak length")
                    return out
                }

                for i := 0; i < int(length); i++ {
                    // log.Printf("Pixel %v, %v color %v", column, row, color)
                    out.SetColorIndex(column, row, color + 1)
                    row += 1
                }
            }
        }
    }

    return out
}

type internalFontInfo struct {
    Height int
    HorizontalSpacing int
    VerticalSpacing int
    Widths []int
    GlyphOffsets []int64
    Glyphs []Glyph
}

type LbxFont struct {
    Height int
    HorizontalSpacing int
    VerticalSpacing int
    Glyphs []Glyph
}

func (font *LbxFont) GlyphForRune(r rune) *Glyph {
    if r < 32 || r >= 128 {
        return nil
    }

    return &font.Glyphs[r - 32]
}

func (font *LbxFont) GlyphCount() int {
    return len(font.Glyphs)
}

func readFontInfo(reader *bytes.Reader) ([]internalFontInfo, error) {
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
        height, err := lbx.ReadUint16(reader)
        if err != nil {
            return nil, err
        }
        fontInfo[i].Height = int(height)
    }

    for i := 0; i < 8; i++ {
        width, err := lbx.ReadUint16(reader)
        if err != nil {
            return nil, err
        }
        fontInfo[i].HorizontalSpacing = int(width)
    }

    for i := 0; i < 8; i++ {
        height, err := lbx.ReadUint16(reader)
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
            offset, err := lbx.ReadUint16(reader)
            if err != nil {
                return nil, err
            }
            fontInfo[i].GlyphOffsets = append(fontInfo[i].GlyphOffsets, int64(offset))
        }
    }

    return fontInfo, nil
}

func readFont(reader *bytes.Reader, fontInfo *internalFontInfo) (*LbxFont, error) {
    font := LbxFont{
        Height: fontInfo.Height,
        HorizontalSpacing: fontInfo.HorizontalSpacing,
        VerticalSpacing: fontInfo.VerticalSpacing,
    }

    for glyphIndex, glyphOffset := range fontInfo.GlyphOffsets {
        reader.Seek(glyphOffset, io.SeekStart)

        if fontInfo.Widths[glyphIndex] == 0 {
            // log.Printf("Empty glyph at font=%v glyph=%v", fontIndex, glyphIndex)
            font.Glyphs = append(font.Glyphs, Glyph{Width: 0})
        } else {
            // FIXME: this doesn't really have to be a copy of the data, we could just as
            // easily create a new reader for the slice at the current position of 'reader'
            // that can read for at most Width*Height bytes
            glyphData := make([]byte, fontInfo.Widths[glyphIndex] * fontInfo.Height)
            n, err := reader.Read(glyphData)
            if err != nil {
                return nil, err
            }

            if n == 0 {
                return nil, fmt.Errorf("unable to read glyph %v offset 0x%x", glyphIndex, glyphOffset)
            }

            /*
            if n != len(glyphData) {
                return nil, fmt.Errorf("unable to read entire glyph size %v font=%v glyph=%v offset=0x%x, read %v", len(glyphData), fontIndex, glyphIndex, glyphOffset, n)
            }
            */

            glyph := Glyph{
                Data: glyphData[0:n],
                Width: fontInfo.Widths[glyphIndex],
                Height: fontInfo.Height,
            }

            font.Glyphs = append(font.Glyphs, glyph)
        }
    }

    return &font, nil
}

func readFonts(reader *bytes.Reader) ([]*LbxFont, error) {
    fontInfos, err := readFontInfo(reader)
    if err != nil {
        return nil, err
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

    var fonts []*LbxFont

    for i, fontInfo := range fontInfos {
        font, err := readFont(reader, &fontInfo)
        if err != nil {
            return nil, fmt.Errorf("Error reading font %v: %v", i, err)
        }
        fonts = append(fonts, font)
    }

    return fonts, nil
}

func ReadFont(lbxFile* lbx.LbxFile, entry int, fontIndex int) (*LbxFont, error) {
    reader, err := lbxFile.GetReader(entry)
    if err != nil {
        return nil, err
    }

    fontInfo, err := readFontInfo(reader)
    if err != nil {
        return nil, err
    }

    if fontIndex < 0 || fontIndex >= len(fontInfo) {
        return nil, fmt.Errorf("invalid font index %v", fontIndex)
    }

    return readFont(reader, &fontInfo[fontIndex])
}

func ReadFonts(lbxFile *lbx.LbxFile, entry int) ([]*LbxFont, error) {
    reader, err := lbxFile.GetReader(entry)
    if err != nil {
        return nil, err
    }

    return readFonts(reader)
}

