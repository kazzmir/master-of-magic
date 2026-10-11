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

func (system *TestSpellSystem) PlaySound(spell spellbook.Spell) {
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
func (system *TestSpellSystem) CreateLifeDrainProjectile(target *ArmyUnit, reduceResistance int, player ArmyPlayer, unitCaster *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateDispelEvilProjectile(target *ArmyUnit, reduceResistance int) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateHealingProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateHolyWordProjectile(target *ArmyUnit, reduceResistance int) *Projectile {
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
func (system *TestSpellSystem) CreateDispelMagicProjectile(target *ArmyUnit, caster ArmyPlayer, dispelStrength int) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateWordOfRecallProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateDisintegrateProjectile(target *ArmyUnit, reduceResistance int) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateDisruptProjectile(x int, y int) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateMagicVortex(team Team, x int, y int) *MagicVortex {
    return nil
}
func (system *TestSpellSystem) CreateWarpWoodProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateDeathSpellProjectile(target *ArmyUnit, reduceResistance int) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateWordOfDeathProjectile(target *ArmyUnit, reduceResistance int) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateSummoningCircle(x int, y int) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateMindStormProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateBlessProjectile(target *ArmyUnit) *Projectile {
    return nil
}
func (system *TestSpellSystem) CreateWeaknessProjectile(target *ArmyUnit, reduceResistance int) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateBlackSleepProjectile(target *ArmyUnit, reduceResistance int) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateVertigoProjectile(target *ArmyUnit, reduceResistance int) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateShatterProjectile(target *ArmyUnit, reduceResistance int) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateWarpCreatureProjectile(target *ArmyUnit, reduceResistance int) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateConfusionProjectile(target *ArmyUnit, reduceResistance int) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreatePossessionProjectile(target *ArmyUnit, reduceResistance int) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateCreatureBindingProjectile(target *ArmyUnit, reduceResistance int) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreatePetrifyProjectile(target *ArmyUnit, reduceResistance int) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateChaosChannelsProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateHeroismProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateHolyArmorProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateHolyWeaponProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateInvulnerabilityProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateLionHeartProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateRighteousnessProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateTrueSightProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateElementalArmorProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateGiantStrengthProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateIronSkinProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateStoneSkinProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateRegenerationProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateResistElementsProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateFlightProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateGuardianWindProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateHasteProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateInvisibilityProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateMagicImmunityProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateResistMagicProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateSpellLockProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateEldritchWeaponProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateFlameBladeProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateImmolationProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateBerserkProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateCloakOfFearProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) CreateWraithFormProjectile(target *ArmyUnit) *Projectile {
    return nil
}

func (system *TestSpellSystem) GetAllSpells() spellbook.Spells {
    return spellbook.Spells{}
}

func TestFireballSpell(test *testing.T){
    defendingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, false, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
    }

    // events are only produced if the player is human
    attackingArmy := &Army{
        Player: playerlib.MakePlayer(setup.WizardCustom{}, true, 1, 1, map[herolib.HeroType]string{}, &playerlib.NoGlobalEnchantments{}),
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
            if target != defendingArmy.units[0] {
                test.Errorf("Expected the defender to be targeted")
            }

            return nil
        },
    }

    combat.InvokeSpell(spellObject, attackingArmy, nil, fireball, func(success bool){
        casted = true
    })

    select {
        case event := <-events:
            selectUnit, ok := event.(*CombatEventSelectUnit)
            if !ok {
                test.Errorf("Expected event to be select unit")
            }

            selectUnit.SelectTarget(defendingArmy.units[0])
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

func TestAITarget(test *testing.T) {
    makeUnit := func(unitType units.Unit, enchantments... data.UnitEnchantment) *ArmyUnit {
        unit := &ArmyUnit{
            Unit: units.MakeOverworldUnit(unitType, 0, 0, data.PlaneArcanus),
        }

        for _, enchantment := range enchantments {
            unit.AddEnchantment(enchantment)
        }

        return unit
    }

    // can target a basic unit without bless
    if !shouldAITargetUnit(makeUnit(units.LizardSwordsmen), spellbook.Spell{Name: "Bless"}) {
        test.Errorf("Expected AI to target basic unit without bless")
    }

    // can not target a basic unit with bless
    if shouldAITargetUnit(makeUnit(units.LizardSwordsmen, data.UnitEnchantmentBless), spellbook.Spell{Name: "Bless"}) {
        test.Errorf("Expected AI not to target basic unit with bless")
    }

    // can not target a basic flying unit with flight
    if shouldAITargetUnit(makeUnit(units.DraconianSwordsmen), spellbook.Spell{Name: "Flight"}) {
        test.Errorf("Expected AI not to target flying unit with flight")
    }

    if shouldAITargetUnit(makeUnit(units.Wraith), spellbook.Spell{Name: "Wraith Form"}) {
        test.Errorf("Expected AI not to target wraith with wraith form")
    }
}
