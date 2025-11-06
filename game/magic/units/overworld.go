package units

import (
    "slices"
    "cmp"
    "math"

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
    BusyStatusStasis // for units under the effect of stasis
)

type EnchantmentProvider interface {
    HasEnchantmentOnly(data.UnitEnchantment) bool
}

type GlobalEnchantmentProvider interface {
    // true if the owner of this unit has the enchantment active
    HasFriendlyEnchantment(data.Enchantment) bool
    // true if any wizard has the enchantment active (useful for enchantments with negative effects)
    HasEnchantment(data.Enchantment) bool
    // true if any wizard other than the owner has this enchantment active
    HasRivalEnchantment(data.Enchantment) bool
}

// a default empty implementation of GlobalEnchantmentProvider
type NoEnchantments struct {}
func (*NoEnchantments) HasFriendlyEnchantment(enchantment data.Enchantment) bool {
    return false
}

func (*NoEnchantments) HasRivalEnchantment(enchantment data.Enchantment) bool {
    return false
}

func (*NoEnchantments) HasEnchantment(enchantment data.Enchantment) bool {
    return false
}

type OverworldUnit struct {
    ExperienceInfo ExperienceInfo
    Unit Unit
    MovesUsed fraction.Fraction
    Banner data.BannerType

    // a reference to an enclosing struct type, or a reference to self
    // The OverworldUnit type should never invoke the same method name on the parent
    // that is, o.Foo() should not call o.Parent.Foo()
    // o.Bar() can call o.Parent.Foo(), though
    // Parent.Foo() can invoke o.Foo() as well
    Parent StackUnit

    Plane data.Plane
    X int
    Y int
    Id uint64
    Damage int
    // Health int
    // to get the level, use the conversion functions in experience.go
    Experience int
    WeaponBonus data.WeaponBonus
    Undead bool

    Busy BusyStatus

    Enchantments []data.UnitEnchantment

    GlobalEnchantments GlobalEnchantmentProvider

    // This should be set during combat to the ArmyUnit, and unset at all other times.
    // This exists because to compute the experience level of a unit we must know if the unit
    // has heroism enchanted on it, but heroism could be cast as a combat spell in which case normally
    // the enchantment would be added to the ArmyUnit enchantment list, which this unit does not have access to.
    // So instead we allow this unit to access the enchanments of the enclosing ArmyUnit to detect
    // if heroism is active.
    ExtraEnchantments EnchantmentProvider
}

func (unit *OverworldUnit) SetEnchantmentProvider(provider EnchantmentProvider) {
    unit.ExtraEnchantments = provider
}

func (unit *OverworldUnit) AddEnchantment(enchantment data.UnitEnchantment) {
    unit.Enchantments = append(unit.Enchantments, enchantment)
    // keep list sorted
    slices.SortFunc(unit.Enchantments, func(a, b data.UnitEnchantment) int {
        return cmp.Compare(int(a), int(b))
    })
}

func (unit *OverworldUnit) SetParent(parent StackUnit) {
    unit.Parent = parent
}

// checks enchantments on the unit itself, ignoring global enchantments
func (unit *OverworldUnit) hasUnitEnchantment(enchantment data.UnitEnchantment) bool {
    return slices.Contains(unit.Enchantments, enchantment) || (unit.ExtraEnchantments != nil && unit.ExtraEnchantments.HasEnchantmentOnly(enchantment))
}

func (unit *OverworldUnit) HasEnchantment(enchantment data.UnitEnchantment) bool {
    if unit.GetRace() != data.RaceFantastic && enchantment == data.UnitEnchantmentHolyWeapon {
        if unit.GlobalEnchantments.HasFriendlyEnchantment(data.EnchantmentHolyArms) {
            return true
        }
    }

    return unit.hasUnitEnchantment(enchantment)
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

func (unit *OverworldUnit) SetUndead() {
    unit.Undead = true
}

func (unit *OverworldUnit) GetAbilityValue(ability data.AbilityType) float32 {
    if ability == data.AbilityFireBreath {
        value := unit.Unit.GetAbilityValue(ability)
        modifier := float32(0)

        if unit.GetRealm() == data.ChaosMagic && unit.GlobalEnchantments.HasEnchantment(data.EnchantmentChaosSurge) {
            modifier += 2
        }

        if value > 0 {
            return value + modifier
        }

        if unit.HasEnchantment(data.UnitEnchantmentChaosChannelsFireBreath) {
            return 2
        }

        return 0
    }

    if ability == data.AbilityThrown {
        value := unit.Unit.GetAbilityValue(ability)
        modifier := float32(0)

        if unit.GetRealm() == data.ChaosMagic && unit.GlobalEnchantments.HasEnchantment(data.EnchantmentChaosSurge) {
            modifier += 2
        }

        if value > 0 {
            return value + modifier
        }

        return 0
    }

    if ability == data.AbilityDoomGaze {
        value := unit.Unit.GetAbilityValue(ability)
        if value == 0 {
            return 0
        }

        modifier := float32(0)

        if unit.GetRealm() == data.ChaosMagic && unit.GlobalEnchantments.HasEnchantment(data.EnchantmentChaosSurge) {
            modifier += 2
        }

        return value + modifier
    }

    return unit.Unit.GetAbilityValue(ability)
}

func (unit *OverworldUnit) RemoveEnchantment(toRemove data.UnitEnchantment) {
    unit.Enchantments = slices.DeleteFunc(unit.Enchantments, func(enchantment data.UnitEnchantment) bool {
        return enchantment == toRemove
    })
}

func (unit *OverworldUnit) GetUpkeepEnchantments() []data.UnitEnchantment {
    return slices.Clone(unit.Enchantments)
}

func (unit *OverworldUnit) GetEnchantments() []data.UnitEnchantment {
    out := slices.Clone(unit.Enchantments)

    if unit.GetRace() != data.RaceFantastic && unit.GlobalEnchantments.HasFriendlyEnchantment(data.EnchantmentHolyArms) {
        out = append(out, data.UnitEnchantmentHolyWeapon)
    }

    return out
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
    return nil
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
    if unit.IsUndead() {
        return data.RaceFantastic
    }

    if unit.IsChaosChanneled() {
        return data.RaceFantastic
    }

    return unit.Unit.Race
}

func (unit *OverworldUnit) IsChaosChanneled() bool {
    return unit.hasUnitEnchantment(data.UnitEnchantmentChaosChannelsDemonWings) ||
       unit.hasUnitEnchantment(data.UnitEnchantmentChaosChannelsDemonSkin) ||
       unit.hasUnitEnchantment(data.UnitEnchantmentChaosChannelsFireBreath)
}

func (unit *OverworldUnit) GetRealm() data.MagicType {
    if unit.IsUndead() {
        return data.DeathMagic
    }

    if unit.IsChaosChanneled() {
        return data.ChaosMagic
    }

    return unit.Unit.Realm
}

func (unit *OverworldUnit) GetPlanePoint() data.PlanePoint {
    return data.PlanePoint{
        X: unit.X,
        Y: unit.Y,
        Plane: unit.Plane,
    }
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

func (unit *OverworldUnit) GetMovesLeft(overworld bool) fraction.Fraction {
    return fraction.Zero().Max(unit.GetMovementSpeed(overworld).Subtract(unit.MovesUsed))
}

func (unit *OverworldUnit) SetMovesLeft(overworld bool, moves fraction.Fraction) {
    unit.MovesUsed = unit.GetMovementSpeed(overworld).Subtract(moves)
}

func (unit *OverworldUnit) IsFlying() bool {
    return unit.Unit.Flying || unit.Parent.HasEnchantment(data.UnitEnchantmentFlight) || unit.Parent.HasEnchantment(data.UnitEnchantmentChaosChannelsDemonWings)
}

func (unit *OverworldUnit) IsInvisible() bool {
    return unit.Parent.HasAbility(data.AbilityInvisibility)
}

func (unit *OverworldUnit) IsSailing() bool {
    return unit.GetRawUnit().Sailing
}

func (unit *OverworldUnit) IsLandWalker() bool {
    if unit.Parent.IsFlying() || unit.Parent.IsSwimmer() || unit.Parent.IsSailing() || unit.Parent.HasAbility(data.AbilityNonCorporeal) {
        return false
    }

    return true
}

func (unit *OverworldUnit) IsSwimmer() bool {
    return unit.Unit.Swimming || unit.Parent.HasEnchantment(data.UnitEnchantmentWaterWalking)
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
    // check the underlying race because an undead unit might still have experience
    if unit.Unit.Race == data.RaceFantastic {
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

// amount is a positive number to heal
func (unit *OverworldUnit) AdjustHealth(amount int) {
    unit.Damage -= amount
    if unit.Damage < 0 {
        unit.Damage = 0
    }

    if unit.Damage > unit.GetMaxHealth() {
        unit.Damage = unit.GetMaxHealth()
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
    return unit.GetMaxHealth() - unit.Damage
}

func (unit *OverworldUnit) GetDamage() int {
    return unit.Damage
}

func (unit *OverworldUnit) GetMaxHealth() int {
    return unit.GetFullHitPoints() * unit.GetCount()
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

    if unit.HasAbility(data.AbilityLucky) {
        base += 10
    }

    if unit.HasEnchantment(data.UnitEnchantmentHolyWeapon) {
        base += 10
    }

    return base
}

func (unit *OverworldUnit) GetToDefend() int {
    base := 30
    modifier := 0

    if unit.HasAbility(data.AbilityLucky) {
        modifier += 10
    }

    return base + modifier
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

    // undead units automatically get all these abilities
    if unit.IsUndead() &&
       (ability == data.AbilityDeathImmunity || ability == data.AbilityPoisonImmunity ||
        ability == data.AbilityIllusionsImmunity || ability == data.AbilityColdImmunity) {
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

func (unit *OverworldUnit) SetGlobalEnchantmentProvider(provider GlobalEnchantmentProvider) {
    unit.GlobalEnchantments = provider
}

func (unit *OverworldUnit) SetExperienceInfo(info ExperienceInfo) {
    unit.ExperienceInfo = info
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

func (unit *OverworldUnit) GetVisibleCount() int {
    return unit.Unit.GetVisibleCount()
}

func (unit *OverworldUnit) VisibleFigures() int {
    health_per_figure := float64(unit.Parent.GetMaxHealth()) / float64(unit.Parent.GetVisibleCount())
    return int(math.Ceil(float64(unit.Parent.GetHealth()) / health_per_figure))
}

func (unit *OverworldUnit) GetUpkeepGold() int {
    if unit.IsUndead() {
        return 0
    }
    return unit.Unit.GetUpkeepGold()
}

func (unit *OverworldUnit) GetUpkeepFood() int {
    if unit.IsUndead() {
        return 0
    }
    return unit.Unit.GetUpkeepFood()
}

func (unit *OverworldUnit) GetUpkeepMana() int {
    mana := unit.Unit.GetUpkeepMana() 
    if unit.IsUndead() && unit.Unit.Race == data.RaceFantastic {
        return mana * 3 / 2
    }
    return mana
}

func (unit *OverworldUnit) GetBaseMovementSpeed(overworld bool) int {
    return unit.Unit.GetMovementSpeed(overworld)
}

// pass in overworld=true if calculating movement on the overworld map, false if in combat
func (unit *OverworldUnit) GetMovementSpeed(overworld bool) fraction.Fraction {
    base := fraction.FromInt(unit.GetBaseMovementSpeed(overworld))

    base = unit.MovementSpeedEnchantmentBonus(base, unit.Enchantments)

    modifier := fraction.FromInt(1)

    if unit.IsSailing() {
        if unit.GlobalEnchantments.HasFriendlyEnchantment(data.EnchantmentWindMastery) {
            modifier = modifier.Add(fraction.Make(1, 2))
        }

        if unit.GlobalEnchantments.HasRivalEnchantment(data.EnchantmentWindMastery) {
            modifier = modifier.Subtract(fraction.Make(1, 2))
        }
    }

    return base.Multiply(modifier)
}

func (unit *OverworldUnit) GetProductionCost() int {
    return unit.Unit.GetProductionCost()
}

func (unit *OverworldUnit) MovementSpeedEnchantmentBonus(base fraction.Fraction, enchantments []data.UnitEnchantment) fraction.Fraction {

    endurance := false
    flying := false
    haste := false

    for _, enchantment := range enchantments {
        switch enchantment {
            case data.UnitEnchantmentEndurance: endurance = true
            case data.UnitEnchantmentFlight: flying = true
            case data.UnitEnchantmentHaste: haste = true
        }
    }

    if endurance {
        base = base.Add(fraction.FromInt(1))
    }

    if flying {
        base = base.Max(fraction.FromInt(3))
    }

    if haste {
        base = base.Multiply(fraction.FromInt(2))
    }

    return base
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

func (unit *OverworldUnit) ResistanceEnchantmentBonus(enchantment data.UnitEnchantment) int {
    switch enchantment {
        case data.UnitEnchantmentBlackChannels: return 1
        case data.UnitEnchantmentLionHeart: return 3
    }
    return 0
}

func (unit *OverworldUnit) DefenseEnchantmentBonus(enchantment data.UnitEnchantment) int {
    switch enchantment {
        case data.UnitEnchantmentBlackChannels: return 1
        case data.UnitEnchantmentChaosChannelsDemonSkin: return 3
        case data.UnitEnchantmentHolyArmor: return 2
        // FIXME: iron/stone skin are mutually exclusive
        case data.UnitEnchantmentIronSkin: return 5
        case data.UnitEnchantmentStoneSkin: return 1
    }
    return 0
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
    // pure fantastic creatures can never gain any levels, but undead units can have experience, so don't use GetRace() here
    if unit.Unit.Race == data.RaceFantastic {
        return ExperienceRecruit
    }

    // FIXME: verify if undead units can be affected by heroism, warlord and crusade
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

    if unit.GetRealm() == data.ChaosMagic && unit.GlobalEnchantments.HasEnchantment(data.EnchantmentChaosSurge) {
        modifier += 2
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

func (unit *OverworldUnit) RangedEnchantmentBonus(enchantment data.UnitEnchantment) int {
    switch enchantment {
        case data.UnitEnchantmentBlackChannels: return 1
        case data.UnitEnchantmentFlameBlade:
            if unit.GetRangedAttackDamageType() == DamageRangedPhysical {
                return 2
            }
        case data.UnitEnchantmentLionHeart:
            if unit.GetRangedAttackDamageType() != DamageRangedMagical {
                return 3
            }
    }

    return 0
}

func (unit *OverworldUnit) GetRangedAttackPower() int {
    base := unit.GetBaseRangedAttackPower()

    if base == 0 {
        return 0
    }

    if unit.GetRangedAttackDamageType() == DamageRangedPhysical || unit.GetRangedAttackDamageType() == DamageRangedBoulder {
        switch unit.WeaponBonus {
            case data.WeaponMythril: base += 1
            case data.WeaponAdamantium: base += 2
        }
    }

    for _, enchantment := range unit.Enchantments {
        base += unit.RangedEnchantmentBonus(enchantment)
    }

    if unit.GetRealm() == data.ChaosMagic && unit.GlobalEnchantments.HasEnchantment(data.EnchantmentChaosSurge) {
        base += 2
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

func (unit *OverworldUnit) GetResistance() int {
    base := unit.GetBaseResistance()

    for _, enchantment := range unit.Enchantments {
        base += unit.ResistanceEnchantmentBonus(enchantment)
    }

    return base
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

    if unit.HasAbility(data.AbilityLucky) {
        base += 1
    }

    return base
}

func (unit *OverworldUnit) GetFullHitPoints() int {
    base := unit.GetBaseHitPoints()
    for _, enchantment := range unit.Enchantments {
        base += unit.HitPointsEnchantmentBonus(enchantment)
    }

    if unit.GlobalEnchantments.HasFriendlyEnchantment(data.EnchantmentCharmOfLife) {
        base = int(math.Ceil(float64(base) * 1.25))
    }

    return base
}

func (unit *OverworldUnit) HitPointsEnchantmentBonus(enchantment data.UnitEnchantment) int {
    switch enchantment {
        case data.UnitEnchantmentBlackChannels: return 1
        case data.UnitEnchantmentLionHeart: return 3
    }
    return 0
}

func (unit *OverworldUnit) GetHitPoints() int {
    return (unit.GetMaxHealth() - unit.GetDamage()) / unit.GetCount()
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
    // FIXME: should the added death abilities from being undead be added here?

    var enchantmentAbilities []data.Ability
    for _, enchantment := range unit.Enchantments {
        enchantmentAbilities = append(enchantmentAbilities, enchantment.Abilities()...)
    }

    return append(unit.Unit.GetAbilities(), enchantmentAbilities...)
}

func MakeOverworldUnit(unit Unit, x int, y int, plane data.Plane) *OverworldUnit {
    return MakeOverworldUnitFromUnit(unit, x, y, plane, data.BannerBrown, &NoExperienceInfo{}, &NoEnchantments{})
}

func MakeOverworldUnitFromUnit(unit Unit, x int, y int, plane data.Plane, banner data.BannerType, experienceInfo ExperienceInfo, globalEnchantment GlobalEnchantmentProvider) *OverworldUnit {
    out := &OverworldUnit{
        Unit: unit,
        Banner: banner,
        Plane: plane,
        ExperienceInfo: experienceInfo,
        GlobalEnchantments: globalEnchantment,
        X: x,
        Y: y,
    }

    // by default the Parent points to self
    out.Parent = out
    return out
}

/* restore health points on the overworld
 */
func (unit *OverworldUnit) NaturalHeal(rate float64) {
    // undead creatures never heal
    if unit.IsUndead() || unit.GetRealm() == data.DeathMagic {
        return
    }

    amount := float64(unit.GetMaxHealth()) * rate
    if amount < 1 {
        amount = 1
    }
    unit.AdjustHealth(int(amount))
}

func (unit *OverworldUnit) ResetMoves() {
    unit.MovesUsed = fraction.Zero()
}

func (unit *OverworldUnit) HasMovesLeft(overworld bool) bool {
    return unit.GetMovesLeft(overworld).GreaterThan(fraction.Zero())
}

func (unit *OverworldUnit) Move(dx int, dy int, cost fraction.Fraction, normalize NormalizeCoordinateFunc){
    unit.X, unit.Y = normalize(unit.X + dx, unit.Y + dy)

    unit.MovesUsed = unit.MovesUsed.Add(cost)
    /*
    if unit.MovesLeft.LessThan(fraction.Zero()) {
        unit.MovesLeft = fraction.Zero()
    }
    */
}

func (unit *OverworldUnit) GetArtifactSlots() []artifact.ArtifactSlot {
    return nil
}

func (unit *OverworldUnit) GetArtifacts() []*artifact.Artifact {
    return nil
}

func (unit *OverworldUnit) GetSightRange() int {
    scouting := unit.Parent.GetAbilityValue(data.AbilityScouting)
    if scouting >= 2 {
        return int(scouting)
    }

    if unit.IsFlying() {
        return 2
    }

    return 1
}
