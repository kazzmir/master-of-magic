package player

import (
    "slices"
    "math"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
)

type ActiveMap map[*units.OverworldUnit]bool

type UnitStack struct {
    units []*units.OverworldUnit
    active ActiveMap

    CurrentPath pathfinding.Path

    // non-zero while animating movement on the overworld
    offsetX float64
    offsetY float64
}

func MakeUnitStack() *UnitStack {
    return MakeUnitStackFromUnits(nil)
}

func MakeUnitStackFromUnits(units []*units.OverworldUnit) *UnitStack {
    stack := &UnitStack{
        units: units,
        active: make(ActiveMap),
    }

    for _, unit := range units {
        stack.active[unit] = true
    }

    return stack
}

func (stack *UnitStack) ResetMoves(){
    for _, unit := range stack.units {
        unit.ResetMoves()
    }
}

func (stack *UnitStack) NaturalHeal(){
    for _, unit := range stack.units {
        unit.NaturalHeal()
    }
}

func (stack *UnitStack) SetOffset(x float64, y float64) {
    stack.offsetX = x
    stack.offsetY = y
}

func (stack *UnitStack) OffsetX() float64 {
    return stack.offsetX
}

func (stack *UnitStack) OffsetY() float64 {
    return stack.offsetY
}

func (stack *UnitStack) IsEmpty() bool {
    return len(stack.units) == 0
}

func (stack *UnitStack) Units() []*units.OverworldUnit {
    return slices.Clone(stack.units)
}

func (stack *UnitStack) ActiveUnits() []*units.OverworldUnit {
    var out []*units.OverworldUnit
    for unit, active := range stack.active {
        if active {
            out = append(out, unit)
        }
    }

    return out
}

func (stack *UnitStack) InactiveUnits() []*units.OverworldUnit {
    var inactive []*units.OverworldUnit
    for unit, active := range stack.active {
        if !active {
            inactive = append(inactive, unit)
        }
    }

    return inactive
}

func (stack *UnitStack) AllFlyers() bool {
    for _, unit := range stack.ActiveUnits() {
        if !unit.Unit.Flying {
            return false
        }
    }

    return true
}

func (stack *UnitStack) ToggleActive(unit *units.OverworldUnit){
    value, ok := stack.active[unit]
    if ok {
        // if unit is active then set to inactive
        // if unit is inactive, then only set to active if the unit has moves left

        if value {
            stack.active[unit] = false
        } else if unit.MovesLeft.GreaterThan(fraction.Zero()) {
            stack.active[unit] = true
            unit.Patrol = false
        }
    }
}

func (stack *UnitStack) AddUnit(unit *units.OverworldUnit){
    stack.units = append(stack.units, unit)
    stack.active[unit] = true
}

func (stack *UnitStack) IsActive(unit *units.OverworldUnit) bool {
    val, ok := stack.active[unit]
    if !ok {
        return false
    }
    return val
}

func (stack *UnitStack) RemoveUnits(units []*units.OverworldUnit){
    for _, unit := range units {
        stack.RemoveUnit(unit)
    }
}

func (stack *UnitStack) RemoveUnit(unit *units.OverworldUnit){
    stack.units = slices.DeleteFunc(stack.units, func(u *units.OverworldUnit) bool {
        return u == unit
    })

    delete(stack.active, unit)
}

func (stack *UnitStack) ContainsUnit(unit *units.OverworldUnit) bool {
    return slices.Contains(stack.units, unit)
}

func (stack *UnitStack) Plane() data.Plane {
    if len(stack.units) > 0 {
        return stack.units[0].Plane
    }

    return data.PlaneArcanus
}

func (stack *UnitStack) ExhaustMoves(){
    for _, unit := range stack.units {
        unit.MovesLeft = fraction.Zero()
        stack.active[unit] = false
    }
}

func (stack *UnitStack) EnableMovers(){
    for _, unit := range stack.units {
        if unit.MovesLeft.GreaterThan(fraction.Zero()) && !unit.Patrol {
            stack.active[unit] = true
        } else {
            stack.active[unit] = false
        }
    }
}

func (stack *UnitStack) Move(dx int, dy int, cost fraction.Fraction){
    for _, unit := range stack.units {
        unit.Move(dx, dy, cost)
    }
}

// true if no unit has any moves left
func (stack *UnitStack) OutOfMoves() bool {
    for _, unit := range stack.units {
        if unit.MovesLeft.GreaterThan(fraction.Zero()) {
            return false
        }
    }

    return true
}

// true if any unit in the stack has moves left
func (stack *UnitStack) HasMoves() bool {
    return !stack.OutOfMoves()
}

func (stack *UnitStack) Leader() *units.OverworldUnit {
    // return the first active unit
    for _, unit := range stack.units {
        if stack.active[unit] {
            return unit
        }
    }

    // otherwise just return any unit
    if len(stack.units) > 0 {
        return stack.units[0]
    }

    return nil
}

func (stack *UnitStack) X() int {
    if len(stack.units) > 0 {
        return stack.units[0].X
    }

    return 0
}

func (stack *UnitStack) Y() int {
    if len(stack.units) > 0 {
        return stack.units[0].Y
    }

    return 0
}

// in the magic screen, power is distributed across the 3 categories
type PowerDistribution struct {
    Mana float64
    Research float64
    Skill float64
}

type Player struct {
    // matrix the same size as the map, where true means the player can see the tile
    // and false means the tile has not yet been discovered
    ArcanusFog [][]bool
    MyrrorFog [][]bool

    TaxRate fraction.Fraction

    Gold int
    Mana int

    Human bool

    // known spells
    KnownSpells spellbook.Spells

    // the full set of spells that can be known by the wizard
    ResearchPoolSpells spellbook.Spells

    // spells that can be researched
    ResearchCandidateSpells spellbook.Spells

    PowerDistribution PowerDistribution

    // total power points put into the casting skill
    CastingSkillPower int
    // how much casting skill remains in this turn
    RemainingCastingSkill int

    ResearchingSpell spellbook.Spell
    ResearchProgress int

    // current spell being cast
    CastingSpell spellbook.Spell
    // how much mana has been put towards the current spell. When this value equals
    // the spell's casting cost, the spell is cast
    CastingSpellProgress int

    // the artifact currently being created by a spell cast of Create Artifact or Enchant Item
    CreateArtifact *artifact.Artifact

    Wizard setup.WizardCustom

    Units []*units.OverworldUnit
    Stacks []*UnitStack
    Cities []*citylib.City

    // counter for the next created unit owned by this player
    UnitId uint64
    SelectedStack *UnitStack
}

func (player *Player) IsAI() bool {
    return !player.Human
}

func (player *Player) IsHuman() bool {
    return player.Human
}

/* return the city that contains the summoning circle */
func (player *Player) FindSummoningCity() *citylib.City {
    for _, city := range player.Cities {
        if city.HasSummoningCircle() {
            return city
        }
    }

    return nil
}

func (player *Player) LearnSpell(spell spellbook.Spell) {
    player.ResearchCandidateSpells.RemoveSpell(spell)
    player.KnownSpells.AddSpell(spell)
    player.UpdateResearchCandidates()

    // if the spell learned is the one being researched, then reset the research spell
    if spell.Name == player.ResearchingSpell.Name {
        player.ResearchingSpell = spellbook.Spell{}
        player.ResearchProgress = 0
    }
}

/* fill up the research candidate spells so that there are at most 8.
 * choose spells from the research pool that are not already known, but preferring
 * lower rarity spells first.
 */
func (player *Player) UpdateResearchCandidates() {
    moreSpells := 8 - len(player.ResearchCandidateSpells.Spells)

    // find the set of potential spells to add to the research candidates
    allSpells := player.ResearchPoolSpells.Copy()
    allSpells.RemoveSpells(player.KnownSpells)
    allSpells.RemoveSpells(player.ResearchCandidateSpells)

    realms := []data.MagicType{
        data.LifeMagic, data.SorceryMagic, data.NatureMagic,
        data.DeathMagic, data.ChaosMagic, data.ArcaneMagic,
    }

    chooseSpell := func (spells *spellbook.Spells) spellbook.Spell {
        // for each realm (chosen randomly), try to find a spell to add to the research candidates
        for _, realmIndex := range rand.Perm(len(realms)) {
            rarities := []spellbook.SpellRarity{
                spellbook.SpellRarityCommon, spellbook.SpellRarityUncommon,
                spellbook.SpellRarityRare, spellbook.SpellRarityVeryRare,
            }

            for _, rarity := range rarities {
                candidates := allSpells.GetSpellsByMagic(realms[realmIndex]).GetSpellsByRarity(rarity)

                if len(candidates.Spells) > 0 {
                    return candidates.Spells[rand.IntN(len(candidates.Spells))]
                }
            }
        }

        return spellbook.Spell{}
    }

    for i := 0; i < moreSpells; i++ {
        if len(allSpells.Spells) > 0 {
            spell := chooseSpell(&allSpells)
            if spell.Valid() {
                player.ResearchCandidateSpells.AddSpell(spell)
                allSpells.RemoveSpell(spell)
            }
        }
    }

    player.ResearchCandidateSpells.SortByRarity()

    // log.Printf("Research candidates: %v", player.ResearchCandidateSpells)
}

func (player *Player) ComputeCastingSkill() int {
    if player.CastingSkillPower == 0 {
        return 0
    }

    bonus := 0
    if player.Wizard.AbilityEnabled(setup.AbilityArchmage) {
        bonus = 10
    }

    return int((math.Sqrt(float64(4 * player.CastingSkillPower - 3)) + 1) / 2) + bonus
}

func (player *Player) CastingSkillPerTurn(power int) int {
    bonus := 1.0

    if player.Wizard.AbilityEnabled(setup.AbilityArchmage) {
        bonus = 1.5
    }

    return int(float64(power) * player.PowerDistribution.Skill * bonus)
}

func (player *Player) SpellResearchPerTurn(power int) float64 {
    research := float64(0)

    for _, city := range player.Cities {
        research += float64(city.ResearchProduction())
    }

    research += float64(power) * player.PowerDistribution.Research

    return research
}

func (player *Player) GoldPerTurn() int {
    gold := 0

    for _, city := range player.Cities {
        gold += city.GoldSurplus()
    }

    for _, unit := range player.Units {
        gold -= unit.Unit.UpkeepGold
    }

    return gold
}

func (player *Player) FoodPerTurn() int {
    food := 0

    for _, city := range player.Cities {
        food += city.SurplusFood()
    }

    for _, unit := range player.Units {
        food -= unit.Unit.UpkeepFood
    }

    return food
}

func (player *Player) ManaPerTurn(power int) int {
    mana := 0

    for _, city := range player.Cities {
        mana += city.ManaSurplus()
    }

    for _, unit := range player.Units {
        mana -= unit.Unit.UpkeepMana
    }

    manaFocusingBonus := float64(1)

    if player.Wizard.AbilityEnabled(setup.AbilityManaFocusing) {
        manaFocusingBonus = 1.25
    }

    mana += int(float64(power) * player.PowerDistribution.Mana * manaFocusingBonus)

    return mana
}

func (player *Player) UpdateTaxRate(rate fraction.Fraction){
    player.TaxRate = rate
    for _, city := range player.Cities {
        city.UpdateTaxRate(rate, player.GetUnits(city.X, city.Y))
    }
}

func (player *Player) GetUnits(x int, y int) []*units.OverworldUnit {
    stack := player.FindStack(x, y)
    if stack != nil {
        return stack.Units()
    }

    return nil
}

func (player *Player) FindCity(x int, y int) *citylib.City {
    for _, city := range player.Cities {
        if city.X == x && city.Y == y {
            return city
        }
    }

    return nil
}

func (player *Player) GetFog(plane data.Plane) [][]bool {
    if plane == data.PlaneArcanus {
        return player.ArcanusFog
    } else {
        return player.MyrrorFog
    }
}

func (player *Player) SetSelectedStack(stack *UnitStack){
    player.SelectedStack = stack
}

func (player *Player) LiftFogSquare(x int, y int, squares int){
    // FIXME: make this a parameter
    fog := player.ArcanusFog

    for dx := -squares; dx <= squares; dx++ {
        for dy := -squares; dy <= squares; dy++ {
            if x + dx < 0 || x + dx >= len(fog) || y + dy < 0 || y + dy >= len(fog[0]) {
                continue
            }

            fog[x + dx][y + dy] = true
        }
    }
}

/* make anything within the given radius viewable by the player */
func (player *Player) LiftFog(x int, y int, radius int){

    // FIXME: make this a parameter
    fog := player.ArcanusFog

    for dx := -radius; dx <= radius; dx++ {
        for dy := -radius; dy <= radius; dy++ {
            if x + dx < 0 || x + dx >= len(fog) || y + dy < 0 || y + dy >= len(fog[0]) {
                continue
            }

            if dx * dx + dy * dy <= radius * radius {
                fog[x + dx][y + dy] = true
            }
        }
    }

}

func (player *Player) FindStackByUnit(unit *units.OverworldUnit) *UnitStack {
    for _, stack := range player.Stacks {
        if stack.ContainsUnit(unit) {
            return stack
        }
    }

    return nil
}

func (player *Player) FindStack(x int, y int) *UnitStack {
    for _, stack := range player.Stacks {
        if stack.X() == x && stack.Y() == y {
            return stack
        }
    }

    return nil
}

func (player *Player) MergeStacks(stack1 *UnitStack, stack2 *UnitStack) *UnitStack {
    stack1.units = append(stack1.units, stack2.units...)

    for unit, active := range stack2.active {
        stack1.active[unit] = active
    }

    player.Stacks = slices.DeleteFunc(player.Stacks, func (s *UnitStack) bool {
        return s == stack2
    })

    return stack1
}

func (player *Player) RemoveUnit(unit *units.OverworldUnit) {
    player.Units = slices.DeleteFunc(player.Units, func (u *units.OverworldUnit) bool {
        return u == unit
    })

    stack := player.FindStack(unit.X, unit.Y)
    if stack != nil {
        stack.RemoveUnit(unit)

        if stack.IsEmpty() {
            player.Stacks = slices.DeleteFunc(player.Stacks, func (s *UnitStack) bool {
                return s == stack
            })
        }
    }
}

func (player *Player) AddCity(city *citylib.City) *citylib.City {
    player.Cities = append(player.Cities, city)
    return city
}

func (player *Player) AddStack(stack *UnitStack){
    player.Stacks = append(player.Stacks, stack)
}

func (player *Player) AddUnit(unit *units.OverworldUnit) *units.OverworldUnit {
    unit.Id = player.UnitId
    player.UnitId += 1
    player.Units = append(player.Units, unit)

    stack := player.FindStack(unit.X, unit.Y)
    if stack == nil {
        stack = MakeUnitStack()
        player.Stacks = append(player.Stacks, stack)
    } else {
    }

    stack.AddUnit(unit)

    return unit
}
