package spellbook

import (
    "math/rand/v2"

    "github.com/kazzmir/master-of-magic/game/magic/data"
)

type RetortOwner interface {
    RetortEnabled(retort data.Retort) bool
}

// Returns dispel chance against 250.
// TargetSpellCost may be either a base cost or an effective cost (https://masterofmagic.fandom.com/wiki/Casting_Cost#Dispelling_Magic)
// If checkTargetSpellOwnerRetorts is true, then Archmage and Mastery retorts will be taken into consideration for dispel resistance
// FIXME: add Runemaster check here (it should increase the dispel strength)
func ComputeDispelChance(dispelStrength int, targetSpellCost int, targetSpellRealm data.MagicType, targetSpellOwner RetortOwner) int {

    dispelResistanceModifier := 1

    if targetSpellOwner.RetortEnabled(data.RetortArchmage) {
        dispelResistanceModifier += 1
    }

    if targetSpellRealm == data.NatureMagic && targetSpellOwner.RetortEnabled(data.RetortNatureMastery) {
        dispelResistanceModifier += 1
    }

    if targetSpellRealm == data.SorceryMagic && targetSpellOwner.RetortEnabled(data.RetortSorceryMastery) {
        dispelResistanceModifier += 1
    }

    if targetSpellRealm == data.ChaosMagic && targetSpellOwner.RetortEnabled(data.RetortChaosMastery) {
        dispelResistanceModifier += 1
    }

    // The original game uses the check in multiples of 250, and not as a percentage (https://masterofmagic.fandom.com/wiki/Casting_Cost#Dispelling_Magic)
    return (250 * dispelStrength) / (dispelStrength + (targetSpellCost * dispelResistanceModifier))
}

// Returns true if dispel is successful.
// The original game uses the check in multiples of 250, and not as a percentage (https://masterofmagic.fandom.com/wiki/Casting_Cost#Dispelling_Magic)
func RollDispelChance(chanceAgainst250 int) bool {
    return rand.N(250) < chanceAgainst250
}
