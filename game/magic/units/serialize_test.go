package units

import (
    "testing"
    "encoding/json"
)

func TestSerialize(test *testing.T) {
    unit := LizardSwordsmen
    serialized := SerializeUnit(unit)

    out, err := json.Marshal(serialized)
    if err != nil {
        test.Fatalf("Failed to serialize unit: %v", err)
    }

    var deserialized SerializedUnit

    err = json.Unmarshal(out, &deserialized)
    if err != nil {
        test.Fatalf("Failed to deserialize unit: %v", err)
    }

    newUnit := DeserializeUnit(deserialized)
    if !newUnit.Equals(unit) {
        test.Fatalf("Deserialized unit does not match original. Got %+v, want %+v", newUnit, unit)
    }
}
