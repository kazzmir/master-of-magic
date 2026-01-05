package serialize

import (
    // "log"
    "io/fs"
    "time"
    "bufio"
    "compress/gzip"
    "encoding/json"
)

type SaveMetadata struct {
    Version int `json:"version"`
    Date time.Time `json:"date"`
    Name string `json:"name"`
}

func LoadMetadata(where fs.FS, path string) (SaveMetadata, bool) {
    file, err := where.Open(path)
    if err != nil {
        // log.Printf("Unable to open %s: %v", path, err)
        return SaveMetadata{}, false
    }
    defer file.Close()

    reader := bufio.NewReader(file)
    gzipReader, err := gzip.NewReader(reader)
    if err != nil {
        return SaveMetadata{}, false
    }
    defer gzipReader.Close()

    var data map[string]any
    decoder := json.NewDecoder(gzipReader)
    err = decoder.Decode(&data)
    if err != nil {
        return SaveMetadata{}, false
    }

    metadata, found := data["metadata"]
    if found {
        metadataMap, ok := metadata.(map[string]any)
        if ok {
            raw, err := json.Marshal(metadataMap)
            if err != nil {
                return SaveMetadata{}, false
            }

            var meta SaveMetadata
            err = json.Unmarshal(raw, &meta)
            if err != nil {
                return SaveMetadata{}, false
            }

            return meta, true
        }
    }

    return SaveMetadata{}, false
}

