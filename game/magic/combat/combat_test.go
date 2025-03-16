package combat

import (
    // "log"
    "testing"
    "math"
    "slices"

    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
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
    unit := units.MakeOverworldUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus)
    armyUnit := ArmyUnit{
        Unit: unit,
    }

    if armyUnit.Figures() != 8 {
        test.Errorf("Error: figures should be 8")
    }

    if armyUnit.GetDamage() != 0 {
        test.Errorf("Error: damage should be 0")
    }

    // each figure has 2 hp, so taking one damage should keep 8 figures
    armyUnit.TakeDamage(1)

    if armyUnit.Figures() != 8 {
        test.Errorf("Error: figures should be 8")
    }

    if armyUnit.GetDamage() != 1 {
        test.Errorf("Error: damage should be 1")
    }

    // kill one figure
    armyUnit.TakeDamage(1)

    if armyUnit.Figures() != 7 {
        test.Errorf("Error: figures should be 7")
    }

    if armyUnit.GetDamage() != 2 {
        test.Errorf("Error: damage should be 2")
    }
}

type TestObserver struct {
    Melee func(attacker *ArmyUnit, defender *ArmyUnit, damageRoll int)
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

func (observer *TestObserver) MeleeAttack(attacker *ArmyUnit, defender *ArmyUnit, damageRoll int){
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
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}),
    }

    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}),
    }

    defender := units.MakeOverworldUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus)
    attacker := units.MakeOverworldUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus)

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
        Melee: func(meleeAttacker *ArmyUnit, meleeDefender *ArmyUnit, damageRoll int){
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

func TestAttackerHaste(test *testing.T){
    defendingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}),
    }

    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}),
    }

    defender := units.MakeOverworldUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus)
    attacker := units.MakeOverworldUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus)

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
        Melee: func(meleeAttacker *ArmyUnit, meleeDefender *ArmyUnit, damageRoll int){
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
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}),
    }

    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}),
    }

    attackerUnit := units.LizardSpearmen
    attackerUnit.Abilities = append(slices.Clone(attackerUnit.Abilities), data.MakeAbility(data.AbilityFirstStrike))
    // ensure attacker can kill the defender in one hit
    attackerUnit.MeleeAttackPower = 10000

    defender := units.MakeOverworldUnit(units.LizardSpearmen, 0, 0, data.PlaneArcanus)
    attacker := units.MakeOverworldUnit(attackerUnit, 0, 0, data.PlaneArcanus)

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
        Melee: func(meleeAttacker *ArmyUnit, meleeDefender *ArmyUnit, damageRoll int){
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
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}),
    }

    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}),
    }

    attackerUnit := units.LizardSpearmen
    attackerUnit.Abilities = append(slices.Clone(attackerUnit.Abilities), data.MakeAbility(data.AbilityFirstStrike))
    // ensure attacker can kill the defender in one hit
    attackerUnit.MeleeAttackPower = 10000

    defenderUnit := units.LizardSpearmen
    defenderUnit.Abilities = append(slices.Clone(defenderUnit.Abilities), data.MakeAbility(data.AbilityNegateFirstStrike))

    defender := units.MakeOverworldUnit(defenderUnit, 0, 0, data.PlaneArcanus)
    attacker := units.MakeOverworldUnit(attackerUnit, 0, 0, data.PlaneArcanus)

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
        Melee: func(meleeAttacker *ArmyUnit, meleeDefender *ArmyUnit, damageRoll int){
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
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}),
    }

    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}),
    }

    attackerUnit := units.LizardSpearmen
    attackerUnit.Abilities = append(slices.Clone(attackerUnit.Abilities), data.MakeAbilityValue(data.AbilityThrown, 10000), data.MakeAbilityValue(data.AbilityToHit, 100))
    // ensure attacker can kill the defender in one hit
    attackerUnit.MeleeAttackPower = 10000

    defenderUnit := units.LizardSpearmen

    defender := units.MakeOverworldUnit(defenderUnit, 0, 0, data.PlaneArcanus)
    attacker := units.MakeOverworldUnit(attackerUnit, 0, 0, data.PlaneArcanus)

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
        Melee: func(meleeAttacker *ArmyUnit, meleeDefender *ArmyUnit, damageRoll int){
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
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}),
    }

    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}),
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

    defender := units.MakeOverworldUnit(defenderUnit, 0, 0, data.PlaneArcanus)
    attacker := units.MakeOverworldUnit(attackerUnit, 0, 0, data.PlaneArcanus)

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
        Melee: func(meleeAttacker *ArmyUnit, meleeDefender *ArmyUnit, damageRoll int){
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
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}),
    }

    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}),
    }

    attackerUnit := units.LizardSpearmen
    attackerUnit.Abilities = append(slices.Clone(attackerUnit.Abilities), data.MakeAbility(data.AbilityCauseFear))

    defenderUnit := units.LizardSwordsmen
    // ensure all units become afraid
    defenderUnit.Resistance = -100

    defender := units.MakeOverworldUnit(defenderUnit, 0, 0, data.PlaneArcanus)
    attacker := units.MakeOverworldUnit(attackerUnit, 0, 0, data.PlaneArcanus)

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
        Melee: func(meleeAttacker *ArmyUnit, meleeDefender *ArmyUnit, damageRoll int){
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
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}),
    }

    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}),
    }

    attackerUnit := units.LizardSpearmen
    defenderUnit := units.LizardSwordsmen

    defender := units.MakeOverworldUnit(defenderUnit, 0, 0, data.PlaneArcanus)
    attacker := units.MakeOverworldUnit(attackerUnit, 0, 0, data.PlaneArcanus)

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
