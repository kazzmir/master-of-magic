package game

import (
    "log"
    "fmt"
    "image"
    "image/color"

    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/mouse"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

type VaultFonts struct {
    ItemName *font.Font
    PowerFont *font.Font
    ResourceFont *font.Font
}

func makeFonts(cache *lbx.LbxCache) *VaultFonts {
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

    return &VaultFonts{
        ItemName: itemName,
        PowerFont: powerFont,
        ResourceFont: resourceFont,
    }
}

func (game *Game) showItemPopup(item *artifact.Artifact, cache *lbx.LbxCache, imageCache *util.ImageCache, vaultFonts *VaultFonts) (func(coroutine.YieldFunc), func (*ebiten.Image, bool)) {
    if vaultFonts == nil {
        vaultFonts = makeFonts(cache)
    }

    counter := uint64(0)

    getAlpha := util.MakeFadeIn(7, &counter)

    mouseNormal, _ := mouse.GetMouseNormal(cache)

    drawer := func (screen *ebiten.Image, drawMouse bool){
        var options ebiten.DrawImageOptions
        options.ColorScale.ScaleAlpha(getAlpha())
        itemBackground, _ := imageCache.GetImage("itemisc.lbx", 25, 0)
        options.GeoM.Translate(32, 48)
        screen.DrawImage(itemBackground, &options)

        itemImage, _ := imageCache.GetImage("items.lbx", item.Image, 0)
        options.GeoM.Translate(10, 8)
        screen.DrawImage(itemImage, &options)

        x, y := options.GeoM.Apply(float64(itemImage.Bounds().Max.X) + 3, 4)

        vaultFonts.ItemName.Print(screen, x, y, 1, options.ColorScale, item.Name)

        dot, _ := imageCache.GetImage("itemisc.lbx", 26, 0)
        savedGeom := options.GeoM
        for i, power := range item.Powers {
            options.GeoM = savedGeom
            options.GeoM.Translate(3, 26)
            options.GeoM.Translate(float64(i / 2 * 80), float64(i % 2 * 13))

            screen.DrawImage(dot, &options)

            x, y := options.GeoM.Apply(float64(dot.Bounds().Dx() + 1), 0)
            vaultFonts.PowerFont.Print(screen, x, y, 1, options.ColorScale, power.String())
        }

        if drawMouse {
            var mouseOptions ebiten.DrawImageOptions
            mouseX, mouseY := ebiten.CursorPosition()
            mouseOptions.GeoM.Translate(float64(mouseX), float64(mouseY))
            screen.DrawImage(mouseNormal, &mouseOptions)
        }
    }

    logic := func (yield coroutine.YieldFunc) {
        quit := false
        for !quit {
            counter += 1
            if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
                quit = true
            }
            yield()
        }

        getAlpha = util.MakeFadeOut(7, &counter)
        for i := 0; i < 7; i++ {
            counter += 1
            yield()
        }
    }

    return logic, drawer
}

func (game *Game) showVaultScreen(createdArtifact *artifact.Artifact, player *playerlib.Player, heroes []*herolib.Hero) (func(coroutine.YieldFunc), func (*ebiten.Image, bool)) {
    imageCache := util.MakeImageCache(game.Cache)

    fonts := makeFonts(game.Cache)

    // mouse should turn into createdArtifact picture

    mouseNormal, _ := mouse.GetMouseNormal(game.Cache)

    selectedItem := createdArtifact

    var artifactImage *ebiten.Image

    drawMouse := false

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            background, _ := imageCache.GetImage("armylist.lbx", 5, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(data.ScreenWidth / 2 - background.Bounds().Dx() / 2), 2)
            screen.DrawImage(background, &options)

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })

            fonts.ResourceFont.PrintRight(screen, 190, 166, 1, options.ColorScale, fmt.Sprintf("%v GP", player.Gold))
            fonts.ResourceFont.PrintRight(screen, 233, 166, 1, options.ColorScale, fmt.Sprintf("%v MP", player.Mana))

            if drawMouse {
                options.GeoM.Reset()
                mouseX, mouseY := ebiten.CursorPosition()
                options.GeoM.Translate(float64(mouseX), float64(mouseY))

                if selectedItem != nil {
                    artifactImage, _ = imageCache.GetImage("items.lbx", selectedItem.Image, 0)
                    screen.DrawImage(artifactImage, &options)
                } else {
                    screen.DrawImage(mouseNormal, &options)
                }
            }
        },
    }

    ui.SetElementsFromArray(nil)

    showItem := make(chan *artifact.Artifact, 10)

    // the 4 equipment slots
    makeEquipmentSlot := func(index int) *uilib.UIElement{
        width := 20
        height := 17
        x1 := 72 + index * width
        y1 := 173
        rect := image.Rect(x1, y1, x1 + width, y1 + height)

        return &uilib.UIElement{
            Rect: rect,
            LeftClick: func(element *uilib.UIElement){
                selectedItem, player.VaultEquipment[index] = player.VaultEquipment[index], selectedItem
            },
            RightClick: func(element *uilib.UIElement){
                if player.VaultEquipment[index] != nil {
                    select {
                        case showItem <- player.VaultEquipment[index]:
                        default:
                    }
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                if player.VaultEquipment[index] != nil {
                    // util.DrawRect(screen, rect, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})

                    equipmentImage, _ := imageCache.GetImage("items.lbx", player.VaultEquipment[index].Image, 0)
                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(73 + index * 20), 173)
                    screen.DrawImage(equipmentImage, &options)
                }
            },
        }
    }

    for i := 0; i < 4; i++ {
        ui.AddElement(makeEquipmentSlot(i))
    }

    ui.AddElement(func () *uilib.UIElement {
        rect := image.Rect(26, 158, 65, 190)
        return &uilib.UIElement{
            Rect: rect,
            LeftClick: func(element *uilib.UIElement){
                if selectedItem != nil {

                    gainedMana := selectedItem.Cost() / 2

                    yes := func(){
                        player.Mana += gainedMana
                        selectedItem = nil
                    }

                    no := func(){
                    }

                    ui.AddElements(uilib.MakeConfirmDialog(ui, game.Cache, &imageCache, fmt.Sprintf("Do you want to destroy your %v and gain %v mana crystals?", selectedItem.Name, gainedMana), yes, no))
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                // util.DrawRect(screen, rect, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})
            },
        }
    }())

    makeHero := func(index int, hero *herolib.Hero) *uilib.UIElement {
        // 3 on left, 3 on right

        x1 := 20 + (index % 2) * 100
        y1 := 2 + (index / 2) * 50

        profile, _ := imageCache.GetImage("portrait.lbx", hero.PortraitIndex(), 0)

        rect := util.ImageRect(x1, y1, profile)

        return &uilib.UIElement{
            Rect: rect,
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
                screen.DrawImage(profile, &options)
            },
        }
    }

    for i, hero := range heroes {
        ui.AddElement(makeHero(i, hero))
    }

    quit := false

    ui.AddElement(func () *uilib.UIElement {
        okImages, _ := imageCache.GetImages("armylist.lbx", 8)
        index := 0
        rect := util.ImageRect(237, 177, okImages[index])
        return &uilib.UIElement{
            Rect: rect,
            LeftClick: func(element *uilib.UIElement){
                if selectedItem == nil {
                    index = 1
                }
            },
            LeftClickRelease: func(element *uilib.UIElement){
                if selectedItem == nil {
                    index = 0
                    // can't quit the screen while holding an item
                    quit = true
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
                if selectedItem != nil {
                    options.ColorScale.SetR(2)
                }
                screen.DrawImage(okImages[index], &options)
            },
        }
    }())

    ui.AddElement(func () *uilib.UIElement {
        images, _ := imageCache.GetImages("armylist.lbx", 7)
        index := 0
        rect := util.ImageRect(237, 157, images[index])
        return &uilib.UIElement{
            Rect: rect,
            LeftClick: func(element *uilib.UIElement){
                index = 1
            },
            LeftClickRelease: func(element *uilib.UIElement){
                index = 0
                // FIXME: show alchemy ui
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
                screen.DrawImage(images[index], &options)
            },
        }
    }())

    drawer := func (screen *ebiten.Image, drawMouseX bool){
        drawMouse = drawMouseX
        ui.Draw(ui, screen)
    }

    showItemPopup := func (yield coroutine.YieldFunc, item *artifact.Artifact){
        itemLogic, itemDraw := game.showItemPopup(item, game.Cache, &imageCache, fonts)

        drawer := game.Drawer
        defer func(){
            game.Drawer = drawer
        }()

        game.Drawer = func (screen *ebiten.Image, game *Game){
            drawer(screen, game)
            itemDraw(screen, drawMouse)
        }

        itemLogic(yield)
    }

    logic := func (yield coroutine.YieldFunc) {
        ebiten.SetCursorMode(ebiten.CursorModeHidden)
        defer ebiten.SetCursorMode(ebiten.CursorModeVisible)
        for !quit {
            ui.StandardUpdate()

            select {
                case item := <-showItem:
                    showItemPopup(yield, item)
                default:
            }

            yield()
        }
    }

    return logic, drawer
}
