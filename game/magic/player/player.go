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

type AIMoveStackDecision struct {
    Stack *UnitStack
    Location image.Point
}

type AICreateUnitDecision struct {
    Unit units.Unit
    X int
    Y int
    Plane data.Plane
}

type PathFinder interface {
    FindPath(oldX int, oldY int, newX int, newY int, stack *UnitStack, fog [][]bool) pathfinding.Path
}

type AIBehavior interface {
    Update(*Player, []*Player, PathFinder) []AIDecision
    NewTurn()
}

type Relationship struct {
    Treaty data.TreatyType
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
    Defeated bool

    Fame int

    // used to seed the random number generator for generating the order of how magic books are drawn
    BookOrderSeed1 uint64
    BookOrderSeed2 uint64

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

    Units []units.StackUnit
    Stacks []*UnitStack
    Cities []*citylib.City

    // counter for the next created unit owned by this player
    UnitId uint64
    SelectedStack *UnitStack
}

func MakePlayer(wizard setup.WizardCustom, human bool, arcanusFog [][]bool, myrrorFog [][]bool) *Player {
    return &Player{
        TaxRate: fraction.FromInt(1),
        ArcanusFog: arcanusFog,
        MyrrorFog: myrrorFog,
        Wizard: wizard,
        Human: human,
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

            hero.Unit = units.MakeOverworldUnitFromUnit(hero.GetRawUnit(), fortressCity.X, fortressCity.Y, fortressCity.Plane, player.Wizard.Banner, player.MakeExperienceInfo())
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
    return experience.Player.Wizard.AbilityEnabled(setup.AbilityWarlord)
}

func (experience *playerExperience) Crusade() bool {
    return experience.Player.GlobalEnchantments.Contains(data.EnchantmentCrusade)
}

func (player *Player) MakeExperienceInfo() units.ExperienceInfo {
    return &playerExperience{
        Player: player,
    }
}

func (player *Player) TotalUnitUpkeepGold() int {
    total := 0

    for _, unit := range player.Units {
        total += unit.GetUpkeepGold()
    }

    total -= player.Fame
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

func (player *Player) ManaPerTurn(power int) int {
    mana := 0

    mana -= player.TotalUnitUpkeepMana()

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

func (player *Player) GetUnits(x int, y int) []units.StackUnit {
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

func (player *Player) WrapX(x int) int {
    fog := player.ArcanusFog
    maximum := len(fog)

    for x < 0 {
        x += maximum
    }

    return x % maximum
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

            fog[mx][my] = true
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
                fog[mx][my] = true
            }
        }
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

    stack := player.FindStack(unit.GetX(), unit.GetY())
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

func (player *Player) AddStack(stack *UnitStack){
    player.Stacks = append(player.Stacks, stack)
}

func (player *Player) AddUnit(unit units.StackUnit) units.StackUnit {
    unit.SetId(player.UnitId)
    player.UnitId += 1
    player.Units = append(player.Units, unit)

    stack := player.FindStack(unit.GetX(), unit.GetY())
    if stack == nil {
        stack = MakeUnitStack()
        player.Stacks = append(player.Stacks, stack)
    } else {
    }

    stack.AddUnit(unit)

    return unit
}
