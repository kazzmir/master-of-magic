package game

import (
    "fmt"

    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
    active bool

    MainHud *ebiten.Image
    GameButtons []*ebiten.Image
    SpellButtons []*ebiten.Image
    ArmyButtons []*ebiten.Image
    CityButtons []*ebiten.Image
    MagicButtons []*ebiten.Image
    InfoButtons []*ebiten.Image
    PlaneButtons []*ebiten.Image

    GoldFoodMagic *ebiten.Image
    NextTurnBackground *ebiten.Image
    NextTurn *ebiten.Image

    // FIXME: need one map for arcanus and one for myrran
    Map *Map
}

func (game *Game) Load(cache *lbx.LbxCache) error {
    mainLbx, err := cache.GetLbxFile("MAIN.LBX")
    if err != nil {
        return fmt.Errorf("Unable to load MAIN.LBX: %v", err)
    }

    var outError error

    loadImages := func(index int) []*ebiten.Image {
        if outError != nil {
            return nil
        }

        sprites, err := mainLbx.ReadImages(index)
        if err != nil {
            outError = fmt.Errorf("Unable to read background image from NEWGAME.LBX: %v", err)
            return nil
        }

        var out []*ebiten.Image
        for i := 0; i < len(sprites); i++ {
            out = append(out, ebiten.NewImageFromImage(sprites[i]))
        }
        return out
    }

    loadImage := func(index int, subIndex int) *ebiten.Image {
        if outError != nil {
            return nil
        }

        sprites, err := mainLbx.ReadImages(index)
        if err != nil {
            outError = fmt.Errorf("Unable to read background image from NEWGAME.LBX: %v", err)
            return nil
        }

        if len(sprites) <= subIndex {
            outError = fmt.Errorf("Unable to read background image from NEWGAME.LBX: index %d out of range", subIndex)
            return nil
        }

        return ebiten.NewImageFromImage(sprites[subIndex])
    }

    game.MainHud = loadImage(0, 0)
    game.GameButtons = loadImages(1)
    game.SpellButtons = loadImages(2)
    game.ArmyButtons = loadImages(3)
    game.CityButtons = loadImages(4)
    game.MagicButtons = loadImages(5)
    game.InfoButtons = loadImages(6)
    game.PlaneButtons = loadImages(7)

    game.GoldFoodMagic = loadImage(34, 0)
    game.NextTurn = loadImage(35, 0)
    game.NextTurnBackground = loadImage(33, 0)

    return outError
}

func MakeGame(wizard setup.WizardCustom) *Game {
    game := &Game{
        active: false,
        Map: MakeMap(),
    }
    return game
}

func (game *Game) IsActive() bool {
    return game.active
}

func (game *Game) Activate() {
    game.active = true
}

func (game *Game) Update(){
}

func (game *Game) Draw(screen *ebiten.Image){
    var options ebiten.DrawImageOptions

    game.Map.Draw(screen)

    // draw hud on top of map
    screen.DrawImage(game.MainHud, &options)

    options.GeoM.Reset()
    x := float64(7)
    y := float64(4)
    options.GeoM.Translate(x, y)
    screen.DrawImage(game.GameButtons[0], &options)

    x += float64(game.GameButtons[0].Bounds().Dx())

    options.GeoM.Translate(float64(game.GameButtons[0].Bounds().Dx()) + 1, 0)
    screen.DrawImage(game.SpellButtons[0], &options)

    options.GeoM.Translate(float64(game.SpellButtons[0].Bounds().Dx()) + 1, 0)
    screen.DrawImage(game.ArmyButtons[0], &options)

    options.GeoM.Translate(float64(game.ArmyButtons[0].Bounds().Dx()) + 1, 0)
    screen.DrawImage(game.CityButtons[0], &options)

    options.GeoM.Translate(float64(game.CityButtons[0].Bounds().Dx()) + 1, 0)
    screen.DrawImage(game.MagicButtons[0], &options)

    options.GeoM.Translate(float64(game.MagicButtons[0].Bounds().Dx()) + 1, 0)
    screen.DrawImage(game.InfoButtons[0], &options)

    options.GeoM.Translate(float64(game.InfoButtons[0].Bounds().Dx()) + 1, 0)
    screen.DrawImage(game.PlaneButtons[0], &options)

    options.GeoM.Reset()
    options.GeoM.Translate(240, 77)
    screen.DrawImage(game.GoldFoodMagic, &options)

    /*
    options.GeoM.Reset()
    options.GeoM.Translate(245, 180)
    screen.DrawImage(game.NextTurnBackground, &options)
    */

    options.GeoM.Reset()
    options.GeoM.Translate(240, 174)
    screen.DrawImage(game.NextTurn, &options)
}
