package mouse

import (
    // "log"
    "fmt"
    "bytes"

    "image/color"

    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
)

// a collection of all mouse images
type MouseData struct {
    Normal *ebiten.Image
    Magic *ebiten.Image
    Error *ebiten.Image
    Arrow *ebiten.Image
    Attack *ebiten.Image
    Wait *ebiten.Image
    Move *ebiten.Image
    Cast []*ebiten.Image
}

func MakeMouseData(cache *lbx.LbxCache) (*MouseData, error) {
    normal, err := GetMouseNormal(cache)
    if err != nil {
        return nil, err
    }
    magic, err := GetMouseMagic(cache)
    if err != nil {
        return nil, err
    }
    errorImage, err := GetMouseError(cache)
    if err != nil {
        return nil, err
    }
    arrow, err := GetMouseArrow(cache)
    if err != nil {
        return nil, err
    }
    attack, err := GetMouseAttack(cache)
    if err != nil {
        return nil, err
    }
    wait, err := GetMouseWait(cache)
    if err != nil {
        return nil, err
    }
    move, err := GetMouseMove(cache)
    if err != nil {
        return nil, err
    }
    cast, err := GetMouseCast(cache)
    if err != nil {
        return nil, err
    }

    return &MouseData{
        Normal: normal,
        Magic: magic,
        Error: errorImage,
        Arrow: arrow,
        Attack: attack,
        Wait: wait,
        Move: move,
        Cast: cast,
    }, nil
}

// pass in an entry from fonts.lbx within range 2-8
func readMousePics(data []byte) ([]*ebiten.Image, error) {
    if len(data) < 5376 {
        return nil, fmt.Errorf("data is too short")
    }

    var mainPalette color.Palette

    paletteData := data[0:256*3]
    for i := 0; i < 256; i++ {
        r := paletteData[i*3]
        g := paletteData[i*3+1]
        b := paletteData[i*3+2]
        // log.Printf("palette[%d] = %d, %d, %d", i, r, g, b)
        mainPalette = append(mainPalette, color.RGBA{R: r, G: g, B: b, A: 255})
    }

    // make transparent
    mainPalette[0] = color.RGBA{R: 0, G: 0, B: 0, A: 0}

    // 32 arrays of 16 colors
    fontColors := data[256*3:256*3 + 1280-768]
    _ = fontColors

    // FIXME: what to do with fontColors and mainPalette?

    // each pointer is 0x100 bytes
    mouseData := data[1280:5376]

    // remap colors start at offset 5376

    length := 0x100

    var mousePics []*ebiten.Image

    usePalette := lbx.GetDefaultPalette()
    for i := 0; i < 16; i++ {
        mouse := mouseData[i*length:i*length + length]
        pic := ebiten.NewImage(16, 16)

        reader := bytes.NewReader(mouse)

        for x := 0; x < 16; x++ {
            for y := 0; y < 16; y++ {
                colorIndex, err := reader.ReadByte()
                if err != nil {
                    return nil, err
                }

                color := usePalette[colorIndex]
                pic.Set(x, y, color)
            }
        }

        mousePics = append(mousePics, pic)
    }

    return mousePics, nil
}

func ReadMouseImages(fontsLbx *lbx.LbxFile, entry int) ([]*ebiten.Image, error) {
    data, err := fontsLbx.RawData(entry)
    if err != nil {
        return nil, err
    }

    return readMousePics(data)
}

func GetMouseImages(cache *lbx.LbxCache, entry int) ([]*ebiten.Image, error){
    fontsLbx, err := cache.GetLbxFile("fonts.lbx")
    if err != nil {
        return nil, err
    }

    return ReadMouseImages(fontsLbx, entry)
}

func GetSingleImage(cache *lbx.LbxCache, entry int, index int) (*ebiten.Image, error) {
    images, err := GetMouseImages(cache, entry)
    if err != nil {
        return nil, err
    }
    if len(images) > index {
        return images[index], nil
    }

    return nil, fmt.Errorf("no image found at index %v", index)
}

func GetMouseNormal(cache *lbx.LbxCache) (*ebiten.Image, error) {
    return GetSingleImage(cache, 2, 0)
}

func GetMouseMagic(cache *lbx.LbxCache) (*ebiten.Image, error) {
    return GetSingleImage(cache, 2, 1)
}

func GetMouseError(cache *lbx.LbxCache) (*ebiten.Image, error) {
    return GetSingleImage(cache, 2, 2)
}

func GetMouseArrow(cache *lbx.LbxCache) (*ebiten.Image, error) {
    return GetSingleImage(cache, 2, 3)
}

func GetMouseAttack(cache *lbx.LbxCache) (*ebiten.Image, error) {
    return GetSingleImage(cache, 2, 4)
}

func GetMouseWait(cache *lbx.LbxCache) (*ebiten.Image, error) {
    return GetSingleImage(cache, 2, 5)
}

func GetMouseMove(cache *lbx.LbxCache) (*ebiten.Image, error) {
    return GetSingleImage(cache, 2, 6)
}

func GetMouseCast(cache *lbx.LbxCache) ([]*ebiten.Image, error) {
    images, err := GetMouseImages(cache, 2)
    if err != nil {
        return nil, err
    }

    if len(images) >= 13 {
        return images[8:13], nil
    }

    return nil, fmt.Errorf("not enough images")
}
