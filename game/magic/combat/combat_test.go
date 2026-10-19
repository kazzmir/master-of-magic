package combat

import (
    "log"
    "net"
    "sync"
    "context"
    "time"
    "testing"
    "math"
    "slices"
    "image"

    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/pathfinding"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/lib/fraction"
    "github.com/kazzmir/master-of-magic/lib/optional"
    "github.com/kazzmir/master-of-magic/lib/set"
)

func TestAngle(test *testing.T){

    if !betweenAngle(0, 0, math.Pi/8){
        test.Errorf("Error check 0 in 0 spread pi/4")
    }

    if !betweenAngle(math.Pi/2, math.Pi/4, math.Pi/2){
        test.Errorf("Error check pi/2 in pi/4 spread pi/2")
    }

    if !betweenAngle(-math.Pi, math.Pi, math.Pi/8){
        test.Errorf("Error check -pi in pi spread pi/8")
    }

    if betweenAngle(math.Pi, 0, math.Pi/8){
        test.Errorf("Error check pi not in 0 spread pi/4")
    }

    if betweenAngle(0, math.Pi, math.Pi/3){
        test.Errorf("Error check 0 not in pi spread pi/3")
    }

}

func tmp(f units.Facing){
    _ = f
}

func BenchmarkAngle(bench *testing.B){
    var final units.Facing
    for range bench.N {
        // for angle := range 360 {
        angle := 32
            radians := float64(angle) * math.Pi / 180
            facing := computeFacing(radians)
            if facing == units.FacingDown {
                final = facing
            }
        // }
    }

    tmp(final)
}

func TestUnitHealth(test *testing.T) {
    unit := units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    armyUnit := ArmyUnit{
        Unit: unit,
    }

    if armyUnit.GetLeadUnitHealth() != 2 {
        test.Errorf("Error: lead unit health should be 2")
    }

    if armyUnit.Figures() != 8 {
        test.Errorf("Error: figures should be 8")
    }

    if armyUnit.GetDamage() != 0 {
        test.Errorf("Error: damage should be 0")
    }

    // each figure has 2 hp, so taking one damage should keep 8 figures
    armyUnit.TakeDamage(1, DamageNormal)

    if armyUnit.GetLeadUnitHealth() != 1 {
        test.Errorf("Error: lead unit health should be 1")
    }

    if armyUnit.Figures() != 8 {
        test.Errorf("Error: figures should be 8")
    }

    if armyUnit.GetDamage() != 1 {
        test.Errorf("Error: damage should be 1")
    }

    // kill one figure
    armyUnit.TakeDamage(1, DamageNormal)

    if armyUnit.GetLeadUnitHealth() != 2 {
        test.Errorf("Error: lead unit health should be 2")
    }

    if armyUnit.Figures() != 7 {
        test.Errorf("Error: figures should be 7")
    }

    if armyUnit.GetDamage() != 2 {
        test.Errorf("Error: damage should be 2")
    }
}

type TestObserver struct {
    Melee func(attacker *ArmyUnit, defender *ArmyUnit, damageRoll []int)
    Throw func(attacker *ArmyUnit, defender *ArmyUnit, defenderDamage int)
    PoisonTouch func(attacker *ArmyUnit, defender *ArmyUnit, damage int)
    Fear func(attacker *ArmyUnit, defender *ArmyUnit, fear int)
}

func (observer *TestObserver) ThrowAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int){
    if observer.Throw != nil {
        observer.Throw(attacker, defender, damage)
    }
}

func (observer *TestObserver) PoisonTouchAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int){
    if observer.PoisonTouch != nil {
        observer.PoisonTouch(attacker, defender, damage)
    }
}

func (observer *TestObserver) LifeStealTouchAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int){
}

func (observer *TestObserver) StoningTouchAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int){
}

func (observer *TestObserver) DispelEvilTouchAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int){
}

func (observer *TestObserver) DeathTouchAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int){
}

func (observer *TestObserver) DestructionAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int){
}

func (observer *TestObserver) StoneGazeAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int){
}

func (observer *TestObserver) DeathGazeAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int){
}

func (observer *TestObserver) DoomGazeAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int){
}

func (observer *TestObserver) FireBreathAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int){
}

func (observer *TestObserver) LightningBreathAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int){
}

func (observer *TestObserver) ImmolationAttack(attacker *ArmyUnit, defender *ArmyUnit, damage int){
}

func (observer *TestObserver) MeleeAttack(attacker *ArmyUnit, defender *ArmyUnit, damageRoll []int){
    if observer.Melee != nil {
        observer.Melee(attacker, defender, damageRoll)
    }
}

func (observer *TestObserver) CauseFear(attacker *ArmyUnit, defender *ArmyUnit, fear int){
    if observer.Fear != nil {
        observer.Fear(attacker, defender, fear)
    }
}

func (observer *TestObserver) WallOfFire(unit *ArmyUnit, damage int){
}

func (observer *TestObserver) UnitKilled(unit *ArmyUnit){
}

func TestBasicMelee(test *testing.T){
    defendingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    defender := units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    attacker := units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})

    defendingArmy.AddUnit(defender)
    attackingArmy.AddUnit(attacker)

    combat := &CombatModel{
        SelectedUnit: nil,
        Tiles: makeTiles(30, 30, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
        Turn: TeamDefender,
        DefendingArmy: defendingArmy,
        AttackingArmy: attackingArmy,
    }

    combat.Initialize(spellbook.Spells{}, 0, 0)

    attackerMelee := false
    defenderMelee := false

    observer := &TestObserver{
        Melee: func(meleeAttacker *ArmyUnit, meleeDefender *ArmyUnit, damageRoll []int){
            if attackingArmy.units[0] == meleeAttacker {
                attackerMelee = true
            } else if defendingArmy.units[0] == meleeAttacker {
                defenderMelee = true
            }
        },
    }
    combat.Observer.AddObserver(observer)

    // even if both units kill each other, they both get to attack
    combat.meleeAttack(attackingArmy.units[0], defendingArmy.units[0])

    if !attackerMelee || !defenderMelee {
        test.Errorf("Error: attacker and defender should have both attacked")
    }
}

// attacker is multi-figure so should do multiple damage rolls
// multiple small damage rolls that are easily blockable should result in 0 damage
func TestMeleeMultiFigure(test *testing.T){
    defendingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    defender := units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    attacker := units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})

    // should easily block all 1 damage rolls
    defender.Unit.Defense = 100

    defendingArmy.AddUnit(defender)
    attackingArmy.AddUnit(attacker)

    combat := &CombatModel{
        SelectedUnit: nil,
        Tiles: makeTiles(30, 30, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
        Turn: TeamDefender,
        DefendingArmy: defendingArmy,
        AttackingArmy: attackingArmy,
    }

    combat.Initialize(spellbook.Spells{}, 0, 0)

    // since each roll does 1 damage, and defender has 100% block, defender should take 0 damage
    // if instead the damage was added up into one number then the defender would have to block 2000 points of damage
    var rolls []int
    for range 2000 {
        rolls = append(rolls, 1) // always roll 1
    }

    hurt, _ := ApplyDamage(defendingArmy.units[0], rolls, units.DamageMeleePhysical, DamageSourceNormal, DamageModifiers{})
    if hurt > 0 {
        test.Errorf("Error: defender should have taken 0 damage, got %d", hurt)
    }
}

func TestAttackerHaste(test *testing.T){
    defendingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    defender := units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    attacker := units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})

    defendingArmy.AddUnit(defender)
    attackingArmy.AddUnit(attacker)

    attacker.AddEnchantment(data.UnitEnchantmentHaste)

    combat := &CombatModel{
        SelectedUnit: nil,
        Tiles: makeTiles(30, 30, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
        Turn: TeamDefender,
        DefendingArmy: defendingArmy,
        AttackingArmy: attackingArmy,
    }

    combat.Initialize(spellbook.Spells{}, 0, 0)

    attackerMelee := 0
    defenderMelee := 0

    observer := &TestObserver{
        Melee: func(meleeAttacker *ArmyUnit, meleeDefender *ArmyUnit, damageRoll []int){
            if attackingArmy.units[0] == meleeAttacker {
                attackerMelee += 1
            } else if defendingArmy.units[0] == meleeAttacker {
                defenderMelee += 1
            }
        },
    }
    combat.Observer.AddObserver(observer)

    // attacker should get to attack twice
    combat.meleeAttack(attackingArmy.units[0], defendingArmy.units[0])

    if attackerMelee != 2 {
        test.Errorf("Error: attacker should have attacked twice")
    }

    if defenderMelee != 1 {
        test.Errorf("Error: defender should have attacked once")
    }
}

// attacker should melee first and cause enough damage to kill the defender
func TestFirstStrike(test *testing.T){
    defendingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackerUnit := units.LizardSpearmen
    attackerUnit.Abilities = append(slices.Clone(attackerUnit.Abilities), data.MakeAbility(data.AbilityFirstStrike))
    // ensure attacker can kill the defender in one hit
    attackerUnit.MeleeAttackPower = 10000

    defender := units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    attacker := units.MakeOverworldUnitFromUnit(attackerUnit, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})

    defendingArmy.AddUnit(defender)
    attackingArmy.AddUnit(attacker)

    combat := &CombatModel{
        SelectedUnit: nil,
        Tiles: makeTiles(30, 30, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
        Turn: TeamDefender,
        DefendingArmy: defendingArmy,
        AttackingArmy: attackingArmy,
    }

    combat.Initialize(spellbook.Spells{}, 0, 0)

    attackerMelee := 0
    defenderMelee := 0

    observer := &TestObserver{
        Melee: func(meleeAttacker *ArmyUnit, meleeDefender *ArmyUnit, damageRoll []int){
            if attackingArmy.units[0] == meleeAttacker {
                attackerMelee += 1
            } else if defendingArmy.units[0] == meleeAttacker {
                defenderMelee += 1
            }
        },
    }
    combat.Observer.AddObserver(observer)

    // attacker should get to attack first
    combat.meleeAttack(attackingArmy.units[0], defendingArmy.units[0])

    if attackerMelee != 1 {
        test.Errorf("Error: attacker should have attacked once")
    }

    if defenderMelee != 0 {
        test.Errorf("Error: defender should have been killed before attacking")
    }
}

// first strike is negated, so units attack each other at the same time
func TestFirstStrikeNegate(test *testing.T){
    defendingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackerUnit := units.LizardSpearmen
    attackerUnit.Abilities = append(slices.Clone(attackerUnit.Abilities), data.MakeAbility(data.AbilityFirstStrike))
    // ensure attacker can kill the defender in one hit
    attackerUnit.MeleeAttackPower = 10000

    defenderUnit := units.LizardSpearmen
    defenderUnit.Abilities = append(slices.Clone(defenderUnit.Abilities), data.MakeAbility(data.AbilityNegateFirstStrike))

    defender := units.MakeOverworldUnitFromUnit(defenderUnit, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    attacker := units.MakeOverworldUnitFromUnit(attackerUnit, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})

    defendingArmy.AddUnit(defender)
    attackingArmy.AddUnit(attacker)

    combat := &CombatModel{
        SelectedUnit: nil,
        Tiles: makeTiles(30, 30, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
        Turn: TeamDefender,
        DefendingArmy: defendingArmy,
        AttackingArmy: attackingArmy,
    }

    combat.Initialize(spellbook.Spells{}, 0, 0)

    attackerMelee := 0
    defenderMelee := 0

    observer := &TestObserver{
        Melee: func(meleeAttacker *ArmyUnit, meleeDefender *ArmyUnit, damageRoll []int){
            if attackingArmy.units[0] == meleeAttacker {
                attackerMelee += 1
            } else if defendingArmy.units[0] == meleeAttacker {
                defenderMelee += 1
            }
        },
    }
    combat.Observer.AddObserver(observer)

    combat.meleeAttack(attackingArmy.units[0], defendingArmy.units[0])

    if attackerMelee != 1 {
        test.Errorf("Error: attacker should have attacked once")
    }

    if defenderMelee != 1 {
        test.Errorf("Error: defender should have attacked once")
    }
}

func TestThrowAttack(test *testing.T){
    defendingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackerUnit := units.LizardSpearmen
    attackerUnit.Abilities = append(slices.Clone(attackerUnit.Abilities), data.MakeAbilityValue(data.AbilityThrown, 10000), data.MakeAbilityValue(data.AbilityToHit, 100))
    // ensure attacker can kill the defender in one hit
    attackerUnit.MeleeAttackPower = 10000

    defenderUnit := units.LizardSpearmen

    defender := units.MakeOverworldUnitFromUnit(defenderUnit, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    attacker := units.MakeOverworldUnitFromUnit(attackerUnit, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})

    defendingArmy.AddUnit(defender)
    attackingArmy.AddUnit(attacker)

    combat := &CombatModel{
        SelectedUnit: nil,
        Tiles: makeTiles(30, 30, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
        Turn: TeamDefender,
        DefendingArmy: defendingArmy,
        AttackingArmy: attackingArmy,
    }

    combat.Initialize(spellbook.Spells{}, 0, 0)

    attackerMelee := 0
    defenderMelee := 0
    attackerThrow := 0

    observer := &TestObserver{
        Throw: func(throwAttacker *ArmyUnit, throwDefender *ArmyUnit, damage int){
            if throwAttacker == attackingArmy.units[0] {
                attackerThrow += 1
            }
        },
        Melee: func(meleeAttacker *ArmyUnit, meleeDefender *ArmyUnit, damageRoll []int){
            if attackingArmy.units[0] == meleeAttacker {
                attackerMelee += 1
            } else if defendingArmy.units[0] == meleeAttacker {
                defenderMelee += 1
            }
        },
    }
    combat.Observer.AddObserver(observer)

    combat.meleeAttack(attackingArmy.units[0], defendingArmy.units[0])

    if attackerThrow != 1 {
        test.Errorf("Error: attacker should have thrown once")
    }

    if attackerMelee != 0 {
        test.Errorf("Error: attacker should have not attacked")
    }

    if defenderMelee != 0 {
        test.Errorf("Error: defender should have not attacked")
    }
}

func TestThrownTouchAttack(test *testing.T){
    defendingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackerUnit := units.LizardSpearmen
    attackerUnit.Abilities = append(slices.Clone(attackerUnit.Abilities),
        data.MakeAbilityValue(data.AbilityThrown, 10000),
        data.MakeAbilityValue(data.AbilityToHit, 100),
        data.MakeAbilityValue(data.AbilityPoisonTouch, 10),
    )
    // ensure attacker can kill the defender in one hit
    attackerUnit.MeleeAttackPower = 10000

    defenderUnit := units.LizardSpearmen

    defender := units.MakeOverworldUnitFromUnit(defenderUnit, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    attacker := units.MakeOverworldUnitFromUnit(attackerUnit, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})

    defendingArmy.AddUnit(defender)
    attackingArmy.AddUnit(attacker)

    combat := &CombatModel{
        SelectedUnit: nil,
        Tiles: makeTiles(30, 30, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
        Turn: TeamDefender,
        DefendingArmy: defendingArmy,
        AttackingArmy: attackingArmy,
    }

    combat.Initialize(spellbook.Spells{}, 0, 0)

    attackerMelee := 0
    defenderMelee := 0
    attackerThrow := 0
    attackerPoison := 0

    observer := &TestObserver{
        Throw: func(throwAttacker *ArmyUnit, throwDefender *ArmyUnit, damage int){
            if throwAttacker == attackingArmy.units[0] {
                attackerThrow += 1
            }
        },
        PoisonTouch: func(poisonAttacker *ArmyUnit, poisonDefender *ArmyUnit, damage int){
            if attackingArmy.units[0] == poisonAttacker {
                attackerPoison += 1
            }
        },
        Melee: func(meleeAttacker *ArmyUnit, meleeDefender *ArmyUnit, damageRoll []int){
            if attackingArmy.units[0] == meleeAttacker {
                attackerMelee += 1
            } else if defendingArmy.units[0] == meleeAttacker {
                defenderMelee += 1
            }
        },
    }
    combat.Observer.AddObserver(observer)

    combat.meleeAttack(attackingArmy.units[0], defendingArmy.units[0])

    if attackerThrow != 1 {
        test.Errorf("Error: attacker should have thrown once")
    }

    if attackerPoison != 1 {
        test.Errorf("Error: attacker should have done poison touch once")
    }

    if attackerMelee != 0 {
        test.Errorf("Error: attacker should have not attacked")
    }

    if defenderMelee != 0 {
        test.Errorf("Error: defender should have not attacked")
    }
}

// attacker causes fear in defender, so defender does not attack
func TestFear(test *testing.T){
    defendingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackerUnit := units.LizardSpearmen
    attackerUnit.Abilities = append(slices.Clone(attackerUnit.Abilities), data.MakeAbility(data.AbilityCauseFear))

    defenderUnit := units.LizardSwordsmen
    // ensure all units become afraid
    defenderUnit.Resistance = -100

    defender := units.MakeOverworldUnitFromUnit(defenderUnit, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    attacker := units.MakeOverworldUnitFromUnit(attackerUnit, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})

    defendingArmy.AddUnit(defender)
    attackingArmy.AddUnit(attacker)

    combat := &CombatModel{
        SelectedUnit: nil,
        Tiles: makeTiles(30, 30, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
        Turn: TeamDefender,
        DefendingArmy: defendingArmy,
        AttackingArmy: attackingArmy,
    }

    combat.Initialize(spellbook.Spells{}, 0, 0)

    attackerMelee := 0
    defenderMelee := 0

    observer := &TestObserver{
        Melee: func(meleeAttacker *ArmyUnit, meleeDefender *ArmyUnit, damageRoll []int){
            if attackingArmy.units[0] == meleeAttacker {
                attackerMelee += 1
            } else if defendingArmy.units[0] == meleeAttacker {
                defenderMelee += 1
            }
        },
        Fear: func(fearAttacker *ArmyUnit, fearDefender *ArmyUnit, fear int){
            if fearAttacker != attackingArmy.units[0] {
                test.Errorf("Error: attacker should have caused fear")
            }

            if fear != fearDefender.Figures() {
                test.Errorf("Error: fear should be equal to defender figures")
            }
        },
    }
    combat.Observer.AddObserver(observer)

    combat.meleeAttack(attackingArmy.units[0], defendingArmy.units[0])

    if attackerMelee != 1 {
        test.Errorf("Error: attacker should have attacked once")
    }

    if defenderMelee != 0 {
        test.Errorf("Error: defender should have not attacked")
    }
}

func TestCounterAttackPenalty(test *testing.T){
    defendingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackerUnit := units.LizardSpearmen
    defenderUnit := units.LizardSwordsmen

    defender := units.MakeOverworldUnitFromUnit(defenderUnit, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    attacker := units.MakeOverworldUnitFromUnit(attackerUnit, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})

    defendingArmy.AddUnit(defender)
    attackingArmy.AddUnit(attacker)

    combat := &CombatModel{
        SelectedUnit: nil,
        Tiles: makeTiles(30, 30, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
        Turn: TeamDefender,
        DefendingArmy: defendingArmy,
        AttackingArmy: attackingArmy,
    }

    combat.Initialize(spellbook.Spells{}, 0, 0)

    if defendingArmy.units[0].GetCounterAttackToHit(attackingArmy.GetUnits()[0]) != 30 {
        test.Errorf("Error: defender should have normal 30%% counter attack to-hit")
    }

    // attack twice
    combat.meleeAttack(attackingArmy.units[0], defendingArmy.units[0])
    combat.meleeAttack(attackingArmy.units[0], defendingArmy.units[0])

    if defendingArmy.units[0].GetCounterAttackToHit(attackingArmy.GetUnits()[0]) != 20 {
        test.Errorf("Error: defender should have 20%% counter attack to-hit")
    }
}

type OverrideToHitMelee struct {
    *units.OverworldUnit
}

func (unit *OverrideToHitMelee) GetToHitMelee() int {
    return 1000
}

func TestRangedAttack(test *testing.T){
    defendingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackerUnit := units.Slingers
    defenderUnit := units.LizardSwordsmen
    attackerUnit.RangedAttackPower = 38

    defender := units.MakeOverworldUnitFromUnit(defenderUnit, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    attacker := units.MakeOverworldUnitFromUnit(attackerUnit, 0, 0, data.PlaneArcanus, data.BannerGreen, &units.NoExperienceInfo{}, &units.NoEnchantments{})

    defendingArmy.AddUnit(defender)
    attackingArmy.AddUnit(&OverrideToHitMelee{attacker})

    combat := &CombatModel{
        SelectedUnit: nil,
        Tiles: makeTiles(30, 30, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
        Turn: TeamDefender,
        DefendingArmy: defendingArmy,
        AttackingArmy: attackingArmy,
    }

    combat.Initialize(spellbook.Spells{}, 0, 0)

    damage := attackingArmy.units[0].ComputeRangeDamage(defendingArmy.units[0], 1)

    if damage != attackerUnit.RangedAttackPower {
        test.Errorf("Error: ranged damage should be %d, got %d", attackerUnit.RangedAttackPower, damage)
    }
}

func TestAttackPower(test *testing.T){
    defendingArmy := Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackingArmy := Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attacker1 := units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})

    attackingArmy.AddUnit(attacker1)

    model := CombatModel{
        SelectedUnit: nil,
        Tiles: makeTiles(30, 30, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
        Turn: TeamDefender,
        DefendingArmy: &defendingArmy,
        AttackingArmy: &attackingArmy,
    }

    model.Initialize(spellbook.Spells{}, 0, 0)

    units1 := attackingArmy.units[0]
    if units1.GetMeleeAttackPower() != units.LizardSpearmen.MeleeAttackPower {
        test.Errorf("Error: melee attack power should be %d, got %d", units.LizardSpearmen.MeleeAttackPower, units1.GetMeleeAttackPower())
    }
}

func TestLeadershipBonus(test *testing.T){
    defendingArmy := Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackingArmy := Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    // melee only
    attacker1 := units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    // ranged attack
    ranged := units.MakeOverworldUnitFromUnit(units.LizardJavelineers, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    boulder := units.MakeOverworldUnitFromUnit(units.Catapult, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    thrown := units.MakeOverworldUnitFromUnit(units.Berserkers, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    fireBreath := units.MakeOverworldUnitFromUnit(units.DraconianSwordsmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    spirit := units.MakeOverworldUnitFromUnit(units.MagicSpirit, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    // valana has regular leadership
    leaderHero := herolib.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroValana, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{}), herolib.HeroValana, "Valana")
    leaderHero.AddExperience(units.ExperienceLord.ExperienceRequired(false, false))

    units1 := attackingArmy.AddUnit(attacker1)
    ranged1 := attackingArmy.AddUnit(ranged)
    boulder1 := attackingArmy.AddUnit(boulder)
    thrown1 := attackingArmy.AddUnit(thrown)
    fireBreath1 := attackingArmy.AddUnit(fireBreath)
    spirit1 := attackingArmy.AddUnit(spirit)
    valana := attackingArmy.AddUnit(leaderHero)

    model := CombatModel{
        SelectedUnit: nil,
        Tiles: makeTiles(30, 30, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
        Turn: TeamDefender,
        DefendingArmy: &defendingArmy,
        AttackingArmy: &attackingArmy,
    }

    model.Initialize(spellbook.Spells{}, 0, 0)

    leadershipBonus := 2

    if units1.GetMeleeAttackPower() != units.LizardSpearmen.MeleeAttackPower + leadershipBonus {
        test.Errorf("Error: melee attack power should be %d, got %d", units.LizardSpearmen.MeleeAttackPower + leadershipBonus, units1.GetMeleeAttackPower())
    }

    if ranged1.GetRangedAttackPower() != units.LizardJavelineers.RangedAttackPower + leadershipBonus / 2 {
        test.Errorf("Error: ranged attack power should be %d, got %d", units.LizardJavelineers.RangedAttackPower + leadershipBonus / 2, ranged1.GetRangedAttackPower())
    }

    if boulder1.GetRangedAttackPower() != units.Catapult.RangedAttackPower + leadershipBonus / 2 {
        test.Errorf("Error: boulder attack power should be %d, got %d", units.Catapult.RangedAttackPower + leadershipBonus / 2, boulder1.GetRangedAttackPower())
    }

    if int(thrown1.GetAbilityValue(data.AbilityThrown)) != int(units.Berserkers.GetAbilityValue(data.AbilityThrown)) + leadershipBonus / 2 {
        test.Errorf("Error: thrown attack power should be %d, got %d", int(units.Berserkers.GetAbilityValue(data.AbilityThrown)) + leadershipBonus / 2, int(thrown1.GetAbilityValue(data.AbilityThrown)))
    }

    if int(fireBreath1.GetAbilityValue(data.AbilityFireBreath)) != int(units.DraconianSwordsmen.GetAbilityValue(data.AbilityFireBreath)) + leadershipBonus / 2 {
        test.Errorf("Error: fire breath attack power should be %d, got %d", int(units.DraconianSwordsmen.GetAbilityValue(data.AbilityFireBreath)) + leadershipBonus / 2, int(fireBreath1.GetAbilityValue(data.AbilityFireBreath)))
    }

    if spirit1.GetMeleeAttackPower() != units.MagicSpirit.MeleeAttackPower {
        test.Errorf("Error: magic spirit melee attack power should be %d, got %d", units.MagicSpirit.MeleeAttackPower, spirit1.GetMeleeAttackPower())
    }

    // +5 for lord level, +2 for leadership
    if valana.GetMeleeAttackPower() != units.HeroValana.MeleeAttackPower + 5 + leadershipBonus {
        test.Errorf("Error: melee attack power should be %d, got %d", units.HeroValana.MeleeAttackPower + 5 + leadershipBonus, valana.GetMeleeAttackPower())
    }
}

// two heroes both with leadership
func TestLeadershipBonusMultiple(test *testing.T){
    defendingArmy := Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackingArmy := Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    // melee only
    attacker1 := units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    // valana has regular leadership
    hero1 := herolib.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroValana, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{}), herolib.HeroValana, "Valana")
    hero1.AddExperience(units.ExperienceLord.ExperienceRequired(false, false))

    hero2 := herolib.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroTorin, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{}), herolib.HeroTorin, "Torin")
    hero2.AddExperience(units.ExperienceLord.ExperienceRequired(false, false))

    units1 := attackingArmy.AddUnit(attacker1)
    valana := attackingArmy.AddUnit(hero1)
    torin := attackingArmy.AddUnit(hero2)

    model := CombatModel{
        SelectedUnit: nil,
        Tiles: makeTiles(30, 30, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
        Turn: TeamDefender,
        DefendingArmy: &defendingArmy,
        AttackingArmy: &attackingArmy,
    }

    model.Initialize(spellbook.Spells{}, 0, 0)

    leadershipBonus := 3

    if units1.GetMeleeAttackPower() != units.LizardSpearmen.MeleeAttackPower + leadershipBonus {
        test.Errorf("Error: spearmen melee attack power should be %d, got %d", units.LizardSpearmen.MeleeAttackPower + leadershipBonus, units1.GetMeleeAttackPower())
    }

    // +5 for lord level, +3 for leadership from torin
    if valana.GetMeleeAttackPower() != units.HeroValana.MeleeAttackPower + 5 + leadershipBonus {
        test.Errorf("Error: valana melee attack power should be %d, got %d", units.HeroValana.MeleeAttackPower + 5 + leadershipBonus, valana.GetMeleeAttackPower())
    }

    // base melee + 5 for lord, + 9 for supermight, + 3 for leadership from torin
    if torin.GetMeleeAttackPower() != units.HeroTorin.MeleeAttackPower + 5 + 9 + leadershipBonus {
        test.Errorf("Error: torin melee attack power should be %d, got %d", units.HeroTorin.MeleeAttackPower + 5 + 9 + leadershipBonus, torin.GetMeleeAttackPower())
    }

}

func TestHolyBonus(test *testing.T){
    defendingArmy := Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    attackingArmy := Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    // melee only
    attacker1 := units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    // valana has regular leadership

    // arch angel has holy bonus +2
    archAngelUnit := units.MakeOverworldUnitFromUnit(units.ArchAngel, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})

    units1 := attackingArmy.AddUnit(attacker1)
    units2 := attackingArmy.AddUnit(archAngelUnit)

    model := CombatModel{
        SelectedUnit: nil,
        Tiles: makeTiles(30, 30, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
        Turn: TeamDefender,
        DefendingArmy: &defendingArmy,
        AttackingArmy: &attackingArmy,
    }

    model.Initialize(spellbook.Spells{}, 0, 0)

    holyBonus := 2

    if units1.GetMeleeAttackPower() != units.LizardSpearmen.MeleeAttackPower + holyBonus {
        test.Errorf("Error: spearmen melee attack power should be %d, got %d", units.LizardSpearmen.MeleeAttackPower + holyBonus, units1.GetMeleeAttackPower())
    }

    if units1.GetDefense() != units.LizardSpearmen.Defense + holyBonus {
        test.Errorf("Error: spearmen defense should be %d, got %d", units.LizardSpearmen.Defense + holyBonus, units1.GetDefense())
    }

    if units1.GetResistance() != units.LizardSpearmen.Resistance + holyBonus {
        test.Errorf("Error: spearmen resistance should be %d, got %d", units.LizardSpearmen.Resistance + holyBonus, units1.GetResistance())
    }

    if units2.GetMeleeAttackPower() != units.ArchAngel.MeleeAttackPower + holyBonus {
        test.Errorf("Error: arch angel melee attack power should be %d, got %d", units.ArchAngel.MeleeAttackPower + holyBonus, units2.GetMeleeAttackPower())
    }

    if units2.GetDefense() != units.ArchAngel.Defense + holyBonus {
        test.Errorf("Error: arch angel defense should be %d, got %d", units.ArchAngel.Defense + holyBonus, units2.GetDefense())
    }

    if units2.GetResistance() != units.ArchAngel.Resistance + holyBonus {
        test.Errorf("Error: arch angel resistance should be %d, got %d", units.ArchAngel.Resistance + holyBonus, units2.GetResistance())
    }
}

type FakeDamageIndicator struct {}
func (f *FakeDamageIndicator) AddDamageIndicator(attacker *ArmyUnit, damage int){
}

func TestSpellEffects(test *testing.T){

    zeroResistance := func(unit units.Unit) units.Unit {
        unit.Resistance = 0
        return unit
    }

    highResistance := func(unit units.Unit) units.Unit {
        unit.Resistance = 1000
        return unit
    }

    lowDefense := func(unit units.Unit) units.Unit {
        unit.Defense = 0
        return unit
    }

    testEffect := func (unitBase units.Unit, doTest func(*CombatModel, *ArmyUnit)) {

        defendingArmy := Army{
            Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
        }

        attackingArmy := Army{
            Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
        }

        unit := units.MakeOverworldUnitFromUnit(unitBase, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
        armyUnit := defendingArmy.AddUnit(unit)

        model := &CombatModel{
            DefendingArmy: &defendingArmy,
            AttackingArmy: &attackingArmy,
            Tiles: makeTiles(1, 1, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
        }

        model.Initialize(spellbook.Spells{}, 0, 0)

        doTest(model, armyUnit)
    }

    testEffect(zeroResistance(units.LizardSpearmen), func (model *CombatModel, unit *ArmyUnit) {
        model.CreateBlackSleepProjectileEffect(0)(unit)
        if !unit.HasCurse(data.UnitCurseBlackSleep) {
            test.Errorf("Error: unit should have black sleep curse")
        }

    })

    testEffect(highResistance(units.LizardSpearmen), func (model *CombatModel, unit *ArmyUnit) {
        model.CreateMindStormProjectileEffect()(unit)
        if !unit.HasCurse(data.UnitCurseMindStorm) {
            test.Errorf("Error: unit should have black sleep curse")
        }
    })

    testEffect(zeroResistance(units.LizardSpearmen), func (model *CombatModel, unit *ArmyUnit) {
        model.CreateBanishProjectileEffect(0, &FakeDamageIndicator{})(unit)
        if unit.GetHealth() != 0 {
            test.Errorf("Error: unit should be banished")
        }
    })

    testEffect(lowDefense(units.LizardSpearmen), func (model *CombatModel, unit *ArmyUnit) {
        model.CreateIceBoltProjectileEffect(10000, &FakeDamageIndicator{})(unit)
        if unit.GetHealth() != 0 {
            test.Errorf("Error: unit should be killed by ice bolt")
        }
    })

    /*
func (model *CombatModel) CreateFireBoltProjectileEffect(strength int, damageIndicator AddDamageIndicators) func(*ArmyUnit) {
func (model *CombatModel) CreateFireballProjectileEffect(strength int, damageIndicator AddDamageIndicators) func(*ArmyUnit) {
func (model *CombatModel) CreateStarFiresProjectileEffect(damageIndicator AddDamageIndicators) func(*ArmyUnit) {
func (model *CombatModel) CreateDispelEvilProjectileEffect(damageIndicator AddDamageIndicators) func(*ArmyUnit) {
func (model *CombatModel) CreatePsionicBlastProjectileEffect(strength int, damageIndicator AddDamageIndicators) func(*ArmyUnit) {
func (model *CombatModel) CreateDoomBoltProjectileEffect(damageIndicator AddDamageIndicators) func(*ArmyUnit) {
func (model *CombatModel) CreateLightningBoltProjectileEffect(strength int, damageIndicator AddDamageIndicators) func(*ArmyUnit) {
func (model *CombatModel) CreateWarpLightningProjectileEffect(damageIndicator AddDamageIndicators) func(*ArmyUnit) {
func (model *CombatModel) CreateLifeDrainProjectileEffect(reduceResistance int, player ArmyPlayer, unitCaster *ArmyUnit, damageIndicator AddDamageIndicators) func(*ArmyUnit) {
func (model *CombatModel) CreateFlameStrikeProjectileEffect(damageIndicator AddDamageIndicators) func(*ArmyUnit) {
func (model *CombatModel) CreateRecallHeroProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateHealingProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateHeroismProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateHolyArmorProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateInvulnerabilityProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateLionHeartProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateTrueSightProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateElementalArmorProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateGiantStrengthProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateIronSkinProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateStoneSkinProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateRegenerationProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateResistElementsProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateRighteousnessProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateHolyWeaponProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateFlightProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateGuardianWindProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateHasteProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateInvisibilityProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateMagicImmunityProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateResistMagicProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateSpellLockProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateEldritchWeaponProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateFlameBladeProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateImmolationProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateBerserkProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateCloakOfFearProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateWraithFormProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateChaosChannelsProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateBlessProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateWeaknessProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateVertigoProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateShatterProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateWarpCreatureProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateConfusionProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreatePossessionProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateCreatureBindingProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreatePetrifyProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateHolyWordProjectileEffect(damageIndicator AddDamageIndicators) func(*ArmyUnit) {
func (model *CombatModel) CreateWebProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateDeathSpellProjectileEffect(damageIndicator AddDamageIndicators) func(*ArmyUnit) {
func (model *CombatModel) CreateWordOfDeathProjectileEffect(damageIndicator AddDamageIndicators) func(*ArmyUnit) {
func (model *CombatModel) CreateWarpWoodProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateDisintegrateProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateWordOfRecallProjectileEffect() func(*ArmyUnit) {
func (model *CombatModel) CreateDispelMagicProjectileEffect(caster ArmyPlayer, dispelStrength int) func(*ArmyUnit) {
func (model *CombatModel) CreateCracksCallProjectileEffect() func(*ArmyUnit) {
     */

}

type noGlobalEnchantments struct {
}

func (*noGlobalEnchantments) HasEnchantment(enchantment data.Enchantment) bool {
    return false
}

func (*noGlobalEnchantments) HasRivalEnchantment(player *playerlib.Player, enchantment data.Enchantment) bool {
    return false
}

// run a full combat simulation
func TestFullCombat(test *testing.T){
    defendingPlayer := playerlib.MakePlayer(setup.WizardCustom{
        Name: "AI-1",
        Banner: data.BannerBrown,
    }, false, 0, 0, nil, &noGlobalEnchantments{})

    attackingPlayer := playerlib.MakePlayer(setup.WizardCustom{
        Name: "AI-2",
        Banner: data.BannerRed,
    }, false, 0, 0, nil, &noGlobalEnchantments{})

    attackingArmy := &Army{
        Player: attackingPlayer,
    }

    defendingArmy := &Army{
        Player: defendingPlayer,
    }

    for range 3 {
        attackingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.GreatDrake, 1, 1, data.PlaneArcanus, attackingPlayer.Wizard.Banner, attackingPlayer.MakeExperienceInfo(), attackingPlayer.MakeUnitEnchantmentProvider()))
    }

    for range 3 {
        defendingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.LizardSwordsmen, 1, 1, data.PlaneArcanus, defendingPlayer.Wizard.Banner, defendingPlayer.MakeExperienceInfo(), defendingPlayer.MakeUnitEnchantmentProvider()))
    }

    var allSpells spellbook.Spells

    model := MakeCombatModel(allSpells, defendingArmy, attackingArmy, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10), nil)

    state := Run(model)
    if state != CombatStateAttackerWin {
        test.Errorf("Error: attacker should have won the combat, got state %d", state)
    }
}

func TestInvisibleEnemy(test *testing.T) {
    // create an ememy army with an invisible unit
    // should not be able to target it with a spell, unless we have a unit
    // with illusion immunity

    check := func(attackerUnit units.Unit) bool {
        defendingPlayer := playerlib.MakePlayer(setup.WizardCustom{
            Name: "AI-1",
            Banner: data.BannerBrown,
        }, false, 0, 0, nil, &noGlobalEnchantments{})

        attackingPlayer := playerlib.MakePlayer(setup.WizardCustom{
            Name: "AI-2",
            Banner: data.BannerRed,
        }, false, 0, 0, nil, &noGlobalEnchantments{})

        attackingArmy := &Army{
            Player: attackingPlayer,
        }

        defendingArmy := &Army{
            Player: defendingPlayer,
        }

        defendingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.NightStalker, 1, 1, data.PlaneArcanus, defendingPlayer.Wizard.Banner, defendingPlayer.MakeExperienceInfo(), defendingPlayer.MakeUnitEnchantmentProvider()))

        attackingArmy.AddUnit(units.MakeOverworldUnitFromUnit(attackerUnit, 1, 1, data.PlaneArcanus, attackingPlayer.Wizard.Banner, attackingPlayer.MakeExperienceInfo(), attackingPlayer.MakeUnitEnchantmentProvider()))

        var allSpells spellbook.Spells

        model := MakeCombatModel(allSpells, defendingArmy, attackingArmy, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10), nil)

        targeted := false
        onTarget := func(unit *ArmyUnit){
            targeted = true
        }

        model.DoAITargetUnitSpell(attackingPlayer, spellbook.Spell{Name: "whatever"}, TeamAttacker, TeamDefender, func(_ *ArmyUnit) bool { return true }, onTarget)
        return targeted
    }


    if check(units.Warlocks) {
        test.Errorf("Error: should not be able to target invisible unit")
    }

    if !check(units.Angel) {
        test.Errorf("Error: should be able to target invisble unit")
    }
}

func TestSpellSkillItemBonus(test *testing.T) {
    defendingPlayer := playerlib.MakePlayer(setup.WizardCustom{
        Name: "AI-1",
        Banner: data.BannerBrown,
    }, false, 0, 0, nil, &noGlobalEnchantments{})

    attackingPlayer := playerlib.MakePlayer(setup.WizardCustom{
        Name: "AI-2",
        Banner: data.BannerRed,
    }, false, 0, 0, nil, &noGlobalEnchantments{})

    attackingArmy := &Army{
        Player: attackingPlayer,
    }

    defendingArmy := &Army{
        Player: defendingPlayer,
    }

    defendingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.ArchAngel, 1, 1, data.PlaneArcanus, defendingPlayer.GetBanner(), defendingPlayer.MakeExperienceInfo(), defendingPlayer.MakeUnitEnchantmentProvider()))

    hero := herolib.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroValana, 0, 0, data.PlaneArcanus, attackingPlayer.GetBanner(), attackingPlayer.MakeExperienceInfo(), attackingPlayer.MakeUnitEnchantmentProvider()), herolib.HeroValana, "Valana")
    hero.Equipment[0] = &artifact.Artifact{
        Name: "whatever",
        Image: 7,
        Type: artifact.ArtifactTypeSword,
        Powers: []artifact.Power{
            {
                Type: artifact.PowerTypeSpellSkill,
                Amount: 10,
                Name: "Spell Skill +10",
            },
        },
    }

    attackingArmy.AddUnit(hero)

    var allSpells spellbook.Spells

    model := MakeCombatModel(allSpells, defendingArmy, attackingArmy, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10), nil)
    _ = model

    // hero should get base (5) + item (10) = 15 spell skill
    attacker1 := attackingArmy.units[0]
    if attacker1.CastingSkill != 15 {
        test.Errorf("Error: spell skill should be 15, got %v", attacker1.CastingSkill)
    }

    // arch angel has 40
    defender1 := defendingArmy.units[0]
    if defender1.CastingSkill != 40 {
        test.Errorf("Error: spell skill should be 40, got %v", defender1.CastingSkill)
    }

}

func TestSpellSavePower(test *testing.T) {
    defendingPlayer := playerlib.MakePlayer(setup.WizardCustom{
        Name: "AI-1",
        Banner: data.BannerBrown,
    }, false, 0, 0, nil, &noGlobalEnchantments{})

    attackingPlayer := playerlib.MakePlayer(setup.WizardCustom{
        Name: "AI-2",
        Banner: data.BannerRed,
    }, false, 0, 0, nil, &noGlobalEnchantments{})

    attackingArmy := &Army{
        Player: attackingPlayer,
    }

    defendingArmy := &Army{
        Player: defendingPlayer,
    }

    // normally colossus is immune to banish
    defender := defendingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.Colossus, 1, 1, data.PlaneArcanus, defendingPlayer.GetBanner(), defendingPlayer.MakeExperienceInfo(), defendingPlayer.MakeUnitEnchantmentProvider()))

    hero := herolib.MakeHero(units.MakeOverworldUnitFromUnit(units.HeroValana, 0, 0, data.PlaneArcanus, attackingPlayer.GetBanner(), attackingPlayer.MakeExperienceInfo(), attackingPlayer.MakeUnitEnchantmentProvider()), herolib.HeroValana, "Valana")
    hero.Equipment[0] = &artifact.Artifact{
        Name: "whatever",
        Image: 7,
        Type: artifact.ArtifactTypeSword,
        Powers: []artifact.Power{
            {
                Type: artifact.PowerTypeSpellSave,
                Amount: 20,
                Name: "-20 Spell Save",
            },
        },
    }

    attackingUnit := attackingArmy.AddUnit(hero)

    var allSpells spellbook.Spells

    model := MakeCombatModel(allSpells, defendingArmy, attackingArmy, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10), nil)

    casted := false
    model.InvokeSpell(&ProxySpellSystem{Model: model}, attackingArmy, attackingUnit, spellbook.Spell{Name: "Banish"}, func(success bool) {
        casted = true
    })

    if !casted {
        test.Errorf("Error: spell should have been cast")
    }

    for _, projectile := range model.Projectiles {
        projectile.Effect(projectile.Target)
    }

    if defender.GetHealth() > 0 {
        test.Errorf("Error: defender should have been banished")
    }
}

func makeTestCombatPlayer(human bool) *playerlib.Player {
    return playerlib.MakePlayer(setup.WizardCustom{}, human, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{})
}

type noopAIActions struct{}

func (noopAIActions) RangeAttack(*ArmyUnit, RangeTarget) {}
func (noopAIActions) MeleeAttack(*ArmyUnit, *ArmyUnit) {}
func (noopAIActions) MeleeAttackWall(*ArmyUnit, int, int) {}
func (noopAIActions) MoveMagicVortex(*MagicVortex, pathfinding.Path) {}
func (noopAIActions) MoveUnit(*ArmyUnit, pathfinding.Path) {}
func (noopAIActions) Teleport(*ArmyUnit, int, int, bool) {}
func (noopAIActions) DoProjectiles() {}

func TestLayoutUnitsStayInsideMap(test *testing.T) {
    player := makeTestCombatPlayer(false)
    model := &CombatModel{
        Tiles: makeTiles(30, 30, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
    }

    for _, team := range []Team{TeamAttacker, TeamDefender} {
        army := &Army{Player: player}
        for range 200 {
            army.AddUnit(units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{}))
        }
        army.LayoutUnits(team, model)
        for _, unit := range army.units {
            if !model.IsInsideMap(unit.X, unit.Y) {
                test.Fatalf("layout placed unit off the combat map at %d,%d", unit.X, unit.Y)
            }
        }
    }
}

func TestAIMovementPathfindingOffMapUnit(test *testing.T) {
    defendingArmy := &Army{Player: makeTestCombatPlayer(false)}
    attackingArmy := &Army{Player: makeTestCombatPlayer(false)}

    defendingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{}))
    attackingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{}))

    model := &CombatModel{
        Tiles: makeTiles(30, 30, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
        DefendingArmy: defendingArmy,
        AttackingArmy: attackingArmy,
    }
    model.Initialize(spellbook.Spells{}, 0, 0)

    defender := defendingArmy.units[0]
    attacker := attackingArmy.units[0]
    defender.X = 30
    defender.Y = 15
    attacker.X = 10
    attacker.Y = 10
    attacker.MovesLeft = fraction.FromInt(2)
    model.Tiles[attacker.Y][attacker.X].Unit = attacker

    doAIMovementPathfinding(model, noopAIActions{}, attacker, defendingArmy)
}

func TestCombatOffMapTileAccessDoesNotPanic(test *testing.T) {
    defendingArmy := &Army{Player: makeTestCombatPlayer(false)}
    attackingArmy := &Army{Player: makeTestCombatPlayer(false)}

    defendingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{}))
    attackingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{}))

    model := &CombatModel{
        Tiles: makeTiles(30, 30, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
        DefendingArmy: defendingArmy,
        AttackingArmy: attackingArmy,
    }
    model.Initialize(spellbook.Spells{}, 0, 0)

    defender := defendingArmy.units[0]
    attacker := attackingArmy.units[0]
    defender.X = 30
    defender.Y = 15
    attacker.X = 10
    attacker.Y = 10
    attacker.MovesLeft = fraction.FromInt(2)

    model.ComputeWallDefense(attacker, defender)
    model.canMeleeAttack(attacker, defender, true)
    model.DestroyWall(30, 15)
    model.MoveUnit(attacker, 30, 15)
    attacker.X = 30
    attacker.Y = 10
    model.Teleport(attacker, 5, 5)
    model.KillUnit(defender)
    model.RemoveUnit(attacker)
}

func TestRemoteRangeDamage(test *testing.T) {

    var allSpells spellbook.Spells

    peer1, peer2 := net.Pipe()

    defer peer1.Close()
    defer peer2.Close()

    // remote side is defender, so false for isAttacker
    remote := MakeRemote(false, false, peer1)

    defendingArmy := &Army{Player: makeTestCombatPlayer(false)}
    attackingArmy := &Army{Player: makeTestCombatPlayer(false)}

    defendingArmy.AddUnit(units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{}))

    // attacking army has a unit that can range attack. grant very high tohit
    attacker := units.MakeOverworldUnitFromUnit(units.Warlocks, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    attackingUnit := attackingArmy.AddUnit(&OverrideToHitMelee{attacker})

    model := MakeCombatModel(allSpells, defendingArmy, attackingArmy, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10), remote)
    effect := model.CreateRangeAttackEffect(attackingUnit, &FakeDamageIndicator{})

    quit, cancel := context.WithCancel(context.Background())
    defer cancel()

    // remote side is attacker, so true for isAttacker
    remoteDefender := MakeRemote(true, true, peer2)

    go remoteDefender.RunReceiveLoop(quit)

    remoteDidDamage := false
    finish := make(chan struct{})

    go func() {
        for quit.Err() == nil {
            select {
                case <-quit.Done():
                    return
                case event := <-remoteDefender.Events:
                    log.Printf("remote received event: %v", event)
                    switch event.GetType() {
                        case RemoteDamageType:
                            event := event.(*RemoteDamageEvent)
                            if event.Id == defendingArmy.units[0].Id && event.Damage > 0 {
                                remoteDidDamage = true
                                close(finish)
                            }

                    }
            }
        }
    }()

    // invoking the effect should send a damage event to the remote
    effect(defendingArmy.units[0])

    select {
        case <-time.After(500 * time.Millisecond):
            test.Errorf("Error: remote did not receive damage event")
        case <-finish:
            if !remoteDidDamage {
                test.Errorf("Error: remote did not receive damage event")
            }
    }

}

type proxySpellSystem struct {
    TestSpellSystem
}

func makeSpellSystem(model *CombatModel) proxySpellSystem {
    makeProjectile := func (target *ArmyUnit, effect func(*ArmyUnit)) *Projectile {
        // a projectile that explodes immediately and invokes the effect
        return &Projectile{
            Animation: util.MakeAnimation(nil, true),
            Explode: util.MakeAnimation(nil, false),
            Exploding: true,
            Target: target,
            Effect: effect,
        }
    }

    return proxySpellSystem{
        createFireballProjectile: func(target *ArmyUnit, cost int) *Projectile {
            return makeProjectile(target, model.CreateFireballProjectileEffect(1000, &FakeDamageIndicator{}))
        },
        createIceBoltProjectile: func(target *ArmyUnit, cost int) *Projectile {
            return makeProjectile(target, model.CreateIceBoltProjectileEffect(1000, &FakeDamageIndicator{}))
        },
        createStarFiresProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateStarFiresProjectileEffect(&FakeDamageIndicator{}))
        },
        createPsionicBlastProjectile: func(target *ArmyUnit, cost int) *Projectile {
            return makeProjectile(target, model.CreatePsionicBlastProjectileEffect(1000, &FakeDamageIndicator{}))
        },
        createDoomBoltProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateDoomBoltProjectileEffect(&FakeDamageIndicator{}))
        },
        createFireBoltProjectile: func(target *ArmyUnit, cost int) *Projectile {
            return makeProjectile(target, model.CreateFireBoltProjectileEffect(1000, &FakeDamageIndicator{}))
        },
        createLightningBoltProjectile: func(target *ArmyUnit, cost int) *Projectile {
            return makeProjectile(target, model.CreateLightningBoltProjectileEffect(1000, &FakeDamageIndicator{}))
        },
        createWarpLightningProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateWarpLightningProjectileEffect(&FakeDamageIndicator{}))
        },
        createLifeDrainProjectile: func(target *ArmyUnit, cost int, player ArmyPlayer, unitCaster *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateLifeDrainProjectileEffect(1000, player, unitCaster, &FakeDamageIndicator{}))
        },
        createDispelEvilProjectile: func(target *ArmyUnit, reduce int) *Projectile {
            return makeProjectile(target, model.CreateDispelEvilProjectileEffect(&FakeDamageIndicator{}, reduce))
        },
        createHealingProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateHealingProjectileEffect())
        },
        createCracksCallProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateCracksCallProjectileEffect())
        },
        createWebProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateWebProjectileEffect())
        },
        createBanishProjectile: func(target *ArmyUnit, reduce int) *Projectile {
            return makeProjectile(target, model.CreateBanishProjectileEffect(reduce, &FakeDamageIndicator{}))
        },
        createDispelMagicProjectile: func(target *ArmyUnit, caster ArmyPlayer, strength int) *Projectile {
            return makeProjectile(target, model.CreateDispelMagicProjectileEffect(caster, strength + 10000))
        },
        createDisintegrateProjectile: func(target *ArmyUnit, reduce int) *Projectile {
            return makeProjectile(target, model.CreateDisintegrateProjectileEffect(reduce))
        },
        createWarpWoodProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateWarpWoodProjectileEffect())
        },
        createWordOfDeathProjectile: func(target *ArmyUnit, reduce int) *Projectile {
            return makeProjectile(target, model.CreateWordOfDeathProjectileEffect(&FakeDamageIndicator{}, 100))
        },
        createCreatureBindingProjectile: func(target *ArmyUnit, reduce int) *Projectile {
            return makeProjectile(target, model.CreateCreatureBindingProjectileEffect(reduce + 100))
        },
        createMindStormProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateMindStormProjectileEffect())
        },
        createBlessProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateBlessProjectileEffect())
        },
        createWeaknessProjectile: func(target *ArmyUnit, reduce int) *Projectile {
            return makeProjectile(target, model.CreateWeaknessProjectileEffect(reduce + 100))
        },
        createBlackSleepProjectile: func(target *ArmyUnit, reduce int) *Projectile {
            return makeProjectile(target, model.CreateBlackSleepProjectileEffect(reduce + 100))
        },
        createVertigoProjectile: func(target *ArmyUnit, reduce int) *Projectile {
            return makeProjectile(target, model.CreateVertigoProjectileEffect(reduce + 100))
        },
        createShatterProjectile: func(target *ArmyUnit, reduce int) *Projectile {
            return makeProjectile(target, model.CreateShatterProjectileEffect(reduce + 100))
        },
        createWarpCreatureProjectile: func(target *ArmyUnit, reduce int) *Projectile {
            return makeProjectile(target, model.CreateWarpCreatureProjectileEffect(reduce + 100))
        },
        createConfusionProjectile: func(target *ArmyUnit, reduce int) *Projectile {
            return makeProjectile(target, model.CreateConfusionProjectileEffect(reduce + 100))
        },
        createPossessionProjectile: func(target *ArmyUnit, reduce int) *Projectile {
            return makeProjectile(target, model.CreatePossessionProjectileEffect(reduce + 100))
        },
        createPetrifyProjectile: func(target *ArmyUnit, reduce int) *Projectile {
            return makeProjectile(target, model.CreatePetrifyProjectileEffect(&FakeDamageIndicator{}, reduce + 100))
        },
        createHeroismProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateHeroismProjectileEffect())
        },
        createHolyArmorProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateHolyArmorProjectileEffect())
        },
        createHolyWeaponProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateHolyWeaponProjectileEffect())
        },
        createInvulnerabilityProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateInvulnerabilityProjectileEffect())
        },
        createLionHeartProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateLionHeartProjectileEffect())
        },
        createRighteousnessProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateRighteousnessProjectileEffect())
        },
        createTrueSightProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateTrueSightProjectileEffect())
        },
        createElementalArmorProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateElementalArmorProjectileEffect())
        },
        createGiantStrengthProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateGiantStrengthProjectileEffect())
        },
        createIronSkinProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateIronSkinProjectileEffect())
        },
        createRegenerationProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateRegenerationProjectileEffect())
        },
        createResistElementsProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateResistElementsProjectileEffect())
        },
        createStoneSkinProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateStoneSkinProjectileEffect())
        },
        createFlightProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateFlightProjectileEffect())
        },
        createGuardianWindProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateGuardianWindProjectileEffect())
        },
        createHasteProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateHasteProjectileEffect())
        },
        createInvisibilityProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateInvisibilityProjectileEffect())
        },
        createMagicImmunityProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateMagicImmunityProjectileEffect())
        },
        createResistMagicProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateResistMagicProjectileEffect())
        },
        createSpellLockProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateSpellLockProjectileEffect())
        },
        createEldritchWeaponProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateEldritchWeaponProjectileEffect())
        },
        createFlameBladeProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateFlameBladeProjectileEffect())
        },
        createImmolationProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateImmolationProjectileEffect())
        },
        createBerserkProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateBerserkProjectileEffect())
        },
        createCloakOfFearProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateCloakOfFearProjectileEffect())
        },
        createWraithFormProjectile: func(target *ArmyUnit) *Projectile {
            return makeProjectile(target, model.CreateWraithFormProjectileEffect())
        },
        createDeathSpellProjectile: func(target *ArmyUnit, reduce int) *Projectile {
            return makeProjectile(target, model.CreateDeathSpellProjectileEffect(&FakeDamageIndicator{}, reduce + 100))
        },
        createHolyWordProjectile: func(target *ArmyUnit, reduce int) *Projectile {
            return makeProjectile(target, model.CreateHolyWordProjectileEffect(&FakeDamageIndicator{}, reduce + 100))
        },
    }
}

func TestRemoteUnitCastProjectile(test *testing.T) {

    type Expecter struct {
        HandleEvent func(RemoteEvent, uint64, string, *CombatModel)
        Assertions func(*testing.T, string)
        Finished func() bool
    }

    isAttacker := func(id uint64, model *CombatModel) bool {
        unit := model.GetUnitById(id)
        if unit == nil {
            return false
        }

        army := model.GetArmy(unit)

        return model.GetTeamForArmy(army) == TeamAttacker
    }

    type TestOptions struct {
        MakeUnit func() *units.OverworldUnit
    }

    doSpellTest := func(spellName string, defender bool, expecter Expecter, testOptions... TestOptions) {
        log.Printf("== Testing spell: %s", spellName)
        var allSpells spellbook.Spells

        peer1, peer2 := net.Pipe()

        defer peer1.Close()
        defer peer2.Close()

        // the attacker's remote, which represents the defender, so false for isAttacker
        remote := MakeRemote(false, false, peer1)

        makeDefendingArmy := func() *Army {
            defendingArmy := &Army{Player: makeTestCombatPlayer(false)}

            if len(testOptions) == 0 {
                // default is hell hounds: fantastic chaos unit
                defendingOverworld := units.MakeOverworldUnitFromUnit(units.HellHounds, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})

                defendingUnit := defendingArmy.AddUnit(defendingOverworld)

                // an enchantment that can be removed via dispel magic
                defendingUnit.AddEnchantment(data.UnitEnchantmentGiantStrength)
            } else {

                // allow a test to override defending unit
                for _, option := range testOptions {
                    if option.MakeUnit != nil {
                        defendingOverworld := option.MakeUnit()
                        defendingUnit := defendingArmy.AddUnit(defendingOverworld)
                        defendingUnit.AddEnchantment(data.UnitEnchantmentGiantStrength)
                    }
                }

            }

            return defendingArmy
        }

        makeAttackingArmy := func() *Army {
            attackingArmy := &Army{Player: makeTestCombatPlayer(false)}

            // attacking army has a unit that can range attack. grant very high tohit
            attacker := units.MakeOverworldUnitFromUnit(units.Warlocks, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
            attackingArmy.AddUnit(&OverrideToHitMelee{attacker})

            // a different unit that can be healed
            attacker2 := units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
            attackingFriend := attackingArmy.AddUnit(attacker2)
            // set health to half so we can test healing
            attackingFriend.TakeDamage(attackingFriend.GetMaxHealth() / 2, DamageNormal)
            return attackingArmy
        }

        attackerModel := MakeCombatModel(allSpells, makeDefendingArmy(), makeAttackingArmy(), CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10), remote)
        attackerModel.CracksCallChance = 100

        // the defender's remote, which represents the attacker, so true for isAttacker
        remoteDefender := MakeRemote(true, true, peer2)

        defenderModel := MakeCombatModel(allSpells, makeDefendingArmy(), makeAttackingArmy(), CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10), remoteDefender)
        // probably don't really need this
        defenderModel.CracksCallChance = 100

        quit, cancel := context.WithCancel(context.Background())
        defer cancel()

        var wait sync.WaitGroup

        wait.Go(func(){ remoteDefender.RunReceiveLoop(quit) })

        finish := make(chan struct{})

        expectedId := defenderModel.DefendingArmy.GetUnits()[0].Id
        if !defender {
            // a bit of a hack, but when casting a friendly spell we target the second unit
            expectedId = attackerModel.AttackingArmy.GetUnits()[1].Id
        }

        wait.Go(func() {
            for quit.Err() == nil {
                select {
                    case <-quit.Done():
                        return
                    case event := <-remoteDefender.Events:
                        expecter.HandleEvent(event, expectedId, spellName, defenderModel)
                }

                if expecter.Finished() {
                    close(finish)
                    break
                }
            }
        })

        spellSystem := makeSpellSystem(attackerModel)

        attackerModel.InvokeSpell(&spellSystem, attackerModel.AttackingArmy, attackerModel.AttackingArmy.GetUnits()[0], spellbook.Spell{Name: spellName}, func(success bool) { })

        var counter uint64
        for attackerModel.UpdateProjectiles(counter, &FakeDamageIndicator{}, &spellSystem) && counter < 10000 {
            counter += 1
        }

        // now the remote side should receive a spell invocation event

        // invoking the effect should send a damage event to the remote
        // effect(defendingArmy.units[0])

        select {
            case <-time.After(500 * time.Millisecond):
            case <-finish:
        }

        expecter.Assertions(test, spellName)

        cancel()
        wait.Wait()
    }

    enemyDamageUnitSpells := []string{
        "Fireball", "Ice Bolt", "Star Fires",
        "Psionic Blast", "Doom Bolt", "Fire Bolt",
        "Lightning Bolt", "Warp Lightning", "Life Drain",
        "Dispel Evil", "Banish", "Word of Death",
        "Petrify",
    }

    friendlyUnitSpells := []string{
        "Healing",
    }

    makeHarmExpecter := func() Expecter {
        didCastSpell := false
        didDamage := false
        didFinishProjectile := false

        var harm Expecter
        harm.HandleEvent = func(event RemoteEvent, expectedId uint64, spellName string, model *CombatModel) {
            switch event.GetType() {
                case RemoteUnitTargetSpellType:
                    event := event.(*RemoteUnitTargetSpellEvent)
                    if event.TargetId == expectedId && event.Spell == spellName {
                        didCastSpell = true
                    }
                case RemoteDamageType:
                    event := event.(*RemoteDamageEvent)
                    if event.Id == expectedId {
                        didDamage = true
                    }
                case RemoteHealType:
                    event := event.(*RemoteHealEvent)
                    if event.Id == expectedId && event.Heal > 0 {
                        // not really damage but we can treat it as such for the purposes of this test
                        didDamage = true
                    }
                case RemoteProjectileFinishedType:
                    didFinishProjectile = true
            }
        }

        harm.Finished = func() bool {
            return didCastSpell && didDamage && didFinishProjectile
        }

        harm.Assertions = func(test *testing.T, spellName string) {
            if !didDamage {
                test.Errorf("Error: remote did not receive damage event for %v", spellName)
            }

            if !didCastSpell {
                test.Errorf("Error: remote did not receive cast spell for %v", spellName)
            }

            if !didFinishProjectile {
                test.Errorf("Error: remote did not receive projectile finished event for %v", spellName)
            }
        }

        return harm
    }

    for _, spell := range enemyDamageUnitSpells {
        doSpellTest(spell, true, makeHarmExpecter())
    }

    for _, spell := range friendlyUnitSpells {
        doSpellTest(spell, false, makeHarmExpecter())
    }

    makeRemoveUnitExpecter := func() Expecter {
        didCastSpell := false
        didRemoveUnit := false
        didFinishProjectile := false

        var remove Expecter
        remove.HandleEvent = func(event RemoteEvent, expectedId uint64, spellName string, model *CombatModel) {
            switch event.GetType() {
                case RemoteUnitTargetSpellType:
                    event := event.(*RemoteUnitTargetSpellEvent)
                    if event.TargetId == expectedId && event.Spell == spellName {
                        didCastSpell = true
                    }
                case RemoteRemoveUnitType:
                    event := event.(*RemoteRemoveUnitEvent)
                    if event.Id == expectedId {
                        didRemoveUnit = true
                    }
                case RemoteProjectileFinishedType:
                    didFinishProjectile = true
            }
        }

        remove.Finished = func() bool {
            return didCastSpell && didRemoveUnit && didFinishProjectile
        }

        remove.Assertions = func(test *testing.T, spellName string) {
            if !didCastSpell {
                test.Errorf("Error: remote did not receive cast spell for %v", spellName)
            }

            if !didRemoveUnit {
                test.Errorf("Error: remote did not receive remove unit event for %v", spellName)
            }

            if !didFinishProjectile {
                test.Errorf("Error: remote did not receive projectile finished event for %v", spellName)
            }
        }

        return remove
    }

    doSpellTest("Cracks Call", true, makeRemoveUnitExpecter())
    doSpellTest("Disintegrate", true, makeRemoveUnitExpecter())

    type CurseUnitExpecter struct {
        Expecter
        DidCurse bool
        DidFinishProjectile bool
    }

    makeCurseUnitExpecter := func(anyCurse... data.UnitEnchantment) Expecter {
        didCastSpell := false
        didCurse := false
        didFinishProjectile := false

        var expect Expecter
        expect.HandleEvent = func(event RemoteEvent, expectedId uint64, spellName string, model *CombatModel) {
            switch event.GetType() {
                case RemoteUnitTargetSpellType:
                    event := event.(*RemoteUnitTargetSpellEvent)
                    if event.TargetId == expectedId && event.Spell == spellName {
                        didCastSpell = true
                    }
                case RemoteCurseUnitType:
                    event := event.(*RemoteCurseUnitEvent)
                    if event.Id == expectedId && slices.Contains(anyCurse, event.Curse) {
                        didCurse = true
                    }
                case RemoteProjectileFinishedType:
                    didFinishProjectile = true
            }
        }

        expect.Finished = func() bool {
            return didCastSpell && didCurse && didFinishProjectile
        }

        expect.Assertions = func(test *testing.T, spellName string) {
            if !didCastSpell {
                test.Errorf("Error: remote did not receive cast spell for %v", spellName)
            }

            if !didCurse {
                test.Errorf("Error: remote did not receive curse unit event for %v", spellName)
            }

            if !didFinishProjectile {
                test.Errorf("Error: remote did not receive projectile finished event for %v", spellName)
            }
        }

        return expect
    }

    doSpellTest("Web", true, makeCurseUnitExpecter(data.UnitCurseWeb))
    doSpellTest("Mind Storm", true, makeCurseUnitExpecter(data.UnitCurseMindStorm))
    doSpellTest("Weakness", true, makeCurseUnitExpecter(data.UnitCurseWeakness))
    doSpellTest("Black Sleep", true, makeCurseUnitExpecter(data.UnitCurseBlackSleep))
    doSpellTest("Vertigo", true, makeCurseUnitExpecter(data.UnitCurseVertigo))
    doSpellTest("Shatter", true, makeCurseUnitExpecter(data.UnitCurseShatter), TestOptions{
        MakeUnit: func() *units.OverworldUnit {
            return units.MakeOverworldUnitFromUnit(units.OrcBowmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
        },
    })
    doSpellTest("Warp Creature", true, makeCurseUnitExpecter(data.UnitCurseWarpCreatureMelee, data.UnitCurseWarpCreatureDefense, data.UnitCurseWarpCreatureResistance), TestOptions{
        MakeUnit: func() *units.OverworldUnit {
            // need a non-fantastic unit
            return units.MakeOverworldUnitFromUnit(units.OrcBowmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
        },
    })
    doSpellTest("Confusion", true, makeCurseUnitExpecter(data.UnitCurseConfusion))

    makeRemoveEnchantmentExpecter := func() Expecter {
        var expect Expecter

        didCastSpell := false
        didFinishProjectile := false
        didRemoveEnchantment := false

        expect.HandleEvent = func(event RemoteEvent, expectedId uint64, spellName string, model *CombatModel) {
            switch event.GetType() {
                case RemoteUnitTargetSpellType:
                    event := event.(*RemoteUnitTargetSpellEvent)
                    if event.TargetId == expectedId && event.Spell == spellName {
                        didCastSpell = true
                    }
                case RemoteRemoveUnitEnchantmentType:
                    event := event.(*RemoteRemoveUnitEnchantmentEvent)
                    if event.Id == expectedId {
                        didRemoveEnchantment = true
                    }
                case RemoteProjectileFinishedType:
                    didFinishProjectile = true
            }
        }

        expect.Finished = func() bool {
            return didCastSpell && didFinishProjectile && didRemoveEnchantment
        }

        expect.Assertions = func(test *testing.T, spellName string) {
            if !didCastSpell {
                test.Errorf("Error: remote did not receive cast spell for %v", spellName)
            }

            if !didRemoveEnchantment {
                test.Errorf("Error: remote did not receive remove unit enchantment event for %v", spellName)
            }

            if !didFinishProjectile {
                test.Errorf("Error: remote did not receive projectile finished event for %v", spellName)
            }
        }

        return expect
    }

    // this test works because only one unit has an enchantment and we know that the AI will only
    // target a unit that has an enchantment on it
    doSpellTest("Dispel Magic True", true, makeRemoveEnchantmentExpecter())
    doSpellTest("Dispel Magic", true, makeRemoveEnchantmentExpecter())

    makeSetRangedAttacksExpecter := func() Expecter {
        var expect Expecter

        didCastSpell := false
        didFinishProjectile := false
        didSetAttacks := false

        expect.HandleEvent = func(event RemoteEvent, expectedId uint64, spellName string, model *CombatModel) {
            switch event.GetType() {
                case RemoteUnitTargetSpellType:
                    event := event.(*RemoteUnitTargetSpellEvent)
                    if event.TargetId == expectedId && event.Spell == spellName {
                        didCastSpell = true
                    }
                case RemoteSetRangedAttacksType:
                    event := event.(*RemoteSetRangedAttacksEvent)
                    if event.Id == expectedId {
                        didSetAttacks = true
                    }
                case RemoteProjectileFinishedType:
                    didFinishProjectile = true
            }
        }

        expect.Finished = func() bool {
            return didCastSpell && didFinishProjectile && didSetAttacks
        }

        expect.Assertions = func(test *testing.T, spellName string) {
            if !didCastSpell {
                test.Errorf("Error: remote did not receive cast spell for %v", spellName)
            }

            if !didSetAttacks {
                test.Errorf("Error: remote did not receive set ranged attacks event for %v", spellName)
            }

            if !didFinishProjectile {
                test.Errorf("Error: remote did not receive projectile finished event for %v", spellName)
            }
        }


        return expect
    }

    doSpellTest("Warp Wood", true, makeSetRangedAttacksExpecter(), TestOptions{
        MakeUnit: func() *units.OverworldUnit {
            return units.MakeOverworldUnitFromUnit(units.OrcBowmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
        },
    })

    makeChangeTeamExpecter := func() Expecter {
        var expect Expecter

        didCastSpell := false
        didFinishProjectile := false
        didSwitchTeam := false

        expect.HandleEvent = func(event RemoteEvent, expectedId uint64, spellName string, model *CombatModel) {
            switch event.GetType() {
                case RemoteUnitTargetSpellType:
                    event := event.(*RemoteUnitTargetSpellEvent)
                    if event.TargetId == expectedId && event.Spell == spellName {
                        didCastSpell = true
                    }
                case RemoteSwitchTeamsType:
                    event := event.(*RemoteSwitchTeamsEvent)
                    if event.Id == expectedId {
                        didSwitchTeam = true
                    }
                case RemoteProjectileFinishedType:
                    didFinishProjectile = true
            }
        }

        expect.Finished = func() bool {
            return didCastSpell && didFinishProjectile && didSwitchTeam
        }

        expect.Assertions = func(test *testing.T, spellName string) {
            if !didCastSpell {
                test.Errorf("Error: remote did not receive cast spell for %v", spellName)
            }

            if !didSwitchTeam {
                test.Errorf("Error: remote did not receive switch team event for %v", spellName)
            }

            if !didFinishProjectile {
                test.Errorf("Error: remote did not receive projectile finished event for %v", spellName)
            }
        }


        return expect
    }

    doSpellTest("Creature Binding", true, makeChangeTeamExpecter())
    doSpellTest("Possession", true, makeChangeTeamExpecter(), TestOptions{
        MakeUnit: func() *units.OverworldUnit {
            return units.MakeOverworldUnitFromUnit(units.OrcBowmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
        },
    })

    makeUnitEnchantmentExpecter := func(enchantment data.UnitEnchantment) Expecter {
        var expect Expecter

        didCastSpell := false
        didFinishProjectile := false
        didEnchant := false

        expect.HandleEvent = func(event RemoteEvent, expectedId uint64, spellName string, model *CombatModel) {
            switch event.GetType() {
                case RemoteUnitTargetSpellType:
                    event := event.(*RemoteUnitTargetSpellEvent)
                    if isAttacker(event.TargetId, model) && event.Spell == spellName {
                        didCastSpell = true
                    }
                case RemoteEnchantmentUnitType:
                    event := event.(*RemoteEnchantmentUnitEvent)
                    // enchantments are always applied to the casters army, which is the attacker in this test
                    if isAttacker(event.Id, model) && event.Enchantment == enchantment {
                        didEnchant = true
                    }
                case RemoteProjectileFinishedType:
                    didFinishProjectile = true
            }
        }

        expect.Finished = func() bool {
            return didCastSpell && didFinishProjectile && didEnchant
        }

        expect.Assertions = func(test *testing.T, spellName string) {
            if !didCastSpell {
                test.Errorf("Error: remote did not receive cast spell for %v", spellName)
            }

            if !didEnchant {
                test.Errorf("Error: remote did not receive enchant unit event for %v", spellName)
            }

            if !didFinishProjectile {
                test.Errorf("Error: remote did not receive projectile finished event for %v", spellName)
            }
        }


        return expect
    }

    doSpellTest("Bless", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentBless))
    doSpellTest("Heroism", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentHeroism))
    doSpellTest("Holy Armor", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentHolyArmor))
    doSpellTest("Holy Weapon", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentHolyWeapon))
    doSpellTest("Invulnerability", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentInvulnerability))
    doSpellTest("Lionheart", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentLionHeart))
    doSpellTest("Righteousness", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentRighteousness))
    doSpellTest("True Sight", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentTrueSight))
    doSpellTest("Elemental Armor", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentElementalArmor))
    doSpellTest("Giant Strength", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentGiantStrength))
    doSpellTest("Iron Skin", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentIronSkin))
    doSpellTest("Regeneration", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentRegeneration))
    doSpellTest("Resist Elements", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentResistElements))
    doSpellTest("Stone Skin", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentStoneSkin))
    doSpellTest("Flight", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentFlight))
    doSpellTest("Guardian Wind", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentGuardianWind))
    doSpellTest("Haste", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentHaste))
    doSpellTest("Invisiblity", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentInvisibility))
    doSpellTest("Magic Immunity", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentMagicImmunity))
    doSpellTest("Resist Magic", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentResistMagic))
    doSpellTest("Spell Lock", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentSpellLock))
    doSpellTest("Eldritch Weapon", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentEldritchWeapon))
    doSpellTest("Flame Blade", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentFlameBlade))
    doSpellTest("Immolation", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentImmolation))
    doSpellTest("Berserk", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentBerserk))
    doSpellTest("Cloak of Fear", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentCloakOfFear))
    doSpellTest("Wraith Form", false, makeUnitEnchantmentExpecter(data.UnitEnchantmentWraithForm))

    makeAllUnitsHarmExpecter := func() Expecter {
        didCastSpell := false
        didDamage := false
        didFinishProjectile := false
        didTargetAll := false

        seenIds := set.NewSet[uint64]()

        var harm Expecter
        harm.HandleEvent = func(event RemoteEvent, expectedId uint64, spellName string, model *CombatModel) {
            switch event.GetType() {
                case RemoteUnitTargetSpellType:
                    event := event.(*RemoteUnitTargetSpellEvent)
                    if event.Spell == spellName {
                        didCastSpell = true
                        seenIds.Insert(event.TargetId)

                        army := model.GetArmy(model.GetUnitById(event.TargetId))
                        if len(army.GetUnits()) == seenIds.Size() {
                            didTargetAll = true
                            for _, unit := range army.GetUnits() {
                                if !seenIds.Contains(unit.Id) {
                                    didTargetAll = false
                                }
                            }
                        }
                    }
                case RemoteDamageType:
                    event := event.(*RemoteDamageEvent)
                    // one of the units should be damaged
                    if event.Id == expectedId {
                        didDamage = true
                    }
                case RemoteProjectileFinishedType:
                    didFinishProjectile = true
            }
        }

        harm.Finished = func() bool {
            return didCastSpell && didDamage && didFinishProjectile && didTargetAll
        }

        harm.Assertions = func(test *testing.T, spellName string) {
            if !didDamage {
                test.Errorf("Error: remote did not receive damage event for %v", spellName)
            }

            if !didCastSpell {
                test.Errorf("Error: remote did not receive cast spell for %v", spellName)
            }

            if !didFinishProjectile {
                test.Errorf("Error: remote did not receive projectile finished event for %v", spellName)
            }

            if !didTargetAll {
                test.Errorf("Error: remote did not target all units for %v", spellName)
            }
        }

        return harm
    }

    makeOrcBowmen := func() *units.OverworldUnit {
        return units.MakeOverworldUnitFromUnit(units.OrcBowmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    }

    doSpellTest("Death Spell", true, makeAllUnitsHarmExpecter(), TestOptions{
        MakeUnit: makeOrcBowmen,
    }, TestOptions{
        MakeUnit: makeOrcBowmen,
    })

    makeArchAngel := func() *units.OverworldUnit {
        return units.MakeOverworldUnitFromUnit(units.ArchAngel, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    }

    doSpellTest("Holy Word", true, makeAllUnitsHarmExpecter(), TestOptions{
        MakeUnit: makeArchAngel,
    }, TestOptions{
        MakeUnit: makeArchAngel,
    })
}

// Exercise actual wall-crossing damage, retrying zero-damage fire rolls so
// the regression does not depend on the confused unit surviving its first roll.
type confusedFireMovementActions struct {
    *ProxyActions
    test *testing.T
    moves int
}

func (actions *confusedFireMovementActions) MoveUnit(unit *ArmyUnit, path pathfinding.Path) {
    actions.test.Helper()
    actions.moves += 1
    if unit.GetHealth() <= 0 {
        actions.test.Error("Confused movement tried to move the unit again after it died")
        // Bound the buggy loop so the remaining turn-state assertions can run.
        unit.MovesLeft = fraction.Zero()
        return
    }
    if len(path) != 1 {
        actions.test.Fatalf("Expected one adjacent step, got %v", path)
    }
    model := actions.Model
    target := path[0]
    if model.InsideWallOfFire(unit.X, unit.Y) || !model.InsideWallOfFire(target.X, target.Y) {
        actions.test.Fatal("Movement did not cross into Wall of Fire")
    }
    startX, startY, movesLeft := unit.X, unit.Y, unit.MovesLeft
    for range 100 {
        if model.MoveUnit(unit, target.X, target.Y) {
            return
        }
        // With one HP and zero defense, surviving means the fire rolled zero
        // damage. Restore the attempted step and retry the actual crossing.
        model.setTileUnit(unit.X, unit.Y, nil)
        unit.X, unit.Y, unit.MovesLeft = startX, startY, movesLeft
        model.setTileUnit(startX, startY, unit)
    }
    actions.test.Fatal("Wall of Fire rolled zero damage on 100 consecutive crossings")
}

func TestRemoteWallOfFireDeath(test *testing.T) {
    for _, movingTeam := range []Team{TeamAttacker, TeamDefender} {
        name := "attacker"
        if movingTeam == TeamDefender {
            name = "defender"
        }
        test.Run(name, func(test *testing.T) {
            peer1, peer2 := net.Pipe()
            defer peer1.Close()
            defer peer2.Close()
            peer1.SetWriteDeadline(time.Now().Add(5 * time.Second))
            peer2.SetWriteDeadline(time.Now().Add(5 * time.Second))
            // The mover's owner is the non-authoritative peer. IsServer()
            // returns the inverse of Remote.IsServer.
            attackerRemote := MakeRemote(movingTeam == TeamAttacker, false, peer1)
            defenderRemote := MakeRemote(movingTeam == TeamDefender, true, peer2)
            makeModel := func(remote *Remote) *CombatModel {
                makeArmy := func() *Army {
                    army := &Army{Player: makeTestCombatPlayer(true)}
                    template := units.LizardSpearmen
                    template.Count = 1
                    template.HitPoints = 1
                    template.Defense = 0
                    army.AddUnit(units.MakeOverworldUnitFromUnit(template, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{}))
                    return army
                }
                return MakeCombatModel(spellbook.Spells{}, makeArmy(), makeArmy(), CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10), remote)
            }
            attackerModel, defenderModel := makeModel(attackerRemote), makeModel(defenderRemote)
            owner, server := attackerModel, defenderModel
            if movingTeam == TeamDefender {
                owner, server = defenderModel, attackerModel
            }
            if owner.IsServer() || !server.IsServer() {
                test.Fatal("Incorrect damage-authority setup")
            }
            id := owner.GetArmyForTeam(movingTeam).units[0].Id
            path := pathfinding.Path{image.Pt(11, 10), image.Pt(12, 10), image.Pt(13, 10)}
            for _, model := range []*CombatModel{owner, server} {
                model.SelectedUnit = nil // Isolate movement/damage from turn advancement.
                for _, point := range path {
                    model.GetTile(point.X, point.Y).InsideFire = true
                }
            }

            quit, cancel := context.WithCancel(context.Background())
            var receiveWait sync.WaitGroup
            for _, remote := range []*Remote{attackerRemote, defenderRemote} {
                receiveWait.Go(func() { remote.RunReceiveLoop(quit) })
            }
            defer func() {
                cancel()
                receiveWait.Wait()
            }()
            receive := func(remote *Remote) RemoteEvent {
                test.Helper()
                select {
                    case event := <-remote.Events:
                        return event
                    case <-time.After(5 * time.Second):
                        test.Fatal("Timed out waiting for movement/fire events")
                        return nil
                }
            }

            // Fire can roll zero damage. Retry those crossings from the same
            // initial state, without assuming the random roll must be lethal.
            for attempt := 0; attempt < 100; attempt++ {
                for _, model := range []*CombatModel{owner, server} {
                    unit := model.GetUnitById(id)
                    model.setTileUnit(unit.X, unit.Y, nil)
                    unit.X, unit.Y = 10, 10
                    unit.MovesLeft = fraction.FromInt(10)
                    model.setTileUnit(unit.X, unit.Y, unit)
                }
                ownerUnit := owner.GetUnitById(id)
                serverUnit := server.GetUnitById(id)
                if err := owner.RemoteMove(ownerUnit, path); err != nil {
                    test.Fatal(err)
                }
                // Simulate normal network delay: local movement finishes before
                // the server replays the command and returns the fire result.

                event, ok := receive(server.Remote).(*RemoteMoveEvent)
                if !ok || event.Id != id {
                    test.Fatal("Expected the mover's remote movement command")
                }

                var waiter sync.WaitGroup

                waiter.Go(func() {
                    (&ProxyActions{Model: owner}).MoveUnit(ownerUnit, path)
                })

                waiter.Go(func() {
                    (&ProxyActions{Model: server}).MoveUnit(serverUnit, event.Path)
                })

                waiter.Wait()

                // A stream marker ensures all damage/kill messages have been
                // applied before comparing models, including zero-damage rolls.
                if err := server.Remote.SendEvent(&RemoteGlobalSpellEvent{Type: RemoteGlobalSpellType, Name: "fire-test-marker"}); err != nil {
                    test.Fatal(err)
                }
                for {
                    event := receive(owner.Remote)
                    if marker, ok := event.(*RemoteGlobalSpellEvent); ok && marker.Name == "fire-test-marker" {
                        break
                    }
                    owner.HandleRemoteEvent(optional.Empty[SpellSystem](), optional.Empty[AddDamageIndicators](), event)
                }
                if serverUnit.GetHealth() > 0 {
                    if ownerUnit.GetHealth() != serverUnit.GetHealth() {
                        test.Fatal("Fire damage diverged between models")
                    }
                    continue
                }

                if ownerUnit.GetHealth() != 0 || owner.GetUnitById(id) != nil || server.GetUnitById(id) != nil {
                    test.Fatal("Lethal fire damage did not remove the mover from both active armies")
                }
                if owner.GetArmyForTeam(movingTeam).GetKilledUnitById(id) != ownerUnit || server.GetArmyForTeam(movingTeam).GetKilledUnitById(id) != serverUnit {
                    test.Fatal("Dead mover missing from a peer's killed-unit list")
                }
                // Synchronization happens after movement finishes, so the
                // non-server owner may advance one extra position on the path.
                movementPositions := append(pathfinding.Path{image.Pt(10, 10)}, path...)
                ownerPosition := slices.Index(movementPositions, image.Pt(ownerUnit.X, ownerUnit.Y))
                serverPosition := slices.Index(movementPositions, image.Pt(serverUnit.X, serverUnit.Y))
                if serverPosition < 0 || ownerPosition < serverPosition || ownerPosition > serverPosition+1 {
                    test.Errorf("Owner moved beyond the allowed extra step: corpse at (%d,%d), server stopped at (%d,%d)", ownerUnit.X, ownerUnit.Y, serverUnit.X, serverUnit.Y)
                }
                if ownerUnit.X == path[len(path)-1].X && ownerUnit.Y == path[len(path)-1].Y {
                    test.Error("Owner completed the path despite dying at the first fire crossing")
                }
                return
            }
            test.Fatal("Wall of Fire rolled zero damage on 100 consecutive crossings")
        })
    }
}

func TestAIUnitDeathDoesNotFinishNextUnitsTurn(test *testing.T) {
    for _, team := range []Team{TeamAttacker, TeamDefender} {
        name := "attacker"
        if team == TeamDefender {
            name = "defender"
        }
        test.Run(name, func(test *testing.T) {
            aiArmy := &Army{Player: makeTestCombatPlayer(false)}
            weak := units.LizardSpearmen
            weak.Count = 1
            weak.HitPoints = 1
            weak.Defense = 0
            weak.MeleeAttackPower = 1
            weak.MovementSpeed = 2
            aiUnit := aiArmy.AddUnit(units.MakeOverworldUnitFromUnit(weak, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{}))
            next := aiArmy.AddUnit(units.MakeOverworldUnitFromUnit(weak, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{}))

            enemyArmy := &Army{Player: makeTestCombatPlayer(true)}
            strong := weak
            strong.HitPoints = 100
            strong.MeleeAttackPower = 100
            // The enemy survives the opening attack and always hits on its
            // counterattack, killing the one-HP AI unit without random retries.
            enemy := enemyArmy.AddUnit(&OverrideToHitMelee{units.MakeOverworldUnitFromUnit(strong, 0, 0, data.PlaneArcanus, data.BannerBlue, &units.NoExperienceInfo{}, &units.NoEnchantments{})})
            attackingArmy, defendingArmy := aiArmy, enemyArmy
            if team == TeamDefender {
                attackingArmy, defendingArmy = enemyArmy, aiArmy
            }
            model := MakeCombatModel(spellbook.Spells{}, defendingArmy, attackingArmy, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10), nil)
            for _, unit := range []*ArmyUnit{aiUnit, next, enemy} {
                model.setTileUnit(unit.X, unit.Y, nil)
            }
            aiUnit.X, aiUnit.Y = 10, 10
            enemy.X, enemy.Y = 11, 10
            next.X, next.Y = 8, 10
            for _, unit := range []*ArmyUnit{aiUnit, next, enemy} {
                model.setTileUnit(unit.X, unit.Y, unit)
            }
            model.Turn = team
            model.SelectedUnit = aiUnit
            initialTurn := model.CurrentTurn
            nextLastTurn, nextMoves := next.LastTurn, next.MovesLeft
            enemyLastTurn := enemy.LastTurn
            if !model.IsAIControlled(aiUnit) || !model.canMeleeAttack(aiUnit, enemy, true) {
                test.Fatal("Expected a locally controlled AI unit able to melee the adjacent enemy")
            }

            model.Update(&ProxySpellSystem{Model: model}, &ProxyActions{Model: model}, false, 0, 0)

            if aiUnit.GetHealth() != 0 || model.GetUnitById(aiUnit.Id) != nil || aiArmy.GetKilledUnitById(aiUnit.Id) != aiUnit {
                test.Fatal("Expected the acting AI unit to die to the melee counterattack")
            }
            if enemy.GetHealth() <= 0 || enemy.Attacked != 1 {
                test.Error("Expected the adjacent enemy to survive one melee attack")
            }
            if model.GetUnit(10, 10) != nil {
                test.Error("Dead AI unit still occupies its tile")
            }
            if model.SelectedUnit != next {
                test.Error("AI unit's death skipped the next unit instead of leaving it selected")
            }
            if next.LastTurn != nextLastTurn || !next.MovesLeft.Equals(nextMoves) {
                test.Error("DoneTurn consumed the next unit's turn or movement after the AI unit died")
            }
            if model.CurrentTurn != initialTurn || model.Turn != team {
                test.Error("AI unit's death incorrectly advanced the army/round")
            }
            if enemy.LastTurn != enemyLastTurn {
                test.Error("AI unit's death incorrectly consumed the enemy's turn")
            }
        })
    }
}

func TestConfusedMovementDeathInWallOfFire(test *testing.T) {
    for _, team := range []Team{TeamAttacker, TeamDefender} {
        name := "attacker"
        if team == TeamDefender {
            name = "defender"
        }
        for _, movement := range []int{1, 4} {
            movementName := "last_movement_point"
            if movement == 4 {
                movementName = "movement_remaining"
            }
            test.Run(name+"/"+movementName, func(test *testing.T) {
                makeArmy := func() *Army {
                    army := &Army{Player: makeTestCombatPlayer(true)}
                    template := units.LizardSpearmen
                    template.Count = 1
                    template.HitPoints = 1
                    template.Defense = 0
                    for range 2 {
                        army.AddUnit(units.MakeOverworldUnitFromUnit(template, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{}))
                    }
                    return army
                }
                model := MakeCombatModel(spellbook.Spells{}, makeArmy(), makeArmy(), CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10), nil)
                army := model.GetArmyForTeam(team)
                confused, next := army.units[0], army.units[1]
                model.setTileUnit(confused.X, confused.Y, nil)
                confused.X, confused.Y = 10, 10
                model.setTileUnit(confused.X, confused.Y, confused)
                confused.Curses = append(confused.Curses, data.UnitCurseConfusion)
                confused.ConfusionAction = ConfusionActionMoveRandomly
                confused.MovesLeft = fraction.FromInt(movement)
                // Every possible random step crosses fire, regardless of direction.
                for dx := -1; dx <= 1; dx++ {
                    for dy := -1; dy <= 1; dy++ {
                        if dx != 0 || dy != 0 {
                            model.GetTile(confused.X + dx, confused.Y + dy).InsideFire = true
                        }
                    }
                }
                model.Turn = team
                model.SelectedUnit = confused
                initialTurn := model.CurrentTurn
                nextLastTurn, nextMoves := next.LastTurn, next.MovesLeft
                actions := &confusedFireMovementActions{
                    ProxyActions: &ProxyActions{Model: model},
                    test: test,
                }

                model.Update(&ProxySpellSystem{Model: model}, actions, false, 0, 0)

                if confused.GetHealth() != 0 || model.GetUnitById(confused.Id) != nil || army.GetKilledUnitById(confused.Id) != confused {
                    test.Error("Confused unit was not killed and moved into the killed-unit list")
                }
                if model.GetUnit(10, 10) != nil {
                    test.Error("Dead confused unit still occupies its tile")
                }
                if actions.moves != 1 {
                    test.Errorf("Movement actions = %d, want exactly one before death", actions.moves)
                }
                if model.SelectedUnit != next {
                    test.Error("Death skipped the next unit instead of leaving it selected")
                }
                if next.LastTurn != nextLastTurn || !next.MovesLeft.Equals(nextMoves) {
                    test.Error("Confused unit's death consumed the next unit's turn or movement")
                }
                if model.CurrentTurn != initialTurn || model.Turn != team {
                    test.Error("Confused unit's death incorrectly advanced the army/round")
                }
            })
        }
    }
}

func TestRemoteTurnEnchantments(test *testing.T) {
    scenarios := []struct {
        name string
        enchantment data.CombatEnchantment
        mana int
        lethal bool
    }{
        {name: "wrack", enchantment: data.CombatEnchantmentWrack, mana: 20},
        {name: "wrack_low_health", enchantment: data.CombatEnchantmentWrack, mana: 20, lethal: true},
        {name: "terror", enchantment: data.CombatEnchantmentTerror, mana: 20},
        {name: "mana_leak", enchantment: data.CombatEnchantmentManaLeak, mana: 20},
        {name: "mana_leak_clamped", enchantment: data.CombatEnchantmentManaLeak, mana: 3},
    }
    for _, scenario := range scenarios {
        for _, affectedTeam := range []Team{TeamAttacker, TeamDefender} {
            for _, authorityTeam := range []Team{TeamAttacker, TeamDefender} {
                teamName := func(team Team) string {
                    if team == TeamAttacker {
                        return "attacker"
                    }
                    return "defender"
                }
                test.Run(scenario.name+"/affected_"+teamName(affectedTeam)+"/authority_"+teamName(authorityTeam), func(test *testing.T) {
                    peer1, peer2 := net.Pipe()
                    defer peer1.Close()
                    defer peer2.Close()
                    peer1.SetWriteDeadline(time.Now().Add(5 * time.Second))
                    peer2.SetWriteDeadline(time.Now().Add(5 * time.Second))
                    // Remote.IsServer is inverted by CombatModel.IsServer().
                    attackerRemote := MakeRemote(authorityTeam != TeamAttacker, false, peer1)
                    defenderRemote := MakeRemote(authorityTeam != TeamDefender, true, peer2)
                    makeModel := func(remote *Remote) *CombatModel {
                        makeArmy := func() *Army {
                            army := &Army{Player: makeTestCombatPlayer(false)}
                            for _, resistance := range []int{-1, 100} {
                                template := units.Warlocks
                                template.Resistance = resistance
                                template.Count = 2
                                template.HitPoints = 5
                                army.AddUnit(units.MakeOverworldUnitFromUnit(template, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{}))
                            }
                            return army
                        }
                        return MakeCombatModel(spellbook.Spells{}, makeArmy(), makeArmy(), CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10), remote)
                    }
                    attackerModel := makeModel(attackerRemote)
                    defenderModel := makeModel(defenderRemote)
                    source, receiver := attackerModel, defenderModel
                    if authorityTeam == TeamDefender {
                        source, receiver = defenderModel, attackerModel
                    }
                    if !source.IsServer() || receiver.IsServer() {
                        test.Fatal("Incorrect authority setup")
                    }

                    quit, cancel := context.WithCancel(context.Background())
                    var receiveWait sync.WaitGroup
                    for _, remote := range []*Remote{attackerRemote, defenderRemote} {
                        receiveWait.Go(func() { remote.RunReceiveLoop(quit) })
                    }
                    defer func() {
                        cancel()
                        receiveWait.Wait()
                    }()

                    // A marker on the same stream bounds the events for each round.
                    // Process model mutations on the test goroutine, not the receive loop.
                    flush := func(from, to *CombatModel) map[RemoteType]int {
                        test.Helper()
                        if err := from.Remote.SendEvent(&RemoteGlobalSpellEvent{Type: RemoteGlobalSpellType, Name: "turn-test-marker"}); err != nil {
                            test.Fatal(err)
                        }
                        counts := make(map[RemoteType]int)
                        timeout := time.NewTimer(5 * time.Second)
                        defer timeout.Stop()
                        for {
                            select {
                                case event := <-to.Remote.Events:
                                    if marker, ok := event.(*RemoteGlobalSpellEvent); ok && marker.Name == "turn-test-marker" {
                                        return counts
                                    }
                                    counts[event.GetType()] += 1
                                    to.HandleRemoteEvent(optional.Empty[SpellSystem](), optional.Empty[AddDamageIndicators](), event)
                                case <-timeout.C:
                                    test.Fatal("Timed out receiving turn-effect events")
                            }
                        }
                    }

                    ids := []uint64{}
                    for _, unit := range source.GetArmyForTeam(affectedTeam).units {
                        ids = append(ids, unit.Id)
                    }
                    initialHealth := source.GetUnitById(ids[0]).GetHealth()
                    for _, model := range []*CombatModel{attackerModel, defenderModel} {
                        affected := model.GetArmyForTeam(affectedTeam)
                        caster := model.GetArmyForTeam(oppositeTeam(affectedTeam))
                        caster.AddEnchantment(scenario.enchantment)
                        affected.ManaPool = scenario.mana
                        affected.Player.(*playerlib.Player).Mana = scenario.mana
                        affected.units[0].CastingSkill = 12
                        affected.units[1].CastingSkill = 3
                        caster.ManaPool = 37
                        caster.Player.(*playerlib.Player).Mana = 37
                        for _, unit := range caster.units {
                            unit.CastingSkill = 9
                        }
                        // Keep lethal Wrack from triggering unrelated selected-unit advancement.
                        model.SelectedUnit = caster.units[0]
                        if scenario.lethal {
                            unit := affected.units[0]
                            unit.Unit.AdjustHealth(1 - unit.GetHealth())
                        }
                        if affected.units[0].GetResistance() != 0 || affected.units[1].GetResistance() < 10 {
                            test.Fatal("Incorrect low/high resistance setup")
                        }
                    }

                    // Resistance is clamped at zero, so low-resistance rolls can
                    // still succeed. Check the actual authoritative outcome rather
                    // than assuming a random roll must fail.
                    for round := 1; round <= 2; round++ {
                        // Reset receiver movement first, before authoritative Terror events.
                        receiver.NextTurn()
                        if counts := flush(receiver, source); len(counts) != 0 {
                            test.Fatalf("Non-authoritative peer emitted turn effects: %v", counts)
                        }
                        activeBefore := len(source.GetArmyForTeam(affectedTeam).units)
                        vulnerable := source.GetUnitById(ids[0])
                        healthBefore, figuresBefore := 0, 0
                        if vulnerable != nil {
                            healthBefore, figuresBefore = vulnerable.GetHealth(), vulnerable.Figures()
                        }
                        source.NextTurn()
                        counts := flush(source, receiver)
                        expected := make(map[RemoteType]int)
                        switch scenario.enchantment {
                            case data.CombatEnchantmentWrack:
                                expected[RemoteDamageType] = activeBefore
                                if killed := activeBefore - len(source.GetArmyForTeam(affectedTeam).units); killed > 0 {
                                    expected[RemoteKillUnitType] = killed
                                }
                                if vulnerable != nil {
                                    damage := healthBefore - vulnerable.GetHealth()
                                    if damage < 0 || damage > figuresBefore {
                                        test.Errorf("Wrack damage = %d, want between 0 and %d", damage, figuresBefore)
                                    }
                                }
                            case data.CombatEnchantmentTerror:
                                if source.GetUnitById(ids[0]).MovesLeft.Equals(fraction.Zero()) {
                                    expected[RemoteUnitMovesLeftType] = 1
                                }
                            case data.CombatEnchantmentManaLeak:
                                expected[RemoteUnitCastSkillType] = 2
                        }
                        if len(counts) != len(expected) {
                            test.Errorf("Round %d: event counts = %v, want %v", round, counts, expected)
                        }
                        for kind, count := range expected {
                            if counts[kind] != count {
                                test.Errorf("Round %d: %s count = %d, want %d", round, kind, counts[kind], count)
                            }
                        }

                        for _, model := range []*CombatModel{attackerModel, defenderModel} {
                            affected := model.GetArmyForTeam(affectedTeam)
                            expectedMana := scenario.mana
                            if scenario.enchantment == data.CombatEnchantmentManaLeak {
                                expectedMana = max(0, expectedMana - 5 * round)
                            }
                            if affected.ManaPool != expectedMana || affected.Player.GetMana() != expectedMana {
                                test.Errorf("Round %d: affected mana pool/player = %d/%d, want %d", round, affected.ManaPool, affected.Player.GetMana(), expectedMana)
                            }
                            for index, id := range ids {
                                unit := model.GetUnitById(id)
                                if scenario.lethal && index == 0 && source.GetUnitById(id) == nil {
                                    corpse := affected.GetKilledUnitById(id)
                                    if unit != nil || corpse == nil || corpse.GetHealth() != 0 {
                                        test.Error("Lethal Wrack did not move the dead unit into the killed-unit list")
                                    }
                                    if corpse != nil && model.GetUnit(corpse.X, corpse.Y) != nil {
                                        test.Error("Lethal Wrack left the dead unit on its tile")
                                    }
                                    continue
                                }
                                if unit == nil {
                                    test.Fatalf("Unit %d unexpectedly removed", id)
                                }
                                expectedHealth := initialHealth
                                if scenario.enchantment == data.CombatEnchantmentWrack && index == 0 {
                                    expectedHealth = source.GetUnitById(id).GetHealth()
                                }
                                if unit.GetHealth() != expectedHealth {
                                    test.Errorf("Round %d, unit %d: health = %d, want %d", round, id, unit.GetHealth(), expectedHealth)
                                }
                                expectedMoves := unit.GetMovementSpeed()
                                if scenario.enchantment == data.CombatEnchantmentTerror && index == 0 {
                                    expectedMoves = source.GetUnitById(id).MovesLeft
                                }
                                if !unit.MovesLeft.Equals(expectedMoves) {
                                    test.Errorf("Round %d, unit %d: moves = %v, want %v", round, id, unit.MovesLeft, expectedMoves)
                                }
                                expectedSkill := float32(12)
                                if index == 1 {
                                    expectedSkill = 3
                                }
                                if scenario.enchantment == data.CombatEnchantmentManaLeak {
                                    expectedSkill = max(0, expectedSkill - float32(5 * round))
                                }
                                if unit.CastingSkill != expectedSkill {
                                    test.Errorf("Round %d, unit %d: casting skill = %v, want %v", round, id, unit.CastingSkill, expectedSkill)
                                }
                            }
                            caster := model.GetArmyForTeam(oppositeTeam(affectedTeam))
                            if caster.ManaPool != 37 || caster.Player.GetMana() != 37 {
                                test.Error("Enchantment drained its own army's mana")
                            }
                            for _, unit := range caster.units {
                                if unit.GetHealth() != initialHealth || unit.CastingSkill != 9 || !unit.MovesLeft.Equals(unit.GetMovementSpeed()) {
                                    test.Error("Enchantment affected its own army's units")
                                }
                            }
                            if model.CurrentTurn != 1 + round {
                                test.Errorf("Current turn = %d, want %d", model.CurrentTurn, 1 + round)
                            }
                        }
                    }
                })
            }
        }
    }
}

func TestRemoteRaiseDead(test *testing.T) {
    for _, team := range []Team{TeamAttacker, TeamDefender} {
        name := "attacker"
        if team == TeamDefender {
            name = "defender"
        }
        test.Run(name, func(test *testing.T) {
            peer1, peer2 := net.Pipe()
            defer peer1.Close()
            defer peer2.Close()
            // Bound synchronous sends as well as event waits if synchronization breaks.
            peer1.SetWriteDeadline(time.Now().Add(5 * time.Second))
            peer2.SetWriteDeadline(time.Now().Add(5 * time.Second))

            attackerRemote := MakeRemote(false, false, peer1)
            defenderRemote := MakeRemote(true, true, peer2)
            makeModel := func(remote *Remote) *CombatModel {
                makeArmy := func() *Army {
                    army := &Army{Player: makeTestCombatPlayer(false)}
                    // Keep another unit alive, and avoid killing the selected unit.
                    for range 2 {
                        army.AddUnit(units.MakeOverworldUnitFromUnit(units.Warlocks, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{}))
                    }
                    return army
                }
                return MakeCombatModel(spellbook.Spells{}, makeArmy(), makeArmy(), CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10), remote)
            }
            attackerModel := makeModel(attackerRemote)
            defenderModel := makeModel(defenderRemote)
            source, receiver := attackerModel, defenderModel
            if team == TeamDefender {
                source, receiver = defenderModel, attackerModel
            }

            quit, cancel := context.WithCancel(context.Background())
            receiveDone := make(chan struct{})
            go func() {
                defer close(receiveDone)
                receiver.Remote.RunReceiveLoop(quit)
            }()
            defer func() {
                cancel()
                <-receiveDone
            }()

            receiverSpells := &ProxySpellSystem{Model: receiver}
            receiveThrough := func(wanted RemoteType) {
                test.Helper()
                timeout := time.NewTimer(5 * time.Second)
                defer timeout.Stop()
                for {
                    select {
                        case event := <-receiver.Remote.Events:
                            receiver.HandleRemoteEvent(optional.Of[SpellSystem](receiverSpells), optional.Empty[AddDamageIndicators](), event)
                            if event.GetType() == wanted {
                                return
                            }
                        case <-timeout.C:
                            test.Fatalf("Timed out waiting for %s", wanted)
                    }
                }
            }

            sourceArmy := source.GetArmyForTeam(team)
            unit := sourceArmy.units[1]
            id := unit.Id
            // Kill a locally controlled unit: this must still notify the other peer.
            if source.IsRemoteUnit(unit) {
                test.Fatal("Expected the casting peer to control the unit")
            }
            damage := unit.GetHealth()
            unit.TakeDamage(damage, DamageNormal)
            if err := source.RemoteDamage(unit, DamageNormal, damage); err != nil {
                test.Fatal(err)
            }
            source.KillUnit(unit)
            receiveThrough(RemoteKillUnitType)
            for _, model := range []*CombatModel{attackerModel, defenderModel} {
                army := model.GetArmyForTeam(team)
                if army.GetKilledUnitById(id) == nil {
                    test.Fatal("Dead unit missing from a peer's killed-unit list")
                }
                if model.GetUnitById(id) != nil {
                    test.Fatal("Dead unit still in a peer's active army")
                }
            }

            casted := false
            source.InvokeSpell(&ProxySpellSystem{Model: source}, sourceArmy, nil, spellbook.Spell{Name: "Raise Dead", Magic: data.LifeMagic}, func(success bool) {
                casted = success
            })
            if !casted {
                test.Fatal("Raise Dead did not succeed")
            }
            receiveThrough(RemoteRaiseDeadUnitType)

            for _, peer := range []struct {
                name string
                model *CombatModel
            }{{"attacker", attackerModel}, {"defender", defenderModel}} {
                army := peer.model.GetArmyForTeam(team)
                revived := peer.model.GetUnitById(id)
                if revived == nil {
                    test.Fatalf("%s model: revived unit missing from active army", peer.name)
                }
                if revived.GetHealth() <= 0 || revived.GetHealth() != unit.GetHealth() {
                    test.Errorf("%s model: revived health = %d, want %d", peer.name, revived.GetHealth(), unit.GetHealth())
                }
                if army.GetKilledUnitById(id) != nil {
                    test.Errorf("%s model: revived unit still in killed-unit list", peer.name)
                }
                if revived.X != unit.X || revived.Y != unit.Y {
                    test.Errorf("%s model: revived unit has different coordinates", peer.name)
                }
                if peer.model.GetUnit(revived.X, revived.Y) != revived {
                    test.Errorf("%s model: revived unit missing from tile", peer.name)
                }
                if len(army.units) != 2 {
                    test.Errorf("%s model: active army has %d units, want 2", peer.name, len(army.units))
                }
            }
        })
    }
}

func TestDisenchantArea(test *testing.T) {
    var allSpells spellbook.Spells

    peer1, peer2 := net.Pipe()

    defer peer1.Close()
    defer peer2.Close()

    // the attacker's remote, which represents the defender, so false for isAttacker
    attackerRemote := MakeRemote(false, false, peer1)
    defenderRemote := MakeRemote(true, true, peer2)

    defendingArmy := &Army{Player: makeTestCombatPlayer(false)}

    // default is hell hounds: fantastic chaos unit
    defendingOverworld := units.MakeOverworldUnitFromUnit(units.HellHounds, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    defendingUnit := defendingArmy.AddUnit(defendingOverworld)
    // an enchantment that can be removed via dispel magic
    defendingUnit.AddEnchantment(data.UnitEnchantmentGiantStrength)

    defendingArmy.AddEnchantment(data.CombatEnchantmentMetalFires)

    attackingArmy := &Army{Player: makeTestCombatPlayer(false)}
    // attacking army has a unit that can range attack. grant very high tohit
    attacker := units.MakeOverworldUnitFromUnit(units.Warlocks, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    attackingArmy.AddUnit(&OverrideToHitMelee{attacker})

    attackerModel := MakeCombatModel(allSpells, defendingArmy, attackingArmy, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10), attackerRemote)
    // defenderModel := MakeCombatModel(allSpells, defendingArmy, attackingArmy, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10), defenderRemote)

    quit, cancel := context.WithCancel(context.Background())
    defer cancel()

    var wait sync.WaitGroup

    wait.Go(func(){ defenderRemote.RunReceiveLoop(quit) })

    spellSystem := makeSpellSystem(attackerModel)

    disenchantAreaSpell := spellbook.Spell{Name: "Disenchant Area", OverrideCost: 1000, Magic: data.ArcaneMagic}

    attackerModel.InvokeSpell(&spellSystem, attackerModel.AttackingArmy, nil, disenchantAreaSpell, func(success bool) { })

    done := make(chan struct{})

    globalSpell := false
    removeEnchantment := false

    wait.Go(func() {
        defer close(done)

        for quit.Err() == nil {
            select {
                case <-quit.Done():
                    break
                case event := <-defenderRemote.Events:
                    switch event.GetType() {
                        case RemoteGlobalSpellType:
                            globalSpell = true
                        case RemoteRemoveArmyEnchantmentType:
                            removeEnchantment = true
                    }
            }

            if globalSpell && removeEnchantment {
                break
            }
        }
    })

    select {
        case <-time.After(500 * time.Millisecond):
            test.Errorf("Error: did not receive expected events for Disenchant Area")
        case <-done:
            if !globalSpell {
                test.Errorf("Error: remote did not receive global spell event for Disenchant Area")
            }

            if !removeEnchantment {
                test.Errorf("Error: remote did not receive remove army enchantment event for Disenchant Area")
            }
    }
}

func TestCounterMagic(test *testing.T) {
    var allSpells spellbook.Spells

    peer1, peer2 := net.Pipe()

    defer peer1.Close()
    defer peer2.Close()

    // the attacker's remote, which represents the defender, so false for isAttacker
    attackerRemote := MakeRemote(false, false, peer1)
    defenderRemote := MakeRemote(true, true, peer2)

    defendingArmy := &Army{Player: makeTestCombatPlayer(false)}

    // default is hell hounds: fantastic chaos unit
    defendingOverworld := units.MakeOverworldUnitFromUnit(units.HellHounds, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    defendingUnit := defendingArmy.AddUnit(defendingOverworld)
    // an enchantment that can be removed via dispel magic
    defendingUnit.AddEnchantment(data.UnitEnchantmentGiantStrength)

    defendingArmy.AddEnchantment(data.CombatEnchantmentMetalFires)

    attackingArmy := &Army{Player: makeTestCombatPlayer(false)}
    // attacking army has a unit that can range attack. grant very high tohit
    attacker := units.MakeOverworldUnitFromUnit(units.Warlocks, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
    attackingArmy.AddUnit(&OverrideToHitMelee{attacker})

    attackerModel := MakeCombatModel(allSpells, defendingArmy, attackingArmy, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10), attackerRemote)
    // defenderModel := MakeCombatModel(allSpells, defendingArmy, attackingArmy, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10), defenderRemote)

    quit, cancel := context.WithCancel(context.Background())
    defer cancel()

    var wait sync.WaitGroup

    wait.Go(func(){ defenderRemote.RunReceiveLoop(quit) })

    spellSystem := makeSpellSystem(attackerModel)

    counterMagicLevel := 1000

    counterMagicSpell := spellbook.Spell{Name: "Counter Magic", OverrideCost: counterMagicLevel, Magic: data.SorceryMagic}

    attackerModel.InvokeSpell(&spellSystem, attackerModel.AttackingArmy, nil, counterMagicSpell, func(success bool) { })

    done := make(chan struct{})

    globalSpell := false
    counterMagic := false

    wait.Go(func() {
        defer close(done)

        for quit.Err() == nil {
            select {
                case <-quit.Done():
                    break
                case event := <-defenderRemote.Events:
                    switch event.GetType() {
                        case RemoteGlobalSpellType:
                            globalSpell = true
                        case RemoteCounterMagicType:
                            event := event.(*RemoteCounterMagicEvent)
                            if event.Value == counterMagicLevel {
                                counterMagic = true
                            }
                    }
            }

            if globalSpell && counterMagic {
                break
            }
        }
    })

    select {
        case <-time.After(500 * time.Millisecond):
            test.Errorf("Error: did not receive expected events for counter magic")
        case <-done:
            if !globalSpell {
                test.Errorf("Error: remote did not receive global spell event for counter magic")
            }

            if !counterMagic {
                test.Errorf("Error: remote did not receive counter magic event")
            }
    }
}

// Mirror CombatScreen.doMelee's local resolution / remote event replay without
// requiring animations, audio, or a window.
type remoteMeleeTestActions struct {
	*ProxyActions
	test *testing.T
}

func (actions *remoteMeleeTestActions) MeleeAttack(attacker, defender *ArmyUnit) {
	actions.test.Helper()
	model := actions.Model
	if !model.IsRemoteUnit(attacker) {
		model.meleeAttack(attacker, defender)
		if err := model.RemoteFinishMeleeAttack(); err != nil {
			actions.test.Fatal(err)
		}
		return
	}

	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()
	for {
		select {
		case event := <-model.Remote.Events:
			if event.GetType() == RemoteFinishMeleeAttackType {
				return
			}
			model.HandleRemoteEvent(optional.Empty[SpellSystem](), optional.Empty[AddDamageIndicators](), event)
		case <-timeout.C:
			actions.test.Fatal("Timed out waiting for remote melee to finish")
		}
	}
}

func TestRemoteDefenderMeleeMutualDeathSelectsNextUnit(test *testing.T) {
	serverPeer, clientPeer := net.Pipe()
	defer serverPeer.Close()
	defer clientPeer.Close()
	serverPeer.SetWriteDeadline(time.Now().Add(5 * time.Second))
	clientPeer.SetWriteDeadline(time.Now().Add(5 * time.Second))

	// Remote flags describe the opposite peer: the server owns the attacker,
	// and the client owns the defender.
	serverRemote := MakeRemote(false, false, serverPeer)
	clientRemote := MakeRemote(true, true, clientPeer)
	makeModel := func(remote *Remote) *CombatModel {
		makeArmy := func(banner data.BannerType) *Army {
			army := &Army{Player: makeTestCombatPlayer(true)}
			template := units.LizardSpearmen
			template.Count = 1
			template.HitPoints = 1
			template.Defense = 1
			template.MeleeAttackPower = 100
			template.MovementSpeed = 2
			for range 2 {
				// Guaranteed hits make simultaneous lethal melee deterministic.
				army.AddUnit(&OverrideToHitMelee{units.MakeOverworldUnitFromUnit(template, 0, 0, data.PlaneArcanus, banner, &units.NoExperienceInfo{}, &units.NoEnchantments{})})
			}
			return army
		}
		model := MakeCombatModel(spellbook.Spells{}, makeArmy(data.BannerRed), makeArmy(data.BannerBlue), CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10), remote)
		for _, army := range []*Army{model.DefendingArmy, model.AttackingArmy} {
			for _, unit := range army.units {
				model.setTileUnit(unit.X, unit.Y, nil)
			}
		}
		for i, unit := range model.DefendingArmy.units {
			unit.X, unit.Y = 10-2*i, 10
			unit.MovesLeft = fraction.FromInt(2)
			unit.LastTurn = model.CurrentTurn - 1
			model.setTileUnit(unit.X, unit.Y, unit)
		}
		for i, unit := range model.AttackingArmy.units {
			unit.X, unit.Y = 11+2*i, 10
			unit.LastTurn = model.CurrentTurn - 1
			model.setTileUnit(unit.X, unit.Y, unit)
		}
		model.Turn = TeamDefender
		model.SelectedUnit = model.DefendingArmy.units[0]
		return model
	}
	server, client := makeModel(serverRemote), makeModel(clientRemote)
	if !server.IsServer() || client.IsServer() {
		test.Fatal("Expected attacker server and defender client")
	}

	clientMover := client.DefendingArmy.units[0]
	clientTarget := client.AttackingArmy.units[0]
	serverMover := server.DefendingArmy.units[0]
	serverTarget := server.AttackingArmy.units[0]
	clientNext := client.DefendingArmy.units[1]
	serverNext := server.DefendingArmy.units[1]
	initialTurn := client.CurrentTurn
	if !client.canMeleeAttack(clientMover, clientTarget, true) || clientMover.GetMovementSpeed().LessThan(fraction.FromInt(2)) {
		test.Fatal("Expected defender to be able to melee twice in one turn")
	}

	quit, cancel := context.WithCancel(context.Background())
	var receivers sync.WaitGroup
	for _, remote := range []*Remote{serverRemote, clientRemote} {
		receivers.Go(func() { remote.RunReceiveLoop(quit) })
	}
	defer func() {
		cancel()
		receivers.Wait()
	}()

	// Initiate the attack through the same selection/update path as a click.
	client.Update(&ProxySpellSystem{Model: client}, &remoteMeleeTestActions{ProxyActions: &ProxyActions{Model: client}, test: test}, true, clientTarget.X, clientTarget.Y)

	// A marker bounds replay of every command, damage, death and turn event.
	if err := clientRemote.SendEvent(&RemoteGlobalSpellEvent{Type: RemoteGlobalSpellType, Name: "mutual-melee-marker"}); err != nil {
		test.Fatal(err)
	}
	// The first queued event is the melee command. Its action consumes the
	// ordered result stream through melee_finished before Update returns.
	server.Update(&ProxySpellSystem{Model: server}, &remoteMeleeTestActions{ProxyActions: &ProxyActions{Model: server}, test: test}, false, 0, 0)
	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()
	for {
		select {
		case event := <-serverRemote.Events:
			if marker, ok := event.(*RemoteGlobalSpellEvent); ok && marker.Name == "mutual-melee-marker" {
				goto assertions
			}
			server.HandleRemoteEvent(optional.Empty[SpellSystem](), optional.Empty[AddDamageIndicators](), event)
		case <-timeout.C:
			test.Fatal("Timed out replaying defender's melee on the server")
		}
	}

assertions:
	for _, peer := range []struct {
		name                string
		model               *CombatModel
		mover, target, next *ArmyUnit
	}{
		{"client", client, clientMover, clientTarget, clientNext},
		{"server", server, serverMover, serverTarget, serverNext},
	} {
		for _, dead := range []*ArmyUnit{peer.mover, peer.target} {
			if dead.GetHealth() != 0 || peer.model.GetUnitById(dead.Id) != nil || peer.model.GetArmyForTeam(dead.Team).GetKilledUnitById(dead.Id) != dead {
				test.Errorf("%s: expected melee participant %d to be dead and removed", peer.name, dead.Id)
			}
		}
		if len(peer.model.DefendingArmy.units) != 1 || len(peer.model.AttackingArmy.units) != 1 {
			test.Errorf("%s: expected one surviving unit in each army", peer.name)
		}
		if peer.model.SelectedUnit != peer.next {
			test.Errorf("%s: expected defender %d selected instead of the dead melee unit (still selected: %v)", peer.name, peer.next.Id, peer.model.SelectedUnit == peer.mover)
		}
		if peer.model.Turn != TeamDefender || peer.model.CurrentTurn != initialTurn {
			test.Errorf("%s: mutual death advanced the team or round", peer.name)
		}
		if peer.next.LastTurn >= initialTurn || !peer.next.MovesLeft.Equals(fraction.FromInt(2)) {
			test.Errorf("%s: mutual death consumed the next defender's turn or movement", peer.name)
		}
	}
}
