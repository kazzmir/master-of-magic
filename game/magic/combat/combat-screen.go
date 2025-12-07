package combat

import (
    "fmt"
    "log"
    "cmp"
    "math"
    "math/rand/v2"
    "image"
    "image/color"
    "time"
    "slices"
    "context"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/mouse"
    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/lib/functional"
    globalMouse "github.com/kazzmir/master-of-magic/game/magic/mouse"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/colorm"
)

type CombatState int

const (
    CombatStateRunning CombatState = iota
    CombatStateAttackerWin
    CombatStateDefenderWin
    CombatStateAttackerFlee
    CombatStateDefenderFlee
    CombatStateNoCombat
)

func (state CombatState) IsWinner(team Team) bool {
    switch state {
        case CombatStateAttackerWin: return team == TeamAttacker
        case CombatStateDefenderWin: return team == TeamDefender
        case CombatStateAttackerFlee: return team == TeamDefender
        case CombatStateDefenderFlee: return team == TeamAttacker
    }

    return false
}

func (state CombatState) String() string {
    switch state {
        case CombatStateRunning: return "Running"
        case CombatStateAttackerWin: return "AttackerWin"
        case CombatStateDefenderWin: return "DefenderWin"
        case CombatStateAttackerFlee: return "AttackerFlee"
        case CombatStateDefenderFlee: return "DefenderFlee"
        case CombatStateNoCombat: return "NoCombat"
    }

    return ""
}

type CombatEvent interface {
}

type CombatEventSelectTile struct {
    SelectTile func(int, int)
    CanTarget func(int, int) bool
    Spell spellbook.Spell
    Selecter Team
}

type CombatEventNextUnit struct {
}

type CombatEventGlobalSpell struct {
    Caster ArmyPlayer
    Magic data.MagicType
    Name string
}

type CombatPlaySound struct {
    Sound int
}

type CombatEventSelectUnit struct {
    SelectTarget func(*ArmyUnit)
    CanTarget func(*ArmyUnit) bool
    Spell spellbook.Spell
    Selecter Team
    SelectTeam Team
}

type CombatSelectTargets struct {
    Title string
    Targets []*ArmyUnit
    Select func (*ArmyUnit)
    Army *Army
}

type CombatEventSummonUnit struct {
    Unit *ArmyUnit
}

type CombatEventMessage struct {
    Message string
}

type CombatCreateWallOfFire struct {
}

type CombatCreateWallOfDarkness struct {
}

type CombatDoSingleAuto struct {
}

// FIXME: kind of ugly to need a specific event like this for one projectile type
type CombatEventCreateLightningBolt struct {
    Target *ArmyUnit
    Strength int
}

type CombatUpdates struct {
    SingleAuto bool
}

type MouseState int
const (
    CombatClickHud MouseState = iota
    CombatMoveOk
    CombatMeleeAttackOk
    CombatRangeAttackOk
    CombatNotOk
    CombatCast
)

type DeathAnimationType int
const (
    DeathAnimationNone DeathAnimationType = iota
    DeathColorFade
)

const (
    LightningBoltSound int = 19
)

/* compute the distance between two tiles by moving in one of the 8 directions
 */
func computeTileDistance(x1 int, y1 int, x2 int, y2 int) int {
    distance := 0

    for x1 != x2 || y1 != y2 {
        xDiff := int(math.Abs(float64(x1 - x2)))
        yDiff := int(math.Abs(float64(y1 - y2)))
        if xDiff > 0 && yDiff > 0 {
            distance += 1
            if x1 < x2 {
                x1 += 1
            } else {
                x1 -= 1
            }
            if y1 < y2 {
                y1 += 1
            } else {
                y1 -= 1
            }
        } else if xDiff > 0 && yDiff == 0 {
            distance += 1
            if x1 < x2 {
                x1 += 1
            } else {
                x1 -= 1
            }
        } else if yDiff > 0 && xDiff == 0 {
            distance += 1
            if y1 < y2 {
                y1 += 1
            } else {
                y1 -= 1
            }
        }
    }

    return distance
}

type DamageIndicator struct {
    X int
    Y int
    Offset int
    Damage int // the damage to show
    Life int // how many more frames to show this indicator, counts down to 0
    Count int
}

type CombatDrawFunc func(*ebiten.Image)

type CombatScreen struct {
    Events chan CombatEvent
    Drawer CombatDrawFunc
    ImageCache util.ImageCache
    Cache *lbx.LbxCache
    AudioCache *audio.AudioCache
    Mouse *mouse.MouseData
    WhitePixel *ebiten.Image
    UI *uilib.UI

    Quit context.Context
    Cancel context.CancelFunc

    Fonts CombatFonts

    DrawRoad bool
    DrawClouds bool
    // order to draw tiles in such that they are drawn from the top of the screen to the bottom (painter's order)
    TopDownOrder []image.Point

    Coordinates ebiten.GeoM
    // ScreenToTile ebiten.GeoM
    MouseState MouseState

    DeathAnimation DeathAnimationType

    CameraScale float64
    ExtraControl bool

    ExtraHighlightedUnit *ArmyUnit
    ShowInfoLevel int

    DamageIndicators []DamageIndicator

    Counter uint64

    MouseTileX int
    MouseTileY int

    // if true then the player should select a tile to cast a spell on
    /*
    SelectTile func(int, int)
    */
    DoSelectTile bool

    // if true then the player should select a unit to cast a spell on
    DoSelectUnit bool

    // true while the player is unable to use the main UI buttons (such as while selecting a tile/unit for a spell)
    ButtonsDisabled bool

    // which team to pick a unit from
    // SelectTeam Team
    // invoke this function on the unit that is selected
    /*
    SelectTarget func(*ArmyUnit)
    CanTarget func(*ArmyUnit) bool
    */

    Model *CombatModel
}

func makePaletteFromBanner(banner data.BannerType) color.Palette {
    topColor := banner.Color()

    // red := color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}

    return color.Palette{
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        color.RGBA{R: 0, G: 0, B: 0, A: 0},
        util.Lighten(topColor, 25), util.Lighten(topColor, 18), util.Lighten(topColor, 12),
        topColor, topColor, topColor,
        topColor, topColor, topColor,
    }
}

type CombatFonts struct {
    DebugFont *font.Font
    HudFont *font.Font
    InfoFont *font.Font
    InfoUIFont *font.Font
    WhiteFont *font.Font
    EnchantmentFont *font.Font
    AttackingWizardFont *font.Font
    DefendingWizardFont *font.Font
}

func MakeCombatFonts(cache *lbx.LbxCache, defendingArmy *Army, attackingArmy *Army) CombatFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Unable to load fonts: %v", err)
        return CombatFonts{}
    }

    fontLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        log.Printf("Unable to read fonts.lbx: %v", err)
        return CombatFonts{}
    }

    fonts, err := font.ReadFonts(fontLbx, 0)
    if err != nil {
        log.Printf("Unable to read fonts from fonts.lbx: %v", err)
        return CombatFonts{}
    }

    defendingWizardFont := font.MakeOptimizedFontWithPalette(fonts[4], makePaletteFromBanner(defendingArmy.Player.GetWizard().Banner))
    attackingWizardFont := font.MakeOptimizedFontWithPalette(fonts[4], makePaletteFromBanner(attackingArmy.Player.GetWizard().Banner))

    return CombatFonts{
        DebugFont: loader(fontslib.SmallFont),
        HudFont: loader(fontslib.SmallBlack),
        InfoFont: loader(fontslib.SmallOrange),
        InfoUIFont: loader(fontslib.LightFontSmall),
        WhiteFont: loader(fontslib.SmallerWhite),
        EnchantmentFont: loader(fontslib.MediumOrange),
        AttackingWizardFont: attackingWizardFont,
        DefendingWizardFont: defendingWizardFont,
    }
}

// player is always the human player
func MakeCombatScreen(cache *lbx.LbxCache, defendingArmy *Army, attackingArmy *Army, player ArmyPlayer, landscape CombatLandscape, plane data.Plane, zone ZoneType, influence data.MagicType, overworldX int, overworldY int) *CombatScreen {
    imageCache := util.MakeImageCache(cache)

    fonts := MakeCombatFonts(cache, defendingArmy, attackingArmy)

    whitePixel := ebiten.NewImage(1, 1)
    whitePixel.Fill(color.RGBA{R: 255, G: 255, B: 255, A: 255})

    mouseData, err := mouse.MakeMouseData(cache, &imageCache)
    if err != nil {
        log.Printf("Error loading mouse data: %v", err)
        return nil
    }

    // FIXME: do layout of armys
    var coordinates ebiten.GeoM

    tile0, _ := imageCache.GetImage("cmbgrass.lbx", 0, 0)

    // the battlefield is rotated by 45 degrees
    coordinates.Rotate(-math.Pi / 4)
    // coordinates.Scale(float64(tile0.Bounds().Dx())/2, float64(tile0.Bounds().Dy())/2)
    // FIXME: this math is hacky, but it works for now
    coordinates.Scale(float64(tile0.Bounds().Dx()) * 3 / 4 - 2, float64(tile0.Bounds().Dy()) * 3 / 4 - 1)
    coordinates.Translate(float64(-220), float64(80))

    events := make(chan CombatEvent, 1000)

    allSpells, err := spellbook.ReadSpellsFromCache(cache)
    if err != nil {
        log.Printf("Error reading spells: %v", err)
        allSpells = spellbook.Spells{}
    }

    quit, cancel := context.WithCancel(context.Background())

    combat := &CombatScreen{
        Events: events,
        Cache: cache,
        Quit: quit,
        Cancel: cancel,
        Counter: 1000, // start at a high number so that existing wall of fire/darkness does not show as being newly cast
        AudioCache: audio.MakeAudioCache(cache),
        ImageCache: imageCache,
        Mouse: mouseData,
        CameraScale: 1,
        DrawRoad: zone.City != nil,
        DrawClouds: zone.City != nil && zone.City.HasEnchantment(data.CityEnchantmentFlyingFortress),
        Fonts: fonts,
        Coordinates: coordinates,
        // ScreenToTile: screenToTile,
        WhitePixel: whitePixel,
        DeathAnimation: DeathColorFade,

        Model: MakeCombatModel(allSpells, defendingArmy, attackingArmy, landscape, plane, zone, influence, overworldX, overworldY, events),
    }

    combat.Drawer = combat.NormalDraw
    combat.TopDownOrder = combat.computeTopDownOrder()

    /*
    log.Printf("Top down order: %v", combat.TopDownOrder)

    for i := 0; i < 5; i++ {
        x1, y1 := combat.Coordinates.Apply(float64(combat.TopDownOrder[i].X), float64(combat.TopDownOrder[i].Y))
        log.Printf("tile %v (%v,%v): %v,%v", i, combat.TopDownOrder[i].X, combat.TopDownOrder[i].Y, x1, y1)
    }
    */

    combat.UI = combat.MakeUI(player)

    return combat
}

func (combat *CombatScreen) GetCameraMatrix() ebiten.GeoM {
    return combat.Coordinates
}

func (combat *CombatScreen) ScreenToTile(x float64, y float64) (float64, float64) {
    // tile0, _ := combat.ImageCache.GetImage("cmbgrass.lbx", 0, 0)
    screenToTile := combat.GetCameraMatrix()
    screenToTile.Scale(scale.ScaleAmount, scale.ScaleAmount)
    screenToTile.Invert()

    // return screenToTile.Apply(x - float64(tile0.Bounds().Dx()/3) * combat.CameraScale, y - float64(tile0.Bounds().Dy()/3) * combat.CameraScale)
    return screenToTile.Apply(x, y)
}

func (combat *CombatScreen) computeTopDownOrder() []image.Point {
    var points []image.Point
    for y := 0; y < len(combat.Model.Tiles); y++ {
        for x := 0; x < len(combat.Model.Tiles[y]); x++ {
            points = append(points, image.Pt(x, y))
        }
    }

    matrix := combat.GetCameraMatrix()

    compare := func(a image.Point, b image.Point) int {
        ax, ay := matrix.Apply(float64(a.X), float64(a.Y))
        bx, by := matrix.Apply(float64(b.X), float64(b.Y))

        if ay < by {
            return -1
        }

        if ay > by {
            return 1
        }

        if ax < bx {
            return -1
        }

        if ax > bx {
            return 1
        }

        return 0
    }

    slices.SortFunc(points, compare)
    return points
}

/* a projectile that shoots down from the sky at an angle
 */
func (combat *CombatScreen) createSkyProjectile(target *ArmyUnit, images []*ebiten.Image, explodeImages []*ebiten.Image, effect ProjectileEffect) *Projectile {
    // find where on the screen the unit is
    matrix := combat.GetCameraMatrix()
    screenX, screenY := matrix.Apply(float64(target.X), float64(target.Y))
    screenY -= 10
    screenX += 2

    x := screenX + 80 + rand.Float64() * 60
    y := -rand.Float64() * 40

    // FIXME: make this a parameter?
    speed := 2.2

    angle := math.Atan2(screenY - y, screenX - x)

    // log.Printf("Create fireball projectile at %v,%v -> %v,%v", x, y, screenX, screenY)

    projectile := &Projectile{
        X: x,
        Y: y,
        Speed: speed,
        Angle: angle,
        Target: target,
        TargetX: screenX,
        TargetY: screenY,
        Animation: util.MakeAnimation(images, true),
        Explode: util.MakeAnimation(explodeImages, false),
        Effect: effect,
    }

    return projectile
}

/* a projectile that shoots down from the sky vertically
 */
func (combat *CombatScreen) createVerticalSkyProjectile(target *ArmyUnit, images []*ebiten.Image, explodeImages []*ebiten.Image, effect ProjectileEffect) *Projectile {
    // find where on the screen the unit is
    matrix := combat.GetCameraMatrix()
    screenX, screenY := matrix.Apply(float64(target.X), float64(target.Y))
    screenY -= 10
    screenX += 2

    x := screenX
    y := -40.0

    // FIXME: make this a parameter?
    speed := 2.5

    angle := math.Pi / 2

    // log.Printf("Create fireball projectile at %v,%v -> %v,%v", x, y, screenX, screenY)

    projectile := &Projectile{
        X: x,
        Y: y,
        Speed: speed,
        Angle: angle,
        Target: target,
        TargetX: screenX,
        TargetY: screenY,
        Animation: util.MakeAnimation(images, true),
        Explode: util.MakeAnimation(explodeImages, false),
        Effect: effect,
    }

    return projectile
}

type UnitPosition int
const (
    UnitPositionMiddle UnitPosition = iota
    UnitPositionUnder
)

type Targeting int
const (
    TargetFriend Targeting = iota
    TargetEnemy
    TargetEither
)

/* needs a new name, but creates a projectile that is already at the target
 */
func (combat *CombatScreen) createUnitProjectile(target *ArmyUnit, explodeImages []*ebiten.Image, position UnitPosition, effect ProjectileEffect) *Projectile {
    // find where on the screen the unit is
    matrix := combat.GetCameraMatrix()

    var geom1 ebiten.GeoM

    useImage := explodeImages[0]

    switch position {
        case UnitPositionMiddle:
            // geom1.Translate(14, 3)
            geom1.Translate(-float64(useImage.Bounds().Dx()/2), -float64(useImage.Bounds().Dy()/2))
        case UnitPositionUnder:
            geom1.Translate(0, 9)
            geom1.Translate(-float64(useImage.Bounds().Dx()/2), -float64(useImage.Bounds().Dy()))
    }

    geom1.Scale(combat.CameraScale, combat.CameraScale)
    tx, ty := matrix.Apply(float64(target.X), float64(target.Y))
    geom1.Translate(tx, ty)

    screenX, screenY := geom1.Apply(float64(useImage.Bounds().Dx())/2, float64(useImage.Bounds().Dy())/2)

    // log.Printf("Create fireball projectile at %v,%v -> %v,%v", x, y, screenX, screenY)

    projectile := &Projectile{
        X: screenX,
        Y: screenY,
        Target: target,
        Speed: 0,
        Angle: 0,
        Effect: effect,
        TargetX: screenX,
        TargetY: screenY,
        Animation: nil,
        Explode: util.MakeAnimation(explodeImages, false),
        // start in an exploding state because there is no other animation to show
        Exploding: true,
    }

    return projectile
}

func (combat *CombatScreen) CreateIceBoltProjectile(target *ArmyUnit, strength int) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 11)

    loopImages := images[0:3]
    explodeImages := images[3:]

    effect := combat.Model.CreateIceBoltProjectileEffect(strength, combat)

    return combat.createSkyProjectile(target, loopImages, explodeImages, effect)
}

func (combat *CombatScreen) CreateFireBoltProjectile(target *ArmyUnit, strength int) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 0)
    loopImages := images[0:3]
    explodeImages := images[3:]

    effect := combat.Model.CreateFireBoltProjectileEffect(strength, combat)

    return combat.createSkyProjectile(target, loopImages, explodeImages, effect)
}

func (combat *CombatScreen) CreateFireballProjectile(target *ArmyUnit, strength int) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 23)

    loopImages := images[0:11]
    explodeImages := images[11:]

    effect := combat.Model.CreateFireballProjectileEffect(strength, combat)

    return combat.createSkyProjectile(target, loopImages, explodeImages, effect)
}

func (combat *CombatScreen) CreateStarFiresProjectile(target *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 9)
    explodeImages := images

    effect := combat.Model.CreateStarFiresProjectileEffect(combat)

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateDispelEvilProjectile(target *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 10)
    explodeImages := images

    effect := combat.Model.CreateDispelEvilProjectileEffect(combat)

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreatePsionicBlastProjectile(target *ArmyUnit, strength int) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 16)
    explodeImages := images

    effect := combat.Model.CreatePsionicBlastProjectileEffect(strength, combat)

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateDoomBoltProjectile(target *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 5)
    loopImages := images[0:3]
    explodeImages := images[3:]

    effect := combat.Model.CreateDoomBoltProjectileEffect(combat)

    return combat.createVerticalSkyProjectile(target, loopImages, explodeImages, effect)
}

func (combat *CombatScreen) CreateLightningBoltProjectile(target *ArmyUnit, strength int) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 24)
    // loopImages := images
    explodeImages := images

    matrix := combat.GetCameraMatrix()
    screenX, screenY := matrix.Apply(float64(target.X), float64(target.Y))

    screenY -= float64(images[0].Bounds().Dy())/2
    screenX += float64(images[0].Bounds().Dx())/2

    effect := combat.Model.CreateLightningBoltProjectileEffect(strength, combat)

    projectile := &Projectile{
        X: screenX,
        Y: screenY,
        Target: target,
        Speed: 0,
        Angle: 0,
        TargetX: screenX,
        TargetY: screenY,
        Animation: util.MakeAnimation(images, true),
        Explode: util.MakeRepeatAnimation(explodeImages, 2),
        Exploding: true,
        Effect: effect,
    }

    return projectile
}

func (combat *CombatScreen) CreateWarpLightningProjectile(target *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 3)
    // loopImages := images
    explodeImages := images

    matrix := combat.GetCameraMatrix()
    screenX, screenY := matrix.Apply(float64(target.X), float64(target.Y))
    // screenY += 13
    screenX += 3

    // screenY -= float64(images[0].Bounds().Dy())

    effect := combat.Model.CreateWarpLightningProjectileEffect(combat)

    projectile := &Projectile{
        X: screenX,
        Y: screenY,
        Target: target,
        Speed: 0,
        Angle: 0,
        TargetX: screenX,
        TargetY: screenY,
        Animation: util.MakeAnimation(images, true),
        Explode: util.MakeRepeatAnimation(explodeImages, 2),
        Exploding: true,
        Effect: effect,
    }

    return projectile
}

// player will never be nil, but unitCaster might be nil if the player is casting the spell
// if a hero/unit is casting the spell then unitCaster will be non-nil
func (combat *CombatScreen) CreateLifeDrainProjectile(target *ArmyUnit, reduceResistance int, player ArmyPlayer, unitCaster *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 6)
    explodeImages := images

    effect := combat.Model.CreateLifeDrainProjectileEffect(reduceResistance, player, unitCaster, combat)

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateFlameStrikeProjectile(target *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 33)
    explodeImages := images

    effect := combat.Model.CreateFlameStrikeProjectileEffect(combat)

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateRecallHeroProjectile(target *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 5)
    explodeImages := images

    effect := combat.Model.CreateRecallHeroProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateHealingProjectile(target *ArmyUnit) *Projectile {
    // FIXME: the images should be mostly with with transparency
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 3)
    explodeImages := images

    effect := combat.Model.CreateHealingProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateHeroismProjectile(target *ArmyUnit) *Projectile {
    // FIXME: the images should be mostly with with transparency
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 3)
    explodeImages := images

    effect := combat.Model.CreateHeroismProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateHolyArmorProjectile(target *ArmyUnit) *Projectile {
    // FIXME: the images should be mostly with with transparency
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 3)
    explodeImages := images

    effect := combat.Model.CreateHolyArmorProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateInvulnerabilityProjectile(target *ArmyUnit) *Projectile {
    // FIXME: the images should be mostly with with transparency
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 3)
    explodeImages := images

    effect := combat.Model.CreateInvulnerabilityProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateLionHeartProjectile(target *ArmyUnit) *Projectile {
    // FIXME: the images should be mostly with with transparency
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 3)
    explodeImages := images

    effect := combat.Model.CreateLionHeartProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateTrueSightProjectile(target *ArmyUnit) *Projectile {
    // FIXME: the images should be mostly with with transparency
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 3)
    explodeImages := images

    effect := combat.Model.CreateTrueSightProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateElementalArmorProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 0)
    explodeImages := images

    effect := combat.Model.CreateElementalArmorProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateGiantStrengthProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 0)
    explodeImages := images

    effect := combat.Model.CreateGiantStrengthProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateIronSkinProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 0)
    explodeImages := images

    effect := combat.Model.CreateIronSkinProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateStoneSkinProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 0)
    explodeImages := images

    effect := combat.Model.CreateStoneSkinProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateRegenerationProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 0)
    explodeImages := images

    effect := combat.Model.CreateRegenerationProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateResistElementsProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 0)
    explodeImages := images

    effect := combat.Model.CreateResistElementsProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateRighteousnessProjectile(target *ArmyUnit) *Projectile {
    // FIXME: the images should be mostly with with transparency
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 3)
    explodeImages := images

    effect := combat.Model.CreateRighteousnessProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateHolyWeaponProjectile(target *ArmyUnit) *Projectile {
    // FIXME: the images should be mostly with with transparency
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 3)
    explodeImages := images

    effect := combat.Model.CreateHolyWeaponProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateFlightProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 1)
    explodeImages := images

    effect := combat.Model.CreateFlightProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateGuardianWindProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 1)
    explodeImages := images

    effect := combat.Model.CreateGuardianWindProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateHasteProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 1)
    explodeImages := images

    effect := combat.Model.CreateHasteProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateInvisibilityProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 1)
    explodeImages := images

    effect := combat.Model.CreateInvisibilityProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateMagicImmunityProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 1)
    explodeImages := images

    effect := combat.Model.CreateMagicImmunityProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateResistMagicProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 1)
    explodeImages := images

    effect := combat.Model.CreateResistMagicProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateSpellLockProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 1)
    explodeImages := images

    effect := combat.Model.CreateSpellLockProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateEldritchWeaponProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 2)
    explodeImages := images

    effect := combat.Model.CreateEldritchWeaponProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateFlameBladeProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 2)
    explodeImages := images

    effect := combat.Model.CreateFlameBladeProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateImmolationProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 2)
    explodeImages := images

    effect := combat.Model.CreateImmolationProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateBerserkProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 4)
    explodeImages := images

    effect := combat.Model.CreateBerserkProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateCloakOfFearProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 4)
    explodeImages := images

    effect := combat.Model.CreateCloakOfFearProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateWraithFormProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify this animation
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 4)
    explodeImages := images

    effect := combat.Model.CreateWraithFormProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateChaosChannelsProjectile(target *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 2)
    explodeImages := images

    effect := combat.Model.CreateChaosChannelsProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateBlessProjectile(target *ArmyUnit) *Projectile {
    // FIXME: the images should be mostly with with transparency
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 3)
    explodeImages := images

    effect := combat.Model.CreateBlessProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateWeaknessProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 5)
    explodeImages := images

    effect := combat.Model.CreateWeaknessProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateCreatureBindingProjectile(target *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 1)
    explodeImages := images

    effect := combat.Model.CreateCreatureBindingProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreatePetrifyProjectile(target *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 12)
    explodeImages := images

    effect := combat.Model.CreatePetrifyProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreatePossessionProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 8)
    explodeImages := images

    effect := combat.Model.CreatePossessionProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateConfusionProjectile(target *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 20)
    explodeImages := images

    effect := combat.Model.CreateConfusionProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateBlackSleepProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 5)
    explodeImages := images

    effect := combat.Model.CreateBlackSleepProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateVertigoProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 17)
    explodeImages := images

    effect := combat.Model.CreateVertigoProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateShatterProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify
    images, _ := combat.ImageCache.GetImages("resource.lbx", 79)
    explodeImages := images

    effect := combat.Model.CreateShatterProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateWarpCreatureProjectile(target *ArmyUnit) *Projectile {
    // FIXME: verify
    images, _ := combat.ImageCache.GetImages("resource.lbx", 81)
    explodeImages := images

    effect := combat.Model.CreateWarpCreatureProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateHolyWordProjectile(target *ArmyUnit) *Projectile {
    // FIXME: the images should be mostly with with transparency
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 3)
    explodeImages := images

    effect := combat.Model.CreateHolyWordProjectileEffect(combat)

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateWebProjectile(target *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 13)
    explodeImages := images

    effect := combat.Model.CreateWebProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateDeathSpellProjectile(target *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 14)
    explodeImages := images

    effect := combat.Model.CreateDeathSpellProjectileEffect(combat)

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateWordOfDeathProjectile(target *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 14)
    explodeImages := images

    effect := combat.Model.CreateWordOfDeathProjectileEffect(combat)

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateWarpWoodProjectile(target *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 2)
    explodeImages := images

    effect := combat.Model.CreateWarpWoodProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateDisintegrateProjectile(target *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 4)
    explodeImages := images

    effect := combat.Model.CreateDisintegrateProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateWordOfRecallProjectile(target *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("specfx.lbx", 1)
    explodeImages := images

    effect := combat.Model.CreateWordOfRecallProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateDispelMagicProjectile(target *ArmyUnit, caster ArmyPlayer, dispelStrength int) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 26)
    explodeImages := images

    effect := combat.Model.CreateDispelMagicProjectileEffect(caster, dispelStrength)

    return combat.createUnitProjectile(target, explodeImages, UnitPositionMiddle, effect)
}

func (combat *CombatScreen) CreateCracksCallProjectile(target *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 15)
    explodeImages := images

    effect := combat.Model.CreateCracksCallProjectileEffect()

    return combat.createUnitProjectile(target, explodeImages, UnitPositionUnder, effect)
}

func (combat *CombatScreen) CreateBanishProjectile(target *ArmyUnit, reduceResistance int) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 19)
    explodeImages := images
    return combat.createUnitProjectile(target, explodeImages, UnitPositionUnder, combat.Model.CreateBanishProjectileEffect(reduceResistance, combat))
}

func (combat *CombatScreen) CreateMindStormProjectile(target *ArmyUnit) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 21)
    explodeImages := images

    return combat.createUnitProjectile(target, explodeImages, UnitPositionUnder, combat.Model.CreateMindStormProjectileEffect())
}

func (combat *CombatScreen) CreateDisruptProjectile(x int, y int) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 1)

    explodeImages := images

    fakeTarget := ArmyUnit{
        X: x,
        Y: y,
    }

    // TODO

    return combat.createUnitProjectile(&fakeTarget, explodeImages, UnitPositionUnder, func (*ArmyUnit){})
}

func (combat *CombatScreen) CreateSummoningCircle(x int, y int) *Projectile {
    images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 22)
    explodeImages := images

    fakeTarget := ArmyUnit{
        X: x,
        Y: y,
    }

    return combat.createUnitProjectile(&fakeTarget, explodeImages, UnitPositionUnder, func (*ArmyUnit){})
}

func (combat *CombatScreen) CreateMagicVortex(x int, y int) *OtherUnit {
    images, _ := combat.ImageCache.GetImages("cmbmagic.lbx", 120)

    unit := &OtherUnit{
        X: x,
        Y: y,
        Animation: util.MakeAnimation(images, true),
    }

    return unit
}

func (combat *CombatScreen) AddSelectTargetsElements(targets []*ArmyUnit, title string, selected func(*ArmyUnit)) {
    var selections []uilib.Selection

    for _, target := range targets {
        selections = append(selections, uilib.Selection{
            Name: target.Unit.GetName(),
            Action: func(){
                selected(target)
            },
        })
    }

    combat.UI.AddElements(uilib.MakeSelectionUI(combat.UI, combat.Cache, &combat.ImageCache, 100, 20, title, selections, true))
}

type UnitCaster struct {
}

func (caster *UnitCaster) ComputeTurnsToCast(cost int) int {
    // shouldn't matter in combat
    return 1
}

func (caster *UnitCaster) ComputeEffectiveResearchPerTurn(research float64, spell spellbook.Spell) int {
    return int(research)
}

func (caster *UnitCaster) ComputeEffectiveSpellCost(spell spellbook.Spell, overland bool) int {
    return spell.Cost(overland)
}

func (combat *CombatScreen) MakeInfoUI(remove func()) *uilib.UIElementGroup {
    group := uilib.MakeGroup()

    boxTop, _ := combat.ImageCache.GetImage("compix.lbx", 58, 0)
    // chop off the bottom few pixels because there is a border there that we don't need
    boxTop = boxTop.SubImage(image.Rect(0, 0, boxTop.Bounds().Dx(), boxTop.Bounds().Dy() - 5)).(*ebiten.Image)
    // a part of the top box that we can replicate as many times as needed
    boxPiece := boxTop.SubImage(image.Rect(0, 10, boxTop.Bounds().Dx(), 30)).(*ebiten.Image)
    boxBottom, _ := combat.ImageCache.GetImage("compix.lbx", 56, 0)

    fader := group.MakeFadeIn(7)

    type Check struct {
        Exists func() bool
        ImageIndex int
        Text string
    }

    hasGlobalEnchantment := func (enchantment data.Enchantment) bool {
        return combat.Model.AttackingArmy.Player.HasEnchantment(enchantment) || combat.Model.DefendingArmy.Player.HasEnchantment(enchantment)
    }

    makeGlobalCheck := func (enchantment data.Enchantment) func() bool {
        return func() bool {
            return hasGlobalEnchantment(enchantment)
        }
    }

    makeTownCheck := func (enchantment data.CityEnchantment) func() bool {
        return func() bool {
            city := combat.Model.Zone.City
            return city != nil && city.HasEnchantment(enchantment)
        }
    }

    checks := []Check{
        Check{
            Exists: makeGlobalCheck(data.EnchantmentCrusade),
            ImageIndex: 42,
            Text: "Crusade",
        },
        Check{
            Exists: makeGlobalCheck(data.EnchantmentHolyArms),
            ImageIndex: 43,
            Text: "Holy Arms",
        },
        Check{
            Exists: makeTownCheck(data.CityEnchantmentHeavenlyLight),
            ImageIndex: 44,
            Text: "Heavenly Light",
        },
        Check{
            Exists: makeGlobalCheck(data.EnchantmentCharmOfLife),
            ImageIndex: 45,
            Text: "Charm of Life",
        },
        Check{
            Exists: makeGlobalCheck(data.EnchantmentChaosSurge),
            ImageIndex: 46,
            Text: "Chaos Surge",
        },
        Check{
            Exists: makeGlobalCheck(data.EnchantmentEternalNight),
            ImageIndex: 49,
            Text: "Eternal Night",
        },
        Check{
            Exists: makeTownCheck(data.CityEnchantmentCloudOfShadow),
            ImageIndex: 50,
            Text: "Cloud of Shadow",
        },
        Check{
            Exists: makeGlobalCheck(data.EnchantmentZombieMastery),
            ImageIndex: 51,
            Text: "Zombie Mastery",
        },
    }

    totalRows := 0
    if combat.Model.InsideMagicNode() {
        totalRows += 1
    }

    countChecks := 0
    for _, check := range checks {
        if check.Exists() {
            countChecks += 1
        }
    }

    totalRows += (countChecks + 1) / 2
    rowSize := 20

    totalHeightNeeded := 10 + totalRows * rowSize + 10

    extraPieces := int(math.Ceil(float64(max(0, totalHeightNeeded - boxTop.Bounds().Dy() - boxBottom.Bounds().Dy())) / float64(boxPiece.Bounds().Dy())))

    // prerender the entire ui since it never changes
    background := ebiten.NewImage(boxTop.Bounds().Dx(), boxTop.Bounds().Dy() + boxPiece.Bounds().Dy() * extraPieces + boxBottom.Bounds().Dy())
    var backgroundOptions ebiten.DrawImageOptions
    background.DrawImage(boxTop, &backgroundOptions)
    backgroundOptions.GeoM.Translate(0, float64(boxTop.Bounds().Dy()))
    for range extraPieces {
        background.DrawImage(boxPiece, &backgroundOptions)
        backgroundOptions.GeoM.Translate(0, float64(boxPiece.Bounds().Dy()))
    }
    background.DrawImage(boxBottom, &backgroundOptions)

    rect := util.ImageRect(30, 40, background)

    row := 0

    if combat.Model.InsideMagicNode() {
        dispelIndex := -1
        auraIndex := -1
        switch combat.Model.Zone.GetMagic() {
            case data.SorceryMagic:
                dispelIndex = 54
                auraIndex = 55
            case data.NatureMagic:
                dispelIndex = 52
                auraIndex = 53
            case data.ChaosMagic:
                dispelIndex = 47
                auraIndex = 48
        }

        y := 10 + row * rowSize
        x := 10

        dispelImage, err := combat.ImageCache.GetImage("compix.lbx", dispelIndex, 0)
        if err == nil {
            backgroundOptions.GeoM.Reset()
            backgroundOptions.GeoM.Translate(float64(x), float64(y))
            background.DrawImage(dispelImage, &backgroundOptions)
            x += dispelImage.Bounds().Dx() + 2
        }

        combat.Fonts.InfoUIFont.PrintOptions(background, float64(x), float64(y + 3), font.FontOptions{DropShadow: true}, fmt.Sprintf("Dispells Non-%v", combat.Model.Zone.GetMagic()))

        x = 120
        auraImage, err := combat.ImageCache.GetImage("compix.lbx", auraIndex, 0)
        if err == nil {
            backgroundOptions.GeoM.Reset()
            backgroundOptions.GeoM.Translate(float64(x), float64(y))
            background.DrawImage(auraImage, &backgroundOptions)
            x += auraImage.Bounds().Dx() + 2
        }

        combat.Fonts.InfoUIFont.PrintOptions(background, float64(x), float64(y + 3), font.FontOptions{DropShadow: true}, fmt.Sprintf("%v Node Aura", combat.Model.Zone.GetMagic()))

        row += 1
    }

    column := 0
    columnSize := 100

    for _, check := range checks {
        if check.Exists() {
            y := 10 + row * rowSize
            x := 10 + column * columnSize
            pic, _ := combat.ImageCache.GetImage("compix.lbx", check.ImageIndex, 0)
            backgroundOptions.GeoM.Reset()
            backgroundOptions.GeoM.Translate(float64(x), float64(y))
            background.DrawImage(pic, &backgroundOptions)

            x += pic.Bounds().Dx() + 3

            combat.Fonts.InfoUIFont.PrintOptions(background, float64(x), float64(y + 2), font.FontOptions{DropShadow: true}, check.Text)

            column += 1
            if column >= 2 {
                row += 1
                column = 0
            }
        }
    }

    clicked := false
    group.AddElement(&uilib.UIElement{
        Layer: 1,
        Rect: rect,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(fader())
            options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
            scale.DrawScaled(screen, background, &options)
        },
        LeftClick: func(element *uilib.UIElement) {
            if !clicked {
                clicked = true
                fader = group.MakeFadeOut(7)
                group.AddDelay(7, remove)
            }
        },
    })

    return group
}

func (combat *CombatScreen) IsSelectingSpell() bool {
    return combat.DoSelectUnit || combat.DoSelectTile
}

func (combat *CombatScreen) MakeUI(player ArmyPlayer) *uilib.UI {
    var elements []*uilib.UIElement

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            hudImage, _ := combat.ImageCache.GetImage("backgrnd.lbx", 3, 0)
            options.GeoM.Reset()
            options.GeoM.Translate(0, float64(200 - hudImage.Bounds().Dy()))
            scale.DrawScaled(screen, hudImage, &options)

            if combat.Model.AttackingArmy.Player == player && (combat.DoSelectUnit || combat.DoSelectTile) {
            } else {
                combat.Fonts.AttackingWizardFont.PrintOptions(screen, 280, 167, font.FontOptions{Justify: font.FontJustifyCenter, Scale: scale.ScaleAmount, DropShadow: true}, combat.Model.AttackingArmy.Player.GetWizard().Name)

                options.GeoM.Reset()
                options.GeoM.Translate(246, 179)
                for _, enchantment := range combat.Model.AttackingArmy.Enchantments {
                    image, _ := combat.ImageCache.GetImage("compix.lbx", enchantment.LbxIndex(), 0)
                    scale.DrawScaled(screen, image, &options)
                    options.GeoM.Translate(float64(image.Bounds().Dx()), 0)
                }
            }

            humanArmy := combat.Model.GetHumanArmy()
            y := 173
            right := 239
            combat.Fonts.HudFont.PrintOptions(screen, float64(200), float64(y), font.FontOptions{Scale: scale.ScaleAmount}, "Skill:")
            combat.Fonts.HudFont.PrintOptions(screen, float64(right), float64(y), font.FontOptions{Scale: scale.ScaleAmount, Justify: font.FontJustifyRight}, fmt.Sprintf("%v", humanArmy.ManaPool))
            y += combat.Fonts.HudFont.Height() + 2

            combat.Fonts.HudFont.PrintOptions(screen, float64(200), float64(y), font.FontOptions{Scale: scale.ScaleAmount}, "Mana:")
            combat.Fonts.HudFont.PrintOptions(screen, float64(right), float64(y), font.FontOptions{Scale: scale.ScaleAmount, Justify: font.FontJustifyRight}, fmt.Sprintf("%v", humanArmy.Player.GetMana()))
            y += combat.Fonts.HudFont.Height() + 2

            combat.Fonts.HudFont.PrintOptions(screen, float64(200), float64(y), font.FontOptions{Scale: scale.ScaleAmount}, "Range:")
            combat.Fonts.HudFont.PrintOptions(screen, float64(right), float64(y), font.FontOptions{Scale: scale.ScaleAmount, Justify: font.FontJustifyRight}, fmt.Sprintf("%vx", humanArmy.Range.ToFloat()))

            if combat.Model.DefendingArmy.Player == player && (combat.DoSelectUnit || combat.DoSelectTile) {
            } else {
                combat.Fonts.DefendingWizardFont.PrintOptions(screen, 40, 167, font.FontOptions{Scale: scale.ScaleAmount, Justify: font.FontJustifyCenter, DropShadow: true}, combat.Model.DefendingArmy.Player.GetWizard().Name)

                options.GeoM.Reset()
                options.GeoM.Translate(float64(7), float64(179))
                for _, enchantment := range combat.Model.DefendingArmy.Enchantments {
                    image, _ := combat.ImageCache.GetImage("compix.lbx", enchantment.LbxIndex(), 0)
                    scale.DrawScaled(screen, image, &options)
                    options.GeoM.Translate(float64(image.Bounds().Dx()), 0)
                }
            }

            if combat.Model.SelectedUnit != nil && combat.IsUnitVisible(combat.Model.SelectedUnit) {

                rightImage, _ := combat.ImageCache.GetImageTransform(combat.Model.SelectedUnit.Unit.GetCombatLbxFile(), combat.Model.SelectedUnit.Unit.GetCombatIndex(units.FacingRight), 0, player.GetWizard().Banner.String(), units.MakeUpdateUnitColorsFunc(player.GetWizard().Banner))
                options.GeoM.Reset()
                options.GeoM.Translate(85, 170)
                scale.DrawScaled(screen, rightImage, &options)

                combat.Fonts.HudFont.PrintOptions(screen, 96, 166, font.FontOptions{Scale: scale.ScaleAmount}, combat.Model.SelectedUnit.Unit.GetName())

                plainAttack, _ := combat.ImageCache.GetImage("compix.lbx", 29, 0)
                options.GeoM.Reset()
                options.GeoM.Translate(130, 173)
                scale.DrawScaled(screen, plainAttack, &options)
                combat.Fonts.HudFont.PrintOptions(screen, 130, 174, font.FontOptions{Scale: scale.ScaleAmount, Justify: font.FontJustifyRight}, fmt.Sprintf("%v", combat.Model.SelectedUnit.GetMeleeAttackPower()))

                if combat.Model.SelectedUnit.CanRangeAttack() {
                    y := float64(180)
                    switch combat.Model.SelectedUnit.Unit.GetRangedAttackDamageType() {
                        case units.DamageRangedPhysical:
                            arrow, _ := combat.ImageCache.GetImage("compix.lbx", 34, 0)
                            options.GeoM.Reset()
                            options.GeoM.Translate(float64(130), y)
                            scale.DrawScaled(screen, arrow, &options)
                            combat.Fonts.HudFont.PrintOptions(screen, 130, y+float64(2), font.FontOptions{Scale: scale.ScaleAmount, Justify: font.FontJustifyRight}, fmt.Sprintf("%v", combat.Model.SelectedUnit.GetRangedAttackPower()))
                        case units.DamageRangedBoulder:
                            boulder, _ := combat.ImageCache.GetImage("compix.lbx", 35, 0)
                            options.GeoM.Reset()
                            options.GeoM.Translate(float64(130), y)
                            scale.DrawScaled(screen, boulder, &options)
                            combat.Fonts.HudFont.PrintOptions(screen, 130, y+float64(2), font.FontOptions{Scale: scale.ScaleAmount, Justify: font.FontJustifyRight}, fmt.Sprintf("%v", combat.Model.SelectedUnit.GetRangedAttackPower()))
                        case units.DamageRangedMagical:
                            magic, _ := combat.ImageCache.GetImage("compix.lbx", 30, 0)
                            options.GeoM.Reset()
                            options.GeoM.Translate(float64(130), y)
                            scale.DrawScaled(screen, magic, &options)
                            combat.Fonts.HudFont.PrintOptions(screen, 130, y+float64(2), font.FontOptions{Scale: scale.ScaleAmount, Justify: font.FontJustifyRight}, fmt.Sprintf("%v", combat.Model.SelectedUnit.GetRangedAttackPower()))
                    }
                }

                var movementImage *ebiten.Image
                if combat.Model.SelectedUnit.IsFlying() {
                    movementImage, _ = combat.ImageCache.GetImage("compix.lbx", 39, 0)
                } else {
                    movementImage, _ = combat.ImageCache.GetImage("compix.lbx", 38, 0)
                }

                options.GeoM.Reset()
                options.GeoM.Translate(130, 188)
                scale.DrawScaled(screen, movementImage, &options)
                combat.Fonts.HudFont.PrintOptions(screen, 130, 190, font.FontOptions{Justify: font.FontJustifyRight, Scale: scale.ScaleAmount}, fmt.Sprintf("%v", combat.Model.SelectedUnit.MovesLeft.ToFloat()))

                combat.DrawHealthBar(screen, 123, 197, 255, combat.Model.SelectedUnit)
            }

            ui.StandardDraw(screen)
        },
    }

    buttonX := float64(144)
    buttonY := float64(168)

    makeButton2 := func(lbxIndex int, buttonDisabledIndex, x int, y int, action func(), alterColor func(*colorm.ColorM)) *uilib.UIElement {
        buttons, _ := combat.ImageCache.GetImages("compix.lbx", lbxIndex)
        buttonDisabled, _ := combat.ImageCache.GetImage("compix.lbx", buttonDisabledIndex, 0)
        rect := image.Rect(0, 0, buttons[0].Bounds().Dx(), buttons[0].Bounds().Dy()).Add(image.Point{int(buttonX) + buttons[0].Bounds().Dx() * x, int(buttonY) + buttons[0].Bounds().Dy() * y})
        index := 0
        var options colorm.DrawImageOptions
        options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
        options.GeoM = scale.ScaleGeom(options.GeoM)
        return &uilib.UIElement{
            Rect: rect,
            LeftClick: func(element *uilib.UIElement){
                if combat.ButtonsDisabled {
                    return
                }

                index = 1
            },
            LeftClickRelease: func(element *uilib.UIElement){
                if combat.ButtonsDisabled {
                    return
                }

                action()
                index = 0
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var extraColor colorm.ColorM
                alterColor(&extraColor)
                if combat.ButtonsDisabled {
                    colorm.DrawImage(screen, buttonDisabled, extraColor, &options)
                } else {
                    colorm.DrawImage(screen, buttons[index], extraColor, &options)
                }
            },
        }
    }

    makeButton := func(lbxIndex int, buttonDisabledIndex, x int, y int, action func()) *uilib.UIElement {
        return makeButton2(lbxIndex, buttonDisabledIndex, x, y, action, func(_ *colorm.ColorM){})
    }

    // spell
    spellPage := 0
    elements = append(elements, makeButton(1, 23, 0, 0, func(){
        // cannot cast if the player is selecting a unit/tile
        if combat.IsSelectingSpell() {
            return
        }

        army := combat.Model.GetArmyForPlayer(player)

        defendingCity := combat.Model.Zone.City != nil && army == combat.Model.DefendingArmy

        doPlayerSpell := func(){
            // FIXME: this check should be done earlier so that we don't even let the player pick a spell
            if army.Casted {
                return
            }

            // the lower of the mana pool (casting skill) or the wizard's mana divided by the range
            minimumMana := min(army.ManaPool, int(float64(army.Player.GetMana()) / army.Range.ToFloat()))

            spellUI := spellbook.MakeSpellBookCastUI(ui, combat.Cache, player.GetKnownSpells().CombatSpells(defendingCity), make(map[spellbook.Spell]int), minimumMana, spellbook.Spell{}, 0, false, player, &spellPage, func (spell spellbook.Spell, picked bool){
                if picked {
                    // player mana and skill should go down accordingly
                    combat.Model.InvokeSpell(combat, combat.Model.GetArmyForPlayer(player), nil, spell, func(success bool){
                        spellCost := player.ComputeEffectiveSpellCost(spell, false)
                        army.Casted = true
                        army.ManaPool -= spellCost
                        player.UseMana(int(float64(spellCost) * army.Range.ToFloat()))
                        if success {
                            combat.Model.AddLogEvent(fmt.Sprintf("%v casts %v", player.GetWizard().Name, spell.Name))
                            combat.PlaySound(spell)
                        }
                    })
                }
            })
            ui.AddElements(spellUI)
        }

        playerOnly := true
        if combat.Model.SelectedUnit != nil {
            // FIXME: if player is out of mana then just select unit spell?

            if combat.Model.SelectedUnit.CanCast() {
                playerOnly = false
                selections := []uilib.Selection{
                    uilib.Selection{
                        Name: player.GetWizard().Name,
                        Action: doPlayerSpell,
                    },
                    uilib.Selection{
                        Name: combat.Model.SelectedUnit.Unit.GetName(),
                        Action: func(){
                            unitSpells := combat.Model.SelectedUnit.Spells
                            caster := combat.Model.SelectedUnit

                            // spell casting range for a unit is always 1

                            doCast := func(spell spellbook.Spell){
                                combat.Model.InvokeSpell(combat, combat.Model.GetArmyForPlayer(player), caster, spell, func(success bool){
                                    charge, hasCharge := caster.SpellCharges[spell]
                                    if hasCharge && charge > 0 {
                                        caster.SpellCharges[spell] -= 1
                                    } else {
                                        // units pay the full cost of a spell with no modifiers
                                        caster.CastingSkill -= float32(spell.Cost(false))
                                    }
                                    caster.Casted = true
                                    if success {
                                        combat.Model.AddLogEvent(fmt.Sprintf("%v casts %v", caster.Unit.GetName(), spell.Name))
                                        combat.PlaySound(spell)
                                    }
                                    caster.MovesLeft = fraction.FromInt(0)
                                    select {
                                        case combat.Events <- &CombatEventNextUnit{}:
                                        default:
                                    }
                                })
                            }

                            if len(unitSpells.Spells) == 0 {
                                available := 0
                                var use spellbook.Spell
                                for spell, charge := range caster.SpellCharges {
                                    if charge > 0 {
                                        available += 1
                                        use = spell
                                    }
                                }
                                if available == 1 {
                                    doCast(use)
                                    return
                                }
                            }

                            // just invoke the one spell
                            if len(unitSpells.Spells) == 1 {
                                spell := unitSpells.Spells[0]
                                doCast(spell)
                                return
                            }

                            // what is casting skill based on for a unit?
                            spellUI := spellbook.MakeSpellBookCastUI(ui, combat.Cache, unitSpells.CombatSpells(defendingCity), caster.SpellCharges, int(caster.CastingSkill), spellbook.Spell{}, 0, false, &UnitCaster{}, &spellPage, func (spell spellbook.Spell, picked bool){
                                if picked {
                                    doCast(spell)
                                }
                            })
                            ui.AddElements(spellUI)
                        },
                    },
                    uilib.Selection{
                        Name: "Cancel",
                        Action: func(){
                        },
                    },
                }

                ui.AddElements(uilib.MakeSelectionUI(ui, combat.Cache, &combat.ImageCache, 100, 50, "Who Will Cast", selections, true))
            }
        }

        if playerOnly {
            doPlayerSpell()
        }
    }))

    // wait
    elements = append(elements, makeButton(2, 24, 1, 0, func(){
        combat.Model.NextUnit()
    }))

    // info
    elements = append(elements, makeButton(20, 25, 0, 1, func(){
        var group *uilib.UIElementGroup
        remove := func(){
            ui.RemoveGroup(group)
        }
        group = combat.MakeInfoUI(remove)
        ui.AddGroup(group)
    }))

    // auto
    elements = append(elements, makeButton2(4, 26, 1, 1, func(){
        if combat.ExtraControl {
            combat.Events <- &CombatDoSingleAuto{}
            return
        }

        if combat.Model.AttackingArmy.Player == player {
            combat.Model.AttackingArmy.Auto = true
        } else {
            combat.Model.DefendingArmy.Auto = true
        }
    }, func(color *colorm.ColorM){
        if combat.ExtraControl {
            color.Translate(0, 0, 0.9, 0)
            color.Scale(0.9, 0.9, 1, 1)
        }
    }))

    // flee
    elements = append(elements, makeButton(21, 27, 0, 2, func(){

        doFlee := func() {
            if combat.Model.AttackingArmy.Player == player {
                combat.Model.AttackingArmy.Fled = true
            } else {
                combat.Model.DefendingArmy.Fled = true
            }
        }

        confirm := uilib.MakeConfirmDialog(ui, combat.Cache, &combat.ImageCache, "Do you wish to flee?", false, doFlee, func(){})
        ui.AddElements(confirm)
    }))

    // done
    elements = append(elements, makeButton(3, 28, 1, 2, func(){
        combat.Model.DoneTurn()
    }))

    ui.SetElementsFromArray(elements)

    return ui
}

/* check that 'check' is between angle-spread and angle+spread
 */
func betweenAngle(check float64, angle float64, spread float64) bool {
    minAngle := angle - spread
    maxAngle := angle + spread

    for minAngle < 0 {
        minAngle += math.Pi * 2
        maxAngle += math.Pi * 2
    }

    // at least make 'check' positive
    for check < 0 {
        check += math.Pi * 2
    }

    // if minAngle is above pi, then try to move 'check' to be in the same range
    if check < minAngle {
        check += math.Pi * 2
    // same for max angle
    } else if check > maxAngle {
        check -= math.Pi * 2
    }

    // now check if 'check' is between min and max
    return check >= minAngle && check <= maxAngle
}

func (combat *CombatScreen) TileIsEmpty(x int, y int) bool {
    unit := combat.Model.GetUnit(x, y)
    if unit != nil && unit.GetHealth() > 0 {
        return false
    }
    /*
    for _, unit := range combat.DefendingArmy.Units {
        if unit.Health > 0 && unit.X == x && unit.Y == y {
            return false
        }
    }

    for _, unit := range combat.AttackingArmy.Units {
        if unit.Health > 0 && unit.X == x && unit.Y == y {
            return false
        }
    }
    */

    return true
}

// angle in radians
func computeFacing(angle float64) units.Facing {
    // right
    if betweenAngle(angle, 0, math.Pi/8){
        return units.FacingRight
    }

    // left
    if betweenAngle(angle, math.Pi, math.Pi/8){
        return units.FacingLeft
    }

    // up
    if betweenAngle(angle, math.Pi/2, math.Pi/8){
        return units.FacingUp
    }

    // up-left
    if betweenAngle(angle, math.Pi - math.Pi/4, math.Pi/8){
        return units.FacingUpLeft
    }

    // up-right
    if betweenAngle(angle, math.Pi/4, math.Pi/8){
        return units.FacingUpRight
    }

    // down-left
    if betweenAngle(angle, math.Pi + math.Pi/4, math.Pi/8){
        return units.FacingDownLeft
    }

    // down
    if betweenAngle(angle, math.Pi + math.Pi/2, math.Pi/8){
        return units.FacingDown
    }

    // down-right
    if betweenAngle(angle, math.Pi + math.Pi/2 + math.Pi/4, math.Pi/8){
        return units.FacingDownRight
    }

    // should be impossible to get here
    return units.FacingRight
}

/*
func (combat *CombatScreen) ContainsOppositeArmy(x int, y int, team Team) bool {
    unit := combat.Model.GetUnit(x, y)
    if unit == nil {
        return false
    }
    return unit.Team != team
}
*/

func faceTowards(x1 int, y1 int, x2 int, y2 int) units.Facing {
    angle := math.Atan2(float64(y2 - y1), float64(x2 - x1))

    // rotate by 45 degrees to get the on screen facing angle
    // have to negate the angle because the y axis is flipped (higher y values are lower on the screen)
    useAngle := -(angle - math.Pi/4)

    // log.Printf("Angle: %v from (%v,%v) to (%v,%v)", useAngle, combat.SelectedUnit.X, combat.SelectedUnit.Y, combat.SelectedUnit.TargetX, combat.SelectedUnit.TargetY)

    return computeFacing(useAngle)
}

func distanceInRange(x1 float64, y1 float64, x2 float64, y2 float64, r float64) bool {
    xDiff := x2 - x1
    yDiff := y2 - y1
    return xDiff * xDiff + yDiff * yDiff <= r*r
}

func distanceAboveRange(x1 float64, y1 float64, x2 float64, y2 float64, r float64) bool {
    xDiff := x2 - x1
    yDiff := y2 - y1
    return xDiff * xDiff + yDiff * yDiff >= r*r
}

func (combat *CombatScreen) doProjectiles(yield coroutine.YieldFunc) {
    for combat.Model.UpdateProjectiles(combat.Counter) {
        combat.Counter += 1
        combat.ProcessInput()
        combat.UpdateDamageIndicators()
        combat.UpdateAnimations()
        if yield() != nil {
            return
        }
    }
}

func (combat *CombatScreen) createUnitToUnitProjectile(attacker *ArmyUnit, target *ArmyUnit, offset image.Point, images []*ebiten.Image, explodeImages []*ebiten.Image, effect ProjectileEffect) *Projectile {
    matrix := combat.GetCameraMatrix()
    // find where on the screen the unit is
    screenX, screenY := matrix.Apply(float64(attacker.X), float64(attacker.Y))
    targetX, targetY := matrix.Apply(float64(target.X), float64(target.Y))

    var useImage *ebiten.Image
    if len(images) > 0 {
        useImage = images[0]
    } else if len(explodeImages) > 0 {
        useImage = explodeImages[0]
    }

    // FIXME: these coordinates should be incorporated into a geom

    var screenGeom ebiten.GeoM
    screenGeom.Translate(screenX, screenY)
    screenGeom.Translate(14, 3)
    screenGeom.Translate(-float64(useImage.Bounds().Dy()/2), -float64(useImage.Bounds().Dy()/2))
    screenGeom.Translate(float64(offset.X), float64(offset.Y))
    // screenGeom.Scale(combat.CameraScale, combat.CameraScale)

    screenX, screenY = screenGeom.Apply(0, 0)

    /*
    screenY += 3 * combat.CameraScale
    screenY -= float64(useImage.Bounds().Dy()/2) * combat.CameraScale
    screenX += 14 * combat.CameraScale
    screenX -= float64(useImage.Bounds().Dx()/2) * combat.CameraScale

    screenY += float64(offset.Y) * combat.CameraScale
    screenX += float64(offset.X) * combat.CameraScale
    */

    targetY += 3 * combat.CameraScale
    targetY -= float64(useImage.Bounds().Dy()/2) * combat.CameraScale
    targetX += 14 * combat.CameraScale
    targetX -= float64(useImage.Bounds().Dx()/2) * combat.CameraScale

    targetY += (rand.Float64() * 6 - 3) * combat.CameraScale
    targetX += (rand.Float64() * 6 - 3) * combat.CameraScale

    /*
    switch position {
        case UnitPositionMiddle:)
            screenY += 3
            screenY -= float64(useImage.Bounds().Dy()/2)
            screenX += 14
            screenX -= float64(useImage.Bounds().Dx()/2)
        case UnitPositionUnder:
            screenY += 15
            screenY -= float64(useImage.Bounds().Dy())
    }
    */

    speed := 2.8 

    angle := math.Atan2(targetY - screenY, targetX - screenX)

    // log.Printf("Create fireball projectile at %v,%v -> %v,%v", x, y, screenX, screenY)

    projectile := &Projectile{
        X: screenX,
        Y: screenY,
        Target: target,
        Speed: speed,
        Angle: angle,
        Effect: effect,
        TargetX: targetX,
        TargetY: targetY,
        Animation: util.MakeAnimation(images, true),
        Explode: util.MakeAnimation(explodeImages, false),
    }

    return projectile
}

func (combat *CombatScreen) createRangeAttack(attacker *ArmyUnit, defender *ArmyUnit){
    index := attacker.Unit.GetCombatRangeIndex(attacker.Facing)
    images, err := combat.ImageCache.GetImages("cmbmagic.lbx", index)
    if err != nil {
        log.Printf("Unable to load attacker range images for %v index %v: %v", attacker.Unit.GetName(), index, err)
        return
    }

    if len(images) != 4 {
        log.Printf("Invalid number of attack range animation images for %v: %v", attacker.Unit.GetName(), len(images))
        return
    }

    animation := images[0:3]
    explode := images[3:]

    tileDistance := computeTileDistance(attacker.X, attacker.Y, defender.X, defender.Y)

    effect := func (target *ArmyUnit){
        if target.GetHealth() <= 0 {
            return
        }

        damage := attacker.ComputeRangeDamage(target, tileDistance)

        // FIXME: for magical damage, set the Magic damage modifier for the proper realm
        appliedDamage, _ := ApplyDamage(target, []int{damage}, attacker.GetRangedAttackDamageType(), attacker.GetDamageSource(), DamageModifiers{WallDefense: combat.Model.ComputeWallDefense(attacker, defender)})

        totalDamage := appliedDamage

        log.Printf("attacker %v rolled %v ranged damage to defender %v, applied %v", attacker.Unit.GetName(), damage, target.Unit.GetName(), appliedDamage)

        if attacker.Unit.CanTouchAttack(attacker.Unit.GetRangedAttackDamageType()) {
            funcs := combat.Model.doTouchAttack(attacker, target, 0)
            for _, f := range funcs {
                totalDamage += f()
            }
        }

        totalDamage += combat.Model.ApplyImmolationDamage(defender, combat.Model.immolationDamage(attacker, defender))

        combat.AddDamageIndicator(target, totalDamage)

        // log.Printf("Ranged attack from %v: damage=%v defense=%v distance=%v", attacker.Unit.Name, damage, defense, tileDistance)

        /*
        damage -= defense
        if damage < 0 {
            damage = 0
        }
        target.TakeDamage(damage)
        */
        if target.GetHealth() <= 0 {
            combat.Model.KillUnit(target)
        }
    }

    for _, offset := range unitview.CombatPoints(attacker.Figures()) {
        combat.Model.Projectiles = append(combat.Model.Projectiles, combat.createUnitToUnitProjectile(attacker, defender, offset, animation, explode, effect))
    }
}

func (combat *CombatScreen) doSelectTile(yield coroutine.YieldFunc, selecter Team, spell spellbook.Spell, canTarget func(int, int) bool, selectTile func(int, int)) {
    combat.ButtonsDisabled = true
    combat.DoSelectTile = true
    defer func(){
        combat.ButtonsDisabled = false
        combat.DoSelectTile = false
    }()

    hudImage, _ := combat.ImageCache.GetImage("cmbtfx.lbx", 28, 0)

    hudY := data.ScreenHeight - hudImage.Bounds().Dy()

    quit := false

    x := 250
    if selecter == TeamDefender {
        x = 3
    }

    y := 168

    var elements []*uilib.UIElement

    removeElements := func(){
        combat.UI.RemoveElements(elements)
    }

    defer removeElements()

    selectElement := &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            combat.Fonts.WhiteFont.PrintWrap(screen, float64(x), float64(y), float64(75), font.FontOptions{Scale: scale.ScaleAmount}, fmt.Sprintf("Select a target for a %v spell.", spell.Name))
        },
    }

    cancelImages, _ := combat.ImageCache.GetImages("compix.lbx", 22)
    cancelRect := image.Rect(0, 0, cancelImages[0].Bounds().Dx(), cancelImages[0].Bounds().Dy()).Add(image.Point{(x + 15), (y + 15)})
    cancelIndex := 0
    cancelElement := &uilib.UIElement{
        Rect: cancelRect,
        LeftClick: func(element *uilib.UIElement){
            cancelIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            cancelIndex = 0
            quit = true
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            scale.DrawScaled(screen, cancelImages[cancelIndex], &options)
        },
    }

    elements = append(elements, selectElement, cancelElement)

    combat.UI.AddElements(elements)

    for !quit {
        combat.Counter += 1

        for _, unit := range combat.Model.OtherUnits {
            if combat.Counter % 6 == 0 {
                unit.Animation.Next()
            }
        }

        combat.UI.StandardUpdate()
        combat.ProcessInput()
        mouseX, mouseY := inputmanager.MousePosition()
        tileX, tileY := combat.ScreenToTile(float64(mouseX), float64(mouseY))
        combat.MouseTileX = int(math.Round(tileX))
        combat.MouseTileY = int(math.Round(tileY))

        if mouseY >= scale.Scale(hudY) {
            combat.MouseState = CombatClickHud
        } else if canTarget(combat.MouseTileX, combat.MouseTileY) {
            combat.MouseState = CombatCast

            if inputmanager.LeftClick() && mouseY < scale.Scale(hudY) {
                selectTile(combat.MouseTileX, combat.MouseTileY)
                yield()
                break
            }
        } else {
            combat.MouseState = CombatNotOk
        }

        combat.UpdateMouseState()

        if yield() != nil {
            return
        }
    }
}

func (combat *CombatScreen) PlaySound(spell spellbook.Spell) {
    sound, err := combat.AudioCache.GetSound(spell.Sound)
    if err == nil {
        sound.Play()
    } else {
        log.Printf("No such sound %v for %v: %v", spell.Sound, spell.Name, err)
    }
}

func (combat *CombatScreen) doSelectUnit(yield coroutine.YieldFunc, selecter Team, spell spellbook.Spell, selectTarget func (*ArmyUnit), canTarget func (*ArmyUnit) bool, selectTeam Team) {
    combat.ButtonsDisabled = true
    combat.DoSelectUnit = true
    defer func(){
        combat.ButtonsDisabled = false
        combat.DoSelectUnit = false
    }()

    hudImage, _ := combat.ImageCache.GetImage("cmbtfx.lbx", 28, 0)

    hudY := data.ScreenHeight - hudImage.Bounds().Dy()

    x := 250
    if selecter == TeamDefender {
        x = 3
    }

    y := 168

    var elements []*uilib.UIElement

    removeElements := func(){
        combat.UI.RemoveElements(elements)
    }

    defer removeElements()

    selectElement := &uilib.UIElement{
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            combat.Fonts.WhiteFont.PrintWrap(screen, float64(x), float64(y), float64(75), font.FontOptions{Scale: scale.ScaleAmount}, fmt.Sprintf("Select a target for a %v spell.", spell.Name))
        },
    }

    quit := false

    cancelImages, _ := combat.ImageCache.GetImages("compix.lbx", 22)
    cancelRect := image.Rect(0, 0, cancelImages[0].Bounds().Dx(), cancelImages[0].Bounds().Dy()).Add(image.Point{(x + 15), (y + 15)})
    cancelIndex := 0
    cancelElement := &uilib.UIElement{
        Rect: cancelRect,
        LeftClick: func(element *uilib.UIElement){
            cancelIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            cancelIndex = 0
            quit = true
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            scale.DrawScaled(screen, cancelImages[cancelIndex], &options)
        },
    }

    elements = append(elements, selectElement, cancelElement)

    combat.UI.AddElements(elements)

    canTargetMemo := functional.Memoize(func (target *ArmyUnit) bool {
        return canTarget(target) && combat.IsUnitVisible(target)
    })

    for !quit {
        combat.Counter += 1

        for _, unit := range combat.Model.OtherUnits {
            if combat.Counter % 6 == 0 {
                unit.Animation.Next()
            }
        }

        combat.UI.StandardUpdate()
        combat.ProcessInput()
        mouseX, mouseY := inputmanager.MousePosition()
        tileX, tileY := combat.ScreenToTile(float64(mouseX), float64(mouseY))
        combat.MouseTileX = int(math.Round(tileX))
        combat.MouseTileY = int(math.Round(tileY))

        combat.MouseState = CombatCast

        combat.Model.HighlightedUnit = combat.Model.GetUnit(combat.MouseTileX, combat.MouseTileY)

        if mouseY >= scale.Scale(hudY) {
            combat.MouseState = CombatClickHud
        } else {
            unit := combat.Model.GetUnit(combat.MouseTileX, combat.MouseTileY)
            if unit == nil || (selectTeam != TeamEither && unit.Team != selectTeam) || !canTargetMemo(unit){
                combat.MouseState = CombatNotOk
            }

            if unit != nil && canTargetMemo(unit) && inputmanager.LeftClick() && mouseY < scale.Scale(hudY) {
                // log.Printf("Click unit at %v,%v -> %v", combat.MouseTileX, combat.MouseTileY, unit)
                if selectTeam == TeamEither || unit.Team == selectTeam {
                    selectTarget(unit)

                    // shouldn't need to set the mouse state here
                    combat.MouseState = CombatClickHud

                    // asborb click
                    yield()
                    return
                }
            }
        }

        combat.UpdateMouseState()

        if yield() != nil {
            return
        }
    }
}

func (combat *CombatScreen) doCastEnchantment(yield coroutine.YieldFunc, caster ArmyPlayer, magic data.MagicType, spellName string) {
    oldDrawer := combat.Drawer
    defer func(){
        combat.Drawer = oldDrawer
    }()

    value := data.GetMagicColor(magic)

    counter := 0
    counterMax := 90

    maxAlpha := 150

    castDescription := fmt.Sprintf("%v cast %v", caster.GetWizard().Name, spellName)

    text := combat.Fonts.EnchantmentFont.MeasureTextWidth(castDescription, 1)

    interpolate := func (counter int) uint8 {
        if counter < counterMax / 2 {
            return uint8(counter * maxAlpha / (counterMax / 2))
        } else {
            return uint8((counterMax - counter) * maxAlpha / (counterMax / 2))
        }
    }

    combat.Drawer = func (screen *ebiten.Image){
        oldDrawer(screen)

        x1 := float64(data.ScreenWidth / 2) - text / 2 - float64(1)
        x2 := float64(data.ScreenWidth / 2) + text / 2 + float64(1)
        y := 4

        vector.FillRect(screen, float32(scale.Scale(x1)), float32(scale.Scale(y)), float32(scale.Scale(x2 - x1)), float32(scale.Scale(combat.Fonts.EnchantmentFont.Height() + 1)), color.RGBA{R: 0, G: 0, B: 0x0, A: 120}, false)
        combat.Fonts.EnchantmentFont.PrintOptions(screen, float64(data.ScreenWidth / 2), float64((y + 1)), font.FontOptions{Scale: scale.ScaleAmount, Justify: font.FontJustifyCenter}, castDescription)

        vector.StrokeRect(screen, float32(scale.Scale(x1)), float32(scale.Scale(y)), float32(scale.Scale(x2 - x1)), float32(scale.Scale(combat.Fonts.EnchantmentFont.Height() + 1)), float32(scale.Scale(1)), color.RGBA{R: 0xff, G: 0xff, B: 0x0, A: 0xff}, false)

        vector.FillRect(screen, 0, 0, float32(screen.Bounds().Dx()), float32(screen.Bounds().Dy()), util.PremultiplyAlpha(value), false)
    }

    for counter < counterMax {
        combat.Counter += 1
        combat.ProcessInput()
        counter += 1
        value.A = interpolate(counter)
        yield()
    }
}

func (combat *CombatScreen) ShowSummon(yield coroutine.YieldFunc, unit *ArmyUnit) {
    for unit.Height < 0 {
        // so that the summoning circle displays
        combat.Model.UpdateProjectiles(combat.Counter)
        combat.Counter += 1

        if combat.Counter % 3 == 0 {
            unit.SetHeight(unit.Height + 1)
        }

        if yield() != nil {
            return
        }
    }
}

func (combat *CombatScreen) ProcessEvents(yield coroutine.YieldFunc) CombatUpdates {

    var updates CombatUpdates

    sounds := set.MakeSet[int]()
    defer func(){
        for _, index := range sounds.Values() {
            sound, err := combat.AudioCache.GetSound(index)
            if err == nil {
                sound.Play()
            } else {
                log.Printf("Unable to play sound %v: %v", index, err)
            }
        }
    }()

    for {
        select {
            case event := <-combat.Events:
                switch event.(type) {
                    case *CombatEventSelectTile:
                        use := event.(*CombatEventSelectTile)
                        combat.doSelectTile(yield, use.Selecter, use.Spell, use.CanTarget, use.SelectTile)
                    case *CombatEventSelectUnit:
                        use := event.(*CombatEventSelectUnit)
                        combat.doSelectUnit(yield, use.Selecter, use.Spell, use.SelectTarget, use.CanTarget, use.SelectTeam)
                    case *CombatEventNextUnit:
                        combat.Model.NextUnit()
                    case *CombatEventGlobalSpell:
                        use := event.(*CombatEventGlobalSpell)
                        combat.doCastEnchantment(yield, use.Caster, use.Magic, use.Name)
                    case *CombatSelectTargets:
                        use := event.(*CombatSelectTargets)
                        if use.Army.Auto {
                            if len(use.Targets) > 0 {
                                pick := use.Targets[rand.N(len(use.Targets))]
                                use.Select(pick)
                            }
                        } else {
                            combat.AddSelectTargetsElements(use.Targets, use.Title, use.Select)
                        }
                    case *CombatEventMessage:
                        use := event.(*CombatEventMessage)
                        combat.UI.AddElement(uilib.MakeErrorElement(combat.UI, combat.Cache, &combat.ImageCache, use.Message, func(){ yield() }))
                    case *CombatEventCreateLightningBolt:
                        bolt := event.(*CombatEventCreateLightningBolt)
                        combat.Model.AddProjectile(combat.CreateLightningBoltProjectile(bolt.Target, bolt.Strength))
                        sounds.Insert(LightningBoltSound)
                    case *CombatEventSummonUnit:
                        summon := event.(*CombatEventSummonUnit)
                        combat.ShowSummon(yield, summon.Unit)

                    case *CombatCreateWallOfFire:
                        createWallOfFire(combat.Model.Tiles, TownCenterX, TownCenterY, 4, combat.Counter)

                    case *CombatCreateWallOfDarkness:
                        createWallOfDarkness(combat.Model.Tiles, TownCenterX, TownCenterY, 4, combat.Counter)

                    case *CombatPlaySound:
                        use := event.(*CombatPlaySound)
                        sounds.Insert(use.Sound)

                    case *CombatDoSingleAuto:
                        updates.SingleAuto = true
                }
            default:
                return updates
        }
    }
}

func (combat *CombatScreen) UpdateAnimations(){
    for _, unit := range combat.Model.OtherUnits {
        if combat.Counter % 6 == 0 {
            unit.Animation.Next()
        }
    }

    updateLost := func (units []*ArmyUnit) {
        for _, unit := range units {
            if unit.LostUnitsTime > 0 {
                unit.LostUnitsTime -= 1
            } else {
                unit.LostUnits = 0
            }
        }
    }

    updateLost(combat.Model.AttackingArmy.units)
    updateLost(combat.Model.DefendingArmy.units)
    updateLost(combat.Model.AttackingArmy.KilledUnits)
    updateLost(combat.Model.DefendingArmy.KilledUnits)
}

func (combat *CombatScreen) doTeleport(yield coroutine.YieldFunc, mover *ArmyUnit, x int, y int, merge bool) {

    sound, err := combat.AudioCache.GetSound(mover.Unit.GetMovementSound().LbxIndex())
    if err == nil && combat.IsUnitVisible(mover) {
        sound.Play()
    }

    mergeCount := 30
    mergeSpeed := 2

    if merge {
        for i := range mergeCount {
            combat.Counter += 1
            combat.UpdateAnimations()
            combat.UpdateDamageIndicators()
            combat.ProcessInput()
            mover.SetHeight(-i/mergeSpeed)
            yield()
        }
    } else {
        for i := range mergeCount {
            combat.Counter += 1
            combat.UpdateAnimations()
            combat.UpdateDamageIndicators()
            combat.ProcessInput()
            mover.SetFade(float32(i)/float32(mergeCount))
            yield()
        }
    }

    combat.Model.Tiles[mover.Y][mover.X].Unit = nil
    mover.X = x
    mover.Y = y
    mover.MovesLeft = mover.MovesLeft.Subtract(fraction.FromInt(1))
    combat.Model.Tiles[mover.Y][mover.X].Unit = mover

    if merge {
        for i := range mergeCount {
            combat.Counter += 1
            combat.UpdateAnimations()
            combat.UpdateDamageIndicators()
            combat.ProcessInput()
            mover.SetHeight(-(mergeCount/mergeSpeed - i/mergeSpeed))
            yield()
        }
        mover.SetHeight(0)
    } else {
        for i := range mergeCount {
            combat.Counter += 1
            combat.UpdateAnimations()
            combat.UpdateDamageIndicators()
            combat.ProcessInput()
            mover.SetFade(float32(mergeCount - i)/float32(mergeCount))
            yield()
        }
        mover.SetFade(0)
    }
}

func (combat *CombatScreen) doMoveUnit(yield coroutine.YieldFunc, mover *ArmyUnit, path pathfinding.Path){
    if len(path) == 0 {
        return
    }

    mover.MovementTick = combat.Counter
    mover.Moving = true
    mover.MoveX = float64(mover.X)
    mover.MoveY = float64(mover.Y)

    quit, cancel := context.WithCancel(combat.Quit)
    defer cancel()

    sound, err := combat.AudioCache.GetSound(mover.Unit.GetMovementSound().LbxIndex())
    if err == nil && combat.IsUnitVisible(mover) {
        // keep playing movement sound in a loop until the unit stops moving
        go func(){
            // defer sound.Pause()
            for quit.Err() == nil {
                err = sound.Rewind()
                if err != nil {
                    log.Printf("Unable to rewind sound for %v: %v", mover.Unit.GetMovementSound(), err)
                }
                sound.Play()
                for sound.IsPlaying() {
                    select {
                    case <-quit.Done():
                        sound.Pause()
                        return
                    case <-time.After(10 * time.Millisecond):
                    }
                }
            }
        }()
    }

    // FIXME: move some of this code into model.go
    for len(path) > 0 && mover.MovesLeft.GreaterThan(fraction.FromInt(0)) {
        mover.CurrentPath = path
        targetX, targetY := path[0].X, path[0].Y

        combat.Model.AddLogEvent(fmt.Sprintf("Moving %v %v,%v -> %v,%v", mover.Unit.GetName(), mover.X, mover.Y, targetX, targetY))

        angle := math.Atan2(float64(targetY) - mover.MoveY, float64(targetX) - mover.MoveX)

        // rotate by 45 degrees to get the on screen facing angle
        // have to negate the angle because the y axis is flipped (higher y values are lower on the screen)
        useAngle := -(angle - math.Pi/4)

        // log.Printf("Angle: %v from (%v,%v) to (%v,%v)", useAngle, combat.SelectedUnit.X, combat.SelectedUnit.Y, combat.SelectedUnit.TargetX, combat.SelectedUnit.TargetY)

        mover.Facing = computeFacing(useAngle)

        // speed := float64(combat.Counter - combat.SelectedUnit.MovementTick) / 4
        speed := float64(0.04)

        reached := false
        for !reached && mover.MovesLeft.GreaterThan(fraction.FromInt(0)) {
            combat.UpdateAnimations()
            combat.UpdateDamageIndicators()
            combat.ProcessInput()
            combat.Counter += 1

            mouseX, mouseY := inputmanager.MousePosition()
            tileX, tileY := combat.ScreenToTile(float64(mouseX), float64(mouseY))
            combat.MouseTileX = int(math.Round(tileX))
            combat.MouseTileY = int(math.Round(tileY))

            mover.MoveX += math.Cos(angle) * speed
            mover.MoveY += math.Sin(angle) * speed

            // log.Printf("Moving %v,%v -> %v,%v", combat.SelectedUnit.X, combat.SelectedUnit.Y, combat.SelectedUnit.MoveX, combat.SelectedUnit.MoveY)

            // if math.Abs(combat.SelectedUnit.MoveX - float64(targetX)) < speed*2 && math.Abs(combat.SelectedUnit.MoveY - float64(targetY)) < 0.5 {
            if distanceInRange(mover.MoveX, mover.MoveY, float64(targetX), float64(targetY), speed * 3) ||
            // a stop gap to ensure the unit doesn't fly off the screen somehow
            distanceAboveRange(float64(mover.X), float64(mover.Y), float64(targetX), float64(targetY), 2.5) {

                // tile where the unit came from is now empty
                combat.Model.Tiles[mover.Y][mover.X].Unit = nil

                mover.MovesLeft = mover.MovesLeft.Subtract(pathCost(image.Pt(mover.X, mover.Y), image.Pt(targetX, targetY)))
                if mover.MovesLeft.LessThan(fraction.FromInt(0)) {
                    mover.MovesLeft = fraction.FromInt(0)
                }

                // unit moves from outside the wall of fire to inside
                if !mover.IsFlying() && combat.Model.InsideWallOfFire(targetX, targetY) && !combat.Model.InsideWallOfFire(mover.X, mover.Y) {
                    combat.Model.ApplyWallOfFireDamage(mover)

                    if mover.GetHealth() <= 0 {
                        // this feels dangerous to do here but it seems to work
                        combat.Model.KillUnit(mover)
                        return
                    }
                }

                mover.X = targetX
                mover.Y = targetY
                mover.MoveX = float64(targetX)
                mover.MoveY = float64(targetY)

                // new tile the unit landed on is now occupied
                combat.Model.Tiles[mover.Y][mover.X].Unit = mover
                path = path[1:]
                reached = true
            }

            if yield() != nil {
                return
            }
        }
    }

    mover.Moving = false
    mover.CurrentPath = nil
    mover.Paths = make(map[image.Point]pathfinding.Path)
}

func (combat *CombatScreen) doRangeAttack(yield coroutine.YieldFunc, attacker *ArmyUnit, defender *ArmyUnit){
    attacker.MovesLeft = attacker.MovesLeft.Subtract(fraction.FromInt(10))
    if attacker.MovesLeft.LessThan(fraction.FromInt(0)) {
        attacker.MovesLeft = fraction.FromInt(0)
    }

    attacker.Facing = faceTowards(attacker.X, attacker.Y, defender.X, defender.Y)

    attacks := 1
    // haste does two ranged attacks
    if attacker.HasEnchantment(data.UnitEnchantmentHaste) {
        // caster's don't get to attack twice
        if attacker.GetRangedAttackDamageType() == units.DamageRangedMagical {
        } else {
            attacks = min(2, attacker.RangedAttacks)
        }
    }

    for range attacks {
        attacker.UseRangeAttack()
        // FIXME: could use a for/yield loop here to update projectiles
        combat.createRangeAttack(attacker, defender)
    }

    sound, err := combat.AudioCache.GetSound(attacker.Unit.GetRangeAttackSound().LbxIndex())
    if err == nil {
        sound.Play()
    }

    combat.doProjectiles(yield)
}

func (combat *CombatScreen) doMelee(yield coroutine.YieldFunc, attacker *ArmyUnit, defender *ArmyUnit){
    attacker.Attacking = true
    defender.Defending = true
    defer func(){
        attacker.Attacking = false
        defender.Defending = false
    }()

    // attacking takes 50% of movement points
    // FIXME: in some cases an extra 0.5 movements points is lost, possibly due to counter attacks?

    pointsUsed := attacker.GetMovementSpeed().Divide(fraction.FromInt(2))
    if pointsUsed.LessThan(fraction.FromInt(1)) {
        pointsUsed = fraction.FromInt(1)
    }

    attacker.MovesLeft = attacker.MovesLeft.Subtract(pointsUsed)
    if attacker.MovesLeft.LessThan(fraction.FromInt(0)) {
        attacker.MovesLeft = fraction.FromInt(0)
    }

    attacker.Facing = faceTowards(attacker.X, attacker.Y, defender.X, defender.Y)
    defender.Facing = faceTowards(defender.X, defender.Y, attacker.X, attacker.Y)

    // FIXME: sound is based on attacker type, and possibly defender type
    sound, err := combat.AudioCache.GetCombatSound(attacker.Unit.GetAttackSound().LbxIndex())
    if err == nil {
        sound.Play()
    }

    combat.Model.AddLogEvent(fmt.Sprintf("%v attacks %v", attacker.Unit.GetName(), defender.Unit.GetName()))

    for i := range 60 {
        combat.Counter += 1
        combat.UpdateAnimations()
        combat.UpdateDamageIndicators()
        combat.ProcessInput()
        combat.ProcessEvents(yield) // ignore return

        // delay the actual melee computation to give time for the sound to play
        if i == 20 {
            attackerDamage, defenderDamage := combat.Model.meleeAttack(attacker, defender)

            combat.AddDamageIndicator(defender, attackerDamage)
            combat.AddDamageIndicator(attacker, defenderDamage)
        }

        if yield() != nil {
            return
        }
    }
}

func (combat *CombatScreen) AddDamageIndicator(unit *ArmyUnit, damage int) {
    offsetWidth := 8
    indicator := DamageIndicator{
        X: unit.X,
        Y: unit.Y,
        Offset: rand.N(offsetWidth * 2 + 1) - offsetWidth,
        Damage: damage,
        Life: 50,
    }

    combat.DamageIndicators = append(combat.DamageIndicators, indicator)
}

type AIUnitActions struct {
    yield coroutine.YieldFunc
    combat *CombatScreen
}

func (actions AIUnitActions) Teleport(mover *ArmyUnit, x int, y int, merge bool) {
    actions.combat.doTeleport(actions.yield, mover, x, y, merge)
}

func (actions AIUnitActions) RangeAttack(attacker *ArmyUnit, defender *ArmyUnit) {
    actions.combat.doRangeAttack(actions.yield, attacker, defender)
}

func (actions AIUnitActions) MeleeAttack(attacker *ArmyUnit, defender *ArmyUnit) {
    actions.combat.doMelee(actions.yield, attacker, defender)
}

func (actions AIUnitActions) MoveUnit(unit *ArmyUnit, path pathfinding.Path) {
    actions.combat.doMoveUnit(actions.yield, unit, path)
}


func (combat *CombatScreen) UpdateMouseState() {
    switch combat.MouseState {
        case CombatMoveOk:
            globalMouse.Mouse.SetImage(combat.Mouse.Move)
        case CombatClickHud:
            globalMouse.Mouse.SetImage(combat.Mouse.Normal)
        case CombatMeleeAttackOk:
            // mouseOptions.GeoM.Translate(-1, -1)
            globalMouse.Mouse.SetImage(combat.Mouse.Attack)
        case CombatRangeAttackOk:
            globalMouse.Mouse.SetImage(combat.Mouse.Arrow)
        case CombatNotOk:
            globalMouse.Mouse.SetImage(combat.Mouse.Error)
        case CombatCast:
            index := (combat.Counter / 8) % uint64(len(combat.Mouse.Cast))
            globalMouse.Mouse.SetImage(combat.Mouse.Cast[index])
    }
}

func (combat *CombatScreen) ProcessInput() {
    combat.ExtraHighlightedUnit = nil
    var keys []ebiten.Key
    keys = inpututil.AppendPressedKeys(keys)
    showInfo := 0
    combat.ExtraControl = false
    for _, key := range keys {
        speed := 0.8
        switch key {
            case ebiten.KeyDown:
                combat.Coordinates.Translate(0, -speed)
            case ebiten.KeyUp:
                combat.Coordinates.Translate(0, speed)
            case ebiten.KeyLeft:
                combat.Coordinates.Translate(speed, 0)
            case ebiten.KeyRight:
                combat.Coordinates.Translate(-speed, 0)
            case ebiten.KeyEqual:
                if combat.CameraScale < 3 {
                    combat.CameraScale *= 1 + 0.01
                    combat.Coordinates.Scale(1.01, 1.01)
                }
            case ebiten.KeyMinus:
                if combat.CameraScale > 0.5 {
                    combat.CameraScale *= 1.0 - 0.01
                    combat.Coordinates.Scale(0.99, 0.99)
                }
            case ebiten.KeySpace:
                normalized := 1 / combat.CameraScale
                combat.CameraScale *= normalized
                combat.Coordinates.Scale(normalized, normalized)
            case ebiten.KeyTab:
                if combat.Model.SelectedUnit != nil && !combat.Model.IsAIControlled(combat.Model.SelectedUnit) {
                    combat.ExtraHighlightedUnit = combat.Model.SelectedUnit
                }
            case ebiten.KeyShift:
                showInfo = 100
            case ebiten.KeyControl:
                combat.ExtraControl = true
        }
    }

    infoStep := 13

    if showInfo > 0 {
        if combat.ShowInfoLevel < showInfo {
            combat.ShowInfoLevel += infoStep
            combat.ShowInfoLevel = min(combat.ShowInfoLevel, showInfo)
        }
    } else {
        combat.ShowInfoLevel = max(0, combat.ShowInfoLevel - infoStep)
    }

    // FIXME: handle right-click drag to move the camera

    _, wheelY := inputmanager.Wheel()

    wheelScale := 1 + float64(wheelY) / 10
    combat.CameraScale *= wheelScale
    combat.Coordinates.Scale(wheelScale, wheelScale)
}

func (combat *CombatScreen) UpdateDamageIndicators() {
    var keepIndicators []DamageIndicator

    for _, indicator := range combat.DamageIndicators {
        indicator.Life -= 1
        indicator.Count += 1
        if indicator.Life > 0 {
            keepIndicators = append(keepIndicators, indicator)
        }
    }

    combat.DamageIndicators = keepIndicators
}

func (combat *CombatScreen) Update(yield coroutine.YieldFunc) CombatState {
    if combat.Model.CurrentTurn >= MAX_TURNS {
        combat.Model.AddLogEvent("Combat exceeded maximum number of turns, defender wins")
        combat.Model.FinishCombat(CombatStateDefenderWin)
        return CombatStateDefenderWin
    }

    if combat.Model.AttackingArmy.Fled {
        combat.Model.flee(combat.Model.AttackingArmy)
        combat.Model.FinishCombat(CombatStateAttackerFlee)
        return CombatStateAttackerFlee
    }

    if combat.Model.DefendingArmy.Fled {
        combat.Model.flee(combat.Model.DefendingArmy)
        combat.Model.FinishCombat(CombatStateDefenderFlee)
        return CombatStateDefenderFlee
    }

    if len(combat.Model.AttackingArmy.units) == 0 {
        combat.Model.AddLogEvent("Defender wins!")
        combat.Model.FinishCombat(CombatStateDefenderWin)
        return CombatStateDefenderWin
    }

    if len(combat.Model.DefendingArmy.units) == 0 {
        combat.Model.AddLogEvent("Attacker wins!")
        combat.Model.FinishCombat(CombatStateAttackerWin)
        return CombatStateAttackerWin
    }

    combat.Counter += 1
    combat.UI.StandardUpdate()

    combat.UpdateMouseState()

    mouseX, mouseY := inputmanager.MousePosition()
    hudImage, _ := combat.ImageCache.GetImage("cmbtfx.lbx", 28, 0)

    tileX, tileY := combat.ScreenToTile(float64(mouseX), float64(mouseY))
    combat.MouseTileX = int(math.Round(tileX))
    combat.MouseTileY = int(math.Round(tileY))

    combat.UpdateDamageIndicators()
    combat.UpdateAnimations()

    // hudY := data.ScreenHeightOriginal - hudImage.Bounds().Dy()
    hudY := (data.ScreenHeight - hudImage.Bounds().Dy())

    combat.ProcessInput()

    updates := combat.ProcessEvents(yield)

    if len(combat.Model.Projectiles) > 0 {
        combat.doProjectiles(yield)
    }

    if combat.Model.SelectedUnit != nil && combat.Model.SelectedUnit.ConfusionAction == ConfusionActionMoveRandomly {
        confusedUnit := combat.Model.SelectedUnit

        // keep moving randomly until the unit is out of moves
        for confusedUnit.MovesLeft.GreaterThan(fraction.Zero()) {
            var points []image.Point
            for x := -1; x <= 1; x++ {
                for y := -1; y <= 1; y++ {
                    if x == 0 && y == 0 {
                        continue
                    }

                    points = append(points, image.Pt(confusedUnit.X + x, confusedUnit.Y + y))
                }
            }

            moved := false
            for _, index := range rand.Perm(len(points)) {
                point := points[index]
                if combat.TileIsEmpty(point.X, point.Y) && combat.Model.CanMoveTo(confusedUnit, point.X, point.Y, false) {
                    path, _ := combat.Model.FindPath(confusedUnit, point.X, point.Y, false)
                    path = path[1:]
                    combat.doMoveUnit(yield, confusedUnit, path)
                    moved = true
                    break
                }
            }

            // unable to move, just quit the loop
            if !moved {
                break
            }
        }

        combat.Model.DoneTurn()

        return CombatStateRunning
    }

    if combat.Model.SelectedUnit != nil && (combat.Model.IsAIControlled(combat.Model.SelectedUnit) || updates.SingleAuto) {
        aiUnit := combat.Model.SelectedUnit

        aiArmy := combat.Model.GetArmy(aiUnit)

        // don't let a single auto unit cast wizard spells
        if combat.Model.IsAIControlled(aiUnit) {
            casted := combat.Model.doAiCast(combat, aiArmy)
            if casted {
                combat.doProjectiles(yield)
            }
        }

        // keep making choices until the unit runs out of moves
        for aiUnit.MovesLeft.GreaterThan(fraction.FromInt(0)) && aiUnit.GetHealth() > 0 {
            doAI(combat.Model, combat, &AIUnitActions{yield: yield, combat: combat}, aiUnit)
        }

        aiUnit.LastTurn = combat.Model.CurrentTurn
        combat.Model.NextUnit()
        return CombatStateRunning
    }

    if combat.UI.GetHighestLayerValue() > 0 || mouseY >= scale.Scale(hudY) {
        combat.MouseState = CombatClickHud
    } else if combat.Model.SelectedUnit != nil && combat.Model.SelectedUnit.Moving {
        combat.MouseState = CombatClickHud
    } else if combat.Model.SelectedUnit != nil {
        who := combat.Model.GetUnit(combat.MouseTileX, combat.MouseTileY)
        if who == nil {
            if combat.Model.CanMoveTo(combat.Model.SelectedUnit, combat.MouseTileX, combat.MouseTileY, combat.ExtraControl) {
                combat.MouseState = CombatMoveOk
            } else {
                combat.MouseState = CombatNotOk
            }
        } else {
            newState := CombatNotOk
            // prioritize range attack over melee
            if combat.Model.canRangeAttack(combat.Model.SelectedUnit, who) && combat.Model.withinArrowRange(combat.Model.SelectedUnit, who) {
                newState = CombatRangeAttackOk
            } else if combat.Model.canMeleeAttack(combat.Model.SelectedUnit, who, true) && combat.Model.withinMeleeRange(combat.Model.SelectedUnit, who) {
                newState = CombatMeleeAttackOk
            }

            combat.MouseState = newState
        }
    }

    // if there is no unit at the tile position then the highlighted unit will be nil
    if combat.UI.GetHighestLayerValue() == 0 {
        combat.Model.HighlightedUnit = combat.Model.GetUnit(combat.MouseTileX, combat.MouseTileY)
    }

    // dont allow clicks into the hud area
    // also don't allow clicks into the game if the ui is showing some overlay
    if combat.UI.GetHighestLayerValue() == 0 &&
       inputmanager.LeftClick() &&
       mouseY < scale.Scale(hudY) {

        if combat.TileIsEmpty(combat.MouseTileX, combat.MouseTileY) && combat.Model.CanMoveTo(combat.Model.SelectedUnit, combat.MouseTileX, combat.MouseTileY, combat.ExtraControl) {
            if combat.Model.SelectedUnit.CanTeleport() {
                combat.doTeleport(yield, combat.Model.SelectedUnit, combat.MouseTileX, combat.MouseTileY, combat.Model.SelectedUnit.HasAbility(data.AbilityMerging))
            } else {
                path, _ := combat.Model.FindPath(combat.Model.SelectedUnit, combat.MouseTileX, combat.MouseTileY, combat.ExtraControl)
                path = path[1:]
                combat.doMoveUnit(yield, combat.Model.SelectedUnit, path)
            }
        } else {

           defender := combat.Model.GetUnit(combat.MouseTileX, combat.MouseTileY)
           attacker := combat.Model.SelectedUnit

           if defender != nil {
               // try a ranged attack first
               if combat.Model.withinArrowRange(attacker, defender) && combat.Model.canRangeAttack(attacker, defender) {
                   combat.doRangeAttack(yield, attacker, defender)
               // then fall back to melee
               } else if combat.Model.withinMeleeRange(attacker, defender) && combat.Model.canMeleeAttack(attacker, defender, true){
                   combat.doMelee(yield, attacker, defender)
                   attacker.Paths = make(map[image.Point]pathfinding.Path)
               }
           }
       }
    }

    if combat.UI.GetHighestLayerValue() == 0 &&
       inputmanager.RightClick() &&
       mouseY < scale.Scale(hudY) {

       showUnit := combat.Model.GetUnit(combat.MouseTileX, combat.MouseTileY)
       if showUnit != nil {
           combat.UI.AddGroup(MakeUnitView(combat.Cache, combat.UI, showUnit))
       }
   }

    // the unit died or is out of moves
    if combat.Model.SelectedUnit != nil && (combat.Model.SelectedUnit.GetHealth() <= 0 || combat.Model.SelectedUnit.MovesLeft.LessThanEqual(fraction.FromInt(0))) {
        combat.Model.DoneTurn()
    }

    // log.Printf("Mouse original %v,%v %v,%v -> %v,%v", mouseX, mouseY, tileX, tileY, combat.MouseTileX, combat.MouseTileY)

    return CombatStateRunning
}

func (combat *CombatScreen) DrawHighlightedTile(screen *ebiten.Image, x int, y int, matrix *ebiten.GeoM, minColor color.NRGBA, maxColor color.NRGBA){
    tile0, _ := combat.ImageCache.GetImage("cmbgrass.lbx", 0, 0)

    var useMatrix ebiten.GeoM

    /*
    tx, ty := matrix.Apply(float64(x), float64(y))
    useMatrix.Scale(combat.CameraScale, combat.CameraScale)
    useMatrix.Translate(tx, ty)
    useMatrix = applyGeomScale(useMatrix)
    */
    // useMatrix = applyGeomScale(ebiten.GeoM{})
    // useMatrix.Concat(*matrix)

    tx, ty := matrix.Apply(float64(x), float64(y))
    useMatrix.Scale(combat.CameraScale, combat.CameraScale)
    useMatrix.Translate(tx, ty)
    useMatrix = scale.ScaleGeom(useMatrix)

    // log.Printf("tx=%v, ty=%v", tx, ty)

    // left
    x1, y1 := useMatrix.Apply(float64(-tile0.Bounds().Dx()/2), 0)
    // top
    x2, y2 := useMatrix.Apply(0, float64(-tile0.Bounds().Dy()/2))
    // right
    x3, y3 := useMatrix.Apply(float64(tile0.Bounds().Dx()/2), 0)
    // bottom
    x4, y4 := useMatrix.Apply(0, float64(tile0.Bounds().Dy()/2))

    gradient := (math.Sin(float64(combat.Counter)/6) + 1)

    lerp := func(minC uint8, maxC uint8) uint8 {
        out := float64(minC) + gradient * float64(maxC - minC)/2
        if out > 255 {
            out = 255
        }
        if out < 0 {
            out = 0
        }

        return uint8(out)
    }

    lineColor := color.NRGBA{
        R: lerp(minColor.R, maxColor.R),
        G: lerp(minColor.G, maxColor.G),
        B: lerp(minColor.B, maxColor.B),
        A: 190}

    var path vector.Path
    path.MoveTo(float32(x1), float32(y1))
    path.LineTo(float32(x2), float32(y2))
    path.LineTo(float32(x3), float32(y3))
    path.LineTo(float32(x4), float32(y4))
    path.Close()

    var fill vector.FillOptions
    var pathOptions vector.DrawPathOptions
    pathOptions.ColorScale.ScaleWithColor(lineColor)

    vector.FillPath(screen, &path, &fill, &pathOptions)
}

func (combat *CombatScreen) ShowUnitInfo(screen *ebiten.Image, unit *ArmyUnit){
    x1 := 255 - 1
    y1 := 5
    width := 65
    height := 45
    vector.FillRect(screen, float32(scale.Scale(x1)), float32(scale.Scale(y1)), float32(scale.Scale(width)), float32(scale.Scale(height)), color.NRGBA{R: 0, G: 0, B: 0, A: 100}, false)
    vector.StrokeRect(screen, float32(scale.Scale(x1)), float32(scale.Scale(y1)), float32(scale.Scale(width)), float32(scale.Scale(height)), float32(scale.Scale(1)), color.NRGBA{R: 0x27, G: 0x4e, B: 0xdc, A: 100}, false)
    combat.Fonts.InfoFont.PrintOptions(screen, float64(x1 + 35), float64(y1 + 2), font.FontOptions{Justify: font.FontJustifyCenter, DropShadow: true, Scale: scale.ScaleAmount}, fmt.Sprintf("%v", unit.Unit.GetName()))

    meleeImage, _ := combat.ImageCache.GetImage("compix.lbx", 61, 0)
    var options ebiten.DrawImageOptions
    options.GeoM.Translate(float64(x1 + 14), float64(y1 + 10))
    scale.DrawScaled(screen, meleeImage, &options)
    ax, ay := options.GeoM.Apply(0, 2)
    combat.Fonts.InfoFont.PrintOptions(screen, ax, ay, font.FontOptions{Justify: font.FontJustifyRight, DropShadow: true, Scale: scale.ScaleAmount}, fmt.Sprintf("%v", unit.GetMeleeAttackPower()))

    switch unit.Unit.GetRangedAttackDamageType() {
        case units.DamageRangedMagical:
            fire, _ := combat.ImageCache.GetImage("compix.lbx", 62, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(x1 + 14), float64(y1 + 18))
            scale.DrawScaled(screen, fire, &options)
            ax, ay := options.GeoM.Apply(0, 2)
            combat.Fonts.InfoFont.PrintOptions(screen, ax, ay, font.FontOptions{Justify: font.FontJustifyRight, DropShadow: true, Scale: scale.ScaleAmount}, fmt.Sprintf("%v", unit.GetRangedAttackPower()))
        case units.DamageRangedBoulder:
            boulder, _ := combat.ImageCache.GetImage("compix.lbx", 67, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(x1 + 14), float64(y1 + 18))
            scale.DrawScaled(screen, boulder, &options)
            ax, ay := options.GeoM.Apply(0, 2)
            combat.Fonts.InfoFont.PrintOptions(screen, ax, ay, font.FontOptions{Justify: font.FontJustifyRight, DropShadow: true, Scale: scale.ScaleAmount}, fmt.Sprintf("%v", unit.GetRangedAttackPower()))

        case units.DamageRangedPhysical:
            arrow, _ := combat.ImageCache.GetImage("compix.lbx", 66, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(x1 + 14), float64(y1 + 18))
            scale.DrawScaled(screen, arrow, &options)
            ax, ay := options.GeoM.Apply(0, 2)
            combat.Fonts.InfoFont.PrintOptions(screen, ax, ay, font.FontOptions{Justify: font.FontJustifyRight, DropShadow: true, Scale: scale.ScaleAmount}, fmt.Sprintf("%v", unit.GetRangedAttackPower()))
    }

    movementImage, _ := combat.ImageCache.GetImage("compix.lbx", 72, 0)
    if unit.IsFlying() {
        movementImage, _ = combat.ImageCache.GetImage("compix.lbx", 73, 0)
    }

    options.GeoM.Reset()
    options.GeoM.Translate(float64(x1 + 14), float64(y1 + 26))
    scale.DrawScaled(screen, movementImage, &options)
    ax, ay = options.GeoM.Apply(0, 2)
    combat.Fonts.InfoFont.PrintOptions(screen, ax, ay, font.FontOptions{Justify: font.FontJustifyRight, DropShadow: true, Scale: scale.ScaleAmount}, fmt.Sprintf("%v", unit.MovesLeft.ToFloat()))

    armorImage, _ := combat.ImageCache.GetImage("compix.lbx", 70, 0)
    options.GeoM.Reset()
    options.GeoM.Translate(float64(x1 + 48), float64(y1 + 10))
    scale.DrawScaled(screen, armorImage, &options)
    ax, ay = options.GeoM.Apply(0, 2)
    combat.Fonts.InfoFont.PrintOptions(screen, ax, ay, font.FontOptions{Justify: font.FontJustifyRight, DropShadow: true, Scale: scale.ScaleAmount}, fmt.Sprintf("%v", unit.GetDefense()))

    resistanceImage, _ := combat.ImageCache.GetImage("compix.lbx", 75, 0)
    options.GeoM.Reset()
    options.GeoM.Translate(float64(x1 + 48), float64(y1 + 18))
    scale.DrawScaled(screen, resistanceImage, &options)
    ax, ay = options.GeoM.Apply(0, 2)
    combat.Fonts.InfoFont.PrintOptions(screen, ax, ay, font.FontOptions{Justify: font.FontJustifyRight, DropShadow: true, Scale: scale.ScaleAmount}, fmt.Sprintf("%v", unit.GetResistance()))

    options.GeoM.Translate(0, 10)
    if unit.GetRangedAttacks() > 0 {
        ax, ay := options.GeoM.Apply(-5, 0)
        combat.Fonts.InfoFont.PrintOptions(screen, ax, ay, font.FontOptions{DropShadow: true, Scale: scale.ScaleAmount}, fmt.Sprintf("%v ammo", unit.GetRangedAttacks()))
    } else if unit.GetCastingSkill() > 0 {
        ax, ay := options.GeoM.Apply(-5, 0)
        combat.Fonts.InfoFont.PrintOptions(screen, ax, ay, font.FontOptions{DropShadow: true, Scale: scale.ScaleAmount}, fmt.Sprintf("%v mp", int(unit.GetCastingSkill())))
    }

    combat.Fonts.InfoFont.PrintOptions(screen, float64(x1 + 14), float64(y1 + 37), font.FontOptions{Justify: font.FontJustifyCenter, DropShadow: true, Scale: scale.ScaleAmount}, "Hits")

    combat.DrawHealthBar(screen, x1 + 25, y1 + 40, 255, unit)

    // draw experience badge
    badge := units.GetExperienceBadge(unit)

    badgeOptions := options
    badgeOptions.GeoM.Translate(-4, 10)
    for range badge.Count {
        pic, _ := combat.ImageCache.GetImage("main.lbx", badge.Badge.IconLbxIndex(), 0)
        scale.DrawScaled(screen, pic, &badgeOptions)
        badgeOptions.GeoM.Translate(5, 0)
    }
}

// draw a horizontal bar that represents the health of the unit
// mostly green if healthy (>66% health)
// yellow if between 33% to 66% health
// otherwise red
func (combat *CombatScreen) DrawHealthBar(screen *ebiten.Image, x int, y int, alpha uint8, unit *ArmyUnit){
    highHealth := color.NRGBA{R: 0, G: 0xff, B: 0, A: alpha}
    mediumHealth := color.NRGBA{R: 0xff, G: 0xff, B: 0, A: alpha}
    lowHealth := color.NRGBA{R: 0xff, G: 0, B: 0, A: alpha}
    healthWidth := 15

    vector.StrokeLine(screen, float32(scale.Scale(x)), float32(scale.Scale(y)), float32(scale.Scale(x + healthWidth)), float32(scale.Scale(y)), float32(scale.Scale(1)), color.NRGBA{R: 0, G: 0, B: 0, A: alpha}, false)

    healthPercent := float64(unit.GetHealth()) / float64(unit.GetMaxHealth())
    healthLength := float64(healthWidth) * healthPercent

    // always show at least one point of health
    if healthLength < 1 {
        healthLength = 1
    }

    useColor := highHealth
    if healthPercent < 0.33 {
        useColor = lowHealth
    } else if healthPercent < 0.66 {
        useColor = mediumHealth
    } else {
        useColor = highHealth
    }

    vector.StrokeLine(screen, float32(scale.Scale(x)), float32(scale.Scale(y)), scale.Scale(float32(x) + float32(healthLength)), float32(scale.Scale(y)), float32(scale.Scale(1)), useColor, false)
}

func (combat *CombatScreen) DrawWall(screen *ebiten.Image, x int, y int, tilePosition func(float64, float64) (float64, float64), animationIndex uint64){
    tile := &combat.Model.Tiles[y][x]
    if tile.Fire == nil && tile.Darkness == nil && tile.Wall == nil {
        return
    }

    var options ebiten.DrawImageOptions

    choose := func(choices []int) int {
        // return a deterministic value based on x,y
        return choices[(x + y) % len(choices)]
    }

    type Order int
    const (
        Order0 Order = iota
        Order1
        Order2
        Order3
        Order4
    )

    // if the tile has fire on the west or north then draw it first, but if the fire is on
    // south or east then draw it last
    // north: fire 0, darkness 1, wall 2
    // south: wall 0, darkness 1, fire 2
    type DrawWallOrder struct {
        Order Order
        Draw func()
    }

    // add things to the list of things to draw, then sort, then draw all by invoking Draw() on each element
    wallDrawOrder := []DrawWallOrder{}

    addDrawWall := func(order Order, draw func()) {
        wallDrawOrder = append(wallDrawOrder, DrawWallOrder{
            Order: order,
            Draw: draw,
        })
    }

    var geom ebiten.GeoM
    geom.Scale(combat.CameraScale, combat.CameraScale)
    tx, ty := tilePosition(float64(x), float64(y))
    geom.Translate(tx, ty)

    drawAnimatedWall := func (index int, activeCounter uint64, dx float64, dy float64){
        options.GeoM.Reset()
        // options.GeoM.Scale(combat.CameraScale, combat.CameraScale)
        // options.GeoM.Translate(tx, ty)
        options.GeoM.Translate(dx, dy)

        var drawImage *ebiten.Image

        activeImages, _ := combat.ImageCache.GetImages("wallrise.lbx", index)

        // FIXME: this /8 should be a parameter to DrawWall or something
        if (combat.Counter - activeCounter) / 8 < uint64(len(activeImages)) {
            index := (combat.Counter - activeCounter) / 8
            drawImage = activeImages[index]
        } else {
            images, _ := combat.ImageCache.GetImages("citywall.lbx", index)
            use := animationIndex % uint64(len(images))
            drawImage = images[use]
        }

        options.GeoM.Translate(-float64(drawImage.Bounds().Dy())/2, -float64(drawImage.Bounds().Dy()/2))

        options.GeoM.Concat(geom)

        scale.DrawScaled(screen, drawImage, &options)
    }

    makeDrawAnimatedWall := func(index int, activeCounter uint64, dx float64, dy float64) func() {
        return func() {
            drawAnimatedWall(index, activeCounter, dx, dy)
        }
    }

    // lbx indices for fire
    fireWest := []int{37, 38, 39}
    fireNorth := []int{40, 41, 42}
    fireSouth := []int{43, 44, 48}
    fireEast := []int{46, 47, 49}

    fire := tile.Fire
    if fire != nil {
        drewNorth := false
        drewSouth := false
        drewEast := false
        drewWest := false

        // draw the same fire animation for a given x,y tile, but choose a different fire
        // animation for other tiles

        // these values are based on a clockwise 45-degree rotation, but the actual
        // combat screen is a counter-clockwise 45-degree rotation.
        // it doesn't matter, as long as the fire animations are consistent
        if fire.Contains(FireSideNorth) && fire.Contains(FireSideWest) {
            addDrawWall(Order0, makeDrawAnimatedWall(36, tile.FireActive, -1, -8))
            drewNorth = true
            drewWest = true
        }

        if fire.Contains(FireSideSouth) && fire.Contains(FireSideEast) {
            addDrawWall(Order2, makeDrawAnimatedWall(45, tile.FireActive, -2, -3))
            drewSouth = true
            drewEast = true
        }

        if !drewSouth && fire.Contains(FireSideSouth) {
            addDrawWall(Order2, makeDrawAnimatedWall(choose(fireSouth), tile.FireActive, -4, -4))
        }

        if !drewWest && fire.Contains(FireSideWest) {
            addDrawWall(Order0, makeDrawAnimatedWall(choose(fireWest), tile.FireActive, -3, -6))
        }

        if !drewNorth && fire.Contains(FireSideNorth) {
            addDrawWall(Order0, makeDrawAnimatedWall(choose(fireNorth), tile.FireActive, 2, -6))
        }

        if !drewEast && fire.Contains(FireSideEast) {
            addDrawWall(Order2, makeDrawAnimatedWall(choose(fireEast), tile.FireActive, 2, -4))
        }
    }

    darknessWest := []int{51, 52, 53}
    darknessNorth := []int{54, 55, 56}
    darknessSouth := []int{57, 58, 62}
    darknessEast := []int{60, 61, 63}

    darkness := tile.Darkness
    if darkness != nil {
        // lbx indices for fire

        drewNorth := false
        drewSouth := false
        drewEast := false
        drewWest := false

        // draw the same fire animation for a given x,y tile, but choose a different fire
        // animation for other tiles

        // these values are based on a clockwise 45-degree rotation, but the actual
        // combat screen is a counter-clockwise 45-degree rotation.
        // it doesn't matter, as long as the fire animations are consistent
        if darkness.Contains(DarknessSideNorth) && darkness.Contains(DarknessSideWest) {
            addDrawWall(Order1, makeDrawAnimatedWall(50, tile.DarknessActive, -1, -8))
            drewNorth = true
            drewWest = true
        }

        if darkness.Contains(DarknessSideSouth) && darkness.Contains(DarknessSideEast) {
            addDrawWall(Order1, makeDrawAnimatedWall(59, tile.DarknessActive, -2, -3))
            drewSouth = true
            drewEast = true
        }

        if !drewSouth && darkness.Contains(DarknessSideSouth) {
            addDrawWall(Order1, makeDrawAnimatedWall(choose(darknessSouth), tile.DarknessActive, -4, -4))
        }

        if !drewWest && darkness.Contains(DarknessSideWest) {
            addDrawWall(Order1, makeDrawAnimatedWall(choose(darknessWest), tile.DarknessActive, -3, -6))
        }

        if !drewNorth && darkness.Contains(DarknessSideNorth) {
            addDrawWall(Order1, makeDrawAnimatedWall(choose(darknessNorth), tile.DarknessActive, 2, -6))
        }

        if !drewEast && darkness.Contains(DarknessSideEast) {
            addDrawWall(Order1, makeDrawAnimatedWall(choose(darknessEast), tile.DarknessActive, 2, -4))
        }
    }

    wall := tile.Wall
    if wall != nil {
        // starting index for the wall. there are 3 types of wall: normal, dark, and grass/ivy covered
        wallBase := []int{0, 12, 24}
        currentWall := 0

        // lbx indices for fire, relative to wallBase
        west := []int{1, 2}
        north := []int{4, 5}
        south := []int{7, 8}
        east := []int{10}
        gate := 11
        northWest := 0
        southWest := 3
        northEast := 6
        southEast := 9

        drawWall := func(index int, dx float64, dy float64){
            options.GeoM.Reset()
            // options.GeoM.Scale(combat.CameraScale, combat.CameraScale)
            // options.GeoM.Translate(tx, ty)
            options.GeoM.Translate(dx, dy)

            // FIXME: a destroyed wall should use index 1 (last argument)
            drawImage, _ := combat.ImageCache.GetImage("citywall.lbx", wallBase[currentWall] + index, 0)
            // use := animationIndex % uint64(len(images))
            // drawImage := images[use]
            options.GeoM.Translate(-float64(drawImage.Bounds().Dy())/2, -float64(drawImage.Bounds().Dy()/2))

            options.GeoM.Concat(geom)

            scale.DrawScaled(screen, drawImage, &options)
        }

        makeDrawWall := func(index int, dx float64, dy float64) func() {
            return func() {
                drawWall(index, dx, dy)
            }
        }

        drewNorth := false
        drewSouth := false
        drewEast := false
        drewWest := false

        // draw the same fire animation for a given x,y tile, but choose a different fire
        // animation for other tiles

        // these values are based on a clockwise 45-degree rotation, but the actual
        // combat screen is a counter-clockwise 45-degree rotation.
        // it doesn't matter, as long as the fire animations are consistent
        if wall.Contains(WallKindNorth) && wall.Contains(WallKindWest) {
            addDrawWall(Order2, makeDrawWall(northWest, -1, -8))
            drewNorth = true
            drewWest = true
        }

        if wall.Contains(WallKindSouth) && wall.Contains(WallKindEast) {
            addDrawWall(Order0, makeDrawWall(southEast, -2, -3))
            drewSouth = true
            drewEast = true
        }

        if wall.Contains(WallKindSouth) && wall.Contains(WallKindWest) {
            addDrawWall(Order2, makeDrawWall(southWest, -2, -3))
            drewSouth = true
            drewWest = true

            if tile.Darkness != nil && tile.Darkness.Contains(DarknessSideSouth) {
                addDrawWall(Order3, makeDrawAnimatedWall(choose(darknessSouth), tile.DarknessActive, -4, -4))
            }

            if tile.Fire != nil && tile.Fire.Contains(FireSideSouth) {
                addDrawWall(Order4, makeDrawAnimatedWall(choose(fireSouth), tile.FireActive, -4, -4))
            }
        }

        if wall.Contains(WallKindNorth) && wall.Contains(WallKindEast) {
            addDrawWall(Order2, makeDrawWall(northEast, -2, -3))
            drewNorth = true
            drewEast = true

            if tile.Darkness != nil && tile.Darkness.Contains(DarknessSideEast) {
                addDrawWall(Order3, makeDrawAnimatedWall(choose(darknessEast), tile.DarknessActive, 2, -4))
            }

            if tile.Fire != nil && tile.Fire.Contains(FireSideEast) {
                addDrawWall(Order4, makeDrawAnimatedWall(choose(fireEast), tile.FireActive, 2, -4))
            }
        }

        if !drewSouth && wall.Contains(WallKindSouth) {
            addDrawWall(Order0, makeDrawWall(choose(south), -4, -4))
        }

        if !drewWest && wall.Contains(WallKindWest) {
            addDrawWall(Order2, makeDrawWall(choose(west), -3, -6))
        }

        if !drewNorth && wall.Contains(WallKindNorth) {
            addDrawWall(Order2, makeDrawWall(choose(north), 2, -6))
        }

        if !drewEast && wall.Contains(WallKindEast) {
            addDrawWall(Order0, makeDrawWall(choose(east), 2, -4))
        }

        if wall.Contains(WallKindGate) {
            addDrawWall(Order0, makeDrawWall(gate, -2, -4))
        }
    }

    slices.SortFunc(wallDrawOrder, func(a, b DrawWallOrder) int {
        return cmp.Compare(a.Order, b.Order)
    })

    for _, draw := range wallDrawOrder {
        draw.Draw()
    }
}

func (combat *CombatScreen) Draw(screen *ebiten.Image){
    combat.Drawer(screen)
}

func (combat *CombatScreen) makeIsUnitVisibleFunc() func(*ArmyUnit) bool {
    teamHasIllusionImmunity := functional.Memoize(func(team Team) bool {
        army := combat.Model.GetArmyForTeam(team)
        for _, unit := range army.units {
            if unit.HasAbility(data.AbilityIllusionsImmunity) {
                return true
            }
        }

        return false
    })

    return func(unit *ArmyUnit) bool {
        if !unit.IsInvisible() {
            return true
        }

        owner := combat.Model.GetArmy(unit)
        return owner.Player.IsHuman() || teamHasIllusionImmunity(oppositeTeam(unit.Team)) || combat.Model.IsAdjacentToEnemy(unit)
    }
}

func (combat *CombatScreen) IsUnitVisible(unit *ArmyUnit) bool {
    return combat.makeIsUnitVisibleFunc()(unit)
}

func getDyingColor(unit *ArmyUnit) color.RGBA {
    if unit.GetRealm() == data.MagicNone {
        return color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff} // red
    }

    return data.GetMagicColor(unit.GetRealm())
}

func (combat *CombatScreen) ShowExtraHighlight(screen *ebiten.Image, unit *ArmyUnit, getTilePoints func(int, int) ([]image.Point)) {
    points := getTilePoints(unit.X, unit.Y)
    p1 := points[0]
    // p2 := points[1]
    p3 := points[2]
    p4 := points[3]

    // draw quad with bottom two points (x3,y3) and (x4,y4), and top two points (x3, 0) and (x4, 0)
    var fillOptions vector.FillOptions

    drawQuad := func(x1, y1, x2, y2, x3, y3, x4, y4 float32, col color.NRGBA) {
        var path vector.Path
        path.MoveTo(x1, y1)
        path.LineTo(x2, y2)
        path.LineTo(x3, y3)
        path.LineTo(x4, y4)
        path.Close()

        var pathOptions vector.DrawPathOptions
        pathOptions.ColorScale.ScaleWithColor(col)

        vector.FillPath(screen, &path, &fillOptions, &pathOptions)
    }

    getAlpha := func(phase uint64) uint8 {
        angle := float64(phase) * (math.Pi / 180)

        var rangeMin float64 = 64
        var rangeMax float64 = 180

        alpha := math.Sin(angle * 2) * (rangeMax - rangeMin) / 2 + (rangeMin + rangeMax) / 2
        if alpha < rangeMin {
            alpha = rangeMin
        }
        if alpha > rangeMax {
            alpha = rangeMax
        }

        return uint8(alpha)
    }

    basePhase := combat.Counter
    drawQuad(float32(p3.X), float32(p3.Y), float32(p4.X), float32(p4.Y), float32(p4.X), float32(0), float32(p3.X), float32(0), color.NRGBA{R: 255, G: 255, B: 255, A: getAlpha(basePhase)})
    drawQuad(float32(p1.X), float32(p1.Y), float32(p4.X), float32(p4.Y), float32(p4.X), float32(0), float32(p1.X), float32(0), color.NRGBA{R: 255, G: 255, B: 255, A: getAlpha(basePhase + 30)})
}

func (combat *CombatScreen) ShowCombatInfo(screen *ebiten.Image) {

    x1 := 2
    y1 := 10
    x2 := data.ScreenWidth - 2
    y2 := data.ScreenHeight - 20

    subScreen := screen.SubImage(image.Rect(scale.Scale(x1), scale.Scale(y1), scale.Scale(x2), scale.Scale(y2))).(*ebiten.Image)

    var alpha uint8 = uint8(180 * combat.ShowInfoLevel / 100)

    var fontOptions ebiten.DrawImageOptions
    fontOptions.ColorScale.ScaleAlpha(float32(alpha) / 255)

    vector.FillRect(subScreen, float32(scale.Scale(x1)), float32(scale.Scale(y1)), float32(scale.Scale(x2)), float32(scale.Scale(y2)), color.NRGBA{R: 0xd7, G: 0xac, B: 0x5a, A: alpha}, false)
    vector.StrokeRect(subScreen, float32(scale.Scale(x1)), float32(scale.Scale(y1)), float32(scale.Scale(x2-x1)), float32(scale.Scale(y2-y1)), 2, color.NRGBA{R: 0, G: 0, B: 0, A: alpha}, true)

    lineX := (x1 + x2) / 2
    vector.StrokeLine(subScreen, float32(scale.Scale(lineX)), float32(scale.Scale(y1)), float32(scale.Scale(lineX)), float32(scale.Scale(y2)), 2, color.RGBA{R: 0, G: 0, B: 0, A: alpha}, false)

    combat.Fonts.AttackingWizardFont.PrintOptions(
        subScreen,
        float64(x1 + 80),
        float64(y1 + 10),
        font.FontOptions{Justify: font.FontJustifyCenter, Scale: scale.ScaleAmount, DropShadow: true, Options: &fontOptions},
        combat.Model.AttackingArmy.Player.GetWizard().Name,
    )

    defendX := lineX + 60
    combat.Fonts.DefendingWizardFont.PrintOptions(
        subScreen,
        float64(defendX),
        float64(y1 + 10),
        font.FontOptions{Justify: font.FontJustifyCenter, Scale: scale.ScaleAmount, DropShadow: true, Options: &fontOptions},
        combat.Model.DefendingArmy.Player.GetWizard().Name,
    )

    showUnits := func(startX int, army *Army) {
        startY := y1 + 10 + 20
        unitFont := combat.Fonts.HudFont
        for i, unit := range army.units {

            unitX := startX + (i % 3) * 49
            unitY := startY + (i / 3) * unitFont.Height() * 5

            unitFont.PrintOptions(subScreen, float64(unitX), float64(unitY), font.FontOptions{Scale: scale.ScaleAmount, Options: &fontOptions}, unit.Unit.GetName())
            unitY += unitFont.Height()
            unitFont.PrintOptions(subScreen, float64(unitX), float64(unitY), font.FontOptions{Scale: scale.ScaleAmount, Options: &fontOptions}, fmt.Sprintf("%v/%v HP", unit.GetHealth(), unit.GetMaxHealth()))
            unitY += unitFont.Height()
            unitImage, err := unitview.GetUnitOverworldImage(&combat.ImageCache, unit.Unit)
            if err == nil {
                var options ebiten.DrawImageOptions
                options.GeoM.Translate(float64(unitX), float64(unitY))
                options.ColorScale.ScaleAlpha(float32(alpha) / 255)
                scale.DrawScaled(subScreen, unitImage, &options)

                for _, enchantment := range unit.GetEnchantments() {
                    util.DrawOutline(subScreen, &combat.ImageCache, unitImage, scale.ScaleGeom(options.GeoM), options.ColorScale, combat.Counter/8, enchantment.Color())
                    break
                }

                combat.DrawHealthBar(subScreen, unitX + unitImage.Bounds().Dx() + 2, unitY + unitImage.Bounds().Dy() / 2, alpha, unit)
                unitY += unitImage.Bounds().Dy() + 2
            }

        }
    }

    showUnits(x1 + 5, combat.Model.AttackingArmy)
    showUnits(x1 + defendX - 57, combat.Model.DefendingArmy)

}

func (combat *CombatScreen) NormalDraw(screen *ebiten.Image) {
    isVisible := functional.Memoize(combat.makeIsUnitVisibleFunc())

    animationIndex := combat.Counter / 8

    var options ebiten.DrawImageOptions

    tile0, _ := combat.ImageCache.GetImage("cmbgrass.lbx", 0, 0)

    /*
    a, b := screenToTile.Apply(160, 100)
    log.Printf("(160,100) -> (%f, %f)", a, b)
    */

    /*
    a, b := coordinates.Apply(3, 0)
    log.Printf("(3,3) -> (%f, %f)", a, b)
    */

    useMatrix := combat.GetCameraMatrix()

    tilePosition := func(x float64, y float64) (float64, float64){
        return useMatrix.Apply(x, y)
    }

    // draw base land first
    for _, point := range combat.TopDownOrder {
        x := point.X
        y := point.Y

        images, _ := combat.ImageCache.GetImages(combat.Model.Tiles[y][x].Lbx, combat.Model.Tiles[y][x].Index)
        image := images[animationIndex % uint64(len(images))]
        options.GeoM.Reset()
        // tx,ty is the middle of the tile
        tx, ty := tilePosition(float64(x), float64(y))
        options.GeoM.Scale(combat.CameraScale, combat.CameraScale)
        options.GeoM.Translate(tx, ty)
        scale.DrawScaled(screen, image, &options)

        if combat.Model.Tiles[y][x].Mud {
            mudTiles, _ := combat.ImageCache.GetImages("cmbtcity.lbx", 118)
            index := animationIndex % uint64(len(mudTiles))
            scale.DrawScaled(screen, mudTiles[index], &options)
        }

        // vector.DrawFilledCircle(screen, float32(tx), float32(ty), 2, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, false)
    }

    if combat.DrawRoad {
        tx, ty := tilePosition(TownCenterX-1, TownCenterY-4)

        road, _ := combat.ImageCache.GetImageTransform("cmbtcity.lbx", 0, 0, "crop", util.AutoCrop)
        options.GeoM.Reset()
        options.GeoM.Scale(combat.CameraScale, combat.CameraScale)
        options.GeoM.Translate(tx, ty)
        options.GeoM.Translate(0, float64(tile0.Bounds().Dy())/2)
        scale.DrawScaled(screen, road, &options)
    }

    if combat.DrawClouds {
        tx, ty := tilePosition(TownCenterX, TownCenterY-5)

        clouds, _ := combat.ImageCache.GetImage("cmbtcity.lbx", 113, 0)
        options.GeoM.Reset()
        options.GeoM.Scale(combat.CameraScale, combat.CameraScale)
        options.GeoM.Translate(tx, ty)
        options.GeoM.Translate(0, float64(tile0.Bounds().Dy())/2)
        scale.DrawScaled(screen, clouds, &options)
    }

    drawExtraObject := func(x int, y int, extra TileTop) {
        if extra.Drawer != nil {
            options.GeoM.Reset()
            tx, ty := tilePosition(float64(x), float64(y))
            options.GeoM.Scale(combat.CameraScale, combat.CameraScale)
            options.GeoM.Translate(tx, ty)

            extra.Drawer(screen, &combat.ImageCache, &options, combat.Counter)
        } else if extra.Index != -1 {
            options.GeoM.Reset()
            // tx,ty is the middle of the tile
            tx, ty := tilePosition(float64(x), float64(y))

            var geom ebiten.GeoM

            geom.Scale(combat.CameraScale, combat.CameraScale)
            geom.Translate(tx, ty)

            extraImages, _ := combat.ImageCache.GetImagesTransform(extra.Lbx, extra.Index, "crop", util.AutoCrop)

            index := animationIndex % uint64(len(extraImages))
            extraImage := extraImages[index]

            switch extra.Alignment {
                case TileAlignBottom:
                    options.GeoM.Translate(0, float64(tile0.Bounds().Dy())/2)
                    options.GeoM.Translate(-float64(extraImage.Bounds().Dy())/2, -float64(extraImage.Bounds().Dy()))
                case TileAlignMiddle:
                    options.GeoM.Translate(-float64(extraImage.Bounds().Dy())/2, -float64(extraImage.Bounds().Dy()/2))
            }

            options.GeoM.Concat(geom)

            scale.DrawScaled(screen, extraImage, &options)

            // vector.DrawFilledCircle(screen, float32(tx), float32(ty), 2, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, false)
        }
    }

    // then draw extra stuff on top
    for _, point := range combat.TopDownOrder {
        x := point.X
        y := point.Y

        extra := combat.Model.Tiles[y][x].ExtraObject
        drawExtraObject(x, y, extra)
    }

    combat.DrawHighlightedTile(screen, combat.MouseTileX, combat.MouseTileY, &useMatrix, color.NRGBA{R: 0, G: 0x67, B: 0x78, A: 255}, color.NRGBA{R: 0, G: 0xef, B: 0xff, A: 255})

    if combat.Model.SelectedUnit != nil && isVisible(combat.Model.SelectedUnit) {
        // if the unit is currently selecting a spell, then don't draw the movement path
        // also don't draw the movement path if the unit can teleport
        if !combat.IsSelectingSpell() && ! combat.Model.SelectedUnit.CanTeleport(){
            var path pathfinding.Path
            ok := false

            if combat.Model.SelectedUnit.Moving {
                path = combat.Model.SelectedUnit.CurrentPath
                ok = true
            } else {
                path, ok = combat.Model.FindPath(combat.Model.SelectedUnit, combat.MouseTileX, combat.MouseTileY, combat.ExtraControl)
                if ok {
                    path = path[1:]
                }
            }

            if ok {
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(0.8)

                moves := combat.Model.SelectedUnit.MovesLeft

                lastX := combat.Model.SelectedUnit.X
                lastY := combat.Model.SelectedUnit.Y

                movementImage, _ := combat.ImageCache.GetImage("compix.lbx", 72, 0)
                showBad := false

                for i := range path {
                    tileX, tileY := path[i].X, path[i].Y

                    tx, ty := tilePosition(float64(tileX), float64(tileY))
                    // tx += float64(tile0.Bounds().Dx())/2
                    // ty += float64(tile0.Bounds().Dy())/2

                    // show boots
                    tx -= float64(movementImage.Bounds().Dx())/2
                    ty -= float64(movementImage.Bounds().Dy())/2

                    options.GeoM.Reset()
                    options.GeoM.Scale(combat.CameraScale, combat.CameraScale)
                    options.GeoM.Translate(tx, ty)

                    if !showBad && moves.LessThanEqual(fraction.FromInt(0)) {
                        showBad = true
                        options.ColorScale.Scale(0.1, 0.1, 0.1, 1)
                    }

                    scale.DrawScaled(screen, movementImage, &options)

                    moves = moves.Subtract(pathCost(image.Pt(lastX, lastY), image.Pt(tileX, tileY)))
                }
            }
        }

        if !combat.Model.SelectedUnit.Moving {
            minColor := color.NRGBA{R: 32, G: 0, B: 0, A: 255}
            maxColor := color.NRGBA{R: 255, G: 0, B: 0, A: 255}

            combat.DrawHighlightedTile(screen, combat.Model.SelectedUnit.X, combat.Model.SelectedUnit.Y, &useMatrix, minColor, maxColor)
        }
    }

    renderUnit := func(unit *ArmyUnit) {
        var unitOptions ebiten.DrawImageOptions
        banner := unit.Unit.GetBanner()
        combatImages, _ := combat.ImageCache.GetImagesTransform(unit.Unit.GetCombatLbxFile(), unit.Unit.GetCombatIndex(unit.Facing), banner.String(), units.MakeUpdateUnitColorsFunc(banner))

        if combatImages != nil {
            var tx float64
            var ty float64

            if unit.Moving {
                tx, ty = tilePosition(unit.MoveX, unit.MoveY)
            } else {
                tx, ty = tilePosition(float64(unit.X), float64(unit.Y))
            }

            /*
            tx, ty := tilePosition(float64(x), float64(y))
            options.GeoM.Scale(combat.CameraScale, combat.CameraScale)
            options.GeoM.Translate(tx, ty)
            */

            unitOptions.GeoM.Scale(combat.CameraScale, combat.CameraScale)
            unitOptions.GeoM.Translate(tx, ty)
            // unitOptions.GeoM.Translate(float64(tile0.Bounds().Dx()/2) * combat.CameraScale, 0)
            // unitOptions.GeoM.Translate(float64(tile0.Bounds().Dx()/2), float64(tile0.Bounds().Dy()/2))
            // unitOptions.GeoM.Translate(0, float64(tile0.Bounds().Dy()/2))

            // unitOptions.GeoM.Translate(0, float64(-unit.Height))

            index := uint64(0)
            // sort of a hack here, but we use unit.Unit.IsFlying() to bypass a webbed unit so that the unit
            // still animates if it was originally flying
            if unit.Unit.IsFlying() || unit.Moving {
                index = animationIndex % (uint64(len(combatImages)) - 1)
            }

            if unit.Attacking || unit.Defending {
                index = 2 + animationIndex % 2
            }

            if combat.Model.SelectedUnit == unit {
                scaleValue := 1.5 + math.Sin(float64(combat.Counter)/6)/2
                unitOptions.ColorScale.Scale(float32(scaleValue), float32(scaleValue), 1, 1)
            }

            if combat.Model.HighlightedUnit == unit {
                scaleValue := 1.5 + math.Sin(float64(combat.Counter)/6)/2
                unitOptions.ColorScale.Scale(float32(scaleValue), 1, 1, 1)
            }

            // for summoning units out of the ground, or the merging ability
            unitImage := combatImages[index]
            if unit.Height < 100 {
                // log.Printf("unit %v has height %v", unit.Unit.GetName(), unit.Height)
                // unitImage = unitImage.SubImage(image.Rect(0, 0, unitImage.Bounds().Dx(), unitImage.Bounds().Dy() + unit.Height)).(*ebiten.Image)
                unitImage = unitImage.SubImage(image.Rect(0, 0, unitImage.Bounds().Dx(), unitImage.Bounds().Dy() + unit.Height)).(*ebiten.Image)
            }

            // for units that teleport
            unitOptions.ColorScale.ScaleAlpha(1 - unit.Fade)

            /*
            x, y := unitOptions.GeoM.Apply(0, 0)
            vector.DrawFilledCircle(screen, float32(x), float32(y), 2, color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}, false)
            x, y = unitOptions.GeoM.Apply(float64(tile0.Bounds().Dx()/2), 0)
            vector.DrawFilledCircle(screen, float32(x), float32(y), 2, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, false)
            */

            lostTime := float64(unit.LostUnitsTime) / LostUnitsMax

            var dying colorm.ColorM
            dyingColor := getDyingColor(unit)
            r, g, b, _ := dyingColor.RGBA()
            dying.Scale(0, 0, 0, lostTime)
            // normalize each color component to [0, 1] range
            dying.Translate(float64(r)/256/256, float64(g)/256/256, float64(b)/256/256, 0)

            // _ = index
            use := util.First(unit.GetEnchantments(), data.UnitEnchantmentNone)
            if unit.IsInvisible() {
                // might not be visible at all, or is semi-visible if next to an enemy unit or if the enemy team has
                // any units with illusions immunity
                canBeSeen := isVisible(unit)
                if canBeSeen {
                    unitview.RenderCombatSemiInvisible(screen, unitImage, unitOptions, unit.VisibleFigures(), unit.LostUnits, &dying, combat.Counter, &combat.ImageCache)
                } else {
                    // if can't be seen then don't render anything at all
                }

            } else if unit.IsAsleep() {
                unitview.RenderCombatUnitGrey(screen, unitImage, unitOptions, unit.VisibleFigures(), unit.LostUnits, &dying, use, combat.Counter, &combat.ImageCache)
            } else {
                warpCreature := false
                for _, curse := range unit.GetCurses() {
                    switch curse {
                        case data.UnitCurseWarpCreatureDefense,
                             data.UnitCurseWarpCreatureMelee,
                             data.UnitCurseWarpCreatureResistance: warpCreature = true
                    }
                    if warpCreature {
                        break
                    }
                }

                var savedColor ebiten.ColorScale
                if warpCreature {
                    savedColor = unitOptions.ColorScale
                    unitOptions.ColorScale.ScaleWithColor(color.RGBA{R: 0xb5, G: 0x5e, B: 0xf3, A: 0xff})
                }

                unitview.RenderCombatUnit(screen, unitImage, unitOptions, unit.VisibleFigures(), unit.LostUnits, &dying, use, combat.Counter, &combat.ImageCache)

                if warpCreature {
                    unitOptions.ColorScale = savedColor
                }
            }

            var curseOptions ebiten.DrawImageOptions

            curseOptions.GeoM.Translate(float64(-unitImage.Bounds().Dx()/2), float64(-unitImage.Bounds().Dy()*3/4))
            curseOptions.GeoM.Concat(unitOptions.GeoM)
            for _, curse := range unit.GetCurses() {
                switch curse {
                    case data.UnitCurseMindStorm:
                        images, _ := combat.ImageCache.GetImages("resource.lbx", 78)
                        index := animationIndex % uint64(len(images))
                        use := images[index]

                        scale.DrawScaled(screen, use, &curseOptions)
                    case data.UnitCurseWeakness:
                        images, _ := combat.ImageCache.GetImages("resource.lbx", 80)
                        index := animationIndex % uint64(len(images))
                        use := images[index]

                        scale.DrawScaled(screen, use, &curseOptions)
                    case data.UnitCurseVertigo:
                        images, _ := combat.ImageCache.GetImages("cmbtfx.lbx", 17)
                        index := animationIndex % uint64(len(images))
                        use := images[index]

                        scale.DrawScaled(screen, use, &curseOptions)
                    case data.UnitCurseShatter:
                        images, _ := combat.ImageCache.GetImages("resource.lbx", 79)
                        index := animationIndex % uint64(len(images))
                        use := images[index]

                        scale.DrawScaled(screen, use, &curseOptions)
                    case data.UnitCurseConfusion:
                        images, _ := combat.ImageCache.GetImages("resource.lbx", 76)
                        index := animationIndex % uint64(len(images))
                        use := images[index]

                        scale.DrawScaled(screen, use, &curseOptions)
                }
            }

            if unit.IsWebbed() {
                image, _ := combat.ImageCache.GetImage("resource.lbx", 82, 0)
                scale.DrawScaled(screen, image, &curseOptions)
            }
        }
    }

    type Drawable struct {
        GetY func() float64
        GetX func() float64
        Render func()
    }

    unitDrawable := func(unit *ArmyUnit) Drawable {
        return Drawable{
            GetX: func() float64 {
                return float64(unit.X)
            },
            GetY: func() float64 {
                return float64(unit.Y)
            },
            Render: func() {
                renderUnit(unit)
            },
        }
    }

    wallDrawable := func(tile TilePoint) Drawable {
        var offset float64 = 0

        if tile.Tile.HasEasternWall() || tile.Tile.HasSouthernWall() {
            offset += 0.1
        } else {
            offset -= 0.1
        }

        return Drawable{
            GetX: func() float64 {
                return float64(tile.X)
            },
            GetY: func() float64 {
                return float64(tile.Y) + offset
            },
            Render: func() {
                combat.DrawWall(screen, tile.X, tile.Y, tilePosition, animationIndex)
            },
        }
    }

    fortressDrawable := func() Drawable {
        extra := TileTop{
            Lbx: "cmbtcity.lbx",
            Index: 17,
            Alignment: TileAlignBottom,
        }

        return Drawable{
            GetX: func() float64 {
                return float64(TownCenterX)
            },
            GetY: func() float64 {
                return float64(TownCenterY)
            },
            Render: func() {
                drawExtraObject(TownCenterX, TownCenterY, extra)
            },
        }
    }

    // sort units in top down order before drawing them
    allDrawables := make([]Drawable, 0, len(combat.Model.DefendingArmy.units) + len(combat.Model.AttackingArmy.units) + 100)

    for _, unit := range combat.Model.AttackingArmy.units {
        allDrawables = append(allDrawables, unitDrawable(unit))
    }

    for _, unit := range combat.Model.DefendingArmy.units {
        allDrawables = append(allDrawables, unitDrawable(unit))
    }

    if combat.Model.Zone.City != nil && combat.Model.Zone.City.HasFortress() {
        allDrawables = append(allDrawables, fortressDrawable())
    }

    for _, unit := range combat.Model.AttackingArmy.KilledUnits {
        if unit.LostUnitsTime > 0 {
            allDrawables = append(allDrawables, unitDrawable(unit))
        }
    }

    for _, unit := range combat.Model.DefendingArmy.KilledUnits {
        if unit.LostUnitsTime > 0 {
            allDrawables = append(allDrawables, unitDrawable(unit))
        }
    }

    for _, tile := range combat.Model.WallTiles() {
        allDrawables = append(allDrawables, wallDrawable(tile))
    }

    compareDrawable := func(drawA Drawable, drawB Drawable) int {
        ax, ay := tilePosition(drawA.GetX(), drawA.GetY())
        bx, by := tilePosition(drawB.GetX(), drawB.GetY())

        if ay < by {
            return -1
        }

        if ay > by {
            return 1
        }

        if ax < bx {
            return -1
        }

        if ax > bx {
            return 1
        }

        return 0
    }

    slices.SortFunc(allDrawables, compareDrawable)

    for _, drawable := range allDrawables {
        drawable.Render()
    }

    for _, unit := range combat.Model.OtherUnits {
        var unitOptions ebiten.DrawImageOptions
        frame := unit.Animation.Frame()
        unitOptions.GeoM.Translate(float64(-frame.Bounds().Dx()/2), float64(-frame.Bounds().Dy()))
        // unitOptions.GeoM.Translate(float64(tile0.Bounds().Dx()/2), float64(tile0.Bounds().Dy()/2))
        unitOptions.GeoM.Translate(0, float64(tile0.Bounds().Dy()/2))

        tx, ty := tilePosition(float64(unit.X), float64(unit.Y))
        unitOptions.GeoM.Scale(combat.CameraScale, combat.CameraScale)
        unitOptions.GeoM.Translate(tx, ty)

        scale.DrawScaled(screen, frame, &unitOptions)
    }

    if combat.ExtraHighlightedUnit != nil {
        getTilePoints := func(x int, y int) ([]image.Point){
            var geom ebiten.GeoM
            tx, ty := useMatrix.Apply(float64(x), float64(y))
            geom.Scale(combat.CameraScale, combat.CameraScale)
            geom.Translate(tx, ty)

            geom = scale.ScaleGeom(geom)

            // left
            x1, y1 := geom.Apply(float64(-tile0.Bounds().Dx()) / 2, 0)

            // top
            x2, y2 := geom.Apply(0, float64(-tile0.Bounds().Dy()) / 2)

            // right
            x3, y3 := geom.Apply(float64(tile0.Bounds().Dx()) / 2, 0)

            // bottom
            x4, y4 := geom.Apply(0, float64(tile0.Bounds().Dy()) / 2)

            return []image.Point{
                image.Point{X: int(x1), Y: int(y1)},
                image.Point{X: int(x2), Y: int(y2)},
                image.Point{X: int(x3), Y: int(y3)},
                image.Point{X: int(x4), Y: int(y4)},
            }
        }

        combat.ShowExtraHighlight(screen, combat.ExtraHighlightedUnit, getTilePoints)
    }

    combat.UI.Draw(combat.UI, screen)

    for _, indicator := range combat.DamageIndicators {
        tx, ty := tilePosition(float64(indicator.X), float64(indicator.Y))
        tx += float64(indicator.Offset)
        ty -= 12
        ty -= float64(indicator.Count) / 5
        var options ebiten.DrawImageOptions
        if indicator.Life < 10 {
            options.ColorScale.ScaleAlpha(float32(indicator.Life) / 10)
        }
        if indicator.Damage > 8 {
            options.ColorScale.Scale(1, 0.5, 0.5, 1)
        } else if indicator.Damage > 3 {
            options.ColorScale.Scale(1, 0.75, 0.75, 1)
        } else {
            options.ColorScale.Scale(1.5, 1.5, 1.5, 1)
        }
        combat.Fonts.InfoFont.PrintOptions(screen, tx, ty, font.FontOptions{Justify: font.FontJustifyCenter, Scale: scale.ScaleAmount, Options: &options, DropShadow: true}, fmt.Sprintf("%d", indicator.Damage))
    }

    if combat.Model.HighlightedUnit != nil && isVisible(combat.Model.HighlightedUnit) {
        combat.ShowUnitInfo(screen, combat.Model.HighlightedUnit)
    }

    for _, projectile := range combat.Model.Projectiles {
        var frame *ebiten.Image
        if projectile.Exploding {
            frame = projectile.Explode.Frame()
        } else {
            frame = projectile.Animation.Frame()
        }
        if frame != nil {
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(-frame.Bounds().Dx()/2), float64(-frame.Bounds().Dy())/2)
            options.GeoM.Scale(combat.CameraScale, combat.CameraScale)
            options.GeoM.Translate(projectile.X, projectile.Y)
            scale.DrawScaled(screen, frame, &options)
        }
    }

    if combat.ShowInfoLevel > 0 {
        combat.ShowCombatInfo(screen)
    }
}
