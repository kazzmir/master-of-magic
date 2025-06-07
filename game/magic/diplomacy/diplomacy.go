package diplomacy

import (
    "log"
    "fmt"
    "image"
    "image/color"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
)

type TalkItem struct {
    Text string
    Action func()
}

type Talk struct {
    UI *uilib.UI
    Font *font.Font
    TitleFont *font.Font
    Items []TalkItem
    Elements []*uilib.UIElement
}

func (talk *Talk) SetTitle(title string) {
    wrap := talk.TitleFont.CreateWrappedText(float64(220), 1, title)

    newElement := &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            talk.TitleFont.RenderWrapped(screen, float64(50), float64(140), wrap, font.FontOptions{Scale: scale.ScaleAmount})
        },
    }

    talk.Elements = append(talk.Elements, newElement)
    talk.UI.AddElement(newElement)
}

func (talk *Talk) Clear() {
    talk.Items = nil
    talk.UI.RemoveElements(talk.Elements)
    talk.Elements = nil
}

func (talk *Talk) AddItem(item string, available bool, action func()){
    talk.Items = append(talk.Items, TalkItem{
        Text: item,
        Action: action,
    })

    posY := 145
    for range len(talk.Elements) {
        posY += (talk.Font.Height() + 1)
    }

    textWidth := talk.Font.MeasureTextWidth(item, 1)

    rect := image.Rect(70, posY, 70 + int(textWidth), posY + talk.Font.Height())

    highlight := false

    newElement := &uilib.UIElement{
        Rect: rect,
        LeftClick: func(element *uilib.UIElement){
            if available {
                action()
            }
        },
        Inside: func(element *uilib.UIElement, x int, y int){
            if available {
                highlight = true
            }
        },
        NotInside: func(element *uilib.UIElement){
            if available {
                highlight = false
            }
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions

            // FIXME: if this option is not available then show the text in grey scale

            if !available {
                options.ColorScale.SetR(0.5)
                options.ColorScale.SetG(0.5)
                options.ColorScale.SetB(0.5)
            }

            if highlight {
                options.ColorScale.SetR(2)
                options.ColorScale.SetG(2)
                options.ColorScale.SetB(2)

                vector.DrawFilledRect(screen, scale.Scale(float32(rect.Min.X)), scale.Scale(float32(rect.Min.Y)), scale.Scale(float32(220)), scale.Scale(float32(talk.Font.Height())), color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 50}, false)

            }

            talk.Font.PrintOptions(screen, 70, float64(posY), font.FontOptions{Options: &options, Scale: scale.ScaleAmount}, item)
        },
    }

    talk.Elements = append(talk.Elements, newElement)
    talk.UI.AddElement(newElement)
}

/* player is talking to enemy
 */
func ShowDiplomacyScreen(cache *lbx.LbxCache, player *playerlib.Player, enemy *playerlib.Player, gameYear int) (func (coroutine.YieldFunc), func (*ebiten.Image)) {

    imageCache := util.MakeImageCache(cache)

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return nil, nil
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return nil, nil
    }

    solid := util.Lighten(color.RGBA{R: 0xca, G: 0x8a, B: 0x4a, A: 0xff}, -10)

    yellowGradient := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        solid,
        util.Lighten(solid, 10),
        util.Lighten(solid, 20),
        util.Lighten(solid, 40),
        util.Lighten(solid, 30),
        util.Lighten(solid, 30),
        solid,
        solid,
        solid,
    }

    solidOrange := color.Palette{
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        color.RGBA{R: 0, G: 0, B: 0x00, A: 0},
        solid, solid, solid,
        solid, solid, solid,
        solid, solid, solid,
    }

    bigFont := font.MakeOptimizedFontWithPalette(fonts[4], yellowGradient)
    _ = bigFont
    normalFont := font.MakeOptimizedFontWithPalette(fonts[1], solidOrange)

    quit := false

    animationIndex := 0
    switch enemy.Wizard.Base {
        case data.WizardMerlin: animationIndex = 0
        case data.WizardRaven: animationIndex = 1
        case data.WizardSharee: animationIndex = 2
        case data.WizardLoPan: animationIndex = 3
        case data.WizardJafar: animationIndex = 4
        case data.WizardOberic: animationIndex = 5
        case data.WizardRjak: animationIndex = 6
        case data.WizardSssra: animationIndex = 7
        case data.WizardTauron: animationIndex = 8
        case data.WizardFreya: animationIndex = 9
        case data.WizardHorus: animationIndex = 10
        case data.WizardAriel: animationIndex = 11
        case data.WizardTlaloc: animationIndex = 12
        case data.WizardKali: animationIndex = 13
    }

    relationship, hasRelationship := enemy.GetDiplomaticRelation(player)

    // the fade in animation
    images, _ := imageCache.GetImages("diplomac.lbx", 38 + animationIndex)
    wizardAnimation := util.MakeAnimation(images, false)

    diplomacLbx, _ := cache.GetLbxFile("diplomac.lbx")
    // the tauron fade in, but any mask will work
    maskSprites, _ := diplomacLbx.ReadImages(46)
    mask := maskSprites[0]

    var makeCutoutMask util.ImageTransformFunc = func (img *image.Paletted) image.Image {
        properImage := img.SubImage(mask.Bounds()).(*image.Paletted)
        imageOut := image.NewPaletted(properImage.Bounds(), properImage.Palette)

        for x := properImage.Bounds().Min.X; x < properImage.Bounds().Max.X; x++ {
            for y := properImage.Bounds().Min.Y; y < properImage.Bounds().Max.Y; y++ {
                maskColor := mask.At(x, y)
                _, _, _, a := maskColor.RGBA()
                if a > 0 {
                    imageOut.Set(x, y, properImage.At(x, y))
                } else {
                    imageOut.SetColorIndex(x, y, 0)
                }
            }
        }

        return imageOut
    }

    var talkMain func()
    var talkTributeSpells func()

    clickedOnce := false

    ui := &uilib.UI{
        Draw: func (ui *uilib.UI, screen *ebiten.Image){
            ui.StandardDraw(screen)
        },
        LeftClick: func(){
            if !clickedOnce {
                talkMain()
                clickedOnce = true
            }
        },
    }

    ui.SetElementsFromArray(nil)

    talk := Talk{
        UI: ui,
        Font: normalFont,
        TitleFont: bigFont,
    }

    doTalk := true

    var talkTribute func()

    willTrade := false

    updateWillTrade := func() {
        if hasRelationship && rand.N(100) < relationship.TradeInterest {
            willTrade = true
        } else {
            willTrade = false
        }
    }

    updateWillTrade()

    makeTalkTradeSpell := func(spell spellbook.Spell) func(){

        var choices spellbook.Spells
        for _, spell := range enemy.KnownSpells.Spells {
            if !player.KnownSpells.Contains(spell) {
                choices.AddSpell(spell)
            }
        }

        if len(choices.Spells) == 0 {
            return talkMain
        }

        choice := rand.N(len(choices.Spells))

        return func(){
            doTalk = true
            talk.Clear()
            talk.SetTitle(fmt.Sprintf("I will trade you %v", choices.Spells[choice].Name))
            talk.AddItem("Yes", true, func(){
                player.KnownSpells.AddSpell(choices.Spells[choice])
                enemy.KnownSpells.AddSpell(spell)

                relationship.TradeInterest = max(-100, relationship.TradeInterest - 20)
                updateWillTrade()

                talkMain()
            })
            talk.AddItem("Forget It", true, talkMain)
        }
    }

    talkSpells := func(){
        doTalk = true
        talk.Clear()

        talk.SetTitle("What spell do you wish to exchange?")

        var choices spellbook.Spells
        for _, spell := range player.KnownSpells.Spells {
            if !enemy.KnownSpells.Contains(spell) {
                choices.AddSpell(spell)
            }
        }

        for i, index := range rand.Perm(len(choices.Spells)) {
            talk.AddItem(choices.Spells[index].Name, true, makeTalkTradeSpell(choices.Spells[index]))
            if i >= 4 {
                break
            }
        }

        talk.AddItem("Forget It", true, talkMain)
    }

    talkEnterPact := func(){
        doTalk = true
        talk.Clear()
        talk.SetTitle(fmt.Sprintf("Let it be known that in the year %v, %v and I have entered into a wizard pact.", gameYear, player.Wizard.Name))
        ui.AddDelay(140, talkMain)
    }

    talkEnterAlliance := func(){
        doTalk = true
        talk.Clear()
        talk.SetTitle(fmt.Sprintf("Let it be known that in the year %v, %v and I have entered into an alliance.", gameYear, player.Wizard.Name))
        ui.AddDelay(140, talkMain)
    }

    talkEnterPeaceTreaty := func(){
        doTalk = true
        talk.Clear()
        talk.SetTitle(fmt.Sprintf("Let it be known that in the year %v, %v and I have entered into a peace treaty.", gameYear, player.Wizard.Name))
        ui.AddDelay(140, talkMain)
    }

    talkTreaty := func(){
        doTalk = true
        talk.Clear()
        talk.SetTitle("You propopse a treaty:")
        talk.AddItem("Wizard Pact", true, func(){
            enemy.PactWithPlayer(player)
            player.PactWithPlayer(enemy)
            talkEnterPact()
        })
        talk.AddItem("Alliance", true, func(){
            enemy.AllianceWithPlayer(player)
            player.AllianceWithPlayer(enemy)
            talkEnterAlliance()
        })
        talk.AddItem("Peace Treaty", true, func(){
            if hasRelationship {
                relationship.PeaceCounter = 50
                talkEnterPeaceTreaty()
            }
        })
        talk.AddItem("Declaration of War on Another Wizard", true, func(){})
        talk.AddItem("Break Alliance With Another Wizard", true, func(){})
        talk.AddItem("Forget It", true, talkMain)
    }

    talkMain = func(){
        doTalk = true
        talk.Clear()

        talk.SetTitle("How may I serve you:")

        talk.AddItem("Propose Treaty", true, talkTreaty)
        talk.AddItem("Threaten/Break Treaty", false, func(){})
        talk.AddItem("Offer Tribute", true, talkTribute)
        talk.AddItem("Exchange spells", willTrade, talkSpells)
        talk.AddItem("Good Bye", true, func(){
            quit = true
        })
    }

    talkThanksTribute := func(){
        doTalk = true
        talk.Clear()
        talk.SetTitle(fmt.Sprintf("I thank you, glorius %v, for your tribute.", player.Wizard.Name))
        enemy.AdjustDiplomaticRelation(player, 5)
        player.AdjustDiplomaticRelation(enemy, 5)
        ui.AddDelay(140, talkMain)
    }

    talkTributeSpells = func(){
        doTalk = true
        talk.Clear()
        talk.SetTitle("What do you offer as tribute?")

        var choices spellbook.Spells
        for _, spell := range player.KnownSpells.Spells {
            if !enemy.KnownSpells.Contains(spell) {
                choices.AddSpell(spell)
            }
        }

        for i, index := range rand.Perm(len(choices.Spells)) {
            talk.AddItem(choices.Spells[index].Name, true, func(){
                log.Printf("Grant spell %v to %v", choices.Spells[index].Name, enemy.Wizard.Name)
                enemy.KnownSpells.AddSpell(choices.Spells[index])
                talkThanksTribute()
            })
            if i >= 4 {
                break
            }
        }

        talk.AddItem("Forget It", true, talkTribute)
    }

    talkTribute = func(){
        doTalk = true
        talk.Clear()
        talk.SetTitle("What do you offer as tribute?")
        gold := int(float64(player.Gold) * 0.1)
        if gold > 5 {
            talk.AddItem(fmt.Sprintf("%v gold", gold), true, func(){
                player.Gold -= gold
                enemy.Gold += gold
                talkThanksTribute()
            })
        }
        talk.AddItem("Spells", true, talkTributeSpells)
        talk.AddItem("Forget It", true, talkMain)
    }

    talk.Clear()
    talk.SetTitle(fmt.Sprintf("Hail, mighty %v. I bear greetings and words of wisdom.", player.Wizard.Name))

    var counter uint64
    logic := func (yield coroutine.YieldFunc) {
        animating := true

        for !quit {
            ui.StandardUpdate()
            counter += 1
            if counter % 7 == 0 {
                if animating && wizardAnimation.Done() {
                    // 0 = happy, 1 = angry, 2 = neutral
                    moodIndex := 0
                    mood, _ := imageCache.GetImageTransform("moodwiz.lbx", animationIndex, moodIndex, "cutout", makeCutoutMask)
                    wizardAnimation = util.MakeAnimation([]*ebiten.Image{mood}, false)
                    animating = false
                }

                wizardAnimation.Next()
            }

            // if rand.N(100) == 0 {
            if doTalk {
                doTalk = false
                // talking
                images, _ := imageCache.GetImagesTransform("diplomac.lbx", 24 + animationIndex, "cutout", makeCutoutMask)
                wizardAnimation = util.MakeAnimation(images, false)
                animating = true
            }

            yield()
        }
    }

    draw := func (screen *ebiten.Image) {
        background, _ := imageCache.GetImage("diplomac.lbx", 0, 0)
        var options ebiten.DrawImageOptions
        scale.DrawScaled(screen, background, &options)

        // choose eye color based on relationship
        eyes := 10
        eyeChoice := 5
        if hasRelationship {
            eyeChoice = (relationship.VisibleRelation + 100) * eyes / 200
            if eyeChoice > eyes {
                eyeChoice = eyes
            }
        }

        // red left eye
        leftEye, _ := imageCache.GetImage("diplomac.lbx", 2 + eyeChoice, 0)
        // red right eye
        rightEye, _ := imageCache.GetImage("diplomac.lbx", 13 + eyeChoice, 0)

        options.GeoM.Translate(63, 58)
        scale.DrawScaled(screen, leftEye, &options)

        options.GeoM.Translate(170, 0)
        scale.DrawScaled(screen, rightEye, &options)

        options.GeoM.Reset()
        options.GeoM.Translate(106, 11)
        scale.DrawScaled(screen, wizardAnimation.Frame(), &options)

        ui.Draw(ui, screen)

        /*
        bigFont.Print(screen, 60, 140, 1, ebiten.ColorScale{}, "How may I serve you:")
        normalFont.Print(screen, 70, 160, 1, ebiten.ColorScale{}, "Good bye")
        */
    }

    return logic, draw
}
