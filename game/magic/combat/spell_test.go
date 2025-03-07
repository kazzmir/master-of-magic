package combat

import (
    "testing"

    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    "github.com/kazzmir/master-of-magic/game/magic/setup"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/data"
)

type TestSpellSystem struct {
    createFireballProjectile func(target *ArmyUnit, cost int) *Projectile
}

func (system *TestSpellSystem) CreateFireballProjectile(target *ArmyUnit, cost int) *Projectile {
    if system.createFireballProjectile != nil {
        return system.createFireballProjectile(target, cost)
    }

    return nil
}

func (system *TestSpellSystem) CreateIceBoltProjectile(target *ArmyUnit, cost int) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateStarFiresProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreatePsionicBlastProjectile(target *ArmyUnit, cost int) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateDoomBoltProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateFireBoltProjectile(target *ArmyUnit, cost int) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateLightningBoltProjectile(target *ArmyUnit, cost int) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateWarpLightningProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateFlameStrikeProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateLifeDrainProjectile(target *ArmyUnit, reduceResistance int, player *playerlib.Player, unitCaster *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateDispelEvilProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateHealingProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateHolyWordProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateRecallHeroProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateCracksCallProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateWebProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateBanishProjectile(target *ArmyUnit, reduceResistance int) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateDispelMagicProjectile(target *ArmyUnit, caster *playerlib.Player, dispelStrength int) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateWordOfRecallProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateDisintegrateProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateDisruptProjectile(x int, y int) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateMagicVortex(x int, y int) *OtherUnit {
    return nil
}
func (system *TestSpellSystem) CreateWarpWoodProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateDeathSpellProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateWordOfDeathProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateSummoningCircle(x int, y int) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateResistElementsProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateMindStormProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateBlessProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateWeaknessProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateBlackSleepProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateVertigoProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) GetAllSpells() spellbook.Spells {
    return spellbook.Spells{}
}

func TestFireballSpell(test *testing.T){
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

    events := make(chan CombatEvent, 10)

    combat := &CombatModel{
        SelectedUnit: nil,
        Tiles: makeTiles(30, 30, CombatLandscapeGrass, data.PlaneArcanus, ZoneType{}),
        Turn: TeamDefender,
        DefendingArmy: defendingArmy,
        AttackingArmy: attackingArmy,
        Events: events,
    }

    combat.Initialize(spellbook.Spells{}, 0, 0)

    fireball := spellbook.Spell{
        Name: "Fireball",
        CastCost: 10,
        Magic: data.ChaosMagic,
    }

    casted := false
    createdFireball := false

    spellObject := &TestSpellSystem{
        createFireballProjectile: func (target *ArmyUnit, cost int) *Projectile {
            createdFireball = true
            if target != defendingArmy.Units[0] {
                test.Errorf("Expected the defender to be targeted")
            }

            return nil
        },
    }

    combat.InvokeSpell(spellObject, attackingArmy.Player, nil, fireball, func(){
        casted = true
    })

    select {
        case event := <-events:
            selectUnit, ok := event.(*CombatEventSelectUnit)
            if !ok {
                test.Errorf("Expected event to be select unit")
            }

            selectUnit.SelectTarget(defendingArmy.Units[0])
        default:
            test.Errorf("Expected select unit event")
    }

    if !casted {
        test.Errorf("Error: fireball should have been cast")
    }

    if !createdFireball {
        test.Errorf("Error: fireball should have created a projectile")
    }
}
