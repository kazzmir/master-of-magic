package fonts

// a place to centralize font creation

import (
    "log"
    "image/color"
    "slices"
    "cmp"
    "maps"

    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/util"
)

type FontLoader func(fonts []*font.LbxFont) *font.Font

var fontLoaders map[string]FontLoader

const LightGradient1 = "LightGradient1"
const PowerFont = "PowerFont"
const ResourceFont = "ResourceFont"
const SmallFont = "SmallFont"

const NormalFont = "NormalFont"
const SmallerFont = "SmallerFont"
const BigFont = "BigFont"
const TitleYellowFont = "TitleYellowFont"
const DescriptionFont = "DescriptionFont"
const SmallWhite = "SmallWhite"
const SmallRed = "SmallRed"
const HelpFont = "HelpFont"
const HelpTitleFont = "HelpTitleFont"
const TitleFont = "TitleFont"
const TitleFontWhite = "TitleFontWhite"
const YellowBig = "YellowBig"
const YellowBig2 = "YellowBig2"
const WhiteBig = "WhiteBig"
const MediumWhite2 = "MediumWhite2"
const NameFont = "NameFont"
const TitleFontOrange = "TitleFontOrange"
const LightFont = "LightFont"
const LightFontSmall = "LightFontSmall"
const SurveyorFont = "SurveyorFont"
const YellowFont = "YellowFont"
const InfoFont = "InfoFont"

const BigRed2 = "BigRed2"
const SmallRed2 = "SmallRed2"
const BigOrangeGradient2 = "BigOrangeGradient2"
const SettingsFont = "SettingsFont"
const SmallYellow = "SmallYellow"
const NormalBlue = "NormalBlue"
const SmallBlue = "SmallBlue"
const SpellFont = "SpellFont"
const TransmuteFont = "TransmuteFont"
const HugeOrange = "HugeOrange"
const HugeRed = "HugeRed"
const NormalYellow = "NormalYellow"

// use util/font-list to see how these fonts are rendered
func init() {
    fontLoaders = make(map[string]FontLoader)

    fontLoaders[LightGradient1] = func (fonts []*font.LbxFont) *font.Font {
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

        return font.MakeOptimizedFontWithPalette(fonts[4], namePalette)
    }

    fontLoaders[PowerFont] = func (fonts []*font.LbxFont) *font.Font {
        orange := color.RGBA{R: 0xc7, G: 0x82, B: 0x1b, A: 0xff}
        // red1 := color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}
        powerPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            util.Lighten(orange, 40),
        }

        return font.MakeOptimizedFontWithPalette(fonts[2], powerPalette)
    }

    fontLoaders[ResourceFont] = func (fonts []*font.LbxFont) *font.Font {
        white := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
        whitePalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            white, white, white, white,
        }

        return font.MakeOptimizedFontWithPalette(fonts[1], whitePalette)
    }

    fontLoaders[SmallFont] = func (fonts []*font.LbxFont) *font.Font {
        translucentWhite := util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 80})
        transmutePalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            translucentWhite, translucentWhite, translucentWhite,
            translucentWhite, translucentWhite, translucentWhite,
        }

        return font.MakeOptimizedFontWithPalette(fonts[0], transmutePalette)
    }

    fontLoaders[NormalFont] = func (fonts []*font.LbxFont) *font.Font {
        normalColor := util.Lighten(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, -30)
        normalPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            normalColor, normalColor, normalColor,
            normalColor, normalColor, normalColor,
        }
        return font.MakeOptimizedFontWithPalette(fonts[1], normalPalette)
    }

    fontLoaders[SmallerFont] = func (fonts []*font.LbxFont) *font.Font {
        normalColor := util.Lighten(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, -30)
        normalPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            normalColor, normalColor, normalColor,
            normalColor, normalColor, normalColor,
        }
        return font.MakeOptimizedFontWithPalette(fonts[0], normalPalette)
    }

    fontLoaders[BigFont] = func (fonts []*font.LbxFont) *font.Font {
        // red := color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}

        yellow := color.RGBA{R: 0xf9, G: 0xdb, B: 0x4c, A: 0xff}
        bigPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            util.RotateHue(yellow, -0.50),
            util.RotateHue(yellow, -0.30),
            util.RotateHue(yellow, -0.10),
            yellow,
        }
        return font.MakeOptimizedFontWithPalette(fonts[4], bigPalette)
    }

    fontLoaders[TitleYellowFont] = func (fonts []*font.LbxFont) *font.Font {
        yellowPalette := color.Palette{
            color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x0},
            color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xd9, G: 0xb9, B: 0x9e, A: 0xff},
            color.RGBA{R: 0xff, G: 0xcb, B: 0x66, A: 0xff},
            color.RGBA{R: 0xff, G: 0xc9, B: 0x26, A: 0xff},
            color.RGBA{R: 0xeb, G: 0xa3, B: 0x28, A: 0xff},
            color.RGBA{R: 0xdb, G: 0x92, B: 0x1d, A: 0xff},
            color.RGBA{R: 0xa3, G: 0x7c, B: 0x55, A: 0xff},
            color.RGBA{R: 0xd4, G: 0x8d, B: 0x17, A: 0xff},
            color.RGBA{R: 0xa1, G: 0x66, B: 0x12, A: 0xff},
            color.RGBA{R: 0x78, G: 0x54, B: 0x23, A: 0xff},
            color.RGBA{R: 0x47, G: 0x37, B: 0x24, A: 0xff},
            color.RGBA{R: 0x69, G: 0x5c, B: 0xc, A: 0xff},
            color.RGBA{R: 0x47, G: 0x37, B: 0x24, A: 0xff},
        }

        return font.MakeOptimizedFontWithPalette(fonts[5], yellowPalette)
    }

    fontLoaders[DescriptionFont] = func (fonts []*font.LbxFont) *font.Font {
        brownPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0xe1, G: 0x8e, B: 0x32, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        }

        return font.MakeOptimizedFontWithPalette(fonts[1], brownPalette)
    }

    fontLoaders[SmallWhite] = func (fonts []*font.LbxFont) *font.Font {
        smallFontPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0x0},
            color.RGBA{R: 128, G: 128, B: 128, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        }

        return font.MakeOptimizedFontWithPalette(fonts[1], smallFontPalette)
    }

    fontLoaders[SmallRed] = func (fonts []*font.LbxFont) *font.Font {
        rubbleFontPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0, A: 0x0},
            color.RGBA{R: 0, G: 0, B: 0, A: 0x0},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        }

        return font.MakeOptimizedFontWithPalette(fonts[1], rubbleFontPalette)
    }

    fontLoaders[HelpFont] = func (fonts []*font.LbxFont) *font.Font {
        helpPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0x5e, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        }

        return font.MakeOptimizedFontWithPalette(fonts[1], helpPalette)
    }

    fontLoaders[HelpTitleFont] = func (fonts []*font.LbxFont) *font.Font {
        titleRed := color.RGBA{R: 0x50, G: 0x00, B: 0x0e, A: 0xff}
        titlePalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            titleRed,
            titleRed,
            titleRed,
            titleRed,
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        }

        return font.MakeOptimizedFontWithPalette(fonts[4], titlePalette)
    }

    fontLoaders[TitleFontWhite] = func (fonts []*font.LbxFont) *font.Font {
        alphaWhite := util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 180})

        whitePalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            alphaWhite, alphaWhite, alphaWhite,
        }

        return font.MakeOptimizedFontWithPalette(fonts[2], whitePalette)
    }

    fontLoaders[YellowBig] = func (fonts []*font.LbxFont) *font.Font {
        yellowGradient := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        }

        return font.MakeOptimizedFontWithPalette(fonts[4], yellowGradient)
    }

    fontLoaders[TitleFont] = func (fonts []*font.LbxFont) *font.Font {
        return font.MakeOptimizedFont(fonts[2])
    }

    descriptionPalette := func () color.Palette {
        return color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 90}),
            util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}),
            util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 200}),
            util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 200}),
            util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 200}),
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        }
    }

    fontLoaders[WhiteBig] = func (fonts []*font.LbxFont) *font.Font {
        return font.MakeOptimizedFontWithPalette(fonts[4], descriptionPalette())
    }

    fontLoaders[MediumWhite2] = func (fonts []*font.LbxFont) *font.Font {
        return font.MakeOptimizedFontWithPalette(fonts[2], descriptionPalette())
    }

    fontLoaders[NameFont] = func (fonts []*font.LbxFont) *font.Font {
        bluish := color.RGBA{R: 0xcf, G: 0xef, B: 0xf9, A: 0xff}
        // red := color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}
        namePalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0, A: 0},
            color.RGBA{R: 0, G: 0, B: 0, A: 0},
            util.Lighten(bluish, -30),
            util.Lighten(bluish, -20),
            util.Lighten(bluish, -10),
            util.Lighten(bluish, 0),
        }

        return font.MakeOptimizedFontWithPalette(fonts[4], namePalette)
    }

    fontLoaders[TitleFontOrange] = func (fonts []*font.LbxFont) *font.Font {
        orange := color.RGBA{R: 0xed, G: 0xa7, B: 0x12, A: 0xff}
        titlePalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0, A: 0},
            color.RGBA{R: 0, G: 0, B: 0, A: 0},
            util.Lighten(orange, -30),
            util.Lighten(orange, -20),
            util.Lighten(orange, -10),
            util.Lighten(orange, 0),
        }

        return font.MakeOptimizedFontWithPalette(fonts[4], titlePalette)
    }

    fontLoaders[LightFont] = func (fonts []*font.LbxFont) *font.Font {
        lightPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0},
            color.RGBA{R: 0xed, G: 0xa4, B: 0x00, A: 0xff},
            color.RGBA{R: 0xff, G: 0xbc, B: 0x00, A: 0xff},
            color.RGBA{R: 0xff, G: 0xd6, B: 0x11, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        }

        return font.MakeOptimizedFontWithPalette(fonts[4], lightPalette)
    }

    fontLoaders[LightFontSmall] = func (fonts []*font.LbxFont) *font.Font {
        lightPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0},
            color.RGBA{R: 0xed, G: 0xa4, B: 0x00, A: 0xff},
            color.RGBA{R: 0xff, G: 0xbc, B: 0x00, A: 0xff},
            color.RGBA{R: 0xff, G: 0xd6, B: 0x11, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        }

        return font.MakeOptimizedFontWithPalette(fonts[2], lightPalette)
    }


    fontLoaders[SurveyorFont] = func (fonts []*font.LbxFont) *font.Font {
        white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
        palette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0, A: 0},
            color.RGBA{R: 0, G: 0, B: 0, A: 0},
            white, white, white,
            white, white, white,
        }

        return font.MakeOptimizedFontWithPalette(fonts[4], palette)
    }

    fontLoaders[YellowFont] = func (fonts []*font.LbxFont) *font.Font {
        yellow := util.RotateHue(color.RGBA{R: 255, G: 255, B: 0, A: 255}, -0.15)

        paletteYellow := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0, A: 0},
            color.RGBA{R: 0, G: 0, B: 0, A: 0},
            yellow, yellow, yellow,
            yellow, yellow, yellow,
        }

        return font.MakeOptimizedFontWithPalette(fonts[1], paletteYellow)
    }

    fontLoaders[InfoFont] = func (fonts []*font.LbxFont) *font.Font {
        white := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
        // red := color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}
        palette := color.Palette{
            color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x0},
            color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x0},
            white,
            white,
            white,
            white,
            white,
            white,
            util.Lighten(white, -20),
            util.Lighten(white, -30),
            util.Lighten(white, -60),
            util.Lighten(white, -40),
            util.Lighten(white, -60),
            util.Lighten(white, -50),
        }

        return font.MakeOptimizedFontWithPalette(fonts[5], palette)
    }

    fontLoaders[YellowBig2] = func (fonts []*font.LbxFont) *font.Font {
        yellow := color.RGBA{R: 0xea, G: 0xb6, B: 0x00, A: 0xff}
        yellowPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0, A: 0},
            color.RGBA{R: 0, G: 0, B: 0, A: 0},
            yellow, yellow, yellow,
            yellow, yellow, yellow,
            yellow, yellow, yellow,
            yellow, yellow, yellow,
        }

        return font.MakeOptimizedFontWithPalette(fonts[4], yellowPalette)
    }

    fontLoaders[BigRed2] = func (fonts []*font.LbxFont) *font.Font {
        red := util.Lighten(color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}, -60)
        redPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0, A: 0},
            color.RGBA{R: 0, G: 0, B: 0, A: 0},
            red, red, red,
            red, red, red,
        }
        return font.MakeOptimizedFontWithPalette(fonts[4], redPalette)
    }

    fontLoaders[SmallRed2] = func (fonts []*font.LbxFont) *font.Font {
        red2 := util.Lighten(color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}, -80)
        redPalette2 := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0, A: 0},
            color.RGBA{R: 0, G: 0, B: 0, A: 0},
            red2, red2, red2,
            red2, red2, red2,
        }

        return font.MakeOptimizedFontWithPalette(fonts[1], redPalette2)
    }

    // this is very similar to TitleOrangeFont
    fontLoaders[BigOrangeGradient2] = func (fonts []*font.LbxFont) *font.Font {
        yellow := util.RotateHue(color.RGBA{R: 0xea, G: 0xb6, B: 0x00, A: 0xff}, -0.1)
        yellowPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0, A: 0},
            color.RGBA{R: 0, G: 0, B: 0, A: 0},
            yellow,
            util.Lighten(yellow, -5),
            util.Lighten(yellow, -15),
            util.Lighten(yellow, -25),
        }

        return font.MakeOptimizedFontWithPalette(fonts[4], yellowPalette)
    }

    fontLoaders[SettingsFont] = func (fonts []*font.LbxFont) *font.Font {
        bluish := util.Lighten(color.RGBA{R: 0x00, G: 0x00, B: 0xff, A: 0xff}, 90)

        optionPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            bluish, bluish, bluish, bluish,
            bluish, bluish, bluish, bluish,
        }

        return font.MakeOptimizedFontWithPalette(fonts[2], optionPalette)
    }

    fontLoaders[SmallYellow] = func (fonts []*font.LbxFont) *font.Font {
        orange := color.RGBA{R: 0xc7, G: 0x82, B: 0x1b, A: 0xff}

        yellowPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            orange,
            orange,
            orange,
            orange,
            orange,
            orange,
        }

        return font.MakeOptimizedFontWithPalette(fonts[0], yellowPalette)
    }

    fontLoaders[NormalBlue] = func (fonts []*font.LbxFont) *font.Font {
        blue := color.RGBA{R: 146, G: 146, B: 166, A: 0xff}
        bluishPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            blue, blue, blue, blue,
        }

        return font.MakeOptimizedFontWithPalette(fonts[2], bluishPalette)
    }

    fontLoaders[SmallBlue] = func (fonts []*font.LbxFont) *font.Font {
        blue := color.RGBA{R: 146, G: 146, B: 166, A: 0xff}
        bluishPalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            blue, blue, blue, blue,
        }

        // normalFont := font.MakeOptimizedFontWithPalette(fonts[2], bluishPalette)

        blue2Palette := bluishPalette
        blue2Palette[1] = color.RGBA{R: 97, G: 97, B: 125, A: 0xff}
        return font.MakeOptimizedFontWithPalette(fonts[1], blue2Palette)
    }

    fontLoaders[SpellFont] = func (fonts []*font.LbxFont) *font.Font {
        yellowishPalette := color.Palette{
            color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00},
            color.RGBA{R: 0xa6, G: 0x6d, B: 0x1c, A: 0xff},
            color.RGBA{R: 0xff, G: 0xb6, B: 0x2c, A: 0xff},
        }

        return font.MakeOptimizedFontWithPalette(fonts[0], yellowishPalette)
    }

    fontLoaders[TransmuteFont] = func (fonts []*font.LbxFont) *font.Font {
        transmute := util.PremultiplyAlpha(color.RGBA{R: 223, G: 150, B: 28, A: 255})
        transmutePalette := color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
            transmute, transmute, transmute,
            transmute, transmute, transmute,
        }

        return font.MakeOptimizedFontWithPalette(fonts[0], transmutePalette)
    }

    fontLoaders[HugeOrange] = func (fonts []*font.LbxFont) *font.Font {
        orangePalette := color.Palette{
            color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x0},
            color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xc6, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xc6, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xc1, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xc1, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xc1, B: 0x0, A: 0xff},
            color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x0},
            color.RGBA{R: 0xff, G: 0xc1, B: 0x0, A: 0xff},
            color.RGBA{R: 0xe3, G: 0xb0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xe3, G: 0xb0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xff, G: 0xc1, B: 0x0, A: 0xff},
            color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x0},
            color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x0},
        }

        return font.MakeOptimizedFontWithPalette(fonts[5], orangePalette)
    }

    fontLoaders[HugeRed] = func (fonts []*font.LbxFont) *font.Font {
        redPalette := color.Palette{
            color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x0},
            color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff},
            color.RGBA{R: 0xcf, G: 0x29, B: 0x0, A: 0xff},
            color.RGBA{R: 0xe8, G: 0x2e, B: 0x0, A: 0xff},
            color.RGBA{R: 0xe0, G: 0x2d, B: 0x0, A: 0xff},
            color.RGBA{R: 0xe8, G: 0x2e, B: 0x0, A: 0xff},
            color.RGBA{R: 0xe8, G: 0x2e, B: 0x0, A: 0xff},
            color.RGBA{R: 0xe8, G: 0x2e, B: 0x0, A: 0xff},
            color.RGBA{R: 0xd4, G: 0x2a, B: 0x0, A: 0xff},
            color.RGBA{R: 0xe8, G: 0x2e, B: 0x0, A: 0xff},
            color.RGBA{R: 0xba, G: 0x25, B: 0x0, A: 0xff},
            color.RGBA{R: 0xe8, G: 0x2e, B: 0x0, A: 0xff},
            color.RGBA{R: 0xe8, G: 0x2e, B: 0x0, A: 0xff},
            color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x0},
        }

        return font.MakeOptimizedFontWithPalette(fonts[5], redPalette)
    }

    fontLoaders[NormalYellow] = func (fonts []*font.LbxFont) *font.Font {
        yellowPalette := color.Palette{
            color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x0},
            color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x0},
            color.RGBA{R: 0xff, G: 0xcc, B: 0x0, A: 0xff},
        }

        return font.MakeOptimizedFontWithPalette(fonts[2], yellowPalette)
    }
}

func GetFontList() []string {
    return slices.SortedFunc(maps.Keys(fontLoaders), cmp.Compare)
}

// return a function that returns a font given by its name
//   loader, err := Loader(cache)
//   font := loader(PowerFont)
func Loader(cache *lbx.LbxCache) (func (string) *font.Font, error) {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        return nil, err
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        return nil, err
    }

    return func (name string) *font.Font {
        loader, ok := fontLoaders[name]
        if ok {
            return loader(fonts)
        }

        return nil
    }, nil
}

// load a bunch of fonts at once in a map, not really much of an optimization over just using Loader()
func LoadFonts(cache *lbx.LbxCache, names ...string) (map[string]*font.Font, error) {
    loader, err := Loader(cache)
    if err != nil {
        return nil, err
    }

    out := make(map[string]*font.Font)

    for _, name := range names {
        out[name] = loader(name)
    }

    return out, nil
}

type VaultFonts struct {
    ItemName *font.Font
    PowerFont *font.Font
    ResourceFont *font.Font
    SmallFont *font.Font
}

func MakeVaultFonts(cache *lbx.LbxCache) *VaultFonts {
    use, err := Loader(cache)
    if err != nil {
        log.Printf("Error loading vault fonts: %v", err)
        return nil
    }

    return &VaultFonts{
        ItemName: use(LightGradient1),
        PowerFont: use(PowerFont),
        ResourceFont: use(ResourceFont),
        SmallFont: use(SmallFont),
    }
}

type CityViewFonts struct {
    BigFont *font.Font
    DescriptionFont *font.Font
    ProducingFont *font.Font
    SmallFont *font.Font
    RubbleFont *font.Font
    CastFont *font.Font
    BannerFonts map[data.BannerType]*font.Font
}

func MakeCityViewFonts(cache *lbx.LbxCache) (*CityViewFonts, error) {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        return nil, err
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil, err
    }

    loader, err := Loader(cache)
    if err != nil {
        log.Printf("Unable to load fonts: %v", err)
        return nil, err
    }

    makeBannerPalette := func(banner data.BannerType) color.Palette {
        var bannerColor color.RGBA

        switch banner {
            case data.BannerBlue: bannerColor = color.RGBA{R: 0x00, G: 0x00, B: 0xff, A: 0xff}
            case data.BannerGreen: bannerColor = color.RGBA{R: 0x00, G: 0xf0, B: 0x00, A: 0xff}
            case data.BannerPurple: bannerColor = color.RGBA{R: 0x8f, G: 0x30, B: 0xff, A: 0xff}
            case data.BannerRed: bannerColor = color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}
            case data.BannerYellow: bannerColor = color.RGBA{R: 0xff, G: 0xff, B: 0x00, A: 0xff}
        }

        return color.Palette{
            color.RGBA{R: 0, G: 0, B: 0x00, A: 0x0},
            color.RGBA{R: 0, G: 0, B: 0, A: 0x0},
            bannerColor, bannerColor, bannerColor,
            bannerColor, bannerColor, bannerColor,
            bannerColor, bannerColor, bannerColor,
        }
    }

    bannerFonts := make(map[data.BannerType]*font.Font)

    for _, banner := range []data.BannerType{data.BannerGreen, data.BannerBlue, data.BannerRed, data.BannerPurple, data.BannerYellow} {
        bannerFonts[banner] = font.MakeOptimizedFontWithPalette(fonts[0], makeBannerPalette(banner))
    }

    return &CityViewFonts{
        BigFont: loader(TitleYellowFont),
        DescriptionFont: loader(DescriptionFont),
        ProducingFont: loader(ResourceFont),
        SmallFont: loader(SmallWhite),
        RubbleFont: loader(SmallRed),
        BannerFonts: bannerFonts,
        CastFont: loader(LightGradient1),
    }, nil
}

type SurveyorFonts struct {
    SurveyorFont *font.Font
    YellowFont *font.Font
    WhiteFont *font.Font
}

func MakeSurveyorFonts(cache *lbx.LbxCache) *SurveyorFonts {
    loader, err := Loader(cache)
    if err != nil {
        log.Printf("Error loading surveyor fonts: %v", err)
        return nil
    }

    return &SurveyorFonts{
        SurveyorFont: loader(SurveyorFont),
        YellowFont: loader(YellowFont),
        WhiteFont: loader(SmallWhite),
    }
}

type SettingsFonts struct {
    OptionFont *font.Font
}

func MakeSettingsFonts(cache *lbx.LbxCache) *SettingsFonts {
    loader, err := Loader(cache)
    if err != nil {
        log.Printf("Error loading settings fonts: %v", err)
        return nil
    }

    return &SettingsFonts{
        OptionFont: loader(SettingsFont),
    }
}

type TreasureFonts struct {
    TreasureFont *font.Font
}

func MakeTreasureFonts(cache *lbx.LbxCache) *TreasureFonts {
    loader, err := Loader(cache)
    if err != nil {
        log.Printf("Error loading fonts: %v", err)
        return nil
    }

    return &TreasureFonts{
        TreasureFont: loader(LightGradient1),
    }
}

type SpellbookFonts struct {
    BigOrange *font.Font
}

func MakeSpellbookFonts(cache *lbx.LbxCache) *SpellbookFonts {
    loader, err := Loader(cache)
    if err != nil {
        log.Printf("Error loading fonts: %v", err)
        return nil
    }

    return &SpellbookFonts{
        BigOrange: loader(LightGradient1),
    }
}

type MagicViewFonts struct {
    NormalFont *font.Font
    SmallerFont *font.Font
    SpellFont *font.Font
    TransmuteFont *font.Font
    BannerBlueFont *font.Font
    BannerGreenFont *font.Font
    BannerPurpleFont *font.Font
    BannerRedFont *font.Font
    BannerYellowFont *font.Font
}

func MakeMagicViewFonts(cache *lbx.LbxCache) *MagicViewFonts {
    loader, err := Loader(cache)
    if err != nil {
        log.Printf("Error loading magic view fonts: %v", err)
        return nil
    }

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Error: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Error: %v", err)
        return nil
    }

    bannerBlue := color.RGBA{R: 0x00, G: 0x00, B: 0xff, A: 0xff}
    bannerGreen := color.RGBA{R: 0x00, G: 0xf0, B: 0x00, A: 0xff}
    bannerPurple := color.RGBA{R: 0x8f, G: 0x30, B: 0xff, A: 0xff}
    bannerRed := color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}
    bannerYellow := color.RGBA{R: 0xff, G: 0xff, B: 0x00, A: 0xff}

    bannerBluePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        bannerBlue, bannerBlue, bannerBlue,
        bannerBlue, bannerBlue, bannerBlue,
    }
    bannerGreenPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        bannerGreen, bannerGreen, bannerGreen,
        bannerGreen, bannerGreen, bannerGreen,
    }
    bannerPurplePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        bannerPurple, bannerPurple, bannerPurple,
        bannerPurple, bannerPurple, bannerPurple,
    }
    bannerRedPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        bannerRed, bannerRed, bannerRed,
        bannerRed, bannerRed, bannerRed,
    }
    bannerYellowPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        bannerYellow, bannerYellow, bannerYellow,
        bannerYellow, bannerYellow, bannerYellow,
    }

    bannerBlueFont := font.MakeOptimizedFontWithPalette(fonts[2], bannerBluePalette)
    bannerGreenFont := font.MakeOptimizedFontWithPalette(fonts[2], bannerGreenPalette)
    bannerPurpleFont := font.MakeOptimizedFontWithPalette(fonts[2], bannerPurplePalette)
    bannerRedFont := font.MakeOptimizedFontWithPalette(fonts[2], bannerRedPalette)
    bannerYellowFont := font.MakeOptimizedFontWithPalette(fonts[2], bannerYellowPalette)

    return &MagicViewFonts{
        NormalFont: loader(NormalBlue),
        SmallerFont: loader(SmallBlue),
        SpellFont: loader(SpellFont),
        TransmuteFont: loader(TransmuteFont),
        BannerBlueFont: bannerBlueFont,
        BannerGreenFont: bannerGreenFont,
        BannerPurpleFont: bannerPurpleFont,
        BannerRedFont: bannerRedFont,
        BannerYellowFont: bannerYellowFont,
    }
}
