package game

import (
    "fmt"
    "log"
    "image"
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    "github.com/kazzmir/master-of-magic/game/magic/music"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/cityview"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/game/magic/inputmanager"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/game/magic/unitview"
    "github.com/kazzmir/master-of-magic/game/magic/audio"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"
    "github.com/kazzmir/master-of-magic/game/magic/camera"
    "github.com/kazzmir/master-of-magic/game/magic/building"

    "github.com/hajimehoshi/ebiten/v2"
)

type LocationType int
const (
    LocationTypeAny LocationType = iota
    LocationTypeLand
    LocationTypeEmptyWater
    LocationTypeFriendlyCity
    LocationTypeEnemyCity
    LocationTypeFriendlyUnit
    LocationTypeEnemyUnit
    LocationTypeChangeTerrain
    LocationTypeTransmute
    LocationTypeRaiseVolcano
    LocationTypeEnemyMeldedNode
    LocationTypeDisenchant
)

type UnitEnchantmentCallback func (units.StackUnit) bool

func noUnitEnchantmentCallback(unit units.StackUnit) bool {
    return true
}

type CityCallback func (*citylib.City) bool

func noCityCallback(city *citylib.City) bool {
    return true
}

func (game *Game) doCastSpell(player *playerlib.Player, spell spellbook.Spell) {
    // FIXME: if the player is AI then invoke some callback that the AI will use to select targets instead of using the GameEventSelectLocationForSpell

    if game.checkInstantFizzleForCastSpell(player, spell) {
        // Fizzle the spell and return
        game.ShowFizzleSpell(spell, player)
        return
    }

    switch spell.Name {
        /*
            SUMMONING SPELLS
        */
        case "Magic Spirit", "Angel", "Arch Angel", "Guardian Spirit",
             "Unicorns", "Basilisk", "Behemoth", "Cockatrices", "Colossus",
             "Earth Elemental", "Giant Spiders", "Gorgons", "Great Wyrm",
             "Sprites", "Stone Giant", "War Bears", "Air Elemental",
             "Djinn", "Nagas", "Phantom Beast", "Phantom Warriors",
             "Sky Drake", "Storm Giant", "Chaos Spawn", "Chimeras", "Doom Bat",
             "Efreet", "Fire Elemental", "Fire Giant", "Gargoyles",
             "Great Drake", "Hell Hounds", "Hydra", "Death Knights",
             "Demon Lord", "Ghouls", "Night Stalker", "Shadow Demons",
             "Skeletons", "Wraiths":
            game.doSummonUnit(player, units.GetUnitByName(spell.Name))
        case "Floating Island":
            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                game.doCastFloatingIsland(yield, player, tileX, tileY)
            }

            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeEmptyWater, SelectedFunc: selected}
        case "Summon Hero":
            game.doSummonHero(player, false)
        case "Summon Champion":
            game.doSummonHero(player, true)
        case "Incarnation":
            game.doIncarnation(player)

        /*
            UNIT ENCHANTMENTS
        */
        case "Bless":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentBless)
        case "Heroism":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentHeroism)
        case "Giant Strength":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentGiantStrength)
        case "Lionheart":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentLionHeart)
        case "Immolation":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentImmolation)
        case "Resist Elements":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentResistElements)
        case "Resist Magic":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentResistMagic)
        case "Elemental Armor":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentElementalArmor)
        case "Righteousness":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentRighteousness)
        case "Cloak of Fear":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentCloakOfFear)
        case "True Sight":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentTrueSight)
        case "Flight":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentFlight)
        case "Chaos Channels":
            choices := []data.UnitEnchantment{
                data.UnitEnchantmentChaosChannelsDemonSkin,
                data.UnitEnchantmentChaosChannelsDemonWings,
                data.UnitEnchantmentChaosChannelsFireBreath,
            }
            game.doCastUnitEnchantment(player, spell, choices[rand.N(len(choices))])
        case "Endurance":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentEndurance)
        case "Holy Armor":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentHolyArmor)
        case "Holy Weapon":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentHolyWeapon)
        case "Invulnerability":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentInvulnerability)
        case "Planar Travel":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentPlanarTravel)
        case "Iron Skin":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentIronSkin)
        case "Path Finding":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentPathFinding)
        case "Regeneration":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentRegeneration)
        case "Stone Skin":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentStoneSkin)
        case "Water Walking":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentWaterWalking)
        case "Guardian Wind":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentGuardianWind)
        // "Invisiblity" is a typo in the game in spelldat.lbx
        case "Invisiblity":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentInvisibility)
        case "Magic Immunity":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentMagicImmunity)
        case "Spell Lock":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentSpellLock)
        case "Wind Walking":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentWindWalking)
        case "Eldritch Weapon":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentEldritchWeapon)
        case "Flame Blade":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentFlameBlade)
        case "Black Channels":
            after := func (unit units.StackUnit) bool {
                unit.SetUndead()
                return true
            }
            game.doCastUnitEnchantmentFull(player, spell, data.UnitEnchantmentBlackChannels, noUnitEnchantmentCallback, after)
        case "Wraith Form":
            game.doCastUnitEnchantment(player, spell, data.UnitEnchantmentWraithForm)
        case "Lycanthropy":
            before := func (unit units.StackUnit) bool {
                if unit.GetRace() == data.RaceFantastic {
                    game.Events <- &GameEventNotice{Message: "Summoned units may not have cast Lycanthropy on them"}
                    return false
                }

                if unit.GetRace() == data.RaceHero {
                    game.Events <- &GameEventNotice{Message: "Heroes units may not have cast Lycanthropy on them"}
                    return false
                }

                if unit.IsUndead() || unit.HasEnchantment(data.UnitEnchantmentBlackChannels) || unit.HasEnchantment(data.UnitEnchantmentChaosChannelsDemonSkin) || unit.HasEnchantment(data.UnitEnchantmentChaosChannelsDemonWings) || unit.HasEnchantment(data.UnitEnchantmentChaosChannelsFireBreath) {
                    game.Events <- &GameEventNotice{Message: "Chaos channeled and Undead units may not have cast Lycanthropy on them"}
                    return false
                }
               return true
            }
            after := func (unit units.StackUnit) bool {
                unit.RemoveEnchantment(data.UnitEnchantmentLycanthropy)
                overworldUnit, ok := unit.(*units.OverworldUnit)
                if ok {
                    damage := overworldUnit.GetMaxHealth() - overworldUnit.Health
                    overworldUnit.Unit = units.WereWolf
                    overworldUnit.Health = overworldUnit.GetMaxHealth() - damage
                    overworldUnit.Experience = 0
                    // unit keeps weapon bonus and enchantments
                }
                return true
            }
            game.doCastUnitEnchantmentFull(player, spell, data.UnitEnchantmentLycanthropy, before, after)

        /*
            TOWN ENCHANTMENTS
                TODO:
                Earth Gate
                Flying Fortress
                Spell Ward
        */
        case "Astral Gate":
            game.doCastCityEnchantment(spell, player, LocationTypeFriendlyCity, data.CityEnchantmentAstralGate)
        case "Heavenly Light":
            game.doCastCityEnchantment(spell, player, LocationTypeFriendlyCity, data.CityEnchantmentHeavenlyLight)
        case "Cloud of Shadow":
            game.doCastCityEnchantment(spell, player, LocationTypeFriendlyCity, data.CityEnchantmentCloudOfShadow)
        case "Prosperity":
            game.doCastCityEnchantment(spell, player, LocationTypeFriendlyCity, data.CityEnchantmentProsperity)
        case "Consecration":
            after := func (city *citylib.City) bool {
                // Remove all the city curses from the targeted city on cast (https://masterofmagic.fandom.com/wiki/Consecration)
                // Wiki specifies only the following ones as city curses
                city.RemoveEnchantments(
                    data.CityEnchantmentChaosRift,
                    data.CityEnchantmentCursedLands,
                    data.CityEnchantmentEvilPresence,
                    data.CityEnchantmentFamine,
                    data.CityEnchantmentPestilence,
                )
                return true
            }
            game.doCastCityEnchantmentFull(spell, player, LocationTypeFriendlyCity, data.CityEnchantmentConsecration, noCityCallback, after)
        case "Inspirations":
            game.doCastCityEnchantment(spell, player, LocationTypeFriendlyCity, data.CityEnchantmentInspirations)
        case "Gaia's Blessing":
            game.doCastCityEnchantment(spell, player, LocationTypeFriendlyCity, data.CityEnchantmentGaiasBlessing)
        case "Nature's Eye":
            after := func (city *citylib.City) bool {
                player.UpdateFogVisibility()
                return true
            }
            game.doCastCityEnchantmentFull(spell, player, LocationTypeFriendlyCity, data.CityEnchantmentNaturesEye, noCityCallback, after)
        case "Dark Rituals":
            game.doCastCityEnchantment(spell, player, LocationTypeFriendlyCity, data.CityEnchantmentDarkRituals)
        case "Stream of Life":
            game.doCastCityEnchantment(spell, player, LocationTypeFriendlyCity, data.CityEnchantmentStreamOfLife)
        case "Altar of Battle":
            game.doCastCityEnchantment(spell, player, LocationTypeFriendlyCity, data.CityEnchantmentAltarOfBattle)
        case "Wall of Fire":
            game.doCastCityEnchantment(spell, player, LocationTypeFriendlyCity, data.CityEnchantmentWallOfFire)
        case "Wall of Darkness":
            game.doCastCityEnchantment(spell, player, LocationTypeFriendlyCity, data.CityEnchantmentWallOfDarkness)

        /*
            TOWN CURSES
                TODO:
                Chaos Rift
        */
        case "Cursed Lands":
            before := func (city *citylib.City) bool {
                if city.HasAnyOfEnchantments(data.CityEnchantmentConsecration, data.CityEnchantmentDeathWard) {
                    game.ShowFizzleSpell(spell, player)
                    return false
                }
                return true
            }
            game.doCastCityEnchantmentFull(spell, player, LocationTypeEnemyCity, data.CityEnchantmentCursedLands, before, noCityCallback)
        case "Famine":
            before := func (city *citylib.City) bool {
                if city.HasAnyOfEnchantments(data.CityEnchantmentConsecration, data.CityEnchantmentDeathWard) {
                    game.ShowFizzleSpell(spell, player)
                    return false
                }
                return true
            }
            game.doCastCityEnchantmentFull(spell, player, LocationTypeEnemyCity, data.CityEnchantmentFamine, before, noCityCallback)
        case "Pestilence":
            before := func (city *citylib.City) bool {
                if city.HasAnyOfEnchantments(data.CityEnchantmentConsecration, data.CityEnchantmentDeathWard) {
                    game.ShowFizzleSpell(spell, player)
                    return false
                }
                return true
            }
            game.doCastCityEnchantmentFull(spell, player, LocationTypeEnemyCity, data.CityEnchantmentPestilence, before, noCityCallback)
        case "Evil Presence":
            before := func (city *citylib.City) bool {
                if city.HasAnyOfEnchantments(data.CityEnchantmentConsecration, data.CityEnchantmentDeathWard) {
                    game.ShowFizzleSpell(spell, player)
                    return false
                }
                return true
            }
            game.doCastCityEnchantmentFull(spell, player, LocationTypeEnemyCity, data.CityEnchantmentEvilPresence, before, noCityCallback)


        /*
            GLOBAL ENCHANTMENTS
                TODO:
                Charm of Life
                Holy Arms
                Planar Seal
                Herb Mastery
                Nature's Wrath
                Aura of Majesty
                Suppress Magic
                Time Stop
                Wind Mastery
                Chaos Surge
                Doom Mastery
                Meteor Storm
                Evil Omens
                Zombie Mastery
        */
        case "Awareness":
            if !player.GlobalEnchantments.Contains(data.EnchantmentAwareness) {
                game.Events <- &GameEventCastGlobalEnchantment{Player: player, Enchantment: data.EnchantmentAwareness}

                player.GlobalEnchantments.Insert(data.EnchantmentAwareness)
                game.doExploreFogForAwareness(player)

                game.RefreshUI()
            }
        case "Nature Awareness":
            if !player.GlobalEnchantments.Contains(data.EnchantmentNatureAwareness) {
                game.Events <- &GameEventCastGlobalEnchantment{Player: player, Enchantment: data.EnchantmentNatureAwareness}

                player.GlobalEnchantments.Insert(data.EnchantmentNatureAwareness)
                player.LiftFogAll(data.PlaneArcanus)
                player.LiftFogAll(data.PlaneMyrror)
                game.RefreshUI()
            }
        case "Crusade":
            if !player.GlobalEnchantments.Contains(data.EnchantmentCrusade) {
                game.Events <- &GameEventCastGlobalEnchantment{Player: player, Enchantment: data.EnchantmentCrusade}

                player.GlobalEnchantments.Insert(data.EnchantmentCrusade)

                game.RefreshUI()
            }
        case "Just Cause":
            if !player.GlobalEnchantments.Contains(data.EnchantmentJustCause) {
                game.Events <- &GameEventCastGlobalEnchantment{Player: player, Enchantment: data.EnchantmentJustCause}

                player.GlobalEnchantments.Insert(data.EnchantmentJustCause)
                player.UpdateUnrest()

                game.RefreshUI()
            }
        case "Life Force":
            if !player.GlobalEnchantments.Contains(data.EnchantmentLifeForce) {
                game.Events <- &GameEventCastGlobalEnchantment{Player: player, Enchantment: data.EnchantmentLifeForce}

                player.GlobalEnchantments.Insert(data.EnchantmentLifeForce)
                game.RefreshUI()
            }
        case "Tranquility":
            if !player.GlobalEnchantments.Contains(data.EnchantmentTranquility) {
                game.Events <- &GameEventCastGlobalEnchantment{Player: player, Enchantment: data.EnchantmentTranquility}

                player.GlobalEnchantments.Insert(data.EnchantmentTranquility)
                game.RefreshUI()
            }
        case "Armageddon":
            if !player.GlobalEnchantments.Contains(data.EnchantmentArmageddon) {
                game.Events <- &GameEventCastGlobalEnchantment{Player: player, Enchantment: data.EnchantmentArmageddon}

                player.GlobalEnchantments.Insert(data.EnchantmentArmageddon)

                for _, player := range game.Players {
                    player.UpdateUnrest()
                }

                game.RefreshUI()
            }
        case "Great Wasting":
            if !player.GlobalEnchantments.Contains(data.EnchantmentGreatWasting) {
                game.Events <- &GameEventCastGlobalEnchantment{Player: player, Enchantment: data.EnchantmentGreatWasting}

                player.GlobalEnchantments.Insert(data.EnchantmentGreatWasting)

                for _, player := range game.Players {
                    player.UpdateUnrest()
                }

                game.RefreshUI()
            }
        case "Detect Magic":
            if !player.GlobalEnchantments.Contains(data.EnchantmentDetectMagic) {
                game.Events <- &GameEventCastGlobalEnchantment{Player: player, Enchantment: data.EnchantmentDetectMagic}

                player.GlobalEnchantments.Insert(data.EnchantmentDetectMagic)

                game.RefreshUI()
            }
        case "Eternal Night":
            if !player.GlobalEnchantments.Contains(data.EnchantmentEternalNight) {
                game.Events <- &GameEventCastGlobalEnchantment{Player: player, Enchantment: data.EnchantmentEternalNight}

                player.GlobalEnchantments.Insert(data.EnchantmentEternalNight)

                game.RefreshUI()
            }
        /*
            INSTANT SPELLS
                TODO:
                Disjunction
                Spell of Mastery
                Spell of Return
                Plane Shift
                Resurrection
                Earthquake
                Ice Storm
                Nature's Cures
                Disjunction True
                Great Unsummoning
                Spell Binding
                Stasis
                Word of Recall
                Fire Storm
                Black Wind
                Cruel Unminding
                Death Wish
                Drain Power
                Subversion
        */
        case "Create Artifact", "Enchant Item":
            game.Events <- &GameEventSummonArtifact{Player: player}
            game.Events <- &GameEventVault{CreatedArtifact: player.CreateArtifact, Player: player}
            player.CreateArtifact = nil
        case "Earth Lore":
            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                game.doCastEarthLore(yield, tileX, tileY, player)
            }

            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeAny, SelectedFunc: selected}
        case "Call the Void":
            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                chosenCity, owner := game.FindCity(tileX, tileY, game.Plane)

                if chosenCity.HasAnyOfEnchantments(data.CityEnchantmentConsecration, data.CityEnchantmentChaosWard) {
                    game.ShowFizzleSpell(spell, player)
                    return
                }

                // FIXME: verify the animation and sound
                game.doCastOnMap(yield, tileX, tileY, 12, false, 72, func (x int, y int, animationFrame int) {})
                game.doCallTheVoid(chosenCity, owner)
            }

            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeEnemyCity, SelectedFunc: selected}
        case "Change Terrain":
            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeChangeTerrain, SelectedFunc: game.doCastChangeTerrain}
        case "Transmute":
            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeTransmute, SelectedFunc: game.doCastTransmute}
        case "Raise Volcano":
            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                chosenCity, _ := game.FindCity(tileX, tileY, game.Plane)

                // FIXME: it's not obvious if Chaos Ward prevents Raise Volcano from being cast on city center. Left it here because it sounds logical
                if chosenCity != nil && chosenCity.HasAnyOfEnchantments(data.CityEnchantmentConsecration, data.CityEnchantmentChaosWard) {
                    game.ShowFizzleSpell(spell, player)
                    return
                }
                game.doCastRaiseVolcano(yield, tileX, tileY, player)
            }

            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeRaiseVolcano, SelectedFunc: selected}
        case "Spell Blast":
            game.doCastSpellBlast(player)
        case "Enchant Road":
            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeAny, SelectedFunc: game.doCastEnchantRoad}
        case "Corruption":
            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                chosenCity, _ := game.FindCity(tileX, tileY, game.Plane)

                // FIXME: it's not obvious if Chaos Ward prevents Corruption from being cast on city center. Left it here because it sounds logical
                if chosenCity != nil && chosenCity.HasAnyOfEnchantments(data.CityEnchantmentConsecration, data.CityEnchantmentChaosWard) {
                    game.ShowFizzleSpell(spell, player)
                    return
                }
                game.doCastCorruption(yield, tileX, tileY)
            }

            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeLand, SelectedFunc: selected}
        case "Warp Node":
            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                game.doCastWarpNode(yield, tileX, tileY, player)
            }
            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeEnemyMeldedNode, SelectedFunc: selected}
        case "Disenchant Area":

            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                game.doDisenchantArea(yield, player, spell, false, tileX, tileY)
            }

            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeDisenchant, SelectedFunc: selected}
        case "Disenchant True":

            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                game.doDisenchantArea(yield, player, spell, true, tileX, tileY)
            }

            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeDisenchant, SelectedFunc: selected}
        case "Wall of Stone":
            after := func (city *citylib.City) bool {
                if city.ProducingBuilding == building.BuildingCityWalls {
                    city.ProducingBuilding = building.BuildingTradeGoods
                }
                return true
            }
            game.doCastNewCityBuilding(spell, player, LocationTypeFriendlyCity, building.BuildingCityWalls, "This city already has a Wall of Stone", after)
        case "Summoning Circle":
            after := func(chosenCity *citylib.City) bool {
                for _, city := range player.Cities {
                    if city != chosenCity && city.Buildings.Contains(building.BuildingSummoningCircle) {
                        city.Buildings.Remove(building.BuildingSummoningCircle)
                        break
                    }
                }
                return true
            }
            game.doCastNewCityBuilding(spell, player, LocationTypeFriendlyCity, building.BuildingSummoningCircle, "Your summoning circle is already in this city", after)
        case "Move Fortress":
            after := func(chosenCity *citylib.City) bool {
                player.Wizard.Race = chosenCity.Race

                for _, city := range player.Cities {
                    if city != chosenCity && city.Buildings.Contains(building.BuildingFortress) {
                        city.Buildings.Remove(building.BuildingFortress)

                        if city.Buildings.Contains(building.BuildingSummoningCircle) {
                            chosenCity.Buildings.Insert(building.BuildingSummoningCircle)
                            city.Buildings.Remove(building.BuildingSummoningCircle)
                        }
                        break
                    }
                }

                player.UpdateUnrest()

                return true
            }
            game.doCastNewCityBuilding(spell, player, LocationTypeFriendlyCity, building.BuildingFortress, "Your fortress is already in this city", after)

        default:
            log.Printf("Warning: casting unhandled spell '%v'", spell.Name)
    }
}

// Returns true if the spell is rolled to be instantly fizzled on cast (caused by spells like Life Force)
func (game *Game) checkInstantFizzleForCastSpell(player *playerlib.Player, spell spellbook.Spell) bool {
    // Tranquility effect: if it's a chaos spell, it should either resist a strength 500 dispel check or fizzle right away.
    if spell.IsOfRealm(data.ChaosMagic) {
        for _, checkingPlayer := range game.Players {
            // FIXME: Not sure if multiple instances of Tranquility stack or are checked separately.
            if checkingPlayer != player && checkingPlayer.GlobalEnchantments.Contains(data.EnchantmentTranquility) {
                return spellbook.RollDispelChance(spellbook.ComputeDispelChance(500, spell.Cost(true), spell.Magic, &player.Wizard))
            }
        }
    }
    // Life Force effect: if it's a death spell, it should either resist a strength 500 dispel check or fizzle right away.
    if spell.IsOfRealm(data.DeathMagic) {
        for _, checkingPlayer := range game.Players {
            // FIXME: Not sure if multiple instances of Life Force stack or are checked separately.
            if checkingPlayer != player && checkingPlayer.GlobalEnchantments.Contains(data.EnchantmentLifeForce) {
                return spellbook.RollDispelChance(spellbook.ComputeDispelChance(500, spell.Cost(true), spell.Magic, &player.Wizard))
            }
        }
    }
    return false
}

func (game *Game) doDisenchantArea(yield coroutine.YieldFunc, player *playerlib.Player, spell spellbook.Spell, disenchantTrue bool, tileX int, tileY int) {
    game.doCastOnMap(yield, tileX, tileY, 9, false, spell.Sound, func (x int, y int, animationFrame int){})

    disenchantStrength := spell.Cost(true)
    if disenchantTrue {
        // each additional point of mana spent increases the disenchant strength by 3
        disenchantStrength = spell.BaseCost(true) + spell.SpentAdditionalCost(true) * 3
    }

    if player.Wizard.RetortEnabled(data.RetortRunemaster) {
        disenchantStrength *= 2
    }

    allSpells := game.AllSpells()

    city, _ := game.FindCity(tileX, tileY, game.Plane)
    if city != nil {
        for _, enchantment := range city.Enchantments.Values() {
            if enchantment.Owner != player.GetBanner() {
                spell := allSpells.FindByName(enchantment.Enchantment.SpellName())
                cost := spell.Cost(true)
                dispellChance := spellbook.ComputeDispelChance(disenchantStrength, cost, spell.Magic, &game.GetPlayerByBanner(enchantment.Owner).Wizard)
                if spellbook.RollDispelChance(dispellChance) {
                    city.RemoveEnchantments(enchantment.Enchantment)
                }
            }
        }
    }

    stack, owner := game.FindStack(tileX, tileY, game.Plane)
    if stack != nil && owner != player {
        for _, unit := range stack.Units() {
            var toRemove []data.UnitEnchantment
            for _, enchantment := range unit.GetEnchantments() {
                spell := allSpells.FindByName(enchantment.SpellName())
                cost := spell.Cost(true)
                dispellChance := spellbook.ComputeDispelChance(disenchantStrength, cost, spell.Magic, &owner.Wizard)
                if rand.N(100) < dispellChance {
                    toRemove = append(toRemove, enchantment)
                }
            }

            for _, enchantment := range toRemove {
                unit.RemoveEnchantment(enchantment)
            }
        }
    }

    mapUse := game.GetMap(game.Plane)
    magicNode := mapUse.GetMagicNode(tileX, tileY)
    if magicNode != nil && magicNode.Warped && magicNode.WarpedOwner != player {
        warpNode := allSpells.FindByName("Warp Node")
        cost := warpNode.Cost(true)
        dispellChance := spellbook.ComputeDispelChance(disenchantStrength, cost, warpNode.Magic, &game.GetPlayerByBanner(magicNode.WarpedOwner.GetBanner()).Wizard)

        if rand.N(100) < dispellChance {
            magicNode.Warped = false
        }
    }
}

func (game *Game) doCastUnitEnchantmentFull(player *playerlib.Player, spell spellbook.Spell, enchantment data.UnitEnchantment, before UnitEnchantmentCallback, after UnitEnchantmentCallback) {
    var selected func (yield coroutine.YieldFunc, tileX int, tileY int)
    selected = func (yield coroutine.YieldFunc, tileX int, tileY int){
        game.doMoveCamera(yield, tileX, tileY)
        stack := player.FindStack(tileX, tileY, game.Plane)
        unit := game.doSelectUnit(yield, player, stack)

        // player didn't select a unit, let them pick a different stack
        if unit == nil {
            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeFriendlyUnit, SelectedFunc: selected}
            return
        }

        if unit.HasEnchantment(enchantment) {
            game.Events <- &GameEventNotice{Message: fmt.Sprintf("That unit already has %v cast on it", spell.Name)}
            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeFriendlyUnit, SelectedFunc: selected}
            return
        }

        if !before(unit) {
            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeFriendlyUnit, SelectedFunc: selected}
            return
        }

        game.doCastOnMap(yield, tileX, tileY, enchantment.CastAnimationIndex(), false, spell.Sound, func (x int, y int, animationFrame int) {})
        unit.AddEnchantment(enchantment)

        after(unit)

        game.RefreshUI()
    }

    game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeFriendlyUnit, SelectedFunc: selected}
}

func (game *Game) doCastUnitEnchantment(player *playerlib.Player, spell spellbook.Spell, enchantment data.UnitEnchantment) {
   game.doCastUnitEnchantmentFull(player, spell, enchantment, noUnitEnchantmentCallback, noUnitEnchantmentCallback)
}

func (game *Game) doSelectUnit(yield coroutine.YieldFunc, player *playerlib.Player, stack *playerlib.UnitStack) units.StackUnit {

    drawer := game.Drawer
    defer func(){
        game.Drawer = drawer
    }()

    quit := false

    ui := uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
        LeftClick: func(){
            quit = true
        },
    }

    var chosen units.StackUnit

    var viewUnits []unitview.UnitView
    for _, unit := range stack.Units() {
        viewUnits = append(viewUnits, unit)
    }

    viewElements := unitview.MakeSmallListView(game.Cache, &ui, viewUnits, fmt.Sprintf("%v Units", player.Wizard.Name), func(unit unitview.UnitView){
        quit = true

        if unit != nil {
            stackUnit, ok := unit.(units.StackUnit)
            if ok {
                chosen = stackUnit
            }
        }
    })
    ui.SetElementsFromArray(viewElements)

    yield()

    game.Drawer = func(screen *ebiten.Image, game *Game){
        drawer(screen, game)
        ui.Draw(&ui, screen)
    }

    for !quit {
        game.Counter += 1
        ui.StandardUpdate()

        if yield() != nil {
            return nil
        }
    }

    return chosen
}

func (game *Game) doSummonHero(player *playerlib.Player, champion bool) {
    var choices []*herolib.Hero
    for _, hero := range player.HeroPool {
        // torin is not summonable through this method
        if hero.Status == herolib.StatusAvailable && hero.IsChampion() == champion && hero.HeroType != herolib.HeroTorin {
            choices = append(choices, hero)
        }
    }

    if len(choices) > 0 {
        hero := choices[rand.N(len(choices))]

        summonEvent := GameEventSummonHero{
            Player: player,
            Champion: false,
        }

        select {
            case game.Events <- &summonEvent:
            default:
        }

        event := GameEventHireHero{
            Hero: hero,
            Player: player,
            Cost: 0,
        }

        select {
            case game.Events <- &event:
            default:
        }

        game.RefreshUI()
    }
}

func (game *Game) doIncarnation(player *playerlib.Player) {
    for _, hero := range player.HeroPool {
        if hero.HeroType == herolib.HeroTorin && hero.Status != herolib.StatusEmployed {
            event := GameEventHireHero{
                Hero: hero,
                Player: player,
                Cost: 0,
            }

            select {
                case game.Events <- &event:
                default:
            }
        }
    }
}

func (game *Game) doSummonUnit(player *playerlib.Player, unit units.Unit) {
    select {
        case game.Events <- &GameEventSummonUnit{Player: player, Unit: unit}:
        default:
    }

    summonCity := player.FindSummoningCity()
    if summonCity != nil {
        overworldUnit := units.MakeOverworldUnitFromUnit(unit, summonCity.X, summonCity.Y, summonCity.Plane, player.Wizard.Banner, player.MakeExperienceInfo())
        player.AddUnit(overworldUnit)
        game.RefreshUI()
    }
}

func (game *Game) showCastNewBuilding(yield coroutine.YieldFunc, city *citylib.City, player *playerlib.Player, newBuilding building.Building, name string) {
    ui, quit, err := cityview.MakeNewBuildingView(game.Cache, city, player, newBuilding, name)
    if err != nil {
        log.Printf("Error making new building view: %v", err)
        return
    }

    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    game.Drawer = func(screen *ebiten.Image, game *Game){
        oldDrawer(screen, game)
        ui.Draw(ui, screen)
    }

    for quit.Err() == nil {
        game.Counter += 1
        ui.StandardUpdate()
        if yield() != nil {
            return
        }
    }

    // absorb left click
    yield()
}

/* return x,y and true/false, where true means cancelled, and false means something was selected */
// FIXME: this copies a lot of code from the surveyor, try to combine the two with shared functions/code
func (game *Game) selectLocationForSpell(yield coroutine.YieldFunc, spell spellbook.Spell, player *playerlib.Player, locationType LocationType) (int, int, bool) {
    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    var cities []*citylib.City
    var citiesMiniMap []maplib.MiniMapCity
    var stacks []*playerlib.UnitStack
    var fog data.FogMap

    for i, player := range game.Players {
        for _, city := range player.Cities {
            if city.Plane == game.Plane {
                cities = append(cities, city)
                citiesMiniMap = append(citiesMiniMap, city)
            }
        }

        for _, stack := range player.Stacks {
            if stack.Plane() == game.Plane {
                stacks = append(stacks, stack)
            }
        }

        if i == 0 {
            fog = player.GetFog(game.Plane)
        }
    }

    fonts := fontslib.MakeSurveyorFonts(game.Cache)
    castingFont := fonts.SurveyorFont
    whiteFont := fonts.WhiteFont

    overworld := Overworld{
        Camera: game.Camera,
        Counter: game.Counter,
        Map: game.CurrentMap(),
        Cities: cities,
        CitiesMiniMap: citiesMiniMap,
        Stacks: stacks,
        SelectedStack: nil,
        ImageCache: &game.ImageCache,
        Fog: fog,
        ShowAnimation: game.State == GameStateUnitMoving,
        FogBlack: game.GetFogImage(),
    }

    cancelBackground, _ := game.ImageCache.GetImage("main.lbx", 47, 0)

    var selectMessage string

    switch locationType {
        case LocationTypeAny, LocationTypeLand, LocationTypeEmptyWater, LocationTypeChangeTerrain,
             LocationTypeTransmute, LocationTypeRaiseVolcano, LocationTypeDisenchant:
            selectMessage = fmt.Sprintf("Select a space as the target for an %v spell.", spell.Name)
        case LocationTypeFriendlyCity:
            selectMessage = fmt.Sprintf("Select a friendly city to cast %v on.", spell.Name)
        case LocationTypeEnemyMeldedNode:
            selectMessage = fmt.Sprintf("Select a magic node as the target for a %v spell.", spell.Name)
        case LocationTypeEnemyCity:
            selectMessage = fmt.Sprintf("Select an enemy city as the target for a %v spell.", spell.Name)
        case LocationTypeFriendlyUnit:
            selectMessage = fmt.Sprintf("Select a friendly unit as the target for a %v spell.", spell.Name)
        default:
            selectMessage = fmt.Sprintf("unhandled location type %v", locationType)
    }

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            mainHud, _ := game.ImageCache.GetImage("main.lbx", 0, 0)
            screen.DrawImage(mainHud, &options)

            landImage, _ := game.ImageCache.GetImage("main.lbx", 57, 0)
            options.GeoM.Translate(float64(240 * data.ScreenScale), float64(77 * data.ScreenScale))
            screen.DrawImage(landImage, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(240 * data.ScreenScale), float64(174 * data.ScreenScale))
            screen.DrawImage(cancelBackground, &options)

            ui.IterateElementsByLayer(func (element *uilib.UIElement){
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })

            game.Fonts.WhiteFont.PrintRight(screen, float64(276 * data.ScreenScale), float64(68 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v GP", game.Players[0].Gold))
            game.Fonts.WhiteFont.PrintRight(screen, float64(313 * data.ScreenScale), float64(68 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, fmt.Sprintf("%v MP", game.Players[0].Mana))

            castingFont.PrintCenter(screen, float64(280 * data.ScreenScale), float64(81 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, "Casting")

            whiteFont.PrintWrapCenter(screen, float64(280 * data.ScreenScale), float64(120 * data.ScreenScale), float64(cancelBackground.Bounds().Dx() - 5 * data.ScreenScale), float64(data.ScreenScale), ebiten.ColorScale{}, selectMessage)
        },
    }

    ui.SetElementsFromArray(nil)

    makeButton := func(lbxIndex int, x int, y int) *uilib.UIElement {
        button, _ := game.ImageCache.GetImage("main.lbx", lbxIndex, 0)
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(x * data.ScreenScale), float64(y * data.ScreenScale))
        return &uilib.UIElement{
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                screen.DrawImage(button, &options)
            },
        }
    }

    // game
    ui.AddElement(makeButton(1, 7, 4))

    // spells
    ui.AddElement(makeButton(2, 47, 4))

    // army button
    ui.AddElement(makeButton(3, 89, 4))

    // cities button
    ui.AddElement(makeButton(4, 140, 4))

    // magic button
    ui.AddElement(makeButton(5, 184, 4))

    // info button
    ui.AddElement(makeButton(6, 226, 4))

    // plane button
    ui.AddElement(makeButton(7, 270, 4))

    quit := false

    // cancel button at bottom
    cancel, _ := game.ImageCache.GetImages("main.lbx", 41)
    cancelIndex := 0
    cancelRect := util.ImageRect(263 * data.ScreenScale, 182 * data.ScreenScale, cancel[0])
    ui.AddElement(&uilib.UIElement{
        Rect: cancelRect,
        LeftClick: func(element *uilib.UIElement){
            cancelIndex = 1
        },
        LeftClickRelease: func(element *uilib.UIElement){
            cancelIndex = 0
            quit = true
        },
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(cancelRect.Min.X), float64(cancelRect.Min.Y))
            screen.DrawImage(cancel[cancelIndex], &options)
        },
    })

    game.Drawer = func(screen *ebiten.Image, game *Game){
        overworld.Camera = game.Camera
        overworld.DrawOverworld(screen, ebiten.GeoM{})

        var miniGeom ebiten.GeoM
        miniGeom.Translate(float64(250 * data.ScreenScale), float64(20 * data.ScreenScale))
        mx, my := miniGeom.Apply(0, 0)
        miniWidth := 60 * data.ScreenScale
        miniHeight := 31 * data.ScreenScale
        mini := screen.SubImage(image.Rect(int(mx), int(my), int(mx) + miniWidth, int(my) + miniHeight)).(*ebiten.Image)
        overworld.DrawMinimap(mini)

        ui.Draw(ui, screen)
    }

    entityInfo := game.ComputeCityStackInfo()

    for !quit {
        if game.Camera.GetZoom() > 0.9 {
            overworld.Counter += 1
        }

        zoomed := game.doInputZoom(yield)
        _ = zoomed
        ui.StandardUpdate()

        x, y := inputmanager.MousePosition()

        // within the viewable area
        if game.InOverworldArea(x, y) {
            tileX, tileY := game.ScreenToTile(float64(x), float64(y))

            // right click should move the camera
            rightClick := inputmanager.RightClick()
            if rightClick /*|| zoomed */ {
                game.doMoveCamera(yield, tileX, tileY)
            }

            if inputmanager.LeftClick() {
                switch locationType {
                    case LocationTypeAny: return tileX, tileY, false
                    case LocationTypeLand:
                        if tileY >= 0 && tileY < overworld.Map.Map.Rows() {
                            tileX = overworld.Map.WrapX(tileX)
                            if player.IsTileExplored(tileX, tileY, game.Plane) {
                                if overworld.Map.GetTile(tileX, tileY).Tile.IsLand() {
                                    return tileX, tileY, false
                                }
                            }
                        }
                    case LocationTypeEmptyWater:
                        if tileY >= 0 && tileY < overworld.Map.Map.Rows() {
                            tileX = overworld.Map.WrapX(tileX)
                            if player.IsTileExplored(tileX, tileY, game.Plane) {
                                if overworld.Map.GetTile(tileX, tileY).Tile.IsWater() {
                                    empty := true
                                    for _, enemy := range game.GetEnemies(player) {
                                        if enemy.FindStack(tileX, tileY, game.Plane) != nil {
                                            empty = false
                                            break
                                        }
                                    }

                                    if empty {
                                        return tileX, tileY, false
                                    }
                                }
                            }
                        }
                    case LocationTypeFriendlyCity:
                        city := player.FindCity(tileX, tileY, game.Plane)
                        if city != nil {
                            return tileX, tileY, false
                        }
                    case LocationTypeChangeTerrain:
                        if tileY >= 0 && tileY < overworld.Map.Map.Rows() {
                            tileX = overworld.Map.WrapX(tileX)

                            if player.IsTileExplored(tileX, tileY, game.Plane) {
                                terrainType := overworld.Map.GetTile(tileX, tileY).Tile.TerrainType()
                                switch terrainType {
                                    case terrain.Desert, terrain.Forest, terrain.Hill,
                                         terrain.Swamp, terrain.Grass, terrain.Volcano,
                                         terrain.Mountain:
                                        return tileX, tileY, false
                                }
                            }
                        }
                    case LocationTypeTransmute:
                        if tileY >= 0 && tileY < overworld.Map.Map.Rows() {
                            tileX = overworld.Map.WrapX(tileX)

                            if player.IsTileExplored(tileX, tileY, game.Plane) {
                                bonusType := overworld.Map.GetBonusTile(tileX, tileY)
                                switch bonusType {
                                    case data.BonusCoal, data.BonusGem, data.BonusIronOre,
                                         data.BonusGoldOre, data.BonusSilverOre, data.BonusMithrilOre:
                                        return tileX, tileY, false
                                }
                            }
                        }
                    case LocationTypeRaiseVolcano:
                        if tileY >= 0 && tileY < overworld.Map.Map.Rows() {
                            tileX = overworld.Map.WrapX(tileX)

                            if player.IsTileExplored(tileX, tileY, game.Plane) {
                                terrainType := overworld.Map.GetTile(tileX, tileY).Tile.TerrainType()
                                switch terrainType {
                                    case terrain.Desert, terrain.Forest, terrain.Swamp, terrain.Grass, terrain.Tundra:
                                        return tileX, tileY, false
                                }
                            }
                        }
                    case LocationTypeEnemyMeldedNode:
                        if tileY >= 0 && tileY < overworld.Map.Map.Rows() {
                            tileX = overworld.Map.WrapX(tileX)

                            if player.IsTileExplored(tileX, tileY, game.Plane) {
                                node := overworld.Map.GetMagicNode(tileX, tileY)
                                if node != nil && node.MeldingWizard != player {
                                    return tileX, tileY, false
                                }
                            }
                        }

                    case LocationTypeDisenchant:
                        // return if the tile has a stack, town, or is a magic node

                        if entityInfo.FindStack(tileX, tileY, game.Plane) != nil ||
                           entityInfo.FindCity(tileX, tileY, game.Plane) != nil ||
                           overworld.Map.GetMagicNode(tileX, tileY) != nil {

                            return tileX, tileY, false
                        }

                    case LocationTypeEnemyCity:
                        for _, enemy := range game.GetEnemies(player) {
                            city := enemy.FindCity(tileX, tileY, game.Plane)
                            if city != nil {
                                return tileX, tileY, false
                            }
                        }

                    case LocationTypeFriendlyUnit:
                        stack := player.FindStack(tileX, tileY, game.Plane)
                        if stack != nil {
                            return tileX, tileY, false
                        }

                    case LocationTypeEnemyUnit:
                        // TODO
                        // FIXME: This should consider only tiles with FogTypeVisible
                }
            }
        }

        if yield() != nil {
            break
        }
    }

    return 0, 0, true
}

func (game *Game) doCastCityEnchantmentFull(spell spellbook.Spell, player *playerlib.Player, locationType LocationType, enchantment data.CityEnchantment, before CityCallback, after CityCallback) {
    var selected func (yield coroutine.YieldFunc, tileX int, tileY int)
    selected = func (yield coroutine.YieldFunc, tileX int, tileY int) {
        // FIXME: Show this only for enemies if detect magic is active and the city is known to the human player
        game.doMoveCamera(yield, tileX, tileY)
        chosenCity, _ := game.FindCity(tileX, tileY, game.Plane)
        if chosenCity == nil {
            return
        }

        if chosenCity.HasEnchantment(enchantment) {
            game.Events <- &GameEventNotice{Message: fmt.Sprintf("This city already has a %v cast on it", spell.Name)}
            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: locationType, SelectedFunc: selected}
            return
        }

        if !before(chosenCity) {
            return
        }

        chosenCity.AddEnchantment(enchantment, player.GetBanner())
        chosenCity.UpdateUnrest()

        yield()

        sound, err := audio.LoadSound(game.Cache, spell.Sound)
        if err == nil {
            sound.Play()
        }

        enchantmentBuilding, ok := buildinglib.EnchantmentBuildings()[enchantment]
        if !ok {
            enchantmentBuilding = buildinglib.BuildingNone
        }

        game.showCastNewBuilding(yield, chosenCity, player, enchantmentBuilding, enchantment.Name())
        game.RefreshUI()

        after(chosenCity)
    }

    game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: locationType, SelectedFunc: selected}
}

func (game *Game) doCastNewCityBuilding(spell spellbook.Spell, player *playerlib.Player, locationType LocationType, newBuilding building.Building, errorMessage string, after CityCallback) {
    var selected func (yield coroutine.YieldFunc, tileX int, tileY int)
    selected = func (yield coroutine.YieldFunc, tileX int, tileY int) {
        // FIXME: Show this only for enemies if detect magic is active and the city is known to the human player
        game.doMoveCamera(yield, tileX, tileY)
        chosenCity, _ := game.FindCity(tileX, tileY, game.Plane)
        if chosenCity == nil {
            return
        }

        if chosenCity.Buildings.Contains(newBuilding) {
            game.Events <- &GameEventNotice{Message: errorMessage}
            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: locationType, SelectedFunc: selected}
            return
        }

        chosenCity.AddBuilding(newBuilding)
        chosenCity.UpdateUnrest()

        yield()

        sound, err := audio.LoadSound(game.Cache, spell.Sound)
        if err == nil {
            sound.Play()
        }

        game.showCastNewBuilding(yield, chosenCity, player, newBuilding, spell.Name)
        game.RefreshUI()

        after(chosenCity)
    }

    game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: locationType, SelectedFunc: selected}
}

func (game *Game) doCastCityEnchantment(spell spellbook.Spell, player *playerlib.Player, locationType LocationType, enchantment data.CityEnchantment) {
    game.doCastCityEnchantmentFull(spell, player, locationType, enchantment, noCityCallback, noCityCallback)
}

type UpdateMapFunction func (tileX int, tileY int, animationFrame int)

func (game *Game) doCastOnMap(yield coroutine.YieldFunc, tileX int, tileY int, animationIndex int, newSound bool, soundIndex int, update UpdateMapFunction) {
    game.Camera.Zoom = camera.ZoomDefault
    game.doMoveCamera(yield, tileX, tileY)

    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    pics, _ := game.ImageCache.GetImages("specfx.lbx", animationIndex)

    animation := util.MakeAnimation(pics, false)

    x, y := game.TileToScreen(tileX, tileY)

    game.Drawer = func(screen *ebiten.Image, game *Game) {
        oldDrawer(screen, game)

        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(x - animation.Frame().Bounds().Dx() / 2), float64(y - animation.Frame().Bounds().Dy() / 2))
        screen.DrawImage(animation.Frame(), &options)
    }

    if newSound {
        sound, err := audio.LoadNewSound(game.Cache, soundIndex)
        if err == nil {
            sound.Play()
        }
    } else {
        sound, err := audio.LoadSound(game.Cache, soundIndex)
        if err == nil {
            sound.Play()
        }
    }

    quit := false
    for !quit {
        game.Counter += 1

        quit = false
        if game.Counter % 6 == 0 {
            update(tileX, tileY, animation.CurrentFrame)
            quit = !animation.Next()
        }

        if yield() != nil {
            return
        }
    }
}

func (game *Game) doCastEnchantRoad(yield coroutine.YieldFunc, tileX int, tileY int) {
    update := func (x int, y int, frame int) {}

    game.doCastOnMap(yield, tileX, tileY, 46, false, 86, update)

    useMap := game.CurrentMap()

    // all roads in a 5x5 square around the target tile should become enchanted
    for dx := -2; dx <= 2; dx++ {
        for dy := -2; dy <= 2; dy++ {
            cx := useMap.WrapX(tileX + dx)
            cy := tileY + dy
            if cy < 0 || cy >= useMap.Height() {
                continue
            }

            if useMap.ContainsRoad(cx, cy) {
                useMap.SetRoad(cx, cy, true)
            }
        }
    }
}

func (game *Game) doCastEarthLore(yield coroutine.YieldFunc, tileX int, tileY int, player *playerlib.Player) {
    update := func (x int, y int, frame int) {}

    game.doCastOnMap(yield, tileX, tileY, 45, true, 18, update)

    player.LiftFogSquare(tileX, tileY, 5, game.Plane)
}

func (game *Game) doCastChangeTerrain(yield coroutine.YieldFunc, tileX int, tileY int) {
    update := func (x int, y int, frame int) {
        if frame == 7 {
            mapObject := game.CurrentMap()
            switch mapObject.GetTile(x, y).Tile.TerrainType() {
                case terrain.Desert, terrain.Forest, terrain.Hill, terrain.Swamp:
                    mapObject.Map.SetTerrainAt(x, y, terrain.Grass, mapObject.Data, mapObject.Plane)
                case terrain.Grass:
                    mapObject.Map.SetTerrainAt(x, y, terrain.Forest, mapObject.Data, mapObject.Plane)
                case terrain.Volcano:
                    mapObject.RemoveVolcano(x, y)
                case terrain.Mountain:
                    mapObject.Map.SetTerrainAt(x, y, terrain.Hill, mapObject.Data, mapObject.Plane)
            }
        }
    }

    game.doCastOnMap(yield, tileX, tileY, 8, false, 28, update)
    game.RefreshUI()
}

func (game *Game) doCastTransmute(yield coroutine.YieldFunc, tileX int, tileY int) {
    update := func (x int, y int, frame int) {
        if frame == 6 {
            mapObject := game.CurrentMap()
            switch mapObject.GetBonusTile(x, y) {
                case data.BonusCoal: mapObject.SetBonus(x, y, data.BonusGem)
                case data.BonusGem: mapObject.SetBonus(x, y, data.BonusCoal)
                case data.BonusIronOre: mapObject.SetBonus(x, y, data.BonusGoldOre)
                case data.BonusGoldOre: mapObject.SetBonus(x, y, data.BonusIronOre)
                case data.BonusSilverOre: mapObject.SetBonus(x, y, data.BonusMithrilOre)
                case data.BonusMithrilOre: mapObject.SetBonus(x, y, data.BonusSilverOre)
            }
        }
    }

    game.doCastOnMap(yield, tileX, tileY, 0, false, 28, update)
    game.RefreshUI()
}

func (game *Game) doCastRaiseVolcano(yield coroutine.YieldFunc, tileX int, tileY int, player *playerlib.Player) {
    update := func (x int, y int, frame int) {
        if frame == 8 {
            mapObject := game.CurrentMap()
            mapObject.Map.SetTerrainAt(x, y, terrain.Grass, mapObject.Data, mapObject.Plane)
            mapObject.SetBonus(tileX, tileY, data.BonusNone)
        }
    }

    game.doCastOnMap(yield, tileX, tileY, 11, false, 98, update)

    mapObject := game.CurrentMap()
    mapObject.SetVolcano(tileX, tileY, player)

    // volcanoes may destroy buildings if cast in a city
    for _, player := range game.Players {
        city := player.FindCity(tileX, tileY, mapObject.Plane)
        if city != nil {
            for _, building := range city.Buildings.Values() {
                if rand.N(100) < 15 {
                    city.Buildings.Remove(building)
                }
            }
        }
    }

    game.RefreshUI()
}

func (game *Game) doCastCorruption(yield coroutine.YieldFunc, tileX int, tileY int) {
    update := func (x int, y int, frame int) {
        if frame == 6 {
            mapObject := game.CurrentMap()
            if y >= 0 || y < mapObject.Map.Rows() {
                x = mapObject.WrapX(x)
                mapObject.SetCorruption(x, y)
            }
        }
    }

    game.doCastOnMap(yield, tileX, tileY, 7, false, 103, update)
    game.RefreshUI()
}

func (game *Game) doCastSpellBlast(player *playerlib.Player) {
    var wizSelectionUiGroup *uilib.UIElementGroup
    onTargetSelectCallback := func(targetPlayer *playerlib.Player) bool {
        if targetPlayer.Defeated || targetPlayer.Banished || !targetPlayer.CastingSpell.Valid() || targetPlayer.CastingSpellProgress > player.Mana {
            return false
        }
        player.Mana -= targetPlayer.CastingSpellProgress
        targetPlayer.InterruptCastingSpell()
        return true
    }
    playersInGame := len(game.Players)
    wizSelectionUiGroup = makeSelectSpellBlastTargetUI(game.HudUI, game.Cache, &game.ImageCache, player, playersInGame, onTargetSelectCallback)
    game.HudUI.AddGroup(wizSelectionUiGroup)
}

func (game *Game) doCastWarpNode(yield coroutine.YieldFunc, tileX int, tileY int, caster *playerlib.Player) {
    update := func (x int, y int, frame int) {}

    game.doCastOnMap(yield, tileX, tileY, 13, true, 5, update)

    node := game.CurrentMap().GetMagicNode(tileX, tileY)
    if node != nil {
        node.Warped = true
        node.WarpedOwner = caster
    }
}

func (game *Game) doCastGlobalEnchantment(yield coroutine.YieldFunc, player *playerlib.Player, enchantment data.Enchantment) {

    song := music.SongNone

    switch enchantment {
        case data.EnchantmentAwareness: song = music.SongLearnSpell
        case data.EnchantmentDetectMagic: song = music.SongLearnSpell
        case data.EnchantmentCharmOfLife: song = music.SongLearnSpell
        case data.EnchantmentCrusade: song = music.SongCrusade
        case data.EnchantmentHolyArms: song = music.SongHolyArms
        case data.EnchantmentJustCause: song = music.SongJustCause
        case data.EnchantmentLifeForce: song = music.SongLifeForce
        case data.EnchantmentPlanarSeal: song = music.SongUncommonSummoningSpell
        case data.EnchantmentTranquility: song = music.SongTranquility
        case data.EnchantmentHerbMastery: song = music.SongHerbMastery
        case data.EnchantmentNatureAwareness: song = music.SongNatureAwareness
        case data.EnchantmentNaturesWrath: song = music.SongNaturesWrath
        case data.EnchantmentAuraOfMajesty: song = music.SongAuraOfMajesty
        case data.EnchantmentSuppressMagic: song = music.SongSuppressMagic
        case data.EnchantmentTimeStop: song = music.SongTimeStop
        case data.EnchantmentWindMastery: song = music.SongWindMastery
        case data.EnchantmentArmageddon: song = music.SongArmageddon
        case data.EnchantmentChaosSurge: song = music.SongChaosSurge
        case data.EnchantmentDoomMastery: song = music.SongDoomMastery
        case data.EnchantmentGreatWasting: song = music.SongGreatWasting
        case data.EnchantmentMeteorStorm: song = music.SongMeteorStorm
        case data.EnchantmentEternalNight: song = music.SongEternalNight
        case data.EnchantmentEvilOmens: song = music.SongEvilOmens
        case data.EnchantmentZombieMastery: song = music.SongZombieMastery
    }

    if song != music.SongNone {
        game.Music.PushSong(song)
        defer game.Music.PopSong()
    }

    fonts := fontslib.MakeGlobalEnchantmentFonts(game.Cache)

    diplomacLbx, _ := game.Cache.GetLbxFile("diplomac.lbx")
    // the tauron fade in, but any mask will work
    maskSprites, _ := diplomacLbx.ReadImages(46)
    mask := maskSprites[0]

    // stolen from diplomacy.go
    var makeCutoutMask util.ImageTransformFunc = func (img *image.Paletted) image.Image {
        properImage := img.SubImage(mask.Bounds()).(*image.Paletted)
        imageOut := image.NewPaletted(properImage.Bounds(), properImage.Palette)

        for x := properImage.Bounds().Min.X; x < properImage.Bounds().Max.X; x++ {
            for y := properImage.Bounds().Min.Y; y < properImage.Bounds().Max.Y; y++ {
                maskColor := mask.At(x, y)
                _, _, _, a := maskColor.RGBA()
                if a > 0 {
                    imageOut.Set(x, y, properImage.At(x, y))
                } else {
                    imageOut.SetColorIndex(x, y, 0)
                }
            }
        }

        return imageOut
    }

    frame, _ := game.ImageCache.GetImage("backgrnd.lbx", 18, 0)

    oldDrawer := game.Drawer
    defer func(){
        game.Drawer = oldDrawer
    }()

    animationIndex := 0
    switch player.Wizard.Base {
        case data.WizardMerlin: animationIndex = 0
        case data.WizardRaven: animationIndex = 1
        case data.WizardSharee: animationIndex = 2
        case data.WizardLoPan: animationIndex = 3
        case data.WizardJafar: animationIndex = 4
        case data.WizardOberic: animationIndex = 5
        case data.WizardRjak: animationIndex = 6
        case data.WizardSssra: animationIndex = 7
        case data.WizardTauron: animationIndex = 8
        case data.WizardFreya: animationIndex = 9
        case data.WizardHorus: animationIndex = 10
        case data.WizardAriel: animationIndex = 11
        case data.WizardTlaloc: animationIndex = 12
        case data.WizardKali: animationIndex = 13
    }

    spellImage, _ := game.ImageCache.GetImage("specfx.lbx", enchantment.LbxIndex(), 0)

    doDraw := 0

    fadeSpeed := 7

    fader := util.MakeFadeIn(uint64(fadeSpeed), &game.Counter)

    offset := -35 * data.ScreenScale

    game.Drawer = func(screen *ebiten.Image, game *Game){
        oldDrawer(screen, game)
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(data.ScreenWidth / 2), float64(data.ScreenHeight / 2))
        options.GeoM.Translate(float64(offset), 0)
        options.GeoM.Translate(float64(-frame.Bounds().Dx() / 2), float64(-frame.Bounds().Dy() / 2))
        screen.DrawImage(frame, &options)

        options.ColorScale.ScaleAlpha(fader())

        // first draw the wizard
        if doDraw == 0 {
            mood, _ := game.ImageCache.GetImageTransform("moodwiz.lbx", animationIndex, 2, "cutout", makeCutoutMask)
            options.GeoM.Translate(float64(13 * data.ScreenScale), float64(13 * data.ScreenScale))
            screen.DrawImage(mood, &options)

            text := "You have finished casting"
            if !player.IsHuman() {
                text = fmt.Sprintf("%v has cast", player.Wizard.Name)
            }

            fonts.InfoFont.PrintCenter(screen, float64(data.ScreenWidth / 2 + offset), float64(data.ScreenHeight / 2 + frame.Bounds().Dy() / 2), float64(data.ScreenScale), options.ColorScale, text)
        } else {
            // then draw the spell image
            options.GeoM.Translate(float64(9 * data.ScreenScale), float64(8 * data.ScreenScale))
            screen.DrawImage(spellImage, &options)

            fonts.InfoFont.PrintCenter(screen, float64(data.ScreenWidth / 2 + offset), float64(data.ScreenHeight / 2 + frame.Bounds().Dy() / 2), float64(data.ScreenScale), options.ColorScale, enchantment.String())
        }

    }

    delay := uint64(3 * 60)

    deadline := game.Counter + delay

    quit := false
    for !quit && game.Counter < deadline {
        game.Counter += 1
        leftClick := inputmanager.LeftClick()
        if leftClick {
            quit = true
        }

        if yield() != nil {
            return
        }
    }

    fader = util.MakeFadeOut(uint64(fadeSpeed), &game.Counter)

    for range fadeSpeed {
        game.Counter += 1
        yield()
    }

    yield()

    doDraw = 1

    fader = util.MakeFadeIn(uint64(fadeSpeed), &game.Counter)

    deadline = game.Counter + delay

    quit = false
    for !quit && game.Counter < deadline {
        game.Counter += 1
        leftClick := inputmanager.LeftClick()
        if leftClick {
            quit = true
        }
        if yield() != nil {
            return
        }
    }

    fader = util.MakeFadeOut(uint64(fadeSpeed), &game.Counter)
    for range fadeSpeed {
        game.Counter += 1
        yield()
    }

    yield()

}

func (game *Game) doCastFloatingIsland(yield coroutine.YieldFunc, player *playerlib.Player, tileX int, tileY int) {
    update := func (x int, y int, frame int) {
        if frame == 5 {
            overworldUnit := units.MakeOverworldUnitFromUnit(units.FloatingIsland, tileX, tileY, game.CurrentMap().Plane, player.Wizard.Banner, player.MakeExperienceInfo())
            player.AddUnit(overworldUnit)
            player.LiftFog(tileX, tileY, 1, game.Plane)
        }
    }

    game.doCastOnMap(yield, tileX, tileY, 1, false, 29, update)
}
