package main

import (
    "os"
    "fmt"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type Editor struct {
    Data *terrain.TerrainData
}

func (editor *Editor) Update() error {
    var keys []ebiten.Key

    keys = make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                return ebiten.Termination
        }
    }


    return nil
}

func (editor *Editor) Draw(screen *ebiten.Image){
}

func (editor *Editor) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func MakeEditor(lbxFile *lbx.LbxFile) *Editor {
    data, err := terrain.ReadTerrainData(lbxFile)
    if err != nil {
        fmt.Printf("Could not read terrain data: %v\n", err)
        os.Exit(0)
    }

    return &Editor{
        Data: data,
    }
}

func main() {
    if len(os.Args) < 2 {
        fmt.Printf("Give the terrain.lbx file as an argument\n")
        return
    }

    path := os.Args[1]
    file, err := os.Open(path)
    if err != nil {
        fmt.Printf("Could not open lbx file %v: %v\n", path, err)
        return
    }
    lbxData, err := lbx.ReadLbx(file)
    if err != nil {
        fmt.Printf("Could read lbx file %v: %v\n", path, err)
        return
    }
    file.Close()

    editor := MakeEditor(&lbxData)

    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("map editor")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    
    err = ebiten.RunGame(editor)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
