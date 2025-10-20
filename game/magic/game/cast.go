package game

import (
    "fmt"
    "log"
    "image"
    "context"
    "slices"
    "cmp"
    "math/rand/v2"
    "errors"

    "github.com/kazzmir/master-of-magic/lib/coroutine"
    "github.com/kazzmir/master-of-magic/lib/set"
    "github.com/kazzmir/master-of-magic/lib/font"
    "github.com/kazzmir/master-of-magic/lib/lbx"
    playerlib "github.com/kazzmir/master-of-magic/game/magic/player"
    herolib "github.com/kazzmir/master-of-magic/game/magic/hero"
    fontslib "github.com/kazzmir/master-of-magic/game/magic/fonts"
    "github.com/kazzmir/master-of-magic/game/magic/music"
    "github.com/kazzmir/master-of-magic/game/magic/units"
    "github.com/kazzmir/master-of-magic/game/magic/maplib"
    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/scale"
    citylib "github.com/kazzmir/master-of-magic/game/magic/city"
    "github.com/kazzmir/master-of-magic/game/magic/cityview"
    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    buildinglib "github.com/kazzmir/master-of-magic/game/magic/building"
    "github.com/kazzmir/master-of-magic/game/magic/combat"
    "github.com/kazzmir/master-of-magic/game/magic/mirror"
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

type SpellSpecialUIFonts struct {
    BigOrange *font.Font
    InfoOrange *font.Font
}

func MakeSpellSpecialUIFonts(cache *lbx.LbxCache) *SpellSpecialUIFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Error loading fonts: %v", err)
        return nil
    }

    return &SpellSpecialUIFonts{
        BigOrange: loader(fontslib.LightGradient1),
        InfoOrange: loader(fontslib.SmallYellow),
    }
}

// the reason a spell was fizzled due to some global enchantment
type FizzleReason struct {
    // the owner of the enchantment
    Owner *playerlib.Player
    // the enchantment that caused the spell to fizzle
    Enchantment data.Enchantment
}

func (game *Game) doCastSpell(player *playerlib.Player, spell spellbook.Spell) {
    // FIXME: if the player is AI then invoke some callback that the AI will use to select targets instead of using the GameEventSelectLocationForSpell

    fizzled, reason := game.checkInstantFizzleForCastSpell(player, spell)
    if fizzled {
        // Fizzle the spell and return
        game.ShowFizzleSpell(spell, player)

        if reason.Owner == game.GetHumanPlayer() || player == game.GetHumanPlayer() {
            game.ShowTranquilityFizzle(reason.Owner, player, spell)
        }

        return
    }

    if spell.IsOfRealm(data.ChaosMagic) || spell.IsOfRealm(data.DeathMagic) {
        game.maybeDoNaturesWrath(player)
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

            choice := choices[rand.N(len(choices))]

            before := func (unit units.StackUnit) bool {
                if unit.GetRace() == data.RaceFantastic {
                    game.Events <- &GameEventNotice{Message: "That unit cannot be targeted"}
                    return false
                }

                for _, enchantment := range choices {
                    if unit.HasEnchantment(enchantment) {
                        game.Events <- &GameEventNotice{Message: "That unit cannot be targeted"}
                        return false
                    }
                }

                return true
            }

            after := func (unit units.StackUnit) bool {
                unit.AddEnchantment(choice)
                return true
            }

            game.doCastOnUnit(player, spell, choice.CastAnimationIndex(), before, after)
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
                overworldUnit, ok := unit.(*units.OverworldUnit)
                if ok {
                    overworldUnit.Unit = units.WereWolf
                    overworldUnit.Experience = 0
                    // unit keeps weapon bonus and enchantments
                }
                return true
            }
            game.doCastOnUnit(player, spell, 4, before, after)

        case "Spell Ward":
            game.doCastSpellWard(player, spell)
        case "Flying Fortress":
            game.doCastCityEnchantment(spell, player, LocationTypeFriendlyCity, data.CityEnchantmentFlyingFortress)
        case "Earth Gate":
            game.doCastCityEnchantment(spell, player, LocationTypeFriendlyCity, data.CityEnchantmentEarthGate)
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
        */
        case "Chaos Rift":
            before := func (city *citylib.City) bool {
                if city.HasAnyOfEnchantments(data.CityEnchantmentConsecration, data.CityEnchantmentChaosWard) {
                    game.ShowFizzleSpell(spell, player)
                    return false
                }
                return true
            }
            game.doCastCityEnchantmentFull(spell, player, LocationTypeEnemyCity, data.CityEnchantmentChaosRift, before, noCityCallback)
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
        */
        case "Aura of Majesty":
            game.castGlobalEnchantment(data.EnchantmentAuraOfMajesty, player)
        case "Time Stop":
            game.castGlobalEnchantment(data.EnchantmentTimeStop, player)
        case "Zombie Mastery":
            game.castGlobalEnchantment(data.EnchantmentZombieMastery, player)
        case "Evil Omens":
            game.castGlobalEnchantment(data.EnchantmentEvilOmens, player)
        case "Meteor Storm":
            game.castGlobalEnchantment(data.EnchantmentMeteorStorm, player)
        case "Doom Mastery":
            game.castGlobalEnchantment(data.EnchantmentDoomMastery, player)
        case "Chaos Surge":
            game.castGlobalEnchantment(data.EnchantmentChaosSurge, player)
        case "Wind Mastery":
            game.castGlobalEnchantment(data.EnchantmentWindMastery, player)
        case "Suppress Magic":
            game.castGlobalEnchantment(data.EnchantmentSuppressMagic, player)
        case "Nature's Wrath":
            game.castGlobalEnchantment(data.EnchantmentNaturesWrath, player)
        case "Charm of Life":
            game.castGlobalEnchantment(data.EnchantmentCharmOfLife, player)
        case "Holy Arms":
            game.castGlobalEnchantment(data.EnchantmentHolyArms, player)
        case "Planar Seal":
            game.castGlobalEnchantment(data.EnchantmentPlanarSeal, player)
        case "Herb Mastery":
            game.castGlobalEnchantment(data.EnchantmentHerbMastery, player)
        case "Awareness":
            game.castGlobalEnchantment(data.EnchantmentAwareness, player)
        case "Nature Awareness":
            game.castGlobalEnchantment(data.EnchantmentNatureAwareness, player)
        case "Crusade":
            game.castGlobalEnchantment(data.EnchantmentCrusade, player)
        case "Just Cause":
            game.castGlobalEnchantment(data.EnchantmentJustCause, player)
        case "Life Force":
            game.castGlobalEnchantment(data.EnchantmentLifeForce, player)
        case "Tranquility":
            game.castGlobalEnchantment(data.EnchantmentTranquility, player)
        case "Armageddon":
            game.castGlobalEnchantment(data.EnchantmentArmageddon, player)
        case "Great Wasting":
            game.castGlobalEnchantment(data.EnchantmentGreatWasting, player)
        case "Detect Magic":
            game.castGlobalEnchantment(data.EnchantmentDetectMagic, player)
        case "Eternal Night":
            game.castGlobalEnchantment(data.EnchantmentEternalNight, player)

        /*
            INSTANT SPELLS
                TODO:
                Spell of Mastery
        */
        case "Spell of Return":
            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                city := player.FindCity(tileX, tileY, game.Plane)
                if city != nil {
                    // FIXME: verify animation
                    game.doCastOnMap(yield, tileX, tileY, 44, spell.Sound, func (x int, y int, animationFrame int) {})
                    player.Banished = false

                    for _, check := range player.Cities {
                        check.RemoveBuilding(building.BuildingSummoningCircle)
                        check.RemoveBuilding(building.BuildingFortress)
                    }

                    city.Buildings.Insert(buildinglib.BuildingFortress)
                    city.Buildings.Insert(buildinglib.BuildingSummoningCircle)
                    player.Wizard.Race = city.Race
                }
            }

            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeFriendlyCity, SelectedFunc: selected}

        case "Plane Shift":
            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                game.doCastOnMap(yield, tileX, tileY, 3, spell.Sound, func (x int, y int, animationFrame int) {})

                stack := player.FindStack(tileX, tileY, game.Plane)
                if stack != nil {
                    err := game.PlaneShift(stack, player)
                    if err != nil {
                        game.Events <- &GameEventNotice{Message: fmt.Sprintf("%v", err)}
                    } else {
                        game.Plane = stack.Plane()
                    }
                }
            }

            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeFriendlyUnit, SelectedFunc: selected}

        case "Subversion":
            uiGroup, quit, err := game.MakeSubversionUI(player, spell)
            if err != nil {
                game.Events <- &GameEventNotice{Message: fmt.Sprintf("%v", err)}
            } else {
                game.Events <- &GameEventRunUI{Group: uiGroup, Quit: quit}
            }

        case "Death Wish":
            after := func() {
                cityStackInfo := game.ComputeCityStackInfo()

                for _, owner := range game.Players {
                    if owner == player {
                        continue
                    }

                    for _, stack := range owner.Stacks {

                        city := cityStackInfo.FindCity(stack.X(), stack.Y(), stack.Plane())
                        if city != nil && !city.CanTarget(spell) {
                            continue
                        }

                        for _, unit := range stack.Units() {
                            ignore := unit.GetRace() == data.RaceFantastic

                            if unit.HasEnchantment(data.UnitEnchantmentChaosChannelsDemonWings) ||
                               unit.HasEnchantment(data.UnitEnchantmentChaosChannelsDemonSkin) ||
                               unit.HasEnchantment(data.UnitEnchantmentChaosChannelsFireBreath) {
                                   ignore = false
                            }

                            if ignore {
                                continue
                            }

                            resistance := combat.GetResistanceFor(unit, data.DeathMagic)
                            if rand.N(10) + 1 > resistance {
                                owner.RemoveUnit(unit)
                            }
                        }
                    }
                }
            }

            game.Events <- &GameEventCastGlobalEnchantment{Player: player, Enchantment: data.DeathWish, After: after}

        case "Black Wind":
            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                stack, owner := game.FindStack(tileX, tileY, game.Plane)
                if stack != nil {

                    city, _ := game.FindCity(tileX, tileY, game.Plane)
                    if city != nil && !city.CanTarget(spell) {
                        game.ShowFizzleSpell(spell, player)
                        return
                    }

                    game.doCastOnMap(yield, tileX, tileY, 14, spell.Sound, func (x int, y int, animationFrame int) {})

                    for _, unit := range stack.Units() {
                        if rand.N(10) + 1 > combat.GetResistanceFor(unit, data.DeathMagic) - 1 {
                            owner.RemoveUnit(unit)
                        }
                    }

                }
            }

            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeEnemyUnit, SelectedFunc: selected}


        case "Stasis":
            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                stack, _ := game.FindStack(tileX, tileY, game.Plane)
                if stack != nil {

                    city, _ := game.FindCity(tileX, tileY, game.Plane)
                    if city != nil && !city.CanTarget(spell) {
                        game.ShowFizzleSpell(spell, player)
                        return
                    }

                    game.doCastOnMap(yield, tileX, tileY, 53, spell.Sound, func (x int, y int, animationFrame int) {})

                    // FIXME: maybe apply a Stasis unit enchantment so the user can see the unit is under the stasis effect?
                    for _, unit := range stack.Units() {
                        unit.SetBusy(units.BusyStatusStasis)
                    }

                }
            }

            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeEnemyUnit, SelectedFunc: selected}
        case "Spell Binding":
            uiGroup, quit, err := game.MakeSpellBindingUI(player, spell)
            if err != nil {
                game.Events <- &GameEventNotice{Message: fmt.Sprintf("%v", err)}
            } else {
                game.Events <- &GameEventRunUI{Group: uiGroup, Quit: quit}
            }

        case "Great Unsummoning":
            after := func(){
                cityStackInfo := game.ComputeCityStackInfo()

                for _, player := range game.Players {
                    for _, stack := range player.Stacks {

                        city := cityStackInfo.FindCity(stack.X(), stack.Y(), stack.Plane())
                        if city != nil && !city.CanTarget(spell) {
                            continue
                        }

                        for _, unit := range stack.Units() {
                            if unit.GetRace() == data.RaceFantastic {

                                if unit.HasEnchantment(data.UnitEnchantmentSpellLock) ||
                                unit.HasEnchantment(data.UnitEnchantmentChaosChannelsDemonWings) ||
                                unit.HasEnchantment(data.UnitEnchantmentChaosChannelsDemonSkin) ||
                                unit.HasEnchantment(data.UnitEnchantmentChaosChannelsFireBreath) ||
                                unit.HasAbility(data.AbilityMagicImmunity) {
                                    continue
                                }

                                resistance := combat.GetResistanceFor(unit, data.SorceryMagic)
                                if rand.N(10) + 1 > resistance - 3 {
                                    player.RemoveUnit(unit)
                                }
                            }
                        }
                    }
                }

                game.RefreshUI()
            }

            game.Events <- &GameEventCastGlobalEnchantment{Player: player, Enchantment: data.GreatUnsummoning, After: after}

        case "Nature's Cures":
            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                stack := player.FindStack(tileX, tileY, game.Plane)
                if stack != nil {
                    // heal all units that aren't undead or death fantastic
                    stack.NaturalHeal(1)
                    game.doCastOnMap(yield, tileX, tileY, 0, spell.Sound, func (x int, y int, animationFrame int) {})
                }
            }

            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeFriendlyUnit, SelectedFunc: selected}

        case "Resurrection":
            heroes := player.GetDeadHeroes()
            if len(heroes) == 0 {
                game.Events <- &GameEventNotice{Message: "No dead heroes to resurrect"}
            } else if player.FreeHeroSlots() == 0 {
                game.Events <- &GameEventNotice{Message: "No free hero slots to resurrect hero"}
            } else {
                // show selection box for all dead heroes

                group, quit := game.MakeResurrectionUI(player, heroes, spell.Sound)

                game.Events <- &GameEventRunUI{
                    Group: group,
                    Quit: quit,
                }
            }

        case "Earthquake":
            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                city, owner := game.FindCity(tileX, tileY, game.Plane)
                if city != nil {
                    sound, err := audio.LoadSound(game.Cache, spell.Sound)
                    if err == nil {
                        sound.Play()
                    }

                    game.showCityEarthquake(yield, city, owner)
                }

                yield()
            }

            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeEnemyCity, SelectedFunc: selected}

        case "Ice Storm":
            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                enemyStack, enemy := game.FindStack(tileX, tileY, game.Plane)

                game.doCastOnMap(yield, tileX, tileY, 10, spell.Sound, func (x int, y int, animationFrame int) {})

                for _, unit := range enemyStack.Units() {
                    combat.ApplyAreaDamage(&UnitDamageWrapper{StackUnit: unit}, 6, units.DamageCold, 0)
                    if unit.GetHealth() <= 0 {
                        enemy.RemoveUnit(unit)
                    }
                }

                yield()
            }

            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeEnemyUnit, SelectedFunc: selected}
        case "Fire Storm":
            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                enemyStack, enemy := game.FindStack(tileX, tileY, game.Plane)

                game.doCastOnMap(yield, tileX, tileY, 6, spell.Sound, func (x int, y int, animationFrame int) {})

                for _, unit := range enemyStack.Units() {
                    combat.ApplyAreaDamage(&UnitDamageWrapper{StackUnit: unit}, 8, units.DamageImmolation, 0)
                    if unit.GetHealth() <= 0 {
                        enemy.RemoveUnit(unit)
                    }
                }

                yield()
            }

            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeEnemyUnit, SelectedFunc: selected}

        case "Disjunction", "Disjunction True":
            uiGroup, quit, err := game.MakeDisjunctionUI(player, spell)
            if err != nil {
                game.Events <- &GameEventNotice{Message: fmt.Sprintf("%v", err)}
            } else {
                game.Events <- &GameEventRunUI{Group: uiGroup, Quit: quit}
            }

        case "Create Artifact", "Enchant Item":
            game.Events <- &GameEventSummonArtifact{Player: player}
            game.Events <- &GameEventVault{CreatedArtifact: player.CreateArtifact, Player: player}
            player.CreateArtifact = nil
        case "Earth Lore":
            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                game.doCastEarthLore(yield, tileX, tileY, player, spell.Sound)
            }

            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeAny, SelectedFunc: selected}
        case "Call the Void":
            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                chosenCity, owner := game.FindCity(tileX, tileY, game.Plane)

                if chosenCity.HasAnyOfEnchantments(data.CityEnchantmentConsecration, data.CityEnchantmentChaosWard) {
                    game.ShowFizzleSpell(spell, player)
                    return
                }

                // FIXME: verify the animation and sound. The spell index is 102
                game.doCastOnMap(yield, tileX, tileY, 12, 72, func (x int, y int, animationFrame int) {})
                game.doCallTheVoid(chosenCity, owner)
            }

            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeEnemyCity, SelectedFunc: selected}
        case "Change Terrain":
            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeChangeTerrain, SelectedFunc: game.doCastChangeTerrain}
        case "Cruel Unminding":
            game.doCastCruelUnminding(player, spell)
        case "Drain Power":
            game.doCastDrainPower(player, spell)
        case "Transmute":
            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeTransmute, SelectedFunc: game.doCastTransmute}
        case "Raise Volcano":
            selected := func (yield coroutine.YieldFunc, tileX int, tileY int){
                chosenCity, _ := game.FindCity(tileX, tileY, game.Plane)

                // unclear if chaos ward makes the spell fizzle or if this tile just can't be selected
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
                game.doCastWarpNode(yield, tileX, tileY, player, spell.Sound)
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
        case "Word of Recall":
            before := func (unit units.StackUnit) bool {
                summonCity := player.FindSummoningCity()
                if summonCity == nil {
                    return false
                }

                if unit.GetX() == summonCity.X && unit.GetY() == summonCity.Y {
                    game.Events <- &GameEventNotice{Message: "Your unit is already in this city"}
                    return false
                }

                if unit.GetPlane() != summonCity.Plane && game.IsGlobalEnchantmentActive(data.EnchantmentPlanarSeal) {
                    game.Events <- &GameEventNotice{Message: "Your unit cannot planar travel to this location"}
                    return false
                }

               return true
            }
            after := func (unit units.StackUnit) bool {
                game.RelocateUnit(player, unit)

                return true
            }
            game.doCastOnUnit(player, spell, 1, before, after)

        default:
            log.Printf("Warning: casting unhandled spell '%v'", spell.Name)
    }
}

func (game *Game) castGlobalEnchantment(enchantment data.Enchantment, player *playerlib.Player) {
    if !player.GlobalEnchantments.Contains(enchantment) {
        game.Events <- &GameEventCastGlobalEnchantment{Player: player, Enchantment: enchantment, After: func(){
            game.ApplyGlobalEnchantment(enchantment, player)
        }}
        player.GlobalEnchantments.Insert(enchantment)
        game.RefreshUI()
    }
}

// apply effects for a global enchantment
func (game *Game) ApplyGlobalEnchantment(enchantment data.Enchantment, player *playerlib.Player) {
    switch enchantment {
        case data.EnchantmentAwareness: game.doExploreFogForAwareness(player)
        case data.EnchantmentNatureAwareness:
            player.LiftFogAll(data.PlaneArcanus)
            player.LiftFogAll(data.PlaneMyrror)
        case data.EnchantmentGreatWasting:
            for _, player := range game.Players {
                player.UpdateUnrest()
            }
        case data.EnchantmentArmageddon:
            for _, player := range game.Players {
                player.UpdateUnrest()
            }
        case data.EnchantmentJustCause:
            player.UpdateUnrest()
        case data.EnchantmentHolyArms:
            // all units implicitly have holy weapon, so the actual enchantment is removed from all units
            for _, unit := range player.Units {
                unit.RemoveEnchantment(data.UnitEnchantmentHolyWeapon)
            }
        case data.EnchantmentAuraOfMajesty:
            for _, other := range player.GetKnownPlayers() {
                other.AdjustDiplomaticRelation(player, 10)
                player.AdjustDiplomaticRelation(other, 10)
            }

    }
}

func (game *Game) MakeResurrectionUI(caster *playerlib.Player, heroes []*herolib.Hero, resurrectionSound int) (*uilib.UIElementGroup, context.Context) {
    group := uilib.MakeGroup()

    quit, cancel := context.WithCancel(context.Background())

    var layer uilib.UILayer = 1

    uiX := 40
    uiY := 1

    specialFonts := MakeSpellSpecialUIFonts(game.Cache)

    var selectedHero *herolib.Hero
    selectedHero = heroes[0]

    // background
    group.AddElement(&uilib.UIElement{
        Layer: layer,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := game.ImageCache.GetImage("spellscr.lbx", 45, 0)
            var options ebiten.DrawImageOptions
            options.GeoM.Translate(float64(uiX), float64(uiY))
            scale.DrawScaled(screen, background, &options)

            specialFonts.BigOrange.PrintOptions(screen, float64(uiX + background.Bounds().Dx() / 2), float64(uiY + 10), font.FontOptions{Justify: font.FontJustifyCenter, Scale: scale.ScaleAmount, DropShadow: true, Options: &options}, "Select hero to resurrect")

            if selectedHero != nil {
                specialFonts.BigOrange.PrintOptions(screen, float64(uiX + background.Bounds().Dx() / 2), float64(uiY + background.Bounds().Dy() - 20), font.FontOptions{Justify: font.FontJustifyCenter, Scale: scale.ScaleAmount, DropShadow: true, Options: &options}, selectedHero.GetFullName())
            }
        },
        Hack: func(element *uilib.UIElement) {
            cancel()
        },
        /*
        NotLeftClicked: func(element *uilib.UIElement) {
            cancel()
        },
        */
    })

    gridX := 0
    gridY := 0
    for _, hero := range slices.SortedFunc(slices.Values(heroes), func (hero1 *herolib.Hero, hero2 *herolib.Hero) int {
        return cmp.Compare(hero1.Name, hero2.Name)
    }) {
        lbxFile, lbxIndex := hero.GetPortraitLbxInfo()
        pic, _ := game.ImageCache.GetImage(lbxFile, lbxIndex, 0)
        if pic == nil {
            log.Printf("Error with hero picture: %v", hero.Name)
            continue
        }

        gridWidth := 18
        gridHeight := 18
        gapX := 7
        gapY := 6

        xPos := uiX + 13 + gridX * (gridWidth + gapX)
        yPos := uiY + 26 + gridY * (gridHeight + gapY)

        rect := image.Rect(xPos, yPos, xPos + gridWidth, yPos + gridHeight)
        lookTime := 0
        maxLookTime := 20
        group.AddElement(&uilib.UIElement{
            Layer: layer,
            Order: 1,
            Rect: rect,
            Inside: func(element *uilib.UIElement, x int, y int) {
                selectedHero = hero
                lookTime = min(lookTime + 1, maxLookTime)
            },
            NotInside: func(element *uilib.UIElement) {
                if selectedHero == hero {
                    selectedHero = nil
                }

                lookTime = max(lookTime - 1, 0)
            },
            LeftClick: func(element *uilib.UIElement) {
                cancel()
                hero.SetStatus(herolib.StatusEmployed)
                caster.AddHeroToSummoningCircle(hero)
                game.ResolveStackAt(hero.GetX(), hero.GetY(), hero.GetPlane())

                summoningCity := caster.FindSummoningCity()
                if summoningCity != nil {
                    game.Plane = summoningCity.Plane
                    game.Events <- &GameEventInvokeRoutine{
                        Routine: func (yield coroutine.YieldFunc) {
                            game.doCastOnMap(yield, summoningCity.X, summoningCity.Y, 3, resurrectionSound, func (x int, y int, animationFrame int) {})
                            game.RefreshUI()
                        },
                    }
                }
            },
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.GeoM.Scale(float64(gridWidth) / float64(pic.Bounds().Dx()), float64(gridHeight) / float64(pic.Bounds().Dy()))
                options.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))

                if lookTime > 0 {
                    options.ColorScale.SetR(1.0 + float32(lookTime) / float32(maxLookTime) / 2)
                    options.ColorScale.SetG(1.0 + float32(lookTime) / float32(maxLookTime) / 2)
                }

                scale.DrawScaled(screen, pic, &options)
            },
        })

        gridX += 1
        if gridX == 6 {
            gridX = 0
            gridY += 1
        }
    }

    return group, quit
}

type SelectedEnchantmentFunc func(data.Enchantment, *playerlib.Player, *string, image.Rectangle, *util.AlphaFadeFunc)

func (game *Game) MakeSpellBindingUI(caster *playerlib.Player, spell spellbook.Spell) (*uilib.UIElementGroup, context.Context, error) {
    var group *uilib.UIElementGroup
    var quit context.Context
    var cancel context.CancelFunc

    // ugly to need this here
    fadeSpeed := 7

    // A func for creating a sparks element when a target is selected
    createSparksElement := func (faceRect image.Rectangle, fader *util.AlphaFadeFunc) *uilib.UIElement {
        sparksCreationTick := group.Counter // Needed for sparks animation
        return &uilib.UIElement{
            Layer: 2,
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                const ticksPerFrame = 5
                frameToShow := int((group.Counter - sparksCreationTick) / ticksPerFrame) % 6
                background, _ := game.ImageCache.GetImage("specfx.lbx", 40, frameToShow)
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha((*fader)())
                options.GeoM.Translate(float64(faceRect.Min.X - 5), float64(faceRect.Min.Y - 10))
                scale.DrawScaled(screen, background, &options)
            },
        }
    }

    selectedEnchantment := func (enchantment data.Enchantment, owner *playerlib.Player, uiTitle *string, faceRect image.Rectangle, fader *util.AlphaFadeFunc) {
        dispelStrength := 20000

        allSpells := game.AllSpells()
        targetSpell := allSpells.FindByName(enchantment.String())

        sound, err := audio.LoadSound(game.Cache, spell.Sound)
        if err == nil {
            sound.Play()
        }

        group.AddElement(createSparksElement(faceRect, fader))
        success := false

        if spellbook.RollDispelChance(spellbook.ComputeDispelChance(dispelStrength, targetSpell.Cost(true), targetSpell.Magic, &owner.Wizard)) {
            success = true
            owner.RemoveEnchantment(enchantment)
            caster.AddEnchantment(enchantment)

            game.ApplyGlobalEnchantment(enchantment, caster)
        }

        group.AddDelay(60, func(){
            if success {
                *uiTitle = fmt.Sprintf("%s has been stolen", enchantment.String())
            } else {
                *uiTitle = "Spell binding failed"
            }
            group.AddDelay(113, func(){
                *fader = group.MakeFadeOut(uint64(fadeSpeed))
                group.AddDelay(7, func(){
                    cancel()
                })
            })
        })

    }

    var err error
    group, quit, cancel, err = game.makeGlobalEnchantmentSelectionUI(caster, spell, selectedEnchantment)
    return group, quit, err
}

func (game *Game) MakeDisjunctionUI(caster *playerlib.Player, spell spellbook.Spell) (*uilib.UIElementGroup, context.Context, error) {
    dispelStrength := spell.Cost(true)
    if spell.Name == "Disjunction True" {
        dispelStrength *= 3
    }

    if caster.Wizard.RetortEnabled(data.RetortRunemaster) {
        dispelStrength *= 2
    }

    var quit context.Context
    var cancel context.CancelFunc
    var group *uilib.UIElementGroup

    // ugly to need this here
    fadeSpeed := 7

    // A func for creating a sparks element when a target is selected
    createSparksElement := func (faceRect image.Rectangle, fader *util.AlphaFadeFunc) *uilib.UIElement {
        sparksCreationTick := group.Counter // Needed for sparks animation
        return &uilib.UIElement{
            Layer: 2,
            Draw: func(element *uilib.UIElement, screen *ebiten.Image) {
                const ticksPerFrame = 5
                frameToShow := int((group.Counter - sparksCreationTick) / ticksPerFrame) % 6
                background, _ := game.ImageCache.GetImage("specfx.lbx", 40, frameToShow)
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha((*fader)())
                options.GeoM.Translate(float64(faceRect.Min.X - 5), float64(faceRect.Min.Y - 10))
                scale.DrawScaled(screen, background, &options)
            },
        }
    }

    selectedEnchantment := func (enchantment data.Enchantment, owner *playerlib.Player, uiTitle *string, faceRect image.Rectangle, fader *util.AlphaFadeFunc) {
        allSpells := game.AllSpells()
        targetSpell := allSpells.FindByName(enchantment.String())

        group.AddElement(createSparksElement(faceRect, fader))
        // FIXME: verify this sound
        sound, err := audio.LoadSound(game.Cache, 29)
        if err == nil {
            sound.Play()
        }

        success := false

        if spellbook.RollDispelChance(spellbook.ComputeDispelChance(dispelStrength, targetSpell.Cost(true), targetSpell.Magic, &owner.Wizard)) {
            // show an animation/play a sound?
            owner.RemoveEnchantment(enchantment)
            success = true
        }

        group.AddDelay(60, func(){
            if success {
                *uiTitle = fmt.Sprintf("%s has been disjuncted", enchantment.String())
            } else {
                *uiTitle = "Disjunction failed"
            }
            group.AddDelay(113, func(){
                *fader = group.MakeFadeOut(uint64(fadeSpeed))
                group.AddDelay(7, func(){
                    cancel()
                })
            })
        })
    }

    var err error
    group, quit, cancel, err = game.makeGlobalEnchantmentSelectionUI(caster, spell, selectedEnchantment)
    return group, quit, err
}

func (game *Game) makeGlobalEnchantmentSelectionUI(caster *playerlib.Player, spell spellbook.Spell, selectedEnchantment SelectedEnchantmentFunc) (*uilib.UIElementGroup, context.Context, context.CancelFunc, error) {
    group := uilib.MakeGroup()
    quit, cancel := context.WithCancel(context.Background())

    fadeSpeed := 7

    fader := group.MakeFadeIn(uint64(fadeSpeed))

    const uiX = 30

    specialFonts := MakeSpellSpecialUIFonts(game.Cache)

    header := "Select a spell to disjunct."

    group.AddElement(&uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := game.ImageCache.GetImage("spellscr.lbx", 1, 0)
            var options ebiten.DrawImageOptions
            options.ColorScale.ScaleAlpha(fader())
            options.GeoM.Translate(uiX, 1)
            scale.DrawScaled(screen, background, &options)

            specialFonts.BigOrange.PrintOptions(screen, float64(uiX + background.Bounds().Dx() / 2), 5, font.FontOptions{Justify: font.FontJustifyCenter, Scale: scale.ScaleAmount, DropShadow: true, Options: &options}, header)
        },
        // a hack to get around go vet complaining that cancel is never called
        Hack: func(element *uilib.UIElement) {
            cancel()
        },
        // maybe click away raises a confirmation box that asks if you want to cancel the spell?
        /*
        NotLeftClicked: func(element *uilib.UIElement) {
            // log.Printf("Cancel ui")
            fader = group.MakeFadeOut(uint64(fadeSpeed))
            group.AddDelay(uint64(fadeSpeed), cancel)
        },
        */
    })

    // count how many enchantments are available to disjunction
    // if this remains 0 then there are no options, so this function will return an error
    enchantmentOptions := 0

    fonts := fontslib.MakeMagicViewFonts(game.Cache)

    enabled := true

    // FIXME: only show enchantments of known players?
    for index, player := range caster.GetKnownPlayers() {
        brokenCrystalPicture, _ := game.ImageCache.GetImage("magic.lbx", 51, 0)
        portrait, _ := game.ImageCache.GetImage("lilwiz.lbx", mirror.GetWizardPortraitIndex(player.Wizard.Base, player.GetBanner()), 0)

        yBase := 15 + 46 * index

        faceRect := util.ImageRect(uiX + 8, yBase, portrait)

        group.AddElement(&uilib.UIElement{
            Layer: 1,
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(fader())
                options.GeoM.Translate(uiX + 8, float64(yBase))
                if player.Defeated {
                    scale.DrawScaled(screen, brokenCrystalPicture, &options)
                } else {
                    scale.DrawScaled(screen, portrait, &options)
                }
            },
        })

        minIndex := 0

        var enchantmentList []*uilib.UIElement

        for enchantmentIndex, enchantment := range slices.SortedFunc(slices.Values(player.GlobalEnchantments.Values()), cmp.Compare) {
            enchantmentOptions += 1

            y := yBase + 5 + 13 * enchantmentIndex
            var options ebiten.DrawImageOptions

            shadow := font.FontOptions{DropShadow: true, Scale: scale.ScaleAmount, Options: &options}

            rect := image.Rect(uiX + 75, y, uiX + 75 + 100, y + 13)
            hover := false
            enchantmentList = append(enchantmentList, &uilib.UIElement{
                Layer: 1,
                Order: 1,
                Rect: rect,
                Inside: func(element *uilib.UIElement, x int, y int) {
                    hover = true
                },
                NotInside: func(element *uilib.UIElement) {
                    hover = false
                },
                RightClick: func(element *uilib.UIElement) {
                    if !enabled {
                        return
                    }

                    helpEntries := game.Help.GetEntriesByName(enchantment.String())
                    if helpEntries != nil {
                        group.AddElement(uilib.MakeHelpElementWithLayer(group, game.Cache, &game.ImageCache, 2, helpEntries[0], helpEntries[1:]...))
                    }
                },
                LeftClick: func(element *uilib.UIElement) {
                    if enabled && enchantmentIndex - minIndex >= 0 && enchantmentIndex - minIndex < 3 {
                        selectedEnchantment(enchantment, player, &header, faceRect, &fader)
                    }
                },
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    if enchantmentIndex - minIndex >= 0 && enchantmentIndex - minIndex < 3 {
                        options.ColorScale.Reset()
                        options.ColorScale.ScaleAlpha(fader())
                        options.ColorScale.SetR(1)
                        options.ColorScale.SetG(1)
                        options.ColorScale.SetB(1)

                        if hover {
                            options.ColorScale.SetR(2)
                            options.ColorScale.SetG(2)
                        }

                        fonts.NormalFont.PrintOptions(screen, float64(element.Rect.Min.X + 2), float64(element.Rect.Min.Y), shadow, enchantment.String())
                    }
                },
            })
        }

        group.AddElements(enchantmentList)

        if len(enchantmentList) > 3 {
            scroll := func (direction int) {
                minIndex += direction
                if minIndex < 0 {
                    minIndex = 0
                }
                if minIndex > len(enchantmentList) - 3 {
                    minIndex = len(enchantmentList) - 3
                }

                for i, element := range enchantmentList {
                    element.Rect.Min.Y = yBase + 5 + 13 * (i - minIndex)
                    element.Rect.Max.Y = element.Rect.Min.Y + 13
                }
            }

            // up arrow
            upClicked := false
            upArrows, _ := game.ImageCache.GetImages("resource.lbx", 32)
            rect := util.ImageRect(uiX + 61, yBase + 5, upArrows[0])
            group.AddElement(&uilib.UIElement{
                Rect: rect,
                Layer: 1,
                Order: 1,
                LeftClick: func(element *uilib.UIElement) {
                    upClicked = true
                },
                LeftClickRelease: func(element *uilib.UIElement) {
                    upClicked = false
                    scroll(-1)
                },
                PlaySoundLeftClick: true,
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    var options ebiten.DrawImageOptions
                    options.ColorScale.ScaleAlpha(fader())
                    options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
                    if upClicked {
                        scale.DrawScaled(screen, upArrows[1], &options)
                    } else {
                        scale.DrawScaled(screen, upArrows[0], &options)
                    }
                },
            })

            // down arrow
            downClicked := false
            downArrows, _ := game.ImageCache.GetImages("resource.lbx", 33)
            downRect := util.ImageRect(uiX + 61, yBase + 29, downArrows[0])
            group.AddElement(&uilib.UIElement{
                Rect: downRect,
                Layer: 1,
                Order: 1,
                LeftClick: func(element *uilib.UIElement) {
                    downClicked = true
                },
                LeftClickRelease: func(element *uilib.UIElement) {
                    downClicked = false
                    scroll(1)
                },
                PlaySoundLeftClick: true,
                Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                    var options ebiten.DrawImageOptions
                    options.ColorScale.ScaleAlpha(fader())
                    options.GeoM.Translate(float64(element.Rect.Min.X), float64(element.Rect.Min.Y))
                    if downClicked {
                        scale.DrawScaled(screen, downArrows[1], &options)
                    } else {
                        scale.DrawScaled(screen, downArrows[0], &options)
                    }
                },
            })
        }
    }

    for i := range 4 - len(caster.GetKnownPlayers()) {
        crystalPicture, _ := game.ImageCache.GetImage("magic.lbx", 6, 0)
        index := i + len(caster.GetKnownPlayers())
        group.AddElement(&uilib.UIElement{
            Layer: 1,
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                var options ebiten.DrawImageOptions
                options.ColorScale.ScaleAlpha(fader())
                options.GeoM.Translate(uiX + 8, float64(15 + 46 * index))
                scale.DrawScaled(screen, crystalPicture, &options)
            },
        })
    }

    if enchantmentOptions == 0 {
        cancel()
        return nil, quit, cancel, errors.New("There are no global spells to disjunct")
    }

    return group, quit, cancel, nil
}

func (game *Game) MakeSubversionUI(caster *playerlib.Player, spell spellbook.Spell) (*uilib.UIElementGroup, context.Context, error) {
    onTargetSelectCallback := func(targetPlayer *playerlib.Player) (bool, string) {
        if targetPlayer.Defeated || targetPlayer.Banished {
            return false, ""
        }

        for _, player := range game.Players {
            // ignore the wizard that cast subversion
            if player == caster {
                continue
            }

            player.AdjustDiplomaticRelation(targetPlayer, -25)
        }

        return true, fmt.Sprintf("%s has been subverted", targetPlayer.Wizard.Name)
    }

    playersInGame := len(game.Players)
    quit, cancel := context.WithCancel(context.Background())
    wizSelectionUiGroup := makeSelectTargetWizardUI(cancel, game.Cache, &game.ImageCache, "Choose target for a Subversion spell", 43, spell.Sound, caster, playersInGame, onTargetSelectCallback)
    return wizSelectionUiGroup, quit, nil

    /*
    knownPlayers := len(caster.GetKnownPlayers())
    alive := 0
    for _, player := range knownPlayers {
        if !player.Defeated {
            alive += 1
        }
    }

    if alive == 0 {
        return nil, nil, errors.New("No known players to subvert")
    }

    group := uilib.MakeGroup()

    quit, cancel := context.WithCancel(context.Background())

    const uiX = 30
    const uiY = 10

    group.AddElement(&uilib.UIElement{
        Layer: 1,
        Draw: func(element *uilib.UIElement, screen *ebiten.Image){
            background, _ := game.ImageCache.GetImage("spellscr.lbx", 0, 0)
        },
    })

    return group, quit, nil
    */
}

// Returns true if the spell is rolled to be instantly fizzled on cast (caused by spells like Life Force)
// also returns the reason why the fizzle occurred (due to tranquility, supress magic, etc
func (game *Game) checkInstantFizzleForCastSpell(player *playerlib.Player, spell spellbook.Spell) (bool, FizzleReason) {
    var dispelChances []FizzleReason

    for _, checkingPlayer := range game.Players {

        // Tranquility effect: if it's a chaos spell, it should either resist a strength 500 dispel check or fizzle right away.
        if spell.IsOfRealm(data.ChaosMagic) {
            // FIXME: Not sure if multiple instances of Tranquility stack or are checked separately.
            if checkingPlayer != player && checkingPlayer.HasEnchantment(data.EnchantmentTranquility) {
                dispelChances = append(dispelChances, FizzleReason{Owner: checkingPlayer, Enchantment: data.EnchantmentTranquility})
            }
        }
        // Life Force effect: if it's a death spell, it should either resist a strength 500 dispel check or fizzle right away.
        if spell.IsOfRealm(data.DeathMagic) {
            // FIXME: Not sure if multiple instances of Life Force stack or are checked separately.
            if checkingPlayer != player && checkingPlayer.HasEnchantment(data.EnchantmentLifeForce) {
                dispelChances = append(dispelChances, FizzleReason{Owner: checkingPlayer, Enchantment: data.EnchantmentLifeForce})
            }
        }

        if checkingPlayer != player && checkingPlayer.HasEnchantment(data.EnchantmentSuppressMagic) {
            dispelChances = append(dispelChances, FizzleReason{Owner: checkingPlayer, Enchantment: data.EnchantmentSuppressMagic})
        }
    }

    for _, reason := range dispelChances {
        if spellbook.RollDispelChance(spellbook.ComputeDispelChance(500, spell.Cost(true), spell.Magic, &player.Wizard)) {
            return true, reason
        }
    }
    return false, FizzleReason{}
}

func (game *Game) doCastSpellWard(player *playerlib.Player, spell spellbook.Spell) {
    wards := []data.CityEnchantment{
        data.CityEnchantmentLifeWard, data.CityEnchantmentSorceryWard,
        data.CityEnchantmentNatureWard, data.CityEnchantmentDeathWard,
        data.CityEnchantmentChaosWard,
    }

    var selectCity func (coroutine.YieldFunc, int, int)
    selectCity = func (yield coroutine.YieldFunc, tileX int, tileY int) {
        // FIXME: Show this only for enemies if detect magic is active and the city is known to the human player
        game.doMoveCamera(yield, tileX, tileY)
        chosenCity, _ := game.FindCity(tileX, tileY, game.Plane)
        if chosenCity == nil {
            return
        }

        choices := set.NewSet(wards...)

        for _, enchantment := range choices.Values() {
            if chosenCity.HasEnchantment(enchantment) {
                choices.Remove(enchantment)
            }
        }

        if choices.Size() == 0 {
            game.Events <- &GameEventNotice{Message: "No wards are available to cast on this city."}
            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeFriendlyCity, SelectedFunc: selectCity}
            return
        }

        quit, cancel := context.WithCancel(context.Background())

        selected := func (ward data.CityEnchantment){
            // invoking cancel removes the selection ui group
            cancel()
            // this yield reference comes from a different time in the ui when the tile was being selected
            // really this should be a GameEvent that shows the enchantment being added
            game.doAddCityEnchantment(yield, chosenCity, player, spell, ward)
        }

        var selections []uilib.Selection

        for _, ward := range slices.SortedFunc(slices.Values(choices.Values()), cmp.Compare) {
            selections = append(selections, uilib.Selection{
                Name: ward.Name(),
                Action: func(){
                    selected(ward)
                },
            })
        }

        uiGroup := uilib.MakeGroup()

        uiGroup.AddElements(uilib.MakeSelectionUI(uiGroup, game.Cache, &game.ImageCache, 40, 10, "Select a Spell Ward to cast", selections, false))
        game.Events <- &GameEventRunUI{
            Group: uiGroup,
            Quit: quit,
        }
    }

    game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeFriendlyCity, SelectedFunc: selectCity}
}

func (game *Game) doDisenchantArea(yield coroutine.YieldFunc, player *playerlib.Player, spell spellbook.Spell, disenchantTrue bool, tileX int, tileY int) {
    game.doCastOnMap(yield, tileX, tileY, 9, spell.Sound, func (x int, y int, animationFrame int){})

    disenchantStrength := spell.Cost(true)
    if disenchantTrue {
        // strength is 3x mana cost
        disenchantStrength = spell.Cost(true) * 3
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

func (game *Game) doCastOnUnit(player *playerlib.Player, spell spellbook.Spell, animationIndex int, before UnitEnchantmentCallback, after UnitEnchantmentCallback) {
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

        if !before(unit) {
            game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeFriendlyUnit, SelectedFunc: selected}
            return
        }

        game.doCastOnMap(yield, tileX, tileY, animationIndex, spell.Sound, func (x int, y int, animationFrame int) {})

        after(unit)

        game.RefreshUI()
    }

    game.Events <- &GameEventSelectLocationForSpell{Spell: spell, Player: player, LocationType: LocationTypeFriendlyUnit, SelectedFunc: selected}
}

func (game *Game) doCastUnitEnchantmentFull(player *playerlib.Player, spell spellbook.Spell, enchantment data.UnitEnchantment, customBefore UnitEnchantmentCallback, customAfter UnitEnchantmentCallback) {
    before := func (unit units.StackUnit) bool {
        if unit.HasEnchantment(enchantment) {
            game.Events <- &GameEventNotice{Message: fmt.Sprintf("That unit already has %v cast on it", spell.Name)}
            return false
        }

        return customBefore(unit)
    }

    after := func (unit units.StackUnit) bool {
        unit.AddEnchantment(enchantment)

        return customAfter(unit)
    }

    game.doCastOnUnit(player, spell, enchantment.CastAnimationIndex(), before, after)
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
            ui.StandardDraw(screen)
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
            Female: hero.IsFemale(),
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
        overworldUnit := units.MakeOverworldUnitFromUnit(unit, summonCity.X, summonCity.Y, summonCity.Plane, player.Wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider())
        player.AddUnit(overworldUnit)
        game.ResolveStackAt(summonCity.X, summonCity.Y, summonCity.Plane)
        game.RefreshUI()
    }
}

func (game *Game) showCityEarthquake(yield coroutine.YieldFunc, city *citylib.City, player *playerlib.Player) {
    ui, quit, updateDestroyed, err := cityview.MakeEarthquakeView(game.Cache, city, player)
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

    counter := game.Counter

    yield()

    for quit.Err() == nil && game.Counter < counter + 120 {
        game.Counter += 1
        ui.StandardUpdate()
        if yield() != nil {
            return
        }
    }

    _, _, buildings := game.doEarthquake(city, player)
    destroyed := set.NewSet(buildings...)

    updateDestroyed(destroyed)

    counter = game.Counter

    for quit.Err() == nil && game.Counter < counter + 180 {
        game.Counter += 1
        ui.StandardUpdate()
        if yield() != nil {
            return
        }
    }

    // absorb left click
    yield()

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

    fonts := fontslib.MakeSurveyorFonts(game.Cache)
    castingFont := fonts.SurveyorFont
    whiteFont := fonts.WhiteFont

    makeOverworld := func () Overworld {
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

        return Overworld{
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
    }

    overworld := makeOverworld()

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
        case LocationTypeEnemyUnit:
            selectMessage = fmt.Sprintf("Select an enemy unit as the target for a %v spell.", spell.Name)
        default:
            selectMessage = fmt.Sprintf("unhandled location type %v", locationType)
    }

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image){
            var options ebiten.DrawImageOptions
            mainHud, _ := game.ImageCache.GetImage("main.lbx", 0, 0)
            scale.DrawScaled(screen, mainHud, &options)

            landImage, _ := game.ImageCache.GetImage("main.lbx", 57, 0)
            options.GeoM.Translate(float64(240), float64(77))
            scale.DrawScaled(screen, landImage, &options)

            options.GeoM.Reset()
            options.GeoM.Translate(float64(240), float64(174))
            scale.DrawScaled(screen, cancelBackground, &options)

            ui.StandardDraw(screen)

            game.Fonts.WhiteFont.PrintRight(screen, float64(276), float64(68), scale.ScaleAmount, ebiten.ColorScale{}, fmt.Sprintf("%v GP", game.Players[0].Gold))
            game.Fonts.WhiteFont.PrintRight(screen, float64(313), float64(68), scale.ScaleAmount, ebiten.ColorScale{}, fmt.Sprintf("%v MP", game.Players[0].Mana))

            castingFont.PrintCenter(screen, float64(280), float64(81), scale.ScaleAmount, ebiten.ColorScale{}, "Casting")

            whiteFont.PrintWrapCenter(screen, float64(280), float64(120), float64(cancelBackground.Bounds().Dx() - 5), scale.ScaleAmount, ebiten.ColorScale{}, selectMessage)
        },
    }

    ui.SetElementsFromArray(nil)

    makeButton2 := func(lbxIndex int, x int, y int, action func()) *uilib.UIElement {
        buttons, _ := game.ImageCache.GetImages("main.lbx", lbxIndex)
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(x), float64(y))
        current := 0
        return &uilib.UIElement{
            Rect: util.ImageRect(x, y, buttons[0]),
            Draw: func(element *uilib.UIElement, screen *ebiten.Image){
                scale.DrawScaled(screen, buttons[current], &options)
            },
            LeftClick: func(element *uilib.UIElement){
                current = 1
            },
            LeftClickRelease: func(element *uilib.UIElement){
                action()
                current = 0
            },
        }
    }

    makeButton := func(lbxIndex int, x int, y int) *uilib.UIElement {
        return makeButton2(lbxIndex, x, y, func(){})
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
    ui.AddElement(makeButton2(7, 270, 4, func (){
        game.SwitchPlane()
        overworld = makeOverworld()
    }))

    quit := false

    // cancel button at bottom
    cancel, _ := game.ImageCache.GetImages("main.lbx", 41)
    cancelIndex := 0
    cancelRect := util.ImageRect(263, 182, cancel[0])
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
            scale.DrawScaled(screen, cancel[cancelIndex], &options)
        },
    })

    game.Drawer = func(screen *ebiten.Image, game *Game){
        overworld.Camera = game.Camera
        overworld.DrawOverworld(screen, ebiten.GeoM{})

        var miniGeom ebiten.GeoM
        miniGeom.Translate(float64(250), float64(20))
        mx, my := miniGeom.Apply(0, 0)
        miniWidth := 60
        miniHeight := 31
        mini := screen.SubImage(scale.ScaleRect(image.Rect(int(mx), int(my), int(mx) + miniWidth, int(my) + miniHeight))).(*ebiten.Image)
        overworld.DrawMinimap(mini)

        ui.Draw(ui, screen)
    }

    entityInfo := game.ComputeCityStackInfo()

    for !quit {
        overworld.Counter += 1

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
                                if !city.CanTarget(spell) {
                                    game.doNotice(yield, ui, fmt.Sprintf("You cannot cast %v on this city", spell.Name))
                                    break
                                } else {
                                    return tileX, tileY, false
                                }
                            }
                        }

                    case LocationTypeFriendlyUnit:
                        stack := player.FindStack(tileX, tileY, game.Plane)
                        if stack != nil {
                            return tileX, tileY, false
                        }

                    case LocationTypeEnemyUnit:
                        // also consider if the unit is in a city with a spell ward that prevents this unit from being targeted
                        if player.IsVisible(tileX, tileY, game.Plane) {
                            stack := entityInfo.FindStack(tileX, tileY, game.Plane)
                            if stack != nil && entityInfo.ContainsEnemy(tileX, tileY, game.Plane, player) {
                                city := entityInfo.FindCity(tileX, tileY, game.Plane)

                                if city != nil && !city.CanTarget(spell) {
                                    game.doNotice(yield, ui, fmt.Sprintf("You cannot cast %v on this unit", spell.Name))
                                    break
                                } else {
                                    return tileX, tileY, false
                                }
                            }
                        }
                }
            }
        }

        if yield() != nil {
            break
        }
    }

    return 0, 0, true
}

func (game *Game) doAddCityEnchantment(yield coroutine.YieldFunc, chosenCity *citylib.City, player *playerlib.Player, spell spellbook.Spell, enchantment data.CityEnchantment) {
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

        game.doAddCityEnchantment(yield, chosenCity, player, spell, enchantment)

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

func (game *Game) doCastOnMap(yield coroutine.YieldFunc, tileX int, tileY int, animationIndex int, soundIndex int, update UpdateMapFunction) {
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
        scale.DrawScaled(screen, animation.Frame(), &options)
    }

    sound, err := audio.LoadSound(game.Cache, soundIndex)
    if err == nil {
        sound.Play()
    } else {
        log.Printf("No such sound %v for spell", soundIndex)
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

    game.doCastOnMap(yield, tileX, tileY, 46, 86, update)

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

func (game *Game) doCastEarthLore(yield coroutine.YieldFunc, tileX int, tileY int, player *playerlib.Player, soundIndex int) {
    update := func (x int, y int, frame int) {}

    game.doCastOnMap(yield, tileX, tileY, 45, soundIndex, update)

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

    game.doCastOnMap(yield, tileX, tileY, 8, 28, update)
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

    game.doCastOnMap(yield, tileX, tileY, 0, 28, update)
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

    game.doCastOnMap(yield, tileX, tileY, 11, 98, update)

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

    game.doCastOnMap(yield, tileX, tileY, 7, 103, update)
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
    quit, cancel := context.WithCancel(context.Background())
    wizSelectionUiGroup = makeSelectSpellBlastTargetUI(cancel, game.Cache, &game.ImageCache, player, playersInGame, onTargetSelectCallback)
    game.Events <- &GameEventRunUI{
        Group: wizSelectionUiGroup,
        Quit: quit,
    }
}

func (game *Game) doCastCruelUnminding(player *playerlib.Player, spell spellbook.Spell) {
    onTargetSelectCallback := func(targetPlayer *playerlib.Player) (bool, string) {
        if targetPlayer.Defeated || targetPlayer.Banished {
            return false, ""
        }
        targetSkill := targetPlayer.ComputeCastingSkill()
        minReduction := max(targetSkill / 100, 1)
        maxReduction := max(targetSkill / 10, 1)
        reductionRandomSpread := max(maxReduction - minReduction, 1) // It should never be zero, or rand.N will crash
        reduction := minReduction + rand.N(reductionRandomSpread)
        actuallyReduced := targetPlayer.ReduceCastingSkill(reduction)
        return true, fmt.Sprintf("%s loses %d points of casting ability", targetPlayer.Wizard.Name, actuallyReduced)
    }
    playersInGame := len(game.Players)
    quit, cancel := context.WithCancel(context.Background())
    wizSelectionUiGroup := makeSelectTargetWizardUI(cancel, game.Cache, &game.ImageCache, "Select target for Cruel Unminding spell", 41, spell.Sound, player, playersInGame, onTargetSelectCallback)
    game.Events <- &GameEventRunUI{
        Group: wizSelectionUiGroup,
        Quit: quit,
    }
}

func (game *Game) doCastDrainPower(player *playerlib.Player, spell spellbook.Spell) {
    onTargetSelectCallback := func(targetPlayer *playerlib.Player) (bool, string) {
        if targetPlayer.Defeated || targetPlayer.Banished {
            return false, ""
        }
        drainAmount := min(targetPlayer.Mana, 50 + rand.N(101)) // random 50-150, but don't drain more than the target has
        targetPlayer.Mana -= drainAmount
        return true, fmt.Sprintf("%s loses %d points of mana", targetPlayer.Wizard.Name, drainAmount)
    }
    playersInGame := len(game.Players)
    quit, cancel := context.WithCancel(context.Background())
    wizSelectionUiGroup := makeSelectTargetWizardUI(cancel, game.Cache, &game.ImageCache, "Select target for Drain Power spell", 42, spell.Sound, player, playersInGame, onTargetSelectCallback)
    game.Events <- &GameEventRunUI{
        Group: wizSelectionUiGroup,
        Quit: quit,
    }
}

func (game *Game) doCastWarpNode(yield coroutine.YieldFunc, tileX int, tileY int, caster *playerlib.Player, soundIndex int) {
    update := func (x int, y int, frame int) {}

    game.doCastOnMap(yield, tileX, tileY, 13, soundIndex, update)

    node := game.CurrentMap().GetMagicNode(tileX, tileY)
    if node != nil {
        node.Warped = true
        node.WarpedOwner = caster
    }
}

type GlobalEnchantmentFonts struct {
    InfoFont *font.Font
}

func MakeGlobalEnchantmentFonts(cache *lbx.LbxCache) *GlobalEnchantmentFonts {
    loader, err := fontslib.Loader(cache)
    if err != nil {
        log.Printf("Error loading global enchantment fonts: %v", err)
        return nil
    }

    return &GlobalEnchantmentFonts{
        InfoFont: loader(fontslib.InfoFont),
    }
}

func (game *Game) doCastGlobalEnchantment(yield coroutine.YieldFunc, player *playerlib.Player, enchantment data.Enchantment, after func()) {

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
        case data.GreatUnsummoning: song = music.SongGreatUnsummoning
        case data.DeathWish: song = music.SongDeathWish
    }

    if song != music.SongNone {
        game.Music.PushSong(song)
        defer game.Music.PopSong()
    }

    fonts := MakeGlobalEnchantmentFonts(game.Cache)

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

    offset := -35

    game.Drawer = func(screen *ebiten.Image, game *Game){
        oldDrawer(screen, game)
        var options ebiten.DrawImageOptions
        options.GeoM.Translate(float64(data.ScreenWidth / 2), float64(data.ScreenHeight / 2))
        options.GeoM.Translate(float64(offset), 0)
        options.GeoM.Translate(float64(-frame.Bounds().Dx() / 2), float64(-frame.Bounds().Dy() / 2))
        scale.DrawScaled(screen, frame, &options)

        options.ColorScale.ScaleAlpha(fader())

        // first draw the wizard
        if doDraw == 0 {
            mood, _ := game.ImageCache.GetImageTransform("moodwiz.lbx", animationIndex, 2, "cutout", makeCutoutMask)
            options.GeoM.Translate(float64(13), float64(13))
            scale.DrawScaled(screen, mood, &options)

            text := "You have finished casting"
            if !player.IsHuman() {
                text = fmt.Sprintf("%v has cast", player.Wizard.Name)
            }

            fonts.InfoFont.PrintCenter(screen, float64(data.ScreenWidth / 2 + offset), float64(data.ScreenHeight / 2 + frame.Bounds().Dy() / 2), scale.ScaleAmount, options.ColorScale, text)
        } else {
            // then draw the spell image
            options.GeoM.Translate(float64(9), float64(8))
            scale.DrawScaled(screen, spellImage, &options)

            fonts.InfoFont.PrintCenter(screen, float64(data.ScreenWidth / 2 + offset), float64(data.ScreenHeight / 2 + frame.Bounds().Dy() / 2), scale.ScaleAmount, options.ColorScale, enchantment.String())
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

    after()
}

func (game *Game) doCastFloatingIsland(yield coroutine.YieldFunc, player *playerlib.Player, tileX int, tileY int) {
    update := func (x int, y int, frame int) {
        if frame == 5 {
            overworldUnit := units.MakeOverworldUnitFromUnit(units.FloatingIsland, tileX, tileY, game.CurrentMap().Plane, player.Wizard.Banner, player.MakeExperienceInfo(), player.MakeUnitEnchantmentProvider())
            player.AddUnit(overworldUnit)
            player.LiftFog(tileX, tileY, 1, game.Plane)
        }
    }

    game.doCastOnMap(yield, tileX, tileY, 1, 29, update)
}
