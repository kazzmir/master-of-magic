package artifact

import (
    "slices"

    "github.com/kazzmir/master-of-magic/game/magic/data"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/lib/set"
)

// NotPremade is CatalogIndex for an item that is not one of the 250
// ITEMDATA slots (Enchant Item / Create Artifact output).
const NotPremade = -1

// Catalog is the 250-slot premade item list plus which slots have not
// yet been awarded this campaign. Identity is the slot index, matching
// ITEMMAKE's "Item #N".
type Catalog struct {
    Slots []*Artifact
    available []bool
}

func MakeCatalog(items []Artifact) *Catalog {
    catalog := &Catalog{
        Slots: make([]*Artifact, len(items)),
        available: make([]bool, len(items)),
    }

    for i, item := range items {
        clone := CloneArtifact(&item)
        clone.CatalogIndex = i
        catalog.Slots[i] = clone
        catalog.available[i] = true
    }

    return catalog
}

func (catalog *Catalog) Len() int {
    if catalog == nil {
        return 0
    }

    return len(catalog.Slots)
}

func (catalog *Catalog) IsAvailable(index int) bool {
    if catalog == nil || index < 0 || index >= len(catalog.available) {
        return false
    }

    return catalog.available[index]
}

// AvailableArtifacts returns the slots that have not been awarded yet,
// in catalog order. Callers must not assume the slice is a live view.
func (catalog *Catalog) AvailableArtifacts() []*Artifact {
    if catalog == nil {
        return nil
    }

    var out []*Artifact
    for i, slot := range catalog.Slots {
        if catalog.available[i] {
            out = append(out, slot)
        }
    }

    return out
}

func (catalog *Catalog) AvailableNames() []string {
    var names []string
    for _, item := range catalog.AvailableArtifacts() {
        names = append(names, item.Name)
    }

    return names
}

func (catalog *Catalog) FindByName(name string) *Artifact {
    if catalog == nil {
        return nil
    }

    for _, slot := range catalog.Slots {
        if slot != nil && slot.Name == name {
            return slot
        }
    }

    return nil
}

// Award marks the premade slot that item came from as used. Crafted items
// (or anything that does not match a slot) are ignored, matching the old
// delete-from-name-map no-op.
func (catalog *Catalog) Award(item *Artifact) {
    if catalog == nil || item == nil {
        return
    }

    for i, slot := range catalog.Slots {
        if slot == item {
            catalog.available[i] = false
            return
        }
    }

    catalog.AwardByName(item.Name)
}

func (catalog *Catalog) AwardByName(name string) {
    if catalog == nil || name == "" {
        return
    }

    for i, slot := range catalog.Slots {
        if catalog.available[i] && slot != nil && slot.Name == name {
            catalog.available[i] = false
            return
        }
    }
}

// Replace copies item into slot index without changing availability and
// without replacing the slot pointer. Awarded copies (hero equipment,
// vault) share that pointer, so an in-game edit updates items already
// in play. Used by the Default Item Editor.
func (catalog *Catalog) Replace(index int, item *Artifact) {
    if catalog == nil || item == nil || index < 0 || index >= len(catalog.Slots) {
        return
    }

    slot := catalog.Slots[index]
    if slot == nil {
        slot = CloneArtifact(item)
        slot.CatalogIndex = index
        catalog.Slots[index] = slot
        return
    }

    slot.Type = item.Type
    slot.Image = item.Image
    slot.Name = item.Name
    slot.Cost = item.Cost
    slot.Powers = slices.Clone(item.Powers)
    slot.Requirements = slices.Clone(item.Requirements)
    slot.CatalogIndex = index
}

// ApplyAvailabilityBitmap applies the 250-byte DOS save mask (1 = still
// available). Shorter or empty slices are ignored so a missing field does
// not wipe the catalog.
func (catalog *Catalog) ApplyAvailabilityBitmap(bits []byte) {
    if catalog == nil || len(bits) == 0 {
        return
    }

    for i := 0; i < len(catalog.Slots) && i < len(bits); i++ {
        catalog.available[i] = bits[i] != 0
    }
}

// WithAvailableNames returns a copy of the catalog whose availability is
// the set of names still remaining. Used to load old saves that only
// stored remaining item names.
func (catalog *Catalog) WithAvailableNames(names []string) *Catalog {
    if catalog == nil {
        return MakeCatalog(nil)
    }

    remaining := make(map[string]bool, len(names))
    for _, name := range names {
        remaining[name] = true
    }

    out := &Catalog{
        Slots: make([]*Artifact, len(catalog.Slots)),
        available: make([]bool, len(catalog.Slots)),
    }

    for i, slot := range catalog.Slots {
        out.Slots[i] = CloneArtifact(slot)
        if slot != nil {
            out.available[i] = remaining[slot.Name]
        }
    }

    return out
}

func (catalog *Catalog) Serialize() []SerializedArtifact {
    if catalog == nil {
        return nil
    }

    out := make([]SerializedArtifact, 0, len(catalog.Slots))
    for _, slot := range catalog.Slots {
        if slot == nil {
            out = append(out, SerializedArtifact{})
            continue
        }
        out = append(out, SerializeArtifact(slot))
    }

    return out
}

func (catalog *Catalog) AvailableMask() []bool {
    if catalog == nil {
        return nil
    }

    return slices.Clone(catalog.available)
}

func ReconstructCatalog(items []SerializedArtifact, available []bool, allSpells spellbook.Spells) *Catalog {
    catalog := &Catalog{
        Slots: make([]*Artifact, len(items)),
        available: make([]bool, len(items)),
    }

    for i, serialized := range items {
        art := ReconstructArtifact(&serialized, allSpells)
        if art != nil {
            art.CatalogIndex = i
        }
        catalog.Slots[i] = art
        if i < len(available) {
            catalog.available[i] = available[i]
        } else {
            catalog.available[i] = true
        }
    }

    return catalog
}

func CloneArtifact(src *Artifact) *Artifact {
    if src == nil {
        return nil
    }

    clone := *src
    clone.Powers = slices.Clone(src.Powers)
    clone.Requirements = slices.Clone(src.Requirements)
    return &clone
}

func isAbilityPowerType(powerType PowerType) bool {
    return powerType == PowerTypeAbility1 || powerType == PowerTypeAbility2 || powerType == PowerTypeAbility3
}

// samePower reports whether two powers are the same pick in the item editor
// (stat amount, ability, or spell-charge pair). Ability1/2/3 are treated as
// the same kind so a catalog record that stored every ability as Ability1
// still matches the itempow grouping.
func samePower(a Power, b Power) bool {
    if isAbilityPowerType(a.Type) && isAbilityPowerType(b.Type) {
        return a.Ability == b.Ability
    }

    if a.Type != b.Type {
        return false
    }

    if a.Type == PowerTypeSpellCharges {
        return a.Spell.Name == b.Spell.Name && a.Amount == b.Amount
    }

    return a.Amount == b.Amount
}

func (artifact *Artifact) ContainsPower(power Power) bool {
    if artifact == nil {
        return false
    }

    for _, existing := range artifact.Powers {
        if samePower(existing, power) {
            return true
        }
    }

    return false
}

// NormalizePowers replaces catalog-style powers with the matching itempow
// entries so the create-artifact UI can highlight and toggle them. Spell
// charges are kept as-is (they are not in itempow).
func NormalizePowers(item *Artifact, powerEntries []Power) {
    if item == nil {
        return
    }

    var normalized []Power
    for _, have := range item.Powers {
        matched := false
        for _, entry := range powerEntries {
            if samePower(have, entry) {
                normalized = append(normalized, entry)
                matched = true
                break
            }
        }

        if !matched {
            normalized = append(normalized, have)
        }
    }

    item.Powers = normalized
}

// RequirementsFromPowers builds the realm book requirements Insecticide
// writes: each realm's amount is the highest book cost among powers of
// that realm. Attribute bonuses do not contribute. This is the opposite
// of official ITEMMAKE, which wrote book counts in power-slot order.
func RequirementsFromPowers(powers []Power) []Requirement {
    amounts := make(map[data.MagicType]int)

    for _, power := range powers {
        if !isAbilityPowerType(power.Type) {
            continue
        }

        if power.Magic == data.MagicNone || power.Amount <= 0 {
            continue
        }

        if power.Amount > amounts[power.Magic] {
            amounts[power.Magic] = power.Amount
        }
    }

    order := []data.MagicType{
        data.NatureMagic,
        data.SorceryMagic,
        data.ChaosMagic,
        data.LifeMagic,
        data.DeathMagic,
    }

    var out []Requirement
    for _, magic := range order {
        if amounts[magic] > 0 {
            out = append(out, Requirement{MagicType: magic, Amount: amounts[magic]})
        }
    }

    return out
}

func FilterPowersForType(item *Artifact, artifactType ArtifactType, compatibilities map[Power]set.Set[ArtifactType]) {
    if item == nil {
        return
    }

    item.Type = artifactType
    item.Powers = slices.DeleteFunc(item.Powers, func (power Power) bool {
        if power.Type == PowerTypeSpellCharges {
            return artifactType != ArtifactTypeWand && artifactType != ArtifactTypeStaff
        }

        allowed, ok := compatibilities[power]
        if !ok {
            return false
        }

        return !allowed.Contains(artifactType)
    })
}

func ImageRange(kind ArtifactType) (int, int) {
    switch kind {
        case ArtifactTypeSword: return 0, 8
        case ArtifactTypeMace: return 9, 19
        case ArtifactTypeAxe: return 20, 28
        case ArtifactTypeBow: return 29, 37
        case ArtifactTypeStaff: return 38, 46
        case ArtifactTypeChain: return 47, 54
        case ArtifactTypePlate: return 55, 61
        case ArtifactTypeShield: return 62, 71
        case ArtifactTypeMisc: return 72, 106
        case ArtifactTypeWand: return 107, 115
    }

    return 0, 0
}

func ClampImageToType(item *Artifact) {
    if item == nil {
        return
    }

    low, high := ImageRange(item.Type)
    if item.Image < low || item.Image > high {
        item.Image = low
    }
}
