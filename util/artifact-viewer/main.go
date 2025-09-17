package main

import (
    "fmt"
    "log"
    "image"
    "image/color"
    "cmp"
    "slices"

    "github.com/kazzmir/master-of-magic/lib/lbx"
    "github.com/kazzmir/master-of-magic/game/magic/artifact"
    "github.com/kazzmir/master-of-magic/game/magic/util"
    "github.com/kazzmir/master-of-magic/util/common"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
    "github.com/hajimehoshi/ebiten/v2/colorm"
    "github.com/hajimehoshi/ebiten/v2/vector"

    "github.com/ebitenui/ebitenui"
    "github.com/ebitenui/ebitenui/widget"
    ui_image "github.com/ebitenui/ebitenui/image"
)

func enlargeTransform(factor int) util.ImageTransformFunc {
    var f util.ImageTransformFunc

    f = func (original *image.Paletted) image.Image {
        newImage := image.NewPaletted(image.Rect(0, 0, original.Bounds().Dx() * factor, original.Bounds().Dy() * factor), original.Palette)

        for y := 0; y < original.Bounds().Dy(); y++ {
            for x := 0; x < original.Bounds().Dx(); x++ {
                colorIndex := original.ColorIndexAt(x, y)

                for dy := 0; dy < factor; dy++ {
                    for dx := 0; dx < factor; dx++ {
                        newImage.SetColorIndex(x * factor + dx, y * factor + dy, colorIndex)
                    }
                }
            }
        }

        return newImage
    }

    return f
}

type Engine struct {
    cache *lbx.LbxCache
    Artifacts []artifact.Artifact
    UI *ebitenui.UI
}

func MakeEngine(cache *lbx.LbxCache) (*Engine, error) {
    artifacts, err := artifact.ReadArtifacts(cache)

    if err != nil {
        return nil, err
    }

    log.Printf("Loaded %d artifacts", len(artifacts))

    engine := &Engine{
        cache: cache,
        Artifacts: artifacts,
    }

    engine.UI = engine.MakeUI()

    return engine, nil
}

func lighten(c color.Color, amount float64) color.Color {
    var change colorm.ColorM
    change.ChangeHSV(0, 1 - amount/100, 1 + amount/100)
    return change.Apply(c)
}

func makeNineImage(img *ebiten.Image, border int) *ui_image.NineSlice {
    width := img.Bounds().Dx()
    return ui_image.NewNineSliceSimple(img, border, width - border * 2)
}

func makeNineRoundedButtonImage(width int, height int, border int, col color.Color) *widget.ButtonImage {
    return &widget.ButtonImage{
        Idle: makeNineImage(makeRoundedButtonImage(width, height, border, col), border),
        Hover: makeNineImage(makeRoundedButtonImage(width, height, border, lighten(col, 50)), border),
        Pressed: makeNineImage(makeRoundedButtonImage(width, height, border, lighten(col, 20)), border),
    }
}

func padding(n int) *widget.Insets {
    return &widget.Insets{Top: n, Bottom: n, Left: n, Right: n}
}

func makeRoundedButtonImage(width int, height int, border int, col color.Color) *ebiten.Image {
    img := ebiten.NewImage(width, height)

    vector.DrawFilledRect(img, float32(border), 0, float32(width - border * 2), float32(height), col, true)
    vector.DrawFilledRect(img, 0, float32(border), float32(width), float32(height - border * 2), col, true)
    vector.DrawFilledCircle(img, float32(border), float32(border), float32(border), col, true)
    vector.DrawFilledCircle(img, float32(width-border), float32(border), float32(border), col, true)
    vector.DrawFilledCircle(img, float32(border), float32(height-border), float32(border), col, true)
    vector.DrawFilledCircle(img, float32(width-border), float32(height-border), float32(border), col, true)

    return img
}

func loadFont(size float64) (text.Face, error) {
    source, err := common.LoadFont()

    if err != nil {
        log.Fatal(err)
        return nil, err
    }

    return &text.GoTextFace{
        Source: source,
        Size:   size,
    }, nil
}

func (engine *Engine) GetArtifact(name string) (artifact.Artifact, bool) {
    for _, artifact := range engine.Artifacts {
        if artifact.Name == name {
            return artifact, true
        }
    }
    return artifact.Artifact{}, false
}

func (engine *Engine) MakeUI() *ebitenui.UI {
    imageCache := util.MakeImageCache(engine.cache)
    face, _ := loadFont(19)
    backgroundImage := ui_image.NewNineSliceColor(color.NRGBA{R: 32, G: 32, B: 32, A: 128})

    rootContainer := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
            widget.RowLayoutOpts.Spacing(12),
            widget.RowLayoutOpts.Padding(padding(5)),
        )),
        widget.ContainerOpts.BackgroundImage(backgroundImage),
        // widget.ContainerOpts.BackgroundImage(backgroundImageNine),
    )

    itemInfo := widget.NewContainer(
        widget.ContainerOpts.Layout(widget.NewRowLayout(
            widget.RowLayoutOpts.Direction(widget.DirectionVertical),
            widget.RowLayoutOpts.Spacing(4),
            widget.RowLayoutOpts.Padding(padding(5)),
        )),
        widget.ContainerOpts.BackgroundImage(makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5)),
    )

    updateItemInfo := func (name string) {
        useArtifact, ok := engine.GetArtifact(name)
        if !ok {
            return
        }

        itemInfo.RemoveChildren()
        itemInfo.AddChild(widget.NewText(widget.TextOpts.Text(useArtifact.Name, &face, color.NRGBA{R: 255, G: 255, B: 255, A: 255})))
        itemInfo.AddChild(widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Cost: %v", useArtifact.Cost), &face, color.NRGBA{R: 255, G: 255, B: 255, A: 255})))

        graphic := widget.NewGraphic(widget.GraphicOpts.Image(ebiten.NewImage(1, 1)))
        itemImage, err := imageCache.GetImageTransform("items.lbx", useArtifact.Image, 0, "enlarge", enlargeTransform(4))

        if err == nil {
            graphic.Image = itemImage
        } else {
            log.Printf("Unable to load artifact image for %v: %v", name, err)
        }

        itemInfo.AddChild(graphic)
        itemInfo.AddChild(widget.NewText(widget.TextOpts.Text(fmt.Sprintf("Type: %v", useArtifact.Type), &face, color.NRGBA{R: 255, G: 255, B: 255, A: 255})))
        for _, power := range useArtifact.Powers {
            itemInfo.AddChild(widget.NewText(widget.TextOpts.Text(power.Name, &face, color.NRGBA{R: 255, G: 255, B: 255, A: 255})))
        }
    }

    fakeImage := ui_image.NewNineSliceColor(color.NRGBA{R: 32, G: 32, B: 32, A: 255})

    artifactList := widget.NewList(
        widget.ListOpts.EntryFontFace(&face),

        widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(
            widget.WidgetOpts.LayoutData(widget.RowLayoutData{
                MaxHeight: 850,
            }),
            widget.WidgetOpts.MinSize(0, 850),
        )),

        widget.ListOpts.SliderParams(&widget.SliderParams{
            TrackImage: &widget.SliderTrackImage{
                Idle: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
                Hover: makeNineImage(makeRoundedButtonImage(20, 20, 5, color.NRGBA{R: 128, G: 128, B: 128, A: 255}), 5),
            },
            HandleImage: makeNineRoundedButtonImage(40, 40, 5, color.NRGBA{R: 0xad, G: 0x8d, B: 0x55, A: 0xff}),
        }),

        widget.ListOpts.HideHorizontalSlider(),

        widget.ListOpts.EntryLabelFunc(
            func (e any) string {
                item := e.(string)
                return item
            },
        ),

        widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
            entry := args.Entry.(string)
            log.Printf("Entry Selected: %v", entry)
            updateItemInfo(entry)
        }),

        widget.ListOpts.EntryColor(&widget.ListEntryColor{
            Selected: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
            Unselected: color.NRGBA{R: 0, G: 255, B: 0, A: 255},
        }),

        widget.ListOpts.ScrollContainerImage(&widget.ScrollContainerImage{
            Idle: ui_image.NewNineSliceColor(color.NRGBA{R: 64, G: 64, B: 64, A: 255}),
            Disabled: fakeImage,
            Mask: fakeImage,
        }),
    )

    for _, artifact := range slices.SortedFunc(slices.Values(engine.Artifacts), func (a artifact.Artifact, b artifact.Artifact) int {
        return cmp.Compare(a.Name, b.Name)
    }) {
        artifactList.AddEntry(artifact.Name)
    }

    itemInfo.AddChild(widget.NewText(widget.TextOpts.Text("Item", &face, color.NRGBA{R: 255, G: 255, B: 255, A: 255})))

    rootContainer.AddChild(artifactList)
    rootContainer.AddChild(itemInfo)

    ui := ebitenui.UI{
        Container: rootContainer,
    }

    return &ui
}

func (engine *Engine) Update() error {
    keys := inpututil.AppendJustPressedKeys(nil)

    for _, key := range keys {
        switch key {
            case ebiten.KeyEscape, ebiten.KeyCapsLock:
                return ebiten.Termination
        }
    }

    engine.UI.Update()

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image) {
    engine.UI.Draw(screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return outsideWidth, outsideHeight
}

func main(){
    cache := lbx.AutoCache()

    engine, err := MakeEngine(cache)
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    ebiten.SetWindowSize(1200, 900)
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
