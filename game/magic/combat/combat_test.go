package combat

import (
    "log"
    "net"
    "context"
    "time"
    "testing"
    "math"
    "slices"

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
    remote := MakeRemote(true, false, peer1)

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
    remoteDefender := MakeRemote(false, true, peer2)

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

func TestRemoteUnitCastProjectile(test *testing.T) {

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

        // remote side is defender, so false for isAttacker
        remote := MakeRemote(true, false, peer1)

        defendingArmy := &Army{Player: makeTestCombatPlayer(false)}
        attackingArmy := &Army{Player: makeTestCombatPlayer(false)}

        useUnit := units.HellHounds
        // hack to get warped wood to work
        useUnit.RangedAttackPower = 1
        useUnit.RangedAttacks = 8
        useUnit.RangedAttackDamageType = units.DamageRangedPhysical
        defendingOverworld := units.MakeOverworldUnitFromUnit(useUnit, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})

        for _, option := range testOptions {
            if option.MakeUnit != nil {
                defendingOverworld = option.MakeUnit()
            }
        }

        defendingUnit := defendingArmy.AddUnit(defendingOverworld)

        // an enchantment that can be removed via dispel magic
        defendingUnit.AddEnchantment(data.UnitEnchantmentGiantStrength)

        // attacking army has a unit that can range attack. grant very high tohit
        attacker := units.MakeOverworldUnitFromUnit(units.Warlocks, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
        attackingUnit := attackingArmy.AddUnit(&OverrideToHitMelee{attacker})

        // a different unit that can be healed
        attacker2 := units.MakeOverworldUnitFromUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus, data.BannerRed, &units.NoExperienceInfo{}, &units.NoEnchantments{})
        attackingFriend := attackingArmy.AddUnit(attacker2)
        // set health to half so we can test healing
        attackingFriend.TakeDamage(attackingFriend.GetMaxHealth() / 2, DamageNormal)

        model := MakeCombatModel(allSpells, defendingArmy, attackingArmy, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10), remote)
        model.CracksCallChance = 100

        quit, cancel := context.WithCancel(context.Background())
        defer cancel()

        // remote side is attacker, so true for isAttacker
        remoteDefender := MakeRemote(false, true, peer2)

        go remoteDefender.RunReceiveLoop(quit)

        finish := make(chan struct{})

        expectedId := defendingUnit.Id
        if !defender {
            expectedId = attackingFriend.Id
        }

        go func() {
            for quit.Err() == nil {
                select {
                    case <-quit.Done():
                        return
                    case event := <-remoteDefender.Events:
                        expecter.HandleEvent(event, expectedId, spellName, model)
                }

                if expecter.Finished() {
                    close(finish)
                    break
                }
            }
        }()

        spellSystem := proxySpellSystem{
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
        }

        model.InvokeSpell(&spellSystem, attackingArmy, attackingUnit, spellbook.Spell{Name: spellName}, func(success bool) { })

        var counter uint64
        for model.UpdateProjectiles(counter, &FakeDamageIndicator{}) && counter < 10000 {
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
    }

    enemyDamageUnitSpells := []string{
        "Fireball", "Ice Bolt", "Star Fires",
        "Psionic Blast", "Doom Bolt", "Fire Bolt",
        "Lightning Bolt", "Warp Lightning", "Life Drain",
        "Dispel Evil", "Banish", "Word of Death",
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

    makeCurseUnitExpecter := func(curse data.UnitEnchantment) Expecter {
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
                    if event.Id == expectedId && event.Curse == curse {
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

    doSpellTest("Warp Wood", true, makeSetRangedAttacksExpecter())

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
}
