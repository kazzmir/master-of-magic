package ui

import (
	"image"
	"math"
	"testing"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestAffordableCheckboxDefaultsChecked(t *testing.T) {
	var changes []bool
	checkbox := Checkbox("Affordable units", nil, true, func(checked bool) {
		changes = append(changes, checked)
	})
	event.ExecuteDeferred()
	if checkbox.State() != widget.WidgetChecked || len(changes) != 0 {
		t.Fatal("checkbox should start checked without triggering a change")
	}
	checkbox.Click()
	event.ExecuteDeferred()
	checkbox.Click()
	event.ExecuteDeferred()
	if len(changes) != 2 || changes[0] || !changes[1] {
		t.Fatalf("expected uncheck then check, got %v", changes)
	}
}

func TestIconStripFitsAvailableSpace(t *testing.T) {
	for _, test := range []struct {
		name                                                  string
		totalWidth, minWidth, maxHeight, count, width, height int
		scale, gap                                            float64
	}{
		{"natural size", 100, 20, 40, 5, 200, 64, 1, 4},
		{"close gaps", 100, 20, 40, 5, 108, 64, 1, 2},
		{"overlap before shrinking", 100, 20, 40, 5, 80, 64, 1, -5},
		{"half visible", 100, 20, 40, 5, 60, 64, 1, -10},
		{"shrink after overlap limit", 100, 20, 40, 5, 30, 64, .5, -5},
		{"mixed widths", 140, 20, 40, 3, 120, 64, 1, -10},
		{"single book", 20, 20, 40, 1, 10, 64, .5, 0},
		{"fit height", 100, 20, 40, 5, 200, 20, .5, 4},
		{"no width", 100, 20, 40, 5, 0, 64, 0, 0},
		{"no books", 0, 0, 0, 0, 200, 64, 0, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			scale, gap := fitIconStrip(test.totalWidth, test.minWidth, test.maxHeight, test.count, test.width, test.height, .5)
			if math.Abs(scale-test.scale) > 1e-9 || math.Abs(gap-test.gap) > 1e-9 {
				t.Fatalf("got scale %v gap %v, want %v and %v", scale, gap, test.scale, test.gap)
			}
			if gap < -float64(test.minWidth)*scale/2-1e-9 {
				t.Fatal("more than half a book is hidden")
			}
		})
	}
}

func TestArmyPreviewOverlapLimit(t *testing.T) {
	// Five 20px icons occupy 80px when each overlaps the next by 25%.
	for _, width := range []int{200, 100, 90, 80, 60, 20, 1} {
		scale, gap := fitIconStrip(100, 20, 40, 5, width, 40, .25)
		if gap < -20*scale*.25-1e-9 {
			t.Fatalf("width %d: icons overlap by more than 25%%", width)
		}
		if used := 100*scale + 4*gap; used > float64(width)+1e-9 {
			t.Fatalf("width %d: icons overflow with used width %v", width, used)
		}
		if width >= 80 && scale != 1 {
			t.Fatalf("width %d: icons shrank before overlap limit", width)
		}
		if width < 80 && scale >= 1 {
			t.Fatalf("width %d: icons must shrink after overlap limit", width)
		}
	}
}

func TestHeaderBooksResizeWithoutOverflow(t *testing.T) {
	header := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(
		widget.GridLayoutOpts.Columns(2),
		widget.GridLayoutOpts.Weights([]float64{0, 1}, []float64{1}),
		widget.GridLayoutOpts.Spacing(8, 0),
	)))
	label := widget.NewGraphic(widget.GraphicOpts.Image(ebiten.NewImage(300, 20)))
	books := NewIconStrip()
	header.AddChild(label, books)
	icon := ebiten.NewImage(20, 40)
	for _, count := range []int{0, 5, 55} {
		images := make([]*ebiten.Image, count)
		for i := range images {
			images[i] = icon
		}
		books.SetImages(images)
		for _, width := range []int{1200, 500, 320, 308} {
			header.SetLocation(image.Rect(0, 0, width, 64))
			header.RequestRelayout()
			header.Render(ebiten.NewImage(width, 64))
			rect := books.GetWidget().Rect
			total, height, n, minWidth := books.imageSize()
			scale, gap := fitIconStrip(total, minWidth, height, n, rect.Dx(), rect.Dy(), books.MaxOverlap)
			used := float64(total)*scale + float64(max(0, n-1))*gap
			if rect.Min.X < 308 || rect.Max.X > width || used > float64(rect.Dx())+1e-9 {
				t.Fatalf("%d books at width %d overflow: bounds %v, used %v", count, width, rect, used)
			}
		}
	}
	if icon.Bounds().Size() != image.Pt(20, 40) {
		t.Fatal("source book image was modified")
	}
}

func TestColumnFillsWidth(t *testing.T) {
	column := Column()
	child := widget.NewGraphic(widget.GraphicOpts.Image(ebiten.NewImage(20, 10)))
	column.AddChild(child)
	column.SetLocation(image.Rect(0, 0, 300, 100))
	column.RequestRelayout()
	column.Render(ebiten.NewImage(300, 100))
	if got := child.GetWidget().Rect.Dx(); got != 288 {
		t.Fatalf("child width = %d, want 288 after padding", got)
	}
}

func TestRecruitmentViewportFillsRemainingHeight(t *testing.T) {
	shop := Column(0, 0, 1)
	controls := Column(0, 1)
	list := Column(1)
	catalog := Column(0, 0, 1)
	row := ScrollRow(1)
	graphic := func(width, height int) *widget.Graphic {
		return widget.NewGraphic(widget.GraphicOpts.Image(ebiten.NewImage(width, height)))
	}
	viewport := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(graphic(100, 2000)),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle: SolidImage(32, 32, 32), Mask: SolidImage(255, 255, 255),
		}),
	)
	scrollbar := graphic(10, 10)
	row.AddChild(viewport, scrollbar)
	catalog.AddChild(graphic(100, 20), graphic(100, 20), row)
	list.AddChild(catalog)
	controls.AddChild(graphic(100, 20), list)
	shop.AddChild(graphic(100, 20), graphic(100, 20), controls)

	for _, height := range []int{600, 1000, 450} {
		shop.SetLocation(image.Rect(0, 0, 800, height))
		shop.Render(ebiten.NewImage(800, height))
		// Four padded columns separate the viewport from the tab's bottom.
		wantBottom := height - 4*6
		if got := viewport.GetWidget().Rect.Max.Y; got != wantBottom {
			t.Fatalf("window height %d: viewport bottom = %d, want %d", height, got, wantBottom)
		}
		if scrollbar.GetWidget().Rect.Min.Y != viewport.GetWidget().Rect.Min.Y || scrollbar.GetWidget().Rect.Max.Y != wantBottom {
			t.Fatalf("scrollbar does not track viewport: %v", scrollbar.GetWidget().Rect)
		}
	}
}

func TestArmyAndUpgradePanelsFillHeight(t *testing.T) {
	graphic := func(width, height int) *widget.Graphic {
		return widget.NewGraphic(widget.GraphicOpts.Image(ebiten.NewImage(width, height)))
	}
	root := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(
		widget.GridLayoutOpts.Columns(2),
		widget.GridLayoutOpts.Weights([]float64{2, 3}, []float64{1}),
		widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(8)),
	)))
	army := Column(0, 1)
	armyRow := ScrollRow(1)
	armyViewport, armyBar := graphic(100, 2000), graphic(10, 10)
	armyRow.AddChild(armyViewport, armyBar)
	army.AddChild(graphic(100, 20), armyRow)

	details := Column(0, 1)
	upgrades := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(
		widget.GridLayoutOpts.Columns(2),
		widget.GridLayoutOpts.Weights([]float64{1, 2}, []float64{1}),
	)))
	applied, available := Column(0, 1), Column(0, 1)
	appliedViewport := graphic(100, 2000)
	availableViewport, upgradeBar := graphic(100, 2000), graphic(10, 10)
	upgradeRow := ScrollRow(1)
	upgradeRow.AddChild(availableViewport, upgradeBar)
	applied.AddChild(graphic(100, 20), appliedViewport)
	available.AddChild(graphic(100, 20), upgradeRow)
	upgrades.AddChild(applied, available)
	details.AddChild(graphic(100, 220), upgrades)
	root.AddChild(army, details)

	for _, height := range []int{600, 1000, 500} {
		root.SetLocation(image.Rect(0, 0, 1200, height))
		root.Render(ebiten.NewImage(1200, height))
		for _, panel := range []struct {
			name   string
			widget widget.PreferredSizeLocateableWidget
			bottom int
		}{
			{"army", armyViewport, height - 14},
			{"army scrollbar", armyBar, height - 14},
			{"applied enchantments", appliedViewport, height - 20},
			{"available enchantments", availableViewport, height - 20},
			{"upgrade scrollbar", upgradeBar, height - 20},
		} {
			if got := panel.widget.GetWidget().Rect.Max.Y; got != panel.bottom {
				t.Errorf("height %d: %s bottom = %d, want %d", height, panel.name, got, panel.bottom)
			}
		}
	}
}

func TestMagicSpellGridFillsRemainingSpace(t *testing.T) {
	shop := Column(0, 0, 0, 1)
	graphic := func(width, height int) *widget.Graphic {
		return widget.NewGraphic(widget.GraphicOpts.Image(ebiten.NewImage(width, height)))
	}
	realm := Column(0, 1)
	row := ScrollRow(1)
	spells := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(
		widget.GridLayoutOpts.Columns(3),
		widget.GridLayoutOpts.Weights([]float64{1, 1, 1}, nil),
		widget.GridLayoutOpts.Spacing(12, 12),
		widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(8)),
	)))
	cards := make([]*widget.Graphic, 6)
	for i := range cards {
		cards[i] = graphic(100, 180)
		spells.AddChild(cards[i])
	}
	viewport := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(spells),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle: SolidImage(32, 32, 32), Mask: SolidImage(255, 255, 255),
		}),
	)
	scrollbar := graphic(10, 10)
	row.AddChild(viewport, scrollbar)
	realm.AddChild(graphic(100, 20), row)
	shop.AddChild(graphic(100, 20), graphic(100, 140), graphic(100, 40), realm)

	for _, size := range []image.Point{{1200, 600}, {1600, 1000}, {1000, 500}} {
		shop.SetLocation(image.Rectangle{Max: size})
		shop.Render(ebiten.NewImage(size.X, size.Y))
		if bottom := viewport.GetWidget().Rect.Max.Y; bottom != size.Y-12 {
			t.Fatalf("size %v: viewport bottom = %d, want %d", size, bottom, size.Y-12)
		}
		if scrollbar.GetWidget().Rect.Max.Y != viewport.GetWidget().Rect.Max.Y {
			t.Fatal("scrollbar does not fill viewport height")
		}
		for i := 0; i < 3; i++ {
			top, lower := cards[i].GetWidget().Rect, cards[i+3].GetWidget().Rect
			if top.Min.X != lower.Min.X || top.Dx() != lower.Dx() || lower.Min.Y != top.Max.Y+12 {
				t.Fatalf("size %v: inconsistent spell column %d", size, i)
			}
		}
		if cards[2].GetWidget().Rect.Max.X != viewport.GetWidget().Rect.Max.X-8 {
			t.Fatalf("size %v: spell columns do not fill viewport width", size)
		}
	}
}

func TestScrollRowCapsContentHeight(t *testing.T) {
	row := ScrollRow()
	content := widget.NewGraphic(
		widget.GraphicOpts.Image(ebiten.NewImage(100, 800)),
		widget.GraphicOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.GridLayoutData{MaxHeight: 400})),
	)
	bar := widget.NewGraphic(widget.GraphicOpts.Image(ebiten.NewImage(10, 400)))
	row.AddChild(content, bar)
	_, height := row.PreferredSize()
	if height != 400 {
		t.Fatalf("row height = %d, want capped height 400", height)
	}
}
