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
        model.CreateBlackSleepProjectileEffect()(unit)
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

    model := MakeCombatModel(allSpells, defendingArmy, attackingArmy, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}, data.MagicNone, 0, 0, make(chan CombatEvent, 10))

    state := Run(model)
    if state != CombatStateAttackerWin {
        test.Errorf("Error: attacker should have won the combat, got state %d", state)
    }
}
