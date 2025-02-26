package units

import (
    "slices"
    "cmp"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/lib/fraction"
)

type NormalizeCoordinateFunc func (int, int) (int, int)

// reasons a unit might be unable to move
type BusyStatus int
const (
    BusyStatusNone BusyStatus = iota
    BusyStatusBuildRoad // for engineers
    BusyStatusPurify // for priests
    BusyStatusPatrol // any unit can patrol
)

type OverworldUnit struct {
    ExperienceInfo ExperienceInfo
    Unit Unit
    MovesLeft fraction.Fraction
    Banner data.BannerType
    Plane data.Plane
    X int
    Y int
    Id uint64
    Health int
    // to get the level, use the conversion functions in experience.go
    Experience int
    WeaponBonus data.WeaponBonus
    Undead bool

    Busy BusyStatus

    Enchantments []data.UnitEnchantment
}

func (unit *OverworldUnit) AddEnchantment(enchantment data.UnitEnchantment) {
    unit.Enchantments = append(unit.Enchantments, enchantment)
    // keep list sorted
    slices.SortFunc(unit.Enchantments, func(a, b data.UnitEnchantment) int {
        return cmp.Compare(int(a), int(b))
    })
}

func (unit *OverworldUnit) HasEnchantment(enchantment data.UnitEnchantment) bool {
    return slices.Contains(unit.Enchantments, enchantment)
}

func (unit *OverworldUnit) GetBusy() BusyStatus {
    return unit.Busy
}

func (unit *OverworldUnit) SetBusy(busy BusyStatus) {
    unit.Busy = busy
}

func (unit *OverworldUnit) GetSpellChargeSpells() map[spellbook.Spell]int {
    return make(map[spellbook.Spell]int)
}

func (unit *OverworldUnit) IsUndead() bool {
    return unit.Undead
}

func (unit *OverworldUnit) GetAbilityValue(ability data.AbilityType) float32 {
    return unit.Unit.GetAbilityValue(ability)
}

func (unit *OverworldUnit) RemoveEnchantment(toRemove data.UnitEnchantment) {
    unit.Enchantments = slices.DeleteFunc(unit.Enchantments, func(enchantment data.UnitEnchantment) bool {
        return enchantment == toRemove
    })
}

func (unit *OverworldUnit) GetEnchantments() []data.UnitEnchantment {
    return slices.Clone(unit.Enchantments)
}

func (unit *OverworldUnit) GetLbxFile() string {
    return unit.Unit.LbxFile
}

func (unit *OverworldUnit) GetTitle() string {
    return ""
}

func (unit *OverworldUnit) GetLbxIndex() int {
    return unit.Unit.Index
}

func (unit *OverworldUnit) GetKnownSpells() []string {
    return unit.Unit.KnownSpells
}

func (unit *OverworldUnit) CanTouchAttack(damage Damage) bool {
    return true
}

func (unit *OverworldUnit) SetWeaponBonus(bonus data.WeaponBonus) {
    unit.WeaponBonus = bonus
}

func (unit *OverworldUnit) GetWeaponBonus() data.WeaponBonus {
    return unit.WeaponBonus
}

func (unit *OverworldUnit) IsHero() bool {
    return false
}

func (unit *OverworldUnit) GetPlane() data.Plane {
    return unit.Plane
}

func (unit *OverworldUnit) SetPlane(plane data.Plane) {
    unit.Plane = plane
}

func (unit *OverworldUnit) GetRace() data.Race {
    return unit.Unit.Race
}

func (unit *OverworldUnit) GetRealm() data.MagicType {
    return unit.Unit.Realm
}

func (unit *OverworldUnit) GetX() int {
    return unit.X
}

func (unit *OverworldUnit) GetY() int {
    return unit.Y
}

func (unit *OverworldUnit) SetX(x int) {
    unit.X = x
}

func (unit *OverworldUnit) SetY(y int) {
    unit.Y = y
}

func (unit *OverworldUnit) SetId(id uint64) {
    unit.Id = id
}

func (unit *OverworldUnit) GetMovesLeft() fraction.Fraction {
    return unit.MovesLeft
}

func (unit *OverworldUnit) SetMovesLeft(moves fraction.Fraction) {
    unit.MovesLeft = moves
}

func (unit *OverworldUnit) IsFlying() bool {
    return unit.Unit.Flying || unit.HasEnchantment(data.UnitEnchantmentFlight) || unit.HasEnchantment(data.UnitEnchantmentChaosChannelsDemonWings)
}

func (unit *OverworldUnit) IsSailing() bool {
    return unit.GetRawUnit().Sailing
}

func (unit *OverworldUnit) IsLandWalker() bool {
    if unit.IsFlying() || unit.IsSwimmer() || unit.IsSailing() || unit.HasAbility(data.AbilityNonCorporeal) {
        return false
    }

    return true
}

func (unit *OverworldUnit) IsSwimmer() bool {
    return unit.Unit.Swimming
}

func (unit *OverworldUnit) AddExperience(amount int) {
    if unit.GetRace() == data.RaceFantastic {
        return
    }

    unit.Experience += amount

    // normal units max out at 120 experience
    if unit.GetRace() != data.RaceHero {
        if unit.Experience > 120 {
            unit.Experience = 120
        }
    }

    if unit.Experience >= 120 && unit.HasEnchantment(data.UnitEnchantmentHeroism) {
        unit.RemoveEnchantment(data.UnitEnchantmentHeroism)
    }
}

func (unit *OverworldUnit) GetExperience() int {
    if unit.GetRace() == data.RaceFantastic {
        return 0
    }
    return unit.Experience
}

func (unit *OverworldUnit) GetExperienceData() ExperienceData {
    level := unit.GetExperienceLevel()
    return &level
}

func (unit *OverworldUnit) GetRawUnit() Unit {
    return unit.Unit
}

func (unit *OverworldUnit) AdjustHealth(amount int) {
    unit.Health += amount
    if unit.Health < 0 {
        unit.Health = 0
    }

    if unit.Health > unit.GetMaxHealth() {
        unit.Health = unit.GetMaxHealth()
    }
}

func (unit *OverworldUnit) GetMovementSound() MovementSound {
    return unit.Unit.MovementSound
}

func (unit *OverworldUnit) GetRangeAttackSound() RangeAttackSound {
    return unit.Unit.RangeAttackSound
}

func (unit *OverworldUnit) GetAttackSound() AttackSound {
    return unit.Unit.AttackSound
}

func (unit *OverworldUnit) GetCombatRangeIndex(facing Facing) int {
    return unit.Unit.GetCombatRangeIndex(facing)
}

func (unit *OverworldUnit) GetHealth() int {
    return unit.Health
}

func (unit *OverworldUnit) GetMaxHealth() int {
    return unit.GetHitPoints() * unit.GetCount()
}

func (unit *OverworldUnit) GetToHitMelee() int {
    base := 30 + int(unit.GetAbilityValue(data.AbilityToHit))

    level := unit.GetExperienceLevel()
    switch level {
        case ExperienceRecruit:
        case ExperienceRegular:
        case ExperienceVeteran:
        case ExperienceElite: base += 10
        case ExperienceUltraElite: base += 20
        case ExperienceChampionNormal: base += 30
    }

    switch unit.WeaponBonus {
        case data.WeaponMagic: base += 10
        case data.WeaponMythril: base += 10
        case data.WeaponAdamantium: base += 10
    }

    return base
}

func (unit *OverworldUnit) GetRangedAttackDamageType() Damage {
    return unit.Unit.RangedAttackDamageType
}

func (unit *OverworldUnit) GetRangedAttacks() int {
    return unit.Unit.RangedAttacks
}

// an ability can either be inherint to the unit or granted by an enchantment
func (unit *OverworldUnit) HasAbility(ability data.AbilityType) bool {
    if unit.Unit.HasAbility(ability) {
        return true
    }

    for _, enchantment := range unit.Enchantments {
        for _, grantedAbility := range enchantment.Abilities() {
            if grantedAbility.Ability == ability {
                return true
            }
        }
    }

    return false
}

func (unit *OverworldUnit) HasItemAbility(ability data.ItemAbility) bool {
    return false
}

func (unit *OverworldUnit) SetBanner(banner data.BannerType) {
    unit.Banner = banner
}

func (unit *OverworldUnit) GetBanner() data.BannerType {
    return unit.Banner
}

func (unit *OverworldUnit) GetName() string {
    return unit.Unit.GetName()
}

func (unit *OverworldUnit) GetFullName() string {
    return unit.GetName()
}

func (unit *OverworldUnit) GetCombatLbxFile() string {
    return unit.Unit.GetCombatLbxFile()
}

func (unit *OverworldUnit) GetCombatIndex(facing Facing) int {
    return unit.Unit.GetCombatIndex(facing)
}

func (unit *OverworldUnit) GetCount() int {
    return unit.Unit.GetCount()
}

func (unit *OverworldUnit) GetUpkeepGold() int {
    return unit.Unit.GetUpkeepGold()
}

func (unit *OverworldUnit) GetUpkeepFood() int {
    return unit.Unit.GetUpkeepFood()
}

func (unit *OverworldUnit) GetUpkeepMana() int {
    return unit.Unit.GetUpkeepMana()
}

func (unit *OverworldUnit) GetMovementSpeed() int {
    return unit.Unit.GetMovementSpeed()
}

func (unit *OverworldUnit) GetProductionCost() int {
    return unit.Unit.GetProductionCost()
}

// apply modifiers for melee power
func (unit *OverworldUnit) MeleeEnchantmentBonus(enchantment data.UnitEnchantment) int {
    switch enchantment {
        case data.UnitEnchantmentGiantStrength: return 1
        case data.UnitEnchantmentBlackChannels: return 2
        case data.UnitEnchantmentFlameBlade: return 2
        case data.UnitEnchantmentLionHeart: return 3
    }

    return 0
}

func (unit *OverworldUnit) DefenseEnchantmentBonus(enchantment data.UnitEnchantment) int {
    switch enchantment {
        case data.UnitEnchantmentBlackChannels: return 1
        case data.UnitEnchantmentHolyArmor: return 2
        // FIXME: iron/stone skin are mutually exclusive
        case data.UnitEnchantmentIronSkin: return 5
        case data.UnitEnchantmentStoneSkin: return 1
    }
    return 0
}

func (unit *OverworldUnit) GetFullMeleeAttackPower() int {
    return unit.GetMeleeAttackPower()
}

func (unit *OverworldUnit) GetBaseMeleeAttackPower() int {
    power := unit.Unit.GetMeleeAttackPower()

    if power == 0 {
        return 0
    }

    level := unit.GetExperienceLevel()
    switch level {
        case ExperienceRecruit:
        case ExperienceRegular: power += 1
        case ExperienceVeteran: power += 1
        case ExperienceElite: power += 2
        case ExperienceUltraElite: power += 2
        case ExperienceChampionNormal: power += 3
    }

    return power
}

func (unit *OverworldUnit) GetHeroExperienceLevel() HeroExperienceLevel {
    return ExperienceHero
}

func (unit *OverworldUnit) GetExperienceLevel() NormalExperienceLevel {
    // fantastic creatures can never gain any levels
    if unit.GetRace() == data.RaceFantastic {
        return ExperienceRecruit
    }

    experience := unit.Experience
    if unit.HasEnchantment(data.UnitEnchantmentHeroism) {
        experience = 120
    }

    if unit.ExperienceInfo != nil {
        return GetNormalExperienceLevel(experience, unit.ExperienceInfo.HasWarlord(), unit.ExperienceInfo.Crusade())
    }

    return ExperienceRecruit
}

func (unit *OverworldUnit) GetMeleeAttackPower() int {
    base := unit.GetBaseMeleeAttackPower()

    if base == 0 {
        return 0
    }

    modifier := 0

    switch unit.WeaponBonus {
        case data.WeaponMythril: modifier += 1
        case data.WeaponAdamantium: modifier += 2
    }

    for _, enchantment := range unit.Enchantments {
        modifier += unit.MeleeEnchantmentBonus(enchantment)
    }

    return base + modifier
}

func (unit *OverworldUnit) GetBaseRangedAttackPower() int {
    base := unit.Unit.GetRangedAttackPower()

    if base == 0 {
        return 0
    }

    level := unit.GetExperienceLevel()
    switch level {
        case ExperienceRecruit:
        case ExperienceRegular: base += 1
        case ExperienceVeteran: base += 1
        case ExperienceElite: base += 2
        case ExperienceUltraElite: base += 2
        case ExperienceChampionNormal: base += 3
    }

    return base
}

func (unit *OverworldUnit) GetFullRangedAttackPower() int {
    return unit.GetRangedAttackPower()
}

func (unit *OverworldUnit) GetRangedAttackPower() int {
    base := unit.GetBaseRangedAttackPower()

    if unit.GetRangedAttackDamageType() == DamageRangedPhysical || unit.GetRangedAttackDamageType() == DamageRangedBoulder {
        switch unit.WeaponBonus {
            case data.WeaponMythril: base += 1
            case data.WeaponAdamantium: base += 2
        }
    }

    return base
}

func (unit *OverworldUnit) GetBaseDefense() int {
    defense := unit.Unit.GetDefense()

    level := unit.GetExperienceLevel()
    switch level {
        case ExperienceRecruit:
        case ExperienceRegular:
        case ExperienceVeteran: defense += 1
        case ExperienceElite: defense += 1
        case ExperienceUltraElite: defense += 2
        case ExperienceChampionNormal: defense += 2
    }

    return defense
}

func (unit *OverworldUnit) GetFullDefense() int {
    return unit.GetDefense()
}

func (unit *OverworldUnit) GetDefense() int {
    base := unit.GetBaseDefense()

    modifier := 0

    switch unit.WeaponBonus {
        case data.WeaponMythril: modifier += 1
        case data.WeaponAdamantium: modifier += 2
    }

    for _, enchantment := range unit.Enchantments {
        modifier += unit.DefenseEnchantmentBonus(enchantment)
    }

    return base + modifier
}

func (unit *OverworldUnit) GetFullResistance() int {
    return unit.GetResistance()
}

func (unit *OverworldUnit) GetResistance() int {
    return unit.GetBaseResistance()
}

func (unit *OverworldUnit) GetBaseResistance() int {
    base := unit.Unit.GetResistance()

    level := unit.GetExperienceLevel()
    switch level {
        case ExperienceRecruit:
        case ExperienceRegular: base += 1
        case ExperienceVeteran: base += 2
        case ExperienceElite: base += 3
        case ExperienceUltraElite: base += 4
        case ExperienceChampionNormal: base += 5
    }

    return base
}

func (unit *OverworldUnit) GetFullHitPoints() int {
    return unit.GetHitPoints()
}

func (unit *OverworldUnit) GetHitPoints() int {
    return unit.GetBaseHitPoints()
}

func (unit *OverworldUnit) GetBaseHitPoints() int {
    base := unit.Unit.GetHitPoints()

    level := unit.GetExperienceLevel()
    switch level {
        case ExperienceRecruit:
        case ExperienceRegular:
        case ExperienceVeteran:
        case ExperienceElite: base += 1
        case ExperienceUltraElite: base += 1
        case ExperienceChampionNormal: base += 2
    }

    return base
}

func (unit *OverworldUnit) GetAbilities() []data.Ability {
    return unit.Unit.GetAbilities()
}

func MakeOverworldUnit(unit Unit, x int, y int, plane data.Plane) *OverworldUnit {
    return MakeOverworldUnitFromUnit(unit, x, y, plane, data.BannerBrown, nil)
}

func MakeOverworldUnitFromUnit(unit Unit, x int, y int, plane data.Plane, banner data.BannerType, experienceInfo ExperienceInfo) *OverworldUnit {
    return &OverworldUnit{
        Unit: unit,
        Banner: banner,
        Plane: plane,
        MovesLeft: fraction.FromInt(unit.MovementSpeed),
        Health: unit.GetMaxHealth(),
        ExperienceInfo: experienceInfo,
        X: x,
        Y: y,
    }
}

/* restore health points on the overworld
 */
func (unit *OverworldUnit) NaturalHeal(rate float64) {
    amount := float64(unit.GetMaxHealth()) * rate
    if amount < 1 {
        amount = 1
    }
    unit.AdjustHealth(int(amount))
}

func (unit *OverworldUnit) ResetMoves() {
    unit.MovesLeft = fraction.FromInt(unit.Unit.MovementSpeed)
}

func (unit *OverworldUnit) HasMovesLeft() bool {
    return unit.MovesLeft.GreaterThan(fraction.Zero())
}

func (unit *OverworldUnit) Move(dx int, dy int, cost fraction.Fraction, normalize NormalizeCoordinateFunc){
    unit.X, unit.Y = normalize(unit.X + dx, unit.Y + dy)

    unit.MovesLeft = unit.MovesLeft.Subtract(cost)
    if unit.MovesLeft.LessThan(fraction.Zero()) {
        unit.MovesLeft = fraction.Zero()
    }
}

func (unit *OverworldUnit) GetArtifactSlots() []artifact.ArtifactSlot {
    return nil
}

func (unit *OverworldUnit) GetArtifacts() []*artifact.Artifact {
    return nil
}

func (unit *OverworldUnit) GetSightRange() int {
    scouting := unit.GetAbilityValue(data.AbilityScouting)
    if scouting >= 2 {
        return int(scouting)
    }

    if unit.IsFlying() {
        return 2
    }

    return 1
}
