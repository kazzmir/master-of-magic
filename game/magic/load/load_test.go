package load

import (
    "os"
    "testing"
    "bytes"
    "io"
)

func BenchmarkGzipLoadFunction(bench *testing.B) {
    file, err := os.Open("files/save1.gam.gz")
    if err != nil {
        bench.Fatal(err)
    }
    defer file.Close()

    var buffer bytes.Buffer
    _, err = io.Copy(&buffer, file)
    if err != nil {
        bench.Fatal(err)
    }

    bench.ResetTimer()
    for bench.Loop() {
        _, err = LoadSaveGame(bytes.NewReader(buffer.Bytes()))
        if err != nil {
            bench.Fatal(err)
        }
    }
}

/*
func BenchmarkUncompressedLoadFunction(bench *testing.B) {
    file, err := os.Open("files/save1.gam")
    if err != nil {
        bench.Fatal(err)
    }

    var buffer bytes.Buffer
    _, err = io.Copy(&buffer, file)
    if err != nil {
        bench.Fatal(err)
    }

    bench.ResetTimer()
    for bench.Loop() {
        _, err = LoadSaveGame(bytes.NewReader(buffer.Bytes()))
        if err != nil {
            bench.Fatal(err)
        }
    }
}
*/
