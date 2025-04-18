package fonts

import (
    "log"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/util"
)

func GetFontList() []string {
    return []string{
        "BigFont1",
        "BigFont2",
        "BigFont3",
        "BigFont4",
    }
}

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

type ArmyViewFonts struct {
    NormalFont *font.Font
    SmallerFont *font.Font
    BigFont *font.Font
}

func MakeArmyViewFonts(cache *lbx.LbxCache) *ArmyViewFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

    normalColor := util.Lighten(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, -30)
    normalPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        normalColor, normalColor, normalColor,
        normalColor, normalColor, normalColor,
    }
    normalFont := font.MakeOptimizedFontWithPalette(fonts[1], normalPalette)

    smallerFont := font.MakeOptimizedFontWithPalette(fonts[0], normalPalette)

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
    bigFont := font.MakeOptimizedFontWithPalette(fonts[4], bigPalette)

    return &ArmyViewFonts{
        NormalFont: normalFont,
        SmallerFont: smallerFont,
        BigFont: bigFont,
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

    bigFont := font.MakeOptimizedFontWithPalette(fonts[5], yellowPalette)

    // FIXME: this palette isn't exactly right. It should be a yellow-orange fade. Probably it exists somewhere else in the codebase
    yellow := color.RGBA{R: 0xef, G: 0xce, B: 0x4e, A: 0xff}
    fadePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        util.RotateHue(yellow, -0.6),
        // color.RGBA{R: 0xd5, G: 0x88, B: 0x25, A: 0xff},
        util.RotateHue(yellow, -0.3),
        util.RotateHue(yellow, -0.1),
        yellow,
    }

    castFont := font.MakeOptimizedFontWithPalette(fonts[4], fadePalette)

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

    // fixme: make shadow font as well
    descriptionFont := font.MakeOptimizedFontWithPalette(fonts[1], brownPalette)

    whitePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    producingFont := font.MakeOptimizedFontWithPalette(fonts[1], whitePalette)

    smallFontPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0x0},
        color.RGBA{R: 128, G: 128, B: 128, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
    }

    // FIXME: this font should have a black outline around all the glyphs
    smallFont := font.MakeOptimizedFontWithPalette(fonts[1], smallFontPalette)

    rubbleFontPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0x0},
        color.RGBA{R: 128, G: 0, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
        color.RGBA{R: 0xff, G: 0x0, B: 0x0, A: 0xff},
    }

    // FIXME: this font should have a black outline around all the glyphs
    rubbleFont := font.MakeOptimizedFontWithPalette(fonts[1], rubbleFontPalette)

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
        BigFont: bigFont,
        DescriptionFont: descriptionFont,
        ProducingFont: producingFont,
        SmallFont: smallFont,
        RubbleFont: rubbleFont,
        BannerFonts: bannerFonts,
        CastFont: castFont,
    }, nil
}

type CityViewResourceFonts struct {
    HelpFont *font.Font
    HelpTitleFont *font.Font
}

func MakeCityViewResourceFonts(cache *lbx.LbxCache) *CityViewResourceFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        return nil
    }

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

    helpFont := font.MakeOptimizedFontWithPalette(fonts[1], helpPalette)

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

    helpTitleFont := font.MakeOptimizedFontWithPalette(fonts[4], titlePalette)

    return &CityViewResourceFonts{
        HelpFont: helpFont,
        HelpTitleFont: helpTitleFont,
    }
}

type BuildScreenFonts struct {
    TitleFont *font.Font
    TitleFontWhite *font.Font
    DescriptionFont *font.Font
    OkCancelFont *font.Font
    SmallFont *font.Font
    MediumFont *font.Font
}

func MakeBuildScreenFonts(cache *lbx.LbxCache) *BuildScreenFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

    titleFont := font.MakeOptimizedFont(fonts[2])

    alphaWhite := util.PremultiplyAlpha(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 180})

    whitePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        alphaWhite, alphaWhite, alphaWhite,
    }

    titleFontWhite := font.MakeOptimizedFontWithPalette(fonts[2], whitePalette)

    descriptionPalette := color.Palette{
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

    descriptionFont := font.MakeOptimizedFontWithPalette(fonts[4], descriptionPalette)

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

    okCancelFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowGradient)

    smallFont := font.MakeOptimizedFontWithPalette(fonts[1], descriptionPalette)

    mediumFont := font.MakeOptimizedFontWithPalette(fonts[2], descriptionPalette)

    return &BuildScreenFonts{
        TitleFont: titleFont,
        TitleFontWhite: titleFontWhite,
        DescriptionFont: descriptionFont,
        OkCancelFont: okCancelFont,
        SmallFont: smallFont,
        MediumFont: mediumFont,
    }
}

type InputFonts struct {
    NameFont *font.Font
    TitleFont *font.Font
}

func MakeInputFonts(cache *lbx.LbxCache) *InputFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

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

    orange := color.RGBA{R: 0xed, G: 0xa7, B: 0x12, A: 0xff}
    titlePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        util.Lighten(orange, -30),
        util.Lighten(orange, -20),
        util.Lighten(orange, -10),
        util.Lighten(orange, 0),
    }

    nameFont := font.MakeOptimizedFontWithPalette(fonts[4], namePalette)
    titleFont := font.MakeOptimizedFontWithPalette(fonts[4], titlePalette)

    return &InputFonts{
        NameFont: nameFont,
        TitleFont: titleFont,
    }
}

type MerchantFonts struct {
    LightFont *font.Font
}

func MakeMerchantFonts(cache *lbx.LbxCache) *MerchantFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

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

    lightFont := font.MakeOptimizedFontWithPalette(fonts[4], lightPalette)

    return &MerchantFonts{
        LightFont: lightFont,
    }
}

type SurveyorFonts struct {
    SurveyorFont *font.Font
    YellowFont *font.Font
    WhiteFont *font.Font
}

func MakeSurveyorFonts(cache *lbx.LbxCache) *SurveyorFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Error reading fonts: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Error reading fonts: %v", err)
        return nil
    }

    white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
    palette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        white, white, white,
        white, white, white,
    }

    surveyorFont := font.MakeOptimizedFontWithPalette(fonts[4], palette)

    yellow := util.RotateHue(color.RGBA{R: 255, G: 255, B: 0, A: 255}, -0.15)

    paletteYellow := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        yellow, yellow, yellow,
        yellow, yellow, yellow,
    }

    yellowFont := font.MakeOptimizedFontWithPalette(fonts[1], paletteYellow)

    paletteWhite := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        white, white, white,
        white, white, white,
    }

    whiteFont := font.MakeOptimizedFontWithPalette(fonts[1], paletteWhite)

    return &SurveyorFonts{
        SurveyorFont: surveyorFont,
        YellowFont: yellowFont,
        WhiteFont: whiteFont,
    }
}

type MercenariesFonts struct {
    DescriptionFont *font.Font
    SmallFont *font.Font
    MediumFont *font.Font
    OkDismissFont *font.Font
}

func MakeMercenariesFonts(cache *lbx.LbxCache) *MercenariesFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Error reading fonts: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Error reading fonts: %v", err)
        return nil
    }

    descriptionPalette := color.Palette{
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

    descriptionFont := font.MakeOptimizedFontWithPalette(fonts[4], descriptionPalette)
    smallFont := font.MakeOptimizedFontWithPalette(fonts[1], descriptionPalette)
    mediumFont := font.MakeOptimizedFontWithPalette(fonts[2], descriptionPalette)

    yellowGradient := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0},
        color.RGBA{R: 0xed, G: 0xa4, B: 0x00, A: 0xff},
        color.RGBA{R: 0xff, G: 0xbc, B: 0x00, A: 0xff},
        color.RGBA{R: 0xff, G: 0xd6, B: 0x11, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
    }

    okDismissFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowGradient)

    return &MercenariesFonts{
        DescriptionFont: descriptionFont,
        SmallFont: smallFont,
        MediumFont: mediumFont,
        OkDismissFont: okDismissFont,
    }
}

type FizzleFonts struct {
    Font *font.Font
}

func MakeFizzleFonts(cache *lbx.LbxCache) *FizzleFonts {
    fonts := MakeMercenariesFonts(cache)
    return &FizzleFonts{
        Font: fonts.OkDismissFont,
    }
}

type GlobalEnchantmentFonts struct {
    InfoFont *font.Font
}

func MakeGlobalEnchantmentFonts(cache *lbx.LbxCache) *GlobalEnchantmentFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Error reading fonts: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Error reading fonts: %v", err)
        return nil
    }

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

    infoFont := font.MakeOptimizedFontWithPalette(fonts[5], palette)

    return &GlobalEnchantmentFonts{
        InfoFont: infoFont,
    }
}

type HireHeroFonts struct {
    DescriptionFont *font.Font
    SmallFont *font.Font
    MediumFont *font.Font
    OkDismissFont *font.Font
}

func MakeHireHeroFonts(cache *lbx.LbxCache) *HireHeroFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

    descriptionPalette := color.Palette{
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

    descriptionFont := font.MakeOptimizedFontWithPalette(fonts[4], descriptionPalette)
    smallFont := font.MakeOptimizedFontWithPalette(fonts[1], descriptionPalette)
    mediumFont := font.MakeOptimizedFontWithPalette(fonts[2], descriptionPalette)

    yellowGradient := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0},
        color.RGBA{R: 0xed, G: 0xa4, B: 0x00, A: 0xff},
        color.RGBA{R: 0xff, G: 0xbc, B: 0x00, A: 0xff},
        color.RGBA{R: 0xff, G: 0xd6, B: 0x11, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
    }

    okDismissFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowGradient)

    return &HireHeroFonts{
        DescriptionFont: descriptionFont,
        SmallFont: smallFont,
        MediumFont: mediumFont,
        OkDismissFont: okDismissFont,
    }
}

type HeroLevelUpFonts struct {
    TitleFont *font.Font
    SmallFont *font.Font
}

func MakeHeroLevelUpFonts(cache *lbx.LbxCache) *HeroLevelUpFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

    yellowGradient := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0},
        color.RGBA{R: 0xed, G: 0xa4, B: 0x00, A: 0xff},
        color.RGBA{R: 0xff, G: 0xbc, B: 0x00, A: 0xff},
        color.RGBA{R: 0xff, G: 0xd6, B: 0x11, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
        color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xff},
    }

    titleFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowGradient)
    smallFont := font.MakeOptimizedFontWithPalette(fonts[2], yellowGradient)

    return &HeroLevelUpFonts{
        TitleFont: titleFont,
        SmallFont: smallFont,
    }
}

type NewBuildingFonts struct {
    BigFont *font.Font
}

func MakeNewBuildingFonts(cache *lbx.LbxCache) *NewBuildingFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

    yellow := color.RGBA{R: 0xea, G: 0xb6, B: 0x00, A: 0xff}
    yellowPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        yellow, yellow, yellow,
        yellow, yellow, yellow,
        yellow, yellow, yellow,
        yellow, yellow, yellow,
    }

    bigFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowPalette)

    return &NewBuildingFonts{
        BigFont: bigFont,
    }
}

type ScrollFonts struct {
    BigFont *font.Font
    SmallFont *font.Font
}

func MakeScrollFonts(cache *lbx.LbxCache) *ScrollFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

    red := util.Lighten(color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}, -60)
    redPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        red, red, red,
        red, red, red,
    }

    red2 := util.Lighten(color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}, -80)
    redPalette2 := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        red2, red2, red2,
        red2, red2, red2,
    }

    bigFont := font.MakeOptimizedFontWithPalette(fonts[4], redPalette)

    smallFont := font.MakeOptimizedFontWithPalette(fonts[1], redPalette2)

    return &ScrollFonts{
        BigFont: bigFont,
        SmallFont: smallFont,
    }
}

type OutpostFonts struct {
    BigFont *font.Font
}

func MakeOutpostFonts(cache *lbx.LbxCache) *OutpostFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

    // red := color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}
    yellow := util.Lighten(util.RotateHue(color.RGBA{R: 0xff, G: 0xff, B: 0x00, A: 0xff}, -0.60), 0)
    yellowPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        yellow,
        util.Lighten(yellow, -20),
        util.Lighten(yellow, -20),
        util.Lighten(yellow, -15),
        util.Lighten(yellow, -30),
        util.Lighten(yellow, -10),
        util.Lighten(yellow, -15),
        util.Lighten(yellow, -10),
        util.Lighten(yellow, -10),
        util.Lighten(yellow, -35),
        util.Lighten(yellow, -45),
        yellow,
        yellow,
        yellow,
    }

    bigFont := font.MakeOptimizedFontWithPalette(fonts[5], yellowPalette)

    return &OutpostFonts{
        BigFont: bigFont,
    }
}

type RandomEventFonts struct {
    BigFont *font.Font
}

func MakeRandomEventFonts(cache *lbx.LbxCache) *RandomEventFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

    yellow := util.RotateHue(color.RGBA{R: 0xea, G: 0xb6, B: 0x00, A: 0xff}, -0.1)
    yellowPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        yellow,
        util.Lighten(yellow, -5),
        util.Lighten(yellow, -15),
        util.Lighten(yellow, -25),
    }

    bigFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowPalette)

    return &RandomEventFonts{
        BigFont: bigFont,
    }
}

type SettingsFonts struct {
    OptionFont *font.Font
}

func MakeSettingsFonts(cache *lbx.LbxCache) *SettingsFonts {
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

    bluish := util.Lighten(color.RGBA{R: 0x00, G: 0x00, B: 0xff, A: 0xff}, 90)

    optionPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        bluish, bluish, bluish, bluish,
        bluish, bluish, bluish, bluish,
    }

    optionFont := font.MakeOptimizedFontWithPalette(fonts[2], optionPalette)

    return &SettingsFonts{
        OptionFont: optionFont,
    }
}

type TreasureFonts struct {
    TreasureFont *font.Font
}

func MakeTreasureFonts(cache *lbx.LbxCache) *TreasureFonts {
    fontLbx, err := cache.GetLbxFile("FONTS.LBX")
    if err != nil {
        log.Printf("Error: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Error: %v", err)
        return nil
    }

    orange := color.RGBA{R: 0xc7, G: 0x82, B: 0x1b, A: 0xff}

    yellowPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        orange,
        util.Lighten(orange, 15),
        util.Lighten(orange, 30),
        util.Lighten(orange, 50),
        orange,
        orange,
    }

    treasureFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowPalette)

    return &TreasureFonts{
        TreasureFont: treasureFont,
    }
}

type GameFonts struct {
    InfoFontYellow *font.Font
    InfoFontRed *font.Font
    WhiteFont *font.Font
}

func MakeGameFonts(cache *lbx.LbxCache) *GameFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        return nil
    }

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

    infoFontYellow := font.MakeOptimizedFontWithPalette(fonts[0], yellowPalette)

    red := color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}
    redPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        red, red, red,
        red, red, red,
    }

    infoFontRed := font.MakeOptimizedFontWithPalette(fonts[0], redPalette)

    whitePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0x90, G: 0x86, B: 0x81, A: 0xff},
        color.White, color.White, color.White, color.White,
    }

    whiteFont := font.MakeOptimizedFontWithPalette(fonts[0], whitePalette)

    return &GameFonts{
        InfoFontYellow: infoFontYellow,
        InfoFontRed: infoFontRed,
        WhiteFont: whiteFont,
    }
}

type SpellbookFonts struct {
    BigOrange *font.Font
}

func MakeSpellbookFonts(cache *lbx.LbxCache) *SpellbookFonts {
    treasureFonts := MakeTreasureFonts(cache)

    return &SpellbookFonts{
        BigOrange: treasureFonts.TreasureFont,
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
    fontLbx, err := cache.GetLbxFile("FONTS.LBX")
    if err != nil {
        log.Printf("Error: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Error: %v", err)
        return nil
    }

    blue := color.RGBA{R: 146, G: 146, B: 166, A: 0xff}
    bluishPalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        blue, blue, blue, blue,
    }

    normalFont := font.MakeOptimizedFontWithPalette(fonts[2], bluishPalette)

    blue2Palette := bluishPalette
    blue2Palette[1] = color.RGBA{R: 97, G: 97, B: 125, A: 0xff}
    smallerFont := font.MakeOptimizedFontWithPalette(fonts[1], blue2Palette)

    yellowishPalette := color.Palette{
        color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00},
        color.RGBA{R: 0xa6, G: 0x6d, B: 0x1c, A: 0xff},
        color.RGBA{R: 0xff, G: 0xb6, B: 0x2c, A: 0xff},
    }

    spellFont := font.MakeOptimizedFontWithPalette(fonts[0], yellowishPalette)

    transmute := util.PremultiplyAlpha(color.RGBA{R: 223, G: 150, B: 28, A: 255})
    transmutePalette := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        transmute, transmute, transmute,
        transmute, transmute, transmute,
    }

    transmuteFont := font.MakeOptimizedFontWithPalette(fonts[0], transmutePalette)

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
        NormalFont: normalFont,
        SmallerFont: smallerFont,
        SpellFont: spellFont,
        TransmuteFont: transmuteFont,
        BannerBlueFont: bannerBlueFont,
        BannerGreenFont: bannerGreenFont,
        BannerPurpleFont: bannerPurpleFont,
        BannerRedFont: bannerRedFont,
        BannerYellowFont: bannerYellowFont,
    }
}
type SpellSpecialUIFonts struct {
    BigOrange *font.Font
    InfoOrange *font.Font
}

func MakeSpellSpecialUIFonts(cache *lbx.LbxCache) *SpellSpecialUIFonts {
    treasureFonts := MakeTreasureFonts(cache)
    gameFonts := MakeGameFonts(cache)

    return &SpellSpecialUIFonts{
        BigOrange: treasureFonts.TreasureFont,
        InfoOrange: gameFonts.InfoFontYellow,
    }
}

type NewWizardFonts struct {
    BigYellowFont *font.Font
}

func MakeNewWizardFonts(cache *lbx.LbxCache) *NewWizardFonts {
    cityViewFonts, err := MakeCityViewFonts(cache)
    if err != nil {
        return nil
    }

    return &NewWizardFonts{
        BigYellowFont: cityViewFonts.BigFont,
    }
}

type SpellOfMasteryFonts struct {
    Font *font.Font
    RedFont *font.Font
}

func MakeSpellOfMasteryFonts(cache *lbx.LbxCache) *SpellOfMasteryFonts {
    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }

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

    orangeFont := font.MakeOptimizedFontWithPalette(fonts[5], orangePalette)

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

    redFont := font.MakeOptimizedFontWithPalette(fonts[5], redPalette)

    return &SpellOfMasteryFonts{
        Font: orangeFont,
        RedFont: redFont,
    }
}

type MainFonts struct {
    Credits *font.Font
}

func MakeMainFonts(cache *lbx.LbxCache) *MainFonts {
    fontLbx, err := cache.GetLbxFile("FONTS.LBX")
    if err != nil {
        log.Printf("Error: %v", err)
        return nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Error: %v", err)
        return nil
    }

    yellowPalette := color.Palette{
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x0},
        color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x0},
        color.RGBA{R: 0xff, G: 0xcc, B: 0x0, A: 0xff},
    }

    credits := font.MakeOptimizedFontWithPalette(fonts[2], yellowPalette)

    return &MainFonts{
        Credits: credits,
    }
}
