package city

import (
    "log"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/hajimehoshi/ebiten/v2"
)

type BuildScreen struct {
    LbxCache *lbx.LbxCache
    ImageCache util.ImageCache
    City *City
    TitleFont *font.Font
}

func MakeBuildScreen(cache *lbx.LbxCache, city *City) *BuildScreen {

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil
    }

    fonts, err := fontLbx.ReadFonts(0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil
    }
    
    titleFont := font.MakeOptimizedFont(fonts[2])

    return &BuildScreen{
        LbxCache: cache,
        ImageCache: util.MakeImageCache(cache),
        City: city,
        TitleFont: titleFont,
    }
}

func (build *BuildScreen) Update() {
}

func (build *BuildScreen) Draw(screen *ebiten.Image) {
    mainInfo, err := build.ImageCache.GetImage("unitview.lbx", 0, 0)
    if err == nil {
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(75, 0)
        screen.DrawImage(mainInfo, &options)
    }

    buildingInfo, err := build.ImageCache.GetImage("unitview.lbx", 31, 0)
    if err == nil {
        possibleBuildings := []Building{BuildingTradeGoods, BuildingHousing, BuildingBarracks, BuildingStables}

        var options ebiten.DrawImageOptions
        options.GeoM.Translate(0, 4)
        for _, building := range possibleBuildings {
            screen.DrawImage(buildingInfo, &options)
            x, y := options.GeoM.Apply(0, 0)
            build.TitleFont.Print(screen, x + 2, y + 1, 1, building.String())

            options.GeoM.Translate(0, float64(buildingInfo.Bounds().Dy() + 1))

        }
    }

    unitInfo, err := build.ImageCache.GetImage("unitview.lbx", 32, 0)
    if err == nil {
        possibleUnits := []units.Unit{units.LizardSpearmen, units.LizardSettlers}

        var options ebiten.DrawImageOptions
        options.GeoM.Translate(240, 4)
        for _, unit := range possibleUnits {
            screen.DrawImage(unitInfo, &options)
            x, y := options.GeoM.Apply(0, 0)
            build.TitleFont.Print(screen, x + 2, y + 1, 1, unit.String())
            options.GeoM.Translate(0, float64(unitInfo.Bounds().Dy() + 1))
        }
    }
}
