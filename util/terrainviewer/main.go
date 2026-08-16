package main

import (
    "os"
    "fmt"
    "image"
    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/util/common"
    "github.com/kazzmir/master-of-magic/game/magic/terrain"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
)

const ScreenWidth = 1024
const ScreenHeight = 768

type ImageGPU struct {
    Raw image.Image
    GPU *ebiten.Image
}

type Viewer struct {
    Images []ImageGPU
    Data *terrain.TerrainData
    Font *text.GoTextFaceSource
    Choice int
    Counter uint64
    StartingRow int
    TileIndex int
}

func MakeViewer(data *terrain.TerrainData) *Viewer {
    var use []ImageGPU

    for _, img := range data.Images {
        use = append(use, ImageGPU{
            Raw: img,
            GPU: nil,
        })
    }

    font, err := common.LoadFont()
    if err != nil {
        fmt.Printf("Could not load font: %v\n", err)
    }

    return &Viewer{
        Data: data,
        Images: use,
        Font: font,
        Choice: 0,
        StartingRow: 0,
        TileIndex: 0,
    }
}

func (viewer *Viewer) TilesPerRow() int {
    // hack: why is +1 needed?
    return (ScreenWidth - 3) / (viewer.Images[0].Raw.Bounds().Dx() + 5) + 1
}

func (viewer *Viewer) TilesPerColumn() int {
    return (ScreenHeight - 110) / (viewer.Images[0].Raw.Bounds().Dy() + 5)
}

func (viewer *Viewer) Update() error {
    viewer.Counter += 1
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendPressedKeys(keys)

    moveRight := false
    moveLeft := false
    moveUp := false
    moveDown := false

    leftShift := inpututil.KeyPressDuration(ebiten.KeyShiftLeft) > 0

    if viewer.Counter % 3 == 0 && leftShift{

        for _, key := range keys {
            switch key {
                case ebiten.KeyRight: moveRight = true
                case ebiten.KeyLeft: moveLeft = true
                case ebiten.KeyUp: moveUp = true
                case ebiten.KeyDown: moveDown = true
            }
        }
    }

    keys = make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        switch key {
            case ebiten.KeyRight: moveRight = true
            case ebiten.KeyLeft: moveLeft = true
            case ebiten.KeyUp: moveUp = true
            case ebiten.KeyDown: moveDown = true
            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                return ebiten.Termination
        }
    }

    needUpdate := false

    if moveRight {
        needUpdate = true
        viewer.Choice += 1
        if viewer.Choice >= len(viewer.Images) {
            viewer.Choice = len(viewer.Images) - 1
        }
    }

    if moveLeft {
        needUpdate = true
        viewer.Choice -= 1
        if viewer.Choice < 0 {
            viewer.Choice = 0
        }
    }

    if moveUp {
        needUpdate = true
        viewer.Choice -= viewer.TilesPerRow()
        if viewer.Choice < 0 {
            viewer.Choice = 0
        }
    }

    if moveDown {
        needUpdate = true
        viewer.Choice += viewer.TilesPerRow()
        if viewer.Choice >= len(viewer.Images) {
            viewer.Choice = len(viewer.Images) - 1
        }
    }

    for viewer.Choice < viewer.StartingRow * viewer.TilesPerRow() {
        viewer.StartingRow -= 1
    }

    for viewer.Choice >= (viewer.StartingRow + viewer.TilesPerColumn()) * viewer.TilesPerRow() {
        viewer.StartingRow += 1
    }

    if needUpdate {
        for i, tile := range viewer.Data.Tiles {
            if tile.ContainsImageIndex(viewer.Choice) {
                viewer.TileIndex = i
                break
            }
        }
    }

    return nil
}

func (viewer *Viewer) Layout(outsideWidth int, outsideHeight int) (int, int) {
    return ScreenWidth, ScreenHeight
}

func (viewer *Viewer) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0x10, 0x10, 0x10, 0xff})

    face := &text.GoTextFace{Source: viewer.Font, Size: 15}
    op := &text.DrawOptions{}
    op.GeoM.Translate(1, 1)
    op.ColorScale.ScaleWithColor(color.White)
    text.Draw(screen, fmt.Sprintf("Terrain entry: %v/%v", viewer.Choice, len(viewer.Images)-1), face, op)

    if viewer.TileIndex != -1 {
        tile := viewer.Data.Tiles[viewer.TileIndex]
        op.GeoM.Translate(0, 20)
        text.Draw(screen, fmt.Sprintf("Tile %v (0x%x)", viewer.TileIndex, viewer.TileIndex), face, op)
        op.GeoM.Translate(0, 20)
        text.Draw(screen, fmt.Sprintf("Center %v", tile.Tile.GetDirection(terrain.Center)), face, op)

        face2 := &text.GoTextFace{Source: viewer.Font, Size: 10}

        op.GeoM.Reset()
        op.GeoM.Translate(ScreenWidth/2, 1)
        text.Draw(screen, fmt.Sprintf("%v", tile.Tile.GetDirection(terrain.North)), face2, op)

        op.GeoM.Reset()
        op.GeoM.Translate(ScreenWidth/2 - 120, 1)
        text.Draw(screen, fmt.Sprintf("%v", tile.Tile.GetDirection(terrain.NorthWest)), face2, op)

        op.GeoM.Reset()
        op.GeoM.Translate(ScreenWidth/2 + 180, 1)
        text.Draw(screen, fmt.Sprintf("%v", tile.Tile.GetDirection(terrain.NorthEast)), face2, op)

        op.GeoM.Reset()
        op.GeoM.Translate(ScreenWidth/2 - 140, 45)
        text.Draw(screen, fmt.Sprintf("%v", tile.Tile.GetDirection(terrain.West)), face2, op)

        op.GeoM.Reset()
        op.GeoM.Translate(ScreenWidth/2 - 125, 90)
        text.Draw(screen, fmt.Sprintf("%v", tile.Tile.GetDirection(terrain.SouthWest)), face2, op)

        op.GeoM.Reset()
        op.GeoM.Translate(ScreenWidth/2, 90)
        text.Draw(screen, fmt.Sprintf("%v", tile.Tile.GetDirection(terrain.South)), face2, op)

        op.GeoM.Reset()
        op.GeoM.Translate(ScreenWidth/2 + 180, 90)
        text.Draw(screen, fmt.Sprintf("%v", tile.Tile.GetDirection(terrain.SouthEast)), face2, op)

        op.GeoM.Reset()
        op.GeoM.Translate(ScreenWidth/2 + 195, 45)
        text.Draw(screen, fmt.Sprintf("%v", tile.Tile.GetDirection(terrain.East)), face2, op)
    }

    var options ebiten.DrawImageOptions
    x := float64(3)
    y := float64(110)

    options.GeoM.Scale(4, 4)
    options.GeoM.Translate(ScreenWidth/2, 15)
    if viewer.Images[viewer.Choice].GPU != nil {
        screen.DrawImage(viewer.Images[viewer.Choice].GPU, &options)
    }

    startPosition := viewer.StartingRow * viewer.TilesPerRow()

    for i := startPosition; i < len(viewer.Images); i++ {
        img := viewer.Images[i]
        if img.GPU == nil {
            img.GPU = ebiten.NewImageFromImage(img.Raw)
            viewer.Images[i] = img
        }

        options.GeoM.Reset()
        options.GeoM.Translate(x, y)
        screen.DrawImage(img.GPU, &options)

        if i == viewer.Choice {
            width := float32(img.Raw.Bounds().Dx())
            height := float32(img.Raw.Bounds().Dy())
            vector.StrokeRect(screen, float32(x-1), float32(y-1), width+2, height+2, 1.5, color.White, true)
        }

        x += float64(img.Raw.Bounds().Dx()) + 5
        if x >= float64(ScreenWidth - img.Raw.Bounds().Dx()) {
            x = 3
            y += float64(img.Raw.Bounds().Dy()) + 5
        }

        if y >= float64(ScreenHeight) {
            break
        }
    }
}

func display(lbxData lbx.LbxFile) error {
    /*
    images, err := lbxData.ReadTerrainImages(0)
    if err != nil {
        return err
    }
    */
    data, err := terrain.ReadTerrainData(&lbxData)
    if err != nil {
        return err
    }

    ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
    ebiten.SetWindowTitle("terrain viewer")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    viewer := MakeViewer(data)

    err = ebiten.RunGame(viewer)

    return err
}

func main(){

    var lbxFile *lbx.LbxFile
    var err error

    if len(os.Args) > 1 {
        file, err := os.Open(os.Args[1])
        if err != nil {
            fmt.Printf("Could not open lbx file: %v\n", err)
            return
        }
        use, err := lbx.ReadLbx(file)
        if err != nil {
            fmt.Printf("Could not read lbx file: %v\n", err)
            return
        }
        lbxFile = &use
        file.Close()
    } else {
        cache := lbx.AutoCache()

        lbxFile, err = cache.GetLbxFile("terrain.lbx")
        if err != nil {
            fmt.Printf("Could not load terrain.lbx: %v\n", err)
            return
        }
    }

    err = display(*lbxFile)
    if err != nil {
        fmt.Printf("Error displaying lbx: %v\n", err)
        return
    }
}
