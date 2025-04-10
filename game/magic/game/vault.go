package game

import (
    "log"
    "fmt"
    "image"

    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"
    "github.com/kazzmir/master-of-magic/game/magic/magicview"
    "github.com/kazzmir/master-of-magic/game/magic/mouse"
    "github.com/kazzmir/master-of-magic/game/magic/fonts"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    helplib "github.com/kazzmir/master-of-magic/game/magic/help"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (game *Game) showItemPopup(item *artifact.Artifact, cache *lbx.LbxCache, imageCache *util.ImageCache, vaultFonts *fonts.VaultFonts) (func(coroutine.YieldFunc), func (*ebiten.Image)) {
    if vaultFonts == nil {
        vaultFonts = fonts.MakeVaultFonts(cache)
    }

    counter := uint64(0)

    getAlpha := util.MakeFadeIn(7, &counter)

    drawer := func (screen *ebiten.Image){
        var options ebiten.DrawImageOptions
        options.ColorScale.ScaleAlpha(getAlpha())
        options.GeoM.Translate(float64(48), float64(48))
        artifact.RenderArtifactBox(screen, imageCache, *item, counter / 8, vaultFonts.ItemName, vaultFonts.PowerFont, options)
    }

    logic := func (yield coroutine.YieldFunc) {
        quit := false
        for !quit {
            counter += 1
            if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
                quit = true
            }

            if yield() != nil {
                return
            }
        }

        getAlpha = util.MakeFadeOut(7, &counter)
        for i := 0; i < 7; i++ {
            counter += 1
            if yield() != nil {
                return
            }
        }
    }

    return logic, drawer
}

func (game *Game) showVaultScreen(createdArtifact *artifact.Artifact, player *playerlib.Player) (func(coroutine.YieldFunc), func (*ebiten.Image)) {
    defer mouse.Mouse.SetImage(game.MouseData.Normal)

    imageCache := util.MakeImageCache(game.Cache)

    fonts := fonts.MakeVaultFonts(game.Cache)

    helpLbx, err := game.Cache.GetLbxFile("help.lbx")
    if err != nil {
        log.Printf("Error: could not load help.lbx: %v", err)
        return func(yield coroutine.YieldFunc){}, func (*ebiten.Image){}
    }

    help, err := helplib.ReadHelp(helpLbx, 2)
    if err != nil {
        log.Printf("Error: could not load help.lbx: %v", err)
        return func(yield coroutine.YieldFunc){}, func (*ebiten.Image){}
    }

    // mouse should turn into createdArtifact picture

    selectedItem := createdArtifact

    ui := &uilib.UI{
        Cache: game.Cache,
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            background, _ := imageCache.GetImage("armylist.lbx", 5, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(data.ScreenWidth / 2 - background.Bounds().Dx() / 2), 2)
            scale.DrawScaled(screen, background, &options)

            fontOptions := font.FontOptions{Justify: font.FontJustifyRight, DropShadow: true, Options: &options, Scale: scale.ScaleAmount}
            fonts.ResourceFont.PrintOptions(screen, 190, 166, fontOptions, fmt.Sprintf("%v GP", player.Gold))
            fonts.ResourceFont.PrintOptions(screen, 233, 166, fontOptions, fmt.Sprintf("%v MP", player.Mana))

            ui.StandardDraw(screen)
        },
    }

    ui.SetElementsFromArray(nil)

    updateMouse := func(){
        if selectedItem != nil {
            mouse.Mouse.SetImageFunc(func (screen *ebiten.Image, options *ebiten.DrawImageOptions){
                artifact.RenderArtifactImage(screen, &imageCache, *selectedItem, ui.Counter / 8, *options)
            })
        } else {
            mouse.Mouse.SetImage(game.MouseData.Normal)
        }
    }

    updateMouse()

    /*
    group := uilib.MakeGroup()
    ui.AddGroup(group)
    */

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
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement){
                selectedItem, player.VaultEquipment[index] = player.VaultEquipment[index], selectedItem
                updateMouse()
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

                    var options ebiten.DrawImageOptions
                    options.GeoM.Translate(float64(73 + index * 20), float64(173))
                    /*
                    equipmentImage, _ := imageCache.GetImage("items.lbx", player.VaultEquipment[index].Image, 0)
                    screen.DrawImage(equipmentImage, &options)
                    */
                    artifact.RenderArtifactImage(screen, &imageCache, *player.VaultEquipment[index], ui.Counter / 8, options)
                }
            },
        }
    }

    for i := 0; i < 4; i++ {
        ui.AddElement(makeEquipmentSlot(i))
    }

    // blacksmith anvil
    ui.AddElement(func () *uilib.UIElement {
        rect := image.Rect(26, 158, 65, 190)
        return &uilib.UIElement{
            Rect: rect,
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement){
                if selectedItem != nil {

                    group := uilib.MakeGroup()

                    gainedMana := selectedItem.Cost
                    if !player.Wizard.RetortEnabled(data.RetortArtificer) {
                        gainedMana /= 2
                    }

                    yes := func(){
                        player.Mana += gainedMana
                        selectedItem = nil
                        updateMouse()
                        ui.RemoveGroup(group)
                    }

                    no := func(){
                        ui.RemoveGroup(group)
                    }

                    group.AddElements(uilib.MakeConfirmDialog(group, game.Cache, &imageCache, fmt.Sprintf("Do you want to destroy your %v and gain %v mana crystals?", selectedItem.Name, gainedMana), true, yes, no))
                    ui.AddGroup(group)
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                // util.DrawRect(screen, rect, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})
            },
        }
    }())

    // returns elements for the hero portrait and the 3 item slots
    makeHero := func(index int, hero *herolib.Hero) []*uilib.UIElement {
        // 3 on left, 3 on right

        x1 := (34 + (index % 2) * 135)
        y1 := (16 + (index / 2) * 46)

        portraitLbx, portraitIndex := hero.GetPortraitLbxInfo()
        profile, _ := imageCache.GetImage(portraitLbx, portraitIndex, 0)
        // FIXME: there are 5 of these frame images, how are they selected?
        frame, _ := imageCache.GetImage("portrait.lbx", 36, 0)

        rect := util.ImageRect(x1, y1, profile)

        var options ebiten.DrawImageOptions
        var baseGeom ebiten.GeoM
        baseGeom.Translate(float64(rect.Min.X), float64(rect.Min.Y))
        options.GeoM = baseGeom

        var elements []*uilib.UIElement

        disband := func(){
            ui.RemoveElements(elements)
            player.RemoveUnit(hero)
        }

        elements = append(elements, &uilib.UIElement{
            Rect: rect,
            RightClick: func(element *uilib.UIElement){
                ui.AddGroup(unitview.MakeUnitContextMenu(game.Cache, ui, hero, disband))
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                scale.DrawScaled(screen, profile, &options)
                scale.DrawScaled(screen, frame, &options)
            },
        })

        for slotIndex, slot := range hero.GetArtifactSlots() {
            slotOptions := options
            slotOptions.GeoM = baseGeom
            slotOptions.GeoM.Translate(float64(profile.Bounds().Dx() + 8), float64(14))
            pic, _ := imageCache.GetImage("itemisc.lbx", slot.ImageIndex(), 0)
            slotOptions.GeoM.Translate(float64((pic.Bounds().Dx() + 11) * slotIndex), 0)

            x, y := slotOptions.GeoM.Apply(0, 0)
            rect := util.ImageRect(int(x), int(y), pic)

            elements = append(elements, &uilib.UIElement{
                Rect: rect,
                RightClick: func(element *uilib.UIElement){
                    if hero.Equipment[slotIndex] != nil {
                        select {
                            case showItem <- hero.Equipment[slotIndex]:
                            default:
                        }
                    }
                },
                PlaySoundLeftClick: true,
                LeftClick: func(element *uilib.UIElement){
                    // if the slot is incompatible with the selected item then do not allow a swap
                    if selectedItem == nil || slot.CompatibleWith(selectedItem.Type) {
                        selectedItem, hero.Equipment[slotIndex] = hero.Equipment[slotIndex], selectedItem
                        updateMouse()
                    }
                },
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    if hero.Equipment[slotIndex] != nil {
                        /*
                        artifactPic, _ := imageCache.GetImage("items.lbx", hero.Equipment[slotIndex].Image, 0)
                        screen.DrawImage(artifactPic, &slotOptions)
                        */
                        artifact.RenderArtifactImage(screen, &imageCache, *hero.Equipment[slotIndex], ui.Counter / 8, slotOptions)
                    } else {
                        scale.DrawScaled(screen, pic, &slotOptions)
                    }
                },
            })
        }

        return elements
    }

    for i, hero := range player.Heroes {
        if hero != nil {
            ui.AddElements(makeHero(i, hero))
        }
    }

    quit := false

    ui.AddElement(func () *uilib.UIElement {
        okImages, _ := imageCache.GetImages("armylist.lbx", 8)
        index := 0
        rect := util.ImageRect(237, 177, okImages[index])
        return &uilib.UIElement{
            Rect: rect,
            PlaySoundLeftClick: true,
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
                scale.DrawScaled(screen, okImages[index], &options)
            },
        }
    }())

    ui.AddElement(func () *uilib.UIElement {
        images, _ := imageCache.GetImages("armylist.lbx", 7)
        index := 0
        rect := util.ImageRect(237, 157, images[index])
        return &uilib.UIElement{
            Rect: rect,
            PlaySoundLeftClick: true,
            LeftClick: func(element *uilib.UIElement){
                index = 1
            },
            LeftClickRelease: func(element *uilib.UIElement){
                index = 0
                transmuteGroup := magicview.MakeTransmuteElements(ui, fonts.SmallFont, player, &help, game.Cache, &imageCache)
                ui.AddGroup(transmuteGroup)
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
                scale.DrawScaled(screen, images[index], &options)
            },
        }
    }())

    drawer := func (screen *ebiten.Image){
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
            itemDraw(screen)
        }

        itemLogic(yield)
    }

    logic := func (yield coroutine.YieldFunc) {
        updateMouse()

        for !quit {
            ui.StandardUpdate()

            select {
                case item := <-showItem:
                    showItemPopup(yield, item)
                default:
            }

            if yield() != nil {
                return
            }
        }
    }

    return logic, drawer
}
