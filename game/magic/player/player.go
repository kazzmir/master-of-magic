package player

import (
    "slices"
    "math"
    "math/rand/v2"
    "image"

    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/set"
)

// in the magic screen, power is distributed across the 3 categories
type PowerDistribution struct {
    Mana float64
    Research float64
    Skill float64
}

type AIDecision interface {
}

type AIProduceDecision struct {
    City *citylib.City
    Building buildinglib.Building
    Unit units.Unit
}

type AIMoveStackDecision struct {
    Stack *UnitStack
    Location image.Point
    Invalid func()
    ConfirmEncounter func(*maplib.ExtraEncounter) bool
}

type AICastSpellDecision struct {
    Spell spellbook.Spell
}

type AICreateUnitDecision struct {
    Unit units.Unit
    X int
    Y int
    Plane data.Plane
}

type AIBuildOutpostDecision struct {
    Stack *UnitStack
}

// choose a new spell to research
type AIResearchSpellDecision struct {
    Spell spellbook.Spell
}

type AIServices interface {
    FindPath(oldX int, oldY int, newX int, newY int, player *Player, stack *UnitStack, fog data.FogMap) pathfinding.Path
    FindSettlableLocations(x int, y int, plane data.Plane, fog data.FogMap) []image.Point
    IsSettlableLocation(x int, y int, plane data.Plane) bool
}

type AIBehavior interface {
    // return a list of decisions to make for the current turn
    Update(*Player, []*Player, AIServices, int) []AIDecision

    // called after all decisions have been processed for an AI player
    PostUpdate(*Player, []*Player)

    // reset any state that needs to be reset at the start of a new turn
    NewTurn(*Player)

    // called when a new unit is produced in the city
    ProducedUnit(*citylib.City, *Player)

    // true to raze, false to occupy
    ConfirmRazeTown(*citylib.City) bool
}

type Relationship struct {
    Treaty data.TreatyType
}

type CityEnchantment struct {
    City *citylib.City
    Enchantment citylib.Enchantment
}

type CityEnchantmentsProvider interface {
    GetCityEnchantmentsByBanner(banner data.BannerType) []CityEnchantment
}

type Player struct {
    // matrix the same size as the map containg information if the tile is explored,
    // unexplored or in the range of sight
    ArcanusFog data.FogMap
    MyrrorFog data.FogMap

    TaxRate fraction.Fraction

    Gold int
    Mana int

    Human bool
    Defeated bool

    // Fame (without just cause and/or heroes)
    Fame int

    // used to seed the random number generator for generating the order of how magic books are drawn
    BookOrderSeed1 uint64
    BookOrderSeed2 uint64

    // if true, the game will only do strategic (non-graphics/non-realtime) based combat if both sides are strategic
    StrategicCombat bool
    // godmode that lets the player interact with enemy cities/units
    Admin bool

    // true if the wizard is currently banished
    Banished bool

    // known spells
    KnownSpells spellbook.Spells

    // the full set of spells that can be known by the wizard
    ResearchPoolSpells spellbook.Spells

    // spells that can be researched
    ResearchCandidateSpells spellbook.Spells

    GlobalEnchantments *set.Set[data.Enchantment]

    PowerDistribution PowerDistribution

    AIBehavior AIBehavior

    // relations with other players (treaties, etc)
    PlayerRelations map[*Player]*Relationship

    // possible heros that can be employed. some heroes might be dead
    HeroPool map[herolib.HeroType]*herolib.Hero

    // currently employed heroes
    Heroes [6]*herolib.Hero

    VaultEquipment [4]*artifact.Artifact

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

    // FIXME: probably remove Units and just use Stacks to track the units
    Units []units.StackUnit

    Stacks []*UnitStack
    Cities []*citylib.City

    // counter for the next created unit owned by this player
    UnitId uint64
    SelectedStack *UnitStack
}

func MakePlayer(wizard setup.WizardCustom, human bool, mapWidth int, mapHeight int, heroNames map[herolib.HeroType]string) *Player {

    makeFog := func() data.FogMap {
        fog := make(data.FogMap, mapWidth)
        for x := 0; x < mapWidth; x++ {
            fog[x] = make([]data.FogType, mapHeight)
        }
        return fog
    }

    return &Player{
        TaxRate: fraction.FromInt(1),
        ArcanusFog: makeFog(),
        MyrrorFog: makeFog(),
        Wizard: wizard,
        Human: human,
        HeroPool: createHeroes(heroNames),
        PlayerRelations: make(map[*Player]*Relationship),
        GlobalEnchantments: set.MakeSet[data.Enchantment](),
        BookOrderSeed1: rand.Uint64(),
        BookOrderSeed2: rand.Uint64(),
        PowerDistribution: PowerDistribution{
            Mana: 1.0/3,
            Research: 1.0/3,
            Skill: 1.0/3,
        },
    }
}

func createHeroes(names map[herolib.HeroType]string) map[herolib.HeroType]*herolib.Hero {
    heroes := make(map[herolib.HeroType]*herolib.Hero)

    for _, heroType := range herolib.AllHeroTypes() {
        hero := herolib.MakeHeroSimple(heroType)
        hero.SetExtraAbilities()

        name, ok := names[heroType]
        if ok {
            hero.SetName(name)
        }

        heroes[heroType] = hero
    }

    return heroes
}

func (player *Player) GetKnownPlayers() []*Player {
    var out []*Player

    for other, _ := range player.PlayerRelations {
        out = append(out, other)
    }

    return out
}

// this player should now be aware of the other player
func (player *Player) AwarePlayer(other *Player) {
    _, ok := player.PlayerRelations[other]
    if !ok {
        player.PlayerRelations[other] = &Relationship{
        }
    }
}

func (player *Player) WarWithPlayer(other *Player) {
    player.AwarePlayer(other)
    player.PlayerRelations[other].Treaty = data.TreatyWar
}

func (player *Player) PactWithPlayer(other *Player) {
    player.AwarePlayer(other)
    player.PlayerRelations[other].Treaty = data.TreatyPact
}

func (player *Player) AllianceWithPlayer(other *Player) {
    player.AwarePlayer(other)
    player.PlayerRelations[other].Treaty = data.TreatyAlliance
}

func (player *Player) IsAI() bool {
    return !player.Human
}

func (player *Player) IsHuman() bool {
    return player.Human
}

func (player *Player) GetBanner() data.BannerType {
    return player.Wizard.Banner
}

// how much gold is stored in this city relative to the player's overall wealth
func (player *Player) ComputePlunderedGold(city *citylib.City) int {
    totalPopulation := 0
    for _, city := range player.Cities {
        totalPopulation += city.Citizens()
    }

    return int(float64(player.Gold) * float64(city.Citizens()) / float64(max(1, totalPopulation)))
}

func (player *Player) IsTileExplored(x int, y int, plane data.Plane) bool {
    fog := player.GetFog(plane)
    x = player.WrapX(x)
    if x < 0 || x >= len(fog) || y < 0 || y >= len(fog[0]) {
        return false
    }

    return fog[x][y] != data.FogTypeUnexplored
}

/* returns true if the hero was actually added to the player
 */
func (player *Player) AddHero(hero *herolib.Hero) bool {
    fortressCity := player.FindFortressCity()
    if fortressCity == nil {
        return false
    }

    for i := 0; i < len(player.Heroes); i++ {
        if player.Heroes[i] == nil {
            player.Heroes[i] = hero

            // keep the current level of the hero when creating a new overworld unit (which stores the experience)
            level := hero.GetHeroExperienceLevel()
            experienceInfo := player.MakeExperienceInfo()

            hero.Unit = units.MakeOverworldUnitFromUnit(hero.GetRawUnit(), fortressCity.X, fortressCity.Y, fortressCity.Plane, player.Wizard.Banner, experienceInfo)
            hero.AdjustHealth(hero.GetMaxHealth())
            hero.AddExperience(level.ExperienceRequired(experienceInfo.HasWarlord(), experienceInfo.Crusade()))

            player.AddUnit(hero)
            return true
        }
    }

    return false
}

func (player *Player) AliveHeroes() []*herolib.Hero {
    var heroes []*herolib.Hero

    for _, hero := range player.Heroes {
        if hero != nil && hero.Status != herolib.StatusDead {
            heroes = append(heroes, hero)
        }
    }

    return heroes
}

/* return the city that contains the summoning circle */
func (player *Player) FindFortressCity() *citylib.City {
    for _, city := range player.Cities {
        if city.HasFortress() {
            return city
        }
    }

    return nil
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

type playerExperience struct {
    Player *Player
}

func (experience *playerExperience) HasWarlord() bool {
    return experience.Player.Wizard.RetortEnabled(data.RetortWarlord)
}

func (experience *playerExperience) Crusade() bool {
    return experience.Player.GlobalEnchantments.Contains(data.EnchantmentCrusade)
}

func (player *Player) MakeExperienceInfo() units.ExperienceInfo {
    return &playerExperience{
        Player: player,
    }
}

func (player *Player) GetGlobalEnchantments() *set.Set[data.Enchantment] {
    return player.GlobalEnchantments
}

func (player *Player) GetFame() int {
    fame := player.Fame

    for _, hero := range player.Heroes {
        if hero != nil && hero.Status == herolib.StatusEmployed {
            fame += hero.GetAbilityFame()
        }
    }

    if fame < 0 {
        fame = 0
    }

    if player.GlobalEnchantments.Contains(data.EnchantmentJustCause) {
        fame += 10
    }

    return fame
}

func (player *Player) TotalUnitUpkeepGold() int {
    total := 0

    for _, unit := range player.Units {
        total += unit.GetUpkeepGold()
    }

    total -= player.GetFame()
    if total < 0 {
        total = 0
    }

    return total
}

func (player *Player) TotalUnitUpkeepFood() int {
    total := 0

    for _, unit := range player.Units {
        total += unit.GetUpkeepFood()
    }

    return total
}

func (player *Player) TotalUnitUpkeepMana() int {
    total := 0

    for _, unit := range player.Units {
        total += unit.GetUpkeepMana()
    }

    return total
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

// This forces the player to stop casting a spell. Resets progress to 0 and resets the spell being cast.
func (player *Player) InterruptCastingSpell() {
    player.CastingSpell = spellbook.Spell{}
    player.CastingSpellProgress = 0
}

// the base casting skill + any heroes with the caster ability in the fortress
func (player *Player) ComputeOverworldCastingSkill() int {
    base := player.ComputeCastingSkill()
    heroes := float32(0)

    // for each hero at the fortress city, add half of their caster ability to the casting skill
    fortressCity := player.FindFortressCity()
    if fortressCity != nil {
        stack := player.FindStack(fortressCity.X, fortressCity.Y, fortressCity.Plane)
        if stack != nil {
            for _, unit := range stack.Units() {
                caster := unit.GetAbilityValue(data.AbilityCaster)
                heroes += caster
            }
        }
    }

    return base + int(heroes) / 2
}

func (player *Player) ComputeCastingSkill() int {
    if player.CastingSkillPower == 0 {
        return 0
    }

    bonus := 0
    if player.Wizard.RetortEnabled(data.RetortArchmage) {
        bonus = 10
    }

    return int((math.Sqrt(float64(4 * player.CastingSkillPower - 3)) + 1) / 2) + bonus
}

// Used by Cruel Unminding. skillReduction is the actual skill to reduce, but the func reduces the power points put into skill. Returns actual value of the resulting skill (not power) reduction.
func (player *Player) ReduceCastingSkill(reduceBy int) int {
    skillBeforeReduction := player.ComputeCastingSkill()
    investedPowerToReduce := reduceBy * (2 * skillBeforeReduction - reduceBy - 1) // Formula: to reduce skill by X, the invested power to reduce is "deltaP = X * (2 * skill - X - 1)"
    // FIXME: Should the power be allowed to be reduced below 0 to nullify Archmage retort? Wiki doesn't know this.
    investedPowerToReduce = min(player.CastingSkillPower, investedPowerToReduce)
    player.CastingSkillPower -= investedPowerToReduce
    // Workaround: reducing the power if the calculated reduction is insufficient due to int rounding errors. This is tested to usually use less than 10 loop iterations unless the initial skill is absurdly high (more than 30000).
    for player.CastingSkillPower > 0 && player.ComputeCastingSkill() != skillBeforeReduction - reduceBy {
        player.CastingSkillPower -= 1
    }
    return skillBeforeReduction - player.ComputeCastingSkill()
}

func (player *Player) CastingSkillPerTurn(power int) int {
    bonus := 1.0

    if player.Wizard.RetortEnabled(data.RetortArchmage) {
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

    // add in sage heroes
    for _, hero := range player.Heroes {
        if hero != nil && hero.Status == herolib.StatusEmployed {
            research += float64(hero.GetAbilityResearch())
        }
    }

    return research
}

func (player *Player) GoldPerTurn() int {
    gold := 0

    for _, city := range player.Cities {
        gold += city.GoldSurplus()
    }

    gold -= player.TotalUnitUpkeepGold()

    gold += player.FoodPerTurn() / 2

    return gold
}

func (player *Player) FoodPerTurn() int {
    food := 0

    for _, city := range player.Cities {
        food += city.SurplusFood()
    }

    food -= player.TotalUnitUpkeepFood()

    return food
}

func (player *Player) TotalEnchantmentUpkeep(cityEnchantmentsProvider CityEnchantmentsProvider) int {
    upkeep := 0

    for _, enchantment := range player.GlobalEnchantments.Values() {
        upkeep += enchantment.UpkeepMana()
    }

    for _, cityEnchanment := range cityEnchantmentsProvider.GetCityEnchantmentsByBanner(player.GetBanner()) {
        upkeep += cityEnchanment.Enchantment.Enchantment.UpkeepMana()
    }

    for _, unit := range player.Units {
        for _, enchantment := range unit.GetEnchantments() {
            upkeep += enchantment.UpkeepMana()
        }
    }

    return upkeep
}

func (player *Player) ManaPerTurn(power int, cityEnchantmentsProvider CityEnchantmentsProvider) int {
    mana := 0

    mana -= player.TotalUnitUpkeepMana()
    mana -= player.TotalEnchantmentUpkeep(cityEnchantmentsProvider)

    manaFocusingBonus := float64(1)

    if player.Wizard.RetortEnabled(data.RetortManaFocusing) {
        manaFocusingBonus = 1.25
    }

    mana += int(float64(power) * player.PowerDistribution.Mana * manaFocusingBonus)

    return mana
}

func (player *Player) UpdateTaxRate(rate fraction.Fraction){
    player.TaxRate = rate
    player.UpdateUnrest()
}

func (player *Player) UpdateUnrest(){
    for _, city := range player.Cities {
        city.UpdateUnrest()
    }
}

func (player *Player) GetUnits(x int, y int, plane data.Plane) []units.StackUnit {
    stack := player.FindStack(x, y, plane)
    if stack != nil {
        return stack.Units()
    }

    return nil
}

func (player *Player) OwnsStack(stack *UnitStack) bool {
    return slices.ContainsFunc(player.Stacks, func (check *UnitStack) bool {
        return check == stack
    })
}

func (player *Player) OwnsCity(city *citylib.City) bool {
    return slices.ContainsFunc(player.Cities, func (check *citylib.City) bool {
        return check == city
    })
}

func (player *Player) FindCity(x int, y int, plane data.Plane) *citylib.City {
    for _, city := range player.Cities {
        if city.X == x && city.Y == y && city.Plane == plane {
            return city
        }
    }

    return nil
}

func (player *Player) GetFog(plane data.Plane) data.FogMap {
    if plane == data.PlaneArcanus {
        return player.ArcanusFog
    } else {
        return player.MyrrorFog
    }
}

func (player *Player) SetSelectedStack(stack *UnitStack){
    player.SelectedStack = stack
}

func (player *Player) WrapX(x int) int {
    fog := player.ArcanusFog
    maximum := len(fog)

    for x < 0 {
        x += maximum
    }

    return x % maximum
}

func (player *Player) LiftFogAll(plane data.Plane){
    fog := player.GetFog(plane)

    for x := 0; x < len(fog); x++ {
        for y := 0; y < len(fog[0]); y++ {
            fog[x][y] = data.FogTypeVisible
        }
    }
}

func (player *Player) IsVisible(x int, y int, plane data.Plane) bool {
    fog := player.GetFog(plane)
    x = player.WrapX(x)
    if x < 0 || x >= len(fog) || y < 0 || y >= len(fog[0]) {
        return false
    }

    return fog[x][y] == data.FogTypeVisible
}

func (player *Player) LiftFogSquare(x int, y int, squares int, plane data.Plane){
    fog := player.GetFog(plane)

    for dx := -squares; dx <= squares; dx++ {
        for dy := -squares; dy <= squares; dy++ {
            mx := player.WrapX(x + dx)
            my := y + dy

            if mx < 0 || mx >= len(fog) || my < 0 || my >= len(fog[0]) {
                continue
            }

            fog[mx][my] = data.FogTypeVisible
        }
    }
}

// Doesn't provide direct visibility, whilst making the tiles explored.
func (player *Player) ExploreFogSquare(x int, y int, squares int, plane data.Plane){
    fog := player.GetFog(plane)

    for dx := -squares; dx <= squares; dx++ {
        for dy := -squares; dy <= squares; dy++ {
            mx := player.WrapX(x + dx)
            my := y + dy

            if mx < 0 || mx >= len(fog) || my < 0 || my >= len(fog[0]) {
                continue
            }

            if fog[mx][my] == data.FogTypeUnexplored {
                fog[mx][my] = data.FogTypeExplored
            }
        }
    }
}

/* make anything within the given radius viewable by the player */
func (player *Player) LiftFog(x int, y int, radius int, plane data.Plane){
    fog := player.GetFog(plane)

    for dx := -radius; dx <= radius; dx++ {
        for dy := -radius; dy <= radius; dy++ {
            mx := player.WrapX(x + dx)
            my := y + dy

            if mx < 0 || mx >= len(fog) || my < 0 || my >= len(fog[0]) {
                continue
            }

            // dx^2 + dy^2 <= (radius + 0.5)^2
            if 4 * (dx * dx + dy * dy) <= 4 * radius * radius + 4 * radius + 1 {
                fog[mx][my] = data.FogTypeVisible
            }
        }
    }

}

func (player *Player) UpdateFogVisibility() {
    // fog should have already been lifted when this enchantment was cast
    if player.GlobalEnchantments.Contains(data.EnchantmentNatureAwareness) {
        return
    }

    fogs := []data.FogMap{player.ArcanusFog, player.MyrrorFog}

    // reset all visible to explored
    for _, fog := range fogs {
        for x, row := range fog {
            for y := range row {
                if fog[x][y] == data.FogTypeVisible {
                    fog[x][y] = data.FogTypeExplored
                }
            }
        }
    }

    // make tiles visible
    for _, unit := range player.Units {
        player.LiftFogSquare(unit.GetX(), unit.GetY(), unit.GetSightRange(), unit.GetPlane())
    }

    for _, city := range player.Cities {
        player.LiftFogSquare(city.X, city.Y, city.GetSightRange(), city.Plane)
    }
}

func (player *Player) FindStackByUnit(unit units.StackUnit) *UnitStack {
    for _, stack := range player.Stacks {
        if stack.ContainsUnit(unit) {
            return stack
        }
    }

    return nil
}

// multiple stacks can be on the same tile
func (player *Player) FindAllStacks(x int, y int, plane data.Plane) []*UnitStack {
    var out []*UnitStack

    for _, stack := range player.Stacks {
        if stack.X() == x && stack.Y() == y && stack.Plane() == plane {
            out = append(out, stack)
        }
    }

    return out
}

func (player *Player) FindStack(x int, y int, plane data.Plane) *UnitStack {
    for _, stack := range player.Stacks {
        if stack.X() == x && stack.Y() == y && stack.Plane() == plane {
            return stack
        }
    }

    return nil
}

func (player *Player) SplitActiveStack(stack *UnitStack) *UnitStack {
    newStack := stack.SplitActiveUnits()
    if newStack != stack {
        player.Stacks = append(player.Stacks, newStack)
    }

    return newStack
}

func (player *Player) MergeStacks(stack1 *UnitStack, stack2 *UnitStack) *UnitStack {
    stack1.units = append(stack1.units, stack2.units...)

    for unit, active := range stack2.active {
        stack1.active[unit] = active
    }

    player.Stacks = slices.DeleteFunc(player.Stacks, func (s *UnitStack) bool {
        return s == stack2
    })

    if player.SelectedStack == stack2 {
        player.SelectedStack = stack1
    }

    return stack1
}

// teleport/move the unit to a new location. The unit should be removed from its current stack and added to a whatever
// stack exists at the given location.
// a presumption is that the unit is already part of the player's Units list. The unit is not added to the Units list in this method
func (player *Player) UpdateUnitLocation(unit units.StackUnit, x int, y int, plane data.Plane) {
    oldStack := player.FindStackByUnit(unit)
    if oldStack != nil {
        oldStack.RemoveUnit(unit)
        if oldStack.IsEmpty() {
            player.Stacks = slices.DeleteFunc(player.Stacks, func (s *UnitStack) bool {
                return s == oldStack
            })
        }
    }

    unit.SetX(x)
    unit.SetY(y)
    unit.SetPlane(plane)

    newStack := player.FindStack(unit.GetX(), unit.GetY(), unit.GetPlane())
    if newStack == nil {
        newStack = MakeUnitStack()
        player.Stacks = append(player.Stacks, newStack)
    }

    newStack.AddUnit(unit)
}

func (player *Player) RemoveUnit(unit units.StackUnit) {

    for i := 0; i < len(player.Heroes); i++ {
        if player.Heroes[i] == unit {
            if player.Heroes[i].Status == herolib.StatusEmployed {
                player.Heroes[i].SetStatus(herolib.StatusAvailable)
                player.Heroes[i].ResetOwner()
            }
            player.Heroes[i] = nil
        }
    }

    player.Units = slices.DeleteFunc(player.Units, func (u units.StackUnit) bool {
        return u == unit
    })

    stack := player.FindStack(unit.GetX(), unit.GetY(), unit.GetPlane())
    if stack != nil {
        stack.RemoveUnit(unit)

        if stack.IsEmpty() {
            player.Stacks = slices.DeleteFunc(player.Stacks, func (s *UnitStack) bool {
                return s == stack
            })

            if player.SelectedStack == stack {
                player.SelectedStack = nil
            }
        }
    }
}

func (player *Player) AddCity(city *citylib.City) *citylib.City {
    player.Cities = append(player.Cities, city)
    return city
}

func (player *Player) RemoveCity(city *citylib.City) {
    player.Cities = slices.DeleteFunc(player.Cities, func (c *citylib.City) bool {
        return c == city
    })
}

func (player *Player) AddStack(stack *UnitStack) *UnitStack {
    player.Stacks = append(player.Stacks, stack)
    return stack
}

func (player *Player) AddUnit(unit units.StackUnit) units.StackUnit {
    unit.SetId(player.UnitId)
    player.UnitId += 1
    player.Units = append(player.Units, unit)

    stack := player.FindStack(unit.GetX(), unit.GetY(), unit.GetPlane())
    if stack == nil {
        stack = MakeUnitStack()
        player.Stacks = append(player.Stacks, stack)
    }

    stack.AddUnit(unit)

    return unit
}

func (player *Player) HasDivinePower() bool {
    return player.Wizard.RetortEnabled(data.RetortDivinePower)
}

func (player *Player) HasInfernalPower() bool {
    return player.Wizard.RetortEnabled(data.RetortInfernalPower)
}

func (player *Player) HasLifeBooks() bool {
    for _, book := range player.Wizard.Books {
        if book.Magic == data.LifeMagic {
            return true
        }
    }

    return false
}

func (player *Player) HasDeathBooks() bool {
    for _, book := range player.Wizard.Books {
        if book.Magic == data.DeathMagic {
            return true
        }
    }

    return false
}

func (player *Player) TotalBooks() int {
    return player.Wizard.TotalBooks()
}

func (player *Player) GetRulingRace() data.Race {
    return player.Wizard.Race
}

func (player *Player) GetTaxRate() fraction.Fraction {
    return player.TaxRate
}

func (player *Player) GetAllCatchmentArea() *set.Set[data.PlanePoint] {
    catchment := set.MakeSet[data.PlanePoint]()

    for _, city := range player.Cities {
        for point, _ := range city.CatchmentProvider.GetCatchmentArea(city.X, city.Y) {
            catchment.Insert(data.PlanePoint{X: point.X, Y: point.Y, Plane: city.Plane})
        }
    }

    return catchment
}
