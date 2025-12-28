package maplib

import (
)

func SerializeMap(useMap *Map) map[string]any {

    return map[string]any{
        "width": useMap.Width(),
        "height": useMap.Height(),
        "map": useMap.Map.Terrain,
    }
}
