package ai

/* EnemyAI is the AI for enemy wizards
 * Actions an enemy can take:
 *  research new spell
 *  cast spell
 *  each city can be producing something, or housing/trade goods
 *  each unit can patrol to defend an area, move to attack an enemy, or move to explore
 *  diplomacy with other wizards
 *
 * given a goal, create a plan. each turn the AI will try to execute part of the plan
 * after exploration, the AI may need to replan
 */

import (
    "fmt"
    "log"
    "slices"
    "cmp"
    "math"
    "math/rand/v2"
    "image"
    "strings"

    "github.com/kazzmir/master-of-magic/lib/functional"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/algorithm"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
)

type Enemy2AI struct {
    playerlib.DefaultAIEvents

    // units that are currently moving towards an enemy
    Attacking map[*playerlib.UnitStack]bool

    // a catalogue of every lair / encounter / magic node this wizard has
    // discovered (explored the tile of). Persists across turns. Once a stack
    // gets a look at the contents (the tile becomes visible) the entry is
    // marked Scouted and its monster Strength is recorded so the AI can decide
    // whether it can field an army strong enough to clear it.
    KnownLairs map[LairKey]*LairInfo

    // a memory of every enemy city this wizard has scouted (explored the tile
    // of). Persists across turns so the AI can mount an offensive toward a city
    // it discovered earlier even after losing sight of it. Entries are forgotten
    // once we can see the tile and the enemy city is no longer there (razed or
    // captured).
    KnownEnemyCities map[CityKey]*EnemyCityInfo

    // a snapshot of the goals/decisions computed on the most recent Update,
    // used by the watch-mode debug overlay (press D). Not used for gameplay.
    LastGoalDebug []GoalDebugInfo

    // a human-readable description of what each stack was last told to do this
    // turn (e.g. "attacking enemy stack", "scouting a lair", "settling a city").
    // Rebuilt every Update; used only by the watch-mode inspector. Not gameplay.
    StackActivity map[*playerlib.UnitStack]string

    // cities that GoalBuildCities has reserved for settler production this turn.
    // GoalDefendCities runs after GoalBuildCities and the engine applies
    // AIProduceDecision with a last-wins override (game.go), so without this
    // reservation a defender build would clobber the just-queued settler and the
    // empire would never actually finish a settler. Rebuilt every Update.
    reservedSettlerCities map[*citylib.City]bool
}

func MakeEnemy2AI() *Enemy2AI {
    return &Enemy2AI{
        KnownLairs: make(map[LairKey]*LairInfo),
        KnownEnemyCities: make(map[CityKey]*EnemyCityInfo),
        StackActivity: make(map[*playerlib.UnitStack]string),
    }
}

// setActivity records what a stack is currently being directed to do, for the
// watch-mode inspector. Purely informational.
func (ai *Enemy2AI) setActivity(stack *playerlib.UnitStack, activity string) {
    if ai.StackActivity == nil {
        ai.StackActivity = make(map[*playerlib.UnitStack]string)
    }
    ai.StackActivity[stack] = activity
}

// StackActivityFor returns a human-readable description of what the given stack
// is currently doing (as last decided by the AI this turn), or "" if unknown.
func (ai *Enemy2AI) StackActivityFor(stack *playerlib.UnitStack) string {
    if ai.StackActivity == nil {
        return ""
    }
    return ai.StackActivity[stack]
}

// describeTravelIntent classifies why a stack is moving toward (destX,destY),
// for the watch-mode inspector. When a stack is just resuming a path the goal
// that originally set it isn't re-run, so we reconstruct the intent from what
// the AI knows about the destination tile (remembered enemy cities, catalogued
// lairs/nodes/towers, or unexplored fog). Purely informational.
func (ai *Enemy2AI) describeTravelIntent(self *playerlib.Player, stack *playerlib.UnitStack, destX int, destY int) string {
    plane := stack.Plane()

    // heading for a city we remember belonging to an enemy?
    if _, ok := ai.KnownEnemyCities[CityKey{X: destX, Y: destY, Plane: plane}]; ok {
        return fmt.Sprintf("marching on enemy city at (%d,%d)", destX, destY)
    }

    // heading for a known lair / node / plane tower?
    if lair, ok := ai.KnownLairs[LairKey{X: destX, Y: destY, Plane: plane}]; ok {
        if lair.Type == maplib.EncounterTypePlaneTower {
            return fmt.Sprintf("traveling to plane tower at (%d,%d)", destX, destY)
        }
        if lair.Scouted {
            return fmt.Sprintf("moving to clear a lair at (%d,%d)", destX, destY)
        }
        return fmt.Sprintf("scouting a lair at (%d,%d)", destX, destY)
    }

    // heading into the fog to reveal new territory?
    if !self.IsExplored(destX, destY, plane) {
        return fmt.Sprintf("scouting unexplored territory near (%d,%d)", destX, destY)
    }

    return fmt.Sprintf("repositioning to (%d,%d)", destX, destY)
}

// LairKey uniquely identifies a discovered lair/encounter by location.
type LairKey struct {
    X, Y int
    Plane data.Plane
}

// LairInfo is the AI's catalogued knowledge about one lair / encounter / node.
type LairInfo struct {
    X, Y int
    Plane data.Plane
    Type maplib.EncounterType
    // Scouted is true once the AI has actually seen the contents (computed a
    // Strength). Before that we only know a lair exists at this location.
    Scouted bool
    // Strength is the combat strength of the defending monsters, valid only
    // when Scouted is true.
    Strength int
}

// CityKey uniquely identifies a remembered enemy city by location.
type CityKey struct {
    X, Y int
    Plane data.Plane
}

// EnemyCityInfo is the AI's remembered knowledge about an enemy city it has
// scouted, so it can march on it later even after losing sight of the tile.
type EnemyCityInfo struct {
    X, Y int
    Plane data.Plane
}

// margin by which a stack's combat strength must exceed a lair's defenders
// before the AI commits to attacking it. The strength estimate ignores some
// factors (combat spells, terrain, exact ability matchups) so we keep a safety
// buffer rather than attacking on a razor-thin advantage. Raised from 1.25 by a
// further 20% because the AI was losing raids too easily.
const lairAttackMargin = 1.5

// requiredLairMargin scales the combat-strength surplus the AI demands before
// committing to a lair/node, based on how strong the defenders are. Weak lairs
// can be cleared opportunistically on a modest advantage, but strong lairs are
// only worth attacking when the AI has "plenty of spare army" — a big surplus —
// because losing an elite stack to a tough lair is far costlier than skipping
// it. Because a lair's recorded strength is recomputed from its current
// defenders whenever the AI can see it, partially-cleared lairs naturally fall
// into a lower band over time and become attractive targets again.
func requiredLairMargin(lairStrength int) float32 {
    switch {
    case lairStrength <= 0:
        return lairAttackMargin
    case lairStrength < 150:
        return 1.25
    case lairStrength < 400:
        return 1.5
    case lairStrength < 800:
        return 2.0
    default:
        return 2.5
    }
}

// GoalDebugInfo is a read-only snapshot of one goal node (and its subgoals)
// produced during a single Update, for display in the watch-mode debug menu.
type GoalDebugInfo struct {
    Goal GoalType
    Weight float32
    // true if this goal was skipped because the same goal type was already
    // processed this turn (the seenGoals dedup)
    Skipped bool
    // human-readable summaries of the decisions this goal produced (not including subgoals)
    Decisions []string
    SubGoals []GoalDebugInfo
}

type GoalType int

const (
    GoalNone GoalType = iota
    GoalDefeatEnemies // defeat enemy wizards
    GoalBuildCities
    GoalExploreTerritory // explore the map
    GoalResearchMagic // research spells
    GoalCastSpellOfMastery
    GoalDefendCities
    GoalBuildArmy
    GoalIncreasePower
    GoalMeldNodes
    GoalConnectCities // build engineers and roads to connect our cities
    GoalEnchantUnits // cast beneficial unit enchantments (scout mobility + elite buffs)
    GoalPlanarTravel // use open plane towers to expand onto the other plane
)

func (goal GoalType) String() string {
    switch goal {
        case GoalNone: return "None"
        case GoalDefeatEnemies: return "Defeat Enemies"
        case GoalBuildCities: return "Build Cities"
        case GoalExploreTerritory: return "Explore Territory"
        case GoalResearchMagic: return "Research Magic"
        case GoalCastSpellOfMastery: return "Cast Spell of Mastery"
        case GoalDefendCities: return "Defend Cities"
        case GoalBuildArmy: return "Build Army"
        case GoalIncreasePower: return "Increase Power"
        case GoalMeldNodes: return "Meld Nodes"
        case GoalConnectCities: return "Connect Cities"
        case GoalEnchantUnits: return "Enchant Units"
        case GoalPlanarTravel: return "Planar Travel"
    }
    return fmt.Sprintf("Goal(%d)", int(goal))
}

// describeDecision returns a short human-readable summary of an AI decision,
// for the watch-mode debug overlay.
func describeDecision(decision playerlib.AIDecision, buildingInfo buildinglib.BuildingInfos) string {
    switch d := decision.(type) {
        case *playerlib.AIProduceDecision:
            if !d.Unit.IsNone() {
                return fmt.Sprintf("%v: produce %v", d.City.Name, d.Unit.Name)
            }
            return fmt.Sprintf("%v: build %v", d.City.Name, buildingInfo.Name(d.Building))
        case *playerlib.AIMoveStackDecision:
            dest := ""
            if len(d.Path) > 0 {
                last := d.Path[len(d.Path)-1]
                dest = fmt.Sprintf(" -> (%v,%v)", last.X, last.Y)
            }
            return fmt.Sprintf("move stack (%v,%v)%v", d.Stack.X(), d.Stack.Y(), dest)
        case *playerlib.AIBuildOutpostDecision:
            return fmt.Sprintf("build outpost (%v,%v)", d.Stack.X(), d.Stack.Y())
        case *playerlib.AIBuildRoadDecision:
            return fmt.Sprintf("build road (%v,%v) -> (%v,%v)", d.Stack.X(), d.Stack.Y(), d.X, d.Y)
        case *playerlib.AIMeldNodeDecision:
            return fmt.Sprintf("meld node (%v,%v)", d.Stack.X(), d.Stack.Y())
        case *playerlib.AIPlaneShiftDecision:
            return fmt.Sprintf("plane shift (%v,%v)", d.Stack.X(), d.Stack.Y())
        case *playerlib.AICastSpellDecision:
            return fmt.Sprintf("cast %v", d.Spell.Name)
        case *playerlib.AICastUnitSpellDecision:
            return fmt.Sprintf("cast %v on %v", d.Spell.Name, d.Target.GetName())
        case *playerlib.AIResearchSpellDecision:
            return fmt.Sprintf("research %v", d.Spell.Name)
    }
    return fmt.Sprintf("%T", decision)
}

type EnemyGoal struct {
    Goal GoalType
    Weight float32 // normalized between 0 and 1

    // subgoals that must be satisfied for the current goal before an action
    // can be taken towards the current goal
    SubGoals []EnemyGoal
}

// weight of this goal, higher weight means higher priority
func (goal *EnemyGoal) GetWeight() float32 {
    return goal.Weight
}

func chance(percent int) bool {
    return rand.N(100) < percent
}

func (ai *Enemy2AI) ComputeGoals(self *playerlib.Player, aiServices playerlib.AIServices) []EnemyGoal {
    var goals []EnemyGoal

    exploreGoal := EnemyGoal{
        Goal: GoalExploreTerritory,
        Weight: 0.4,
    }

    buildArmyGoal := EnemyGoal{
        Goal: GoalBuildArmy,
        Weight: 0.5,
    }

    // only build an army if we have few stacks compared to the turn number, so
    // the standing-army target grows as the game goes on (more stacks = more
    // aggression / defense capacity).
    armyTarget := max(5, aiServices.GetTurnNumber() / 5)
    if armyTarget < 2 {
        armyTarget = 2
    }
    if uint64(len(self.Stacks)) < armyTarget {
        exploreGoal.SubGoals = append(exploreGoal.SubGoals, buildArmyGoal)
    }

    // base goal priorities. These weights surface in the watch-mode debug overlay
    // so the AI's current disposition (aggressive vs. builder vs. turtle) is
    // visible while spectating.
    goals = []EnemyGoal{
        EnemyGoal{
            Goal: GoalDefeatEnemies,
            Weight: 0.9,
            SubGoals: []EnemyGoal{
                exploreGoal,
            },
        },
        EnemyGoal{
            Goal: GoalBuildCities,
            Weight: 0.7,
        },
        EnemyGoal{
            Goal: GoalIncreasePower,
            Weight: 0.6,
        },
        EnemyGoal{
            Goal: GoalDefendCities,
            Weight: 0.8,
        },
    }

    // once we have a small empire, dedicate some effort to connecting cities
    // with roads (1 engineer per 3 cities) for faster troop movement and trade
    if len(self.Cities) >= 3 {
        goals = append(goals, EnemyGoal{
            Goal: GoalConnectCities,
            Weight: 0.5,
        })
    }

    // capture magic nodes with spirits for extra magic power. The goal self-guards
    // (does nothing unless we know a spirit spell and there are reachable, cleared,
    // unowned nodes), so it is always available once we have a city to summon from.
    if knowsSpiritSpell(self) {
        goals = append(goals, EnemyGoal{
            Goal: GoalMeldNodes,
            Weight: 0.55,
        })
    }

    // when we know beneficial unit enchantments and have spare mana, buff our
    // scouts (mobility) and best stacks (combat power). Placed last and gated on
    // spare mana so it never steals the casting channel from summoning/expansion.
    if knowsUnitEnchantment(self) {
        goals = append(goals, EnemyGoal{
            Goal: GoalEnchantUnits,
            Weight: 0.45,
        })
    }

    // ferry spare stacks through open plane towers to gain a foothold on the
    // other plane. Self-guards: only acts when we have spare forces, a city on
    // the source plane, and no established presence on the destination plane.
    goals = append(goals, EnemyGoal{
        Goal: GoalPlanarTravel,
        Weight: 0.5,
    })

    // goals = append(goals, exploreGoal)

    /*
    goals = append(goals, EnemyGoal{
        Goal: GoalDefeatEnemies,
        SubGoals: []EnemyGoal{
            exploreGoal,
            EnemyGoal{
                Goal: GoalBuildArmy,
            },
        },
    })

    goals = append(goals, EnemyGoal{
        Goal: GoalCastSpellOfMastery,
        SubGoals: []EnemyGoal{
            EnemyGoal{
                Goal: GoalResearchMagic,
            },
            EnemyGoal{
                Goal: GoalIncreasePower,
                SubGoals: []EnemyGoal{
                    EnemyGoal{
                        Goal: GoalMeldNodes,
                    },
                },
            },
        },
    })
    */

    return goals
}

// stop producing that unit
func (ai *Enemy2AI) ProducedUnit(city *citylib.City, player *playerlib.Player) {
    city.ProducingBuilding = buildinglib.BuildingTradeGoods
    city.ProducingUnit = units.UnitNone
}

type AIData struct {
    FoodPerTurn func() int
    GoldPerTurn func() int
}

func rawUnitAttackPower(units_... units.Unit) int {
    total := 0
    for _, unit := range units_ {
        overworld := units.MakeOverworldUnit(unit, 0, 0, data.PlaneArcanus)
        total += unitAttackPower(overworld)
    }
    return total
}

func unitAttackPower(unit ...units.StackUnit) int {
    total := 0
    for _, u := range unit {
        meleePower := float32(u.GetMeleeAttackPower()) * float32(u.GetToHitMelee()) / 100
        rangedPower := float32(u.GetRangedAttackPower()) * 0.3
        if u.GetRangedAttackDamageType() == units.DamageRangedMagical {
            rangedPower *= 1.5
        }

        best := max(meleePower, rangedPower)
        // to-hit can shrink a 1-attack unit below 1; keep a floor so they
        // still count as combatants. True 0-attack units (settlers) stay 0.
        power := 0
        if best > 0 {
            power = int(max(1, best)) * u.GetCount()
        }

        total += power
    }

    return total
}

func stackAttackPower(stack *playerlib.UnitStack) int {
    return unitAttackPower(stack.Units()...)
}

// combatProfile is the set of combat-relevant numbers used to estimate a
// unit's fighting value. It is extracted uniformly from any units.StackUnit so
// that the defenders of a lair and our own units are scored by the exact same
// model.
type combatProfile struct {
    count int

    meleeAttack   int
    rangedAttack  int
    rangedAttacks int // ammo: number of ranged shots available
    rangedMagical bool

    toHit    int // percent chance per attack point to land a hit
    toDefend int // percent chance per defense point to block a hit
    defense  int
    hitPoints  int
    resistance int

    firstStrike    bool
    armorPiercing  bool
    lifeSteal      bool
    weaponImmunity bool
    missileImmunity bool
    magicImmunity  bool
    largeShield    bool
    regeneration   bool
    invisibility   bool

    // expected extra per-figure damage from special attacks (breath, immolation, thrown, poison)
    bonusAttack float64
    // gaze / touch abilities that can destroy figures regardless of hit points
    deadlyTouch bool
}

// unitCombatProfile reads a StackUnit into a combatProfile.
func unitCombatProfile(unit units.StackUnit) combatProfile {
    p := combatProfile{
        count:           unit.GetCount(),
        meleeAttack:     unit.GetMeleeAttackPower(),
        rangedAttack:    unit.GetRangedAttackPower(),
        rangedAttacks:   unit.GetRangedAttacks(),
        rangedMagical:   unit.GetRangedAttackDamageType() == units.DamageRangedMagical,
        toHit:           unit.GetToHitMelee(),
        toDefend:        unit.GetToDefend(),
        defense:         unit.GetDefense(),
        hitPoints:       unit.GetHitPoints(),
        resistance:      unit.GetResistance(),
        firstStrike:     unit.HasAbility(data.AbilityFirstStrike),
        armorPiercing:   unit.HasAbility(data.AbilityArmorPiercing),
        lifeSteal:       unit.HasAbility(data.AbilityLifeSteal),
        weaponImmunity:  unit.HasAbility(data.AbilityWeaponImmunity),
        missileImmunity: unit.HasAbility(data.AbilityMissileImmunity),
        magicImmunity:   unit.HasAbility(data.AbilityMagicImmunity),
        largeShield:     unit.HasAbility(data.AbilityLargeShield),
        regeneration:    unit.HasAbility(data.AbilityRegeneration),
        invisibility:    unit.HasAbility(data.AbilityInvisibility),
    }

    if p.toHit <= 0 {
        p.toHit = 30
    }
    if p.toDefend <= 0 {
        p.toDefend = 30
    }

    // special offensive abilities add expected damage per figure
    p.bonusAttack += float64(unit.GetAbilityValue(data.AbilityFireBreath))
    p.bonusAttack += float64(unit.GetAbilityValue(data.AbilityLightningBreath))
    p.bonusAttack += float64(unit.GetAbilityValue(data.AbilityThrown))
    p.bonusAttack += float64(unit.GetAbilityValue(data.AbilityPoisonTouch))
    if unit.HasAbility(data.AbilityImmolation) {
        p.bonusAttack += 4
    }

    if unit.HasAbility(data.AbilityDeathGaze) || unit.HasAbility(data.AbilityStoningGaze) ||
        unit.HasAbility(data.AbilityDoomGaze) || unit.HasAbility(data.AbilityDeathTouch) ||
        unit.HasAbility(data.AbilityStoningTouch) {
        p.deadlyTouch = true
    }

    return p
}

// profileStrength turns a combatProfile into a single fighting value. The value
// combines expected damage output (offense) with effective hit points
// (survivability) so that a unit which both hits hard and lives long is worth
// disproportionately more than one that only does one of those.
func profileStrength(p combatProfile) float64 {
    if p.count <= 0 {
        return 0
    }
    offense, effHP, _ := profileStrengthBreakdown(p)
    return offense * effHP * float64(p.count)
}

// profileStrengthBreakdown computes the per-figure offense and effective-HP
// factors used by profileStrength, and (optionally) returns human-readable
// lines explaining each step. profileStrength multiplies offense * effHP *
// count. Keeping the math here ensures the explanation never drifts from the
// number actually used by the AI.
func profileStrengthBreakdown(p combatProfile) (offense float64, effHP float64, lines []string) {
    // --- offense: expected damage output per figure per round ---
    pHit := clampFloat(float64(p.toHit)/100, 0.1, 0.9)

    melee := float64(p.meleeAttack)
    ranged := float64(p.rangedAttack)
    if p.rangedMagical {
        ranged *= 1.5
    }
    if p.rangedAttacks <= 0 {
        ranged = 0
    }

    primary := math.Max(melee, ranged)
    // a unit with both melee and ranged is more flexible; credit the lesser one a little
    secondary := math.Min(melee, ranged) * 0.25

    rawAttack := primary + secondary + p.bonusAttack
    damage := rawAttack * pHit

    var offMults []string
    if p.firstStrike {
        damage *= 1.15
        offMults = append(offMults, "firstStrike x1.15")
    }
    if p.armorPiercing {
        damage *= 1.25
        offMults = append(offMults, "armorPiercing x1.25")
    }
    if p.lifeSteal {
        damage *= 1.2
        offMults = append(offMults, "lifeSteal x1.2")
    }
    if p.deadlyTouch {
        damage += 5 // flat per-figure kill potential, independent of enemy HP
        offMults = append(offMults, "deadlyTouch +5")
    }

    offense = math.Max(damage, 0.5) // floor so a 0-attack support unit still counts

    offLine := fmt.Sprintf("  offense %.2f = attack %.1f x toHit %d%%", offense, rawAttack, p.toHit)
    if p.bonusAttack > 0 {
        offLine += fmt.Sprintf(" (incl +%.1f special)", p.bonusAttack)
    }
    if len(offMults) > 0 {
        offLine += " x [" + strings.Join(offMults, ", ") + "]"
    }
    lines = append(lines, offLine)

    // --- survivability: effective hit points per figure ---
    pDef := clampFloat(float64(p.toDefend)/100, 0.1, 0.6)

    defense := float64(p.defense)
    if p.largeShield {
        defense += 2
    }

    // each defense point blocks pDef of an incoming hit -> raises effective HP
    effHP = float64(p.hitPoints) * (1 + defense*pDef)
    // resistance wards off gaze / poison / spell effects
    effHP *= 1 + float64(p.resistance)*0.015

    var hpMults []string
    if p.resistance > 0 {
        hpMults = append(hpMults, fmt.Sprintf("resist %d x%.3f", p.resistance, 1+float64(p.resistance)*0.015))
    }
    if p.weaponImmunity {
        effHP *= 1.8 // effectively defense 10 against normal weapons
        hpMults = append(hpMults, "weaponImm x1.8")
    }
    if p.magicImmunity {
        effHP *= 1.4
        hpMults = append(hpMults, "magicImm x1.4")
    }
    if p.regeneration {
        effHP *= 1.3
        hpMults = append(hpMults, "regen x1.3")
    }
    if p.invisibility {
        effHP *= 1.2
        hpMults = append(hpMults, "invis x1.2")
    }
    if p.missileImmunity {
        effHP *= 1.1
        hpMults = append(hpMults, "missileImm x1.1")
    }

    hpLine := fmt.Sprintf("  effHP %.2f = hp %d x (1 + def %.0f x toDef %d%%)", effHP, p.hitPoints, defense, p.toDefend)
    if len(hpMults) > 0 {
        hpLine += " x [" + strings.Join(hpMults, ", ") + "]"
    }
    lines = append(lines, hpLine)

    return offense, effHP, lines
}

// ExplainUnitStrength returns a fielded unit's combat-strength score along with
// human-readable lines explaining how the score was derived. For the watch-mode
// inspector.
func ExplainUnitStrength(unit units.StackUnit) (int, []string) {
    p := unitCombatProfile(unit)
    offense, effHP, lines := profileStrengthBreakdown(p)
    total := 0.0
    if p.count > 0 {
        total = offense * effHP * float64(p.count)
    }
    lines = append(lines, fmt.Sprintf("  strength = offense %.2f x effHP %.2f x count %d = %d",
        offense, effHP, p.count, int(total)))
    return int(total), lines
}

// ExplainUnitDataStrength scores a raw lair defender (units.Unit) and explains
// the calculation, wrapping it exactly as combat does.
func ExplainUnitDataStrength(unit units.Unit) (int, []string) {
    wrapped := units.MakeOverworldUnit(unit, 0, 0, data.PlaneArcanus)
    return ExplainUnitStrength(wrapped)
}

// LairAttackMargin is the exported safety margin a stack's strength must exceed
// a lair's defender strength by before the AI commits to attacking.
const LairAttackMargin = lairAttackMargin

func clampFloat(value, low, high float64) float64 {
    if value < low {
        return low
    }
    if value > high {
        return high
    }
    return value
}

// unitDataStrength scores the defenders found in a lair. They are raw
// units.Unit values, so we wrap each one exactly as combat does
// (MakeOverworldUnit) to expose the full StackUnit interface and score it with
// the same model as our own units.
func unitDataStrength(unit units.Unit) int {
    wrapped := units.MakeOverworldUnit(unit, 0, 0, data.PlaneArcanus)
    return int(profileStrength(unitCombatProfile(wrapped)))
}

// stackUnitStrength scores a single fielded unit (units.StackUnit).
func stackUnitStrength(unit units.StackUnit) int {
    return int(profileStrength(unitCombatProfile(unit)))
}

// StackCombatStrength is an exported wrapper around the AI's combat-strength
// estimate for a whole stack, for use by the watch-mode inspector UI.
func StackCombatStrength(stack *playerlib.UnitStack) int {
    return stackCombatStrength(stack)
}

// UnitCombatStrength is an exported wrapper around the AI's combat-strength
// estimate for a single unit, for use by the watch-mode inspector UI.
func UnitCombatStrength(unit units.StackUnit) int {
    return stackUnitStrength(unit)
}

// lairStrength is the total combat strength of an encounter's defenders.
func lairStrength(encounter *maplib.ExtraEncounter) int {
    if encounter == nil {
        return 0
    }
    total := 0
    for _, unit := range encounter.Units {
        total += unitDataStrength(unit)
    }
    return total
}

// LairStrength is an exported wrapper around lairStrength for the watch-mode
// inspector UI.
func LairStrength(encounter *maplib.ExtraEncounter) int {
    return lairStrength(encounter)
}

// combatUnits returns the units in the stack that can actually fight (nonzero
// attack power). Settlers and other support units are excluded so the AI never
// drags them into an attack on a lair or enemy.
func combatUnits(stack *playerlib.UnitStack) []units.StackUnit {
    var out []units.StackUnit
    for _, unit := range stack.Units() {
        if unitAttackPower(unit) > 0 {
            out = append(out, unit)
        }
    }
    return out
}

// unitsCombatStrength is the total combat strength of a slice of units.
func unitsCombatStrength(list []units.StackUnit) int {
    total := 0
    for _, unit := range list {
        total += stackUnitStrength(unit)
    }
    return total
}

// stackCombatStrength is the combat strength of a stack, counting only the
// units that can fight (so a settler-only stack scores 0 and never attacks).
func stackCombatStrength(stack *playerlib.UnitStack) int {
    return unitsCombatStrength(combatUnits(stack))
}

func absInt(value int) int {
    if value < 0 {
        return -value
    }
    return value
}

// countEngineers counts how many road-building (Construction) units the player
// owns across all stacks.
func countEngineers(self *playerlib.Player) int {
    total := 0
    for _, stack := range self.Stacks {
        for _, unit := range stack.Units() {
            if unit.HasAbility(data.AbilityConstruction) {
                total += 1
            }
        }
    }
    return total
}

// stackHasEngineer reports whether the stack contains a Construction unit.
func stackHasEngineer(stack *playerlib.UnitStack) bool {
    for _, unit := range stack.Units() {
        if unit.HasAbility(data.AbilityConstruction) {
            return true
        }
    }
    return false
}

// stackHasSettler reports whether the stack contains a settler (a unit that can
// found a city). Such stacks are owned by the BuildCities goal: they are never
// sent to attack/raid an enemy or lair, and are not dragged off to explore, so
// any escort that travels with the settler stays to protect it.
func stackHasSettler(stack *playerlib.UnitStack) bool {
    for _, unit := range stack.Units() {
        if unit.HasAbility(data.AbilityCreateOutpost) {
            return true
        }
    }
    return false
}

// stackHasMelder reports whether the stack contains a spirit (a unit with the
// Meld ability). Such stacks belong to the MeldNodes goal: they are never sent
// to attack/raid or explore, so the spirit travels to a node to capture it.
func stackHasMelder(stack *playerlib.UnitStack) bool {
    for _, unit := range stack.Units() {
        if unit.HasAbility(data.AbilityMeld) {
            return true
        }
    }
    return false
}

// melderUnits returns the meld-capable units (spirits) in a stack.
func melderUnits(stack *playerlib.UnitStack) []units.StackUnit {
    var out []units.StackUnit
    for _, unit := range stack.ActiveUnits() {
        if unit.HasAbility(data.AbilityMeld) {
            out = append(out, unit)
        }
    }
    return out
}

// countMelders counts how many meld-capable units (spirits) the player owns,
// whether idle or already travelling toward a node.
func countMelders(self *playerlib.Player) int {
    total := 0
    for _, stack := range self.Stacks {
        for _, unit := range stack.Units() {
            if unit.HasAbility(data.AbilityMeld) {
                total += 1
            }
        }
    }
    return total
}

// findSpiritSpell returns the cheapest known summoning spell whose summoned unit
// can meld a node (Magic Spirit, Guardian Spirit). Prefers the cheaper spell so
// the AI usually summons Magic Spirit (cost 30) over Guardian Spirit (cost 80).
func findSpiritSpell(self *playerlib.Player) (spellbook.Spell, bool) {
    var best spellbook.Spell
    found := false
    summoning := self.KnownSpells.GetSpellsBySection(spellbook.SectionSummoning)
    for _, spell := range summoning.Spells {
        unit := units.GetUnitByName(spell.Name)
        if unit.IsNone() || !unit.HasAbility(data.AbilityMeld) {
            continue
        }
        if !found || spell.Cost(true) < best.Cost(true) {
            best = spell
            found = true
        }
    }
    return best, found
}

// knowsSpiritSpell reports whether the wizard knows any spell that summons a
// meld-capable spirit.
func knowsSpiritSpell(self *playerlib.Player) bool {
    _, ok := findSpiritSpell(self)
    return ok
}

// knownUnitEnchantmentSpells returns the wizard's known, overland-castable spells
// that apply a beneficial (non-curse) enchantment to a single unit. These are the
// spells the AI can use to buff scouts and elite stacks on the overworld.
func knownUnitEnchantmentSpells(self *playerlib.Player) []spellbook.Spell {
    var out []spellbook.Spell
    for _, spell := range self.KnownSpells.OverlandSpells().Spells {
        if spell.GetUnitEnchantment() != data.UnitEnchantmentNone {
            out = append(out, spell)
        }
    }
    return out
}

// knowsUnitEnchantment reports whether the wizard knows any unit-enchantment spell.
func knowsUnitEnchantment(self *playerlib.Player) bool {
    return len(knownUnitEnchantmentSpells(self)) > 0
}

// findEnchantSpell returns the named spell from a list of candidate spells.
func findEnchantSpell(spells []spellbook.Spell, name string) (spellbook.Spell, bool) {
    for _, spell := range spells {
        if spell.Name == name {
            return spell, true
        }
    }
    return spellbook.Spell{}, false
}

// strongestUnit returns the highest combat-strength unit in a slice, or nil.
func strongestUnit(list []units.StackUnit) units.StackUnit {
    var best units.StackUnit
    bestStrength := -1
    for _, unit := range list {
        strength := stackUnitStrength(unit)
        if strength > bestStrength {
            bestStrength = strength
            best = unit
        }
    }
    return best
}

// aiHasCityOnPlane reports whether the wizard owns at least one city on a plane.
func aiHasCityOnPlane(self *playerlib.Player, plane data.Plane) bool {
    for _, city := range self.Cities {
        if city.Plane == plane {
            return true
        }
    }
    return false
}

// countStacksOnPlane counts the wizard's stacks currently on a plane.
func countStacksOnPlane(self *playerlib.Player, plane data.Plane) int {
    n := 0
    for _, stack := range self.Stacks {
        if stack.Plane() == plane {
            n += 1
        }
    }
    return n
}

// minCitiesBeforeOffPlaneExpansion is the empire size at which the AI starts
// pushing onto the other plane even if it could still settle more cities at
// home. Below this it concentrates on its starting plane.
const minCitiesBeforeOffPlaneExpansion = 4

// aiNoRoomToSettle reports whether the wizard can no longer find a settlable
// location reachable from any of its cities on the given plane (settler
// expansion there is exhausted). Used to decide it is time to look to the other
// plane.
func aiNoRoomToSettle(self *playerlib.Player, aiServices playerlib.AIServices, plane data.Plane) bool {
    fog := self.GetFog(plane)
    for _, city := range self.Cities {
        if city.Plane != plane {
            continue
        }
        if len(aiServices.FindSettlableLocations(city.X, city.Y, city.Plane, fog)) > 0 {
            return false
        }
    }
    return true
}

// aiReadyToExpandOffPlane reports whether the wizard is mature enough on `from`
// to start expanding onto the other plane: it is established there (has a city)
// and either commands a sizable empire (>= minCitiesBeforeOffPlaneExpansion) or
// has run out of room to settle more cities on `from`.
func aiReadyToExpandOffPlane(self *playerlib.Player, aiServices playerlib.AIServices, from data.Plane) bool {
    if !aiHasCityOnPlane(self, from) {
        return false
    }
    return len(self.Cities) >= minCitiesBeforeOffPlaneExpansion ||
        aiNoRoomToSettle(self, aiServices, from)
}

// aiWantsPlaneFoothold reports whether the wizard wants to ferry forces from
// `from` to its opposite plane: it must be ready to expand off `from` (see
// aiReadyToExpandOffPlane), have no city on the opposite plane yet, and hold
// fewer than two stacks there. The "no city on the opposite plane" rule
// structurally prevents a ferried stack from shifting straight back (its new
// plane has no city, so it fails the test).
func aiWantsPlaneFoothold(self *playerlib.Player, aiServices playerlib.AIServices, from data.Plane) bool {
    opposite := from.Opposite()
    return aiReadyToExpandOffPlane(self, aiServices, from) &&
        !aiHasCityOnPlane(self, opposite) &&
        countStacksOnPlane(self, opposite) < 2
}

// aiWantsToShift reports whether the given stack is a spare combat stack standing
// on an open plane tower whose opposite plane the wizard wants a foothold on. Used
// both to trigger the shift and to keep the explore goal from dragging the stack
// off the tower first.
func aiWantsToShift(self *playerlib.Player, aiServices playerlib.AIServices, stack *playerlib.UnitStack) bool {
    if stackHasSettler(stack) || stackHasMelder(stack) {
        return false
    }
    if stackCombatStrength(stack) <= 0 {
        return false
    }
    useMap := aiServices.GetMap(stack.Plane())
    if useMap == nil || !useMap.HasOpenTower(stack.X(), stack.Y()) {
        return false
    }
    return aiWantsPlaneFoothold(self, aiServices, stack.Plane())
}

// nearestReachableOpenTower returns the shortest path from the stack to an
// explored open plane tower on its current plane, or nil if none is reachable.
// Used to route both spare armies and settlers toward a portal to the other
// plane.
func nearestReachableOpenTower(self *playerlib.Player, aiServices playerlib.AIServices, stack *playerlib.UnitStack) pathfinding.Path {
    useMap := aiServices.GetMap(stack.Plane())
    if useMap == nil {
        return nil
    }
    fog := self.GetFog(stack.Plane())
    var best pathfinding.Path
    for _, tower := range useMap.GetOpenTowerLocations() {
        if !self.IsExplored(tower.X, tower.Y, stack.Plane()) {
            continue
        }
        path, ok := aiServices.FindPath(stack.X(), stack.Y(), tower.X, tower.Y, self, stack, fog)
        if ok && len(path) > 0 {
            if best == nil || len(path) < len(best) {
                best = path
            }
        }
    }
    return best
}

// stackOnOwnedNode reports whether the stack is currently standing on a magic
// node this wizard has melded (i.e. it is acting as that node's garrison).
func stackOnOwnedNode(self *playerlib.Player, aiServices playerlib.AIServices, stack *playerlib.UnitStack) bool {
    useMap := aiServices.GetMap(stack.Plane())
    if useMap == nil {
        return false
    }
    node := useMap.GetMagicNode(stack.X(), stack.Y())
    return node != nil && node.MeldingWizard == self
}

// spareGarrisonUnits returns the combat units a stack can spare to garrison a
// node. A loose field stack can spare all of its fighters; a stack sitting in
// one of our cities only spares the fighters above the city's minimum defense.
func spareGarrisonUnits(stack *playerlib.UnitStack, maybeCity *citylib.City) []units.StackUnit {
    combat := combatUnits(stack)
    if maybeCity == nil {
        return combat
    }

    minHomePower := getCityMinimumAttackPower(maybeCity)
    homePower := unitAttackPower(combat...)
    var spare []units.StackUnit
    for _, unit := range combat {
        if homePower-unitAttackPower(unit) < minHomePower {
            continue
        }
        spare = append(spare, unit)
        homePower -= unitAttackPower(unit)
    }
    return spare
}

// findCheapDefender picks the lowest-cost ordinary military unit a city can
// build to defend a node (skips settlers, engineers and spirits). Returns false
// if the city has nothing suitable.
func findCheapDefender(possible []units.Unit) (units.Unit, bool) {
    var best units.Unit
    found := false
    for _, unit := range possible {
        if unit.IsSettlers() || unit.HasAbility(data.AbilityConstruction) || unit.HasAbility(data.AbilityMeld) {
            continue
        }
        // must be able to fight or take a hit
        if rawUnitAttackPower(unit) == 0 && unit.Defense == 0 {
            continue
        }
        if !found || unit.ProductionCost < best.ProductionCost {
            best = unit
            found = true
        }
    }
    return best, found
}

// countSettlerPipeline returns how many settlers the empire is committed to:
// settler units already in the field plus cities currently producing one. Used
// to cap settler production so cities don't all spam settlers at once.
func countSettlerPipeline(self *playerlib.Player) int {
    count := 0
    for _, stack := range self.Stacks {
        for _, unit := range stack.Units() {
            if unit.HasAbility(data.AbilityCreateOutpost) {
                count++
            }
        }
    }
    for _, city := range self.Cities {
        if city.ProducingUnit.IsSettlers() {
            count++
        }
    }
    return count
}

// settlerTravelUnits returns the subset of a stack that should accompany a
// settler when it sets out to found a city: the settler(s) themselves plus a
// small escort of fighters for protection. If the stack is sitting in one of
// our cities, enough attack power is left behind to keep the city defended. The
// result is used as the Units subset of an AIMoveStackDecision so the rest of a
// city garrison stays home rather than marching off with the settler. The
// second return value reports whether the settler actually has a fighter escort
// (callers use this to avoid sending a settler off alone from a city).
func settlerTravelUnits(stack *playerlib.UnitStack, maybeCity *citylib.City) ([]units.StackUnit, bool) {
    const maxEscort = 2

    var settlers []units.StackUnit
    var fighters []units.StackUnit
    for _, unit := range stack.ActiveUnits() {
        if unit.HasAbility(data.AbilityCreateOutpost) {
            settlers = append(settlers, unit)
        } else if unitAttackPower(unit) > 0 {
            fighters = append(fighters, unit)
        }
    }

    // when leaving a city, keep enough behind to defend it: retain at least one
    // fighter and at least half the desired minimum attack power. A settler in
    // the wild (no city) can keep all its fighters together.
    minHomePower := 0
    if maybeCity != nil {
        minHomePower = getCityMinimumAttackPower(maybeCity) / 2
    }

    homePower := unitAttackPower(fighters...)
    var escort []units.StackUnit
    for _, fighter := range fighters {
        if len(escort) >= maxEscort {
            break
        }
        // never take the last fighter from a city garrison
        if maybeCity != nil && len(fighters)-len(escort)-1 < 1 {
            break
        }
        // keep the city defended: only borrow a fighter if the remaining
        // garrison still meets the minimum defensive power
        if homePower-unitAttackPower(fighter) < minHomePower {
            continue
        }
        escort = append(escort, fighter)
        homePower -= unitAttackPower(fighter)
    }

    return append(settlers, escort...), len(escort) > 0
}

// aiEnemiesNearby reports whether the AI can currently see any enemy unit stack
// within `radius` tiles of (x,y) on the given plane. Used to decide whether a
// settler needs to wait for an escort before leaving a city: if the surrounding
// area is clear, the settler is allowed to set out alone rather than sit idle.
func aiEnemiesNearby(self *playerlib.Player, aiServices playerlib.AIServices, x int, y int, plane data.Plane, radius int) bool {
    useMap := aiServices.GetMap(plane)
    if useMap == nil {
        return false
    }

    for _, enemyPlayer := range aiServices.GetEnemies(self) {
        for _, enemyStack := range enemyPlayer.Stacks {
            if enemyStack.Plane() != plane {
                continue
            }
            // only react to threats we can actually see; remembered/fogged
            // positions shouldn't pin a settler in town forever
            if !self.IsVisible(enemyStack.X(), enemyStack.Y(), plane) {
                continue
            }
            if useMap.TileDistance(x, y, enemyStack.X(), enemyStack.Y()) <= radius {
                return true
            }
        }
    }

    return false
}

// stackBusyBuildingRoad reports whether any engineer in the stack is already
// committed to building a road (busy or has a pending road path).
func stackBusyBuildingRoad(stack *playerlib.UnitStack) bool {
    for _, unit := range stack.Units() {
        if unit.GetBusy() == units.BusyStatusBuildRoad {
            return true
        }
        if len(unit.GetBuildRoadPath()) > 0 {
            return true
        }
    }
    return false
}

// findConstructionUnit returns the first buildable unit with the Construction
// ability (an engineer), if any.
func findConstructionUnit(possible []units.Unit) (units.Unit, bool) {
    for _, unit := range possible {
        if unit.HasAbility(data.AbilityConstruction) {
            return unit, true
        }
    }
    return units.UnitNone, false
}

// citiesRoadConnected reports whether the two points are joined by a contiguous
// chain of road tiles (city tiles act as road junctions). Used to avoid sending
// an engineer to build a road that already exists.
func citiesRoadConnected(useMap *maplib.Map, cityPoints map[image.Point]bool, ax, ay, bx, by int) bool {
    if useMap == nil {
        return false
    }
    width := useMap.Width()
    height := useMap.Height()

    start := image.Pt(useMap.WrapX(ax), ay)
    goal := image.Pt(useMap.WrapX(bx), by)
    if start == goal {
        return true
    }

    // a tile is part of the network if it has a road or is one of our cities
    isNetwork := func(x, y int) bool {
        x = useMap.WrapX(x)
        if y < 0 || y >= height {
            return false
        }
        if useMap.ContainsRoad(x, y) {
            return true
        }
        return cityPoints[image.Pt(x, y)]
    }

    visited := make(map[image.Point]bool)
    queue := []image.Point{start}
    visited[start] = true

    for len(queue) > 0 {
        cur := queue[0]
        queue = queue[1:]

        for dy := -1; dy <= 1; dy++ {
            for dx := -1; dx <= 1; dx++ {
                if dx == 0 && dy == 0 {
                    continue
                }
                nx := useMap.WrapX(cur.X + dx)
                ny := cur.Y + dy
                if ny < 0 || ny >= height || nx < 0 || nx >= width {
                    continue
                }
                next := image.Pt(nx, ny)
                if visited[next] {
                    continue
                }
                if next == goal {
                    return true
                }
                if isNetwork(nx, ny) {
                    visited[next] = true
                    queue = append(queue, next)
                }
            }
        }
    }

    return false
}

// updateLairCatalogue refreshes ai.KnownLairs from the map: it records every
// encounter on a tile the wizard has explored, marks an entry Scouted (and
// records the defenders' Strength) once the tile is visible, and forgets
// entries whose encounter has been cleared.
func (ai *Enemy2AI) updateLairCatalogue(self *playerlib.Player, aiServices playerlib.AIServices) {
    if ai.KnownLairs == nil {
        ai.KnownLairs = make(map[LairKey]*LairInfo)
    }

    for _, plane := range []data.Plane{data.PlaneArcanus, data.PlaneMyrror} {
        useMap := aiServices.GetMap(plane)
        if useMap == nil {
            continue
        }

        present := set.MakeSet[LairKey]()

        for _, point := range useMap.GetEncounterLocations() {
            // we only learn an encounter exists once we've explored its tile
            if !self.IsExplored(point.X, point.Y, plane) {
                continue
            }

            key := LairKey{X: point.X, Y: point.Y, Plane: plane}
            present.Insert(key)

            encounter := useMap.GetEncounter(point.X, point.Y)
            if encounter == nil {
                continue
            }

            info, ok := ai.KnownLairs[key]
            if !ok {
                info = &LairInfo{X: point.X, Y: point.Y, Plane: plane}
                ai.KnownLairs[key] = info
            }
            info.Type = encounter.Type

            // we can read the defenders only when we can currently see the tile
            if self.IsVisible(point.X, point.Y, plane) {
                info.Strength = lairStrength(encounter)
                info.Scouted = true
            }
        }

        // forget lairs that no longer exist on tiles we've explored (cleared)
        for key := range ai.KnownLairs {
            if key.Plane != plane || present.Contains(key) {
                continue
            }
            if self.IsExplored(key.X, key.Y, plane) {
                delete(ai.KnownLairs, key)
            }
        }
    }
}

// updateScoutMemory refreshes ai.KnownEnemyCities from what the wizard can see:
// it remembers every enemy city on a tile we have explored, and forgets a
// remembered city once we can see its tile and the enemy city is gone (razed,
// captured, or it was ours). This lets the AI plan offensives toward cities it
// scouted earlier even after the tile falls back under fog.
func (ai *Enemy2AI) updateScoutMemory(self *playerlib.Player, aiServices playerlib.AIServices) {
    if ai.KnownEnemyCities == nil {
        ai.KnownEnemyCities = make(map[CityKey]*EnemyCityInfo)
    }

    // record enemy cities on tiles we have explored
    for _, enemy := range aiServices.GetEnemies(self) {
        for _, city := range enemy.Cities {
            if !self.IsExplored(city.X, city.Y, city.Plane) {
                continue
            }
            key := CityKey{X: city.X, Y: city.Y, Plane: city.Plane}
            ai.KnownEnemyCities[key] = &EnemyCityInfo{X: city.X, Y: city.Y, Plane: city.Plane}
        }
    }

    // forget remembered cities we can currently see are no longer an enemy's
    for key := range ai.KnownEnemyCities {
        if !self.IsVisible(key.X, key.Y, key.Plane) {
            continue
        }
        city, owner := aiServices.FindCity(key.X, key.Y, key.Plane)
        if city == nil || owner == self {
            delete(ai.KnownEnemyCities, key)
        }
    }
}

// findScoutPath returns the shortest path that sends the given stack toward the
// nearest discovered-but-unscouted lair on its plane, so the AI can reveal what
// is guarding it. Returns nil if there is nothing to scout.
func (ai *Enemy2AI) findScoutPath(self *playerlib.Player, aiServices playerlib.AIServices, stack *playerlib.UnitStack) pathfinding.Path {
    var shortestPath pathfinding.Path
    fog := self.GetFog(stack.Plane())
    for _, lair := range ai.KnownLairs {
        if lair.Scouted || lair.Plane != stack.Plane() {
            continue
        }
        path, ok := aiServices.FindPath(stack.X(), stack.Y(), lair.X, lair.Y, self, stack, fog)
        if ok && len(path) > 0 {
            if len(shortestPath) == 0 || len(path) < len(shortestPath) {
                shortestPath = path
            }
        }
    }
    return shortestPath
}

// findNavalExplorePath returns a path that sends a ship out toward the unknown
// sea so triremes leave port and scout instead of idling. The generic explore
// logic samples mostly-land destinations that a sailing unit can never reach,
// so naval stacks need their own water-only target search. It collects nearby
// water tiles that are either unexplored or on the frontier (explored water
// adjacent to fog), prefers the closest, and returns the first reachable one.
func (ai *Enemy2AI) findNavalExplorePath(self *playerlib.Player, aiServices playerlib.AIServices, stack *playerlib.UnitStack) pathfinding.Path {
    useMap := aiServices.GetMap(stack.Plane())
    if useMap == nil {
        return nil
    }
    fog := self.GetFog(stack.Plane())
    plane := stack.Plane()
    sx, sy := stack.X(), stack.Y()

    type cand struct {
        x, y, dist int
    }
    var candidates []cand

    const radius = 12
    for dy := -radius; dy <= radius; dy++ {
        for dx := -radius; dx <= radius; dx++ {
            if dx == 0 && dy == 0 {
                continue
            }
            nx := useMap.WrapX(sx + dx)
            ny := sy + dy
            tile := useMap.GetTile(nx, ny)
            if !tile.Valid() || !useMap.IsWater(nx, ny) {
                continue
            }

            // target unexplored water directly, or explored water sitting next
            // to a fogged tile (the edge of what we've seen) so the ship keeps
            // pushing into the unknown even once its home waters are charted
            target := !self.IsExplored(nx, ny, plane)
            if !target {
                for ddy := -1; ddy <= 1 && !target; ddy++ {
                    for ddx := -1; ddx <= 1; ddx++ {
                        if !self.IsExplored(useMap.WrapX(nx+ddx), ny+ddy, plane) {
                            target = true
                            break
                        }
                    }
                }
            }
            if !target {
                continue
            }

            candidates = append(candidates, cand{x: nx, y: ny, dist: useMap.TileDistance(sx, sy, nx, ny)})
        }
    }

    slices.SortFunc(candidates, func(a, b cand) int {
        return a.dist - b.dist
    })

    tried := 0
    for _, c := range candidates {
        if tried >= 20 {
            break
        }
        tried++
        path, ok := aiServices.FindPath(sx, sy, c.x, c.y, self, stack, fog)
        if ok && len(path) > 0 {
            return path
        }
    }

    return nil
}

// return a subset of moveUnits that can move while still leaving minimumAttackPower behind
func getMoveableUnits(moveUnits []units.StackUnit, minimumAttackPower int) []units.StackUnit {

    // mostly settlers
    var weakUnits []units.StackUnit
    nonWeakUnits := make([]units.StackUnit, 0, len(moveUnits))
    for _, unit := range moveUnits {
        if unitAttackPower(unit) == 0 {
            weakUnits = append(weakUnits, unit)
        } else {
            nonWeakUnits = append(nonWeakUnits, unit)
        }
    }
    moveUnits = nonWeakUnits

    lessUnits := make([]units.StackUnit, 0, len(moveUnits))
    for i := len(moveUnits) - 1; i > 0; i-- {
        lessUnits = lessUnits[:0]

        // choose random elements from i
        for _, j := range rand.Perm(i)[:i] {
            lessUnits = append(lessUnits, moveUnits[j])
        }

        if unitAttackPower(moveUnits...) - unitAttackPower(lessUnits...) >= minimumAttackPower {
            return append(weakUnits, lessUnits...)
        }
    }

    return weakUnits
}

func getCityMinimumAttackPower(city *citylib.City) int {
    // this number is an arbitrary choice, but make sure we have enough attack power to defend against weak enemies
    const attack_power_per_citizen = 3

    // cities with more population should have more defense
    return city.Citizens() * attack_power_per_citizen
}

// cityDefenseTarget is the garrison attack power the AI actively builds toward
// for a city, as opposed to getCityMinimumAttackPower (the bare floor used when
// deciding how much to leave behind when units wander off). It is deliberately
// higher and scales with the turn number because rampaging raider packs grow
// stronger as the game goes on (see raider.go createMonsters, whose monster
// budget ramps up to ~3x by turn 200), so standing garrisons must keep pace or
// frontier towns get overrun.
func cityDefenseTarget(city *citylib.City, turn uint64) int {
    base := getCityMinimumAttackPower(city)

    // growth tracks how far past turn 50 we are, capped at turn 200
    growth := 0
    if turn > 50 {
        growth = int(turn - 50)
        if growth > 150 {
            growth = 150
        }
    }

    // 1.5x the floor early, ramping to 3x of the floor by turn 200
    scale := 1.5 + 1.5*float64(growth)/150.0
    target := int(float64(base) * scale)

    // even a brand-new one-citizen town should keep a real garrison so a single
    // wandering raider pack can't walk in unopposed
    const minimumGarrison = 24
    if target < minimumGarrison {
        target = minimumGarrison
    }

    return target
}

// the decisions to make for this goal
// seenGoals is a set used to avoid computing the same goal twice
func (ai *Enemy2AI) GoalDecisions(self *playerlib.Player, aiServices playerlib.AIServices, goal EnemyGoal, seenGoals *set.Set[GoalType], aiData *AIData) ([]playerlib.AIDecision, GoalDebugInfo) {

    info := GoalDebugInfo{Goal: goal.Goal, Weight: goal.Weight}

    // if we've seen this goal then just return nil
    if seenGoals.Contains(goal.Goal) {
        info.Skipped = true
        return nil, info
    }

    seenGoals.Insert(goal.Goal)

    var decisions []playerlib.AIDecision

    // recursively satisfy subgoals first
    for _, subGoal := range goal.SubGoals {
        subDecisions, subInfo := ai.GoalDecisions(self, aiServices, subGoal, seenGoals, aiData)
        decisions = append(decisions, subDecisions...)
        info.SubGoals = append(info.SubGoals, subInfo)
    }

    // decisions produced by this goal's own logic (excluding subgoals) start here
    ownStart := len(decisions)

    switch goal.Goal {
        case GoalDefeatEnemies:
            // find possible enemy targets
            var possibleTarget []*playerlib.UnitStack
            var possibleCities []*citylib.City
            for _, enemyPlayer := range aiServices.GetEnemies(self) {
                // FIXME: if there is a diplomatic treaty with the enemy then do not attack them

                for _, enemyStack := range enemyPlayer.Stacks {
                    if self.IsVisible(enemyStack.X(), enemyStack.Y(), enemyStack.Plane()) {
                        possibleTarget = append(possibleTarget, enemyStack)
                    }
                }

                for _, enemyCity := range enemyPlayer.Cities {
                    // in theory we can see cities that on tiles that we have explored in the past
                    if self.IsVisible(enemyCity.X, enemyCity.Y, enemyCity.Plane) {
                        possibleCities = append(possibleCities, enemyCity)
                    }
                }
            }

            if len(possibleTarget) > 0 {
                for _, stack := range self.Stacks {
                    // never march a settler (or its escort) off to attack
                    if stackHasSettler(stack) {
                        continue
                    }
                    // spirits are reserved for capturing magic nodes
                    if stackHasMelder(stack) {
                        continue
                    }

                    stackPower := stackAttackPower(stack)

                    if stackPower > 0 && stack.HasMoves() {

                        var shortestPath pathfinding.Path

                        for _, city := range possibleCities {
                            if city.Plane == stack.Plane() {
                                // just assume the city has some power in it
                                targetPower := 15
                                if stackPower > targetPower - rand.N(10) {
                                    pathToCity, ok := aiServices.FindPath(stack.X(), stack.Y(), city.X, city.Y, self, stack, self.GetFog(stack.Plane()))
                                    if ok {
                                        if len(shortestPath) == 0 || len(pathToCity) < len(shortestPath) {
                                            shortestPath = pathToCity
                                        }
                                    }
                                }
                            }
                        }

                        if len(shortestPath) == 0 {
                            for _, target := range possibleTarget {
                                if target.Plane() == stack.Plane() {

                                    targetPower := stackAttackPower(target)

                                    if stackPower > targetPower - rand.N(10) {

                                        pathToEnemy, ok := aiServices.FindPath(stack.X(), stack.Y(), target.X(), target.Y(), self, stack, self.GetFog(stack.Plane()))
                                        if ok {
                                            if len(shortestPath) == 0 || len(pathToEnemy) < len(shortestPath) {
                                                shortestPath = pathToEnemy
                                            }
                                        }
                                    }
                                }
                            }
                        }

                        // fall back to a remembered enemy city we scouted earlier
                        // but can't currently see: march a strong stack on it so
                        // the AI keeps pressing an offensive using its scouting
                        // memory. Gated on a healthy attack power so we don't send
                        // weak stacks blindly into an unknown garrison.
                        if len(shortestPath) == 0 && stackPower >= 20 {
                            for _, known := range ai.KnownEnemyCities {
                                if known.Plane != stack.Plane() {
                                    continue
                                }
                                // visible cities were already considered above
                                if self.IsVisible(known.X, known.Y, known.Plane) {
                                    continue
                                }
                                pathToCity, ok := aiServices.FindPath(stack.X(), stack.Y(), known.X, known.Y, self, stack, self.GetFog(stack.Plane()))
                                if ok {
                                    if len(shortestPath) == 0 || len(pathToCity) < len(shortestPath) {
                                        shortestPath = pathToCity
                                    }
                                }
                            }
                        }

                        if len(shortestPath) > 0 {
                            // log.Printf("AI %v moving stack at %v,%v to attack enemy via %v", self.Wizard.Name, stack.X(), stack.Y(), shortestPath)
                            // only send the fighters; leave settlers/support behind
                            decisions = append(decisions, &playerlib.AIMoveStackDecision{
                                Stack: stack,
                                Path: shortestPath,
                                Units: combatUnits(stack),
                            })

                            ai.Attacking[stack] = true
                            dest := shortestPath[len(shortestPath)-1]
                            ai.setActivity(stack, fmt.Sprintf("attacking enemy at (%d,%d)", dest.X, dest.Y))
                        }
                    }
                }
            }

            // clear out monster lairs / nodes we have scouted and can beat
            for _, stack := range self.Stacks {
                if ai.Attacking[stack] || !stack.HasMoves() {
                    continue
                }

                // never send a settler (or its escort) to raid a lair/node
                if stackHasSettler(stack) {
                    continue
                }
                // spirits are reserved for capturing magic nodes
                if stackHasMelder(stack) {
                    continue
                }

                stackStrength := stackCombatStrength(stack)
                if stackStrength <= 0 {
                    continue
                }

                var shortestPath pathfinding.Path
                for _, lair := range ai.KnownLairs {
                    if !lair.Scouted || lair.Plane != stack.Plane() {
                        continue
                    }

                    // require a strength surplus over the lair before raiding it,
                    // scaled by how dangerous the lair is.
                    margin := requiredLairMargin(lair.Strength)
                    if float32(stackStrength) < float32(lair.Strength) * margin {
                        continue
                    }

                    path, ok := aiServices.FindPath(stack.X(), stack.Y(), lair.X, lair.Y, self, stack, self.GetFog(stack.Plane()))
                    if ok && len(path) > 0 {
                        if len(shortestPath) == 0 || len(path) < len(shortestPath) {
                            shortestPath = path
                        }
                    }
                }

                if len(shortestPath) > 0 {
                    // only send the fighters; leave settlers/support behind
                    decisions = append(decisions, &playerlib.AIMoveStackDecision{
                        Stack: stack,
                        Path: shortestPath,
                        Units: combatUnits(stack),
                    })
                    ai.Attacking[stack] = true
                    dest := shortestPath[len(shortestPath)-1]
                    ai.setActivity(stack, fmt.Sprintf("clearing lair/node at (%d,%d)", dest.X, dest.Y))
                }
            }

        case GoalDefendCities:

            stacksOnContinent := functional.Memoize3(func(x int, y int, plane data.Plane) []*playerlib.UnitStack {
                return aiServices.FindStacksOnContinent(x, y, plane, self)
            })

            turn := aiServices.GetTurnNumber()

            for _, city := range self.Cities {
                // don't override a settler this empire is trying to build: either
                // one BuildCities just queued this turn, or one already in
                // progress from a previous turn. Otherwise expansion gets starved
                // because defender production would clobber it every turn.
                if ai.reservedSettlerCities[city] || city.ProducingUnit.IsSettlers() {
                    continue
                }

                garrison := city.GetGarrison()
                totalAttackPower := 0
                for _, unit := range garrison {
                    totalAttackPower += unitAttackPower(unit)
                }

                // the garrison strength we want to maintain per city
                defenseTarget := cityDefenseTarget(city, turn)
                if totalAttackPower < defenseTarget {
                    // 1. find a roaming unit that can reach this city and enter patrol
                    // 2. produce a defending unit in this city

                    var candidates []*playerlib.UnitStack
                    for _, stack := range stacksOnContinent(city.X, city.Y, city.Plane) {
                        // stack is already on a city
                        if self.FindCity(stack.X(), stack.Y(), stack.Plane()) != nil {
                            continue
                        }

                        path := stack.CurrentPath
                        /*
                        if len(path) > 0 {
                            last := path[len(path) - 1]
                            // stack is headed towards a city
                            if self.FindCity(last.X, last.Y, stack.Plane()) != nil {
                                continue
                            }
                        }
                        */
                        // ignore stacks that are going somewhere
                        if len(path) > 0 {
                            continue
                        }

                        if stackAttackPower(stack) > 0 && stack.HasMoves() {
                            candidates = append(candidates, stack)
                        }
                    }

                    if len(candidates) > 0 {
                        // choose a random stack to come save us
                        choice := candidates[rand.N(len(candidates))]

                        path, ok := aiServices.FindPath(choice.X(), choice.Y(), city.X, city.Y, self, choice, self.GetFog(choice.Plane()))
                        if ok {
                            decisions = append(decisions, &playerlib.AIMoveStackDecision{
                                Stack: choice,
                                Path: path,
                            })
                            ai.setActivity(choice, fmt.Sprintf("moving to defend %s (%d,%d)", city.Name, city.X, city.Y))
                        }
                    } else {
                        // have to build a unit
                        if aiData.FoodPerTurn() > 0 && aiData.GoldPerTurn() > 0 && self.Gold > 5 {
                            possibleUnits := city.ComputePossibleUnits()

                            possibleUnits = slices.DeleteFunc(possibleUnits, func(unit units.Unit) bool {
                                if rawUnitAttackPower(unit) == 0 {
                                    return true
                                }
                                return false
                            })

                            if len(possibleUnits) > 0 {
                                decisions = append(decisions, &playerlib.AIProduceDecision{
                                    City: city,
                                    Building: buildinglib.BuildingNone,
                                    Unit: possibleUnits[rand.N(len(possibleUnits))],
                                })
                            }

                            // if there are no possible units to build then the city is going to be undefended..
                        }
                    }
                }
            }

        case GoalBuildCities:
            // to achieve this goal the AI should move settlers towards settlable locations
            // if a settler is at a settlable location, build an outpost
            // if there are no settlers and there is enough food, produce more settlers

            for _, stack := range self.Stacks {
                if stack.HasMoves() && stack.ActiveUnitsHasAbility(data.AbilityCreateOutpost) {
                    // found a stack with a settler that is not currently moving

                    // Once the empire is mature (or out of room at home), ferry a
                    // settler onto the other plane through an open plane tower to
                    // found our first city there, instead of settling locally.
                    // aiWantsPlaneFoothold stops firing as soon as a city exists
                    // on the far plane, so this bootstraps the foothold and then
                    // hands expansion back to the normal local-settle logic.
                    if aiWantsPlaneFoothold(self, aiServices, stack.Plane()) {
                        maybeCity := self.FindCity(stack.X(), stack.Y(), stack.Plane())
                        travelUnits, hasEscort := settlerTravelUnits(stack, maybeCity)
                        // require an escort to leave a city only when enemies are nearby;
                        // otherwise let the settler ferry out alone rather than stay stuck.
                        if maybeCity == nil || hasEscort || !aiEnemiesNearby(self, aiServices, stack.X(), stack.Y(), stack.Plane(), 6) {
                            useMap := aiServices.GetMap(stack.Plane())
                            if useMap != nil && useMap.HasOpenTower(stack.X(), stack.Y()) {
                                // standing on the tower: cross to the other plane now
                                decisions = append(decisions, &playerlib.AIPlaneShiftDecision{Stack: stack})
                                ai.setActivity(stack, fmt.Sprintf("settler planar travel to %v", stack.Plane().Opposite()))
                                continue
                            }
                            if path := nearestReachableOpenTower(self, aiServices, stack); len(path) > 0 {
                                decisions = append(decisions, &playerlib.AIMoveStackDecision{
                                    Stack: stack,
                                    Path: path,
                                    Units: travelUnits,
                                })
                                dest := path[len(path)-1]
                                ai.setActivity(stack, fmt.Sprintf("settler heading to plane tower at (%d,%d)", dest.X, dest.Y))
                                continue
                            }
                        }
                    }

                    findNewPath := len(stack.CurrentPath) == 0

                    // if headed to some location, check that we can still settle there
                    if len(stack.CurrentPath) > 0 {
                        lastPoint := stack.CurrentPath[len(stack.CurrentPath) - 1]
                        if !aiServices.IsSettlableLocation(lastPoint.X, lastPoint.Y, stack.Plane()) {
                            findNewPath = true
                        }
                    }

                    if findNewPath {
                        // search through all explored locations on the current continent for settlable locations
                        fog := self.GetFog(stack.Plane())
                        locations := aiServices.FindSettlableLocations(stack.X(), stack.Y(), stack.Plane(), fog)
                        citiesOnContinent := aiServices.FindCitiesOnContinent(stack.X(), stack.Y(), stack.Plane(), self)

                        // determine if there are any cities on this continent adjacent to a shore,
                        // which means we can build a navy on this continent
                        hasShoreCity := false
                        useMap := aiServices.GetMap(stack.Plane())
                        for _, city := range citiesOnContinent {
                            if useMap.OnShore(city.X, city.Y) {
                                hasShoreCity = true
                                break
                            }
                        }

                        type PathResult struct {
                            Path pathfinding.Path
                            Ok bool
                        }

                        pathTo := functional.Memoize(func(location image.Point) PathResult {
                            path, ok := aiServices.FindPath(stack.X(), stack.Y(), location.X, location.Y, self, stack, fog)
                            return PathResult{Path: path, Ok: ok}
                        })

                        maximumPopulation := functional.Memoize(func(location image.Point) int {
                            return aiServices.ComputeMaximumPopulation(location.X, location.Y, stack.Plane())
                        })

                        // filter out all locations we cannot reach
                        locations = slices.DeleteFunc(locations, func(location image.Point) bool {
                            return pathTo(location).Ok == false
                        })

                        score := func(location image.Point) int {
                            total := maximumPopulation(location)

                            // prioritize shore locations if we don't have a city on this continent adjacent to a shore
                            if !hasShoreCity && useMap.OnShore(location.X, location.Y) {
                                total += 10
                            }

                            return total
                        }

                        slices.SortFunc(locations, func(a, b image.Point) int {
                            return cmp.Compare(score(b), score(a))
                        })

                        // log.Printf("AI possible settlable locations: %v", locations)

                        if len(locations) > 0 {
                            // search through locations and either return a decision to build an output
                            // because the stack is already at that location, or return a decision to move to that location
                            getDecision := func() (playerlib.AIDecision, bool) {
                                // if standing on a settlable location, then return immediately
                                for _, location := range locations {
                                    path := pathTo(location)

                                    if len(path.Path) == 0 {
                                        if aiServices.IsSettlableLocation(stack.X(), stack.Y(), stack.Plane()) {
                                            ai.setActivity(stack, "founding a new city (building outpost)")
                                            return &playerlib.AIBuildOutpostDecision{
                                                Stack: stack,
                                            }, true
                                        }
                                    }
                                }

                                // otherwise move towards the best settlable location.
                                // take the settler plus a small escort for protection,
                                // leaving the rest of any city garrison behind.
                                for _, location := range locations {
                                    path := pathTo(location)
                                    // log.Printf("AI moving settler stack at %v,%v to settlable location %v,%v via %v", stack.X(), stack.Y(), location.X, location.Y, path)
                                    maybeCity := self.FindCity(stack.X(), stack.Y(), stack.Plane())
                                    travelUnits, hasEscort := settlerTravelUnits(stack, maybeCity)
                                    // only hold a settler in a city for an escort when there are
                                    // enemies in the surrounding area; otherwise let it set out
                                    // alone so settlers aren't stuck inside towns indefinitely.
                                    // (when there is no escort, travelUnits is just the settler(s),
                                    // so the rest of the garrison stays home either way)
                                    if maybeCity != nil && !hasEscort && aiEnemiesNearby(self, aiServices, stack.X(), stack.Y(), stack.Plane(), 6) {
                                        ai.setActivity(stack, "settling: waiting at city for an escort (enemies nearby)")
                                        return nil, false
                                    }
                                    ai.setActivity(stack, fmt.Sprintf("settling: moving to found a city at (%d,%d)", location.X, location.Y))
                                    return &playerlib.AIMoveStackDecision{
                                        Stack: stack,
                                        Path: path.Path,
                                        Units: travelUnits,
                                    }, true
                                }

                                return nil, false
                            }

                            decision, ok := getDecision()
                            if ok {
                                decisions = append(decisions, decision)
                            }
                        }
                    }
                }
            }

            // Cap how many settlers the empire pursues at once so cities don't
            // all spam settlers simultaneously. Count settlers already in the
            // field plus those in production, and only start one new settler per
            // turn while under the cap. The cap grows with the empire so larger
            // empires keep expanding.
            maxConcurrentSettlers := 1 + len(self.Cities)/3
            if maxConcurrentSettlers < 1 {
                maxConcurrentSettlers = 1
            }
            settlerChance := 60
            if aiData.FoodPerTurn() > 0 && countSettlerPipeline(self) < maxConcurrentSettlers {
                for _, city := range self.Cities {
                    if !isMakingSomething(city) && chance(settlerChance) {
                        locations := aiServices.FindSettlableLocations(city.X, city.Y, city.Plane, self.GetFog(city.Plane))
                        if len(locations) > 0 {
                            decisions = append(decisions, &playerlib.AIProduceDecision{
                                City: city,
                                Building: buildinglib.BuildingNone,
                                Unit: units.GetSettlerUnit(city.Race),
                            })
                            // reserve this city so GoalDefendCities (which runs
                            // later and would otherwise override the production)
                            // leaves the settler build alone
                            if ai.reservedSettlerCities == nil {
                                ai.reservedSettlerCities = make(map[*citylib.City]bool)
                            }
                            ai.reservedSettlerCities[city] = true
                            // only one new settler per turn across the empire
                            break
                        }
                    }
                }
            }

        case GoalExploreTerritory:
            // in order to explore territory we need units available that are not busy

            // if there are units that are not busy and not moving, then move them to some unexplored location nearby
            // if there are no units available, then see if we can produce more units
            // if we can't sustain more units then cities should wait to grow in population via housing
            // or create more cities with settlers

            // goal 1 might want to take one action for a city, but goal 2 might want to take a conflicting action
            // choose the goal that has a higher weight

            for _, stack := range self.Stacks {
                if stack.HasMoves() {

                    // once the ConnectCities goal is active (>=3 cities) leave engineer
                    // stacks for it to dispatch as road builders instead of sending them
                    // off to scout/explore (engineers have a nominal attack power of 1
                    // and would otherwise pass the scout filter below)
                    if len(self.Cities) >= 3 && stackHasEngineer(stack) {
                        continue
                    }

                    // settler stacks belong to the BuildCities goal; don't drag a
                    // settler (or its escort) off to explore
                    if stackHasSettler(stack) {
                        continue
                    }

                    // spirit stacks belong to the MeldNodes goal; don't send them
                    // off to explore instead of capturing a node
                    if stackHasMelder(stack) {
                        continue
                    }

                    // a stack standing on one of our melded nodes is its garrison;
                    // leave it there to hold the node rather than wandering off to
                    // explore (the attack/defense goals can still reclaim it)
                    if len(stack.CurrentPath) == 0 && stackOnOwnedNode(self, aiServices, stack) {
                        continue
                    }

                    // a spare stack sitting on an open tower we intend to plane
                    // shift through belongs to the PlanarTravel goal; don't move
                    // it off the tower before it can travel
                    if len(stack.CurrentPath) == 0 && aiWantsToShift(self, aiServices, stack) {
                        continue
                    }

                    if len(stack.CurrentPath) > 0 {
                        decisions = append(decisions, &playerlib.AIMoveStackDecision{
                            Stack: stack,
                            Path: stack.CurrentPath,
                        })
                        dest := stack.CurrentPath[len(stack.CurrentPath)-1]
                        if _, exists := ai.StackActivity[stack]; !exists {
                            ai.setActivity(stack, ai.describeTravelIntent(self, stack, dest.X, dest.Y))
                        }
                        continue
                    }

                    if len(stack.CurrentPath) == 0 {
                        // naval stacks can't path over land, so the generic explore
                        // below (which samples mostly-land destinations) always fails
                        // and leaves triremes idling in port. Split any ship out of a
                        // mixed city garrison and send it to scout the open sea.
                        if stack.HasSailingUnits(true) && !stack.CanMoveOnLand(false) {
                            var shipUnits []units.StackUnit
                            for _, unit := range stack.ActiveUnits() {
                                if unit.IsSailing() && !unit.IsFlying() {
                                    shipUnits = append(shipUnits, unit)
                                }
                            }
                            if len(shipUnits) > 0 {
                                if navalPath := ai.findNavalExplorePath(self, aiServices, stack); len(navalPath) > 0 {
                                    decisions = append(decisions, &playerlib.AIMoveStackDecision{
                                        Stack: stack,
                                        Path: navalPath,
                                        Units: shipUnits,
                                    })
                                    dest := navalPath[len(navalPath)-1]
                                    ai.setActivity(stack, fmt.Sprintf("scouting the seas towards (%d,%d)", dest.X, dest.Y))
                                } else {
                                    ai.setActivity(stack, "naval: no unexplored seas reachable")
                                }
                            }
                            continue
                        }

                        moveUnits := stack.ActiveUnits()

                        useMap := aiServices.GetMap(stack.Plane())

                        // FIXME: its probably also ok if all units can fly
                        if useMap.IsWater(stack.X(), stack.Y()) {
                            moveUnits = stack.Units()
                        } else {
                            maybeCity := self.FindCity(stack.X(), stack.Y(), stack.Plane())
                            if maybeCity != nil {
                                // if this stack leaving would cause the city to be considered undefended, then stay still
                                // log.Printf("city stack before: %v", moveUnits)
                                moveUnits = getMoveableUnits(moveUnits, getCityMinimumAttackPower(maybeCity))
                                // log.Printf("city stack after: %v", moveUnits)
                            }
                        }

                        if len(moveUnits) == 0 {
                            continue
                        }

                        // if we know of a lair we haven't scouted yet, send the
                        // fighters (never settlers) to reveal its contents so the
                        // DefeatEnemies goal can later decide whether to clear it
                        var scoutUnits []units.StackUnit
                        for _, unit := range moveUnits {
                            if unitAttackPower(unit) > 0 {
                                scoutUnits = append(scoutUnits, unit)
                            }
                        }
                        if len(scoutUnits) > 0 {
                            if scoutPath := ai.findScoutPath(self, aiServices, stack); len(scoutPath) > 0 {
                                decisions = append(decisions, &playerlib.AIMoveStackDecision{
                                    Stack: stack,
                                    Path: scoutPath,
                                    Units: scoutUnits,
                                })
                                dest := scoutPath[len(scoutPath)-1]
                                ai.setActivity(stack, fmt.Sprintf("scouting a lair at (%d,%d)", dest.X, dest.Y))
                                continue
                            }
                        }

                        // split the stack in half
                        /*
                        if len(stack.Units()) > 1 && rand.N(4) == 0 {
                            stackUnits := stack.Units()
                            stack = self.SplitStack(stack, stackUnits[0:len(stackUnits) / 2])
                        }
                        */
                        // FIXME: maybe emite SplitStack decision

                        var path pathfinding.Path
                        fog := self.GetFog(stack.Plane())

                        // try upto 5 times to find a path
                        distance := 3
                        for range 5 {
                            distance += 3
                            newX, newY := stack.X() + rand.N(distance) - distance / 2, stack.Y() + rand.N(distance) - distance / 2

                            tile := useMap.GetTile(newX, newY)
                            if !tile.Valid() || self.IsExplored(newX, newY, stack.Plane()) {
                                continue
                            }

                            path, _ = aiServices.FindPath(stack.X(), stack.Y(), newX, newY, self, stack, fog)
                            if len(path) != 0 {
                                // log.Printf("Explore new location %v,%v via %v", newX, newY, path)
                                break
                            }
                        }

                        // just go somewhere random
                        if len(path) == 0 {
                            newX, newY := stack.X() + rand.N(5) - 2, stack.Y() + rand.N(5) - 2
                            path, _ = aiServices.FindPath(stack.X(), stack.Y(), newX, newY, self, stack, fog)
                        }

                        if len(path) > 0 {
                            decisions = append(decisions, &playerlib.AIMoveStackDecision{
                                Stack: stack,
                                Path: path,
                                Units: moveUnits,
                            })
                            dest := path[len(path)-1]
                            ai.setActivity(stack, fmt.Sprintf("exploring towards (%d,%d)", dest.X, dest.Y))
                        }
                    }
                }
            }

        case GoalBuildArmy:

            computeTransportUnits := functional.Memoize(func(plane data.Plane) int {
                return self.TransportUnits(plane)
            })

            for _, city := range self.Cities {
                if !isMakingSomething(city) {

                    useMap := aiServices.GetMap(city.Plane)
                    buildNavy := useMap.OnShore(city.X, city.Y) && computeTransportUnits(city.Plane) < 2

                    cityDecision := func() (playerlib.AIDecision, bool) {
                        if buildNavy && chance(50) {
                            // build buildings towards a ship building if necessary
                            // otherwise if this city can build a ship then do so
                            possibleUnits := city.ComputePossibleUnits()
                            possibleUnits = slices.DeleteFunc(possibleUnits, func(unit units.Unit) bool {
                                return !unit.HasAbility(data.AbilityTransport)
                            })

                            // FIXME: sort the transport units by their strength. prefer warship over trieme

                            if len(possibleUnits) > 0 {
                                // log.Printf("AI %v building navy unit in city %v", self.Wizard.Name, city.Name)
                                return &playerlib.AIProduceDecision{
                                    City: city,
                                    Building: buildinglib.BuildingNone,
                                    Unit: possibleUnits[rand.N(len(possibleUnits))],
                                }, true
                            } else {
                                transportBuildings := set.NewSet(
                                    buildinglib.BuildingShipYard,
                                    buildinglib.BuildingShipwrightsGuild,
                                    buildinglib.BuildingMaritimeGuild,
                                )

                                // shipyard for draconian builds an airship, which does not have transport
                                if city.Race == data.RaceDraconian {
                                    transportBuildings.Remove(buildinglib.BuildingShipYard)
                                }

                                // get full set of dependencies
                                for _, building := range transportBuildings.Values() {
                                    dependencies := city.BuildingInfo.Dependencies(building)
                                    transportBuildings.InsertMany(dependencies...)
                                }

                                // try to choose one of the transport buildings or one of its dependencies to build
                                possibleBuildings := city.ComputePossibleBuildings(true)
                                for _, building := range possibleBuildings.Values() {
                                    if transportBuildings.Contains(building) {
                                        // log.Printf("AI %v building transport building %v in city %v", self.Wizard.Name, building, city.Name)
                                        return &playerlib.AIProduceDecision{
                                            City: city,
                                            Building: building,
                                            Unit: units.UnitNone,
                                        }, true
                                    }
                                }

                            }
                        }

                        if aiData.FoodPerTurn() > 0 && aiData.GoldPerTurn() > 0 && self.Gold > 50 && chance(30) {
                            possibleUnits := city.ComputePossibleUnits()

                            possibleUnits = slices.DeleteFunc(possibleUnits, func(unit units.Unit) bool {
                                if unit.IsSettlers() {
                                    return true
                                }
                                return false
                            })

                            // bias towards stronger units
                            attacks := make([]int, 0, len(possibleUnits))
                            for _, unit := range possibleUnits {
                                attacks = append(attacks, rawUnitAttackPower(unit))
                            }

                            if len(possibleUnits) > 0 {
                                return &playerlib.AIProduceDecision{
                                    City: city,
                                    Building: buildinglib.BuildingNone,
                                    Unit: algorithm.ChoseRandomWeightedElement(possibleUnits, attacks),
                                }, true
                            }
                        }

                        return nil, false
                    }

                    decision, ok := cityDecision()
                    if ok {
                        decisions = append(decisions, decision)
                    }
                }
            }
        case GoalIncreasePower:
            type DependencyFunc func(d data.Race, building buildinglib.Building) bool
            type BuildingCheckFunc func(building buildinglib.Building) bool

            checkDependency := func(kind BuildingCheckFunc) DependencyFunc { 

                return func(race data.Race, building buildinglib.Building) bool {
                    infos := aiServices.GetBuildingInfos()

                    dependencies := set.NewSet[buildinglib.Building]()
                    for _, building := range buildinglib.RacialBuildings(race).Values() {
                        if kind(building) {
                            dependencies.InsertMany(infos.Dependencies(building)...)
                        }
                    }

                    return dependencies.Contains(building)
                }
            }

            isReligiousDependency := functional.Memoize2(checkDependency(buildinglib.Building.IsReligious))
            isEconomicDependency := functional.Memoize2(checkDependency(buildinglib.Building.IsEconomic))
            isFoodDependency := functional.Memoize2(checkDependency(buildinglib.Building.ProducesFood))

            // feels awkward to build buildings in cities here
            for _, city := range self.Cities {
                // don't clobber a settler that GoalBuildCities reserved this turn:
                // both goals target start-of-turn-idle cities and the engine
                // applies production with a last-wins override, so without this
                // guard the building queued here would replace the settler and the
                // empire would never expand (the city looks idle because the
                // settler decision hasn't been applied yet)
                if ai.reservedSettlerCities[city] {
                    continue
                }
                if !isMakingSomething(city) {
                    // create housing
                    switch {
                        case city.Citizens() < 3:
                            decisions = append(decisions, &playerlib.AIProduceDecision{
                                City: city,
                                Building: buildinglib.BuildingHousing,
                                Unit: units.UnitNone,
                            })
                        case aiData.GoldPerTurn() < 0 || self.Gold < 10 - aiData.GoldPerTurn():
                            decisions = append(decisions, &playerlib.AIProduceDecision{
                                City: city,
                                Building: buildinglib.BuildingTradeGoods,
                                Unit: units.UnitNone,
                            })
                        // always try to build something rather than sitting idle
                        default:

                            // FIXME: if unrest is high then build a shrine/temple/etc
                            // if money production is low then build a marketplace/bank/etc
                            // if food production/population growth is low then build a granary/farmers market/etc
                            // otherwise build a random building

                            possibleBuildings := city.ComputePossibleBuildings(true)
                            values := possibleBuildings.Values()

                            goldSurplus := city.GoldSurplus()

                            needsFood := city.Citizens() < city.MaximumCitySize() / 2 || city.PopulationGrowthRate() < 30

                            weights := make([]int, 0, possibleBuildings.Size())
                            for _, building := range values {
                                weight := 1

                                if city.Rebels > 0 {
                                    if building.IsReligious() {
                                        weight += 2
                                    } else if isReligiousDependency(city.Race, building) {
                                        weight += 1
                                    }
                                }

                                if goldSurplus < 0 {
                                    if building.IsEconomic() {
                                        weight += 2
                                    } else if isEconomicDependency(city.Race, building) {
                                        weight += 1
                                    }
                                }

                                if needsFood {
                                    if building.ProducesFood() {
                                        weight += 2
                                    } else if isFoodDependency(city.Race, building) {
                                        weight += 1
                                    }
                                }

                                // FIXME: also consider dependencies of buildings that produce religion/gold/food

                                weights = append(weights, weight)
                            }

                            if possibleBuildings.Size() > 0 {
                                // choose a random building to create
                                decisions = append(decisions, &playerlib.AIProduceDecision{
                                    City: city,
                                    Building: algorithm.ChoseRandomWeightedElement(values, weights),
                                    Unit: units.UnitNone,
                                })
                            }
                    }
                }
            }

        case GoalConnectCities:
            // need a few cities before roads are worth the engineer investment
            if len(self.Cities) < 3 {
                break
            }

            // budget: one engineer per three cities
            engineerBudget := len(self.Cities) / 3
            currentEngineers := countEngineers(self)

            // produce another engineer if we are under budget and can afford it
            if currentEngineers < engineerBudget && aiData.GoldPerTurn() > 0 && self.Gold > 20 {
                for _, city := range self.Cities {
                    if isMakingSomething(city) || ai.reservedSettlerCities[city] {
                        continue
                    }
                    if engineer, ok := findConstructionUnit(city.ComputePossibleUnits()); ok {
                        decisions = append(decisions, &playerlib.AIProduceDecision{
                            City: city,
                            Building: buildinglib.BuildingNone,
                            Unit: engineer,
                        })
                        break // only queue one engineer per turn
                    }
                }
            }

            // index our city tiles per plane for the road-connectivity test
            cityPointsByPlane := make(map[data.Plane]map[image.Point]bool)
            for _, city := range self.Cities {
                m := cityPointsByPlane[city.Plane]
                if m == nil {
                    m = make(map[image.Point]bool)
                    cityPointsByPlane[city.Plane] = m
                }
                m[image.Pt(city.X, city.Y)] = true
            }

            // dispatch idle engineers to connect a not-yet-connected city pair
            for _, stack := range self.Stacks {
                if !stack.HasMoves() || !stackHasEngineer(stack) {
                    continue
                }
                if stackBusyBuildingRoad(stack) || len(stack.CurrentPath) > 0 {
                    continue
                }

                useMap := aiServices.GetMap(stack.Plane())
                if useMap == nil {
                    continue
                }
                cityPoints := cityPointsByPlane[stack.Plane()]

                // home = nearest owned city to the engineer on this plane
                var home *citylib.City
                homeDist := math.MaxInt
                for _, city := range self.Cities {
                    if city.Plane != stack.Plane() {
                        continue
                    }
                    d := absInt(useMap.XDistance(city.X, stack.X())) + absInt(city.Y-stack.Y())
                    if d < homeDist {
                        homeDist = d
                        home = city
                    }
                }
                if home == nil {
                    continue
                }

                // target = nearest owned city not yet road-connected to home and
                // reachable on foot from the engineer's position
                var bestTarget *citylib.City
                bestDist := math.MaxInt
                for _, city := range self.Cities {
                    if city.Plane != stack.Plane() || city == home {
                        continue
                    }
                    if citiesRoadConnected(useMap, cityPoints, home.X, home.Y, city.X, city.Y) {
                        continue
                    }
                    path, ok := aiServices.FindPath(stack.X(), stack.Y(), city.X, city.Y, self, stack, self.GetFog(stack.Plane()))
                    if !ok || len(path) == 0 {
                        continue
                    }
                    if len(path) < bestDist {
                        bestDist = len(path)
                        bestTarget = city
                    }
                }

                if bestTarget != nil {
                    decisions = append(decisions, &playerlib.AIBuildRoadDecision{
                        Stack: stack,
                        X: bestTarget.X,
                        Y: bestTarget.Y,
                    })
                    ai.setActivity(stack, fmt.Sprintf("building road to %s (%d,%d)", bestTarget.Name, bestTarget.X, bestTarget.Y))
                }
            }

        case GoalMeldNodes:
            spiritSpell, knowSpirit := findSpiritSpell(self)
            if !knowSpirit {
                break
            }

            // gather candidate target nodes: explored, cleared of guardians (no
            // encounter), not warped, and not yet melded by anyone.
            type nodeTarget struct {
                X, Y int
                Plane data.Plane
            }
            var targets []nodeTarget
            for _, plane := range []data.Plane{data.PlaneArcanus, data.PlaneMyrror} {
                useMap := aiServices.GetMap(plane)
                if useMap == nil {
                    continue
                }
                for _, point := range useMap.GetMagicNodeLocations() {
                    if !self.IsExplored(point.X, point.Y, plane) {
                        continue
                    }
                    // a node still guarded by monsters must be cleared first (the
                    // DefeatEnemies goal handles that); a spirit can't walk onto it
                    if useMap.GetEncounter(point.X, point.Y) != nil {
                        continue
                    }
                    node := useMap.GetMagicNode(point.X, point.Y)
                    if node == nil || node.Warped || node.MeldingWizard != nil {
                        continue
                    }
                    targets = append(targets, nodeTarget{X: point.X, Y: point.Y, Plane: plane})
                }
            }

            // claimed tracks nodes a spirit is already heading to this turn so
            // two spirits don't converge on the same node
            claimed := make(map[image.Point]bool)

            // move existing spirits onto nodes, or meld the one they stand on
            for _, stack := range self.Stacks {
                if !stack.HasMoves() || !stackHasMelder(stack) {
                    continue
                }
                if len(stack.CurrentPath) > 0 {
                    continue
                }

                useMap := aiServices.GetMap(stack.Plane())
                if useMap == nil {
                    continue
                }

                // already standing on a meldable node? capture it.
                if node := useMap.GetMagicNode(stack.X(), stack.Y()); node != nil &&
                    !node.Warped && node.MeldingWizard == nil &&
                    useMap.GetEncounter(stack.X(), stack.Y()) == nil {
                    decisions = append(decisions, &playerlib.AIMeldNodeDecision{Stack: stack})
                    ai.setActivity(stack, fmt.Sprintf("melding the node at (%d,%d)", stack.X(), stack.Y()))
                    continue
                }

                // otherwise head to the nearest reachable unclaimed target node
                var bestPath pathfinding.Path
                var bestPoint image.Point
                for _, target := range targets {
                    if target.Plane != stack.Plane() {
                        continue
                    }
                    pt := image.Pt(target.X, target.Y)
                    if claimed[pt] {
                        continue
                    }
                    path, ok := aiServices.FindPath(stack.X(), stack.Y(), target.X, target.Y, self, stack, self.GetFog(stack.Plane()))
                    if !ok || len(path) == 0 {
                        continue
                    }
                    if len(bestPath) == 0 || len(path) < len(bestPath) {
                        bestPath = path
                        bestPoint = pt
                    }
                }

                if len(bestPath) > 0 {
                    claimed[bestPoint] = true
                    decisions = append(decisions, &playerlib.AIMoveStackDecision{
                        Stack: stack,
                        Path: bestPath,
                        Units: melderUnits(stack),
                    })
                    ai.setActivity(stack, fmt.Sprintf("melding: traveling to node at (%d,%d)", bestPoint.X, bestPoint.Y))
                }
            }

            // summon another spirit if there are still more open nodes than we
            // have spirits, we are not already casting, can pay the upkeep, can
            // actually cast overland, and have a city to summon into.
            if len(targets) > countMelders(self) &&
                self.CastingSpell.Invalid() &&
                self.FindSummoningCity() != nil &&
                self.ComputeOverworldCastingSkill() > 0 {

                spiritUnit := units.GetUnitByName(spiritSpell.Name)
                manaPerTurn := self.ManaPerTurn(aiServices.ComputePower(self), aiServices)
                if !spiritUnit.IsNone() && manaPerTurn >= spiritUnit.UpkeepMana {
                    decisions = append(decisions, &playerlib.AICastSpellDecision{Spell: spiritSpell})
                }
            }

            // --- keep captured nodes defended when we can afford it ---
            // Find our melded nodes that currently have no real defender.
            type garrisonTarget struct {
                X, Y int
                Plane data.Plane
            }
            var undefendedNodes []garrisonTarget
            for _, plane := range []data.Plane{data.PlaneArcanus, data.PlaneMyrror} {
                useMap := aiServices.GetMap(plane)
                if useMap == nil {
                    continue
                }
                for _, point := range useMap.GetMagicNodeLocations() {
                    node := useMap.GetMagicNode(point.X, point.Y)
                    if node == nil || node.MeldingWizard != self {
                        continue
                    }
                    if existing := self.FindStack(point.X, point.Y, plane); existing != nil && stackCombatStrength(existing) > 0 {
                        continue
                    }
                    undefendedNodes = append(undefendedNodes, garrisonTarget{X: point.X, Y: point.Y, Plane: plane})
                }
            }

            // Dispatch spare units to hold the nodes. These are existing units
            // (already paid for) and are NOT reserved here, so a higher-priority
            // goal that runs earlier next turn (mounting an attack, or defending a
            // threatened city) can reclaim them.
            claimedGarrison := make(map[*playerlib.UnitStack]bool)
            for _, target := range undefendedNodes {
                var bestStack *playerlib.UnitStack
                var bestSpare []units.StackUnit
                var bestPath pathfinding.Path
                for _, stack := range self.Stacks {
                    if claimedGarrison[stack] || ai.Attacking[stack] {
                        continue
                    }
                    if !stack.HasMoves() || len(stack.CurrentPath) > 0 || stack.Plane() != target.Plane {
                        continue
                    }
                    if stackHasSettler(stack) || stackHasMelder(stack) || stackHasEngineer(stack) {
                        continue
                    }
                    // a stack already holding one of our nodes stays put
                    if stackOnOwnedNode(self, aiServices, stack) {
                        continue
                    }

                    maybeCity := self.FindCity(stack.X(), stack.Y(), stack.Plane())
                    spare := spareGarrisonUnits(stack, maybeCity)
                    if len(spare) == 0 {
                        continue
                    }

                    path, ok := aiServices.FindPath(stack.X(), stack.Y(), target.X, target.Y, self, stack, self.GetFog(stack.Plane()))
                    if !ok || len(path) == 0 {
                        continue
                    }
                    if bestStack == nil || len(path) < len(bestPath) {
                        bestStack = stack
                        bestSpare = spare
                        bestPath = path
                    }
                }
                if bestStack != nil {
                    claimedGarrison[bestStack] = true
                    decisions = append(decisions, &playerlib.AIMoveStackDecision{
                        Stack: bestStack,
                        Path: bestPath,
                        Units: bestSpare,
                    })
                    ai.setActivity(bestStack, fmt.Sprintf("garrisoning the node at (%d,%d)", target.X, target.Y))
                }
            }

            // If nodes are still undefended and we can afford the extra upkeep,
            // train one cheap defender in the city nearest an open node.
            if len(undefendedNodes) > 0 && aiData.GoldPerTurn() > 0 && self.Gold > 20 && aiData.FoodPerTurn() > 0 {
                var bestCity *citylib.City
                bestDist := math.MaxInt
                for _, city := range self.Cities {
                    if isMakingSomething(city) || ai.reservedSettlerCities[city] {
                        continue
                    }
                    useMap := aiServices.GetMap(city.Plane)
                    if useMap == nil {
                        continue
                    }
                    for _, target := range undefendedNodes {
                        if target.Plane != city.Plane {
                            continue
                        }
                        d := absInt(useMap.XDistance(city.X, target.X)) + absInt(city.Y-target.Y)
                        if d < bestDist {
                            bestDist = d
                            bestCity = city
                        }
                    }
                }
                if bestCity != nil {
                    if defender, ok := findCheapDefender(bestCity.ComputePossibleUnits()); ok {
                        decisions = append(decisions, &playerlib.AIProduceDecision{
                            City: bestCity,
                            Building: buildinglib.BuildingNone,
                            Unit: defender,
                        })
                    }
                }
            }

        case GoalEnchantUnits:
            // Buff our own units with beneficial enchantments when we have spare
            // casting capacity and a healthy mana cushion. Only one overworld
            // spell can be queued at a time, so this emits at most one cast per
            // turn and naturally yields the casting channel to summoning/expansion
            // (those goals run earlier and set CastingSpell first).
            if !self.CastingSpell.Invalid() || self.ComputeOverworldCastingSkill() <= 0 {
                break
            }

            manaPerTurn := self.ManaPerTurn(aiServices.ComputePower(self), aiServices)
            if manaPerTurn <= 0 {
                break
            }

            enchantSpells := knownUnitEnchantmentSpells(self)
            if len(enchantSpells) == 0 {
                break
            }

            // keep a reserve so buffs never drain the mana pool that summoning
            // and city/overworld casting rely on.
            const enchantManaReserve = 40
            affordable := func(spell spellbook.Spell) bool {
                return self.Mana >= self.ComputeEffectiveSpellCost(spell, true) + enchantManaReserve
            }

            // emit a single enchant cast on a target unit; returns true if emitted.
            castOn := func(spell spellbook.Spell, target units.StackUnit) bool {
                if target == nil || !affordable(spell) {
                    return false
                }
                if target.HasEnchantment(spell.GetUnitEnchantment()) {
                    return false
                }
                decisions = append(decisions, &playerlib.AICastUnitSpellDecision{
                    Spell: spell,
                    Target: target,
                })
                if stack := self.FindStackByUnit(target); stack != nil {
                    ai.setActivity(stack, fmt.Sprintf("casting %v on %v", spell.Name, target.GetName()))
                }
                return true
            }

            cast := false

            // --- (b) scout mobility ---
            // A lone roaming scout benefits most from movement enchantments:
            // Water Walking lets it cross otherwise-impassable ocean, Path Finding
            // ignores terrain movement penalties, so it reveals far more map per
            // turn. We only target single-unit, non-city stacks because a unit
            // enchantment helps that one unit's movement, not a mixed stack's.
            mobilitySpells := []string{"Water Walking", "Path Finding", "Pathfinding"}
            for _, name := range mobilitySpells {
                if cast {
                    break
                }
                spell, ok := findEnchantSpell(enchantSpells, name)
                if !ok {
                    continue
                }
                for _, stack := range self.Stacks {
                    if ai.Attacking[stack] || stackHasSettler(stack) || stackHasMelder(stack) {
                        continue
                    }
                    // only lone scouts roaming outside a city
                    if self.FindCity(stack.X(), stack.Y(), stack.Plane()) != nil {
                        continue
                    }
                    fighters := combatUnits(stack)
                    if len(fighters) != 1 {
                        continue
                    }
                    if castOn(spell, fighters[0]) {
                        cast = true
                        break
                    }
                }
            }

            // --- (e) elite stack combat buffs ---
            // As we accumulate spare mana, reinforce our single strongest army so
            // it can take on tougher lairs and enemies. Each turn we add the most
            // impactful enchantment its lead unit is still missing, so buffs stack
            // up over several turns.
            if !cast {
                var eliteStack *playerlib.UnitStack
                eliteStrength := 0
                for _, stack := range self.Stacks {
                    if stackHasSettler(stack) || stackHasMelder(stack) {
                        continue
                    }
                    // a lone scout isn't our elite army
                    if len(combatUnits(stack)) < 2 {
                        continue
                    }
                    strength := stackCombatStrength(stack)
                    if strength > eliteStrength {
                        eliteStrength = strength
                        eliteStack = stack
                    }
                }

                if eliteStack != nil {
                    best := strongestUnit(combatUnits(eliteStack))
                    // priority order of beneficial combat enchantments
                    combatBuffs := []string{
                        "Holy Weapon", "Holy Armor", "Giant Strength", "Bless",
                        "Heroism", "Lionheart", "Iron Skin", "Endurance",
                        "Elemental Armor", "Resist Magic", "True Sight", "Regeneration",
                    }
                    for _, name := range combatBuffs {
                        spell, ok := findEnchantSpell(enchantSpells, name)
                        if !ok {
                            continue
                        }
                        if castOn(spell, best) {
                            cast = true
                            break
                        }
                    }
                }
            }

        case GoalPlanarTravel:
            // Expand to the other plane through open plane towers. The lair-clear
            // goal already attacks guarded towers; once a tower is cleared the
            // victorious stack is left standing on it, ready to travel.
            //
            // We only ferry when we have spare forces and want a foothold on the
            // far plane (see aiWantsPlaneFoothold), so this never strips our core
            // army or ping-pongs stacks back and forth.
            if len(self.Stacks) < 3 {
                break
            }

            // 1) shift any spare stack already standing on a wanted open tower
            shifted := false
            for _, stack := range self.Stacks {
                if ai.Attacking[stack] {
                    continue
                }
                if aiWantsToShift(self, aiServices, stack) {
                    decisions = append(decisions, &playerlib.AIPlaneShiftDecision{Stack: stack})
                    ai.setActivity(stack, fmt.Sprintf("planar travel to %v", stack.Plane().Opposite()))
                    shifted = true
                }
            }

            // 2) otherwise route one spare stack to the nearest explored open
            //    tower leading to a plane we want a foothold on
            if !shifted {
                for _, plane := range []data.Plane{data.PlaneArcanus, data.PlaneMyrror} {
                    if !aiWantsPlaneFoothold(self, aiServices, plane) {
                        continue
                    }
                    useMap := aiServices.GetMap(plane)
                    if useMap == nil {
                        continue
                    }

                    var towers []image.Point
                    for _, point := range useMap.GetOpenTowerLocations() {
                        if self.IsExplored(point.X, point.Y, plane) {
                            towers = append(towers, point)
                        }
                    }
                    if len(towers) == 0 {
                        continue
                    }

                    var bestStack *playerlib.UnitStack
                    var bestPath pathfinding.Path
                    for _, stack := range self.Stacks {
                        if stack.Plane() != plane || ai.Attacking[stack] {
                            continue
                        }
                        if stackHasSettler(stack) || stackHasMelder(stack) || stackHasEngineer(stack) {
                            continue
                        }
                        if stackCombatStrength(stack) <= 0 {
                            continue
                        }
                        // stacks already on a tower are handled by step 1
                        if useMap.HasOpenTower(stack.X(), stack.Y()) {
                            continue
                        }
                        // never empty a city below its minimum garrison
                        if city := self.FindCity(stack.X(), stack.Y(), plane); city != nil {
                            if len(getMoveableUnits(stack.Units(), getCityMinimumAttackPower(city))) == 0 {
                                continue
                            }
                        }
                        for _, tower := range towers {
                            path, ok := aiServices.FindPath(stack.X(), stack.Y(), tower.X, tower.Y, self, stack, self.GetFog(plane))
                            if ok && len(path) > 0 {
                                if bestStack == nil || len(path) < len(bestPath) {
                                    bestStack = stack
                                    bestPath = path
                                }
                            }
                        }
                    }

                    if bestStack != nil {
                        decisions = append(decisions, &playerlib.AIMoveStackDecision{
                            Stack: bestStack,
                            Path: bestPath,
                            Units: combatUnits(bestStack),
                        })
                        dest := bestPath[len(bestPath)-1]
                        ai.setActivity(bestStack, fmt.Sprintf("traveling to plane tower at (%d,%d)", dest.X, dest.Y))
                        break
                    }
                }
            }

        default:
            log.Printf("WARNING: unhandled goal %v", goal.Goal)
    }

    // attribute the decisions this goal produced (after its subgoals) for the debug overlay
    buildingInfo := aiServices.GetBuildingInfos()
    for _, decision := range decisions[ownStart:] {
        info.Decisions = append(info.Decisions, describeDecision(decision, buildingInfo))
    }

    return decisions, info
}

func (ai *Enemy2AI) Update(self *playerlib.Player, aiServices playerlib.AIServices) []playerlib.AIDecision {
    var decisions []playerlib.AIDecision

    // refresh our knowledge of nearby lairs / nodes / enemy cities before planning
    ai.updateLairCatalogue(self, aiServices)
    ai.updateScoutMemory(self, aiServices)

    goals := ai.ComputeGoals(self, aiServices)

    // compute these values once for all goals
    aiData := AIData{
        FoodPerTurn: functional.Memoize0(func() int {
            return self.FoodPerTurn()
        }),
        GoldPerTurn: functional.Memoize0(func() int {
            return self.GoldPerTurn()
        }),
    }

    seenGoals := set.MakeSet[GoalType]()
    ai.LastGoalDebug = nil
    // rebuild per-stack activity descriptions this turn (watch-mode inspector)
    ai.StackActivity = make(map[*playerlib.UnitStack]string)
    // clear the per-turn settler-production reservations so DefendCities can see
    // which cities BuildCities has earmarked for a settler this turn
    ai.reservedSettlerCities = make(map[*citylib.City]bool)
    for _, goal := range goals {
        goalDecisions, info := ai.GoalDecisions(self, aiServices, goal, seenGoals, &aiData)
        decisions = append(decisions, goalDecisions...)
        ai.LastGoalDebug = append(ai.LastGoalDebug, info)
    }

    if self.ResearchingSpell.Invalid() {
        if len(self.ResearchCandidateSpells.Spells) > 0 {
            // choose cheapest research cost spell
            choice := self.ResearchCandidateSpells.Spells[0]
            for _, spell := range self.ResearchCandidateSpells.Spells {
                if spell.ResearchCost < choice.ResearchCost {
                    choice = spell
                }
            }
            decisions = append(decisions, &playerlib.AIResearchSpellDecision{
                Spell: choice,
            })
        }
    }

    return decisions
}

func (ai *Enemy2AI) ConfirmEncounter(stack *playerlib.UnitStack, encounter *maplib.ExtraEncounter) bool {
    if stack == nil || encounter == nil {
        return false
    }

    // only attack a lair if our army is comfortably stronger than its defenders.
    // refusing leaves the stack one tile away, which reveals the lair's contents
    // (so a too-weak stack effectively scouts it for a future, stronger attack).
    strength := lairStrength(encounter)
    return float32(stackCombatStrength(stack)) >= float32(strength) * requiredLairMargin(strength)
}

func (ai *Enemy2AI) InvalidMove(stack *playerlib.UnitStack) {
}

func (ai *Enemy2AI) MovedStack(stack *playerlib.UnitStack, path pathfinding.Path) pathfinding.Path {
    // after moving towards an enemy, clear the current path so a new path can be computed next turn
    // maybe the enemy is no longer there, so there is no point in moving towards it
    _, ok := ai.Attacking[stack]
    if ok {
        return nil
    }

    return path
}

func (ai *Enemy2AI) PostUpdate(self *playerlib.Player, aiServices playerlib.AIServices) {

    // merge stacks that are on top of each other
    type Location struct {
        X, Y int
        Plane data.Plane
    }

    var stackLocations []Location

    for _, stack := range self.Stacks {
        stackLocations = append(stackLocations, Location{X: stack.X(), Y: stack.Y(), Plane: stack.Plane()})
    }

    for _, location := range stackLocations {
        stacks := self.FindAllStacks(location.X, location.Y, location.Plane)
        for len(stacks) > 1 {
            self.MergeStacks(stacks[0], stacks[1])
            stacks = self.FindAllStacks(location.X, location.Y, location.Plane)
        }
    }

    // make sure food is balanced at the end
    self.RebalanceFood()
}

func (ai *Enemy2AI) PreTurn(player *playerlib.Player) {
}

func (ai *Enemy2AI) NewTurn(player *playerlib.Player) {
    // maybe make this dynamic based on gold and current unrest levels
    player.TaxRate = fraction.FromInt(1)

    // make sure cities have enough farmers
    for _, city := range player.Cities {
        // try to maximize the number of workers
        city.Workers = city.Citizens()
        city.ResetCitizens()
    }

    player.RebalanceFood()

    ok := true
    for ok && player.Gold + player.GoldPerTurn() * 3 < 0 {
        var newRate fraction.Fraction
        newRate, ok = player.NextTaxRate(player.TaxRate)
        if ok {
            player.UpdateTaxRate(newRate)
        }
    }

    if player.Gold < 50 && player.Mana > 200 {
        ai.ConvertManaToGold(player)
    }

    ai.Attacking = make(map[*playerlib.UnitStack]bool)
}

func (ai *Enemy2AI) ConvertManaToGold(self *playerlib.Player) {
    manaToConvert := int(float64(self.Mana) * 0.2)

    alchemyConversion := 0.5
    if self.Wizard.RetortEnabled(data.RetortAlchemy) {
        alchemyConversion = 1
    }

    self.Mana -= manaToConvert
    self.Gold += int(float64(manaToConvert) * alchemyConversion)
}

func (ai *Enemy2AI) ConfirmRazeTown(city *citylib.City) bool {
    return false
}

func (ai *Enemy2AI) HandleMerchantItem(self *playerlib.Player, item *artifact.Artifact, cost int) bool {
    if self.Gold >= cost {
        for _, hero := range self.Heroes {
            if hero != nil && hero.Status == herolib.StatusEmployed {
                slots := hero.GetArtifactSlots()
                for i := range hero.Equipment {
                    if hero.Equipment[i] == nil && slots[i].CompatibleWith(item.Type) {
                        hero.Equipment[i] = item
                        log.Printf("AI %v bought artifact %v for %v gold, and gave it to hero %v", self.Wizard.Name, item.Name, cost, hero.Name)
                        return true
                    }
                }
            }
        }

        for i := range self.VaultEquipment {
            // FIXME: possibly replace an artifact
            if self.VaultEquipment[i] == nil {
                self.VaultEquipment[i] = item
                self.Gold -= cost
                log.Printf("AI %v bought artifact %v for %v gold, and placed it in the vault", self.Wizard.Name, item.Name, cost)
                return true
            }
        }
    }

    return false
}

func (ai *Enemy2AI) HandleHireHero(self *playerlib.Player, hero *herolib.Hero, cost int, atFortress bool, point data.PlanePoint){
    goldPerTurn := self.GoldPerTurn()

    // always try to hire a hero if we can afford it
    if self.Gold >= cost && goldPerTurn >= hero.GetUpkeepGold() {
        added := false
        if atFortress {
            added = self.AddHeroToFortress(hero)
        } else {
            added = self.AddHero(hero, point.X, point.Y, point.Plane)
        }

        if added {
            log.Printf("AI %v hired hero %v for %v gold", self.Wizard.Name, hero.Name, cost)
            self.Gold -= cost
            hero.SetStatus(herolib.StatusEmployed)

            // FIXME: consider invoking this method
            // game.ResolveStackAt(hero.GetX(), hero.GetY(), hero.GetPlane())

            if self.SelectedStack == nil {
                self.SelectedStack = self.FindStack(hero.GetX(), hero.GetY(), hero.GetPlane())
            }
        }
    }
}

func (ai *Enemy2AI) HandleHireMercenaries(self *playerlib.Player, mercenaries []*units.OverworldUnit, cost int) {
    neededGoldPerTurn := 0
    for _, mercenary := range mercenaries {
        neededGoldPerTurn += mercenary.GetUpkeepGold()
    }

    if self.Gold >= cost && self.GoldPerTurn() > neededGoldPerTurn && self.FoodPerTurn() > 0 {
        log.Printf("AI %v hired %v mercenaries for %v gold", self.Wizard.Name, len(mercenaries), cost)
        for _, unit := range mercenaries {
            self.AddUnit(unit)
            // FIXME: consider invoking this method
            // game.ResolveStackAt(unit.GetX(), unit.GetY(), unit.GetPlane())
        }
        self.Gold -= cost
    }
}
