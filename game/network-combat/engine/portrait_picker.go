package engine

import (
	"image"
	"image/color"

	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
)

func (editor *gameSetupUI) portraitImageFor(id int) *ebiten.Image {
	if img := editor.images[id]; img != nil {
		return img
	}
	if editor.imageCache != nil {
		if img, err := editor.imageCache.GetImage("wizards.lbx", id, 0); err == nil {
			editor.images[id] = img
		}
	}
	if editor.images[id] == nil {
		editor.images[id] = ebiten.NewImage(64, 80)
	}
	return editor.images[id]
}

func (editor *gameSetupUI) selectPortrait(id int) {
	if editor.portraitButton.GetWidget().Disabled || id < 0 || id >= len(protocol.PortraitNames) {
		return
	}
	if editor.draft.Portrait != id {
		editor.draft.Portrait = id
		editor.dirty = true
		editor.showPortrait()
		editor.refresh()
	}
	if editor.portraitWindow != nil {
		editor.portraitWindow.Close()
	}
}

func (editor *gameSetupUI) openPortraitPicker() {
	game := editor.match.game
	if editor.portraitButton.GetWidget().Disabled {
		return
	}
	if editor.portraitWindow == nil {
		editor.makePortraitPicker()
	}
	if game.ui.IsWindowOpen(editor.portraitWindow) {
		return
	}
	rect := game.ui.Container.GetWidget().Rect
	if rect.Empty() {
		rect = image.Rect(0, 0, 1100, 700)
	}
	width, height := min(500, max(360, rect.Dx()-40)), min(620, max(300, rect.Dy()-40))
	x, y := max(0, (rect.Dx()-width)/2), max(0, (rect.Dy()-height)/2)
	editor.portraitWindow.SetLocation(image.Rect(x, y, x+width, y+height))
	game.ui.AddWindow(editor.portraitWindow)
}

func (editor *gameSetupUI) makePortraitPicker() {
	theme := editor.match.game.ui.PrimaryTheme
	label := func(value string) *widget.Text { return widget.NewText(widget.TextOpts.Text(value, nil, color.White)) }
	editor.portraitGrid = widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(
		widget.GridLayoutOpts.Columns(3), widget.GridLayoutOpts.Weights([]float64{1, 1, 1}, nil),
		widget.GridLayoutOpts.Spacing(8, 8), widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(8)),
	)))
	editor.portraitGrid.GetWidget().SetTheme(theme)
	for id, name := range protocol.PortraitNames {
		button := newPortraitPreview(editor.portraitImageFor(id), func() { editor.selectPortrait(id) })
		editor.portraitChoices = append(editor.portraitChoices, button)
		cell := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1), widget.GridLayoutOpts.Weights([]float64{1}, nil), widget.GridLayoutOpts.Spacing(0, 4),
		)))
		nameLabel := label(name)
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter)(nameLabel)
		cell.AddChild(button, nameLabel)
		editor.portraitGrid.AddChild(cell)
	}
	background := uiimage.NewNineSliceColor(color.NRGBA{R: 24, G: 28, B: 36, A: 255})
	scroller := widget.NewScrollContainer(widget.ScrollContainerOpts.Content(editor.portraitGrid), widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{Idle: background, Disabled: background, Mask: uiimage.NewNineSliceColor(color.White)}))
	slider := widget.NewSlider(widget.SliderOpts.Direction(widget.DirectionVertical), widget.SliderOpts.MinMax(0, 1000),
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.MinSize(14, 40)), widget.SliderOpts.PageSizeFunc(func() int { return 200 }),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) { scroller.ScrollTop = float64(args.Current) / 1000 }))
	slider.Current = 0
	scroller.GetWidget().ScrolledEvent.AddHandler(func(args any) {
		if event, ok := args.(*widget.WidgetScrolledEventArgs); ok {
			slider.Current = max(0, min(1000, slider.Current-int(event.Y*100)))
			scroller.ScrollTop = float64(slider.Current) / 1000
		}
	})
	viewport := setupRow([]float64{1, 0})
	viewport.AddChild(scroller, slider)
	contents := lobbyColumn(1, 0)
	contents.GetWidget().SetTheme(theme)
	contents.SetBackgroundImage(background)
	closeButton := widget.NewButton(widget.ButtonOpts.Text("Close", nil, nil), widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { editor.portraitWindow.Close() }))
	contents.AddChild(viewport, closeButton)
	title := lobbyColumn(0)
	title.GetWidget().SetTheme(theme)
	title.SetBackgroundImage(uiimage.NewNineSliceColor(color.NRGBA{R: 42, G: 50, B: 64, A: 255}))
	title.AddChild(label("Choose wizard portrait"))
	editor.portraitWindow = widget.NewWindow(widget.WindowOpts.Contents(contents), widget.WindowOpts.TitleBar(title, 36),
		widget.WindowOpts.Draggable(), widget.WindowOpts.Resizeable(), widget.WindowOpts.Modal(), widget.WindowOpts.MinSize(360, 300))
	frame := editor.portraitWindow.GetContainer().(*widget.Container)
	frame.GetWidget().SetTheme(theme)
	frame.SetBackgroundImage(uiimage.NewBorderedNineSliceColor(color.NRGBA{}, color.NRGBA{R: 230, G: 215, B: 140, A: 255}, 2))
	widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1),
		widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}), widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(2))))(frame)
}
