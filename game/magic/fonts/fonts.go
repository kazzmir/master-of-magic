package fonts

import (
    "log"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/util"
)

// a place to centralize font creation

type VaultFonts struct {
    ItemName *font.Font
    PowerFont *font.Font
    ResourceFont *font.Font
    SmallFont *font.Font
}

func MakeVaultFonts(cache *lbx.LbxCache) *VaultFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Could not read fonts: %v", err)
        return &VaultFonts{}
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Could not read fonts: %v", err)
        return &VaultFonts{}
    }

    orange := color.RGBA{R: 0xc7, G: 0x82, B: 0x1b, A: 0xff}
    namePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        util.Lighten(orange, 0),
        util.Lighten(orange, 20),
        util.Lighten(orange, 50),
        util.Lighten(orange, 80),
        orange,
        orange,
    }

    // red1 := color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}
    powerPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        util.Lighten(orange, 40),
    }

    itemName := font.MakeOptimizedFontWithPalette(fonts[4], namePalette)
    powerFont := font.MakeOptimizedFontWithPalette(fonts[2], powerPalette)

    white := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
    whitePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        white, white, white, white,
    }

    resourceFont := font.MakeOptimizedFontWithPalette(fonts[1], whitePalette)

    translucentWhite := util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 80})
    transmutePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        translucentWhite, translucentWhite, translucentWhite,
        translucentWhite, translucentWhite, translucentWhite,
    }

    transmuteFont := font.MakeOptimizedFontWithPalette(fonts[0], transmutePalette)

    return &VaultFonts{
        ItemName: itemName,
        PowerFont: powerFont,
        ResourceFont: resourceFont,
        SmallFont: transmuteFont,
    }
}


